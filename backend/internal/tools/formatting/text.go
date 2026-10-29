package formatting

import (
	"strings"
	"unicode/utf8"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// TextInput 是纯文本输入。
//
// 与 `tools.Bytes` 的区别：文本工具关心的是「字符」而不是「字节」，
// 因此**不**提供 hex / base64 之类的传输编码选项。
type TextInput struct {
	// Data 是文本内容。
	Data string `json:"data"`
}

// Text 返回文本内容。
func (t TextInput) Text() (string, error) {
	if t.Data == "" {
		return "", tools.MissingField("input.data")
	}
	return t.Data, nil
}

// Validate 校验文本输入。
func (t TextInput) Validate(field string) error {
	if t.Data == "" {
		return tools.MissingField(field + ".data")
	}
	return nil
}

// positionOf 把字节偏移换算成行列。
//
// offset 是 1-based 的字节偏移（标准库的 json.SyntaxError 口径）；
// 越界或为 0 时返回 (0, 0)，表示「无法定位」。
func positionOf(source string, offset int) (int, int) {
	if offset <= 0 || offset > len(source)+1 {
		return 0, 0
	}

	before := source[:offset-1]
	line := strings.Count(before, "\n") + 1
	lastNewline := strings.LastIndex(before, "\n")

	column := utf8.RuneCountInString(before) + 1
	if lastNewline >= 0 {
		column = utf8.RuneCountInString(before[lastNewline+1:]) + 1
	}
	return line, column
}
