package cryptosvc

import (
	"encoding/asn1"
	"math/big"
	"strings"

	"github.com/blue-cloud-net/tongsuo-go/asym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 签名算法名。
const (
	// SignSM2 是 SM2withSM3（GB/T 32918.2），摘要固定为 SM3。
	SignSM2 = "SM2"
	// SignRSA 是 RSA（PKCS#1 v1.5 / PSS）。
	SignRSA = "RSA"
	// SignECDSA 是 ECDSA，摘要固定为 SHA-256。
	SignECDSA = "ECDSA"
	// SignED25519 是 Ed25519（RFC 8032），纯签名。
	SignED25519 = "ED25519"
	// SignED448 是 Ed448（RFC 8032），纯签名。
	SignED448 = "ED448"
)

// 签名摘要名。
const (
	// HashSHA256 是 SHA-256。
	HashSHA256 = "SHA256"
	// HashSHA384 是 SHA-384。
	HashSHA384 = "SHA384"
	// HashSHA512 是 SHA-512。
	HashSHA512 = "SHA512"
)

// RSA 填充方式。
const (
	// RSAPaddingPSS 是 RSA-PSS（默认）。
	RSAPaddingPSS = "pss"
	// RSAPaddingPKCS1 是 RSA-PKCS#1 v1.5。
	RSAPaddingPKCS1 = "pkcs1"
)

// Ed25519SigSize 是 Ed25519 签名长度（字节）。
const Ed25519SigSize = 64

// Ed448SigSize 是 Ed448 签名长度（字节）。
const Ed448SigSize = 114

// SignAlgorithms 返回受支持的签名算法名（顺序稳定）。
func SignAlgorithms() []string {
	return []string{SignSM2, SignRSA, SignECDSA, SignED25519, SignED448}
}

// SignHashes 返回某签名算法支持的摘要名。
//
// 规则：
//   - SM2 摘要固定为 SM3，**不**接受 hash 参数，返回 nil；
//   - ECDSA 摘要固定为 SHA-256，返回 nil；
//   - Ed25519 / Ed448 是纯签名，**禁止**预哈希，返回 nil；
//   - RSA 支持 SHA-256 / SHA-384 / SHA-512。
//
// 返回 nil 表示该算法不接受摘要参数。
func SignHashes(algorithm string) []string {
	switch algorithm {
	case SignRSA:
		return []string{HashSHA256, HashSHA384, HashSHA512}
	default:
		return nil
	}
}

// SignPaddings 返回某签名算法支持的填充方式。
//
// 目前只有 RSA 有填充概念；其余算法返回 nil。
func SignPaddings(algorithm string) []string {
	switch algorithm {
	case SignRSA:
		return []string{RSAPaddingPSS, RSAPaddingPKCS1}
	default:
		return nil
	}
}

// DefaultSM2UserID 是 SM2 的默认用户标识符（GM/T 0003-2012 规定值）。
const DefaultSM2UserID = "1234567812345678"

// Sign 对 data 签名。
//
// hash 为空时按算法默认处理：
//   - SM2 → SM3（固定）；
//   - ECDSA → SHA-256（固定）；
//   - RSA → SHA-256。
//
// padding 仅对 RSA 有效，缺省为 PSS。
//
// Ed25519 / Ed448 遵循 RFC 8032 的**纯签名**语义：消息原样参与签名，
// 调用方**不得**预先哈希（否则会破坏与其它实现的互通）。
//
// sm2UserID 为空时使用 DefaultSM2UserID。
func Sign(algorithm, hash, padding string, priv *PrivateKey, data []byte, sm2UserID string) ([]byte, error) {
	if priv == nil || priv.key == nil {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "key"})
	}
	digestName, err := resolveSignDigest(algorithm, hash)
	if err != nil {
		return nil, err
	}
	pad, err := resolveSignPadding(algorithm, padding)
	if err != nil {
		return nil, err
	}

	switch algorithm {
	case SignSM2:
		id := sm2UserID
		if id == "" {
			id = DefaultSM2UserID
		}
		sig, err := asym.SignWithID(priv.key, data, []byte(id))
		if err != nil {
			return nil, MapError(err)
		}
		return sig, nil
	case SignRSA:
		if pad == RSAPaddingPSS {
			// PSS 的 saltLen 缺省取摘要长度（asym.PSSSaltLenDigest）。
			sig, err := asym.SignPSS(priv.key, data, asym.PSSSaltLenDigest, strings.ToLower(digestName))
			if err != nil {
				return nil, MapError(err)
			}
			return sig, nil
		}
		sig, err := asym.SignPKCS1v15(priv.key, data, strings.ToLower(digestName))
		if err != nil {
			return nil, MapError(err)
		}
		return sig, nil
	case SignECDSA:
		sig, err := asym.SignECDSA(priv.key, data)
		if err != nil {
			return nil, MapError(err)
		}
		return sig, nil
	case SignED25519:
		sig, err := asym.SignEd25519(priv.key, data)
		if err != nil {
			return nil, MapError(err)
		}
		return sig, nil
	case SignED448:
		sig, err := asym.SignEd448(priv.key, data)
		if err != nil {
			return nil, MapError(err)
		}
		return sig, nil
	default:
		return nil, unsupportedAlgorithm(algorithm, SignAlgorithms())
	}
}

// Verify 验签。
//
// 返回 (true, nil) 表示验签通过；
// 返回 (false, nil) 表示**验签不通过**——这是**业务结论**，
// 不是错误（见 docs/standards/03-crypto.md §6）；
// 返回 err 表示签名**结构非法**（无法解析为 DER / 长度不符）。
//
// ⚠️ 底层库对「验签不通过」与「签名结构非法」可能给出相近的原生错误，
// 因此本函数**先做结构校验**再验签：结构不合法直接返回 SIGNATURE_INVALID，
// 结构合法但验签失败一律视为业务结论。
func Verify(algorithm, hash, padding string, pub *PublicKey, data, sig []byte, sm2UserID string) (bool, error) {
	if pub == nil || pub.key == nil {
		return false, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "publicKey"})
	}
	digestName, err := resolveSignDigest(algorithm, hash)
	if err != nil {
		return false, err
	}
	pad, err := resolveSignPadding(algorithm, padding)
	if err != nil {
		return false, err
	}
	if err := checkSignatureStructure(algorithm, pad, pub, sig); err != nil {
		return false, err
	}

	var verifyErr error
	switch algorithm {
	case SignSM2:
		id := sm2UserID
		if id == "" {
			id = DefaultSM2UserID
		}
		verifyErr = asym.VerifyWithID(pub.key, data, sig, []byte(id))
	case SignRSA:
		if pad == RSAPaddingPSS {
			verifyErr = asym.VerifyPSS(pub.key, data, sig, asym.PSSSaltLenDigest, strings.ToLower(digestName))
		} else {
			verifyErr = asym.VerifyPKCS1v15(pub.key, data, sig, strings.ToLower(digestName))
		}
	case SignECDSA:
		verifyErr = asym.VerifyECDSA(pub.key, data, sig)
	case SignED25519:
		verifyErr = asym.VerifyEd25519(pub.key, data, sig)
	case SignED448:
		verifyErr = asym.VerifyEd448(pub.key, data, sig)
	default:
		return false, unsupportedAlgorithm(algorithm, SignAlgorithms())
	}

	if verifyErr != nil {
		// 结构已确认合法，验签失败是**业务结论**。
		return false, nil
	}
	return true, nil
}

// checkSignatureStructure 校验签名的结构是否合法。
//
// 结构规则（与底层库的实际输出**逐字一致**）：
//   - SM2 / ECDSA：ASN.1 DER 的 SEQUENCE { INTEGER, INTEGER }；
//   - RSA（PSS 与 PKCS#1 v1.5 均是）：**裸值**，长度等于模数字节数；
//   - Ed25519：64 字节；Ed448：114 字节。
//
// 违反规则返回 SIGNATURE_INVALID。
func checkSignatureStructure(algorithm, padding string, pub *PublicKey, sig []byte) error {
	switch algorithm {
	case SignED25519:
		if len(sig) != Ed25519SigSize {
			return toolerr.Newf(toolerr.CodeSignatureInvalid, map[string]any{
				"expected": Ed25519SigSize,
				"actual":   len(sig),
			})
		}
		return nil
	case SignED448:
		if len(sig) != Ed448SigSize {
			return toolerr.Newf(toolerr.CodeSignatureInvalid, map[string]any{
				"expected": Ed448SigSize,
				"actual":   len(sig),
			})
		}
		return nil
	case SignRSA:
		// RSA 签名是**裸值**而非 DER：长度必须等于模数字节数。
		params, err := asym.Params(pub.key)
		if err != nil || params.N == nil {
			return nil
		}
		want := (params.N.BitLen() + 7) / 8
		if len(sig) != want {
			return toolerr.Newf(toolerr.CodeSignatureInvalid, map[string]any{
				"expected": want,
				"actual":   len(sig),
			})
		}
		return nil
	}

	// SM2 / ECDSA：DER 签名 SEQUENCE { INTEGER, INTEGER }。
	var parsed struct {
		R *big.Int
		S *big.Int
	}
	rest, err := asn1.Unmarshal(sig, &parsed)
	if err != nil || len(rest) != 0 || parsed.R == nil || parsed.S == nil {
		return toolerr.Newf(toolerr.CodeSignatureInvalid, map[string]any{
			"expected": "der",
			"actual":   len(sig),
		})
	}
	return nil
}

// resolveSignDigest 解析签名摘要。
//
// 算法不接受摘要参数（SM2 / ECDSA / Ed 系列）而调用方传了 hash 时，
// 返回 UNSUPPORTED_ALGORITHM；hash 为空时返回该算法的默认摘要。
func resolveSignDigest(algorithm, hash string) (string, error) {
	supported := SignHashes(algorithm)
	if len(supported) == 0 {
		if hash != "" {
			return "", unsupportedAlgorithm(algorithm+"/"+hash, []string{algorithm})
		}
		switch algorithm {
		case SignSM2, SignECDSA, SignED25519, SignED448:
			return "", nil
		default:
			return "", unsupportedAlgorithm(algorithm, SignAlgorithms())
		}
	}
	if hash == "" {
		return supported[0], nil
	}
	for _, h := range supported {
		if h == hash {
			return h, nil
		}
	}
	return "", unsupportedAlgorithm(algorithm+"/"+hash, supported)
}

// resolveSignPadding 解析签名填充方式。
//
// 只有 RSA 有填充概念；其余算法传了 padding 时返回 UNSUPPORTED_ALGORITHM。
func resolveSignPadding(algorithm, padding string) (string, error) {
	supported := SignPaddings(algorithm)
	if len(supported) == 0 {
		if padding != "" {
			return "", unsupportedAlgorithm(algorithm+"/"+padding, []string{algorithm})
		}
		return "", nil
	}
	if padding == "" {
		return supported[0], nil
	}
	for _, p := range supported {
		if p == padding {
			return p, nil
		}
	}
	return "", unsupportedAlgorithm(algorithm+"/"+padding, supported)
}
