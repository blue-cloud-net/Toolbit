package formatting

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestColorConvertHex 断言十六进制转换。
func TestColorConvertHex(t *testing.T) {
	result, err := ColorConvert(t.Context(), &ColorConvertRequest{
		Input:         "#3B82F6",
		InputFormat:   ColorFormatHex,
		OutputFormats: []string{ColorFormatHex, ColorFormatRGB, ColorFormatHSL},
	})
	data := result.Data.(ColorConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.Hex != "#3B82F6" {
		t.Errorf("HEX = %q", data.Hex)
	}
	if data.RGB == "" {
		t.Error("应给出 RGB 表示")
	}
	if data.ContrastOnWhite <= 1 {
		t.Errorf("对比度 = %v", data.ContrastOnWhite)
	}
}

// TestColorConvertInvalid 断言非法颜色返回 INVALID_FIELD。
func TestColorConvertInvalid(t *testing.T) {
	_, err := ColorConvert(t.Context(), &ColorConvertRequest{
		Input:       "not-a-color",
		InputFormat: ColorFormatHex,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
