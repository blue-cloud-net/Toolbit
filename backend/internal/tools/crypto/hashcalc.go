package crypto

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// HashCalcRequest 是摘要计算工具的请求。
type HashCalcRequest struct {
	// Algorithms 是要计算的摘要算法列表。
	Algorithms []string `json:"algorithms"`
	// Input 是待摘要的内容。
	Input ByteInput `json:"input"`
	// OutputEncoding 是输出编码；默认 hex。
	OutputEncoding string `json:"outputEncoding"`
}

// DigestValue 是单个摘要结果。
type DigestValue struct {
	// Algorithm 是摘要算法名。
	Algorithm string `json:"algorithm"`
	// Value 是摘要值。
	Value string `json:"value"`
}

// HashCalcResult 是摘要计算工具的响应。
type HashCalcResult struct {
	// Digests 是各算法的摘要值。
	Digests []DigestValue `json:"digests"`
	// InputBytes 是输入字节数。
	InputBytes int `json:"inputBytes"`
}

// Validate 做字段级校验。
func (r *HashCalcRequest) Validate() error {
	if len(r.Algorithms) == 0 {
		return tools.MissingField("algorithms")
	}
	supported := cryptosvc.DigestAlgorithms()
	for _, alg := range r.Algorithms {
		if !containsString(supported, alg) {
			return tools.UnsupportedAlgorithm(alg, supported)
		}
	}
	if err := r.Input.Validate("input"); err != nil {
		return err
	}
	if _, err := codecsvc.Normalize(r.OutputEncoding); err != nil {
		return tools.InvalidField("outputEncoding", codecsvc.Encodings(), r.OutputEncoding)
	}
	return nil
}

// HashCalc 计算摘要。
func HashCalc(_ context.Context, req *HashCalcRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := req.Input.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	// 摘要值是二进制，缺省按十六进制输出。
	outputEncoding := req.OutputEncoding
	if outputEncoding == "" {
		outputEncoding = codecsvc.EncodingHex
	}

	result := HashCalcResult{
		Digests:    make([]DigestValue, 0, len(req.Algorithms)),
		InputBytes: len(raw),
	}
	var warnings []toolerr.Warning
	for _, alg := range req.Algorithms {
		sum, err := cryptosvc.Digest(alg, raw)
		if err != nil {
			return tools.Result{}, err
		}
		value, err := codecsvc.Encode(sum, outputEncoding)
		if err != nil {
			return tools.Result{}, err
		}
		result.Digests = append(result.Digests, DigestValue{Algorithm: alg, Value: value})

		if cryptosvc.IsWeakDigest(alg) {
			warnings = append(warnings, toolerr.NewWarning(toolerr.WarnWeakHash, map[string]any{
				"algorithm": alg,
			}))
		}
	}

	return tools.Result{Data: result, Warnings: warnings}, nil
}

// containsString 报告列表中是否含有该字符串。
func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
