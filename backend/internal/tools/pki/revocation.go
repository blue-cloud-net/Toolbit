package pki

import (
	"context"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// RevocationCheckRequest 是吊销状态检查工具的请求。
type RevocationCheckRequest struct {
	// Cert 是待检查的证书。
	Cert PemBundle `json:"cert"`
	// Crl 是吊销列表。
	Crl PemBundle `json:"crl"`
	// Issuer 是签发者证书（可选）。
	//
	// 提供时**先验签再判定**；缺省时只确认「未被收录」，
	// 状态最多给到 unverified，**绝不**返回 good
	// （见 docs/standards/03-crypto.md §8.4）。
	Issuer *PemBundle `json:"issuer"`
	// AtTime 是判定时刻（RFC3339）；为空时按当前时间。
	AtTime string `json:"atTime"`
}

// RevocationCheckResult 是吊销状态检查工具的响应。
type RevocationCheckResult struct {
	// Status 取值：good / revoked / unknown / unverified。
	Status string `json:"status"`
	// SignatureVerified 报告 CRL 签名是否已验证。
	SignatureVerified bool `json:"signatureVerified"`
	// Serial 是证书序列号（大写十六进制）。
	Serial string `json:"serial"`
	// RevokedAt 是吊销时间；未吊销时为零值。
	RevokedAt time.Time `json:"revokedAt"`
	// ReasonCode 是吊销原因码；未吊销为 -1。
	ReasonCode int `json:"reasonCode"`
	// ReasonText 是吊销原因名。
	ReasonText string `json:"reasonText"`
	// CrlNumber 是 CRL 编号。
	CrlNumber int64 `json:"crlNumber"`
	// CrlThisUpdate / CrlNextUpdate 是 CRL 的时间窗。
	CrlThisUpdate time.Time `json:"crlThisUpdate"`
	CrlNextUpdate time.Time `json:"crlNextUpdate"`
}

// Validate 做字段级校验。
func (r *RevocationCheckRequest) Validate() error {
	if err := r.Cert.Validate("cert"); err != nil {
		return err
	}
	if err := r.Crl.Validate("crl"); err != nil {
		return err
	}
	if r.Issuer != nil {
		if err := r.Issuer.Validate("issuer"); err != nil {
			return err
		}
	}
	return nil
}

// RevocationCheck 判定证书是否被吊销。
func RevocationCheck(_ context.Context, req *RevocationCheckRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	at, err := parseAtTime(req.AtTime)
	if err != nil {
		return tools.Result{}, err
	}

	cert, err := loadPemCertificate(req.Cert)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = cert.Close() }()

	crlSource, crlIsPEM, err := req.Crl.Resolve()
	if err != nil {
		return tools.Result{}, err
	}
	crl, err := pkisvc.LoadRevocationList(crlSource, crlIsPEM)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = crl.Close() }()

	var issuer *pkisvc.Certificate
	if req.Issuer != nil && req.Issuer.Data != "" {
		issuer, err = loadPemCertificate(*req.Issuer)
		if err != nil {
			return tools.Result{}, err
		}
		defer func() { _ = issuer.Close() }()
	}

	out, err := pkisvc.CheckRevocation(cert, crl, issuer, at)
	if err != nil {
		return tools.Result{}, err
	}

	result := &RevocationCheckResult{
		Status:            out.Status,
		SignatureVerified: out.SignatureVerified,
		Serial:            out.Serial,
		RevokedAt:         out.RevokedAt,
		ReasonCode:        out.ReasonCode,
		ReasonText:        out.ReasonText,
		CrlNumber:         out.CRLNumber,
		CrlThisUpdate:     out.CRLThisUpdate,
		CrlNextUpdate:     out.CRLNextUpdate,
	}

	return tools.Result{Data: result, Warnings: revocationWarnings(out)}, nil
}

// revocationWarnings 汇总吊销判定的警告。
//
// 未验签时**必须**给出警告：否则用户会把「未被收录」误读成「确认有效」。
func revocationWarnings(out *pkisvc.RevocationResult) []toolerr.Warning {
	if out == nil || out.SignatureVerified {
		return nil
	}
	return []toolerr.Warning{{
		Code: toolerr.WarnCRLUnverified,
		Params: map[string]any{
			"status": out.Status,
		},
	}}
}
