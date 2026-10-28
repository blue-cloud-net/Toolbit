package stats

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// newTestStore 创建一个临时目录上的统计库。
func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "stats.db"))
	if err != nil {
		t.Fatalf("打开统计库失败: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestOpenRunsMigration(t *testing.T) {
	store := newTestStore(t)

	var version int
	if err := store.DB().QueryRowContext(context.Background(), "SELECT MAX(version) FROM schema_version").Scan(&version); err != nil {
		t.Fatalf("读取迁移版本失败: %v", err)
	}
	if version != SchemaVersion {
		t.Errorf("迁移版本 = %d，期望 %d", version, SchemaVersion)
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stats.db")

	// 重复打开同一个库，迁移不应报错，也不应重复应用。
	for i := 0; i < 3; i++ {
		store, err := Open(path)
		if err != nil {
			t.Fatalf("第 %d 次打开失败: %v", i+1, err)
		}
		var count int
		if err := store.DB().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM schema_version").Scan(&count); err != nil {
			t.Fatalf("读取迁移记录数失败: %v", err)
		}
		if count != 1 {
			t.Errorf("第 %d 次打开后迁移记录数 = %d，期望 1", i+1, count)
		}
		_ = store.Close()
	}
}

func TestOpenFailsOnEmptyPath(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Error("空路径应当返回错误")
	}
}

func TestWritable(t *testing.T) {
	store := newTestStore(t)
	if err := store.Writable(context.Background()); err != nil {
		t.Errorf("统计库应可写: %v", err)
	}
}

func TestWritableFailsAfterClose(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "stats.db"))
	if err != nil {
		t.Fatalf("打开统计库失败: %v", err)
	}
	_ = store.Close()
	if err := store.Writable(context.Background()); err == nil {
		t.Error("关闭后不应报告可写")
	}
}

func TestWriterFlushesOnClose(t *testing.T) {
	store := newTestStore(t)
	writer := NewWriter(store, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	day := time.Now().UTC().Format("2006-01-02")
	for i := 0; i < 10; i++ {
		writer.Write(Event{
			Slug:       "base64",
			Category:   "encoding",
			Status:     200,
			OK:         true,
			DurationMs: 2,
			ReqBytes:   10,
			RespBytes:  20,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := writer.Close(ctx); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	n, err := store.CountCalls(context.Background(), day)
	if err != nil {
		t.Fatalf("统计调用数失败: %v", err)
	}
	if n != 10 {
		t.Errorf("落盘条数 = %d，期望 10", n)
	}
	if got := writer.Written(); got != 10 {
		t.Errorf("Written() = %d，期望 10", got)
	}
}

func TestWriterDropsWhenQueueFull(t *testing.T) {
	store := newTestStore(t)
	// 队列容量固定，这里塞入远超容量的事件，验证丢弃计数与「不阻塞」。
	writer := NewWriter(store, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	defer func() { _ = writer.Close(context.Background()) }()

	for i := 0; i < QueueSize*2; i++ {
		writer.Write(Event{Slug: "base64", Category: "encoding", Status: 200, OK: true})
	}
	if writer.Dropped() == 0 {
		t.Error("队列满时应当丢弃事件并计数")
	}
}

func TestWriteAfterCloseIsNoop(t *testing.T) {
	store := newTestStore(t)
	writer := NewWriter(store, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err := writer.Close(context.Background()); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}
	// 关闭后再投递不得 panic。
	writer.Write(Event{Slug: "base64", Category: "encoding", Status: 200, OK: true})
}

func TestCleanupRemovesOldRecords(t *testing.T) {
	store := newTestStore(t)

	old := time.Now().UTC().AddDate(0, 0, -RetentionDays-1)
	recent := time.Now().UTC().AddDate(0, 0, -1)

	events := []Event{
		{Slug: "old", Category: "encoding", Status: 200, OK: true, At: old},
		{Slug: "recent", Category: "encoding", Status: 200, OK: true, At: recent},
	}
	if err := store.insertEvents(context.Background(), events); err != nil {
		t.Fatalf("写入事件失败: %v", err)
	}

	removed, err := store.CleanupExpired(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("清理失败: %v", err)
	}
	if removed != 1 {
		t.Errorf("删除条数 = %d，期望 1", removed)
	}

	oldCount, err := store.CountCalls(context.Background(), old.UTC().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("统计旧记录失败: %v", err)
	}
	if oldCount != 0 {
		t.Errorf("过期记录数 = %d，期望 0", oldCount)
	}

	recentCount, err := store.CountCalls(context.Background(), recent.UTC().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("统计新记录失败: %v", err)
	}
	if recentCount != 1 {
		t.Errorf("保留记录数 = %d，期望 1", recentCount)
	}
}

func TestCleanupKeepsBoundaryDay(t *testing.T) {
	store := newTestStore(t)

	boundary := time.Now().UTC().AddDate(0, 0, -RetentionDays)
	event := Event{Slug: "boundary", Category: "encoding", Status: 200, OK: true, At: boundary}
	if err := store.insertEvents(context.Background(), []Event{event}); err != nil {
		t.Fatalf("写入事件失败: %v", err)
	}

	removed, err := store.CleanupExpired(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("清理失败: %v", err)
	}
	if removed != 0 {
		t.Errorf("分界日当天的记录不应被删除，删除了 %d 条", removed)
	}
}

func TestConcurrentWrites(t *testing.T) {
	store := newTestStore(t)
	writer := NewWriter(store, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	var wg sync.WaitGroup
	const workers = 10
	const perWorker = 50
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				writer.Write(Event{Slug: "base64", Category: "encoding", Status: 200, OK: true})
			}
		}()
	}
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := writer.Close(ctx); err != nil {
		t.Fatalf("关闭写入器失败: %v", err)
	}

	n, err := store.CountCalls(context.Background(), time.Now().UTC().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("统计调用数失败: %v", err)
	}
	if n != workers*perWorker {
		t.Errorf("落盘条数 = %d，期望 %d", n, workers*perWorker)
	}
}

func TestVersionOf(t *testing.T) {
	cases := map[string]struct {
		version int
		ok      bool
	}{
		"0001_init.sql":              {1, true},
		"0002_visitor_and_daily.sql": {2, true},
		"invalid.sql":                {0, false},
	}
	for name, want := range cases {
		got, ok := versionOf(name)
		if ok != want.ok {
			t.Errorf("versionOf(%q) ok = %t，期望 %t", name, ok, want.ok)
			continue
		}
		if ok && got != want.version {
			t.Errorf("versionOf(%q) = %d，期望 %d", name, got, want.version)
		}
	}
}
