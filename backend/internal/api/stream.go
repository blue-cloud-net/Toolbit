package api

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/network"
)

// 流式端点的时间参数。
const (
	// firstFrameTimeout 是等待客户端首帧的上限。
	firstFrameTimeout = 5 * time.Second
	// streamBudget 是单次任务的总预算。
	streamBudget = 30 * time.Second
	// streamIdleTimeout 是空闲多久后关闭。
	streamIdleTimeout = 60 * time.Second
	// streamPingInterval 是心跳间隔。
	streamPingInterval = 30 * time.Second
	// maxStreamSlotsPerIP 是同一 IP 的并发流式连接上限。
	maxStreamSlotsPerIP = 1
)

// streamFrame 是服务端帧。
//
// 三种 `type`：`progress` / `result` / `error`，与 docs/api/README.md §11.2 一致。
type streamFrame struct {
	// Type 是帧类型。
	Type string `json:"type"`
	// Phase 是进度阶段（仅 progress）。
	Phase string `json:"phase,omitempty"`
	// Done 是已完成阶段数（仅 progress）。
	Done int `json:"done,omitempty"`
	// Total 是阶段总数（仅 progress）。
	Total int `json:"total,omitempty"`
	// Data 是结果（仅 result）。
	Data any `json:"data,omitempty"`
	// Error 是错误（仅 error）。
	Error *streamError `json:"error,omitempty"`
}

// streamError 是错误帧的载荷。
//
// 与统一信封的 error 结构一致：只含错误码、注册表文案与结构化参数。
type streamError struct {
	// Code 是错误码。
	Code string `json:"code"`
	// Message 是注册表文案。
	Message string `json:"message"`
	// Params 是可插值参数。
	Params map[string]any `json:"params,omitempty"`
}

// streamLimiter 限制同一 IP 的并发流式连接数。
//
// 规范要求「同一 IP 同时最多 1 条」：流式连接会长期占用 goroutine 与出网连接，
// 必须独立于速率限流再限一层。
type streamLimiter struct {
	mu     sync.Mutex
	active map[string]int
}

// newStreamLimiter 构造并发流限流器。
func newStreamLimiter() *streamLimiter {
	return &streamLimiter{active: make(map[string]int)}
}

// acquire 尝试占用一个槽位。
func (l *streamLimiter) acquire(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.active[key] >= maxStreamSlotsPerIP {
		return false
	}
	l.active[key]++
	return true
}

// release 归还槽位。
func (l *streamLimiter) release(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.active[key] <= 1 {
		delete(l.active, key)
		return
	}
	l.active[key]--
}

// handleToolStream 是流式工具端点。
//
// 目前只有 `ssl-check` 使用：`GET /api/v1/tools/ssl-check/stream`。
//
// 生命周期按 docs/api/README.md §11.4：
// 首帧 5 秒超时、总预算 30 秒、心跳 30 秒、空闲 60 秒关闭、
// 同一 IP 同时最多 1 条连接。
func (s *Server) handleToolStream(c *gin.Context) {
	slug := c.Param("slug")
	httpx.SetSlug(c, slug)

	tool, ok := s.catalog.Tool(slug)
	if !ok {
		httpx.Fail(c, toolerr.Newf(toolerr.CodeToolNotFound, map[string]any{"slug": slug}))
		return
	}
	if !tool.Enabled {
		httpx.Fail(c, toolerr.Newf(toolerr.CodeToolDisabled, map[string]any{"slug": slug}))
		return
	}
	if tool.Transport != catalog.TransportWebSocket {
		// 非流式工具没有 /stream 端点。
		httpx.Fail(c, toolerr.Newf(toolerr.CodeToolNotFound, map[string]any{"slug": slug}))
		return
	}
	if slug != "ssl-check" {
		httpx.Fail(c, toolerr.Newf(toolerr.CodeToolNotFound, map[string]any{"slug": slug}))
		return
	}

	key := c.ClientIP()
	if !s.streams.acquire(key) {
		httpx.Fail(c, toolerr.New(toolerr.CodeRateLimited))
		return
	}
	defer s.streams.release(key)

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: false,
	})
	if err != nil {
		// Accept 失败时已经写过响应（例如来源校验不通过）。
		s.logger.Warn("流式握手失败", "slug", slug)
		return
	}
	defer func() { _ = conn.CloseNow() }()

	ctx, cancel := context.WithTimeout(c.Request.Context(), streamBudget)
	defer cancel()

	go s.pingLoop(ctx, conn, time.Now())

	firstCtx, firstCancel := context.WithTimeout(ctx, firstFrameTimeout)
	_, payload, readErr := conn.Read(firstCtx)
	firstCancel()
	if readErr != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, "missing first frame")
		return
	}

	var input streamInput
	if unmarshalErr := json.Unmarshal(payload, &input); unmarshalErr != nil {
		s.writeErrorFrame(ctx, conn, toolerr.New(toolerr.CodeInvalidJSON))
		_ = conn.Close(websocket.StatusPolicyViolation, "invalid first frame")
		return
	}

	opts := input.toOptions()
	if validErr := network.ValidateTLSProbe(&opts); validErr != nil {
		s.writeErrorFrame(ctx, conn, tools.Normalize(validErr))
		_ = conn.Close(websocket.StatusPolicyViolation, "invalid input")
		return
	}

	result, probeErr := network.TLSProbe(ctx, s.network, opts, func(progress network.TLSProgress) {
		s.writeFrame(ctx, conn, streamFrame{
			Type:  "progress",
			Phase: progress.Phase,
			Done:  progress.Done,
			Total: progress.Total,
		})
	})
	if probeErr != nil {
		s.writeErrorFrame(ctx, conn, tools.Normalize(probeErr))
		if errors.Is(probeErr, context.DeadlineExceeded) {
			_ = conn.Close(websocket.StatusInternalError, "timeout")
			return
		}
		_ = conn.Close(websocket.StatusPolicyViolation, "probe rejected")
		return
	}

	s.writeFrame(ctx, conn, streamFrame{Type: "result", Data: result})
	_ = conn.Close(websocket.StatusNormalClosure, "")
}

// streamInput 是 ssl-check 的客户端首帧。
type streamInput struct {
	// Host 是目标主机。
	Host string `json:"host"`
	// Port 是目标端口。
	Port int `json:"port"`
	// Sni 是 SNI。
	Sni string `json:"sni"`
	// Depth 是链深度上限。
	Depth int `json:"depth"`
	// Versions 是协议版本白名单。
	Versions []string `json:"versions"`
	// Ciphers 是密码套件白名单。
	Ciphers []string `json:"ciphers"`
	// TimeoutMs 是超时。
	TimeoutMs int `json:"timeoutMs"`
}

// toOptions 转成探测参数。
func (in *streamInput) toOptions() network.TLSProbeOptions {
	return network.TLSProbeOptions{
		Host:      in.Host,
		Port:      in.Port,
		SNI:       in.Sni,
		Depth:     in.Depth,
		Versions:  in.Versions,
		Ciphers:   in.Ciphers,
		TimeoutMs: in.TimeoutMs,
	}
}

// writeFrame 写一帧；失败即放弃（对端已断开）。
func (s *Server) writeFrame(ctx context.Context, conn *websocket.Conn, frame streamFrame) {
	payload, err := json.Marshal(frame)
	if err != nil {
		return
	}
	_ = conn.Write(ctx, websocket.MessageText, payload)
}

// writeErrorFrame 写一个错误帧。
func (s *Server) writeErrorFrame(ctx context.Context, conn *websocket.Conn, err *toolerr.Error) {
	if err == nil {
		err = toolerr.New(toolerr.CodeInternalError)
	}
	s.writeFrame(ctx, conn, streamFrame{
		Type: "error",
		Error: &streamError{
			Code:    string(err.Code),
			Message: err.Error(),
			Params:  err.Params,
		},
	})
}

// pingLoop 按固定间隔发送心跳，并在超过空闲上限时强制断开。
//
// 对端断开后 Ping 会返回错误，此时退出即可 —— 关闭动作由主流程负责。
// 空闲上限是**兜底**：正常路径在任务结束后立即关闭。
func (s *Server) pingLoop(ctx context.Context, conn *websocket.Conn, start time.Time) {
	ticker := time.NewTicker(streamPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if time.Since(start) > streamIdleTimeout {
				_ = conn.CloseNow()
				return
			}
			pingCtx, cancel := context.WithTimeout(ctx, firstFrameTimeout)
			pingErr := conn.Ping(pingCtx)
			cancel()
			if pingErr != nil {
				return
			}
		}
	}
}
