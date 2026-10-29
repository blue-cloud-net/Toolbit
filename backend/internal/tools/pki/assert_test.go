package pki

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// assertCode 断言错误的错误码。
func assertCode(t *testing.T, err error, want toolerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望错误 %s，实际无错误", want)
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("期望 *toolerr.Error，实际 %T", err)
	}
	if toolErr.Code != want {
		t.Errorf("错误码 = %s，期望 %s", toolErr.Code, want)
	}
}
