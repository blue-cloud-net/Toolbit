package keys

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// KeyParseRequest 是密钥解析工具的请求。
type KeyParseRequest struct {
	// Key 是待解析的密钥。
	Key KeyInput `json:"key"`
	// Passphrase 是加密私钥口令。
	Passphrase string `json:"passphrase"`
}

// KeyParseResult 是密钥解析工具的响应。
type KeyParseResult struct {
	// Algorithm 是算法标识。
	Algorithm string `json:"algorithm"`
	// KeyType 是 private 或 public。
	KeyType string `json:"keyType"`
	// Format 是识别到的编码（pkcs1 / pkcs8 / spki）。
	Format string `json:"format"`
	// Encrypted 报告是否为加密私钥。
	Encrypted bool `json:"encrypted"`
	// Bits 是 RSA 位数；非 RSA 为 0。
	Bits int `json:"bits"`
	// Curve 是曲线名；非椭圆曲线为空。
	Curve string `json:"curve"`
	// FingerprintSHA256 是公钥 SPKI DER 的 SHA-256 指纹。
	FingerprintSHA256 string `json:"fingerprintSha256"`
	// Params 是 RSA 私钥参数；仅私钥且算法为 RSA 时返回。
	Params *RSAParamsOutput `json:"params,omitempty"`
	// NormalizedPem 是规范化后的 PEM。
	//
	// 源为加密私钥时**保持加密**（不会顺手解密暴露私钥）。
	NormalizedPem string `json:"normalizedPem"`
}

// Validate 做字段级校验。
func (r *KeyParseRequest) Validate() error {
	if err := r.Key.Validate("key"); err != nil {
		return err
	}
	// 口令可写在顶层 `passphrase` 或 `key.passphrase`，两者等价。
	if r.Passphrase == "" {
		r.Passphrase = r.Key.Passphrase
	}
	return nil
}

// KeyParse 解析密钥内容并展示其元信息。
//
// 输入可以是私钥、公钥或证书：证书按其公钥解析。
func KeyParse(_ context.Context, req *KeyParseRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := decodeInput(req.Key)
	if err != nil {
		return tools.Result{}, err
	}

	info, normalized, err := pkisvc.Inspect(source, isPEM, req.Passphrase)
	if err != nil {
		return tools.Result{}, err
	}

	result := &KeyParseResult{
		Algorithm:         info.Algorithm,
		KeyType:           info.KeyType,
		Format:            info.Format,
		Encrypted:         info.Encrypted,
		Bits:              info.Bits,
		Curve:             info.Curve,
		FingerprintSHA256: info.FingerprintSHA256,
		NormalizedPem:     encodePEM(normalized),
	}
	if info.RSA != nil {
		result.Params = &RSAParamsOutput{
			N:  info.RSA.N,
			E:  info.RSA.E,
			D:  info.RSA.D,
			P:  info.RSA.P,
			Q:  info.RSA.Q,
			Dp: info.RSA.Dp,
			Dq: info.RSA.Dq,
			Qi: info.RSA.Qi,
		}
	}

	return tools.Result{Data: result}, nil
}
