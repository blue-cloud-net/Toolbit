// Package text 实现文本类工具。
//
// 本包只做「解码输入 → 校验字段 → 计算 → 编码 DTO」。
package text

import (
	"regexp"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 正则标志。
const (
	// RegexFlagIgnoreCase 忽略大小写。
	RegexFlagIgnoreCase = "i"
	// RegexFlagMultiline 多行模式。
	RegexFlagMultiline = "m"
	// RegexFlagDotAll 让 `.` 匹配换行。
	RegexFlagDotAll = "s"
	// RegexFlagExtended 忽略空白与注释。
	RegexFlagExtended = "x"
)

// RegexFlags 返回受支持的正则标志（顺序稳定）。
func RegexFlags() []string {
	return []string{RegexFlagIgnoreCase, RegexFlagMultiline, RegexFlagDotAll, RegexFlagExtended}
}

// maxRegexMatches 是匹配结果条数上限。
//
// 超过上限返回 TOO_MANY_ITEMS：正则在病态输入上可能匹配出
// 数量爆炸的结果，全部返回会耗尽内存。
const maxRegexMatches = 1000

// RegexTesterRequest 是 regex-tester 工具的请求。
type RegexTesterRequest struct {
	// Pattern 是正则表达式。
	Pattern string `json:"pattern"`
	// Flags 是标志名列表。
	Flags []string `json:"flags"`
	// Input 是待匹配的文本。
	Input TextInput `json:"input"`
	// MaxMatches 是匹配条数上限；默认 1000。
	MaxMatches int `json:"maxMatches"`
	// GroupNames 是捕获组名称（按顺序对应第 1、2 … 组）。
	GroupNames []string `json:"groupNames"`
}

// TextInput 是纯文本输入。
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

// RegexMatch 是一条匹配。
type RegexMatch struct {
	// Index 是匹配在原文中的起始字节偏移。
	Index int `json:"index"`
	// Text 是匹配的文本。
	Text string `json:"text"`
	// Line 是匹配所在行号，从 1 开始。
	Line int `json:"line"`
	// Column 是匹配所在列号，从 1 开始。
	Column int `json:"column"`
	// Groups 是捕获组内容；未参与匹配的组为空串。
	Groups []string `json:"groups"`
	// Names 是捕获组的命名（与 GroupNames 对应）。
	Names map[string]string `json:"names,omitempty"`
}

// RegexTesterResult 是 regex-tester 工具的响应。
type RegexTesterResult struct {
	// Matches 是匹配列表。
	Matches []RegexMatch `json:"matches"`
	// Count 是匹配总数。
	Count int `json:"count"`
	// Truncated 为真表示结果被上限截断。
	Truncated bool `json:"truncated"`
	// Pattern 是实际使用的正则。
	Pattern string `json:"pattern"`
}

// RegexTester 测试正则表达式。
//
// 服务端使用 Go 的 RE2 语法：**线性时间**，不支持反向引用与前后向断言。
// 这类语法在编译期即被拒绝，返回 INVALID_REGEX。
func RegexTester(req *RegexTesterRequest) (RegexTesterResult, error) {
	if err := req.Validate(); err != nil {
		return RegexTesterResult{}, err
	}

	source, err := req.Input.Text()
	if err != nil {
		return RegexTesterResult{}, err
	}

	expr, err := compileRegex(req.Pattern, req.Flags)
	if err != nil {
		return RegexTesterResult{}, err
	}

	limit := req.MaxMatches
	if limit <= 0 || limit > maxRegexMatches {
		limit = maxRegexMatches
	}

	matches := expr.FindAllStringSubmatchIndex(source, limit+1)
	truncated := len(matches) > limit
	if truncated {
		matches = matches[:limit]
	}

	out := make([]RegexMatch, 0, len(matches))
	for _, idx := range matches {
		groups := make([]string, 0, len(idx)/2-1)
		for i := 1; i*2 < len(idx); i++ {
			if idx[i*2] < 0 {
				groups = append(groups, "")
				continue
			}
			groups = append(groups, source[idx[i*2]:idx[i*2+1]])
		}

		line, column := lineColumn(source, idx[0])
		match := RegexMatch{
			Index:  idx[0],
			Text:   source[idx[0]:idx[1]],
			Line:   line,
			Column: column,
			Groups: groups,
		}

		if len(req.GroupNames) > 0 {
			match.Names = map[string]string{}
			for i, name := range req.GroupNames {
				if i < len(groups) {
					match.Names[name] = groups[i]
				}
			}
		}

		out = append(out, match)
	}

	result := RegexTesterResult{
		Matches:   out,
		Count:     len(out),
		Truncated: truncated,
		Pattern:   expr.String(),
	}

	return result, nil
}

// Validate 做字段级校验。
func (r *RegexTesterRequest) Validate() error {
	if strings.TrimSpace(r.Pattern) == "" {
		return tools.MissingField("pattern")
	}
	for _, flag := range r.Flags {
		if !isKnownFlag(flag) {
			return tools.InvalidField("flags", RegexFlags(), flag)
		}
	}
	if r.MaxMatches < 0 || r.MaxMatches > maxRegexMatches {
		return tools.InvalidField("maxMatches", []int{0, maxRegexMatches}, r.MaxMatches)
	}
	return r.Input.Validate("input")
}

// isKnownFlag 报告标志是否受支持。
func isKnownFlag(flag string) bool {
	switch flag {
	case RegexFlagIgnoreCase, RegexFlagMultiline, RegexFlagDotAll, RegexFlagExtended:
		return true
	default:
		return false
	}
}

// compileRegex 编译正则表达式。
//
// 不受支持的构造（反向引用、前后向断言）会被 Go 的解析器拒绝，
// 统一映射为 INVALID_REGEX。
func compileRegex(pattern string, flags []string) (*regexp.Regexp, error) {
	prefix := ""
	for _, flag := range flags {
		switch flag {
		case RegexFlagIgnoreCase:
			prefix += "i"
		case RegexFlagMultiline:
			prefix += "m"
		case RegexFlagDotAll:
			prefix += "s"
		case RegexFlagExtended:
			prefix += "x"
		}
	}

	source := pattern
	if prefix != "" {
		source = "(?" + prefix + ")" + pattern
	}

	expr, err := regexp.Compile(source)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidRegex, map[string]any{
			"reason": "unsupported or invalid regular expression",
		})
	}
	return expr, nil
}

// hasDialectDifference 报告表达式是否用了 RE2 不支持的构造。
//
// RE2 在编译期就会拒绝这些构造，因此走到这里说明**编译成功**，
// 只能用于提示「这条表达式在别的方言里也许可用」。
func hasDialectDifference(pattern string) bool {
	return strings.Contains(pattern, `(?=`) ||
		strings.Contains(pattern, `(?!`) ||
		strings.Contains(pattern, `(?<=`) ||
		strings.Contains(pattern, `(?<!`)
}

// lineColumn 计算字节偏移对应的行列。
func lineColumn(source string, offset int) (int, int) {
	if offset < 0 || offset > len(source) {
		return 1, 1
	}
	before := source[:offset]
	line := strings.Count(before, "\n") + 1
	lastNewline := strings.LastIndex(before, "\n")
	column := len(before) + 1
	if lastNewline >= 0 {
		column = len(before) - lastNewline
	}
	return line, column
}
