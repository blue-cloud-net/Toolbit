package stats

import (
	"context"
	"fmt"
	"time"
)

// RetentionDays 是明细数据的保留天数。
//
// 超过保留期的明细会被每日清理任务删除；日聚合表（本版尚未启用）永久保留。
const RetentionDays = 90

// Cleanup 删除保留期之前的明细。
//
// cutoff 是保留期分界日：该日之前的记录会被删除，该日及之后的保留。
// 返回被删除的记录数。
//
// 清理**不得**阻塞请求路径，由后台任务每日执行一次。
func (s *Store) Cleanup(ctx context.Context, cutoff time.Time) (int64, error) {
	day := cutoff.UTC().Format("2006-01-02")
	res, err := s.db.ExecContext(ctx, "DELETE FROM tool_calls WHERE day < ?", day)
	if err != nil {
		return 0, fmt.Errorf("清理过期统计失败: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, nil
	}
	return n, nil
}

// CleanupExpired 按保留期清理过期明细。
//
// now 是可注入的时钟，便于测试跨日边界。
func (s *Store) CleanupExpired(ctx context.Context, now time.Time) (int64, error) {
	return s.Cleanup(ctx, now.UTC().AddDate(0, 0, -RetentionDays))
}

// StartRetentionLoop 启动每日清理任务。
//
// 该协程随进程存活，收到 ctx 取消后退出；关闭前会执行一次清理，
// 保证进程重启后不会遗漏上一天的清理。
func (s *Store) StartRetentionLoop(ctx context.Context, logger interface {
	Error(msg string, args ...any)
}) {
	go func() {
		// 启动时先补一次清理。
		if _, err := s.CleanupExpired(ctx, s.now()); err != nil && ctx.Err() == nil {
			logger.Error("统计清理失败", "error", err)
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := s.CleanupExpired(ctx, s.now()); err != nil {
					logger.Error("统计清理失败", "error", err)
				}
			}
		}
	}()
}
