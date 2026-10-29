package pki

import (
	"context"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CRLGenRequest 是吊销列表生成工具的请求。
type CRLGenRequest struct {
	// IssuerCert 是签发者证书（须为 CA 且带 SKID）。
	IssuerCert PemBundle `json:"issuerCert"`
	// IssuerKey 是签发者私钥。
	IssuerKey KeyInput `json:"issuerKey"`
	// Number 是 CRL 编号（十六进制）；为空时自动生成。
	Number string `json:"number"`
	// ThisUpdate / NextUpdate 是有效期（RFC3339）。
	ThisUpdate string `json:"thisUpdate"`
	NextUpdate string `json:"nextUpdate"`
	// Revoked 是待吊销的证书列表。
	//
	// ⚠️ 只接受**证书**：上游 `CRLBuilder.Revoke` 没有「按序列号吊销」的入口
	// （见 docs/upstream/tongsuo-go-issues.md I-20）。
	Revoked []CRLRevokeInput `json:"revoked"`
}

// CRLRevokeInput 是一条吊销项的输入形态。
type CRLRevokeInput struct {
	// Cert 是被吊销的证书。
	Cert PemBundle `json:"cert"`
	// RevokedAt 是吊销生效时间（RFC3339）；为空时用 thisUpdate。
	RevokedAt string `json:"revokedAt"`
	// ReasonCode 是吊销原因码。
	ReasonCode int `json:"reasonCode"`
}

// CRLGenResult 是吊销列表生成工具的响应。
type CRLGenResult struct {
	// CrlPem 是 CRL 的 PEM。
	CrlPem string `json:"crlPem"`
	// CrlDerBase64 是 CRL 的 DER（base64）。
	CrlDerBase64 string `json:"crlDerBase64"`
	// IssuerText 是签发者的单行文本。
	IssuerText string `json:"issuerText"`
	// Number 是 CRL 编号。
	Number int64 `json:"number"`
	// ThisUpdate / NextUpdate 是有效期。
	ThisUpdate time.Time `json:"thisUpdate"`
	NextUpdate time.Time `json:"nextUpdate"`
	// RevokedCount 是写入的吊销条目数。
	RevokedCount int `json:"revokedCount"`
}

// Validate 做字段级校验。
func (r *CRLGenRequest) Validate() error {
	if err := r.IssuerCert.Validate("issuerCert"); err != nil {
		return err
	}
	if err := r.IssuerKey.Validate("issuerKey"); err != nil {
		return err
	}
	if len(r.Revoked) > pkisvc.MaxCRLEntries {
		return tools.InvalidField("revoked", []int{pkisvc.MaxCRLEntries}, len(r.Revoked))
	}
	for i := range r.Revoked {
		if err := r.Revoked[i].Cert.Validate("revoked.cert"); err != nil {
			return err
		}
		if !containsInt(pkisvc.RevocationReasons(), r.Revoked[i].ReasonCode) {
			return tools.InvalidField("revoked.reasonCode",
				pkisvc.RevocationReasons(), r.Revoked[i].ReasonCode)
		}
	}
	return nil
}

// CRLGen 签发证书吊销列表。
func CRLGen(_ context.Context, req *CRLGenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	issuerCert, err := loadPemCertificate(req.IssuerCert)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = issuerCert.Close() }()

	issuerKey, err := loadPemPrivateKey(req.IssuerKey)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = issuerKey.Close() }()

	items, err := loadRevokeItems(req.Revoked)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(certificatesOf(items))

	number, err := resolveSerial(req.Number)
	if err != nil {
		return tools.Result{}, err
	}
	if req.Number == "" {
		number = 0 // 交给服务层自动生成
	}

	thisUpdate, err := parseAtTime(req.ThisUpdate)
	if err != nil {
		return tools.Result{}, err
	}
	nextUpdate, err := parseAtTime(req.NextUpdate)
	if err != nil {
		return tools.Result{}, err
	}

	out, err := pkisvc.CreateCRL(issuerCert, issuerKey, pkisvc.CRLGenOptions{
		Number:     number,
		ThisUpdate: thisUpdate,
		NextUpdate: nextUpdate,
	}, items)
	if err != nil {
		return tools.Result{}, err
	}

	result := &CRLGenResult{
		CrlPem:       string(out.PEM),
		CrlDerBase64: encodeDER(out.DER),
		IssuerText:   out.Info.IssuerText,
		Number:       out.Info.Number,
		ThisUpdate:   out.Info.ThisUpdate,
		NextUpdate:   out.Info.NextUpdate,
		RevokedCount: len(out.Info.Revoked),
	}
	return tools.Result{Data: result}, nil
}

// loadRevokeItems 加载待吊销的证书列表。
func loadRevokeItems(inputs []CRLRevokeInput) ([]pkisvc.CRLRevokeItem, error) {
	out := make([]pkisvc.CRLRevokeItem, 0, len(inputs))
	for i := range inputs {
		cert, err := loadPemCertificate(inputs[i].Cert)
		if err != nil {
			closeAll(certificatesOf(out))
			return nil, err
		}
		at, err := parseAtTime(inputs[i].RevokedAt)
		if err != nil {
			_ = cert.Close()
			closeAll(certificatesOf(out))
			return nil, err
		}
		out = append(out, pkisvc.CRLRevokeItem{
			Cert:      cert,
			RevokedAt: at,
			Reason:    inputs[i].ReasonCode,
		})
	}
	return out, nil
}

// certificatesOf 取出吊销项里的证书句柄，供统一释放。
func certificatesOf(items []pkisvc.CRLRevokeItem) []*pkisvc.Certificate {
	out := make([]*pkisvc.Certificate, 0, len(items))
	for _, item := range items {
		if item.Cert != nil {
			out = append(out, item.Cert)
		}
	}
	return out
}

// loadPemCertificate 从 PemBundle 加载一张证书。
func loadPemCertificate(bundle PemBundle) (*pkisvc.Certificate, error) {
	source, isPEM, err := bundle.Resolve()
	if err != nil {
		return nil, err
	}
	return pkisvc.LoadCertificate(source, isPEM)
}

// loadPemPrivateKey 从 KeyInput 加载私钥。
func loadPemPrivateKey(input KeyInput) (*pkisvc.PrivateKey, error) {
	source, isPEM, err := input.Resolve()
	if err != nil {
		return nil, err
	}
	if !isPEM {
		return nil, tools.UnsupportedFormat("der", []string{"pem"})
	}
	return pkisvc.LoadPrivateKey(source, input.Passphrase)
}
