package crypto

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 非对称加解密的操作类型。
const (
	// AsymOpEncrypt 是加密。
	AsymOpEncrypt = "encrypt"
	// AsymOpDecrypt 是解密。
	AsymOpDecrypt = "decrypt"
)

// RSA 填充方式。
const (
	// RSAPaddingOAEP 是 RSA-OAEP（默认）。
	RSAPaddingOAEP = "oaep"
	// RSAPaddingPKCS1 是 RSA-PKCS#1 v1.5。
	RSAPaddingPKCS1 = "pkcs1"
)

// AsymmetricCryptRequest 是非对称加解密工具的请求。
type AsymmetricCryptRequest struct {
	// Algorithm 是 RSA 或 SM2。
	Algorithm string `json:"algorithm"`
	// Operation 是 encrypt 或 decrypt。
	Operation string `json:"operation"`
	// Key 是公钥（加密）或私钥（解密）。
	Key KeyInput `json:"key"`
	// Input 是明文或密文。
	Input ByteInput `json:"input"`
	// RSAPadding 是 RSA 填充；默认 oaep。
	RSAPadding string `json:"rsaPadding"`
	// OAEPHash 是 OAEP 摘要；默认 SHA256。
	OAEPHash string `json:"oaepHash"`
	// OutputEncoding 是输出编码；默认 base64。
	OutputEncoding string `json:"outputEncoding"`
}

// AsymmetricCryptResult 是非对称加解密工具的响应。
type AsymmetricCryptResult struct {
	// Output 是密文或明文。
	Output string `json:"output"`
	// Algorithm 是实际使用的算法。
	Algorithm string `json:"algorithm"`
	// Order 是 SM2 密文顺序。
	Order string `json:"order,omitempty"`
	// Format 是 SM2 密文格式。
	Format string `json:"format,omitempty"`
}

// Validate 做字段级校验。
func (r *AsymmetricCryptRequest) Validate() error {
	switch r.Operation {
	case AsymOpEncrypt, AsymOpDecrypt:
	default:
		return tools.InvalidField("operation", []string{AsymOpEncrypt, AsymOpDecrypt}, r.Operation)
	}

	algorithms := cryptosvc.AsymAlgorithms()
	if !containsString(algorithms, r.Algorithm) {
		return tools.UnsupportedAlgorithm(r.Algorithm, algorithms)
	}

	if r.RSAPadding != "" {
		switch r.RSAPadding {
		case RSAPaddingOAEP, RSAPaddingPKCS1:
		default:
			return tools.UnsupportedPadding(r.RSAPadding, []string{RSAPaddingOAEP, RSAPaddingPKCS1})
		}
	}

	if err := r.Key.Validate("key"); err != nil {
		return err
	}
	if err := r.Input.Validate("input"); err != nil {
		return err
	}
	if _, err := codecsvc.Normalize(r.OutputEncoding); err != nil {
		return tools.InvalidField("outputEncoding", codecsvc.Encodings(), r.OutputEncoding)
	}
	return nil
}

// AsymmetricCrypt 执行非对称加解密。
func AsymmetricCrypt(_ context.Context, req *AsymmetricCryptRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	outputEncoding := req.OutputEncoding
	if outputEncoding == "" {
		outputEncoding = codecsvc.EncodingBase64
	}

	input, err := req.Input.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	var output []byte
	if req.Operation == AsymOpEncrypt {
		pemBytes, err := req.Key.PublicKeyPEM()
		if err != nil {
			return tools.Result{}, err
		}
		pub, err := cryptosvc.LoadPublicKey(pemBytes)
		if err != nil {
			return tools.Result{}, err
		}
		output, err = cryptosvc.AsymEncrypt(req.Algorithm, req.RSAPadding, req.OAEPHash, pub, input)
		if err != nil {
			return tools.Result{}, err
		}
	} else {
		pemBytes, err := req.Key.PrivateKeyPEM()
		if err != nil {
			return tools.Result{}, err
		}
		priv, err := cryptosvc.LoadPrivateKey(pemBytes, req.Key.Passphrase)
		if err != nil {
			return tools.Result{}, err
		}
		defer func() { _ = priv.Close() }()

		output, err = cryptosvc.AsymDecrypt(req.Algorithm, req.RSAPadding, req.OAEPHash, priv, input)
		if err != nil {
			return tools.Result{}, err
		}
	}

	text, err := codecsvc.Encode(output, outputEncoding)
	if err != nil {
		return tools.Result{}, err
	}

	result := AsymmetricCryptResult{
		Output:    text,
		Algorithm: req.Algorithm,
	}
	if req.Algorithm == cryptosvc.AsymSM2 {
		result.Order = cryptosvc.SM2OrderC1C3C2
		result.Format = "raw"
	}

	return tools.Result{Data: result}, nil
}
