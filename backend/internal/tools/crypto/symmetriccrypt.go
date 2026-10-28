package crypto

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 对称加解密的操作类型。
const (
	// SymOpEncrypt 是加密。
	SymOpEncrypt = "encrypt"
	// SymOpDecrypt 是解密。
	SymOpDecrypt = "decrypt"
)

// SymmetricCryptRequest 是对称加解密工具的请求。
type SymmetricCryptRequest struct {
	// Algorithm 是 SM4 或 AES。
	Algorithm string `json:"algorithm"`
	// Operation 是 encrypt 或 decrypt。
	Operation string `json:"operation"`
	// Mode 是加密模式。
	Mode string `json:"mode"`
	// Padding 是填充方式；默认 pkcs7。
	Padding string `json:"padding"`
	// Key 是对称密钥。
	Key ByteInput `json:"key"`
	// IV 是初始向量；GCM 为 nonce，其余为 16 字节 IV。
	IV ByteInput `json:"iv"`
	// AAD 是 GCM 的附加认证数据。
	AAD ByteInput `json:"aad"`
	// Tag 是 GCM 的认证标签；解密时必填。
	Tag ByteInput `json:"tag"`
	// Input 是待处理内容。
	Input ByteInput `json:"input"`
	// OutputEncoding 是输出编码；默认 base64。
	OutputEncoding string `json:"outputEncoding"`
}

// SymmetricCryptResult 是对称加解密工具的响应。
type SymmetricCryptResult struct {
	// Output 是密文或明文。
	Output string `json:"output"`
	// Tag 是 GCM 的认证标签（十六进制）。
	Tag string `json:"tag,omitempty"`
	// Algorithm 是实际使用的算法。
	Algorithm string `json:"algorithm"`
	// Mode 是实际使用的模式。
	Mode string `json:"mode"`
	// Padding 是实际使用的填充方式。
	Padding string `json:"padding"`
	// KeyBytes 是密钥长度（字节）。
	KeyBytes int `json:"keyBytes"`
	// IVBytes 是 IV 长度（字节）。
	IVBytes int `json:"ivBytes"`
	// IV 是自动生成的 nonce（仅 GCM 加密且未提供 IV 时返回）。
	IV string `json:"iv,omitempty"`
}

// Validate 做字段级校验。
func (r *SymmetricCryptRequest) Validate() error {
	switch r.Operation {
	case SymOpEncrypt, SymOpDecrypt:
	default:
		return tools.InvalidField("operation", []string{SymOpEncrypt, SymOpDecrypt}, r.Operation)
	}

	algorithms := cryptosvc.SymAlgorithms()
	if !containsString(algorithms, r.Algorithm) {
		return tools.UnsupportedAlgorithm(r.Algorithm, algorithms)
	}

	modes := cryptosvc.SymModes(r.Algorithm)
	if !containsString(modes, r.Mode) {
		return tools.UnsupportedMode(r.Mode, modes)
	}

	padding := r.Padding
	if padding == "" {
		padding = cryptosvc.PaddingPKCS7
	}
	paddings := cryptosvc.SymPaddings(r.Algorithm, r.Mode)
	if !containsString(paddings, padding) {
		return tools.UnsupportedPadding(padding, paddings)
	}

	if err := r.Key.Validate("key"); err != nil {
		return err
	}
	if err := r.IV.Validate("iv"); err != nil {
		return err
	}
	if err := r.AAD.Validate("aad"); err != nil {
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

// SymmetricCrypt 执行对称加解密。
func SymmetricCrypt(_ context.Context, req *SymmetricCryptRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	padding := req.Padding
	if padding == "" {
		padding = cryptosvc.PaddingPKCS7
	}
	outputEncoding := req.OutputEncoding
	if outputEncoding == "" {
		outputEncoding = codecsvc.EncodingBase64
	}

	key, err := req.Key.Decode()
	if err != nil {
		return tools.Result{}, err
	}
	defer cryptosvc.Zero(key)

	var iv []byte
	if req.IV.Data != "" {
		iv, err = req.IV.Decode()
		if err != nil {
			return tools.Result{}, err
		}
	} else if req.Mode == cryptosvc.ModeGCM && req.Operation == SymOpEncrypt {
		// GCM 加密缺省 nonce 时生成随机值，并在响应里回传（解密需复用）。
		iv, err = cryptosvc.RandomBytes(cryptosvc.GCMNonceSize)
		if err != nil {
			return tools.Result{}, err
		}
	}

	var aad []byte
	if req.AAD.Data != "" {
		aad, err = req.AAD.Decode()
		if err != nil {
			return tools.Result{}, err
		}
	}

	input, err := req.Input.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	var (
		output []byte
		tag    []byte
	)
	generatedIV := req.IV.Data == "" && req.Mode == cryptosvc.ModeGCM && req.Operation == SymOpEncrypt
	if req.Operation == SymOpEncrypt {
		output, tag, err = cryptosvc.SymEncrypt(req.Algorithm, req.Mode, padding, key, iv, input, aad)
	} else {
		var inputTag []byte
		if req.Tag.Data != "" {
			inputTag, err = req.Tag.Decode()
			if err != nil {
				return tools.Result{}, err
			}
		}
		output, err = cryptosvc.SymDecrypt(req.Algorithm, req.Mode, padding, key, iv, input, inputTag, aad)
	}
	if err != nil {
		return tools.Result{}, err
	}

	outText, err := codecsvc.Encode(output, outputEncoding)
	if err != nil {
		return tools.Result{}, err
	}

	result := SymmetricCryptResult{
		Output:    outText,
		Algorithm: req.Algorithm,
		Mode:      req.Mode,
		Padding:   padding,
		KeyBytes:  len(key),
		IVBytes:   len(iv),
	}
	if len(tag) > 0 {
		result.Tag = hexEncode(tag)
	}
	if generatedIV {
		result.IV = hexEncode(iv)
	}

	var warnings []toolerr.Warning
	if req.Mode == cryptosvc.ModeECB {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnInsecureECB, nil))
	} else if req.Mode != cryptosvc.ModeGCM {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnNoAuthentication, nil))
	}

	return tools.Result{Data: result, Warnings: warnings}, nil
}
