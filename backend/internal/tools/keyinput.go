package tools

import (
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 密钥输入格式。
const (
	// FormatAuto 自动识别：PEM → DER → base64 → hex。
	FormatAuto = "auto"
	// FormatPEM 是 PEM 文本。
	FormatPEM = "pem"
	// FormatDER 是 DER 二进制的 base64 表示。
	FormatDER = "der"
	// FormatBase64 是 base64 文本。
	FormatBase64 = "base64"
	// FormatHex 是十六进制文本。
	FormatHex = "hex"
)

// KeyFormats 返回受支持的密钥格式名（顺序稳定）。
//
// 该列表是**传输层**的取值域（用户怎么把密钥贴进来），
// 与 pkisvc 的**编码层**（pkcs1 / pkcs8 / spki）不是一回事。
func KeyFormats() []string {
	return []string{FormatAuto, FormatPEM, FormatDER, FormatBase64, FormatHex}
}

// KeyInput 是密钥输入。
//
// 本类型是**跨分类共享**的：`crypto`（加密类）与 `keys`（密钥类）
// 两个工具包都用类型别名复用它，避免每个包各写一套解析逻辑。
type KeyInput struct {
	// Data 是密钥文本。
	Data string `json:"data"`
	// Format 是 auto / pem / der / base64 / hex。
	Format string `json:"format"`
	// Passphrase 是私钥口令；仅对加密私钥有效。
	Passphrase string `json:"passphrase,omitempty"`
}

// Validate 校验密钥格式名。
func (in KeyInput) Validate(field string) error {
	switch in.Format {
	case "", FormatAuto, FormatPEM, FormatDER, FormatBase64, FormatHex:
		return nil
	default:
		return InvalidField(field+".format", KeyFormats(), in.Format)
	}
}

// Resolve 把密钥输入解析为原始字节，并报告它是 PEM 文本还是 DER 二进制。
//
// format 为 auto 时按 **PEM → base64 → hex** 的顺序识别：
//
//   - 文本形如 PEM 块 → 直接作为 PEM 字节；
//   - base64 / hex 解码后是 DER → 作为 DER 字节；
//   - base64 / hex 解码后是 PEM 文本 → 作为 PEM 字节。
//
// 全部失败返回 INVALID_PEM。
//
// 返回的字节是**新分配**的切片，调用方负责在使用后清零。
func (in KeyInput) Resolve() (data []byte, isPEM bool, err error) {
	if strings.TrimSpace(in.Data) == "" {
		return nil, false, MissingField("key.data")
	}

	switch in.Format {
	case FormatPEM:
		return []byte(in.Data), true, nil
	case FormatDER:
		raw, err := codecsvc.Decode(in.Data, codecsvc.EncodingBase64)
		if err != nil {
			return nil, false, err
		}
		return raw, false, nil
	case FormatBase64:
		raw, err := codecsvc.Decode(in.Data, codecsvc.EncodingBase64)
		if err != nil {
			return nil, false, err
		}
		return classifyKeyBytes(raw)
	case FormatHex:
		raw, err := codecsvc.Decode(in.Data, codecsvc.EncodingHex)
		if err != nil {
			return nil, false, err
		}
		return classifyKeyBytes(raw)
	default:
		if codecsvc.LooksLikePEM(in.Data) {
			return []byte(in.Data), true, nil
		}
		for _, enc := range []string{codecsvc.EncodingBase64, codecsvc.EncodingHex} {
			raw, err := codecsvc.Decode(in.Data, enc)
			if err != nil {
				continue
			}
			data, isPEM, err := classifyKeyBytes(raw)
			if err == nil {
				return data, isPEM, nil
			}
		}
		return nil, false, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field":    "key.data",
			"expected": KeyFormats(),
		})
	}
}

// classifyKeyBytes 判定解码后的字节是 PEM 文本还是 DER 二进制。
func classifyKeyBytes(raw []byte) ([]byte, bool, error) {
	if codecsvc.LooksLikeDER(raw) {
		return raw, false, nil
	}
	if codecsvc.LooksLikePEM(string(raw)) {
		return raw, true, nil
	}
	return nil, false, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
		"field":    "key.data",
		"expected": KeyFormats(),
	})
}

// PrivateKeyPEM 解析私钥输入，返回 PEM 字节。
//
// format 为 auto 时按 **PEM → DER → base64 → hex** 的顺序识别；
// 全部失败返回 INVALID_PEM。
//
// 返回的字节是**新分配**的切片，调用方负责在使用后清零。
func (in KeyInput) PrivateKeyPEM() ([]byte, error) {
	return in.resolvePEM(codecsvc.BlockPrivateKey)
}

// PublicKeyPEM 解析公钥输入，返回 PEM 字节。
func (in KeyInput) PublicKeyPEM() ([]byte, error) {
	return in.resolvePEM(codecsvc.BlockPublicKey)
}

// resolvePEM 解析密钥输入为 PEM 字节。
//
// blockType 是 DER 形式无法判定块类型时使用的默认块类型。
func (in KeyInput) resolvePEM(blockType string) ([]byte, error) {
	data, isPEM, err := in.Resolve()
	if err != nil {
		return nil, err
	}
	if isPEM {
		return data, nil
	}
	return codecsvc.ToPEM(data, blockType)
}
