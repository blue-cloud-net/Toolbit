package catalog

import (
	"strings"
	"testing"
)

func TestLoadEmbeddedCatalog(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("加载内嵌目录失败: %v", err)
	}
	if c.Version != Version {
		t.Errorf("目录版本 = %d，期望 %d", c.Version, Version)
	}
	if got, want := len(c.Tools), 70; got != want {
		t.Errorf("工具数量 = %d，期望 %d", got, want)
	}
	if got, want := len(c.Categories), 10; got != want {
		t.Errorf("分类数量 = %d，期望 %d", got, want)
	}
}

func TestCatalogInvariants(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("加载内嵌目录失败: %v", err)
	}

	categorySlugs := make(map[string]bool, len(c.Categories))
	for _, cat := range c.Categories {
		categorySlugs[cat.Slug] = true
	}

	seen := make(map[string]bool, len(c.Tools))
	orders := make(map[string]bool, len(c.Tools))
	for _, tool := range c.Tools {
		if seen[tool.Slug] {
			t.Errorf("工具 slug 重复：%s", tool.Slug)
		}
		seen[tool.Slug] = true

		if !slugPattern.MatchString(tool.Slug) {
			t.Errorf("工具 slug 不符合 kebab-case：%s", tool.Slug)
		}
		if !categorySlugs[tool.Category] {
			t.Errorf("工具 %s 引用了不存在的分类：%s", tool.Slug, tool.Category)
		}
		if !validTransports[tool.Transport] {
			t.Errorf("工具 %s 的 transport 非法：%s", tool.Slug, tool.Transport)
		}
		if !validRateClasses[tool.RateClass] {
			t.Errorf("工具 %s 的 rateClass 非法：%s", tool.Slug, tool.RateClass)
		}
		if tool.MaxBodyBytes <= 0 {
			t.Errorf("工具 %s 的 maxBodyBytes 非法：%d", tool.Slug, tool.MaxBodyBytes)
		}
		if tool.TimeoutMs <= 0 {
			t.Errorf("工具 %s 的 timeoutMs 非法：%d", tool.Slug, tool.TimeoutMs)
		}
		if want := "tools." + tool.Slug; tool.I18nKey != want {
			t.Errorf("工具 %s 的 i18nKey = %s，期望 %s", tool.Slug, tool.I18nKey, want)
		}

		key := tool.Category + "#" + tool.Slug
		if orders[key] {
			t.Errorf("工具 %s 的分类内 order 重复", tool.Slug)
		}
		orders[key] = true
	}
}

func TestToolLookupIsCaseSensitive(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("加载内嵌目录失败: %v", err)
	}
	if _, ok := c.Tool("base64"); !ok {
		t.Error("应能按 slug 找到 base64")
	}
	if _, ok := c.Tool("Base64"); ok {
		t.Error("slug 查询必须大小写敏感")
	}
	if _, ok := c.Tool("BASE64"); ok {
		t.Error("slug 查询必须大小写敏感")
	}
}

func TestParseRejectsInvalidCatalog(t *testing.T) {
	cases := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "非 kebab-case 的 slug",
			yaml:    catalogYAML("Bad_Slug", "encoding", 1, "json", "light"),
			wantErr: "kebab-case",
		},
		{
			name:    "不存在的分类",
			yaml:    catalogYAML("demo", "nope", 1, "json", "light"),
			wantErr: "不存在的分类",
		},
		{
			name:    "非法 transport",
			yaml:    catalogYAML("demo", "encoding", 1, "graphql", "light"),
			wantErr: "transport",
		},
		{
			name:    "非法 rateClass",
			yaml:    catalogYAML("demo", "encoding", 1, "json", "extreme"),
			wantErr: "rateClass",
		},
		{
			name:    "i18nKey 与 slug 不匹配",
			yaml:    catalogYAML("demo", "encoding", 1, "json", "light", "tools.other"),
			wantErr: "i18nKey",
		},
		{
			name:    "maxBodyBytes 非正",
			yaml:    catalogYAML("demo", "encoding", 1, "json", "light", "tools.demo", "0"),
			wantErr: "maxBodyBytes",
		},
		{
			name:    "timeoutMs 非正",
			yaml:    catalogYAML("demo", "encoding", 1, "json", "light", "tools.demo", "1024", "0"),
			wantErr: "timeoutMs",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.yaml))
			if err == nil {
				t.Fatal("期望校验失败，但通过了")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("错误信息 = %q，期望包含 %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestParseRejectsDuplicateSlug(t *testing.T) {
	yaml := `
version: 1
categories:
  - slug: encoding
    name: 编码
    nameEn: Encoding
    order: 1
    icon: code
tools:
  - slug: demo
    category: encoding
    order: 1
    name: 示例
    nameEn: Demo
    enabled: true
    transport: json
    rateClass: light
    maxBodyBytes: 1024
    timeoutMs: 1000
    sensitive: false
    requiresNetwork: false
    autoRun: false
    wide: false
    i18nKey: tools.demo
  - slug: demo
    category: encoding
    order: 2
    name: 示例二
    nameEn: Demo Two
    enabled: true
    transport: json
    rateClass: light
    maxBodyBytes: 1024
    timeoutMs: 1000
    sensitive: false
    requiresNetwork: false
    autoRun: false
    wide: false
    i18nKey: tools.demo
`
	_, err := Parse([]byte(yaml))
	if err == nil || !strings.Contains(err.Error(), "重复") {
		t.Errorf("重复 slug 应当报错，实际：%v", err)
	}
}

func TestEnabledToolsAndGroups(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("加载内嵌目录失败: %v", err)
	}
	enabled := c.EnabledTools()
	if len(enabled) == 0 {
		t.Fatal("应至少有一个已启用的工具")
	}
	groups := c.ToolsByCategory()
	if len(groups) == 0 {
		t.Fatal("分类分组不应为空")
	}
	for slug, list := range groups {
		for i := 1; i < len(list); i++ {
			if list[i-1].Order > list[i].Order {
				t.Errorf("分类 %s 内工具未按 order 升序", slug)
				break
			}
		}
	}
}

func TestCategorySlugsSortedByOrder(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("加载内嵌目录失败: %v", err)
	}
	slugs := c.CategorySlugs()
	if len(slugs) != len(c.Categories) {
		t.Fatalf("分类 slug 数量 = %d，期望 %d", len(slugs), len(c.Categories))
	}
	for i := 1; i < len(slugs); i++ {
		if c.categoryOrder(slugs[i-1]) > c.categoryOrder(slugs[i]) {
			t.Errorf("分类未按 order 升序：%s 在 %s 之前", slugs[i-1], slugs[i])
		}
	}
}

// catalogYAML 构造一个最小可用的目录 YAML，参数可按需覆盖。
func catalogYAML(slug, category string, order int, transport, rateClass string, extra ...string) string {
	fields := []string{
		"  - slug: " + slug,
		"    category: " + category,
		"    order: " + itoa(order),
		"    name: 示例",
		"    nameEn: Demo",
		"    enabled: true",
		"    transport: " + transport,
		"    rateClass: " + rateClass,
	}
	// 依次填充可选覆盖项：i18nKey、maxBodyBytes、timeoutMs。
	defaults := []string{"tools." + slug, "1024", "1000"}
	for i := 0; i < len(defaults); i++ {
		val := defaults[i]
		if i < len(extra) {
			val = extra[i]
		}
		fields = append(fields, "    "+[]string{"i18nKey", "maxBodyBytes", "timeoutMs"}[i]+": "+val)
	}
	fields = append(fields,
		"    sensitive: false",
		"    requiresNetwork: false",
		"    autoRun: false",
		"    wide: false",
	)
	return "version: 1\ncategories:\n  - slug: encoding\n    name: 编码\n    nameEn: Encoding\n    order: 1\n    icon: code\ntools:\n" +
		strings.Join(fields, "\n") + "\n"
}

// itoa 把正整数转成字符串，避免为测试引入 strconv 之外的依赖。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
