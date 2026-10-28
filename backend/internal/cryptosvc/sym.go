package cryptosvc

import (
	"fmt"

	"github.com/blue-cloud-net/tongsuo-go/sym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 对称算法名。
const (
	// SymSM4 是 SM4（GB/T 32907），密钥 16 字节。
	SymSM4 = "SM4"
	// SymAES 是 AES，密钥 16 / 24 / 32 字节。
	SymAES = "AES"
)

// 对称加密模式。
const (
	// ModeECB 是电子密码本模式（**无扩散**，仅用于兼容既有数据）。
	ModeECB = "ECB"
	// ModeCBC 是密码分组链接模式。
	ModeCBC = "CBC"
	// ModeCTR 是计数器模式。
	ModeCTR = "CTR"
	// ModeOFB 是输出反馈模式。
	ModeOFB = "OFB"
	// ModeCFB 是密文反馈模式。
	ModeCFB = "CFB"
	// ModeGCM 是带认证的 Galois/Counter 模式（AEAD）。
	ModeGCM = "GCM"
)

// 填充方式。
const (
	// PaddingPKCS7 是 PKCS#7 填充（默认）。
	PaddingPKCS7 = "pkcs7"
	// PaddingZero 是零填充。
	PaddingZero = "zero"
	// PaddingNone 是不填充。
	PaddingNone = "none"
)

// GCMNonceSize 是 GCM 推荐 nonce 长度（字节）。
const GCMNonceSize = 12

// SymAlgorithms 返回受支持的对称算法名（顺序稳定）。
func SymAlgorithms() []string {
	return []string{SymSM4, SymAES}
}

// SymModes 返回某算法支持的模式列表。
//
// 算法不受支持时返回 nil。
func SymModes(algorithm string) []string {
	switch algorithm {
	case SymSM4:
		return []string{ModeECB, ModeCBC, ModeCTR, ModeOFB, ModeCFB, ModeGCM}
	case SymAES:
		return []string{ModeECB, ModeCBC, ModeCTR, ModeGCM}
	default:
		return nil
	}
}

// SymPaddings 返回某算法与模式组合支持的填充方式。
//
// ⚠️ 这是本层**自行维护**的支持矩阵：tongsuo-go 的 by-name 入口
// 会对 AES **静默忽略** padding 参数（恒用 pkcs7），
// 若不显式拦截，`padding: none` 的 AES 密文将无法按预期解密。
//
// 规则：
//   - GCM 是 AEAD，没有填充概念，只接受 pkcs7；
//   - SM4 的 ECB / CBC 支持 pkcs7 / zero / none；
//   - 其余组合（SM4 流式模式、AES 全部模式）只支持 pkcs7。
func SymPaddings(algorithm, mode string) []string {
	switch {
	case mode == ModeGCM:
		return []string{PaddingPKCS7}
	case algorithm == SymSM4 && (mode == ModeECB || mode == ModeCBC):
		return []string{PaddingPKCS7, PaddingZero, PaddingNone}
	default:
		return []string{PaddingPKCS7}
	}
}

// SymEncrypt 加密。
//
// mode 为 GCM 时返回**分离**的认证标签；其余模式 tag 为 nil。
// aad 仅对 GCM 有效。
//
// padding 省略时按 pkcs7 处理。
func SymEncrypt(algorithm, mode, padding string, key, iv, plaintext, aad []byte) ([]byte, []byte, error) {
	name, err := symName(algorithm, mode, len(key))
	if err != nil {
		return nil, nil, err
	}
	if err := checkSymPadding(algorithm, mode, padding); err != nil {
		return nil, nil, err
	}
	if err := checkSymIV(mode, iv); err != nil {
		return nil, nil, err
	}
	if padding == PaddingNone {
		if err := checkBlockAligned(plaintext); err != nil {
			return nil, nil, err
		}
	}

	if mode == ModeGCM {
		ct, tag, err := sym.EncryptAEAD(name, key, iv, plaintext, aad)
		if err != nil {
			return nil, nil, MapError(err)
		}
		return ct, tag, nil
	}

	ct, err := sym.Encrypt(name, key, iv, plaintext, &sym.Options{Padding: padding})
	if err != nil {
		return nil, nil, MapError(err)
	}
	return ct, nil, nil
}

// SymDecrypt 解密。
//
// mode 为 GCM 时 tag **必填**；缺失返回 TAG_REQUIRED，长度不符返回 TAG_INVALID，
// 认证失败返回 TAG_INVALID。
func SymDecrypt(algorithm, mode, padding string, key, iv, ciphertext, tag, aad []byte) ([]byte, error) {
	name, err := symName(algorithm, mode, len(key))
	if err != nil {
		return nil, err
	}
	if err := checkSymPadding(algorithm, mode, padding); err != nil {
		return nil, err
	}
	if err := checkSymIV(mode, iv); err != nil {
		return nil, err
	}
	if padding == PaddingNone {
		if err := checkBlockAligned(ciphertext); err != nil {
			return nil, err
		}
	}

	if mode == ModeGCM {
		if len(tag) == 0 {
			return nil, toolerr.Newf(toolerr.CodeTagRequired, map[string]any{
				"expected": sym.TagSize,
			})
		}
		if len(tag) != sym.TagSize {
			return nil, toolerr.Newf(toolerr.CodeTagInvalid, map[string]any{
				"expected": sym.TagSize,
				"actual":   len(tag),
			})
		}
		pt, err := sym.DecryptAEAD(name, key, iv, ciphertext, tag, aad)
		if err != nil {
			return nil, toolerr.Wrap(toolerr.CodeTagInvalid, err)
		}
		return pt, nil
	}

	pt, err := sym.Decrypt(name, key, iv, ciphertext, &sym.Options{Padding: padding})
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeDecryptFailed, err)
	}
	return pt, nil
}

// symName 构造 tongsuo-go 的「算法-模式」名，并校验密钥长度。
//
// AES 的具体位数（128 / 192 / 256）由**密钥长度**决定。
func symName(algorithm, mode string, keyLen int) (string, error) {
	modes := SymModes(algorithm)
	if modes == nil {
		return "", unsupportedAlgorithm(algorithm, SymAlgorithms())
	}
	known := false
	for _, m := range modes {
		if m == mode {
			known = true
			break
		}
	}
	if !known {
		return "", unsupportedMode(mode, modes)
	}

	switch algorithm {
	case SymSM4:
		if keyLen != sym.SM4KeySize {
			return "", toolerr.Newf(toolerr.CodeKeySizeMismatch, map[string]any{
				"algorithm": algorithm,
				"expected":  sym.SM4KeySize,
				"actual":    keyLen,
			})
		}
		return "SM4-" + mode, nil
	case SymAES:
		var bits int
		switch keyLen {
		case sym.AES128KeySize:
			bits = 128
		case sym.AES192KeySize:
			bits = 192
		case sym.AES256KeySize:
			bits = 256
		default:
			return "", toolerr.Newf(toolerr.CodeKeySizeMismatch, map[string]any{
				"algorithm": algorithm,
				"expected":  []int{sym.AES128KeySize, sym.AES192KeySize, sym.AES256KeySize},
				"actual":    keyLen,
			})
		}
		return fmt.Sprintf("AES-%d-%s", bits, mode), nil
	default:
		return "", unsupportedAlgorithm(algorithm, SymAlgorithms())
	}
}

// checkSymPadding 校验填充方式在该算法与模式下是否受支持。
func checkSymPadding(algorithm, mode, padding string) error {
	if padding == "" {
		padding = PaddingPKCS7
	}
	supported := SymPaddings(algorithm, mode)
	for _, p := range supported {
		if p == padding {
			return nil
		}
	}
	return unsupportedPadding(padding, supported)
}

// checkSymIV 校验 IV 长度。
//
// ECB 不需要 IV；GCM 需要 12 字节 nonce；其余模式需要 16 字节 IV。
func checkSymIV(mode string, iv []byte) error {
	want, required := symIVSize(mode)
	if !required {
		return nil
	}
	if len(iv) != want {
		return toolerr.Newf(toolerr.CodeIVSizeMismatch, map[string]any{
			"mode":     mode,
			"expected": want,
			"actual":   len(iv),
		})
	}
	return nil
}

// symIVSize 返回该模式要求的 IV 长度，以及是否需要 IV。
func symIVSize(mode string) (int, bool) {
	switch mode {
	case ModeECB:
		return 0, false
	case ModeGCM:
		return sym.NonceSize, true
	default:
		return sym.BlockSize, true
	}
}

// checkBlockAligned 校验数据长度是分组长度的整数倍。
//
// `padding: none` 时必须自行校验（见 docs/standards/03-crypto.md §7）。
func checkBlockAligned(data []byte) error {
	if len(data)%sym.BlockSize != 0 {
		return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "input",
			"expected": fmt.Sprintf("multiple of %d bytes", sym.BlockSize),
			"actual":   len(data),
		})
	}
	return nil
}
