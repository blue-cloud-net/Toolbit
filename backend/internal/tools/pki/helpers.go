package pki

import (
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// parseAtTime 解析 RFC3339 时刻。
//
// 为空时返回零值（表示「按当前时间」）。
// 格式非法时返回 INVALID_TIME。
func parseAtTime(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, nil
	}
	at, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
			"field":    "atTime",
			"expected": "RFC3339",
		})
	}
	return at, nil
}

// containsString 判断字符串切片是否包含目标值。
func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
