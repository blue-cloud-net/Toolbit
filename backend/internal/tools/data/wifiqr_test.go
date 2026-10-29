package data

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestWiFiQRPayload 断言 WPA 载荷格式。
func TestWiFiQRPayload(t *testing.T) {
	result, err := WiFiQR(t.Context(), &WiFiQRRequest{
		SSID:     "Guest-Network",
		Password: "guest-password",
		Security: WiFiSecurityWPA,
		Size:     128,
	})
	data := result.Data.(WiFiQRResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.HasPrefix(data.Payload, "WIFI:T:WPA;S:Guest-Network;P:guest-password;") {
		t.Errorf("载荷 = %q", data.Payload)
	}
	if data.PNG == "" {
		t.Error("应产出 PNG")
	}
}

// TestWiFiQROpenNetwork 断言开放网络不带口令。
func TestWiFiQROpenNetwork(t *testing.T) {
	result, err := WiFiQR(t.Context(), &WiFiQRRequest{
		SSID:     "Open",
		Security: WiFiSecurityNone,
		Size:     128,
	})
	data := result.Data.(WiFiQRResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Payload, "T:nopass;") {
		t.Errorf("载荷 = %q", data.Payload)
	}
}

// TestWiFiQRMissingSSID 断言缺少 SSID 返回 MISSING_FIELD。
func TestWiFiQRMissingSSID(t *testing.T) {
	_, err := WiFiQR(t.Context(), &WiFiQRRequest{
		Password: "x",
		Security: WiFiSecurityWPA,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}
