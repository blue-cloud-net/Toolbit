package system

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 防火墙工具链。
const (
	// FirewallToolUFW 是 UFW。
	FirewallToolUFW = "ufw"
	// FirewallToolFirewalld 是 firewalld。
	FirewallToolFirewalld = "firewalld"
	// FirewallToolIPTables 是 iptables。
	FirewallToolIPTables = "iptables"
	// FirewallToolNFTables 是 nftables。
	FirewallToolNFTables = "nftables"
)

// FirewallTool 返回受支持的防火墙工具（顺序稳定）。
func FirewallTool() []string {
	return []string{
		FirewallToolUFW, FirewallToolFirewalld,
		FirewallToolIPTables, FirewallToolNFTables,
	}
}

// 防火墙场景。
const (
	// FirewallScenarioHTTP 是放行 HTTP 服务。
	FirewallScenarioHTTP = "http"
	// FirewallScenarioHTTPS 是放行 HTTPS 服务。
	FirewallScenarioHTTPS = "https"
	// FirewallScenarioSSH 是放行 SSH 管理端口。
	FirewallScenarioSSH = "ssh"
	// FirewallScenarioDatabase 是放行数据库端口。
	FirewallScenarioDatabase = "database"
	// FirewallScenarioDenyAll 是默认拒绝入站。
	FirewallScenarioDenyAll = "denyAll"
	// FirewallScenarioAllowAll 是允许全部入站。
	FirewallScenarioAllowAll = "allowAll"
)

// FirewallScenario 返回受支持的场景（顺序稳定）。
func FirewallScenario() []string {
	return []string{
		FirewallScenarioHTTP, FirewallScenarioHTTPS, FirewallScenarioSSH,
		FirewallScenarioDatabase, FirewallScenarioDenyAll, FirewallScenarioAllowAll,
	}
}

// FirewallRule 是一条规则。
type FirewallRule struct {
	// Action 是动作：allow / deny。
	Action string `json:"action"`
	// Port 是端口或端口范围。
	Port string `json:"port"`
	// Protocol 是协议：tcp / udp / any。
	Protocol string `json:"protocol"`
	// Source 是来源地址；空表示任意。
	Source string `json:"source"`
	// Description 是规则说明。
	Description string `json:"description"`
}

// FirewallRulesRequest 是 firewall-rules 工具的请求。
type FirewallRulesRequest struct {
	// Tool 是防火墙工具；默认 ufw。
	Tool string `json:"tool"`
	// Scenario 是场景；默认 http。
	Scenario string `json:"scenario"`
	// Ports 是自定义端口列表；提供时覆盖场景默认端口。
	Ports []string `json:"ports"`
	// Protocol 是协议：tcp / udp / any；默认 tcp。
	Protocol string `json:"protocol"`
	// Source 是来源地址；空表示任意。
	Source string `json:"source"`
}

// FirewallRulesResult 是 firewall-rules 工具的响应。
type FirewallRulesResult struct {
	// Tool 是实际使用的防火墙工具。
	Tool string `json:"tool"`
	// Scenario 是实际使用的场景。
	Scenario string `json:"scenario"`
	// Rules 是生成的规则列表。
	Rules []FirewallRule `json:"rules"`
	// Commands 是可直接执行的命令。
	Commands []string `json:"commands"`
	// Notes 是使用提示。
	Notes []string `json:"notes"`
}

// 场景默认端口。
var firewallScenarioPorts = map[string][]string{
	FirewallScenarioHTTP:     {"80"},
	FirewallScenarioHTTPS:    {"443"},
	FirewallScenarioSSH:      {"22"},
	FirewallScenarioDatabase: {"3306", "5432"},
	FirewallScenarioDenyAll:  {},
	FirewallScenarioAllowAll: {},
}

// 场景默认说明。
var firewallScenarioNotes = map[string][]string{
	FirewallScenarioHTTP:     {"放行 HTTP 端口 80"},
	FirewallScenarioHTTPS:    {"放行 HTTPS 端口 443"},
	FirewallScenarioSSH:      {"放行 SSH 端口 22；远程修改规则前请先确认当前会话已放行，避免把自己锁在门外"},
	FirewallScenarioDatabase: {"放行常见数据库端口 3306 与 5432；生产环境建议限制来源地址"},
	FirewallScenarioDenyAll:  {"默认拒绝全部入站流量；务必先放行 SSH 再启用，否则会失去远程访问能力"},
	FirewallScenarioAllowAll: {"允许全部入站流量；仅建议在受控网络中使用"},
}

// FirewallRules 生成防火墙规则与命令。
//
// 这是**纯模板拼装**：只生成文本，不触碰宿主机防火墙。
func FirewallRules(_ context.Context, req *FirewallRulesRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	toolName := req.Tool
	if toolName == "" {
		toolName = FirewallToolUFW
	}
	scenario := req.Scenario
	if scenario == "" {
		scenario = FirewallScenarioHTTP
	}
	protocol := req.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	ports := req.Ports
	if len(ports) == 0 {
		ports = firewallScenarioPorts[scenario]
	}

	rules := []FirewallRule{}
	for _, port := range ports {
		rules = append(rules, FirewallRule{
			Action:      "allow",
			Port:        port,
			Protocol:    protocol,
			Source:      req.Source,
			Description: "放行 " + port + "/" + protocol,
		})
	}
	if scenario == FirewallScenarioDenyAll {
		rules = append(rules, FirewallRule{
			Action:      "deny",
			Port:        "any",
			Protocol:    "any",
			Source:      req.Source,
			Description: "默认拒绝入站",
		})
	}
	if scenario == FirewallScenarioAllowAll {
		rules = append(rules, FirewallRule{
			Action:      "allow",
			Port:        "any",
			Protocol:    "any",
			Source:      req.Source,
			Description: "允许全部入站",
		})
	}

	commands := buildFirewallCommands(toolName, rules)
	notes := append([]string{}, firewallScenarioNotes[scenario]...)
	if req.Source != "" {
		notes = append(notes, "规则已限制来源为指定地址，其他来源不受影响")
	}

	return tools.Result{Data: FirewallRulesResult{
		Tool:     toolName,
		Scenario: scenario,
		Rules:    rules,
		Commands: commands,
		Notes:    notes,
	}}, nil
}

// buildFirewallCommands 按工具生成可执行命令。
func buildFirewallCommands(toolName string, rules []FirewallRule) []string {
	commands := []string{}

	switch toolName {
	case FirewallToolUFW:
		commands = append(commands, "ufw --force reset")
		for _, rule := range rules {
			if rule.Action == "deny" {
				commands = append(commands, "ufw default deny incoming")
				continue
			}
			commands = append(commands, buildUFWCommand(rule))
		}
		commands = append(commands, "ufw enable")
	case FirewallToolFirewalld:
		for _, rule := range rules {
			if rule.Action != "allow" || rule.Port == "any" {
				continue
			}
			if service := firewalldService(rule.Port); service != "" {
				commands = append(commands, "firewall-cmd --permanent --add-service="+service)
				continue
			}
			commands = append(commands,
				"firewall-cmd --permanent --add-port="+rule.Port+"/"+rule.Protocol)
		}
		commands = append(commands, "firewall-cmd --reload")
	case FirewallToolIPTables:
		for _, rule := range rules {
			if rule.Action == "deny" {
				commands = append(commands, "iptables -A INPUT -j DROP")
				continue
			}
			commands = append(commands, buildIPTablesCommand(rule))
		}
	case FirewallToolNFTables:
		commands = append(commands, "table inet filter {")
		commands = append(commands, "  chain input {")
		commands = append(commands, "    type filter hook input priority 0; policy accept;")
		for _, rule := range rules {
			if rule.Action == "deny" {
				commands = append(commands, "    # 默认拒绝策略：policy drop;")
				continue
			}
			commands = append(commands, buildNFTablesRule(rule))
		}
		commands = append(commands, "  }")
		commands = append(commands, "}")
	}

	return commands
}

// buildUFWCommand 生成 ufw 命令。
func buildUFWCommand(rule FirewallRule) string {
	parts := []string{"ufw allow"}
	if rule.Protocol != "any" && rule.Protocol != "" {
		parts = append(parts, rule.Protocol)
	}
	if rule.Port != "any" {
		parts = append(parts, rule.Port)
	}
	if rule.Source != "" {
		parts = append(parts, "from", rule.Source)
	}
	return strings.Join(parts, " ")
}

// buildIPTablesCommand 生成 iptables 命令。
func buildIPTablesCommand(rule FirewallRule) string {
	parts := []string{"iptables -A INPUT"}
	if rule.Protocol != "any" && rule.Protocol != "" {
		parts = append(parts, "-p", rule.Protocol)
	}
	if rule.Port != "any" {
		parts = append(parts, "--dport", rule.Port)
	}
	if rule.Source != "" {
		parts = append(parts, "-s", rule.Source)
	}
	parts = append(parts, "-j ACCEPT")
	return strings.Join(parts, " ")
}

// buildNFTablesRule 生成 nftables 规则行。
func buildNFTablesRule(rule FirewallRule) string {
	parts := []string{"    "}
	if rule.Protocol != "any" && rule.Protocol != "" {
		parts = append(parts, rule.Protocol+" ")
	}
	if rule.Port != "any" {
		parts = append(parts, "dport "+rule.Port+" ")
	}
	if rule.Source != "" {
		parts = append(parts, "ip saddr "+rule.Source+" ")
	}
	parts = append(parts, "accept")
	return strings.Join(parts, "")
}

// Validate 做字段级校验。
func (r *FirewallRulesRequest) Validate() error {
	toolName := r.Tool
	if toolName == "" {
		toolName = FirewallToolUFW
	}
	if !containsString(FirewallTool(), toolName) {
		return tools.InvalidField("tool", FirewallTool(), toolName)
	}

	scenario := r.Scenario
	if scenario == "" {
		scenario = FirewallScenarioHTTP
	}
	if !containsString(FirewallScenario(), scenario) {
		return tools.InvalidField("scenario", FirewallScenario(), scenario)
	}

	protocol := r.Protocol
	if protocol == "" {
		protocol = "tcp"
	}
	if !containsString([]string{"tcp", "udp", "any"}, protocol) {
		return tools.InvalidField("protocol", []string{"tcp", "udp", "any"}, protocol)
	}

	if len(r.Ports) > 32 {
		return tools.InvalidField("ports", []int{0, 32}, len(r.Ports))
	}
	for _, port := range r.Ports {
		if !isValidPortSpec(port) {
			return tools.InvalidField("ports", "port or port range", "invalid format")
		}
	}
	if len(r.Source) > 64 {
		return tools.InvalidField("source", []int{0, 64}, len(r.Source))
	}
	return nil
}

// isValidPortSpec 判断端口写法是否合法。
//
// 允许单个端口（80）与端口范围（8000-8100）。
func isValidPortSpec(spec string) bool {
	if spec == "" {
		return false
	}
	parts := strings.Split(spec, "-")
	if len(parts) > 2 {
		return false
	}
	for _, part := range parts {
		if !isAllDigits(part) {
			return false
		}
	}
	return true
}

// isAllDigits 判断是否全部是数字。
func isAllDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
