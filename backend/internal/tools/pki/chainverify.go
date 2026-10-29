package pki

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ChainVerifyRequest 是证书链验证工具的请求。
type ChainVerifyRequest struct {
	// Bundle 是证书链材料（可含多张证书）。
	Bundle PemBundle `json:"bundle"`
	// TrustedRoots 是受信任根证书；为空时用链内的自签名根作隐含信任锚。
	TrustedRoots []PemBundle `json:"trustedRoots"`
	// AtTime 是验证时刻（RFC3339）；为空时按当前时间。
	AtTime string `json:"atTime"`
	// Host 是要校验的主机名；为空时不做名字匹配。
	Host string `json:"host"`
}

// ChainVerifyResult 是证书链验证工具的响应。
type ChainVerifyResult struct {
	// Valid 报告整链是否可信。
	//
	// **验证不通过是业务结果**：返回 200 + valid: false。
	Valid bool `json:"valid"`
	// Errors 是失败原因列表。
	Errors []VerifyErrorOutput `json:"errors"`
	// Details 是逐张证书的结论。
	Details []CertDetailOutput `json:"details"`
	// HostnameMatch 报告主机名是否匹配。
	HostnameMatch bool `json:"hostnameMatch"`
	// HostnameError 是主机名不匹配的原因。
	HostnameError string `json:"hostnameError"`
	// ImplicitTrust 报告是否使用了「链内自签名根」作为隐含信任锚。
	ImplicitTrust bool `json:"implicitTrust"`
}

// VerifyErrorOutput 是链验证失败原因的输出形态。
type VerifyErrorOutput struct {
	// Code 是 X509_V_ERR_* 错误码。
	Code string `json:"code"`
	// Message 是错误说明。
	Message string `json:"message"`
	// Depth 是出错深度（0 为待验证证书本身）。
	Depth int `json:"depth"`
}

// CertDetailOutput 是链上一张证书的验证结论。
type CertDetailOutput struct {
	// Subject 是主体文本。
	Subject string `json:"subject"`
	// Issuer 是签发者文本。
	Issuer string `json:"issuer"`
	// Valid 报告这张证书本身是否通过验证。
	Valid bool `json:"valid"`
	// Reason 是失败原因（通过时为空）。
	Reason string `json:"reason"`
}

// Validate 做字段级校验。
func (r *ChainVerifyRequest) Validate() error {
	if err := r.Bundle.Validate("bundle"); err != nil {
		return err
	}
	for i := range r.TrustedRoots {
		if err := r.TrustedRoots[i].Validate("trustedRoots"); err != nil {
			return err
		}
	}
	return nil
}

// ChainVerify 验证证书链。
//
// **没有系统信任根可用**（上游未暴露默认信任库，见 I-17）：
// trustedRoots 为空时把链内的自签名根当作隐含信任锚，
// 此时 valid: true 只说明**链内自洽**，不代表被外部信任
// （响应中的 implicitTrust 会标注这一点）。
func ChainVerify(_ context.Context, req *ChainVerifyRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	at, err := parseAtTime(req.AtTime)
	if err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Bundle.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	certs, err := pkisvc.ParseCertificates(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(certs)

	roots, err := loadCertificates(req.TrustedRoots)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(roots)

	// 链内证书按「叶子 → 根」排序：首张是叶证书，其余是中间证书。
	ordered, err := pkisvc.OrderChain(certs)
	if err != nil {
		return tools.Result{}, err
	}

	leaf, intermediates := pkisvc.SplitChain(ordered)
	if leaf == nil {
		return tools.Result{}, toolerr.Newf(toolerr.CodeInvalidCert, map[string]any{
			"reason": "no certificate in bundle",
		})
	}

	out, err := pkisvc.VerifyChain(leaf, intermediates, roots, at, req.Host)
	if err != nil {
		return tools.Result{}, err
	}

	result := &ChainVerifyResult{
		Valid:         out.Valid,
		Errors:        verifyErrorsOutput(out.Errors),
		Details:       certDetailsOutput(out.Details),
		HostnameMatch: out.HostnameMatch,
		HostnameError: out.HostnameError,
		ImplicitTrust: out.ImplicitTrust,
	}

	return tools.Result{Data: result, Warnings: verifyWarnings(out)}, nil
}

// loadCertificates 加载一组证书材料。
func loadCertificates(bundles []PemBundle) ([]*pkisvc.Certificate, error) {
	var out []*pkisvc.Certificate
	for _, bundle := range bundles {
		source, isPEM, err := bundle.Resolve()
		if err != nil {
			closeAll(out)
			return nil, err
		}
		certs, err := pkisvc.ParseCertificates(source, isPEM)
		if err != nil {
			closeAll(out)
			return nil, err
		}
		out = append(out, certs...)
	}
	return out, nil
}

// verifyErrorsOutput 把失败原因转成输出形态。
func verifyErrorsOutput(errs []pkisvc.VerifyErrorItem) []VerifyErrorOutput {
	out := make([]VerifyErrorOutput, 0, len(errs))
	for _, item := range errs {
		out = append(out, VerifyErrorOutput{
			Code:    item.Code,
			Message: item.Message,
			Depth:   item.Depth,
		})
	}
	return out
}

// certDetailsOutput 把逐张证书的结论转成输出形态。
func certDetailsOutput(details []pkisvc.CertDetail) []CertDetailOutput {
	out := make([]CertDetailOutput, 0, len(details))
	for _, detail := range details {
		out = append(out, CertDetailOutput{
			Subject: detail.Subject,
			Issuer:  detail.Issuer,
			Valid:   detail.Valid,
			Reason:  detail.Reason,
		})
	}
	return out
}

// verifyWarnings 汇总与验证结果相关的警告。
//
// UNTRUSTED_CHAIN 只在**验证失败**时给出：
// 隐含信任锚（链内自签名根）本身不是「不可信」，
// 它的语义由响应里的 implicitTrust 字段表达。
func verifyWarnings(out *pkisvc.VerifyResult) []toolerr.Warning {
	var warnings []toolerr.Warning

	for _, item := range out.Errors {
		switch item.Code {
		case "X509_V_ERR_CERT_HAS_EXPIRED":
			warnings = append(warnings, toolerr.NewWarning(toolerr.WarnExpiredCert, nil))
		case "X509_V_ERR_CERT_NOT_YET_VALID":
			warnings = append(warnings, toolerr.NewWarning(toolerr.WarnNotYetValid, nil))
		}
	}

	if !out.Valid {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnUntrustedChain, nil))
	}

	return warnings
}
