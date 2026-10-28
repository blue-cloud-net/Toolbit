package formatting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// JSONFormatRequest 是 json-format 工具的请求。
type JSONFormatRequest struct {
	// Input 是待处理的 JSON 文本。
	Input TextInput `json:"input"`
	// Action 是 format、minify 或 validate。
	Action string `json:"action"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
	// SortKeys 是否按键排序。
	SortKeys bool `json:"sortKeys"`
	// EscapeUnicode 是否转义非 ASCII 字符。
	EscapeUnicode bool `json:"escapeUnicode"`
}

// JSONFormatResult 是 json-format 工具的响应。
type JSONFormatResult struct {
	// Output 是处理结果；validate 时为空串。
	Output string `json:"output"`
	// Valid 是内容是否合法。
	Valid bool `json:"valid"`
	// Error 是语法错误的位置与原因；合法时为 nil。
	Error *SyntaxError `json:"error,omitempty"`
}

// JSONFormat 美化、压缩或校验 JSON。
func JSONFormat(req *JSONFormatRequest) (JSONFormatResult, error) {
	if err := req.Validate(); err != nil {
		return JSONFormatResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return JSONFormatResult{}, err
	}

	value, synErr := parseJSON(text)
	if synErr != nil {
		return JSONFormatResult{
			Output: "",
			Valid:  false,
			Error:  synErr,
		}, nil
	}

	if action(req.Action) == ActionValidate {
		return JSONFormatResult{Valid: true}, nil
	}

	var out string
	switch action(req.Action) {
	case ActionMinify:
		out, err = minifyJSON(value, req.EscapeUnicode)
	default:
		out, err = formatJSON(value, indentWidth(req.Indent), req.SortKeys, req.EscapeUnicode)
	}
	if err != nil {
		return JSONFormatResult{}, err
	}

	return JSONFormatResult{Output: out, Valid: true}, nil
}

// Validate 做字段级校验。
func (r *JSONFormatRequest) Validate() error {
	if err := validateAction(r.Action); err != nil {
		return err
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	return r.Input.Validate("input")
}

// jsonMember 是 JSON 对象的一个成员（**有序**）。
type jsonMember struct {
	// Key 是键名。
	Key string
	// Value 是值。
	Value any
}

// jsonObject 是保序的 JSON 对象。
type jsonObject struct {
	// members 是成员列表，按原文顺序。
	members []jsonMember
	// index 是键到成员的索引，用于查重。
	index map[string]int
}

// jsonArray 是 JSON 数组。
type jsonArray struct {
	// items 是元素列表。
	items []any
}

// parseJSON 解析 JSON，同时检测重复键。
//
// 标准库**允许**重复键（后者覆盖前者），但对工具使用者来说
// 重复键几乎总是拼写错误，因此显式检出并报错。
//
// 解析走 `json.Decoder.Token()` 流式接口而不是 `Unmarshal`：
// 后者把对象解成 `map[string]any`，**键序会被打乱**，
// 而键序是用户数据，格式化工具不能把它弄丢。
func parseJSON(text string) (any, *SyntaxError) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()

	value, err := decodeJSONValue(dec)
	if err != nil {
		se := jsonSyntaxError(text, err)
		return nil, &se
	}

	// 拒绝尾随内容：请求体只允许一个 JSON 值。
	if _, err := dec.Token(); err != io.EOF {
		se := jsonSyntaxError(text, fmt.Errorf("trailing content"))
		return nil, &se
	}

	return value, nil
}

// decodeJSONValue 递归解析 JSON 值，并检查对象内的重复键。
func decodeJSONValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return decodeJSONObject(dec)
		case '[':
			return decodeJSONArray(dec)
		default:
			return nil, fmt.Errorf("unexpected delimiter")
		}
	default:
		return tok, nil
	}
}

// decodeJSONObject 解析 JSON 对象，保持键序。
func decodeJSONObject(dec *json.Decoder) (*jsonObject, error) {
	out := &jsonObject{index: make(map[string]int, 8)}

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("object key is not a string")
		}
		if _, dup := out.index[key]; dup {
			return nil, fmt.Errorf("duplicate key %q", key)
		}

		value, err := decodeJSONValue(dec)
		if err != nil {
			return nil, err
		}
		out.index[key] = len(out.members)
		out.members = append(out.members, jsonMember{Key: key, Value: value})
	}

	// 消费对象结束符。
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return out, nil
}

// decodeJSONArray 解析 JSON 数组。
func decodeJSONArray(dec *json.Decoder) (*jsonArray, error) {
	out := &jsonArray{}

	for dec.More() {
		value, err := decodeJSONValue(dec)
		if err != nil {
			return nil, err
		}
		out.items = append(out.items, value)
	}

	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return out, nil
}

// formatJSON 美化 JSON。
func formatJSON(value any, indent int, sortKeys bool, escapeUnicode bool) (string, error) {
	var buf bytes.Buffer
	if err := writeFormattedJSON(&buf, value, indent, escapeUnicode, 0, sortKeys); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// minifyJSON 压缩 JSON。
func minifyJSON(value any, escapeUnicode bool) (string, error) {
	var buf bytes.Buffer
	if err := writeFormattedJSON(&buf, value, 0, escapeUnicode, 0, false); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// writeFormattedJSON 把值写入缓冲区。
//
// sortKeys 为真时对象按键排序；否则保持原文顺序。
func writeFormattedJSON(buf *bytes.Buffer, value any, indent int, escapeUnicode bool, depth int, sortKeys bool) error {
	pad := indentString(indent)
	inner := strings.Repeat(pad, depth+1)
	outer := strings.Repeat(pad, depth)

	switch v := value.(type) {
	case *jsonObject:
		if len(v.members) == 0 {
			buf.WriteString("{}")
			return nil
		}
		members := v.members
		if sortKeys {
			members = sortedMembers(members)
		}

		buf.WriteString("{")
		if indent > 0 {
			buf.WriteString("\n")
		}
		for i, m := range members {
			buf.WriteString(inner)
			buf.WriteString(jsonString(m.Key, escapeUnicode))
			buf.WriteString(":")
			if indent > 0 {
				buf.WriteString(" ")
			}
			if err := writeFormattedJSON(buf, m.Value, indent, escapeUnicode, depth+1, sortKeys); err != nil {
				return err
			}
			if i < len(members)-1 {
				buf.WriteString(",")
			}
			if indent > 0 {
				buf.WriteString("\n")
			}
		}
		buf.WriteString(outer)
		buf.WriteString("}")

	case *jsonArray:
		if len(v.items) == 0 {
			buf.WriteString("[]")
			return nil
		}
		buf.WriteString("[")
		if indent > 0 {
			buf.WriteString("\n")
		}
		for i, item := range v.items {
			buf.WriteString(inner)
			if err := writeFormattedJSON(buf, item, indent, escapeUnicode, depth+1, sortKeys); err != nil {
				return err
			}
			if i < len(v.items)-1 {
				buf.WriteString(",")
			}
			if indent > 0 {
				buf.WriteString("\n")
			}
		}
		buf.WriteString(outer)
		buf.WriteString("]")

	default:
		encoded, err := marshalScalar(value, escapeUnicode)
		if err != nil {
			return err
		}
		buf.WriteString(encoded)
	}
	return nil
}

// sortedMembers 返回按键排序的成员副本（原切片不变）。
func sortedMembers(members []jsonMember) []jsonMember {
	out := make([]jsonMember, len(members))
	copy(out, members)

	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Key < out[j-1].Key; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// marshalScalar 编码标量值。
func marshalScalar(value any, escapeUnicode bool) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(!escapeUnicode)
	if err := enc.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}

// jsonString 把字符串编码为 JSON 字面量。
func jsonString(s string, escapeUnicode bool) string {
	encoded, err := marshalScalar(s, escapeUnicode)
	if err != nil {
		return `""`
	}
	return encoded
}

// jsonSyntaxError 把 JSON 解析错误转成带行列的结构化错误。
func jsonSyntaxError(source string, err error) SyntaxError {
	se := SyntaxError{Message: "invalid JSON"}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		se.Line, se.Column = positionOf(source, int(syntaxErr.Offset))
		return se
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Offset > 0 {
			se.Line, se.Column = positionOf(source, int(typeErr.Offset))
		}
		se.Message = "invalid JSON value type"
		return se
	}

	// 重复键等自定义错误：消息里带字段名，**不含**输入内容。
	if msg := err.Error(); strings.HasPrefix(msg, "duplicate key ") {
		se.Message = "duplicate object key"
	}
	return se
}
