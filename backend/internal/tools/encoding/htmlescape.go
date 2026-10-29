package encoding

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// HTML 转义模式。
const (
	// HTMLEscapeEncode 是转义为实体。
	HTMLEscapeEncode = "encode"
	// HTMLEscapeDecode 是把实体还原为字符。
	HTMLEscapeDecode = "decode"
)

// HTMLEscapeMode 返回受支持的模式（顺序稳定）。
func HTMLEscapeMode() []string {
	return []string{HTMLEscapeEncode, HTMLEscapeDecode}
}

// HTMLEscapeRequest 是 html-escape 工具的请求。
type HTMLEscapeRequest struct {
	// Mode 是模式：encode / decode；默认 encode。
	Mode string `json:"mode"`
	// Input 是待处理文本。
	Input string `json:"input"`
}

// HTMLEscapeResult 是 html-escape 工具的响应。
type HTMLEscapeResult struct {
	// Mode 是实际使用的模式。
	Mode string `json:"mode"`
	// Output 是处理结果。
	Output string `json:"output"`
	// ChangedCount 是被替换的字符数。
	ChangedCount int `json:"changedCount"`
}

// 需要转义的 HTML 实体。
var htmlEntities = []struct {
	char   rune
	entity string
}{
	{'&', "&amp;"},
	{'<', "&lt;"},
	{'>', "&gt;"},
	{'"', "&quot;"},
	{'\'', "&#39;"},
}

// maxHTMLEscapeBytes 是输入的字节上限。
const maxHTMLEscapeBytes = 1024 * 1024

// HTMLEscape 在文本与 HTML 实体之间转换。
//
// 这是**纯编码转换**：不涉及密码学，也不联网。
func HTMLEscape(_ context.Context, req *HTMLEscapeRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = HTMLEscapeEncode
	}

	var output string
	var changed int
	if mode == HTMLEscapeDecode {
		output, changed = decodeHTMLEntities(req.Input)
	} else {
		output, changed = encodeHTMLEntities(req.Input)
	}

	return tools.Result{Data: HTMLEscapeResult{
		Mode:         mode,
		Output:       output,
		ChangedCount: changed,
	}}, nil
}

// encodeHTMLEntities 把特殊字符转成实体。
func encodeHTMLEntities(input string) (string, int) {
	var builder strings.Builder
	changed := 0
	for _, r := range input {
		replaced := false
		for _, entity := range htmlEntities {
			if r == entity.char {
				builder.WriteString(entity.entity)
				changed++
				replaced = true
				break
			}
		}
		if !replaced {
			builder.WriteRune(r)
		}
	}
	return builder.String(), changed
}

// decodeHTMLEntities 把实体还原成字符。
//
// 支持命名实体与数字实体（十进制与十六进制）。
func decodeHTMLEntities(input string) (string, int) {
	output := input
	changed := 0

	// 命名实体：&amp; 最后替换，避免二次转义。
	for _, entity := range []struct {
		name string
		char string
	}{
		{"&quot;", "\""},
		{"&#39;", "'"},
		{"&apos;", "'"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&nbsp;", " "},
		{"&amp;", "&"},
	} {
		count := strings.Count(output, entity.name)
		if count > 0 {
			output = strings.ReplaceAll(output, entity.name, entity.char)
			changed += count
		}
	}

	// 数字实体：&#NN; 与 &#xHH;。
	for {
		start := strings.Index(output, "&#")
		if start < 0 {
			break
		}
		end := strings.Index(output[start:], ";")
		if end < 0 {
			break
		}
		end += start
		body := output[start+2 : end]
		code, ok := parseEntityCode(body)
		if !ok {
			// 不是合法实体，跳过这一处继续找下一个。
			next := strings.Index(output[start+2:], "&#")
			if next < 0 {
				break
			}
			output = output[:start+2] + output[start+2+next:]
			continue
		}
		output = output[:start] + string(rune(code)) + output[end+1:]
		changed++
	}

	return output, changed
}

// parseEntityCode 解析数字实体的码点。
func parseEntityCode(body string) (int, bool) {
	if body == "" {
		return 0, false
	}
	if body[0] == 'x' || body[0] == 'X' {
		if len(body) == 1 {
			return 0, false
		}
		code := 0
		for _, c := range body[1:] {
			digit := hexDigitValue(c)
			if digit < 0 {
				return 0, false
			}
			code = code*16 + digit
			if code > 0x10FFFF {
				return 0, false
			}
		}
		return code, true
	}
	code := 0
	for _, c := range body {
		if c < '0' || c > '9' {
			return 0, false
		}
		code = code*10 + int(c-'0')
		if code > 0x10FFFF {
			return 0, false
		}
	}
	return code, true
}

// hexDigitValue 返回十六进制数字的值。
func hexDigitValue(c rune) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}

// Validate 做字段级校验。
func (r *HTMLEscapeRequest) Validate() error {
	mode := r.Mode
	if mode == "" {
		mode = HTMLEscapeEncode
	}
	if !containsString(HTMLEscapeMode(), mode) {
		return tools.InvalidField("mode", HTMLEscapeMode(), mode)
	}
	if r.Input == "" {
		return tools.MissingField("input")
	}
	if len(r.Input) > maxHTMLEscapeBytes {
		return tools.InvalidField("input", []int{0, maxHTMLEscapeBytes}, len(r.Input))
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
