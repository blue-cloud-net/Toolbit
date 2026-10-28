package encoding

import (
	"context"
	"net/url"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// URL 编解码操作。
const (
	// URLEncode 把输入编码为百分号编码。
	URLEncode = "encode"
	// URLDecode 把输入解码为原文。
	URLDecode = "decode"
	// URLParse 解析 URL 的各组成部分。
	URLParse = "parse"
)

// URL 组件模式。
const (
	// URLComponentFull 整体编码：保留 URI 保留字符。
	URLComponentFull = "full"
	// URLComponentPath 按路径段编码。
	URLComponentPath = "path"
	// URLComponentQuery 按查询值编码。
	URLComponentQuery = "query"
	// URLComponentFragment 按片段编码。
	URLComponentFragment = "fragment"
)

// URLCodecRequest 是 url-codec 工具的请求。
type URLCodecRequest struct {
	// Operation 是 encode、decode 或 parse。
	Operation string `json:"operation"`
	// Input 是待处理的 URL 文本。
	Input tools.Bytes `json:"input"`
	// Component 是组件模式；默认 full。
	Component string `json:"component"`
	// PlusAsSpace 是否把 `+` 视为空格；默认 true，仅 decode 时生效。
	PlusAsSpace *bool `json:"plusAsSpace"`
}

// URLParam 是一个查询参数。
type URLParam struct {
	// Key 是参数名。
	Key string `json:"key"`
	// Value 是参数值。
	Value string `json:"value"`
}

// URLParts 是 URL 的组成部分。
type URLParts struct {
	// Scheme 是协议名，不含 `://`。
	Scheme string `json:"scheme"`
	// Host 是主机名或 IP。
	Host string `json:"host"`
	// Port 是端口号；未显式指定时为空串。
	Port string `json:"port"`
	// Path 是路径部分。
	Path string `json:"path"`
	// Query 是查询串原文，不含 `?`。
	Query string `json:"query"`
	// Params 是查询参数列表；同名参数保留多个条目。
	Params []URLParam `json:"params"`
}

// URLCodecResult 是 url-codec 工具的响应。
type URLCodecResult struct {
	// Output 是结果文本。
	Output string `json:"output"`
	// Parsed 是 parse 时的结构化结果；其余操作为 nil。
	Parsed *URLParts `json:"parsed,omitempty"`
}

// URLCodec 执行 URL 编解码与解析。
func URLCodec(_ context.Context, req *URLCodecRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := req.Input.Decode()
	if err != nil {
		return tools.Result{}, err
	}
	text := string(raw)

	switch req.Operation {
	case URLEncode:
		out := encodeComponent(text, req.Component)
		return tools.Result{Data: URLCodecResult{Output: out}}, nil

	case URLDecode:
		out, err := decodeComponent(text, req.plusAsSpace())
		if err != nil {
			return tools.Result{}, err
		}
		return tools.Result{Data: URLCodecResult{Output: out}}, nil

	case URLParse:
		parts, err := parseURL(text)
		if err != nil {
			return tools.Result{}, err
		}
		return tools.Result{Data: URLCodecResult{Output: text, Parsed: parts}}, nil

	default:
		return tools.Result{}, tools.InvalidField("operation",
			[]string{URLEncode, URLDecode, URLParse}, req.Operation)
	}
}

// plusAsSpace 返回 `+` 是否视为空格；缺省为 true。
func (r *URLCodecRequest) plusAsSpace() bool {
	if r.PlusAsSpace == nil {
		return true
	}
	return *r.PlusAsSpace
}

// Validate 做字段级校验。
func (r *URLCodecRequest) Validate() error {
	switch r.Operation {
	case URLEncode, URLDecode, URLParse:
	default:
		return tools.InvalidField("operation",
			[]string{URLEncode, URLDecode, URLParse}, r.Operation)
	}

	switch r.Component {
	case "", URLComponentFull, URLComponentPath, URLComponentQuery, URLComponentFragment:
	default:
		return tools.InvalidField("component", []string{
			URLComponentFull, URLComponentPath, URLComponentQuery, URLComponentFragment,
		}, r.Component)
	}

	return r.Input.Validate("input")
}

// encodeComponent 按组件语义做百分号编码。
//
// - `full`：保留 URI 保留字符（`:` `/` `?` `#` 等），只转义非法字符；
// - `path`：按路径段编码，`/` 被转义；
// - `query`：按查询值编码，空格转成 `+`；
// - `fragment`：按片段编码，`#` 与 `?` 被转义。
func encodeComponent(s string, component string) string {
	switch component {
	case URLComponentPath:
		return url.PathEscape(s)
	case URLComponentQuery:
		return url.QueryEscape(s)
	case URLComponentFragment:
		return fragmentEscape(s)
	case URLComponentFull, "":
		return fullEscape(s)
	default:
		return fullEscape(s)
	}
}

// decodeComponent 做百分号解码。
//
// plusAsSpace 为真时，`+` 先被替换成空格再解码；
// 为假时 `+` 是普通字符，原样保留。
func decodeComponent(s string, plusAsSpace bool) (string, error) {
	if plusAsSpace {
		s = strings.ReplaceAll(s, "+", " ")
	}
	out, err := url.PathUnescape(s)
	if err != nil {
		return "", toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
			"reason": "invalid percent encoding",
		})
	}
	return out, nil
}

// fullEscape 按「整体 URL」语义编码。
//
// 保留字符集与 `encodeURI` 一致：A-Z a-z 0-9 与
// `- _ . ! ~ * ' ( ) ; , / ? : @ & = + $ #`。
func fullEscape(s string) string {
	const safe = "-_.!~*'();,/?:@&=+#$"
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isURIUnreserved(c) || strings.IndexByte(safe, c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteString(percentEncodeByte(c))
	}
	return b.String()
}

// fragmentEscape 按「片段」语义编码。
//
// 片段允许 `/` 与 `?`，但 `&`、`=`、`#` 会被转义。
func fragmentEscape(s string) string {
	const safe = "-_.!~*'()/?"
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isURIUnreserved(c) || strings.IndexByte(safe, c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteString(percentEncodeByte(c))
	}
	return b.String()
}

// isURIUnreserved 报告字节是否为 URI 的「非保留字符」。
func isURIUnreserved(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	case c == '-' || c == '_' || c == '.' || c == '~':
		return true
	default:
		return false
	}
}

// percentEncodeByte 把单字节转成 `%XX`（大写十六进制）。
func percentEncodeByte(c byte) string {
	const hexDigits = "0123456789ABCDEF"
	return string([]byte{
		'%',
		hexDigits[c>>4],
		hexDigits[c&0x0F],
	})
}

// parseURL 解析 URL 的组成部分。
func parseURL(raw string) (*URLParts, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "input.data",
			"reason": "not a valid URL",
		})
	}

	parts := &URLParts{
		Scheme: u.Scheme,
		Host:   u.Hostname(),
		Port:   u.Port(),
		Path:   u.Path,
		Query:  u.RawQuery,
		Params: []URLParam{},
	}

	parts.Params = parseQueryParams(u.RawQuery)
	return parts, nil
}

// parseQueryParams 解析查询串。
//
// 同名参数**保留多个条目**，顺序与原文一致。
func parseQueryParams(raw string) []URLParam {
	params := []URLParam{}
	if raw == "" {
		return params
	}

	for _, pair := range strings.Split(raw, "&") {
		if pair == "" {
			continue
		}
		key, value, _ := strings.Cut(pair, "=")
		// 解码失败时保留原文：查询串本身已经被 url.Parse 接受，
		// 个别百分号序列非法不应让整个解析失败。
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			decodedKey = key
		}
		decodedValue, err := url.QueryUnescape(value)
		if err != nil {
			decodedValue = value
		}
		params = append(params, URLParam{Key: decodedKey, Value: decodedValue})
	}
	return params
}
