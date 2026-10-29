package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestRegexValidateValid 断言合法表达式通过校验。
func TestRegexValidateValid(t *testing.T) {
	result, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: `^\d{3}-\d{4}$`,
	})
	data := result.Data.(RegexValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if !data.Valid {
		t.Error("合法表达式 valid 应为 true")
	}
	if data.Dialect != "RE2" {
		t.Errorf("方言 = %q，期望 RE2", data.Dialect)
	}
}

// TestRegexValidateInvalid 断言非法表达式返回错误位置。
func TestRegexValidateInvalid(t *testing.T) {
	result, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: `([a-z`,
	})
	data := result.Data.(RegexValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.Valid {
		t.Error("非法表达式 valid 应为 false")
	}
	if data.Error == nil {
		t.Fatal("非法表达式应给出错误信息")
	}
	// Go 的正则错误只在部分场景带 `^` 标记，
	// 有标记时必须换算出行列位置。
	if data.Error.Offset >= 0 && data.Error.Position == "" {
		t.Errorf("偏移 %d 应换算出位置", data.Error.Offset)
	}
}

// TestRegexValidateCaptureGroups 断言捕获组数量。
func TestRegexValidateCaptureGroups(t *testing.T) {
	result, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: `(\w+)@(\w+)`,
	})
	data := result.Data.(RegexValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.CaptureGroups != 2 {
		t.Errorf("捕获组数 = %d，期望 2", data.CaptureGroups)
	}
}

// TestRegexValidateMissingPattern 断言缺少模式返回 MISSING_FIELD。
func TestRegexValidateMissingPattern(t *testing.T) {
	_, err := RegexValidate(t.Context(), &RegexValidateRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestRegexValidateUnknownFlag 断言未知标志返回 INVALID_FIELD。
func TestRegexValidateUnknownFlag(t *testing.T) {
	_, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: "abc",
		Flags:   []string{"z"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
