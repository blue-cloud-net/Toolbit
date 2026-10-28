package cryptosvc

import (
	"encoding/pem"

	"github.com/blue-cloud-net/tongsuo-go/asym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// PrivateKey 是非对称私钥句柄。
//
// 它是**本层**对底层密钥对象的封装：tools 层**不**直接接触 asym 的类型，
// 从而保证「只有 cryptosvc 触碰 Tongsuo」这条不变量。
//
// 私钥持有原生句柄，**必须**显式释放（见 docs/standards/03-crypto.md §3）。
type PrivateKey struct {
	// key 是底层私钥；Close 之后置为 nil。
	key asym.PrivateKey
	// source 是加载时的 PEM 源缓冲；加载成功后立即清零。
	source []byte
}

// PublicKey 是非对称公钥句柄。
type PublicKey struct {
	// key 是底层公钥。
	key asym.PublicKey
}

// LoadPrivateKey 从 PEM 加载私钥。
//
// passphrase 非空时按**加密私钥**处理（AES-256-CBC + PBKDF2 的 PKCS#8）。
// passphrase 为空但 PEM 是加密块时返回 PASSPHRASE_REQUIRED。
//
// 加载成功后**立即**清零 PEM 源缓冲。
func LoadPrivateKey(pemBytes []byte, passphrase string) (*PrivateKey, error) {
	if len(pemBytes) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "key"})
	}

	var (
		key asym.PrivateKey
		err error
	)
	if passphrase != "" {
		key, err = asym.LoadEncryptedPrivateKeyPEM(pemBytes, passphrase)
		if err != nil {
			return nil, toolerr.Wrap(toolerr.CodePassphraseIncorrect, err)
		}
	} else {
		key, err = asym.LoadPrivateKeyPEM(pemBytes)
		if err != nil {
			// 加密私钥在缺少口令时也走这条路。
			if IsEncryptedPrivateKey(pemBytes) {
				return nil, toolerr.Newf(toolerr.CodePassphraseRequired, map[string]any{
					"field": "key.passphrase",
				})
			}
			return nil, toolerr.Wrap(toolerr.CodeInvalidKey, err)
		}
	}

	// 源缓冲用后立即清零。
	Zero(pemBytes)

	return &PrivateKey{key: key}, nil
}

// LoadPublicKey 从 PEM 加载公钥（SubjectPublicKeyInfo）。
func LoadPublicKey(pemBytes []byte) (*PublicKey, error) {
	if len(pemBytes) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "key"})
	}
	key, err := asym.LoadPublicKeyPEM(pemBytes)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidKey, err)
	}
	Zero(pemBytes)
	return &PublicKey{key: key}, nil
}

// Close 释放私钥持有的原生句柄。
//
// **幂等**：重复调用只返回首次的结果。
func (k *PrivateKey) Close() error {
	if k == nil || k.key == nil {
		return nil
	}
	err := asym.Close(k.key)
	k.key = nil
	return err
}

// PublicKeyPEM 导出与私钥配对的公钥（SubjectPublicKeyInfo PEM）。
//
// 这里**刻意**不返回可持有的公钥对象：底层 `priv.Public()` 与私钥
// **共享同一个句柄**，单独关闭会连带释放私钥
// （见 docs/standards/03-crypto.md §3）。
func (k *PrivateKey) PublicKeyPEM() ([]byte, error) {
	if k == nil || k.key == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidKey, map[string]any{
			"reason": "key already closed",
		})
	}
	pub := k.key.Public()
	if pub == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidKey, map[string]any{
			"reason": "public key unavailable",
		})
	}
	pemBytes, err := pub.MarshalPublicKeyPEM()
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeCryptoOperationFailed, err)
	}
	return pemBytes, nil
}

// IsEncryptedPrivateKey 判断 PEM 是否是加密私钥块。
func IsEncryptedPrivateKey(pemBytes []byte) bool {
	return containsBlock(pemBytes, "ENCRYPTED PRIVATE KEY")
}

// containsBlock 判断 PEM 数据中是否含指定类型的块。
//
// 这里只用标准库做**容器解析**（不是密码学运算），因此允许。
func containsBlock(pemBytes []byte, blockType string) bool {
	block, _ := pem.Decode(pemBytes)
	return block != nil && block.Type == blockType
}
