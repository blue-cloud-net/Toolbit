package stats

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// 批量写入的触发条件。
const (
	// BatchSize 是触发批量落盘的事件条数。
	BatchSize = 100
	// FlushInterval 是触发批量落盘的时间间隔。
	FlushInterval = 2 * time.Second
	// QueueSize 是内存队列容量；队列满时**丢弃**事件，**不阻塞**请求。
	QueueSize = 4096
)

// Event 是一次工具调用的统计事件。
//
// 只包含**非敏感**的元信息；**禁止**放入请求体、参数内容或 IP 原值。
type Event struct {
	// Slug 是工具标识。
	Slug string
	// Category 是工具分类。
	Category string
	// Status 是 HTTP 状态码。
	Status int
	// OK 表示调用是否成功。
	OK bool
	// ErrorCode 是机器可读错误码；成功时为空。
	ErrorCode string
	// DurationMs 是处理耗时（毫秒）。
	DurationMs int64
	// ReqBytes 是请求体字节数。
	ReqBytes int64
	// RespBytes 是响应体字节数。
	RespBytes int64
	// At 是事件发生时间；零值时由写入器填充。
	At time.Time
}

// Writer 负责把事件异步批量写入统计库。
type Writer struct {
	store *Store
	queue chan Event
	now   func() time.Time

	dropped  atomic.Int64
	written  atomic.Int64
	closed   atomic.Bool
	closeOne sync.Once
	done     chan struct{}
	// nowFn 允许测试注入固定时钟。
	logger *slog.Logger
}

// NewWriter 创建写入器并启动后台刷新协程。
func NewWriter(store *Store, logger *slog.Logger) *Writer {
	if logger == nil {
		logger = slog.Default()
	}
	w := &Writer{
		store:  store,
		queue:  make(chan Event, QueueSize),
		now:    time.Now,
		done:   make(chan struct{}),
		logger: logger,
	}
	go w.loop()
	return w
}

// Write 投递一个事件。
//
// 队列已满时**丢弃**该事件并累加丢弃计数，**绝不**阻塞调用方。
func (w *Writer) Write(e Event) {
	if w.closed.Load() {
		return
	}
	if e.At.IsZero() {
		e.At = w.now()
	}
	select {
	case w.queue <- e:
	default:
		w.dropped.Add(1)
	}
}

// Dropped 返回被丢弃的事件数。
func (w *Writer) Dropped() int64 {
	return w.dropped.Load()
}

// Written 返回已落盘的事件数。
func (w *Writer) Written() int64 {
	return w.written.Load()
}

// Close 停止写入器并落盘剩余事件。
func (w *Writer) Close(ctx context.Context) error {
	var err error
	w.closeOne.Do(func() {
		w.closed.Store(true)
		close(w.queue)
		select {
		case <-w.done:
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return err
}

// loop 是后台刷新协程。
func (w *Writer) loop() {
	defer close(w.done)

	ticker := time.NewTicker(FlushInterval)
	defer ticker.Stop()

	batch := make([]Event, 0, BatchSize)
	for {
		select {
		case e, ok := <-w.queue:
			if !ok {
				// 队列已关闭：落盘剩余事件后退出。
				w.flush(context.Background(), batch)
				return
			}
			batch = append(batch, e)
			if len(batch) >= BatchSize {
				w.flush(context.Background(), batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				w.flush(context.Background(), batch)
				batch = batch[:0]
			}
		}
	}
}

// flush 把一批事件写入数据库。
func (w *Writer) flush(ctx context.Context, batch []Event) {
	if len(batch) == 0 {
		return
	}
	if err := w.store.insertEvents(ctx, batch); err != nil {
		w.logger.Error("统计写入失败", "error", err, "count", len(batch))
		return
	}
	w.written.Add(int64(len(batch)))
}

// insertEvents 在一个事务中写入一批事件。
func (s *Store) insertEvents(ctx context.Context, events []Event) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO tool_calls
		(ts, day, slug, category, status, ok, error_code, duration_ms, req_bytes, resp_bytes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for _, e := range events {
		at := e.At
		if at.IsZero() {
			at = s.now()
		}
		ok := 0
		if e.OK {
			ok = 1
		}
		if _, err := stmt.ExecContext(ctx,
			at.UnixMilli(),
			at.UTC().Format("2006-01-02"),
			e.Slug,
			e.Category,
			e.Status,
			ok,
			e.ErrorCode,
			e.DurationMs,
			e.ReqBytes,
			e.RespBytes,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// CountCalls 返回指定日期的调用总数，供测试与聚合使用。
func (s *Store) CountCalls(ctx context.Context, day string) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_calls WHERE day = ?", day).Scan(&n)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return n, nil
}

// EventCount 返回已记录的事件总数。
func (s *Store) EventCount(ctx context.Context) int64 {
	var n int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_calls").Scan(&n); err != nil {
		return 0
	}
	return n
}
