// Package cryptosvc 是密码学服务层。
//
// 本包是**唯一**允许 import tongsuo-go 的层：所有密码学运算都在这里
// **转调 Tongsuo**。其它层（tools / api）只经由本包的服务方法使用，
// 不得自行拼装密码学运算，也不得直接触碰原生库。
//
// 生命周期模型见 docs/standards/03-crypto.md：
// 每个请求在 api 层创建一个 Session，句柄登记到 Session；
// 请求结束时 Close() 按**注册逆序**释放，且必须**幂等**。
package cryptosvc

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// Session 是单个请求的密码学句柄作用域。
//
// Session **不**感知具体工具，也不感知 HTTP：它只管理句柄释放、警告累积
// 与重操作并发许可。
type Session struct {
	// ctx 是创建 Session 时的请求上下文。
	//
	// 释放器与重操作许可都据此判断请求是否已超时。
	ctx context.Context

	// sems 指向重操作信号量集合；默认为全局单例。
	sems *semaphores

	mu       sync.Mutex
	warnings []toolerr.Warning
	closers  []func() error
	orphans  []Orphan
	closed   bool
	closeErr error

	// heavyStart 记录当前持有的重操作许可的起始时刻。
	heavyStart time.Time
}

// Orphan 记录一次「请求已超时但底层运算仍在跑」的重操作。
//
// cgo 调用**不可中断**：超时后请求立即返回 504，底层工作继续跑完。
// 调用方据此写一条日志（只记操作名与耗时，不记任何输入）。
type Orphan struct {
	// Operation 是操作名，例如 "rsa-keygen"。
	Operation string
	// DurationMs 是从取得许可到释放许可的时长（毫秒）。
	DurationMs int64
}

// NewSession 创建一个 Session。
//
// ctx 会被存入 Session，并用于后续的 ctx 注入（见 Context）。
func NewSession(ctx context.Context) *Session {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Session{
		ctx:  ctx,
		sems: defaultSemaphores,
	}
}

// Context 返回带 Session 的上下文。
//
// api 层应把它传给 handler，handler 再用 FromContext 取回 Session。
func (s *Session) Context() context.Context {
	return WithSession(s.ctx, s)
}

// Defer 登记一个释放器。
//
// 释放器按**注册逆序**执行（LIFO）。重复登记同一个函数是允许的。
func (s *Session) Defer(closer func() error) {
	if closer == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		// Session 已关闭：立即释放，避免泄漏。
		_ = closer()
		return
	}
	s.closers = append(s.closers, closer)
}

// RegisterKey 把私钥句柄登记到 Session。
//
// 请求结束时 Session 会自动释放它，因此 handler **不**需要（也不应）再手动关闭。
func (s *Session) RegisterKey(k *PrivateKey) {
	if s == nil || k == nil {
		return
	}
	s.Defer(func() error { return k.Close() })
}

// AddWarning 追加一条警告。
//
// 警告**不**改变 HTTP 状态，只出现在响应的 meta.warnings 中。
func (s *Session) AddWarning(code toolerr.WarningCode, params map[string]any) {
	if code == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.warnings = append(s.warnings, toolerr.NewWarning(code, params))
}

// Warnings 返回已累积的警告副本。
func (s *Session) Warnings() []toolerr.Warning {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]toolerr.Warning, len(s.warnings))
	copy(out, s.warnings)
	return out
}

// Orphans 返回超时后仍在运行的重操作记录。
func (s *Session) Orphans() []Orphan {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Orphan, len(s.orphans))
	copy(out, s.orphans)
	return out
}

// Close 释放全部句柄。
//
// 释放顺序与注册顺序**相反**（LIFO），且**幂等**：重复调用只返回首次的结果。
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		err := s.closeErr
		s.mu.Unlock()
		return err
	}
	s.closed = true
	closers := s.closers
	s.closers = nil
	s.mu.Unlock()

	var errs []error
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i](); err != nil {
			errs = append(errs, err)
		}
	}

	joined := errors.Join(errs...)

	s.mu.Lock()
	s.closeErr = joined
	s.mu.Unlock()
	return joined
}

// Acquire 申请一个重操作许可。
//
// 重操作（密钥生成、证书签发等）并发上限为 2（见 docs/standards/03-crypto.md §12）。
// 申请**必须**响应 ctx 取消；取不到许可且 ctx 已取消时返回 TIMEOUT。
func (s *Session) Acquire(ctx context.Context) error {
	if ctx == nil {
		ctx = s.ctx
	}
	// 快路径：许可空闲。
	select {
	case s.sems.heavy <- struct{}{}:
		s.markHeavyStart()
		return nil
	default:
	}
	// 慢路径：等待许可，期间响应取消。
	select {
	case s.sems.heavy <- struct{}{}:
		s.markHeavyStart()
		return nil
	case <-ctx.Done():
		return toolerr.Newf(toolerr.CodeTimeout, map[string]any{
			"reason": "waiting for crypto semaphore",
		})
	}
}

// Release 归还重操作许可。
//
// cgo 调用不可中断：若请求在任务期间超时，任务仍会跑完，
// 此时记录一条 orphan 供调用方写日志。许可**一定**被归还。
func (s *Session) Release(operation string) {
	s.mu.Lock()
	select {
	case <-s.sems.heavy:
	default:
		// 未持有许可：直接返回，避免误归还。
		s.mu.Unlock()
		return
	}
	if s.ctx.Err() != nil {
		s.orphans = append(s.orphans, Orphan{
			Operation: operation,
			DurationMs: time.Since(s.heavyStart).Milliseconds(),
		})
	}
	s.mu.Unlock()
}

// markHeavyStart 记录重操作起始时刻。
func (s *Session) markHeavyStart() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.heavyStart = time.Now()
}
