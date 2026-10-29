package pki

import (
	"context"
	"encoding/base64"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// PEMDerRequest 是 PEM ↔ DER 转换工具的请求。
type PEMDerRequest struct {
	// Input 是待转换的材料。
	Input PemBundle `json:"input"`
	// Type 是类型过滤值（cert / key / csr / crl / any）；默认 any。
	Type string `json:"type"`
	// To 是目标格式（pem / der）。
	To string `json:"to"`
	// Label 是输出 PEM 的标签；为空时沿用源标签。
	Label string `json:"label"`
}

// PEMDerResult 是 PEM ↔ DER 转换工具的响应。
type PEMDerResult struct {
	// Output 是转换结果。
	Output PEMOutput `json:"output"`
	// Labels 是输入中识别到的 PEM 标签。
	Labels []string `json:"labels"`
	// BlockCount 是输出的 PEM 块数。
	BlockCount int `json:"blockCount"`
}

// PEMOutput 是转换结果的输出形态。
type PEMOutput struct {
	// Data 是输出文本：目标为 der 时是 base64，否则是 PEM 文本。
	Data string `json:"data"`
	// Format 是输出格式（pem / der）。
	Format string `json:"format"`
	// Encoding 是 Data 的传输编码（base64 / utf8）。
	Encoding string `json:"encoding"`
}

// Validate 做字段级校验。
func (r *PEMDerRequest) Validate() error {
	if err := r.Input.Validate("input"); err != nil {
		return err
	}
	if r.To == "" {
		return tools.MissingField("to")
	}
	if !containsString(pkisvc.ConvertTargets(), r.To) {
		return tools.UnsupportedFormat(r.To, pkisvc.ConvertTargets())
	}
	if r.Type == "" {
		r.Type = TypeAny
	}
	if !containsString(pkisvc.PEMTypes(), r.Type) {
		return tools.UnsupportedFormat(r.Type, pkisvc.PEMTypes())
	}
	return nil
}

// PEMDer 在 PEM 与 DER 之间转换。
//
// 目标为 der 时**只允许一个块**：DER 装不下多个对象。
func PEMDer(_ context.Context, req *PEMDerRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Input.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	out, err := pkisvc.ConvertPEMDER(pkisvc.PEMConvertOptions{
		Source:      source,
		SourceIsPEM: isPEM,
		Type:        req.Type,
		To:          req.To,
		Label:       req.Label,
	})
	if err != nil {
		return tools.Result{}, err
	}

	result := &PEMDerResult{
		Output: PEMOutput{
			Data:     out.Data,
			Format:   out.Format,
			Encoding: out.Encoding,
		},
		Labels:     out.Labels,
		BlockCount: out.BlockCount,
	}

	return tools.Result{Data: result}, nil
}

// encodeDER 把 DER 字节编码为 base64 文本。
func encodeDER(der []byte) string {
	return base64.StdEncoding.EncodeToString(der)
}
