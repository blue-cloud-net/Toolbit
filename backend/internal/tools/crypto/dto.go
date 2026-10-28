// Package crypto 实现加密类工具。
//
// 本包**不**实现算法：只做「解码输入 → 校验字段 → 调用 cryptosvc → 编码 DTO」。
//
// 所有密码学运算都经 cryptosvc 转调 Tongsuo；本包**不得**直接 import tongsuo-go，
// 也**不得**用标准库 crypto/* 拼装密码学运算。
package crypto

import (
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
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
func KeyFormats() []string {
	return []string{FormatAuto, FormatPEM, FormatDER, FormatBase64, FormatHex}
}

// ByteInput 是字节输入。
//
// data 始终是字符串，字节序列由 encoding 解释。
type ByteInput struct {
	// Data 是数据文本。
	Data string `json:"data"`
	// Encoding 是 utf8 / hex / base64 / base64url。
	Encoding string `json:"encoding"`
}

// KeyInput 是密钥输入。
type KeyInput struct {
	// Data 是密钥文本。
	Data string `json:"data"`
	// Format 是 auto / pem / der / base64 / hex。
	Format string `json:"format"`
	// Passphrase 是私钥口令；仅对加密私钥有效。
	Passphrase string `json:"passphrase,omitempty"`
}

// Decode 把字节输入解释为字节序列。
func (in ByteInput) Decode() ([]byte, error) {
	if in.Data == "" {
		return nil, tools.MissingField("input.data")
	}
	return codecsvc.Decode(in.Data, in.Encoding)
}

// Decode 把字节输入编码为指定格式的文本。
func (in ByteInput) EncodeTo(enc string) (string, error) {
	raw, err := in.Decode()
	if err != nil {
		return "", err
	}
	return codecsvc.Encode(raw, enc)
}

// Validate 校验字节输入的编码名。
func (in ByteInput) Validate(field string) error {
	_, err := codecsvc.Normalize(in.Encoding)
	if err != nil {
		return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    field + ".encoding",
			"expected": codecsvc.Encodings(),
		})
	}
	return nil
}

// Validate 校验密钥格式名。
func (in KeyInput) Validate(field string) error {
	switch in.Format {
	case "", FormatAuto, FormatPEM, FormatDER, FormatBase64, FormatHex:
		return nil
	default:
		return tools.InvalidField(field+".format", KeyFormats(), in.Format)
	}
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
	if strings.TrimSpace(in.Data) == "" {
		return nil, tools.MissingField("key.data")
	}

	switch in.Format {
	case FormatPEM:
		return []byte(in.Data), nil
	case FormatDER:
		raw, err := codecsvc.Decode(in.Data, codecsvc.EncodingBase64)
		if err != nil {
			return nil, err
		}
		return codecsvc.ToPEM(raw, blockType)
	case FormatBase64:
		raw, err := codecsvc.Decode(in.Data, codecsvc.EncodingBase64)
		if err != nil {
			return nil, err
		}
		return derOrPEM(raw, blockType)
	case FormatHex:
		raw, err := codecsvc.Decode(in.Data, codecsvc.EncodingHex)
		if err != nil {
			return nil, err
		}
		return derOrPEM(raw, blockType)
	default:
		return in.resolveAuto(blockType)
	}
}

// resolveAuto 按 PEM → DER → base64 → hex 的顺序自动识别。
func (in KeyInput) resolveAuto(blockType string) ([]byte, error) {
	// 1. PEM 头。
	if codecsvc.LooksLikePEM(in.Data) {
		return []byte(in.Data), nil
	}

	// 2 & 3. base64 / hex 解码后是否是 DER。
	for _, enc := range []string{codecsvc.EncodingBase64, codecsvc.EncodingHex} {
		raw, err := codecsvc.Decode(in.Data, enc)
		if err != nil {
			continue
		}
		if codecsvc.LooksLikeDER(raw) {
			return codecsvc.ToPEM(raw, blockType)
		}
	}

	return nil, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
		"field":    "key",
		"expected": []string{FormatPEM, FormatDER, FormatBase64, FormatHex},
	})
}

// derOrPEM 把已解码的字节判定为 PEM 文本或 DER 二进制。
func derOrPEM(raw []byte, blockType string) ([]byte, error) {
	if codecsvc.LooksLikePEM(string(raw)) {
		return raw, nil
	}
	return codecsvc.ToPEM(raw, blockType)
}
