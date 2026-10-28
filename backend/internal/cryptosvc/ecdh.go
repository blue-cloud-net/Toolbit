package cryptosvc

import (
	"github.com/blue-cloud-net/tongsuo-go/ecdh"
	"github.com/blue-cloud-net/tongsuo-go/kdf"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// ECDH 曲线名。
const (
	// CurveP256 是 NIST P-256。
	CurveP256 = "P-256"
	// CurveP384 是 NIST P-384。
	CurveP384 = "P-384"
	// CurveP521 是 NIST P-521。
	CurveP521 = "P-521"
	// CurveX25519 是 X25519（RFC 7748）。
	CurveX25519 = "X25519"
	// CurveX448 是 X448（RFC 7748）。
	CurveX448 = "X448"
)

// MaxDerivedKeyBytes 是 KDF 派生输出的长度上限（字节）。
const MaxDerivedKeyBytes = 1 << 20

// ECDHCurves 返回受支持的曲线名（顺序稳定）。
//
// ⚠️ **不含 SM2**：SM2 的曲线不是 ECDH 可协商曲线，
// 上游 ecdh 包会返回 ErrUnsupportedKey。
func ECDHCurves() []string {
	return []string{CurveP256, CurveP384, CurveP521, CurveX25519, CurveX448}
}

// ECDHKDFHashes 返回受支持的 KDF 摘要名（顺序稳定）。
func ECDHKDFHashes() []string {
	return []string{HashSHA256, HashSHA384, HashSHA512, "SM3"}
}

// ECDHSharedSecret 计算 ECDH 共享密钥。
//
// privPEM 与 pubPEM 必须是**未加密**的 PEM 私钥 / 公钥。
// 输出是敏感数据，调用方使用后**必须**清零。
func ECDHSharedSecret(curve string, privPEM, pubPEM []byte) ([]byte, error) {
	c, err := ecdh.CurveByName(curve)
	if err != nil {
		return nil, unsupportedAlgorithm(curve, ECDHCurves())
	}

	priv, err := ecdh.LoadPrivateKeyPEM(c, privPEM)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidKey, err)
	}
	peer, err := ecdh.LoadPublicKeyPEM(c, pubPEM)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidKey, err)
	}

	secret, err := priv.ECDH(peer)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeCryptoOperationFailed, err)
	}
	return secret, nil
}

// ECDHDerive 用 HKDF 从共享密钥派生新密钥。
//
// length 不得超过 MaxDerivedKeyBytes；超限返回 TOO_LARGE。
// 输出是敏感数据，调用方使用后**必须**清零。
func ECDHDerive(hash string, secret, salt, info []byte, length int) ([]byte, error) {
	if length <= 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "kdf.length",
			"expected": "positive",
		})
	}
	if length > MaxDerivedKeyBytes {
		return nil, toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"maxBytes": MaxDerivedKeyBytes,
			"actual":   length,
		})
	}
	if len(secret) == 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "sharedSecret",
			"expected": "non-empty",
		})
	}

	md, ok := kdfHash(hash)
	if !ok {
		return nil, unsupportedAlgorithm(hash, ECDHKDFHashes())
	}

	out, err := kdf.HKDF(md, secret, salt, info, length)
	if err != nil {
		return nil, MapError(err)
	}
	return out, nil
}

// kdfHash 把摘要名映射为 kdf 包的常量。
func kdfHash(hash string) (kdf.Hash, bool) {
	switch hash {
	case HashSHA256:
		return kdf.HashSHA256, true
	case HashSHA384:
		return kdf.HashSHA384, true
	case HashSHA512:
		return kdf.HashSHA512, true
	case "SM3":
		return kdf.HashSM3, true
	default:
		return "", false
	}
}
