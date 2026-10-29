package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDataMaskPhone 断言手机号脱敏。
func TestDataMaskPhone(t *testing.T) {
	result, err := DataMask(t.Context(), &DataMaskRequest{
		Text:      "手机号 13800138000",
		Types:     []string{MaskTypePhone},
		KeepCount: 3,
	})
	data := result.Data.(DataMaskResult)
	if err != nil {
		t.Fatalf("脱敏失败：%v", err)
	}
	if data.Total != 1 {
		t.Fatalf("命中数 = %d，期望 1", data.Total)
	}
	if data.Masked == "手机号 13800138000" {
		t.Error("脱敏结果不应保留原文")
	}
}

// TestDataMaskKeepTail 断言保留尾部。
func TestDataMaskKeepTail(t *testing.T) {
	result, err := DataMask(t.Context(), &DataMaskRequest{
		Text:      "110101199003078888",
		Types:     []string{MaskTypeIDCard},
		Position:  MaskKeepTail,
		KeepCount: 3,
	})
	data := result.Data.(DataMaskResult)
	if err != nil {
		t.Fatalf("脱敏失败：%v", err)
	}
	if data.Masked != "***************888" {
		t.Errorf("脱敏结果 = %q", data.Masked)
	}
}

// TestDataMaskMissingText 断言缺少文本返回 MISSING_FIELD。
func TestDataMaskMissingText(t *testing.T) {
	_, err := DataMask(t.Context(), &DataMaskRequest{Types: []string{MaskTypePhone}})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestDataMaskUnknownType 断言未知脱敏类型返回 INVALID_FIELD。
func TestDataMaskUnknownType(t *testing.T) {
	_, err := DataMask(t.Context(), &DataMaskRequest{
		Text:  "abc",
		Types: []string{"unknown"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
