package formatting

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// XMLFormatRequest 是 xml-format 工具的请求。
type XMLFormatRequest struct {
	// Input 是待处理的 XML 文本。
	Input TextInput `json:"input"`
	// Action 是 format、minify 或 validate。
	Action string `json:"action"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
	// KeepComments 是否保留注释。
	KeepComments bool `json:"keepComments"`
}

// XMLFormatResult 是 xml-format 工具的响应。
type XMLFormatResult struct {
	// Output 是处理结果；validate 时为空串。
	Output string `json:"output"`
	// Valid 是内容是否合法。
	Valid bool `json:"valid"`
	// Error 是语法错误的位置与原因；合法时为 nil。
	Error *SyntaxError `json:"error,omitempty"`
}

// XMLFormat 美化、压缩或校验 XML。
//
// 用 `encoding/xml` 的**流式**解码器逐 token 处理：
// 这样既保住了属性顺序与原始文本内容，
// 也不会把整棵树物化成 `map[string]any` 而丢掉结构信息。
func XMLFormat(req *XMLFormatRequest) (XMLFormatResult, error) {
	if err := req.Validate(); err != nil {
		return XMLFormatResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return XMLFormatResult{}, err
	}

	tokens, synErr := tokenizeXML(text)
	if synErr != nil {
		return XMLFormatResult{Output: "", Valid: false, Error: synErr}, nil
	}

	if action(req.Action) == ActionValidate {
		return XMLFormatResult{Valid: true}, nil
	}

	var out string
	switch action(req.Action) {
	case ActionMinify:
		out = minifyXML(tokens)
	default:
		out = formatXML(tokens, indentWidth(req.Indent), req.KeepComments)
	}

	return XMLFormatResult{Output: out, Valid: true}, nil
}

// Validate 做字段级校验。
func (r *XMLFormatRequest) Validate() error {
	if err := validateAction(r.Action); err != nil {
		return err
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	return r.Input.Validate("input")
}

// xmlToken 是 XML 的一个词法单元。
type xmlToken struct {
	// kind 是 token 类型。
	kind xmlTokenKind
	// text 是 token 文本。
	text string
	// attrs 是属性列表（有序）。
	attrs []xmlAttr
}

// xmlAttr 是一个属性。
type xmlAttr struct {
	// Name 是属性名。
	Name string
	// Value 是属性值。
	Value string
}

// xmlTokenKind 是 token 类型。
type xmlTokenKind int

const (
	// xmlTokenStart 是开始标签。
	xmlTokenStart xmlTokenKind = iota
	// xmlTokenEnd 是结束标签。
	xmlTokenEnd
	// xmlTokenText 是文本内容。
	xmlTokenText
	// xmlTokenComment 是注释。
	xmlTokenComment
	// xmlTokenDirective 是处理指令或 XML 声明。
	xmlTokenDirective
)

// tokenizeXML 把 XML 文本拆成词法单元。
func tokenizeXML(text string) ([]xmlToken, *SyntaxError) {
	dec := xml.NewDecoder(strings.NewReader(text))
	dec.Strict = true

	var tokens []xmlToken

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			se := xmlSyntaxError(err)
			return nil, &se
		}

		switch t := tok.(type) {
		case xml.StartElement:
			attrs := make([]xmlAttr, 0, len(t.Attr))
			for _, a := range t.Attr {
				attrs = append(attrs, xmlAttr{
					Name:  a.Name.Local,
					Value: a.Value,
				})
			}
			tokens = append(tokens, xmlToken{
				kind:  xmlTokenStart,
				text:  t.Name.Local,
				attrs: attrs,
			})
		case xml.EndElement:
			tokens = append(tokens, xmlToken{
				kind: xmlTokenEnd,
				text: t.Name.Local,
			})
		case xml.CharData:
			content := string(t)
			if strings.TrimSpace(content) == "" {
				// 纯空白文本在缩进输出里没有意义。
				continue
			}
			tokens = append(tokens, xmlToken{
				kind: xmlTokenText,
				text: content,
			})
		case xml.Comment:
			tokens = append(tokens, xmlToken{
				kind: xmlTokenComment,
				text: string(t),
			})
		case xml.ProcInst:
			tokens = append(tokens, xmlToken{
				kind: xmlTokenDirective,
				text: t.Target + " " + string(t.Inst),
			})
		case xml.Directive:
			tokens = append(tokens, xmlToken{
				kind: xmlTokenDirective,
				text: string(t),
			})
		}
	}

	return tokens, nil
}

// xmlElement 是一个 XML 元素。
type xmlElement struct {
	// name 是元素名。
	name string
	// attrs 是属性列表。
	attrs []xmlAttr
	// text 是元素自身的文本内容。
	text string
	// hasText 标记是否有文本。
	hasText bool
	// children 是子元素。
	children []*xmlElement
	// comments 是元素内的注释。
	comments []string
}

// buildXMLTree 把 token 流组装成元素树。
func buildXMLTree(tokens []xmlToken) []*xmlElement {
	var stack []*xmlElement
	var roots []*xmlElement

	for _, tok := range tokens {
		switch tok.kind {
		case xmlTokenStart:
			elem := &xmlElement{name: tok.text, attrs: tok.attrs}
			if len(stack) == 0 {
				roots = append(roots, elem)
			} else {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, elem)
			}
			stack = append(stack, elem)

		case xmlTokenEnd:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xmlTokenText:
			if len(stack) == 0 {
				continue
			}
			current := stack[len(stack)-1]
			current.text += strings.TrimSpace(tok.text)
			current.hasText = true

		case xmlTokenComment:
			if len(stack) == 0 {
				continue
			}
			current := stack[len(stack)-1]
			current.comments = append(current.comments, tok.text)
		}
	}

	return roots
}

// formatXML 美化 XML。
//
// 只有**纯文本**元素保持内联（`<a>1</a>`）：
// 把文本换到下一行会在元素内容里插入空白，
// 等于**改写**了数据。
func formatXML(tokens []xmlToken, indent int, keepComments bool) string {
	pad := indentString(indent)
	var buf bytes.Buffer

	for _, elem := range buildXMLTree(tokens) {
		writeXMLElement(&buf, elem, pad, keepComments, 0)
	}

	return strings.TrimRight(buf.String(), "\n")
}

// writeXMLElement 递归输出 XML 元素。
func writeXMLElement(buf *bytes.Buffer, elem *xmlElement, pad string, keepComments bool, depth int) {
	prefix := strings.Repeat(pad, depth)

	// 纯文本元素：内联输出。
	if len(elem.children) == 0 {
		buf.WriteString(prefix + "<" + elem.name + renderXMLAttrs(elem.attrs) + ">")
		if elem.hasText {
			buf.WriteString(escapeXMLText(elem.text))
		}
		buf.WriteString("</" + elem.name + ">")
		buf.WriteString("\n")
		return
	}

	buf.WriteString(prefix + "<" + elem.name + renderXMLAttrs(elem.attrs) + ">")
	buf.WriteString("\n")

	if elem.hasText {
		buf.WriteString(strings.Repeat(pad, depth+1))
		buf.WriteString(escapeXMLText(elem.text))
		buf.WriteString("\n")
	}

	for _, comment := range elem.comments {
		if !keepComments {
			continue
		}
		buf.WriteString(strings.Repeat(pad, depth+1))
		buf.WriteString("<!--" + comment + "-->")
		buf.WriteString("\n")
	}

	for _, child := range elem.children {
		writeXMLElement(buf, child, pad, keepComments, depth+1)
	}

	buf.WriteString(prefix + "</" + elem.name + ">")
	buf.WriteString("\n")
}

// minifyXML 压缩 XML：去掉标签之间的空白。
func minifyXML(tokens []xmlToken) string {
	var buf bytes.Buffer

	for _, tok := range tokens {
		switch tok.kind {
		case xmlTokenStart:
			buf.WriteString("<" + tok.text + renderXMLAttrs(tok.attrs) + ">")
		case xmlTokenEnd:
			buf.WriteString("</" + tok.text + ">")
		case xmlTokenText:
			buf.WriteString(strings.TrimSpace(tok.text))
		case xmlTokenComment:
			buf.WriteString("<!--" + tok.text + "-->")
		case xmlTokenDirective:
			buf.WriteString("<?" + tok.text + "?>")
		}
	}

	return buf.String()
}

// renderXMLAttrs 渲染属性列表。
func renderXMLAttrs(attrs []xmlAttr) string {
	if len(attrs) == 0 {
		return ""
	}

	var b strings.Builder
	for _, a := range attrs {
		b.WriteString(" ")
		b.WriteString(a.Name)
		b.WriteString(`="`)
		b.WriteString(escapeXMLAttr(a.Value))
		b.WriteString(`"`)
	}
	return b.String()
}

// escapeXMLAttr 转义属性值里的特殊字符。
func escapeXMLAttr(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return replacer.Replace(s)
}

// escapeXMLText 转义文本内容里的特殊字符。
func escapeXMLText(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(s)
}

// xmlSyntaxError 把 XML 解析错误转成带行列的结构化错误。
func xmlSyntaxError(err error) SyntaxError {
	se := SyntaxError{Message: "invalid XML"}

	var syntaxErr *xml.SyntaxError
	if errors.As(err, &syntaxErr) {
		se.Line = syntaxErr.Line
	}
	return se
}
