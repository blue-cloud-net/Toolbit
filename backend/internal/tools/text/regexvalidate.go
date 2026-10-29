package text

import (
	"context"
	"regexp"
	"regexp/syntax"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// RegexValidateRequest 是 regex-validate 工具的请求。
type RegexValidateRequest struct {
	// Pattern 是待校验的正则表达式。
	Pattern string `json:"pattern"`
	// Flags 是标志名列表。
	Flags []string `json:"flags"`
}

// RegexError 是正则表达式的语法错误。
type RegexError struct {
	// Code 是错误码标识（syntax）。
	Code string `json:"code"`
	// Message 是错误说明。
	Message string `json:"message"`
	// Offset 是错误在表达式中的字节偏移；无偏移时为 -1。
	Offset int `json:"offset"`
	// Position 是错误所在的行列位置。
	Position string `json:"position"`
}

// RegexSuggestion 是一条改进建议。
type RegexSuggestion struct {
	// Code 是建议标识。
	Code string `json:"code"`
	// Message 是建议说明。
	Message string `json:"message"`
	// Priority 是优先级：high / medium / low。
	Priority string `json:"priority"`
}

// RegexValidateResult 是 regex-validate 工具的响应。
type RegexValidateResult struct {
	// Valid 是否是合法表达式。
	Valid bool `json:"valid"`
	// Dialect 是正则方言；本工具固定为 RE2。
	Dialect string `json:"dialect"`
	// Flags 是实际生效的标志。
	Flags []string `json:"flags"`
	// Error 是语法错误；合法时为空。
	Error *RegexError `json:"error"`
	// Suggestions 是改进建议。
	Suggestions []RegexSuggestion `json:"suggestions"`
	// CaptureGroups 是捕获组数量。
	CaptureGroups int `json:"captureGroups"`
}

// maxRegexPatternBytes 是正则表达式的字节上限。
const maxRegexPatternBytes = 4096

// RegexValidate 校验正则表达式的语法。
//
// 本工具**只校验不匹配**：不接收待匹配文本，因此不存在超时的可能。
func RegexValidate(_ context.Context, req *RegexValidateRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	flags, err := resolveRegexFlags(req.Flags)
	if err != nil {
		return tools.Result{}, err
	}

	compiled, compileErr := regexp.Compile(buildRegexPattern(req.Pattern, flags))
	if compileErr != nil {
		return tools.Result{Data: RegexValidateResult{
			Valid:         false,
			Dialect:       "RE2",
			Flags:         flags,
			Error:         buildRegexError(req.Pattern, compileErr),
			Suggestions:   suggestRegexFix(req.Pattern),
			CaptureGroups: 0,
		}}, nil
	}

	captureGroups := 0
	if parsed, parseErr := syntax.Parse(req.Pattern, syntax.Perl); parseErr == nil {
		captureGroups = parsed.MaxCap()
	} else {
		captureGroups = compiled.NumSubexp()
	}

	return tools.Result{Data: RegexValidateResult{
		Valid:         true,
		Dialect:       "RE2",
		Flags:         flags,
		Error:         nil,
		Suggestions:   suggestRegexImprovements(req.Pattern, compiled),
		CaptureGroups: captureGroups,
	}}, nil
}

// buildRegexPattern 把表达式与标志拼成 Go 的正则字面量。
func buildRegexPattern(pattern string, flags []string) string {
	prefix := ""
	if containsString(flags, RegexFlagIgnoreCase) {
		prefix += "i"
	}
	if containsString(flags, RegexFlagMultiline) {
		prefix += "m"
	}
	if containsString(flags, RegexFlagDotAll) {
		prefix += "s"
	}
	if containsString(flags, RegexFlagExtended) {
		prefix += "x"
	}
	return "(?" + prefix + ")" + pattern
}

// resolveRegexFlags 校验并归一化标志列表。
func resolveRegexFlags(flags []string) ([]string, error) {
	supported := RegexFlags()
	seen := map[string]bool{}
	out := []string{}
	for _, flag := range flags {
		flag = strings.TrimSpace(flag)
		if flag == "" {
			continue
		}
		normalized := strings.ToLower(flag)
		if !containsString(supported, normalized) {
			return nil, tools.InvalidField("flags", supported, normalized)
		}
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out, nil
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

// buildRegexError 从编译错误中提取位置信息。
func buildRegexError(pattern string, err error) *RegexError {
	message := err.Error()
	offset := -1

	// Go 的正则错误形如 `error parsing regexp: ...: <expr>`；
	// 末段的下标就是错误位置。
	if idx := strings.LastIndex(message, ": "); idx >= 0 {
		tail := message[idx+2:]
		if len(tail) > 0 && tail[0] == '^' {
			for pos := 1; pos < len(tail); pos++ {
				if tail[pos] != '^' {
					offset = pos - 1
					break
				}
			}
		}
	}

	position := ""
	if offset >= 0 && offset <= len(pattern) {
		position = offsetToPosition(pattern, offset)
	}

	return &RegexError{
		Code:     "syntax",
		Message:  message,
		Offset:   offset,
		Position: position,
	}
}

// offsetToPosition 把字节偏移转成「第 N 行第 M 列」。
func offsetToPosition(text string, offset int) string {
	line := 1
	column := 1
	for i := 0; i < offset && i < len(text); i++ {
		if text[i] == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return "第 " + itoa(line) + " 行第 " + itoa(column) + " 列"
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

// suggestRegexFix 在表达式非法时给出修复方向。
func suggestRegexFix(pattern string) []RegexSuggestion {
	suggestions := []RegexSuggestion{}

	if strings.Count(pattern, "(") != strings.Count(pattern, ")") {
		suggestions = append(suggestions, RegexSuggestion{
			Code:     "unbalanced-paren",
			Message:  "左右括号数量不匹配，可能缺少右括号",
			Priority: "high",
		})
	}
	if strings.Count(pattern, "[") != strings.Count(pattern, "]") {
		suggestions = append(suggestions, RegexSuggestion{
			Code:     "unbalanced-bracket",
			Message:  "左右方括号数量不匹配，可能缺少右方括号",
			Priority: "high",
		})
	}
	if strings.Count(pattern, "{") != strings.Count(pattern, "}") {
		suggestions = append(suggestions, RegexSuggestion{
			Code:     "unbalanced-brace",
			Message:  "左右花括号数量不匹配，可能缺少右花括号",
			Priority: "high",
		})
	}
	if strings.Contains(pattern, "**") {
		suggestions = append(suggestions, RegexSuggestion{
			Code:     "double-quantifier",
			Message:  "量词不能连续出现，两个星号应写成一个",
			Priority: "high",
		})
	}
	return suggestions
}

// suggestRegexImprovements 在表达式合法时给出改进建议。
func suggestRegexImprovements(pattern string, compiled *regexp.Regexp) []RegexSuggestion {
	suggestions := []RegexSuggestion{}

	if compiled.NumSubexp() == 0 && strings.Contains(pattern, "()") {
		suggestions = append(suggestions, RegexSuggestion{
			Code:     "unused-capture",
			Message:  "包含分组但没有捕获组，考虑改用非捕获分组 (?:…)",
			Priority: "low",
		})
	}
	if strings.Contains(pattern, ".*.*") {
		suggestions = append(suggestions, RegexSuggestion{
			Code:     "greedy-repeat",
			Message:  "连续的 .* 可能引发回溯放大，考虑拆分为更具体的结构",
			Priority: "medium",
		})
	}
	if len(pattern) > 1000 {
		suggestions = append(suggestions, RegexSuggestion{
			Code:     "long-pattern",
			Message:  "表达式较长，匹配开销可能上升，建议拆分并加锚点",
			Priority: "low",
		})
	}
	return suggestions
}

// Validate 做字段级校验。
func (r *RegexValidateRequest) Validate() error {
	if strings.TrimSpace(r.Pattern) == "" {
		return tools.MissingField("pattern")
	}
	if len(r.Pattern) > maxRegexPatternBytes {
		return tools.InvalidField("pattern", []int{0, maxRegexPatternBytes}, len(r.Pattern))
	}
	return nil
}
