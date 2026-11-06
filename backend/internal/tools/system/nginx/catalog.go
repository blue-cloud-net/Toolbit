package nginx

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"
)

// snapshotFS 内嵌指令目录快照。
//
// 快照由 `scripts/nginx-docs/gen-directives.mjs` 从官方文档抓取**事实字段**生成，
// 随二进制一起分发：工具**不**在运行时出网抓取文档。
//
//go:embed data/directives.json
var snapshotFS embed.FS

// Snapshot 是一次 Nginx 指令目录快照。
type Snapshot struct {
	// DocsVersion 是快照对应的官方文档版本。
	DocsVersion string `json:"docsVersion"`
	// StableVersion 是快照记录时的稳定版版本号。
	StableVersion string `json:"stableVersion"`
	// MainlineVersion 是快照记录时的主线版版本号。
	MainlineVersion string `json:"mainlineVersion"`
	// Source 是官方文档地址。
	Source string `json:"source"`
	// Modules 是全部模块。
	Modules []Module `json:"modules"`
	// Directives 是全部指令。
	Directives []Directive `json:"directives"`
}

// Module 是 Nginx 模块的元信息。
type Module struct {
	// Name 是模块标识，如 ngx_http_ssl_module。
	Name string `json:"name"`
	// Title 是模块标题。
	Title string `json:"title"`
	// DocURL 是官方文档链接。
	DocURL string `json:"docUrl"`
	// BuildOption 是启用该模块的编译开关；为空表示默认编译。
	BuildOption string `json:"buildOption"`
	// DefaultBuild 是否默认编译。
	DefaultBuild bool `json:"defaultBuild"`
	// Commercial 是否属于商业订阅提供的模块。
	Commercial bool `json:"commercial"`
}

// Directive 是一条指令的元信息。
type Directive struct {
	// ID 是指令的唯一标识：模块与指令名。
	ID string `json:"id"`
	// Name 是指令名。
	Name string `json:"name"`
	// Module 是所属模块。
	Module string `json:"module"`
	// Syntax 是语法提示。
	Syntax string `json:"syntax"`
	// Default 是默认值。
	Default string `json:"default"`
	// Contexts 是官方给出的可用上下文。
	Contexts []string `json:"contexts"`
	// Block 是否是块指令。
	Block bool `json:"block"`
	// Since 是首次出现的版本；为空表示官方未标注。
	Since string `json:"since"`
	// Deprecated 是废弃的版本。
	Deprecated string `json:"deprecated"`
	// Removed 是移除的版本。
	Removed string `json:"removed"`
	// CommercialParams 表示该指令含有仅商业订阅可用的参数。
	//
	// 它**不**表示指令本身不可用于开源版：官方文档里的商业版声明
	// 通常出现在某个参数的说明中（例如 resolver 的 status_zone）。
	CommercialParams bool `json:"commercialParams"`
	// DocURL 是官方文档链接。
	DocURL string `json:"docUrl"`
	// SummaryZh 是中文简述（人工维护的自有文案）。
	SummaryZh string `json:"summaryZh"`
	// SummaryEn 是英文简述（人工维护的自有文案）。
	SummaryEn string `json:"summaryEn"`
}

// 快照加载状态。
var (
	loadOnce  sync.Once
	loaded    *Snapshot
	loadEmpty = &Snapshot{}
)

// Load 读取内嵌的指令目录快照。
//
// 内嵌文件在编译期就已确定存在；解析失败按空目录处理，
// 让查询返回空结果而不是让整个进程崩溃。
func Load() *Snapshot {
	loadOnce.Do(func() {
		raw, err := snapshotFS.ReadFile("data/directives.json")
		if err != nil {
			loaded = loadEmpty
			return
		}
		snapshot := &Snapshot{}
		if err := json.Unmarshal(raw, snapshot); err != nil {
			loaded = loadEmpty
			return
		}
		loaded = snapshot
	})
	if loaded == nil {
		return loadEmpty
	}
	return loaded
}

// Versions 返回可选的目标版本（稳定版与主线版）。
func Versions() []string {
	snapshot := Load()
	out := make([]string, 0, 2)
	if snapshot.StableVersion != "" {
		out = append(out, snapshot.StableVersion)
	}
	if snapshot.MainlineVersion != "" && snapshot.MainlineVersion != snapshot.StableVersion {
		out = append(out, snapshot.MainlineVersion)
	}
	return out
}

// CompareVersion 比较两个版本号；a 小于 b 返回负数。
//
// 缺失或格式非法的版本按 0 处理，便于把「官方未标注」当作最早可用。
func CompareVersion(a, b string) int {
	pa := parseVersion(a)
	pb := parseVersion(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}
	return 0
}

// parseVersion 把版本号拆成三段数字。
func parseVersion(value string) [3]int {
	var out [3]int
	for i, part := range strings.SplitN(value, ".", 3) {
		if i > 2 {
			break
		}
		num := 0
		for _, r := range part {
			if r < '0' || r > '9' {
				num = 0
				break
			}
			num = num*10 + int(r-'0')
		}
		out[i] = num
	}
	return out
}

// AvailableIn 判断指令在目标版本中是否可用。
//
// 规则：官方未标注引入版本的指令视为一直可用；
// 已标注的按 since 与 removed 判定。
func AvailableIn(directive Directive, version string) bool {
	if version == "" {
		return true
	}
	if directive.Since != "" && CompareVersion(directive.Since, version) > 0 {
		return false
	}
	if directive.Removed != "" && CompareVersion(directive.Removed, version) <= 0 {
		return false
	}
	return true
}

// DeprecatedIn 判断指令在目标版本中是否已废弃。
func DeprecatedIn(directive Directive, version string) bool {
	if directive.Deprecated == "" || version == "" {
		return false
	}
	return CompareVersion(directive.Deprecated, version) <= 0
}

// FindDirective 按指令名查找目录条目。
//
// 同名指令可能分属不同模块，返回全部匹配项。
func FindDirective(snapshot *Snapshot, name string) []Directive {
	out := []Directive{}
	for _, directive := range snapshot.Directives {
		if directive.Name == name {
			out = append(out, directive)
		}
	}
	return out
}

// AllowsContext 判断指令是否可用于给定上下文。
//
// 官方把 http、mail、stream 三套 server 上下文写作同一个名字，
// 因此这里按「是否存在同名且允许该上下文的条目」判定。
//
// `include` 的官方上下文是 `any`：它在任何层级都可用，
// 因此这里放行全部上下文。
func AllowsContext(snapshot *Snapshot, name, context string) bool {
	if name == "include" {
		return true
	}
	for _, directive := range FindDirective(snapshot, name) {
		for _, ctx := range directive.Contexts {
			if NormalizeContext(ctx) == context {
				return true
			}
		}
	}
	return false
}

// NormalizeContext 把官方上下文写法归一化为工具使用的上下文名。
func NormalizeContext(context string) string {
	switch context {
	case "if in location":
		return "if"
	case "server", "location", "http", "upstream", "stream", "mail", "main", "limit_except":
		return context
	default:
		return context
	}
}
