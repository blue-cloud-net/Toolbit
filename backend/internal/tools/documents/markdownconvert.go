package documents

import (
	"bytes"
	"context"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Markdown 转换方向。
const (
	// MarkdownToHTML 是 Markdown 转 HTML。
	MarkdownToHTML = "toHtml"
	// HTMLToMarkdown 是 HTML 转 Markdown。
	HTMLToMarkdown = "toMarkdown"
)

// MarkdownConvertMode 返回受支持的转换方向（顺序稳定）。
func MarkdownConvertMode() []string {
	return []string{MarkdownToHTML, HTMLToMarkdown}
}

// MarkdownConvertRequest 是 markdown-convert 工具的请求。
type MarkdownConvertRequest struct {
	// Mode 是转换方向：toHtml / toMarkdown；默认 toHtml。
	Mode string `json:"mode"`
	// Input 是待转换的内容。
	Input string `json:"input"`
	// GFM 是否启用 GFM 扩展（表格、删除线等）。
	GFM bool `json:"gfm"`
}

// MarkdownConvertResult 是 markdown-convert 工具的响应。
type MarkdownConvertResult struct {
	// Mode 是实际使用的转换方向。
	Mode string `json:"mode"`
	// Output 是转换结果。
	Output string `json:"output"`
	// HeadingCount 是标题数量。
	HeadingCount int `json:"headingCount"`
	// LinkCount 是链接数量。
	LinkCount int `json:"linkCount"`
	// CodeBlockCount 是代码块数量。
	CodeBlockCount int `json:"codeBlockCount"`
	// WordCount 是正文字数。
	WordCount int `json:"wordCount"`
}

// maxMarkdownConvertBytes 是输入内容的字节上限。
const maxMarkdownConvertBytes = 1 << 20

// MarkdownConvert 在 Markdown 与 HTML 之间转换。
//
// Markdown 解析复用既有的 goldmark 依赖，HTML 解析复用 golang.org/x/net/html。
func MarkdownConvert(_ context.Context, req *MarkdownConvertRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = MarkdownToHTML
	}

	if mode == HTMLToMarkdown {
		return htmlToMarkdown(req)
	}
	return markdownToHTML(req)
}

// markdownToHTML 把 Markdown 转成 HTML。
func markdownToHTML(req *MarkdownConvertRequest) (tools.Result, error) {
	extensions := []goldmark.Extender{}
	if req.GFM {
		extensions = append(extensions, extension.GFM)
	}

	converter := goldmark.New(goldmark.WithExtensions(extensions...))
	source := []byte(req.Input)
	document := converter.Parser().Parse(
		text.NewReader(source),
		parser.WithContext(parser.NewContext()),
	)

	var buffer bytes.Buffer
	if err := converter.Renderer().Render(&buffer, source, document); err != nil {
		return tools.Result{}, tools.InvalidField("input", "valid Markdown", "render error")
	}

	stats := collectMarkdownStats(source, document)

	return tools.Result{Data: MarkdownConvertResult{
		Mode:           MarkdownToHTML,
		Output:         buffer.String(),
		HeadingCount:   stats.headings,
		LinkCount:      stats.links,
		CodeBlockCount: stats.codeBlocks,
		WordCount:      stats.words,
	}}, nil
}

// htmlToMarkdown 把 HTML 转成 Markdown。
func htmlToMarkdown(req *MarkdownConvertRequest) (tools.Result, error) {
	root, err := html.Parse(strings.NewReader(req.Input))
	if err != nil {
		return tools.Result{}, tools.InvalidField("input", "valid HTML", "parse error")
	}

	var builder strings.Builder
	convertHTMLNodes(&builder, root, 0)

	output := builder.String()
	stats := countHTMLStats(output)

	return tools.Result{Data: MarkdownConvertResult{
		Mode:           HTMLToMarkdown,
		Output:         output,
		HeadingCount:   stats.headings,
		LinkCount:      stats.links,
		CodeBlockCount: stats.codeBlocks,
		WordCount:      stats.words,
	}}, nil
}

// markdownStats 是 Markdown 文档的统计。
type markdownStats struct {
	// headings 是标题数量。
	headings int
	// links 是链接数量。
	links int
	// codeBlocks 是代码块数量。
	codeBlocks int
	// words 是正文字数。
	words int
}

// collectMarkdownStats 统计 Markdown 文档。
func collectMarkdownStats(source []byte, document ast.Node) markdownStats {
	stats := markdownStats{}
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node.Kind() {
		case ast.KindHeading:
			stats.headings++
		case ast.KindLink:
			stats.links++
		case ast.KindCodeBlock, ast.KindFencedCodeBlock:
			stats.codeBlocks++
		}
		if textNode, ok := node.(*ast.Text); ok {
			stats.words += len([]rune(string(textNode.Value(source))))
		}
		return ast.WalkContinue, nil
	})
	return stats
}

// htmlStats 是 HTML 转换结果的统计。
type htmlStats struct {
	// headings 是标题数量。
	headings int
	// links 是链接数量。
	links int
	// codeBlocks 是代码块数量。
	codeBlocks int
	// words 是正文字数。
	words int
}

// countHTMLStats 统计转换后的 Markdown 文本。
func countHTMLStats(output string) htmlStats {
	stats := htmlStats{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#"):
			stats.headings++
		case strings.Contains(trimmed, "]("):
			stats.links++
		case strings.HasPrefix(trimmed, "```"):
			stats.codeBlocks++
		}
	}
	stats.words = len([]rune(strings.TrimSpace(output)))
	return stats
}

// convertHTMLNodes 把 HTML 子节点转成 Markdown 文本。
func convertHTMLNodes(builder *strings.Builder, node *html.Node, depth int) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		convertHTMLNode(builder, child, depth)
	}
}

// convertHTMLNode 把单个 HTML 节点转成 Markdown 文本。
func convertHTMLNode(builder *strings.Builder, node *html.Node, depth int) {
	switch node.Type {
	case html.TextNode:
		builder.WriteString(node.Data)
		return
	case html.ElementNode:
	default:
		// 注释等其他节点直接跳过。
		return
	}

	switch node.Data {
	case "html", "head", "body":
		// `html.Parse` 会补出 html / head / body 三层容器，
		// 它们的子树才是正文，必须继续向下走。
		convertHTMLNodes(builder, node, depth+1)
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(node.Data[1] - '0')
		builder.WriteString(strings.Repeat("#", level) + " ")
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("\n\n")
	case "p", "div", "section", "article", "header", "footer":
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("\n\n")
	case "br":
		builder.WriteString("\n")
	case "strong", "b":
		builder.WriteString("**")
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("**")
	case "em", "i":
		builder.WriteString("*")
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("*")
	case "code":
		if isInsidePre(node) {
			convertHTMLNodes(builder, node, depth)
			return
		}
		builder.WriteString("`")
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("`")
	case "pre":
		builder.WriteString("```\n")
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("\n```\n\n")
	case "blockquote":
		builder.WriteString("> ")
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("\n")
	case "ul", "ol":
		convertHTMLList(builder, node, depth)
	case "li":
		convertHTMLNodes(builder, node, depth)
		builder.WriteString("\n")
	case "a":
		href := ""
		text := ""
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				href = attr.Val
			}
		}
		var inner strings.Builder
		convertHTMLNodes(&inner, node.FirstChild, depth)
		text = inner.String()
		if href != "" {
			builder.WriteString("[" + text + "](" + href + ")")
		} else {
			builder.WriteString(text)
		}
	case "img":
		src := ""
		alt := ""
		for _, attr := range node.Attr {
			switch attr.Key {
			case "src":
				src = attr.Val
			case "alt":
				alt = attr.Val
			}
		}
		builder.WriteString("![" + alt + "](" + src + ")")
	case "hr":
		builder.WriteString("---\n\n")
	case "table":
		convertHTMLTable(builder, node, depth)
	case "script", "style":
		// 不转换脚本与样式内容。
		return
	default:
		convertHTMLNodes(builder, node, depth)
	}
}

// isInsidePre 判断节点是否在 pre 元素内。
func isInsidePre(node *html.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode && parent.Data == "pre" {
			return true
		}
	}
	return false
}

// convertHTMLList 转换列表。
func convertHTMLList(builder *strings.Builder, node *html.Node, depth int) {
	ordered := node.Data == "ol"
	index := 1
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode || child.Data != "li" {
			continue
		}
		prefix := "- "
		if ordered {
			prefix = itoa(index) + ". "
			index++
		}
		builder.WriteString(prefix)
		convertHTMLNodes(builder, child.FirstChild, depth+1)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

// convertHTMLTable 转换表格。
func convertHTMLTable(builder *strings.Builder, node *html.Node, depth int) {
	rows := [][]string{}
	var walk func(target *html.Node)
	walk = func(target *html.Node) {
		for child := target.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != html.ElementNode {
				continue
			}
			switch child.Data {
			case "tr":
				cells := []string{}
				for cell := child.FirstChild; cell != nil; cell = cell.NextSibling {
					if cell.Type != html.ElementNode || (cell.Data != "td" && cell.Data != "th") {
						continue
					}
					var inner strings.Builder
					convertHTMLNodes(&inner, cell.FirstChild, depth)
					cells = append(cells, strings.TrimSpace(inner.String()))
				}
				rows = append(rows, cells)
			default:
				walk(child)
			}
		}
	}
	walk(node)

	if len(rows) == 0 {
		return
	}

	columns := 0
	for _, row := range rows {
		if len(row) > columns {
			columns = len(row)
		}
	}

	builder.WriteString("| " + strings.Join(padRow(rows[0], columns), " | ") + " |\n")
	separators := make([]string, columns)
	for i := range separators {
		separators[i] = "---"
	}
	builder.WriteString("| " + strings.Join(separators, " | ") + " |\n")
	for _, row := range rows[1:] {
		builder.WriteString("| " + strings.Join(padRow(row, columns), " | ") + " |\n")
	}
	builder.WriteString("\n")
}

// padRow 补齐行内单元格数量。
func padRow(row []string, columns int) []string {
	out := make([]string, columns)
	for i := range out {
		if i < len(row) {
			out[i] = row[i]
		}
	}
	return out
}

// itoa 把整数转成字符串。
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// Validate 做字段级校验。
func (r *MarkdownConvertRequest) Validate() error {
	mode := r.Mode
	if mode == "" {
		mode = MarkdownToHTML
	}
	if !containsString(MarkdownConvertMode(), mode) {
		return tools.InvalidField("mode", MarkdownConvertMode(), mode)
	}
	if strings.TrimSpace(r.Input) == "" {
		return tools.MissingField("input")
	}
	if len(r.Input) > maxMarkdownBytes {
		return tools.InvalidField("input", []int{0, maxMarkdownBytes}, len(r.Input))
	}
	if len(r.Input) > maxMarkdownConvertBytes {
		return tools.InvalidField("input", []int{0, maxMarkdownConvertBytes}, len(r.Input))
	}
	return nil
}

// containsString 判断切片是否包含某值。
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
