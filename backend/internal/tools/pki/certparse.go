package pki

import (
	"context"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CertParseRequest 是证书解析工具的请求。
type CertParseRequest struct {
	// Input 是待解析的证书。
	Input PemBundle `json:"input"`
	// AtTime 是判断有效性的时刻（RFC3339）；为空时按当前时间。
	AtTime string `json:"atTime"`
	// Fingerprints 是要计算的指纹算法；为空时只算 SHA-256。
	Fingerprints []string `json:"fingerprints"`
}

// CertParseResult 是证书解析工具的响应。
type CertParseResult struct {
	// Certificate 是证书元信息。
	Certificate CertInfoOutput `json:"certificate"`
	// Validity 是有效性判定。
	Validity ValidityOutput `json:"validity"`
	// Fingerprints 是计算出的指纹。
	Fingerprints []FingerprintOutput `json:"fingerprints"`
	// SelfSigned 报告是否自签名。
	SelfSigned bool `json:"selfSigned"`
	// RawDerBase64 是证书的原始 DER（base64）。
	RawDerBase64 string `json:"rawDerBase64"`
}

// Validate 做字段级校验。
func (r *CertParseRequest) Validate() error {
	if err := r.Input.Validate("input"); err != nil {
		return err
	}
	for _, alg := range r.Fingerprints {
		if !containsString(pkisvc.FingerprintAlgorithms(), alg) {
			return tools.UnsupportedAlgorithm(alg, pkisvc.FingerprintAlgorithms())
		}
	}
	return nil
}

// CertParse 解析一张证书。
//
// 输入可以是 PEM 文本或 DER 二进制；PEM bundle 时只取第一块证书。
func CertParse(_ context.Context, req *CertParseRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	at, err := parseAtTime(req.AtTime)
	if err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Input.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	cert, err := pkisvc.LoadCertificate(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = cert.Close() }()

	info, err := pkisvc.Describe(cert, at, req.Fingerprints)
	if err != nil {
		return tools.Result{}, err
	}

	fps, err := pkisvc.Fingerprints(cert, req.Fingerprints)
	if err != nil {
		return tools.Result{}, err
	}

	der, err := cert.MarshalDER()
	if err != nil {
		return tools.Result{}, err
	}

	result := &CertParseResult{
		Certificate:  certInfoOutput(info),
		Validity:     validityOutput(pkisvc.CheckValidity(cert, at)),
		Fingerprints: fingerprintsOutput(fps),
		SelfSigned:   info.SelfSigned,
		RawDerBase64: encodeDER(der),
	}

	return tools.Result{Data: result, Warnings: certWarnings(cert, info, at)}, nil
}

// validityOutput 把有效性判定转成输出形态。
func validityOutput(info pkisvc.ValidityInfo) ValidityOutput {
	return ValidityOutput{
		Status:        info.Status,
		DaysRemaining: info.DaysRemaining,
	}
}

// fingerprintsOutput 把指纹列表转成输出形态。
func fingerprintsOutput(fps []pkisvc.FingerprintInfo) []FingerprintOutput {
	out := make([]FingerprintOutput, 0, len(fps))
	for _, fp := range fps {
		out = append(out, FingerprintOutput{Algorithm: fp.Algorithm, Value: fp.Value})
	}
	return out
}

// certWarnings 汇总与证书状态相关的警告。
//
// 警告**不**改变 HTTP 状态：证书已过期仍然是「解析成功」。
func certWarnings(cert *pkisvc.Certificate, info *pkisvc.CertInfo, at time.Time) []toolerr.Warning {
	var warnings []toolerr.Warning

	validity := pkisvc.CheckValidity(cert, at)
	switch validity.Status {
	case pkisvc.ValidityExpired:
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnExpiredCert, nil))
	case pkisvc.ValidityNotYetValid:
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnNotYetValid, nil))
	}

	if info != nil && info.SelfSigned {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnSelfSignedCert, nil))
	}

	return warnings
}
