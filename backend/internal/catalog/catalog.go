// Package catalog 解析并校验工具目录 YAML。
//
// 本包**只**认识 YAML，不认识任何工具实现；反过来，工具实现也不认识本包。
// 目录的元数据（分类、顺序、传输方式、限流档位、限额）**只**在这里定义，
// 代码与文档都不得另立一份。
package catalog

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/blue-cloud-net/Toolbit/backend/catalog"
)

// Version 是目录文件的结构版本。
const Version = 1

// Transport 表示工具的传输方式。
type Transport string

// 传输方式取值。
const (
	// TransportJSON 走 JSON 请求体。
	TransportJSON Transport = "json"
	// TransportMultipart 走 multipart/form-data。
	TransportMultipart Transport = "multipart"
	// TransportWebSocket 走 WebSocket 流式端点。
	TransportWebSocket Transport = "websocket"
	// TransportClient 纯前端实现，**没有**后端端点。
	TransportClient Transport = "client"
)

// RateClass 表示限流档位。
type RateClass string

// 限流档位取值。
const (
	RateLight    RateClass = "light"
	RateStandard RateClass = "standard"
	RateHeavy    RateClass = "heavy"
	RateFile     RateClass = "file"
	RateNetwork  RateClass = "network"
)

// Category 是一个工具分类。
type Category struct {
	Slug   string `yaml:"slug" json:"slug"`
	Name   string `yaml:"name" json:"name"`
	NameEn string `yaml:"nameEn" json:"nameEn"`
	Order  int    `yaml:"order" json:"order"`
	Icon   string `yaml:"icon" json:"icon"`
}

// Tool 是一个工具的元数据。
//
// 字段与 backend/catalog/tools.yaml 一一对应。
type Tool struct {
	Slug            string    `yaml:"slug" json:"slug"`
	Category        string    `yaml:"category" json:"category"`
	Order           int       `yaml:"order" json:"order"`
	Name            string    `yaml:"name" json:"name"`
	NameEn          string    `yaml:"nameEn" json:"nameEn"`
	Enabled         bool      `yaml:"enabled" json:"enabled"`
	Transport       Transport `yaml:"transport" json:"transport"`
	RateClass       RateClass `yaml:"rateClass" json:"rateClass"`
	MaxBodyBytes    int64     `yaml:"maxBodyBytes" json:"maxBodyBytes"`
	TimeoutMs       int       `yaml:"timeoutMs" json:"timeoutMs"`
	Sensitive       bool      `yaml:"sensitive" json:"sensitive"`
	RequiresNetwork bool      `yaml:"requiresNetwork" json:"requiresNetwork"`
	AutoRun         bool      `yaml:"autoRun" json:"autoRun"`
	Wide            bool      `yaml:"wide" json:"wide"`
	I18nKey         string    `yaml:"i18nKey" json:"i18nKey"`
}

// Catalog 是解析后的工具目录。
type Catalog struct {
	Version    int        `yaml:"version" json:"version"`
	Categories []Category `yaml:"categories" json:"categories"`
	Tools      []Tool     `yaml:"tools" json:"tools"`
}

// slugPattern 是合法的 slug 形式：小写 kebab-case。
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// 合法取值表。
var (
	validTransports = map[Transport]bool{
		TransportJSON:      true,
		TransportMultipart: true,
		TransportWebSocket: true,
		TransportClient:    true,
	}
	validRateClasses = map[RateClass]bool{
		RateLight:    true,
		RateStandard: true,
		RateHeavy:    true,
		RateFile:     true,
		RateNetwork:  true,
	}
)

// Parse 解析并校验目录 YAML。
//
// 任何校验失败都返回带上下文的错误，错误信息里**只**包含字段名与取值，
// 不包含用户输入。
func Parse(data []byte) (*Catalog, error) {
	var c Catalog
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("解析目录 YAML 失败: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Load 解析内嵌的目录 YAML。
func Load() (*Catalog, error) {
	return Parse(catalog.ToolsYAML)
}

// validate 校验目录的结构一致性。
func (c *Catalog) validate() error {
	if c.Version != Version {
		return fmt.Errorf("目录版本 = %d，期望 %d", c.Version, Version)
	}
	if len(c.Categories) == 0 {
		return fmt.Errorf("目录缺少分类")
	}
	if len(c.Tools) == 0 {
		return fmt.Errorf("目录缺少工具")
	}

	categorySlugs := make(map[string]bool, len(c.Categories))
	categoryOrders := make(map[int]string, len(c.Categories))
	for _, cat := range c.Categories {
		if !slugPattern.MatchString(cat.Slug) {
			return fmt.Errorf("分类 slug 不符合 kebab-case：%s", cat.Slug)
		}
		if categorySlugs[cat.Slug] {
			return fmt.Errorf("分类 slug 重复：%s", cat.Slug)
		}
		if cat.Order <= 0 {
			return fmt.Errorf("分类 %s 的 order 非法：%d", cat.Slug, cat.Order)
		}
		if prev, dup := categoryOrders[cat.Order]; dup {
			return fmt.Errorf("分类 order %d 重复：%s 与 %s", cat.Order, prev, cat.Slug)
		}
		categorySlugs[cat.Slug] = true
		categoryOrders[cat.Order] = cat.Slug
	}

	toolSlugs := make(map[string]bool, len(c.Tools))
	toolOrders := make(map[string]int, len(c.Tools))
	for i := range c.Tools {
		t := &c.Tools[i]
		if !slugPattern.MatchString(t.Slug) {
			return fmt.Errorf("工具 slug 不符合 kebab-case：%s", t.Slug)
		}
		if toolSlugs[t.Slug] {
			return fmt.Errorf("工具 slug 重复：%s", t.Slug)
		}
		toolSlugs[t.Slug] = true

		if !categorySlugs[t.Category] {
			return fmt.Errorf("工具 %s 引用了不存在的分类：%s", t.Slug, t.Category)
		}
		if t.Order <= 0 {
			return fmt.Errorf("工具 %s 的 order 非法：%d", t.Slug, t.Order)
		}
		key := t.Category + "#" + t.Slug
		if prev, dup := toolOrders[key]; dup {
			return fmt.Errorf("工具 %s 的 order %d 与同分类内其他工具重复", t.Slug, prev)
		}
		toolOrders[key] = t.Order

		if !validTransports[t.Transport] {
			return fmt.Errorf("工具 %s 的 transport 非法：%s", t.Slug, t.Transport)
		}
		if !validRateClasses[t.RateClass] {
			return fmt.Errorf("工具 %s 的 rateClass 非法：%s", t.Slug, t.RateClass)
		}
		if t.MaxBodyBytes <= 0 {
			return fmt.Errorf("工具 %s 的 maxBodyBytes 非法：%d", t.Slug, t.MaxBodyBytes)
		}
		if t.TimeoutMs <= 0 {
			return fmt.Errorf("工具 %s 的 timeoutMs 非法：%d", t.Slug, t.TimeoutMs)
		}
		if want := "tools." + t.Slug; t.I18nKey != want {
			return fmt.Errorf("工具 %s 的 i18nKey = %s，期望 %s", t.Slug, t.I18nKey, want)
		}
	}

	return nil
}

// Tool 按 slug 查找工具。
//
// 查找**大小写敏感**：slug 是契约的一部分。
func (c *Catalog) Tool(slug string) (Tool, bool) {
	for _, t := range c.Tools {
		if t.Slug == slug {
			return t, true
		}
	}
	return Tool{}, false
}

// EnabledTools 返回全部已启用的工具。
func (c *Catalog) EnabledTools() []Tool {
	out := make([]Tool, 0, len(c.Tools))
	for _, t := range c.Tools {
		if t.Enabled {
			out = append(out, t)
		}
	}
	return out
}

// CategorySlugs 返回按 order 升序排列的分类 slug。
func (c *Catalog) CategorySlugs() []string {
	out := make([]string, 0, len(c.Categories))
	for _, cat := range c.Categories {
		out = append(out, cat.Slug)
	}
	sort.Slice(out, func(i, j int) bool {
		return c.categoryOrder(out[i]) < c.categoryOrder(out[j])
	})
	return out
}

// categoryOrder 返回分类的 order。
func (c *Catalog) categoryOrder(slug string) int {
	for _, cat := range c.Categories {
		if cat.Slug == slug {
			return cat.Order
		}
	}
	return 0
}

// ToolsByCategory 返回按分类分组的工具，每组内按 order 升序。
func (c *Catalog) ToolsByCategory() map[string][]Tool {
	groups := make(map[string][]Tool, len(c.Categories))
	for _, t := range c.Tools {
		groups[t.Category] = append(groups[t.Category], t)
	}
	for slug := range groups {
		list := groups[slug]
		sort.Slice(list, func(i, j int) bool { return list[i].Order < list[j].Order })
		groups[slug] = list
	}
	return groups
}

// String 返回目录的简短摘要，用于日志。
func (c *Catalog) String() string {
	return fmt.Sprintf("catalog v%d: %d 个工具, %d 个分类", c.Version, len(c.Tools), len(c.Categories))
}

// Slugs 返回全部工具 slug，按字典序。
func (c *Catalog) Slugs() []string {
	out := make([]string, 0, len(c.Tools))
	for _, t := range c.Tools {
		out = append(out, t.Slug)
	}
	sort.Strings(out)
	return out
}

// NormalizeSlug 归一化 slug：去除首尾空白并转小写。
//
// **仅**用于**日志与错误提示**，路由匹配一律使用原始 slug。
func NormalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
