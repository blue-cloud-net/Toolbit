package network

import (
	"context"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// uaParser 执行一次 UA 解析并取出结果。
func uaParser(t *testing.T, req *UAParserRequest) UAParserResult {
	t.Helper()
	result, err := UAParser(context.Background(), req)
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	out, ok := result.Data.(UAParserResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

func TestUAParserChromeOnWindows(t *testing.T) {
	t.Parallel()

	out := uaParser(t, &UAParserRequest{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})
	if out.Browser != "Chrome" {
		t.Errorf("浏览器 = %q，期望 Chrome", out.Browser)
	}
	if out.BrowserVersion != "120.0.0.0" {
		t.Errorf("浏览器版本 = %q，期望 120.0.0.0", out.BrowserVersion)
	}
	if out.OS != "Windows" {
		t.Errorf("操作系统 = %q，期望 Windows", out.OS)
	}
	if out.OSVersion != "10" {
		t.Errorf("系统版本 = %q，期望 10", out.OSVersion)
	}
	if out.DeviceType != "desktop" {
		t.Errorf("设备类型 = %q，期望 desktop", out.DeviceType)
	}
	if out.IsBot {
		t.Error("不应判为爬虫")
	}
	if out.Engine != "Blink/WebKit" {
		t.Errorf("渲染引擎 = %q，期望 Blink/WebKit", out.Engine)
	}
}

func TestUAParserMobileSafari(t *testing.T) {
	t.Parallel()

	out := uaParser(t, &UAParserRequest{
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) " +
			"AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1",
	})
	if out.Browser != "Safari" {
		t.Errorf("浏览器 = %q，期望 Safari", out.Browser)
	}
	if out.OS != "iOS" {
		t.Errorf("操作系统 = %q，期望 iOS", out.OS)
	}
	if out.DeviceType != "mobile" {
		t.Errorf("设备类型 = %q，期望 mobile", out.DeviceType)
	}
}

func TestUAParserDetectsBot(t *testing.T) {
	t.Parallel()

	out := uaParser(t, &UAParserRequest{
		UserAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	})
	if !out.IsBot {
		t.Error("应判为爬虫")
	}
	if out.DeviceType != "bot" {
		t.Errorf("设备类型 = %q，期望 bot", out.DeviceType)
	}
}

func TestUAParserUnknownFallsBack(t *testing.T) {
	t.Parallel()

	out := uaParser(t, &UAParserRequest{UserAgent: "Toolbit/1.0"})
	if out.Browser != "unknown" {
		t.Errorf("浏览器 = %q，期望 unknown", out.Browser)
	}
	if out.OS != "unknown" {
		t.Errorf("操作系统 = %q，期望 unknown", out.OS)
	}
	if out.DeviceType != "unknown" {
		t.Errorf("设备类型 = %q，期望 unknown", out.DeviceType)
	}
}

func TestUAParserMissingUserAgent(t *testing.T) {
	t.Parallel()

	_, err := UAParser(context.Background(), &UAParserRequest{UserAgent: "   "})
	assertCode(t, err, toolerr.CodeMissingField)
}

func TestUAParserTooLongUserAgent(t *testing.T) {
	t.Parallel()

	_, err := UAParser(context.Background(), &UAParserRequest{
		UserAgent: strings.Repeat("a", maxUserAgentLength+1),
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
