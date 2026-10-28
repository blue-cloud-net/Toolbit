// Package codecsvc 提供编码归一化与 PEM/DER 识别。
//
// 本包只做**表示层**的转换（utf8 / hex / base64 / base64url 与 PEM/DER 容器），
// **不涉及任何密码学运算**：算法一律由 cryptosvc 转调 Tongsuo。
//
// 允许依赖：标准库。
package codecsvc

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 编码标识。
const (
	// EncodingUTF8 表示 UTF-8 文本。
	EncodingUTF8 = "utf8"
	// EncodingHex 表示十六进制串。
	EncodingHex = "hex"
	// EncodingBase64 表示标准 Base64。
	EncodingBase64 = "base64"
	// EncodingBase64URL 表示 URL 安全 Base64。
	EncodingBase64URL = "base64url"
)

// Encodings 返回受支持的编码名（顺序稳定）。
func Encodings() []string {
	return []string{EncodingUTF8, EncodingHex, EncodingBase64, EncodingBase64URL}
}

// Normalize 把编码名归一化。
//
// 空串按 utf8 处理；大小写不敏感；未登记的编码返回 INVALID_ENCODING。
func Normalize(enc string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(enc)) {
	case "", EncodingUTF8:
		return EncodingUTF8, nil
	case EncodingHex:
		return EncodingHex, nil
	case EncodingBase64:
		return EncodingBase64, nil
	case EncodingBase64URL:
		return EncodingBase64URL, nil
	default:
		return "", toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
			"encoding": enc,
			"supported": Encodings(),
		})
	}
}

// Decode 把文本按编码解释为字节。
//
// base64 解码**忽略**空白与换行，并兼容有无填充；
// hex 解码忽略空白。
func Decode(s string, enc string) ([]byte, error) {
	normalized, err := Normalize(enc)
	if err != nil {
		return nil, err
	}
	switch normalized {
	case EncodingUTF8:
		return []byte(s), nil
	case EncodingHex:
		raw, err := hex.DecodeString(stripSpace(s))
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeInvalidHex, map[string]any{
				"reason": "invalid hex input",
			})
		}
		return raw, nil
	case EncodingBase64:
		raw, err := decodeBase64(s, base64.StdEncoding)
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeInvalidBase64, map[string]any{
				"reason": "invalid base64 input",
			})
		}
		return raw, nil
	case EncodingBase64URL:
		raw, err := decodeBase64(s, base64.URLEncoding)
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeInvalidBase64, map[string]any{
				"reason": "invalid base64 input",
			})
		}
		return raw, nil
	default:
		return nil, toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
			"encoding": enc,
			"supported": Encodings(),
		})
	}
}

// Encode 把字节编码为文本。
func Encode(b []byte, enc string) (string, error) {
	normalized, err := Normalize(enc)
	if err != nil {
		return "", err
	}
	switch normalized {
	case EncodingUTF8:
		return string(b), nil
	case EncodingHex:
		return hex.EncodeToString(b), nil
	case EncodingBase64:
		return base64.StdEncoding.EncodeToString(b), nil
	case EncodingBase64URL:
		return base64.URLEncoding.EncodeToString(b), nil
	default:
		return "", toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
			"encoding": enc,
			"supported": Encodings(),
		})
	}
}

// decodeBase64 解码 base64，忽略空白与填充差异。
func decodeBase64(s string, enc *base64.Encoding) ([]byte, error) {
	cleaned := stripSpace(s)
	cleaned = strings.TrimRight(cleaned, "=")
	raw, err := enc.WithPadding(base64.NoPadding).DecodeString(cleaned)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// stripSpace 去除空白字符。
func stripSpace(s string) string {
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

// PEM 块类型常量。
const (
	// BlockPrivateKey 是 PKCS#8 私钥。
	BlockPrivateKey = "PRIVATE KEY"
	// BlockEncryptedPrivateKey 是加密的 PKCS#8 私钥。
	BlockEncryptedPrivateKey = "ENCRYPTED PRIVATE KEY"
	// BlockPublicKey 是 SPKI 公钥。
	BlockPublicKey = "PUBLIC KEY"
	// BlockCertificate 是 X.509 证书。
	BlockCertificate = "CERTIFICATE"
)

// LooksLikePEM 判断文本是否是 PEM 块。
func LooksLikePEM(s string) bool {
	return strings.Contains(s, "-----BEGIN ")
}

// ToPEM 把 DER 字节包装为 PEM。
func ToPEM(der []byte, blockType string) ([]byte, error) {
	if len(der) == 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidDER, map[string]any{
			"reason": "empty der",
		})
	}
	return pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), nil
}

// FromPEM 取出 PEM 块的内容与类型。
func FromPEM(pemBytes []byte) ([]byte, string, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, "", toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"reason": "no pem block",
		})
	}
	return block.Bytes, block.Type, nil
}

// LooksLikeDER 判断字节是否像 DER 编码的 ASN.1 结构。
//
// 判据是首字节为 0x30（SEQUENCE）——这是启发式判断，
// 只用于 `format: auto` 的**候选筛选**，最终仍由解析器判定合法性。
func LooksLikeDER(b []byte) bool {
	return len(b) >= 2 && b[0] == 0x30
}
