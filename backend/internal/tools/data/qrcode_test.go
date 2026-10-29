package data

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestQRCodeGenerate 断言生成二维码。
func TestQRCodeGenerate(t *testing.T) {
	result, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: QROpGenerate,
		Text:      "https://example.com",
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.HasPrefix(result.SVG, "<svg") {
		t.Errorf("SVG 格式不正确：%s", result.SVG[:min(40, len(result.SVG))])
	}
	if result.PNGBase64 == "" {
		t.Error("缺少 PNG")
	}
	raw, err := base64.StdEncoding.DecodeString(result.PNGBase64)
	if err != nil {
		t.Fatalf("PNG base64 解析失败：%v", err)
	}
	if len(raw) < 8 || string(raw[1:4]) != "PNG" {
		t.Error("PNG 魔数不正确")
	}
}

// TestQRCodeRoundTrip 断言生成后能识别回来。
func TestQRCodeRoundTrip(t *testing.T) {
	const content = "Toolbit 二维码往返测试"

	generated, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: QROpGenerate,
		Text:      content,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}

	decoded, err := QRCode(context.Background(), &QRCodeRequest{
		Operation:   QROpDecode,
		ImageBase64: generated.PNGBase64,
	})
	if err != nil {
		t.Fatalf("识别失败：%v", err)
	}
	if decoded.Content != content {
		t.Errorf("识别内容 = %q，期望 %q", decoded.Content, content)
	}
}

// TestQRCodeCustomColors 断言自定义颜色生效。
func TestQRCodeCustomColors(t *testing.T) {
	size := 8
	margin := 2
	result, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: QROpGenerate,
		Text:      "颜色",
		Options: QROptions{
			Size:    &size,
			Margin:  &margin,
			ECLevel: QRECLevelH,
			Fg:      "#123456",
			Bg:      "#abcdef",
		},
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(result.SVG, "#123456") {
		t.Error("SVG 缺少前景色")
	}
	if !strings.Contains(result.SVG, "#abcdef") {
		t.Error("SVG 缺少背景色")
	}
}

// TestQRCodeMissingText 断言缺少内容时报 MISSING_FIELD。
func TestQRCodeMissingText(t *testing.T) {
	_, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: QROpGenerate,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestQRCodeMissingImage 断言缺少图片时报 MISSING_FIELD。
func TestQRCodeMissingImage(t *testing.T) {
	_, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: QROpDecode,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestQRCodeInvalidOperation 断言未知操作报 INVALID_FIELD。
func TestQRCodeInvalidOperation(t *testing.T) {
	_, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: "render",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestQRCodeInvalidColor 断言非法颜色报 INVALID_FIELD。
func TestQRCodeInvalidColor(t *testing.T) {
	_, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: QROpGenerate,
		Text:      "x",
		Options:   QROptions{Fg: "red"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestQRCodeTextTooLong 断言超长内容报 TOO_LARGE。
func TestQRCodeTextTooLong(t *testing.T) {
	_, err := QRCode(context.Background(), &QRCodeRequest{
		Operation: QROpGenerate,
		Text:      strings.Repeat("a", maxQRTextBytes+1),
	})
	assertCode(t, err, toolerr.CodeTooLarge)
}

// TestQRCodeDecodeNotAnImage 断言非图片内容报 UNSUPPORTED_FORMAT。
func TestQRCodeDecodeNotAnImage(t *testing.T) {
	_, err := QRCode(context.Background(), &QRCodeRequest{
		Operation:   QROpDecode,
		ImageBase64: base64.StdEncoding.EncodeToString([]byte("not an image")),
	})
	assertCode(t, err, toolerr.CodeUnsupportedFormat)
}

// min 返回两数中的较小者。
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
