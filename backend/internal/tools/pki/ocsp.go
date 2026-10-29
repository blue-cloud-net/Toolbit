package pki

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// --- ocsp-request ---

// OCSPRequestInput 是 OCSP 请求生成工具的请求。
type OCSPRequestInput struct {
	// Cert 是待查询的证书。
	Cert PemBundle `json:"cert"`
	// Issuer 是签发者证书。
	Issuer PemBundle `json:"issuer"`
	// Hash 是 CertID 的哈希算法（sha1 / sha256 / sm3）；为空时用 sha1。
	Hash string `json:"hash"`
}

// OCSPRequestOutput 是 OCSP 请求生成工具的响应。
type OCSPRequestOutput struct {
	// Hash 是实际使用的哈希算法。
	Hash string `json:"hash"`
	// DerBase64 是请求的 DER（base64）。
	DerBase64 string `json:"derBase64"`
	// DerHex 是请求的 DER（小写十六进制）。
	DerHex string `json:"derHex"`
	// Length 是字节长度。
	Length int `json:"length"`
}

// Validate 做字段级校验。
func (r *OCSPRequestInput) Validate() error {
	if err := r.Cert.Validate("cert"); err != nil {
		return err
	}
	if err := r.Issuer.Validate("issuer"); err != nil {
		return err
	}
	if r.Hash != "" && !containsString(pkisvc.OCSPHashes(), r.Hash) {
		return tools.UnsupportedAlgorithm(r.Hash, pkisvc.OCSPHashes())
	}
	return nil
}

// OCSPRequest 生成 OCSP 请求。
func OCSPRequest(_ context.Context, req *OCSPRequestInput) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	cert, err := loadPemCertificate(req.Cert)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = cert.Close() }()

	issuer, err := loadPemCertificate(req.Issuer)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = issuer.Close() }()

	out, err := pkisvc.CreateOCSPRequestDER(cert, issuer, req.Hash)
	if err != nil {
		return tools.Result{}, err
	}

	result := &OCSPRequestOutput{
		Hash:      out.Hash,
		DerBase64: encodeDER(out.DER),
		DerHex:    hex.EncodeToString(out.DER),
		Length:    len(out.DER),
	}
	return tools.Result{Data: result}, nil
}

// --- ocsp-response ---

// OCSPResponseInput 是 OCSP 响应解析工具的请求。
type OCSPResponseInput struct {
	// Response 是 OCSP 响应（PEM / DER / base64 / hex 自动识别）。
	Response PemBundle `json:"response"`
	// Cert 是被查询的证书。
	Cert PemBundle `json:"cert"`
	// Issuer 是签发者证书。
	Issuer PemBundle `json:"issuer"`
	// TrustedRoots 是响应签名的信任根；为空时**不验签**。
	TrustedRoots []PemBundle `json:"trustedRoots"`
}

// OCSPResponseOutput 是 OCSP 响应解析工具的响应。
type OCSPResponseOutput struct {
	// Status 是响应级状态码。
	Status int `json:"status"`
	// StatusText 是响应级状态说明。
	StatusText string `json:"statusText"`
	// ProducedAt 是响应生成时间。
	ProducedAt time.Time `json:"producedAt"`
	// CertStatus 是目标证书状态：good / revoked / unknown。
	CertStatus string `json:"certStatus"`
	// RevocationTime 是吊销时间。
	RevocationTime time.Time `json:"revocationTime"`
	// ReasonCode 是吊销原因码；无原因时为 -1。
	ReasonCode int `json:"reasonCode"`
	// ReasonText 是吊销原因名。
	ReasonText string `json:"reasonText"`
	// ThisUpdate / NextUpdate 是该状态的有效窗口。
	ThisUpdate time.Time `json:"thisUpdate"`
	NextUpdate time.Time `json:"nextUpdate"`
	// ResponderCount 是响应内嵌的响应者证书数量。
	ResponderCount int `json:"responderCount"`
	// SignatureVerified 报告响应签名是否已验证。
	SignatureVerified bool `json:"signatureVerified"`
}

// Validate 做字段级校验。
func (r *OCSPResponseInput) Validate() error {
	if err := r.Response.Validate("response"); err != nil {
		return err
	}
	if err := r.Cert.Validate("cert"); err != nil {
		return err
	}
	if err := r.Issuer.Validate("issuer"); err != nil {
		return err
	}
	for i := range r.TrustedRoots {
		if err := r.TrustedRoots[i].Validate("trustedRoots"); err != nil {
			return err
		}
	}
	return nil
}

// OCSPResponse 解析 OCSP 响应。
func OCSPResponse(_ context.Context, req *OCSPResponseInput) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Response.Resolve()
	if err != nil {
		return tools.Result{}, err
	}
	der, err := pkisvc.NormalizeToDER(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}

	cert, err := loadPemCertificate(req.Cert)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = cert.Close() }()

	issuer, err := loadPemCertificate(req.Issuer)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = issuer.Close() }()

	roots, err := loadCertificates(req.TrustedRoots)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(roots)

	resp, err := pkisvc.ParseOCSPResponseDER(der, cert, issuer)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = resp.Close() }()

	if err := resp.Verify(roots); err != nil {
		return tools.Result{}, err
	}

	info := resp.Info()
	result := &OCSPResponseOutput{
		Status:            info.Status,
		StatusText:        info.StatusText,
		ProducedAt:        info.ProducedAt,
		CertStatus:        ocspCertStatusName(info.CertStatus),
		RevocationTime:    info.RevocationTime,
		ReasonCode:        info.ReasonCode,
		ReasonText:        info.ReasonText,
		ThisUpdate:        info.ThisUpdate,
		NextUpdate:        info.NextUpdate,
		ResponderCount:    info.ResponderCount,
		SignatureVerified: info.SignatureVerified,
	}

	return tools.Result{Data: result, Warnings: ocspWarnings(info)}, nil
}

// ocspCertStatusName 把证书状态码转成可读名。
func ocspCertStatusName(status int) string {
	switch status {
	case pkisvc.OCSPCertGood:
		return "good"
	case pkisvc.OCSPCertRevoked:
		return "revoked"
	case pkisvc.OCSPCertUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// ocspWarnings 汇总 OCSP 解析的警告。
func ocspWarnings(info pkisvc.OCSPResponseInfo) []toolerr.Warning {
	if info.SignatureVerified {
		return nil
	}
	return []toolerr.Warning{{
		Code:   toolerr.WarnCRLUnverified,
		Params: map[string]any{"status": ocspCertStatusName(info.CertStatus)},
	}}
}
