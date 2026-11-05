package system

import (
	"context"
	"errors"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// exitCodes 执行一次退出码查询并取出结果。
func exitCodes(t *testing.T, req *ExitCodeRequest) ExitCodeResult {
	t.Helper()
	result, err := ExitCodes(context.Background(), req)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	out, ok := result.Data.(ExitCodeResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

func TestExitCodesByNumber(t *testing.T) {
	t.Parallel()

	out := exitCodes(t, &ExitCodeRequest{Query: "127"})
	if !out.Matched {
		t.Fatal("127 应当命中")
	}
	if out.Kind != ExitCodeQueryCode {
		t.Errorf("kind = %s，期望 %s", out.Kind, ExitCodeQueryCode)
	}
	if out.Code != 127 {
		t.Errorf("code = %d，期望 127", out.Code)
	}
	if out.Type != ExitCodeTypeFailure {
		t.Errorf("type = %s，期望 %s", out.Type, ExitCodeTypeFailure)
	}
	if out.Description == "" {
		t.Error("应给出中文说明")
	}
}

func TestExitCodesByKeyword(t *testing.T) {
	t.Parallel()

	out := exitCodes(t, &ExitCodeRequest{Query: "段错误"})
	if out.Kind != ExitCodeQueryKeyword {
		t.Errorf("kind = %s，期望 %s", out.Kind, ExitCodeQueryKeyword)
	}
	if !out.Matched || out.Code != 139 {
		t.Fatalf("「段错误」应当命中 139，实际 matched=%v code=%d", out.Matched, out.Code)
	}
}

func TestExitCodesAllEntries(t *testing.T) {
	t.Parallel()

	out := exitCodes(t, &ExitCodeRequest{})
	if out.Kind != ExitCodeQueryAll {
		t.Errorf("kind = %s，期望 %s", out.Kind, ExitCodeQueryAll)
	}
	if out.Total != out.DatasetTotal || out.Total == 0 {
		t.Fatalf("total = %d，datasetTotal = %d，两者应相等且非空", out.Total, out.DatasetTotal)
	}
	for _, entry := range out.Matches {
		if entry.Type == "" {
			t.Fatalf("退出码 %d 缺少类别", entry.Code)
		}
		if entry.Description == "" {
			t.Fatalf("退出码 %d 缺少中文说明", entry.Code)
		}
	}
}

func TestExitCodesTypeFilter(t *testing.T) {
	t.Parallel()

	out := exitCodes(t, &ExitCodeRequest{Type: ExitCodeTypeSignal})
	if out.Total == 0 {
		t.Fatal("信号终止类别不应为空")
	}
	for _, entry := range out.Matches {
		if entry.Type != ExitCodeTypeSignal {
			t.Fatalf("退出码 %d 的类别 = %s，期望 %s", entry.Code, entry.Type, ExitCodeTypeSignal)
		}
	}
}

func TestExitCodesUnknownCode(t *testing.T) {
	t.Parallel()

	out := exitCodes(t, &ExitCodeRequest{Query: "42"})
	if out.Matched {
		t.Error("未收录的退出码不应命中")
	}
	if out.Total != 0 {
		t.Errorf("total = %d，期望 0", out.Total)
	}
}

func TestExitCodesInvalidType(t *testing.T) {
	t.Parallel()

	_, err := ExitCodes(context.Background(), &ExitCodeRequest{Type: "bogus"})
	if err == nil {
		t.Fatal("非法类别应当报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) || toolErr.Code != toolerr.CodeInvalidField {
		t.Fatalf("错误码不符：%v", err)
	}
}
