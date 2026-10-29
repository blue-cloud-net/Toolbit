package system

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestNginxConfigSite 断言站点配置含 server_name。
func TestNginxConfigSite(t *testing.T) {
	result, err := NginxConfig(t.Context(), &NginxConfigRequest{
		Mode:       NginxModeSite,
		ServerName: "example.com",
		ListenPort: 80,
		Root:       "/var/www/html",
	})
	data := result.Data.(NginxConfigResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Config, "example.com") {
		t.Errorf("配置应含域名：%s", data.Config)
	}
	if data.Mode != NginxModeSite {
		t.Errorf("模式 = %q", data.Mode)
	}
}

// TestNginxConfigUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestNginxConfigUnsupportedMode(t *testing.T) {
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{Mode: "stream"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
