package tools

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
)

// Registry 是 slug 到实现的映射。
type Registry struct {
	bySlug map[string]Definition
}

// NewRegistry 构造注册表。
func NewRegistry(defs []Definition) *Registry {
	r := &Registry{bySlug: make(map[string]Definition, len(defs))}
	for _, d := range defs {
		r.bySlug[d.Slug] = d
	}
	return r
}

// Lookup 按 slug 查找实现。
//
// 查询**大小写敏感**：slug 是契约的一部分。
func (r *Registry) Lookup(slug string) (Definition, bool) {
	d, ok := r.bySlug[slug]
	return d, ok
}

// Slugs 返回全部已注册的 slug，按字典序。
func (r *Registry) Slugs() []string {
	out := make([]string, 0, len(r.bySlug))
	for slug := range r.bySlug {
		out = append(out, slug)
	}
	sort.Strings(out)
	return out
}

// Len 返回已注册的实现数量。
func (r *Registry) Len() int {
	return len(r.bySlug)
}

// CrossCheck 交叉校验 catalog 与实现注册表。
//
// 返回两类不一致：
//   - Missing：catalog 中已启用但没有实现的 slug；
//   - Orphan：有实现但 catalog 中没有对应条目的 slug。
//
// 已禁用（enabled: false）的工具**不**算缺失。
func CrossCheck(c *catalog.Catalog, r *Registry) (missing []string, orphan []string) {
	implemented := make(map[string]bool, r.Len())
	for _, slug := range r.Slugs() {
		implemented[slug] = true
	}

	for _, t := range c.Tools {
		if !t.Enabled {
			continue
		}
		if !implemented[t.Slug] {
			missing = append(missing, t.Slug)
		}
	}

	for _, slug := range r.Slugs() {
		if _, ok := c.Tool(slug); !ok {
			orphan = append(orphan, slug)
		}
	}

	sort.Strings(missing)
	sort.Strings(orphan)
	return missing, orphan
}

// Report 返回交叉校验结果的文本摘要，用于启动日志。
func Report(c *catalog.Catalog, r *Registry) string {
	missing, orphan := CrossCheck(c, r)
	var b strings.Builder
	fmt.Fprintf(&b, "已实现 %d 个工具", r.Len())
	if len(missing) > 0 {
		fmt.Fprintf(&b, "；未实现 %d 个", len(missing))
	}
	if len(orphan) > 0 {
		fmt.Fprintf(&b, "；孤儿实现 %d 个", len(orphan))
	}
	return b.String()
}
