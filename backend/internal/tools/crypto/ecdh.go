package crypto

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// KDF 类型。
const (
	// KDFTypeHKDF 是 HKDF（唯一支持的类型）。
	KDFTypeHKDF = "hkdf"
)

// KDFOptions 是密钥派生参数。
type KDFOptions struct {
	// Type 是派生类型；当前只支持 hkdf。
	Type string `json:"type"`
	// Hash 是摘要名。
	Hash string `json:"hash"`
	// Salt 是盐值。
	Salt ByteInput `json:"salt"`
	// Info 是上下文信息。
	Info ByteInput `json:"info"`
	// Length 是派生长度（字节）。
	Length int `json:"length"`
}

// ECDHRequest 是 ECDH 密钥协商工具的请求。
type ECDHRequest struct {
	// Curve 是曲线名。
	Curve string `json:"curve"`
	// PrivateKey 是本方私钥。
	PrivateKey KeyInput `json:"privateKey"`
	// PeerPublicKey 是对端公钥。
	PeerPublicKey KeyInput `json:"peerPublicKey"`
	// KDF 是可选的派生参数。
	KDF *KDFOptions `json:"kdf"`
	// OutputEncoding 是输出编码；默认 base64。
	OutputEncoding string `json:"outputEncoding"`
}

// ECDHResult 是 ECDH 密钥协商工具的响应。
type ECDHResult struct {
	// SharedSecret 是共享密钥。
	SharedSecret string `json:"sharedSecret"`
	// DerivedKey 是派生密钥；提供 kdf 时返回。
	DerivedKey string `json:"derivedKey,omitempty"`
	// Curve 是实际使用的曲线。
	Curve string `json:"curve"`
	// Length 是共享密钥字节数。
	Length int `json:"length"`
}

// Validate 做字段级校验。
func (r *ECDHRequest) Validate() error {
	curves := cryptosvc.ECDHCurves()
	if !containsString(curves, r.Curve) {
		return tools.UnsupportedAlgorithm(r.Curve, curves)
	}
	if err := r.PrivateKey.Validate("privateKey"); err != nil {
		return err
	}
	if err := r.PeerPublicKey.Validate("peerPublicKey"); err != nil {
		return err
	}
	if r.KDF != nil {
		if r.KDF.Type != "" && r.KDF.Type != KDFTypeHKDF {
			return tools.UnsupportedFormat(r.KDF.Type, []string{KDFTypeHKDF})
		}
		hashes := cryptosvc.ECDHKDFHashes()
		if r.KDF.Hash != "" && !containsString(hashes, r.KDF.Hash) {
			return tools.UnsupportedAlgorithm(r.KDF.Hash, hashes)
		}
		if r.KDF.Length <= 0 {
			return tools.MissingField("kdf.length")
		}
		if err := r.KDF.Salt.Validate("kdf.salt"); err != nil {
			return err
		}
		if err := r.KDF.Info.Validate("kdf.info"); err != nil {
			return err
		}
	}
	if _, err := codecsvc.Normalize(r.OutputEncoding); err != nil {
		return tools.InvalidField("outputEncoding", codecsvc.Encodings(), r.OutputEncoding)
	}
	return nil
}

// ECDH 执行密钥协商与可选派生。
func ECDH(_ context.Context, req *ECDHRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	outputEncoding := req.OutputEncoding
	if outputEncoding == "" {
		outputEncoding = codecsvc.EncodingBase64
	}

	privPEM, err := req.PrivateKey.PrivateKeyPEM()
	if err != nil {
		return tools.Result{}, err
	}
	peerPEM, err := req.PeerPublicKey.PublicKeyPEM()
	if err != nil {
		return tools.Result{}, err
	}

	secret, err := cryptosvc.ECDHSharedSecret(req.Curve, privPEM, peerPEM)
	if err != nil {
		return tools.Result{}, err
	}
	defer cryptosvc.Zero(secret)

	secretText, err := codecsvc.Encode(secret, outputEncoding)
	if err != nil {
		return tools.Result{}, err
	}

	result := ECDHResult{
		SharedSecret: secretText,
		Curve:        req.Curve,
		Length:       len(secret),
	}

	if req.KDF != nil {
		hash := req.KDF.Hash
		if hash == "" {
			hash = cryptosvc.HashSHA256
		}
		var salt, info []byte
		if req.KDF.Salt.Data != "" {
			salt, err = req.KDF.Salt.Decode()
			if err != nil {
				return tools.Result{}, err
			}
		}
		if req.KDF.Info.Data != "" {
			info, err = req.KDF.Info.Decode()
			if err != nil {
				return tools.Result{}, err
			}
		}
		derived, err := cryptosvc.ECDHDerive(hash, secret, salt, info, req.KDF.Length)
		if err != nil {
			return tools.Result{}, err
		}
		defer cryptosvc.Zero(derived)

		derivedText, err := codecsvc.Encode(derived, outputEncoding)
		if err != nil {
			return tools.Result{}, err
		}
		result.DerivedKey = derivedText
	}

	return tools.Result{Data: result}, nil
}
