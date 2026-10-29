package keys

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// KeyConvertRequest 是密钥格式转换工具的请求。
type KeyConvertRequest struct {
	// Key 是待转换的密钥。
	Key KeyInput `json:"key"`
	// From 是源格式；默认 auto。
	From string `json:"from"`
	// To 是目标格式；必填。
	To string `json:"to"`
	// Passphrase 是加密私钥的现有口令。
	Passphrase string `json:"passphrase"`
}

// KeyConvertResult 是密钥格式转换工具的响应。
type KeyConvertResult struct {
	// Key 是转换后的密钥。
	Key PEMOutput `json:"key"`
	// Detected 是识别到的源密钥元信息。
	Detected DetectedInfo `json:"detected"`
}

// Validate 做字段级校验。
func (r *KeyConvertRequest) Validate() error {
	if err := r.Key.Validate("key"); err != nil {
		return err
	}
	if r.From == "" {
		r.From = pkisvc.FormatAuto
	}
	switch r.From {
	case pkisvc.FormatAuto, pkisvc.FormatPEM, pkisvc.FormatDER,
		pkisvc.FormatPKCS1, pkisvc.FormatPKCS8, pkisvc.FormatSPKI:
	default:
		return tools.UnsupportedFormat(r.From, append(pkisvc.KeyFormats(), pkisvc.FormatAuto))
	}
	if r.To == "" {
		return tools.MissingField("to")
	}
	if !containsString(pkisvc.KeyFormats(), r.To) {
		return tools.UnsupportedFormat(r.To, pkisvc.KeyFormats())
	}
	// 口令可写在顶层 `passphrase` 或 `key.passphrase`，两者等价。
	if r.Passphrase == "" {
		r.Passphrase = r.Key.Passphrase
	}
	return nil
}

// KeyConvert 转换密钥格式。
func KeyConvert(_ context.Context, req *KeyConvertRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := decodeInput(req.Key)
	if err != nil {
		return tools.Result{}, err
	}

	converted, err := pkisvc.Convert(pkisvc.ConvertOptions{
		Source:      source,
		SourceIsPEM: isPEM,
		Passphrase:  req.Passphrase,
		From:        req.From,
		To:          req.To,
	})
	if err != nil {
		return tools.Result{}, err
	}

	result := &KeyConvertResult{
		Key: PEMOutput{PEM: encodePEM(converted.PEM)},
		Detected: DetectedInfo{
			Algorithm: converted.Algorithm,
			Format:    converted.Detected,
			Encrypted: converted.Encrypted,
			Bits:      converted.Bits,
			Curve:     converted.Curve,
		},
	}
	if len(converted.DER) > 0 {
		result.Key.DERBase64 = encodeDER(converted.DER)
		result.Key.PEM = ""
	}

	return tools.Result{Data: result}, nil
}
