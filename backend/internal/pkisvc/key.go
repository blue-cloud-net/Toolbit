package pkisvc

import (
	"encoding/pem"

	"github.com/blue-cloud-net/tongsuo-go/asym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 密钥格式标识。
const (
	// FormatAuto 自动识别。
	FormatAuto = "auto"
	// FormatPKCS1 是 PKCS#1 私钥（仅 RSA）。
	FormatPKCS1 = "pkcs1"
	// FormatPKCS8 是 PKCS#8 私钥。
	FormatPKCS8 = "pkcs8"
	// FormatSPKI 是 SubjectPublicKeyInfo 公钥。
	FormatSPKI = "spki"
	// FormatDER 是二进制 DER。
	FormatDER = "der"
	// FormatPEM 是 PEM 文本。
	FormatPEM = "pem"
)

// KeyFormats 返回受支持的密钥格式名（顺序稳定）。
//
// 该列表是**转换目标**的取值域；FormatAuto 只作为**源**格式出现。
func KeyFormats() []string {
	return []string{FormatPKCS1, FormatPKCS8, FormatSPKI, FormatDER, FormatPEM}
}

// 密钥类型标识。
const (
	// KeyTypePrivate 是私钥。
	KeyTypePrivate = "private"
	// KeyTypePublic 是公钥。
	KeyTypePublic = "public"
)

// PrivateKey 是非对称私钥句柄。
//
// 它持有原生句柄，**必须**显式 Close（见 docs/standards/03-crypto.md §3）。
type PrivateKey struct {
	// key 是底层私钥；Close 之后置为 nil。
	key asym.PrivateKey
}

// PublicKey 是非对称公钥句柄。
type PublicKey struct {
	// key 是底层公钥。
	key asym.PublicKey
}

// KeyPair 是一对密钥。
//
// Public 是**独立持有**的公钥句柄（由私钥的 SPKI 重新加载而来），
// 因此两者可以各自 Close；这与 `priv.Public()` 的语义不同——
// 后者与私钥**共享**底层句柄，单独关闭会连带释放私钥。
type KeyPair struct {
	// Private 是私钥句柄。
	Private *PrivateKey
	// Public 是公钥句柄。
	Public *PublicKey
}

// Close 释放公钥与私钥。
//
// **幂等**：重复调用只返回首次的结果。
func (p *KeyPair) Close() error {
	if p == nil {
		return nil
	}
	var firstErr error
	if p.Public != nil {
		if err := p.Public.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		p.Public = nil
	}
	if p.Private != nil {
		if err := p.Private.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		p.Private = nil
	}
	return firstErr
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

// Close 释放公钥持有的原生句柄。
//
// **幂等**：重复调用只返回首次的结果。
func (k *PublicKey) Close() error {
	if k == nil || k.key == nil {
		return nil
	}
	err := asym.Close(k.key)
	k.key = nil
	return err
}

// Algorithm 返回算法标识（对外友好名）。
func (k *PrivateKey) Algorithm() string {
	if k == nil || k.key == nil {
		return ""
	}
	return algorithmFromUpstream(k.key.Algorithm().String())
}

// Algorithm 返回算法标识（对外友好名）。
func (k *PublicKey) Algorithm() string {
	if k == nil || k.key == nil {
		return ""
	}
	return algorithmFromUpstream(k.key.Algorithm().String())
}

// MarshalPKCS8PEM 导出私钥为 PKCS#8 PEM。
func (k *PrivateKey) MarshalPKCS8PEM() ([]byte, error) {
	if k == nil || k.key == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidKey, map[string]any{
			"reason": "key already closed",
		})
	}
	out, err := k.key.MarshalPrivateKeyPEM()
	if err != nil {
		return nil, MapError(err)
	}
	return out, nil
}

// MarshalPKCS1PEM 导出私钥为 PKCS#1 PEM。
//
// 仅 RSA 支持；其它算法返回 UNSUPPORTED_FORMAT。
func (k *PrivateKey) MarshalPKCS1PEM() ([]byte, error) {
	if k == nil || k.key == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidKey, map[string]any{
			"reason": "key already closed",
		})
	}
	out, err := asym.MarshalRSAPrivateKeyPKCS1PEM(k.key)
	if err != nil {
		return nil, unsupportedFormat(FormatPKCS1, KeyFormats())
	}
	return out, nil
}

// MarshalPublicKeyPEM 导出配对的公钥为 SPKI PEM。
//
// 这里**不**返回可持有的公钥对象：底层 `priv.Public()` 与私钥
// **共享同一个句柄**，单独关闭会连带释放私钥
// （见 docs/standards/03-crypto.md §3）。
func (k *PrivateKey) MarshalPublicKeyPEM() ([]byte, error) {
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
	out, err := pub.MarshalPublicKeyPEM()
	if err != nil {
		return nil, MapError(err)
	}
	return out, nil
}

// MarshalPEM 导出公钥为 SPKI PEM。
func (k *PublicKey) MarshalPEM() ([]byte, error) {
	if k == nil || k.key == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidKey, map[string]any{
			"reason": "key already closed",
		})
	}
	out, err := k.key.MarshalPublicKeyPEM()
	if err != nil {
		return nil, MapError(err)
	}
	return out, nil
}

// PublicKeyPEM 返回公钥的 SPKI PEM。
//
// 公钥句柄不可用时**不**报错，而是返回空串：调用方据此判断
// 「这不是一把公钥」。
func (k *PublicKey) PublicKeyPEM() ([]byte, error) {
	if k == nil || k.key == nil {
		return nil, nil
	}
	out, err := k.key.MarshalPublicKeyPEM()
	if err != nil {
		return nil, MapError(err)
	}
	return out, nil
}

// Generate 生成密钥对。
//
// algorithm 必须是 KeyAlgorithms 之一；
// RSA 用 bits（必须是 RSABits 之一），EC / SM2 用 curve。
//
// 返回的公钥是**独立持有**的句柄，可与私钥分别释放。
func Generate(algorithm string, bits int, curve string) (*KeyPair, error) {
	var (
		priv asym.PrivateKey
		err  error
	)

	switch algorithm {
	case AlgRSA:
		if bits == 0 {
			bits = 2048
		}
		if !containsInt(RSABits(), bits) {
			return nil, toolerr.Newf(toolerr.CodeKeySizeMismatch, map[string]any{
				"bits":      bits,
				"supported": RSABits(),
			})
		}
		priv, err = asym.GenerateRSA(bits)
	case AlgSM2:
		if curve != "" && curve != CurveSM2 {
			return nil, unsupportedFormat(curve, ECCurves())
		}
		priv, err = asym.GenerateSM2()
	case AlgEC:
		upstream, ok := curveToUpstream(curve)
		if !ok {
			return nil, unsupportedFormat(curve, ECCurves())
		}
		priv, err = asym.GenerateEC(upstream)
	case AlgEd25519:
		priv, err = asym.GenerateEd25519()
	case AlgEd448:
		priv, err = asym.GenerateEd448()
	case AlgX25519:
		priv, err = asym.GenerateX25519()
	case AlgX448:
		priv, err = asym.GenerateX448()
	default:
		return nil, unsupportedAlgorithm(algorithm, KeyAlgorithms())
	}

	if err != nil {
		return nil, MapError(err)
	}

	pair := &KeyPair{Private: &PrivateKey{key: priv}}

	// 由私钥导出 SPKI 后**重新加载**公钥，得到独立句柄。
	pubPEM, err := pair.Private.MarshalPublicKeyPEM()
	if err != nil {
		_ = pair.Close()
		return nil, err
	}
	pub, err := asym.LoadPublicKeyPEM(pubPEM)
	if err != nil {
		_ = pair.Close()
		return nil, MapError(err)
	}
	pair.Public = &PublicKey{key: pub}

	return pair, nil
}

// LoadPrivateKey 从 PEM 加载私钥。
//
// passphrase 非空时按**加密私钥**处理（PKCS#8 + PBES2）。
// PEM 是加密块但未给口令时返回 PASSPHRASE_REQUIRED。
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
			if IsEncryptedPrivateKey(pemBytes) {
				return nil, toolerr.Newf(toolerr.CodePassphraseRequired, map[string]any{
					"field": "key.passphrase",
				})
			}
			return nil, toolerr.Wrap(toolerr.CodeInvalidKey, err)
		}
	}

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
	return &PublicKey{key: key}, nil
}

// IsEncryptedPrivateKey 判断 PEM 是否是加密私钥块。
func IsEncryptedPrivateKey(pemBytes []byte) bool {
	return containsBlock(pemBytes, "ENCRYPTED PRIVATE KEY")
}

// containsBlock 判断 PEM 中是否出现指定的块类型。
//
// 只看**首个**块：密钥材料是单块的，多块输入由调用方先行拆分。
func containsBlock(pemBytes []byte, blockType string) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}
	return block.Type == blockType
}

// algorithmFromUpstream 把底层库的算法名映射为对外友好名。
func algorithmFromUpstream(alg string) string {
	switch alg {
	case "ED25519":
		return AlgEd25519
	case "ED448":
		return AlgEd448
	default:
		return alg
	}
}

// containsInt 判断整数切片是否包含目标值。
func containsInt(list []int, target int) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// PublicKeyFromCertificate 从证书里取出公钥句柄。
//
// 返回的是**独立持有**的句柄（上游 `Certificate.PublicKey()` 已做 Dup），
// 调用方必须显式 Close。
func PublicKeyFromCertificate(cert *Certificate) (*PublicKey, error) {
	if cert == nil || cert.cert == nil {
		return nil, toolerrInvalidCert("certificate unavailable")
	}
	pub, err := cert.cert.PublicKey()
	if err != nil {
		return nil, toolerrInvalidCert("cannot read public key")
	}
	return &PublicKey{key: pub}, nil
}

// PublicKeyFromPrivateKey 从私钥派生公钥，返回**独立持有**的句柄。
//
// 与 `PrivateKey.MarshalPublicKeyPEM` 的区别是它返回一个可独立释放的对象：
// 内部经 SPKI PEM 往返重新加载，避免使用与私钥共享句柄的 `priv.Public()`。
func PublicKeyFromPrivateKey(priv *PrivateKey) (*PublicKey, error) {
	if priv == nil || priv.key == nil {
		return nil, toolerrInvalidKey("key already closed")
	}
	pemBytes, err := priv.MarshalPublicKeyPEM()
	if err != nil {
		return nil, err
	}
	return LoadPublicKey(pemBytes)
}
