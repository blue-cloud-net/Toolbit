package nginx

import (
	"strings"
	"testing"
)

// TestSnapshotLoads 断言内嵌快照可解析且规模合理。
func TestSnapshotLoads(t *testing.T) {
	snapshot := Load()
	if snapshot == nil {
		t.Fatalf("快照为空")
	}
	if len(snapshot.Directives) < 500 {
		t.Errorf("指令数 = %d，期望至少 500", len(snapshot.Directives))
	}
	if len(snapshot.Modules) < 30 {
		t.Errorf("模块数 = %d，期望至少 30", len(snapshot.Modules))
	}
	if snapshot.StableVersion == "" || snapshot.MainlineVersion == "" {
		t.Errorf("缺少版本信息：stable=%q mainline=%q", snapshot.StableVersion, snapshot.MainlineVersion)
	}
	if snapshot.Source == "" {
		t.Errorf("缺少官方文档地址")
	}
}

// TestSnapshotCoreDirectives 断言核心模块指令已收录。
//
// 全局索引只覆盖 http / stream / mail，核心模块需要单独补齐；
// 漏掉它会让 `worker_processes` 之类的指令被误判为未知。
func TestSnapshotCoreDirectives(t *testing.T) {
	snapshot := Load()
	for _, name := range []string{"worker_processes", "worker_connections", "events", "include", "user", "pid"} {
		if len(FindDirective(snapshot, name)) == 0 {
			t.Errorf("核心指令 %q 未收录", name)
		}
	}
}

// TestSnapshotCommonDirectives 断言常用 HTTP 指令已收录。
func TestSnapshotCommonDirectives(t *testing.T) {
	snapshot := Load()
	for _, name := range []string{
		"listen", "server_name", "root", "index", "try_files", "error_page", "return",
		"rewrite", "add_header", "expires", "charset", "gzip", "proxy_pass",
		"proxy_set_header", "ssl_certificate", "ssl_protocols", "fastcgi_pass", "upstream",
	} {
		if len(FindDirective(snapshot, name)) == 0 {
			t.Errorf("常用指令 %q 未收录", name)
		}
	}
}

// TestSnapshotDirectiveFields 断言每条指令的关键字段齐备。
func TestSnapshotDirectiveFields(t *testing.T) {
	snapshot := Load()
	for _, directive := range snapshot.Directives {
		if directive.Name == "" {
			t.Errorf("指令缺少名称：%v", directive.ID)
			continue
		}
		if directive.Module == "" {
			t.Errorf("指令 %q 缺少模块", directive.Name)
		}
		if directive.ID == "" {
			t.Errorf("指令 %q 缺少标识", directive.Name)
		}
		if len(directive.Contexts) == 0 {
			t.Errorf("指令 %q 缺少上下文", directive.Name)
		}
		if directive.DocURL == "" {
			t.Errorf("指令 %q 缺少文档链接", directive.Name)
		}
	}
}

// TestSnapshotUniqueIDs 断言指令标识唯一。
func TestSnapshotUniqueIDs(t *testing.T) {
	snapshot := Load()
	seen := map[string]bool{}
	for _, directive := range snapshot.Directives {
		if seen[directive.ID] {
			t.Errorf("指令标识重复：%s", directive.ID)
		}
		seen[directive.ID] = true
	}
}

// TestCompareVersion 断言版本号比较。
func TestCompareVersion(t *testing.T) {
	cases := []struct {
		a    string
		b    string
		want int
	}{
		{"1.30.6", "1.31.8", -1},
		{"1.31.8", "1.30.6", 1},
		{"1.30.6", "1.30.6", 0},
		{"1.9.0", "1.10.0", -1},
		{"", "1.0.0", -1},
	}
	for _, c := range cases {
		got := CompareVersion(c.a, c.b)
		if (got < 0) != (c.want < 0) || (got == 0) != (c.want == 0) {
			t.Errorf("CompareVersion(%q, %q) = %d，期望符号 %d", c.a, c.b, got, c.want)
		}
	}
}

// TestAvailableIn 断言按版本判定指令可用性。
func TestAvailableIn(t *testing.T) {
	// 官方文档：absolute_redirect 自 1.11.8 起提供。
	since := Directive{Name: "absolute_redirect", Since: "1.11.8"}
	if AvailableIn(since, "1.10.0") {
		t.Errorf("1.10.0 不应包含 1.11.8 引入的指令")
	}
	if !AvailableIn(since, "1.30.6") {
		t.Errorf("1.30.6 应包含 absolute_redirect")
	}

	// 官方文档：ssl 指令在 1.25.1 移除。
	removed := Directive{Name: "ssl", Removed: "1.25.1"}
	if AvailableIn(removed, "1.25.1") {
		t.Errorf("1.25.1 起 ssl 指令已移除")
	}
	if !AvailableIn(removed, "1.24.0") {
		t.Errorf("1.24.0 仍应有 ssl 指令")
	}

	// 未标注版本的指令视为一直可用。
	plain := Directive{Name: "server_name"}
	if !AvailableIn(plain, "1.0.0") {
		t.Errorf("未标注版本的指令应始终可用")
	}
}

// TestAllowsContextInclude 断言 include 在任意上下文都可用。
func TestAllowsContextInclude(t *testing.T) {
	snapshot := Load()
	for _, context := range []string{"main", "http", "server", "location", "upstream"} {
		if !AllowsContext(snapshot, "include", context) {
			t.Errorf("include 应适用于 %s 上下文", context)
		}
	}
}

// TestNormalizeContext 断言上下文归一化。
func TestNormalizeContext(t *testing.T) {
	if NormalizeContext("if in location") != "if" {
		t.Errorf("if in location 应归一化为 if")
	}
	if NormalizeContext("server") != "server" {
		t.Errorf("server 应保持不变")
	}
}

// TestVersions 断言可选版本列表。
func TestVersions(t *testing.T) {
	versions := Versions()
	if len(versions) < 1 {
		t.Fatalf("版本列表为空")
	}
	for _, version := range versions {
		if !strings.Contains(version, ".") {
			t.Errorf("版本号格式异常：%q", version)
		}
	}
}
