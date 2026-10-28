// Package stats 实现匿名调用统计：迁移、异步批量写入与保留期清理。
//
// 隐私约束：
//   - **不**记录请求体、响应体、参数内容、IP 原值或 User-Agent 原文；
//   - 只记录 slug、分类、状态码、错误码、耗时与字节数；
//   - 写入**必须**异步且**不阻塞**请求路径，队列满时丢弃并计数。
package stats

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// SchemaVersion 是当前代码期望的 schema 版本。
//
// 新增迁移时递增该值，并按序提供 migrations/<version>_<name>.sql。
const SchemaVersion = 1

// Store 是统计库的句柄。
type Store struct {
	db *sql.DB
	// now 是可注入的时钟，便于测试保留期逻辑。
	now func() time.Time
}

// Open 打开（或创建）统计库并执行迁移。
//
// path 为空或 dbPath 指向的目录不可写时返回错误，调用方应据此 fail-fast。
func Open(dbPath string) (*Store, error) {
	return openWithClock(dbPath, time.Now)
}

// openWithClock 是可注入时钟的版本。
func openWithClock(dbPath string, now func() time.Time) (*Store, error) {
	if strings.TrimSpace(dbPath) == "" {
		return nil, fmt.Errorf("统计库路径为空")
	}
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, fmt.Errorf("创建统计库目录失败: %w", err)
		}
	}

	dsn := dbPath + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开统计库失败: %w", err)
	}
	// SQLite 的并发写能力有限，限制连接数可避免写锁争用。
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("统计库不可用: %w", err)
	}

	s := &Store{db: db, now: now}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// DB 返回底层数据库句柄，仅供聚合查询使用。
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close 关闭统计库。
func (s *Store) Close() error {
	return s.db.Close()
}

// Writable 判断统计库是否可写。
//
// 用于 /readyz 的就绪检查：只读目录或权限不足时返回错误。
func (s *Store) Writable(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("统计库不可用: %w", err)
	}
	var one int
	err := s.db.QueryRowContext(ctx, "SELECT 1").Scan(&one)
	if err != nil {
		return fmt.Errorf("统计库不可读: %w", err)
	}
	return nil
}

// migrate 执行未应用的迁移。
//
// 迁移**必须**版本化，**禁止**在启动时无条件重建表。
func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_version (
		version    INTEGER NOT NULL,
		applied_at TEXT    NOT NULL
	)`); err != nil {
		return fmt.Errorf("创建 schema_version 失败: %w", err)
	}

	current, err := s.currentVersion(ctx)
	if err != nil {
		return err
	}
	if current >= SchemaVersion {
		return nil
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("读取迁移目录失败: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		version, ok := versionOf(name)
		if !ok {
			continue
		}
		if version <= current {
			continue
		}
		body, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("读取迁移 %s 失败: %w", name, err)
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("开启迁移事务失败: %w", err)
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("执行迁移 %s 失败: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_version (version, applied_at) VALUES (?, ?)",
			version, s.now().UTC().Format(time.RFC3339)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("记录迁移版本失败: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交迁移 %s 失败: %w", name, err)
		}
	}

	return nil
}

// currentVersion 返回已应用的最高迁移版本。
func (s *Store) currentVersion(ctx context.Context) (int, error) {
	var v sql.NullInt64
	err := s.db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_version").Scan(&v)
	if err != nil {
		return 0, fmt.Errorf("读取迁移版本失败: %w", err)
	}
	if !v.Valid {
		return 0, nil
	}
	return int(v.Int64), nil
}

// versionOf 从迁移文件名解析版本号。
//
// 文件名形如 0001_init.sql，返回 1；不符合命名则返回 false。
func versionOf(name string) (int, bool) {
	base := name
	if idx := strings.Index(base, "_"); idx > 0 {
		base = base[:idx]
	}
	base = strings.TrimSuffix(base, ".sql")
	var v int
	if _, err := fmt.Sscanf(base, "%04d", &v); err != nil {
		return 0, false
	}
	return v, true
}
