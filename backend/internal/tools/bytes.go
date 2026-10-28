package tools

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// Bytes 是字节输入。
//
// data 始终是字符串，字节序列由 encoding 解释。
//
// 本类型是**跨分类共享**的：各工具包用类型别名复用它
// （`encoding.ByteInput`、`crypto.ByteInput` 等），
// 避免每个包各写一套结构体与解码逻辑。
type Bytes struct {
	// Data 是数据文本。
	Data string `json:"data"`
	// Encoding 是 utf8 / hex / base64 / base64url。
	Encoding string `json:"encoding"`
}

// Decode 把字节输入解释为字节序列。
//
// 空内容视为缺失字段：工具输入**不**接受「有编码而无内容」。
func (b Bytes) Decode() ([]byte, error) {
	if b.Data == "" {
		return nil, MissingField("input.data")
	}
	return codecsvc.Decode(b.Data, b.Encoding)
}

// EncodeTo 把字节输入编码为指定格式的文本。
func (b Bytes) EncodeTo(enc string) (string, error) {
	raw, err := b.Decode()
	if err != nil {
		return "", err
	}
	return codecsvc.Encode(raw, enc)
}

// Validate 校验字节输入的编码名。
//
// field 是调用方给该输入起的字段路径（如 "input"），
// 错误里回显的是**字段名**与**期望取值**，不回显用户输入。
func (b Bytes) Validate(field string) error {
	_, err := codecsvc.Normalize(b.Encoding)
	if err != nil {
		return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    field + ".encoding",
			"expected": codecsvc.Encodings(),
		})
	}
	return nil
}
