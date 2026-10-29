// Package documents 实现文档类工具。
//
// 本包只做「解码输入 → 校验字段 → 计算 → 编码 DTO」。
package documents

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Markdown 方言。
const (
	// FlavorGFM 是 GitHub Flavored Markdown。
	FlavorGFM = "gfm"
	// FlavorCommonMark 是 CommonMark。
	FlavorCommonMark = "commonmark"
)

// MarkdownFlavors 返回受支持的方言（顺序稳定）。
func MarkdownFlavors() []string {
	return []string{FlavorGFM, FlavorCommonMark}
}

// maxMarkdownBytes 是 Markdown 源文的字节上限。
const maxMarkdownBytes = 1 << 20

// maxTocEntries 是目录条目上限。
const maxTocEntries = 500

// MarkdownPreviewRequest 是 markdown-preview 工具的请求。
type MarkdownPreviewRequest struct {
	// Markdown 是 Markdown 源文。
	Markdown string `json:"markdown"`
	// Flavor 是方言：gfm / commonmark；缺省 gfm。
	Flavor string `json:"flavor"`
	// AllowHTML 是否允许内嵌 HTML；缺省 false。
	AllowHTML bool `json:"allowHtml"`
	// Breaks 是否把单换行视为换行；缺省 false。
	Breaks bool `json:"breaks"`
}

// TocEntry 是一个目录条目。
type TocEntry struct {
	// Level 是标题层级（1 到 6）。
	Level int `json:"level"`
	// Text 是标题文本。
	Text string `json:"text"`
	// ID 是标题锚点。
	ID string `json:"id"`
}

// MarkdownPreviewResult 是 markdown-preview 工具的响应。
type MarkdownPreviewResult struct {
	// HTML 是净化后的 HTML。
	HTML string `json:"html"`
	// TOC 是目录。
	TOC []TocEntry `json:"toc"`
}

// MarkdownPreview 把 Markdown 渲染为 HTML 并生成目录。
//
// 输出 HTML **一定**经过净化：禁止 `script`、事件属性与
// `javascript:` URL。原始 HTML 默认被转义。
func MarkdownPreview(ctx context.Context, req *MarkdownPreviewRequest) (MarkdownPreviewResult, error) {
	if err := req.Validate(); err != nil {
		return MarkdownPreviewResult{}, err
	}

	source := []byte(req.Markdown)
	md, parseCtx := buildMarkdown(req)

	// **只解析一次**：标题锚点由解析器生成，
	// 渲染与目录必须基于同一棵树，否则锚点会对不上。
	doc := md.Parser().Parse(text.NewReader(source), parser.WithContext(parseCtx))

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, source, doc); err != nil {
		return MarkdownPreviewResult{}, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "markdown",
			"reason": "cannot render markdown",
		})
	}

	return MarkdownPreviewResult{
		HTML: sanitizer().Sanitize(buf.String()),
		TOC:  collectTOC(doc, source),
	}, nil
}

// Validate 做字段级校验。
func (r *MarkdownPreviewRequest) Validate() error {
	switch r.Flavor {
	case "", FlavorGFM, FlavorCommonMark:
	default:
		return tools.InvalidField("flavor", MarkdownFlavors(), r.Flavor)
	}
	if strings.TrimSpace(r.Markdown) == "" {
		return tools.MissingField("markdown")
	}
	if len(r.Markdown) > maxMarkdownBytes {
		return toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"field":    "markdown",
			"maxBytes": maxMarkdownBytes,
		})
	}
	return nil
}

// buildMarkdown 按请求构造渲染器与锚点生成器。
//
// `allowHtml: false` 时**不**开启 `html.WithUnsafe()`：
// goldmark 会把原始 HTML 换成注释，等价于转义。
func buildMarkdown(req *MarkdownPreviewRequest) (goldmark.Markdown, parser.Context) {
	ids := newHeadingIDs()

	parserOptions := []parser.Option{
		parser.WithAutoHeadingID(),
		parser.WithHeadingAttribute(),
	}
	rendererOptions := []renderer.Option{}

	if req.Breaks {
		rendererOptions = append(rendererOptions, html.WithHardWraps())
	}
	if req.AllowHTML {
		rendererOptions = append(rendererOptions, html.WithUnsafe())
	}

	options := []goldmark.Option{
		goldmark.WithParserOptions(parserOptions...),
	}
	if len(rendererOptions) > 0 {
		options = append(options, goldmark.WithRendererOptions(rendererOptions...))
	}
	if req.Flavor != FlavorCommonMark {
		options = append(options, goldmark.WithExtensions(extension.GFM))
	}

	return goldmark.New(options...), parser.NewContext(parser.WithIDs(ids))
}

// headingIDs 生成 GitHub 风格的标题锚点。
//
// goldmark 的默认实现**丢弃**非 ASCII 字符，
// 于是中文标题会全部变成 `heading` / `heading-1`。
// 本实现保留 Unicode 字母与数字，更贴近 GitHub 的实际表现。
type headingIDs struct {
	// mu 保护 values。
	mu sync.Mutex
	// values 是已使用的锚点集合。
	values map[string]bool
}

// newHeadingIDs 创建一个锚点生成器。
//
// 每个请求**必须**新建一个：去重表不能跨请求共享。
func newHeadingIDs() *headingIDs {
	return &headingIDs{values: map[string]bool{}}
}

// Generate 生成标题锚点。
func (h *headingIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	slug := slugifyHeading(value)
	if slug == "" {
		if kind == ast.KindHeading {
			slug = "heading"
		} else {
			slug = "id"
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.values[slug] {
		h.values[slug] = true
		return []byte(slug)
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		if !h.values[candidate] {
			h.values[candidate] = true
			return []byte(candidate)
		}
	}
}

// Put 登记一个已显式指定的锚点。
func (h *headingIDs) Put(value []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.values[string(value)] = true
}

// slugifyHeading 把标题文本转成锚点。
//
// 保留字母（含中文）与数字；空格与连字符折叠为 `-`。
func slugifyHeading(value []byte) string {
	var b strings.Builder
	for _, r := range string(value) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// sanitizer 返回 HTML 净化策略。
//
// 基于 UGC 策略：允许常见排版标签，剥离脚本、样式、事件属性
// 与危险 URL 协议。
func sanitizer() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("id").OnElements("h1", "h2", "h3", "h4", "h5", "h6")
	policy.AllowAttrs("class").OnElements("code", "pre", "div", "span", "table", "ul", "ol", "li")
	policy.AllowElements("input")
	policy.AllowAttrs("type", "checked", "disabled").OnElements("input")
	policy.AllowAttrs("checked", "disabled").OnElements("li")
	return policy
}

// collectTOC 收集标题目录。
//
// 锚点 ID 与渲染时一致：两者读的是同一棵解析树上的
// `id` 属性（自动标题 ID 在**解析期**生成）。
func collectTOC(doc ast.Node, source []byte) []TocEntry {
	entries := []TocEntry{}
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := node.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		if len(entries) >= maxTocEntries {
			return ast.WalkStop, nil
		}
		entries = append(entries, TocEntry{
			Level: heading.Level,
			Text:  headingText(heading, source),
			ID:    headingID(heading),
		})
		return ast.WalkContinue, nil
	})

	return entries
}

// headingText 提取标题的纯文本。
func headingText(heading *ast.Heading, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(heading, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := node.(type) {
		case *ast.Text:
			b.Write(n.Value(source))
		case *ast.String:
			b.Write(n.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// headingID 提取标题锚点。
func headingID(heading *ast.Heading) string {
	for _, attr := range heading.Attributes() {
		if string(attr.Name) != "id" {
			continue
		}
		switch value := attr.Value.(type) {
		case []byte:
			return string(value)
		case string:
			return value
		}
	}
	return ""
}
