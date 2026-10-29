package pkisvc

import (
	"github.com/blue-cloud-net/tongsuo-go/asym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 私钥加密算法标识。
const (
	// PBEAES256CBC 是 AES-256-CBC + PBKDF2（默认）。
	PBEAES256CBC = "AES-256-CBC"
	// PBEDESEDE3CBC 是 DES-EDE3-CBC + PBKDF2。
	//
	// ⚠️ DES-EDE3 属**遗留算法**，仅用于兼容既有数据。
	PBEDESEDE3CBC = "DES-EDE3-CBC"
)

// PBECiphers 返回受支持的私钥加密算法（顺序稳定）。
func PBECiphers() []string {
	return []string{PBEAES256CBC, PBEDESEDE3CBC}
}

// pbeToUpstream 把对外的加密算法名映射为底层库的通用名。
func pbeToUpstream(cipher string) (string, bool) {
	switch cipher {
	case "", PBEAES256CBC:
		return "aes-256-cbc", true
	case PBEDESEDE3CBC:
		return "des-ede3-cbc", true
	default:
		return "", false
	}
}

// EncryptPrivateKey 用口令加密私钥，导出 PKCS#8 加密 PEM。
//
// cipher 取 PBECiphers 之一；为空时用 AES-256-CBC。
// 输入必须是**未加密**的私钥 PEM。
func EncryptPrivateKey(pemBytes []byte, passphrase, cipher string) ([]byte, error) {
	if passphrase == "" {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{
			"field": "passphrase",
		})
	}
	upstream, ok := pbeToUpstream(cipher)
	if !ok {
		return nil, unsupportedFormat(cipher, PBECiphers())
	}

	priv, err := LoadPrivateKey(pemBytes, "")
	if err != nil {
		return nil, err
	}
	defer func() { _ = priv.Close() }()

	out, err := asym.MarshalEncryptedPrivateKeyPEMWithCipher(priv.key, upstream, passphrase)
	if err != nil {
		return nil, MapError(err)
	}
	return out, nil
}

// DecryptPrivateKey 解密加密私钥，导出 PKCS#8 明文 PEM。
//
// 口令错误返回 PASSPHRASE_INCORRECT。
func DecryptPrivateKey(pemBytes []byte, passphrase string) ([]byte, error) {
	if passphrase == "" {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{
			"field": "passphrase",
		})
	}

	priv, err := LoadPrivateKey(pemBytes, passphrase)
	if err != nil {
		return nil, err
	}
	defer func() { _ = priv.Close() }()

	out, err := priv.MarshalPKCS8PEM()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChangePassphrase 修改加密私钥的口令。
//
// oldPass 必须正确；否则返回 PASSPHRASE_INCORRECT。
func ChangePassphrase(pemBytes []byte, oldPass, newPass string) ([]byte, error) {
	if oldPass == "" {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{
			"field": "passphrase",
		})
	}
	if newPass == "" {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{
			"field": "newPassphrase",
		})
	}

	out, err := asym.ChangePassword(pemBytes, oldPass, newPass)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodePassphraseIncorrect, err)
	}
	return out, nil
}

// ExtractPublicKey 从私钥导出配对的公钥（SPKI PEM）。
//
// 该操作**不需要**口令：口令只保护私钥材料，不影响公钥导出。
func ExtractPublicKey(pemBytes []byte) ([]byte, error) {
	priv, err := LoadPrivateKey(pemBytes, "")
	if err != nil {
		return nil, err
	}
	defer func() { _ = priv.Close() }()

	out, err := priv.MarshalPublicKeyPEM()
	if err != nil {
		return nil, err
	}
	return out, nil
}
