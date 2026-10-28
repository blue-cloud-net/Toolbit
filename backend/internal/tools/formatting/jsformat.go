package formatting

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/tdewolff/minify/v2"
	minjs "github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// jsLineRE 匹配 JS 解析错误消息里的行号。
var jsLineRE = regexp.MustCompile(`line (\d+)`)

// JSFormatRequest 是 js-format 工具的请求。
type JSFormatRequest struct {
	// Input 是待处理的 JavaScript 源码。
	Input TextInput `json:"input"`
	// Action 是 format 或 minify。
	Action string `json:"action"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
	// Quotes 是引号风格：double（默认）或 single。
	Quotes string `json:"quotes"`
	// Semi 是否保留分号。
	Semi *bool `json:"semi"`
}

// JSFormatResult 是 js-format 工具的响应。
type JSFormatResult struct {
	// Output 是处理结果；validate 时为空串。
	Output string `json:"output"`
	// Valid 是内容是否合法。
	Valid bool `json:"valid"`
	// Error 是语法错误的位置与原因；合法时为 nil。
	Error *SyntaxError `json:"error,omitempty"`
}

// JS 引号风格。
const (
	// JSQuotesDouble 使用双引号。
	JSQuotesDouble = "double"
	// JSQuotesSingle 使用单引号。
	JSQuotesSingle = "single"
)

// JSFormat 美化、压缩或校验 JavaScript。
//
// 解析与打印由 `tdewolff/parse` 完成——它实现的是 ECMAScript 语法，
// **不执行**任何代码。
//
// 打印器输出的缩进固定为 4 空格，因此这里在**打印结果**上做缩进重映射：
// 只调整行首空白，不动任何 token。
func JSFormat(req *JSFormatRequest) (JSFormatResult, error) {
	if err := req.Validate(); err != nil {
		return JSFormatResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return JSFormatResult{}, err
	}

	ast, parseErr := js.Parse(parse.NewInputString(text), js.Options{})
	if parseErr != nil {
		se := jsSyntaxError(parseErr)
		return JSFormatResult{Output: "", Valid: false, Error: &se}, nil
	}

	if action(req.Action) == ActionValidate {
		return JSFormatResult{Valid: true}, nil
	}

	switch action(req.Action) {
	case ActionMinify:
		out, err := minifyJS(text)
		if err != nil {
			se := jsSyntaxError(err)
			return JSFormatResult{Output: "", Valid: false, Error: &se}, nil
		}
		return JSFormatResult{Output: out, Valid: true}, nil
	default:
		out := formatJS(ast.JSString(), indentWidth(req.Indent), req.Quotes, req.semi())
		return JSFormatResult{Output: out, Valid: true}, nil
	}
}

// Validate 做字段级校验。
func (r *JSFormatRequest) Validate() error {
	if err := validateAction(r.Action); err != nil {
		return err
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	switch r.Quotes {
	case "", JSQuotesDouble, JSQuotesSingle:
	default:
		return tools.InvalidField("quotes", []string{JSQuotesDouble, JSQuotesSingle}, r.Quotes)
	}
	return r.Input.Validate("input")
}

// semi 返回是否保留分号；缺省为 true。
func (r *JSFormatRequest) semi() bool {
	if r.Semi == nil {
		return true
	}
	return *r.Semi
}

// formatJS 调整打印结果的缩进与引号风格。
func formatJS(printed string, indent int, quotes string, semi bool) string {
	lines := strings.Split(printed, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		out = append(out, remapIndent(line, indent))
	}

	text := strings.Join(out, "\n")
	if quotes == JSQuotesSingle {
		text = swapQuotes(text)
	}
	if !semi {
		text = stripSemicolons(text)
	}
	return strings.TrimRight(text, "\n")
}

// jsPrinterIndent 是打印器的固定缩进宽度。
const jsPrinterIndent = 4

// remapIndent 把行首缩进从 4 空格重映射为目标宽度。
//
// 打印器输出的行首空格数必然是 4 的倍数（每层一个缩进），
// 因此可以安全地按比例缩放。若某行不是 4 的倍数
// （例如多行模板字符串的续行），则**原样保留**该行。
func remapIndent(line string, indent int) string {
	spaces := 0
	for spaces < len(line) && line[spaces] == ' ' {
		spaces++
	}
	if spaces == 0 || spaces%jsPrinterIndent != 0 {
		return line
	}
	levels := spaces / jsPrinterIndent
	return strings.Repeat(indentString(indent), levels) + line[spaces:]
}

// swapQuotes 把字符串字面量的双引号换成单引号。
//
// 只处理**整条**都是字符串字面量的情况：即整行去掉首尾空白后
// 以 `"` 开头且以 `"` 结尾，且内部没有未转义的双引号。
// 这样可以避免误伤代码中的其它双引号。
func swapQuotes(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' &&
			!strings.Contains(trimmed[1:len(trimmed)-1], `"`) {
			prefix := line[:len(line)-len(strings.TrimLeft(line, " "))]
			out = append(out, prefix+"'"+trimmed[1:len(trimmed)-1]+"'")
			continue
		}
		out = append(out, line)
	}

	return strings.Join(out, "\n")
}

// stripSemicolons 去掉行尾分号。
//
// 只处理分号**恰好是行尾最后一个非空白字符**的情况；
// 出现在字符串里的分号不受影响（因为那不会在行尾）。
func stripSemicolons(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if strings.HasSuffix(trimmed, ";") && !strings.HasSuffix(trimmed, `";`) &&
			!strings.HasSuffix(trimmed, `';`) {
			out = append(out, strings.TrimRight(trimmed[:len(trimmed)-1], " \t"))
			continue
		}
		out = append(out, line)
	}

	return strings.Join(out, "\n")
}

// minifyJS 压缩 JavaScript。
//
// 用 `tdewolff/minify` 的 JS 压缩器：它会去掉注释、多余空白与
// 可省略的符号，是真正的压缩而非「逐行去缩进」。
//
// **必须**关掉变量重命名：重命名虽不改变语义（局部作用域），
// 但会让压缩结果与原文对不上，调试时无法阅读。
func minifyJS(src string) (string, error) {
	m := minify.New()
	m.AddFunc("application/javascript", (&minjs.Minifier{KeepVarNames: true}).Minify)

	var buf bytes.Buffer
	if err := m.MinifyMimetype([]byte("application/javascript"), &buf, strings.NewReader(src), nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// jsSyntaxError 把 JS 解析错误转成带行列的结构化错误。
func jsSyntaxError(err error) SyntaxError {
	se := SyntaxError{Message: "invalid JavaScript"}

	msg := err.Error()
	if loc := jsLineRE.FindStringSubmatch(msg); loc != nil {
		if line, convErr := atoiSafe(loc[1]); convErr == nil {
			se.Line = line
			se.Column = 1
		}
	}
	return se
}
