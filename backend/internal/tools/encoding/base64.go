// Package encoding 实现编码类工具。
//
// 本包**不**实现算法，只做「解码输入 → 校验字段 → 调用服务 → 编码 DTO」。
package encoding

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Base64 操作类型。
const (
	// Base64Encode 把输入编码为 Base64。
	Base64Encode = "encode"
	// Base64Decode 把 Base64 解码为原始数据。
	Base64Decode = "decode"
)

// Base64 变体。
const (
	// Base64VariantStandard 使用标准字符集（+/）。
	Base64VariantStandard = "standard"
	// Base64VariantURL 使用 URL 安全字符集（-_）。
	Base64VariantURL = "url"
)

// Base64 支持的输入编码。
const (
	// EncodingUTF8 表示输入是 UTF-8 文本。
	EncodingUTF8 = "utf8"
	// EncodingHex 表示输入是十六进制串。
	EncodingHex = "hex"
)

// Base64Request 是 base64 工具的请求。
type Base64Request struct {
	// Operation 是 encode 或 decode。
	Operation string `json:"operation"`
	// Input 是待处理内容。
	Input ByteInput `json:"input"`
	// Variant 是 standard 或 url；默认 standard。
	Variant string `json:"variant"`
	// Padding 是否保留 = 填充；默认 true。
	Padding *bool `json:"padding"`
	// OutputEncoding 是输出编码；解码时默认 utf8。
	OutputEncoding string `json:"outputEncoding"`
}

// ByteInput 是字节输入。
//
// data 始终是字符串，字节序列由 encoding 解释。
type ByteInput struct {
	// Data 是数据文本。
	Data string `json:"data"`
	// Encoding 是 utf8 或 hex。
	Encoding string `json:"encoding"`
}

// Base64Result 是 base64 工具的响应。
type Base64Result struct {
	// Output 是结果文本。
	Output string `json:"output"`
	// Encoding 是输出编码。
	Encoding string `json:"encoding"`
}

// Validate 做字段级校验。
func (r *Base64Request) Validate() error {
	switch r.Operation {
	case Base64Encode, Base64Decode:
	default:
		return tools.InvalidField("operation", []string{Base64Encode, Base64Decode}, r.Operation)
	}

	switch r.Input.Encoding {
	case "", EncodingUTF8, EncodingHex:
	default:
		return tools.InvalidField("input.encoding", []string{EncodingUTF8, EncodingHex}, r.Input.Encoding)
	}

	switch r.Variant {
	case "", Base64VariantStandard, Base64VariantURL:
	default:
		return tools.InvalidField("variant", []string{Base64VariantStandard, Base64VariantURL}, r.Variant)
	}

	switch r.OutputEncoding {
	case "", EncodingUTF8, EncodingHex, "base64", "base64url":
	default:
		return tools.InvalidField("outputEncoding",
			[]string{EncodingUTF8, EncodingHex, "base64", "base64url"}, r.OutputEncoding)
	}

	return nil
}

// Base64 执行 Base64 编解码。
func Base64(_ context.Context, req *Base64Request) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := decodeInput(req.Input)
	if err != nil {
		return tools.Result{}, err
	}

	switch req.Operation {
	case Base64Encode:
		out, enc := encodeBase64(raw, req.Variant, req.Padding)
		return tools.Result{Data: Base64Result{Output: out, Encoding: enc}}, nil
	case Base64Decode:
		out, enc, err := decodeBase64(raw, req.Variant, req.OutputEncoding)
		if err != nil {
			return tools.Result{}, err
		}
		return tools.Result{Data: Base64Result{Output: out, Encoding: enc}}, nil
	default:
		return tools.Result{}, tools.InvalidField("operation",
			[]string{Base64Encode, Base64Decode}, req.Operation)
	}
}

// decodeInput 把请求输入解释为字节序列。
func decodeInput(in ByteInput) ([]byte, error) {
	switch in.Encoding {
	case EncodingHex:
		raw, err := hexDecode(in.Data)
		if err != nil {
			return nil, err
		}
		return raw, nil
	default:
		return []byte(in.Data), nil
	}
}

// encodeBase64 把字节序列编码为 Base64。
func encodeBase64(raw []byte, variant string, padding *bool) (string, string) {
	enc := base64.StdEncoding
	name := "base64"
	if variant == Base64VariantURL {
		enc = base64.URLEncoding
		name = "base64url"
	}
	if padding != nil && !*padding {
		enc = enc.WithPadding(base64.NoPadding)
	}
	return enc.EncodeToString(raw), name
}

// decodeBase64 把 Base64 文本解码为指定编码的输出。
//
// 解码时**忽略**换行、空白与填充：'=' 一律去掉后用无填充编码解析，
// 因此有填充与无填充的输入都能正确解码。
func decodeBase64(raw []byte, variant string, outputEncoding string) (string, string, error) {
	text := stripWhitespace(string(raw))
	text = strings.TrimRight(text, "=")

	enc := base64.RawStdEncoding
	if variant == Base64VariantURL {
		enc = base64.RawURLEncoding
	}

	decoded, err := enc.DecodeString(text)
	if err != nil {
		return "", "", toolerr.Newf(toolerr.CodeInvalidBase64, map[string]any{
			"reason": "invalid base64 input",
		})
	}

	switch outputEncoding {
	case EncodingHex:
		return hexEncode(decoded), EncodingHex, nil
	case "base64":
		return base64.StdEncoding.EncodeToString(decoded), "base64", nil
	case "base64url":
		return base64.URLEncoding.EncodeToString(decoded), "base64url", nil
	default:
		return string(decoded), EncodingUTF8, nil
	}
}

// stripWhitespace 去除空白字符。
func stripWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '\r':
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
