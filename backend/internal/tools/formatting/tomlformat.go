package formatting

import (
	"bytes"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// TOMLFormatRequest 是 toml-format 工具的请求。
type TOMLFormatRequest struct {
	// Input 是待处理的 TOML 文本。
	Input TextInput `json:"input"`
	// Action 是 format、minify 或 validate。
	Action string `json:"action"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
}

// TOMLFormatResult 是 toml-format 工具的响应。
type TOMLFormatResult struct {
	// Output 是处理结果；validate 时为空串。
	Output string `json:"output"`
	// Valid 是内容是否合法。
	Valid bool `json:"valid"`
	// Error 是语法错误的位置与原因；合法时为 nil。
	Error *SyntaxError `json:"error,omitempty"`
	// CommentsLost 标记注释是否丢失。
	CommentsLost bool `json:"commentsLost,omitempty"`
}

// TOMLFormat 美化、压缩或校验 TOML。
//
// TOML 解析器**不保留注释**（见 docs/api/tools/formatting.md），
// 因此结果里带 `commentsLost` 标记，由调用方提示用户。
func TOMLFormat(req *TOMLFormatRequest) (TOMLFormatResult, error) {
	if err := req.Validate(); err != nil {
		return TOMLFormatResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return TOMLFormatResult{}, err
	}

	var value map[string]any
	if err := toml.Unmarshal([]byte(text), &value); err != nil {
		se := tomlSyntaxError(err)
		return TOMLFormatResult{Output: "", Valid: false, Error: &se}, nil
	}

	if action(req.Action) == ActionValidate {
		return TOMLFormatResult{Valid: true, CommentsLost: hasTOMLComment(text)}, nil
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(value); err != nil {
		se := SyntaxError{Message: "invalid TOML"}
		return TOMLFormatResult{Output: "", Valid: false, Error: &se}, nil
	}

	out := strings.TrimRight(buf.String(), "\n")
	if action(req.Action) == ActionMinify {
		out = minifyTOML(out)
	}

	return TOMLFormatResult{
		Output:       out,
		Valid:        true,
		CommentsLost: hasTOMLComment(text),
	}, nil
}

// Validate 做字段级校验。
func (r *TOMLFormatRequest) Validate() error {
	if err := validateAction(r.Action); err != nil {
		return err
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	return r.Input.Validate("input")
}

// minifyTOML 压缩 TOML：去掉空行。
func minifyTOML(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, strings.TrimRight(line, " \t"))
	}
	return strings.Join(out, "\n")
}

// hasTOMLComment 判断原文是否含注释。
func hasTOMLComment(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if idx := strings.Index(line, "#"); idx >= 0 {
			// `#` 出现在字符串里不算注释；这里只做粗判，
			// 宁可多报也不漏报。
			if idx == 0 || line[idx-1] == ' ' || line[idx-1] == '\t' {
				return true
			}
		}
	}
	return false
}

// tomlSyntaxError 把 go-toml 的错误转成带行列的结构化错误。
//
// go-toml 的错误消息形如：
// `toml: line 3: expected ...`
func tomlSyntaxError(err error) SyntaxError {
	se := SyntaxError{Message: "invalid TOML"}

	msg := err.Error()
	if loc := yamlLineRE.FindStringSubmatch(msg); loc != nil {
		if line, convErr := atoiSafe(loc[1]); convErr == nil {
			se.Line = line
			se.Column = 1
		}
	}
	return se
}
