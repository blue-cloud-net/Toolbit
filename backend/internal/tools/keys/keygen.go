package keys

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// EncryptOptions 是生成密钥时的加密参数。
type EncryptOptions struct {
	// Cipher 是加密算法；为空时用 AES-256-CBC。
	Cipher string `json:"cipher"`
	// Passphrase 是口令。
	Passphrase string `json:"passphrase"`
}

// KeyGenRequest 是密钥生成工具的请求。
type KeyGenRequest struct {
	// Algorithm 是算法标识。
	Algorithm string `json:"algorithm"`
	// Bits 是 RSA 密钥位数；为空时默认 2048。
	Bits int `json:"bits"`
	// Curve 是曲线名；EC / SM2 时必填。
	Curve string `json:"curve"`
	// PrivateFormat 是私钥输出格式；默认 pkcs8。
	PrivateFormat string `json:"privateFormat"`
	// PublicFormat 是公钥输出格式；默认 spki。
	PublicFormat string `json:"publicFormat"`
	// Encrypt 是可选的口令加密参数。
	Encrypt *EncryptOptions `json:"encrypt"`
	// Count 是生成数量；当前只支持 1。
	Count int `json:"count"`
}

// KeyGenResult 是密钥生成工具的响应。
type KeyGenResult struct {
	// PrivateKey 是私钥。
	PrivateKey PEMOutput `json:"privateKey"`
	// PublicKey 是公钥。
	PublicKey PEMOutput `json:"publicKey"`
	// Algorithm 是实际算法。
	Algorithm string `json:"algorithm"`
	// Bits 是密钥位数（RSA）。
	Bits int `json:"bits"`
	// Curve 是曲线（EC / SM2）。
	Curve string `json:"curve"`
}

// Validate 做字段级校验。
func (r *KeyGenRequest) Validate() error {
	if r.Algorithm == "" {
		return tools.MissingField("algorithm")
	}
	if !containsString(pkisvc.KeyAlgorithms(), r.Algorithm) {
		return tools.UnsupportedAlgorithm(r.Algorithm, pkisvc.KeyAlgorithms())
	}

	switch r.Algorithm {
	case pkisvc.AlgRSA:
		if r.Bits == 0 {
			r.Bits = 2048
		}
		if !containsInt(pkisvc.RSABits(), r.Bits) {
			return tools.InvalidField("bits", pkisvc.RSABits(), r.Bits)
		}
	case pkisvc.AlgEC, pkisvc.AlgSM2:
		if r.Curve == "" {
			return tools.MissingField("curve")
		}
		if !containsString(pkisvc.ECCurves(), r.Curve) {
			return tools.InvalidField("curve", pkisvc.ECCurves(), r.Curve)
		}
	}

	if r.PrivateFormat == "" {
		r.PrivateFormat = PrivateFormatPKCS8
	}
	if r.PrivateFormat != PrivateFormatPKCS8 && r.PrivateFormat != PrivateFormatPKCS1 {
		return tools.InvalidField("privateFormat",
			[]string{PrivateFormatPKCS8, PrivateFormatPKCS1}, r.PrivateFormat)
	}

	if r.PublicFormat == "" {
		r.PublicFormat = PublicFormatSPKI
	}
	if r.PublicFormat != PublicFormatSPKI {
		return tools.UnsupportedFormat(r.PublicFormat, []string{PublicFormatSPKI})
	}

	if r.Encrypt != nil {
		if r.Encrypt.Cipher == "" {
			r.Encrypt.Cipher = pkisvc.PBEAES256CBC
		}
		if !containsString(pkisvc.PBECiphers(), r.Encrypt.Cipher) {
			return tools.InvalidField("encrypt.cipher", pkisvc.PBECiphers(), r.Encrypt.Cipher)
		}
		if r.Encrypt.Passphrase == "" {
			return tools.MissingField("encrypt.passphrase")
		}
	}

	if r.Count == 0 {
		r.Count = 1
	}
	if r.Count != 1 {
		// 响应结构只能表达**一把**密钥；批量生成需要新增响应字段，
		// 属于破坏性契约变更。0.x 轨道允许，但本版不实现：
		// 显式拒绝比悄悄忽略更安全（见 docs/roadmap.md §7.4 K7）。
		return tools.InvalidField("count", []int{1}, r.Count)
	}

	return nil
}

// KeyGen 生成密钥对。
//
// 重操作：走 Session 的重操作许可（并发上限 2）。
func KeyGen(ctx context.Context, req *KeyGenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	session := cryptosvc.FromContext(ctx)
	if session != nil {
		if err := session.Acquire(ctx); err != nil {
			return tools.Result{}, err
		}
		defer session.Release("keygen")
	}

	pair, err := pkisvc.Generate(req.Algorithm, req.Bits, req.Curve)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = pair.Close() }()

	privatePEM, err := marshalPrivate(pair, req.PrivateFormat)
	if err != nil {
		return tools.Result{}, err
	}

	publicPEM, err := pair.Public.MarshalPEM()
	if err != nil {
		return tools.Result{}, err
	}

	if req.Encrypt != nil {
		privatePEM, err = pkisvc.EncryptPrivateKey(privatePEM, req.Encrypt.Passphrase, req.Encrypt.Cipher)
		if err != nil {
			return tools.Result{}, err
		}
	}

	result := &KeyGenResult{
		PrivateKey: PEMOutput{PEM: encodePEM(privatePEM)},
		PublicKey:  PEMOutput{PEM: encodePEM(publicPEM)},
		Algorithm:  req.Algorithm,
		Bits:       req.Bits,
		Curve:      req.Curve,
	}
	if req.Algorithm != pkisvc.AlgRSA {
		result.Bits = 0
	}
	if req.Algorithm != pkisvc.AlgEC && req.Algorithm != pkisvc.AlgSM2 {
		result.Curve = ""
	}

	return tools.Result{Data: result}, nil
}

// marshalPrivate 按目标格式导出私钥。
func marshalPrivate(pair *pkisvc.KeyPair, format string) ([]byte, error) {
	switch format {
	case PrivateFormatPKCS1:
		return pair.Private.MarshalPKCS1PEM()
	default:
		return pair.Private.MarshalPKCS8PEM()
	}
}
