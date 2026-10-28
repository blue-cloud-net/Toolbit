package crypto

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// HmacGenRequest 是 HMAC 计算工具的请求。
type HmacGenRequest struct {
	// Algorithm 是 HMAC 算法名，如 HMAC-SM3。
	Algorithm string `json:"algorithm"`
	// Key 是消息认证码密钥。
	Key ByteInput `json:"key"`
	// Message 是待计算认证码的消息。
	Message ByteInput `json:"message"`
	// OutputEncoding 是输出编码；默认 hex。
	OutputEncoding string `json:"outputEncoding"`
}

// HmacGenResult 是 HMAC 计算工具的响应。
type HmacGenResult struct {
	// MAC 是认证码值。
	MAC string `json:"mac"`
	// Algorithm 是实际使用的算法。
	Algorithm string `json:"algorithm"`
}

// Validate 做字段级校验。
func (r *HmacGenRequest) Validate() error {
	supported := cryptosvc.MACAlgorithms()
	if !containsString(supported, r.Algorithm) {
		return tools.UnsupportedAlgorithm(r.Algorithm, supported)
	}
	if err := r.Key.Validate("key"); err != nil {
		return err
	}
	if err := r.Message.Validate("message"); err != nil {
		return err
	}
	if _, err := codecsvc.Normalize(r.OutputEncoding); err != nil {
		return tools.InvalidField("outputEncoding", codecsvc.Encodings(), r.OutputEncoding)
	}
	return nil
}

// HmacGen 计算消息认证码。
func HmacGen(_ context.Context, req *HmacGenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	key, err := req.Key.Decode()
	if err != nil {
		return tools.Result{}, err
	}
	message, err := req.Message.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	mac, err := cryptosvc.MAC(req.Algorithm, key, message)
	if err != nil {
		return tools.Result{}, err
	}
	cryptosvc.Zero(key)

	// 认证码是二进制，缺省按十六进制输出。
	outputEncoding := req.OutputEncoding
	if outputEncoding == "" {
		outputEncoding = codecsvc.EncodingHex
	}
	value, err := codecsvc.Encode(mac, outputEncoding)
	if err != nil {
		return tools.Result{}, err
	}

	result := HmacGenResult{MAC: value, Algorithm: req.Algorithm}
	var warnings []toolerr.Warning
	if cryptosvc.IsWeakMAC(req.Algorithm) {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnWeakHash, map[string]any{
			"algorithm": req.Algorithm,
		}))
	}

	return tools.Result{Data: result, Warnings: warnings}, nil
}
