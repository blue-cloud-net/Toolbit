package system

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestFirewallRulesUFW 断言生成 ufw 规则。
func TestFirewallRulesUFW(t *testing.T) {
	result, err := FirewallRules(t.Context(), &FirewallRulesRequest{
		Tool:     FirewallToolUFW,
		Scenario: FirewallScenarioHTTPS,
		Protocol: "tcp",
	})
	data := result.Data.(FirewallRulesResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.Commands) == 0 {
		t.Fatal("应给出可执行命令")
	}
	if data.Tool != FirewallToolUFW {
		t.Errorf("工具 = %q", data.Tool)
	}
}

// TestFirewallRulesUnsupportedTool 断言未知工具返回 INVALID_FIELD。
func TestFirewallRulesUnsupportedTool(t *testing.T) {
	_, err := FirewallRules(t.Context(), &FirewallRulesRequest{
		Tool:     "pf",
		Scenario: FirewallScenarioHTTP,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
