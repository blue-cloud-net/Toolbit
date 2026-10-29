package network

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDeviceHeadersGenerate 断言生成 Chrome 请求头。
func TestDeviceHeadersGenerate(t *testing.T) {
	result, err := DeviceHeaders(t.Context(), &DeviceHeadersRequest{
		Mode:     DeviceModeGenerate,
		Platform: DevicePlatformChrome,
		Version:  "120.0.0.0",
	})
	data := result.Data.(DeviceHeadersResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.Headers) == 0 {
		t.Fatal("应给出请求头")
	}
	if !strings.Contains(data.UserAgent, "Chrome/120.0.0.0") {
		t.Errorf("UA = %q", data.UserAgent)
	}
}

// TestDeviceHeadersParseUA 断言解析 User-Agent。
func TestDeviceHeadersParseUA(t *testing.T) {
	result, err := DeviceHeaders(t.Context(), &DeviceHeadersRequest{
		Mode: DeviceModeParseUA,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})
	data := result.Data.(DeviceHeadersResult)
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if data.Parsed == nil {
		t.Fatal("应给出解析结果")
	}
	if data.Parsed.Browser != "Chrome" {
		t.Errorf("浏览器 = %q", data.Parsed.Browser)
	}
}

// TestDeviceHeadersUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestDeviceHeadersUnsupportedMode(t *testing.T) {
	_, err := DeviceHeaders(t.Context(), &DeviceHeadersRequest{Mode: "rewrite"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
