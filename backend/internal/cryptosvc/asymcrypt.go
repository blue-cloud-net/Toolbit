package cryptosvc

import (
	"github.com/blue-cloud-net/tongsuo-go/asym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 非对称加密算法名。
const (
	// AsymRSA 是 RSA。
	AsymRSA = "RSA"
	// AsymSM2 是 SM2（GB/T 32918）。
	AsymSM2 = "SM2"
)

// SM2 密文顺序。
const (
	// SM2OrderC1C3C2 是 C1‖C3‖C2 顺序（GM/T 0009-2012 之后的标准顺序）。
	SM2OrderC1C3C2 = "c1c3c2"
	// SM2OrderC1C2C3 是 C1‖C2‖C3 顺序。
	SM2OrderC1C2C3 = "c1c2c3"
)

// AsymAlgorithms 返回受支持的非对称加密算法名（顺序稳定）。
func AsymAlgorithms() []string {
	return []string{AsymRSA, AsymSM2}
}

// SM2Orders 返回受支持的 SM2 密文顺序（顺序稳定）。
func SM2Orders() []string {
	return []string{SM2OrderC1C3C2, SM2OrderC1C2C3}
}

// AsymEncrypt 用公钥加密。
//
// SM2 的输出是**裸格式**且 C1 为未压缩点（含 `04` 前缀），
// 与 BouncyCastle 等主流实现兼容。
// RSA 的 padding 只能是 oaep（默认）或 pkcs1。
//
// rsaHash 为空时按 SHA-256 处理。
func AsymEncrypt(algorithm, padding, rsaHash string, pub *PublicKey, data []byte) ([]byte, error) {
	if pub == nil || pub.key == nil {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "key"})
	}
	switch algorithm {
	case AsymSM2:
		ct, err := asym.EncryptWithOrder(pub.key, data, SM2OrderC1C3C2)
		if err != nil {
			return nil, MapError(err)
		}
		return ct, nil
	case AsymRSA:
		if padding == "" {
			padding = "oaep"
		}
		switch padding {
		case "oaep":
			digest := rsaDigest(rsaHash)
			if err := checkRSAPlaintextLength(pub.key, digest, len(data)); err != nil {
				return nil, err
			}
			ct, err := asym.EncryptOAEP(pub.key, data, digest)
			if err != nil {
				return nil, MapError(err)
			}
			return ct, nil
		case "pkcs1":
			digest := rsaDigest(rsaHash)
			if err := checkRSAPlaintextLength(pub.key, digest, len(data)); err != nil {
				return nil, err
			}
			ct, err := asym.EncryptPKCS1v15(pub.key, data)
			if err != nil {
				return nil, MapError(err)
			}
			return ct, nil
		default:
			return nil, unsupportedPadding(padding, []string{"oaep", "pkcs1"})
		}
	default:
		return nil, unsupportedAlgorithm(algorithm, AsymAlgorithms())
	}
}

// AsymDecrypt 用私钥解密。
//
// SM2 的输入**自动识别**：
//   - ASN.1 DER（首字节 0x30）；
//   - 裸格式 C1‖C3‖C2（`04` 前缀）；
//   - 裸格式 C1‖C2‖C3（`04` 前缀）。
//
// ⚠️ 两种裸格式**长度相同**，无法静态区分，因此按默认顺序先试、
// 失败再试另一种；两者都失败时统一返回 DECRYPT_FAILED
// （不区分「顺序不对」与「私钥不匹配」，避免形成解密预言机）。
//
// 压缩点（`02` / `03` 前缀）当前**不受支持**，返回 UNSUPPORTED_FORMAT。
//
// RSA 的 padding 与加密侧必须一致。
func AsymDecrypt(algorithm, padding, rsaHash string, priv *PrivateKey, data []byte) ([]byte, error) {
	if priv == nil || priv.key == nil {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "key"})
	}
	switch algorithm {
	case AsymSM2:
		return decryptSM2(priv.key, data)
	case AsymRSA:
		if padding == "" {
			padding = "oaep"
		}
		switch padding {
		case "oaep":
			pt, err := asym.DecryptOAEP(priv.key, data, rsaDigest(rsaHash))
			if err != nil {
				return nil, toolerr.Wrap(toolerr.CodeDecryptFailed, err)
			}
			return pt, nil
		case "pkcs1":
			pt, err := asym.DecryptPKCS1v15(priv.key, data)
			if err != nil {
				return nil, toolerr.Wrap(toolerr.CodeDecryptFailed, err)
			}
			return pt, nil
		default:
			return nil, unsupportedPadding(padding, []string{"oaep", "pkcs1"})
		}
	default:
		return nil, unsupportedAlgorithm(algorithm, AsymAlgorithms())
	}
}

// decryptSM2 解密 SM2 密文，格式自动识别。
func decryptSM2(priv asym.PrivateKey, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "input",
			"expected": "non-empty",
		})
	}

	switch data[0] {
	case 0x30:
		// ASN.1 DER。
		pt, err := asym.Decrypt(priv, data)
		if err != nil {
			return nil, toolerr.Wrap(toolerr.CodeDecryptFailed, err)
		}
		return pt, nil
	case 0x04:
		// 裸格式、未压缩点：两种顺序长度相同，依次尝试。
		pt, err := asym.DecryptWithOrder(priv, data, SM2OrderC1C3C2)
		if err == nil {
			return pt, nil
		}
		pt, err2 := asym.DecryptWithOrder(priv, data, SM2OrderC1C2C3)
		if err2 == nil {
			return pt, nil
		}
		return nil, toolerr.Wrap(toolerr.CodeDecryptFailed, err2)
	case 0x02, 0x03:
		// 压缩点：上游的 SM2 格式转换会丢失点坐标，无法正确解密。
		return nil, toolerr.Newf(toolerr.CodeUnsupportedFormat, map[string]any{
			"format":    "sm2-compressed-point",
			"supported": []string{"der", "c1c3c2", "c1c2c3"},
		})
	default:
		return nil, toolerr.Newf(toolerr.CodeUnsupportedFormat, map[string]any{
			"format":    "unknown",
			"supported": []string{"der", "c1c3c2", "c1c2c3"},
		})
	}
}

// rsaDigest 返回 OAEP 使用的摘要名（小写）。
func rsaDigest(hash string) string {
	switch hash {
	case HashSHA384:
		return "sha384"
	case HashSHA512:
		return "sha512"
	default:
		return "sha256"
	}
}

// checkRSAPlaintextLength 校验 RSA 明文长度。
//
// RSA 只能加密极少量数据：OAEP 的最大明文长度是 k − 2·hLen − 2；
// PKCS#1 v1.5 是 k − 11。超限返回 PLAINTEXT_TOO_LONG。
func checkRSAPlaintextLength(pub asym.PublicKey, digest string, plainLen int) error {
	params, err := asym.Params(pub)
	if err != nil || params.N == nil {
		// 拿不到模数时交由底层库报错。
		return nil
	}
	k := (params.N.BitLen() + 7) / 8
	hLen := 32
	if digest == "sha384" {
		hLen = 48
	} else if digest == "sha512" {
		hLen = 64
	}
	max := k - 2*hLen - 2
	if plainLen > max {
		return toolerr.Newf(toolerr.CodePlaintextTooLong, map[string]any{
			"maxBytes": max,
			"actual":   plainLen,
		})
	}
	return nil
}
