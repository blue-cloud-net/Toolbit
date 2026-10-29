package tools

import (
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// PemBundle 是 PEM 文本 / DER 二进制输入。
//
// 本类型是**跨分类共享**的：证书类工具（certificates）与密钥类工具
// （keys）都要读证书 / 证书签名请求 / 吊销列表，因此共用同一套
// 「传输层解码」逻辑，避免第三份实现。
//
// 与 KeyInput 的差别只在于**语义**：KeyInput 面向密钥，
// PemBundle 面向证书族材料（允许一个输入里含多个 PEM 块）。
type PemBundle struct {
	// Data 是材料文本。
	Data string `json:"data"`
	// Format 是 auto / pem / der / base64 / hex。
	Format string `json:"format"`
}

// Validate 校验输入格式名。
func (in PemBundle) Validate(field string) error {
	switch in.Format {
	case "", FormatAuto, FormatPEM, FormatDER, FormatBase64, FormatHex:
		return nil
	default:
		return InvalidField(field+".format", KeyFormats(), in.Format)
	}
}

// Resolve 把输入解析为原始字节，并报告它是 PEM 文本还是 DER 二进制。
//
// format 为 auto 时按 **PEM → base64 → hex** 的顺序识别；
// 全部失败返回 INVALID_PEM。
//
// 返回的字节是**新分配**的切片。
func (in PemBundle) Resolve() (data []byte, isPEM bool, err error) {
	if strings.TrimSpace(in.Data) == "" {
		return nil, false, MissingField("input.data")
	}
	return resolveInput(in.Data, in.Format, "input.data")
}

// resolveInput 按格式名解码输入，是 KeyInput 与 PemBundle 的共同实现。
//
// field 是调用方给该输入起的字段路径，用于错误回显**字段名**（不回显内容）。
func resolveInput(data string, format string, field string) ([]byte, bool, error) {
	switch format {
	case FormatPEM:
		return []byte(data), true, nil
	case FormatDER:
		raw, err := codecsvc.Decode(data, codecsvc.EncodingBase64)
		if err != nil {
			return nil, false, err
		}
		return raw, false, nil
	case FormatBase64:
		raw, err := codecsvc.Decode(data, codecsvc.EncodingBase64)
		if err != nil {
			return nil, false, err
		}
		return classifyInputBytes(raw, field)
	case FormatHex:
		raw, err := codecsvc.Decode(data, codecsvc.EncodingHex)
		if err != nil {
			return nil, false, err
		}
		return classifyInputBytes(raw, field)
	default:
		if codecsvc.LooksLikePEM(data) {
			return []byte(data), true, nil
		}
		for _, enc := range []string{codecsvc.EncodingBase64, codecsvc.EncodingHex} {
			raw, err := codecsvc.Decode(data, enc)
			if err != nil {
				continue
			}
			decoded, isPEM, err := classifyInputBytes(raw, field)
			if err == nil {
				return decoded, isPEM, nil
			}
		}
		return nil, false, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field":    field,
			"expected": KeyFormats(),
		})
	}
}

// classifyInputBytes 判定解码后的字节是 PEM 文本还是 DER 二进制。
func classifyInputBytes(raw []byte, field string) ([]byte, bool, error) {
	if codecsvc.LooksLikeDER(raw) {
		return raw, false, nil
	}
	if codecsvc.LooksLikePEM(string(raw)) {
		return raw, true, nil
	}
	return nil, false, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
		"field":    field,
		"expected": KeyFormats(),
	})
}
