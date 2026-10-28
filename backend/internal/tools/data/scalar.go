package data

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Scalar 是既可写作字符串又可写作数字的标量。
//
// 某些字段（如时间戳）两种写法都常见，工具不应因为
// 客户端传了数字就报错。
type Scalar struct {
	// Value 是文本形式；未给出时为空。
	Value string
	// Set 报告该字段是否出现过。
	Set bool
}

// UnmarshalJSON 解析字符串或数字。
func (s *Scalar) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*s = Scalar{}
		return nil
	}

	s.Set = true

	if trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return err
		}
		s.Value = text
		return nil
	}

	// 数字：保留原始字面量，避免浮点精度损失。
	s.Value = string(trimmed)
	return nil
}

// MarshalJSON 始终输出字符串。
func (s Scalar) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}

// String 返回文本值。
func (s Scalar) String() string {
	return s.Value
}

// Int64 把值解析为整数。
//
// 非整数或超出范围时返回 false。
func (s Scalar) Int64() (int64, bool) {
	if s.Value == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(s.Value, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
