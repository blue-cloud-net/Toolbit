// Package formatting 实现格式化类工具。
//
// 本包只做「解析 → 重新排版 → 编码 DTO」：
// 解析交给标准库与专用解析器，本包**不**实现任何解析算法。
//
// 8 个工具共用同一套约定：
//
//   - `action` 三态：format（美化）、minify（压缩）、validate（只校验）；
//   - `indent` 统一表示缩进宽度，默认 2；
//   - 语法错误统一返回 `{ line, column, message }`，**不**回显用户输入。
package formatting

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 格式化操作。
const (
	// ActionFormat 美化：重新排版并加缩进。
	ActionFormat = "format"
	// ActionMinify 压缩：去掉所有非必要空白。
	ActionMinify = "minify"
	// ActionValidate 校验：只判断是否合法，不改写内容。
	ActionValidate = "validate"
)

// Actions 返回受支持的操作名（顺序稳定）。
func Actions() []string {
	return []string{ActionFormat, ActionMinify, ActionValidate}
}

// DefaultIndent 是默认缩进宽度。
const DefaultIndent = 2

// MinIndent 是最小缩进宽度。
const MinIndent = 0

// MaxIndent 是最大缩进宽度。
const MaxIndent = 10

// SyntaxError 是语法错误的结构化描述。
//
// 只放**位置**与**原因**，不回显用户输入。
type SyntaxError struct {
	// Line 是出错行号，从 1 开始；无法定位时为 0。
	Line int `json:"line"`
	// Column 是出错列号，从 1 开始；无法定位时为 0。
	Column int `json:"column"`
	// Message 是错误原因（英文，不含用户输入）。
	Message string `json:"message"`
}

// SyntaxErr 把 SyntaxError 包成工具错误。
//
// code 是该格式对应的语法错误码（INVALID_JSON / INVALID_YAML …）。
func SyntaxErr(code toolerr.Code, se SyntaxError) error {
	params := map[string]any{
		"line":    se.Line,
		"column":  se.Column,
		"message": se.Message,
	}
	return toolerr.Newf(code, params)
}

// action 返回归一化后的操作；缺省为 format。
func action(raw string) string {
	if raw == "" {
		return ActionFormat
	}
	return raw
}

// validateAction 校验操作名。
func validateAction(raw string) error {
	switch raw {
	case "", ActionFormat, ActionMinify, ActionValidate:
		return nil
	default:
		return tools.InvalidField("action", Actions(), raw)
	}
}

// indentWidth 返回缩进宽度；缺省为 DefaultIndent。
func indentWidth(raw int) int {
	if raw == 0 {
		return DefaultIndent
	}
	return raw
}

// validateIndent 校验缩进宽度。
func validateIndent(raw int) error {
	if raw < MinIndent || raw > MaxIndent {
		return tools.InvalidField("indent",
			[]int{MinIndent, MaxIndent}, raw)
	}
	return nil
}

// indentString 返回缩进字符串。
func indentString(width int) string {
	if width <= 0 {
		return ""
	}
	out := make([]byte, 0, width)
	for i := 0; i < width; i++ {
		out = append(out, ' ')
	}
	return string(out)
}
