package formatting

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// HTMLFormatRequest 是 html-format 工具的请求。
type HTMLFormatRequest struct {
	// Input 是待处理的 HTML 文本。
	Input TextInput `json:"input"`
	// Action 是 format、minify 或 validate。
	Action string `json:"action"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
}

// HTMLFormatResult 是 html-format 工具的响应。
type HTMLFormatResult struct {
	// Output 是处理结果；validate 时为空串。
	Output string `json:"output"`
	// Valid 是内容是否合法。
	Valid bool `json:"valid"`
	// Error 是语法错误的位置与原因；合法时为 nil。
	Error *SyntaxError `json:"error,omitempty"`
}

// HTMLFormat 美化、压缩或校验 HTML。
//
// 用 `x/net/html` 解析：它是容错解析器，会自动补全未闭合的标签，
// 因此**几乎不会**报语法错误——`validate` 的结论是「能否解析出结构」。
//
// 解析走 **fragment** 模式：不会给输入硬套上 `<html><head><body>`，
// 因为那等于往用户的内容里插入元素。
//
// 输出**不是**净化结果：格式化后的 HTML **不要**直接嵌入网页，
// 需要净化请另行处理。
func HTMLFormat(req *HTMLFormatRequest) (HTMLFormatResult, error) {
	if err := req.Validate(); err != nil {
		return HTMLFormatResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return HTMLFormatResult{}, err
	}

	context := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, parseErr := html.ParseFragment(strings.NewReader(text), context)
	if parseErr != nil {
		se := SyntaxError{Message: "invalid HTML"}
		return HTMLFormatResult{Output: "", Valid: false, Error: &se}, nil
	}

	if action(req.Action) == ActionValidate {
		return HTMLFormatResult{Valid: true}, nil
	}

	var out string
	switch action(req.Action) {
	case ActionMinify:
		out = minifyHTML(nodes)
	default:
		out = formatHTML(nodes, indentWidth(req.Indent))
	}

	return HTMLFormatResult{Output: out, Valid: true}, nil
}

// Validate 做字段级校验。
func (r *HTMLFormatRequest) Validate() error {
	if err := validateAction(r.Action); err != nil {
		return err
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	return r.Input.Validate("input")
}

// voidElements 是没有结束标签的 HTML 元素。
var voidElements = map[atom.Atom]bool{
	atom.Area:   true,
	atom.Base:   true,
	atom.Br:     true,
	atom.Col:    true,
	atom.Embed:  true,
	atom.Hr:     true,
	atom.Img:    true,
	atom.Input:  true,
	atom.Link:   true,
	atom.Meta:   true,
	atom.Param:  true,
	atom.Source: true,
	atom.Track:  true,
	atom.Wbr:    true,
}

// formatHTML 美化 HTML。
func formatHTML(nodes []*html.Node, indent int) string {
	pad := indentString(indent)
	var buf bytes.Buffer

	for _, node := range nodes {
		writeHTMLNode(&buf, node, pad, 0)
	}

	return strings.TrimRight(buf.String(), "\n")
}

// writeHTMLNode 递归写出 HTML 节点。
func writeHTMLNode(buf *bytes.Buffer, node *html.Node, pad string, depth int) {
	switch node.Type {
	case html.TextNode:
		text := strings.TrimSpace(node.Data)
		if text == "" {
			return
		}
		buf.WriteString(strings.Repeat(pad, depth))
		buf.WriteString(escapeHTMLText(text))
		buf.WriteString("\n")

	case html.CommentNode:
		buf.WriteString(strings.Repeat(pad, depth))
		buf.WriteString("<!--" + strings.TrimSpace(node.Data) + "-->")
		buf.WriteString("\n")

	case html.DoctypeNode:
		buf.WriteString(strings.Repeat(pad, depth))
		buf.WriteString("<!DOCTYPE " + strings.TrimSpace(node.Data) + ">")
		buf.WriteString("\n")

	case html.ElementNode:
		inline := htmlHasOnlyText(node)

		buf.WriteString(strings.Repeat(pad, depth))
		buf.WriteString("<" + node.Data)
		for _, attr := range node.Attr {
			buf.WriteString(" " + attr.Key)
			if attr.Val != "" {
				buf.WriteString(`="` + escapeHTMLAttr(attr.Val) + `"`)
			}
		}
		buf.WriteString(">")

		if inline {
			// 纯文本元素保持内联：换行会在内容里插入空白。
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type != html.TextNode {
					continue
				}
				buf.WriteString(escapeHTMLText(strings.TrimSpace(child.Data)))
			}
		} else {
			buf.WriteString("\n")
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				writeHTMLNode(buf, child, pad, depth+1)
			}
			buf.WriteString(strings.Repeat(pad, depth))
		}

		if !voidElements[node.DataAtom] {
			buf.WriteString("</" + node.Data + ">")
		}
		buf.WriteString("\n")

	case html.DocumentNode:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			writeHTMLNode(buf, child, pad, depth)
		}
	}
}

// htmlHasOnlyText 判断元素是否只包含文本内容。
//
// 这类元素（如 `<p>hello</p>`）输出时保持内联：
// 换行会在元素内容里插入空白，等于**改写**了数据。
func htmlHasOnlyText(node *html.Node) bool {
	hasChild := false
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			if strings.TrimSpace(child.Data) == "" {
				continue
			}
			hasChild = true
			continue
		}
		return false
	}
	return hasChild
}

// minifyHTML 压缩 HTML：去掉标签之间的空白。
func minifyHTML(nodes []*html.Node) string {
	var buf bytes.Buffer
	for _, node := range nodes {
		writeHTMLMinified(&buf, node)
	}
	return buf.String()
}

// writeHTMLMinified 递归写出压缩后的 HTML 节点。
func writeHTMLMinified(buf *bytes.Buffer, node *html.Node) {
	switch node.Type {
	case html.TextNode:
		text := strings.TrimSpace(node.Data)
		if text != "" {
			buf.WriteString(escapeHTMLText(text))
		}

	case html.CommentNode:
		buf.WriteString("<!--" + strings.TrimSpace(node.Data) + "-->")

	case html.DoctypeNode:
		buf.WriteString("<!DOCTYPE " + strings.TrimSpace(node.Data) + ">")

	case html.ElementNode:
		buf.WriteString("<" + node.Data)
		for _, attr := range node.Attr {
			buf.WriteString(" " + attr.Key)
			if attr.Val != "" {
				buf.WriteString(`="` + escapeHTMLAttr(attr.Val) + `"`)
			}
		}
		buf.WriteString(">")

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			writeHTMLMinified(buf, child)
		}

		if !voidElements[node.DataAtom] {
			buf.WriteString("</" + node.Data + ">")
		}

	case html.DocumentNode:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			writeHTMLMinified(buf, child)
		}
	}
}

// escapeHTMLAttr 转义 HTML 属性值里的特殊字符。
func escapeHTMLAttr(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return replacer.Replace(s)
}

// escapeHTMLText 转义 HTML 文本内容里的特殊字符。
func escapeHTMLText(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(s)
}
