package data

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestUUIDGenV4 断言 v4 的格式与唯一性。
func TestUUIDGenV4(t *testing.T) {
	result, err := UUIDGen(context.Background(), &UUIDGenRequest{
		Version: UUIDv4,
		Count:   100,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(result.UUIDs) != 100 {
		t.Fatalf("数量 = %d，期望 100", len(result.UUIDs))
	}

	seen := map[string]bool{}
	for _, id := range result.UUIDs {
		if !isUUIDShape(id) {
			t.Fatalf("UUID 格式不正确：%q", id)
		}
		if seen[id] {
			t.Fatalf("UUID 重复：%q", id)
		}
		seen[id] = true
	}
}

// TestUUIDGenV7Monotonic 断言 v7 在同一毫秒内不重复且递增。
func TestUUIDGenV7Monotonic(t *testing.T) {
	result, err := UUIDGen(context.Background(), &UUIDGenRequest{
		Version: UUIDv7,
		Count:   500,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}

	seen := map[string]bool{}
	previous := ""
	for _, id := range result.UUIDs {
		if !isUUIDShape(id) {
			t.Fatalf("UUID 格式不正确：%q", id)
		}
		if seen[id] {
			t.Fatalf("UUID 重复：%q", id)
		}
		seen[id] = true
		if previous != "" && id <= previous {
			t.Fatalf("v7 未保持递增：%q 不大于 %q", id, previous)
		}
		previous = id
	}
}

// TestUUIDGenV1Monotonic 断言 v1 不重复。
func TestUUIDGenV1Monotonic(t *testing.T) {
	result, err := UUIDGen(context.Background(), &UUIDGenRequest{
		Version: UUIDv1,
		Count:   200,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}

	seen := map[string]bool{}
	for _, id := range result.UUIDs {
		if !isUUIDShape(id) {
			t.Fatalf("UUID 格式不正确：%q", id)
		}
		if seen[id] {
			t.Fatalf("UUID 重复：%q", id)
		}
		seen[id] = true
	}
}

// TestUUIDGenFormat 断言排版选项。
func TestUUIDGenFormat(t *testing.T) {
	noHyphen := false
	result, err := UUIDGen(context.Background(), &UUIDGenRequest{
		Count:     1,
		Hyphens:   &noHyphen,
		Braces:    true,
		Uppercase: true,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	got := result.UUIDs[0]
	// 32 位十六进制 + 两个花括号。
	if len(got) != 34 {
		t.Fatalf("长度 = %d，期望 34", len(got))
	}
	if got[0] != '{' || got[len(got)-1] != '}' {
		t.Errorf("缺少花括号：%q", got)
	}
	for _, c := range got {
		if c >= 'a' && c <= 'z' {
			t.Errorf("未转成大写：%q", got)
			break
		}
	}
}

// TestUUIDGenTooMany 断言数量超限报 INVALID_FIELD。
func TestUUIDGenTooMany(t *testing.T) {
	_, err := UUIDGen(context.Background(), &UUIDGenRequest{Count: 1001})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestUUIDGenInvalidVersion 断言未知版本报 INVALID_FIELD。
func TestUUIDGenInvalidVersion(t *testing.T) {
	_, err := UUIDGen(context.Background(), &UUIDGenRequest{Version: 5})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// isUUIDShape 判断文本是否符合标准 UUID 形状。
func isUUIDShape(id string) bool {
	if len(id) != 36 {
		return false
	}
	for _, i := range []int{8, 13, 18, 23} {
		if id[i] != '-' {
			return false
		}
	}
	for i, c := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		default:
			return false
		}
	}
	return true
}
