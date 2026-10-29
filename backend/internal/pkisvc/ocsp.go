package pkisvc

import (
	"time"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// OCSP 哈希算法（上游只认小写名）。
const (
	// OCSPHashSHA1 是 SHA-1（上游默认）。
	OCSPHashSHA1 = "sha1"
	// OCSPHashSHA256 是 SHA-256。
	OCSPHashSHA256 = "sha256"
	// OCSPHashSM3 是 SM3。
	OCSPHashSM3 = "sm3"
)

// OCSPHashes 返回受支持的 OCSP 哈希算法（顺序稳定）。
//
// 注意：OCSP 的 CertID 哈希**只**支持这三种，
// 与证书指纹的六种算法不是同一集合。
func OCSPHashes() []string {
	return []string{OCSPHashSHA1, OCSPHashSHA256, OCSPHashSM3}
}

// OCSP 证书状态码。
const (
	// OCSPCertGood 表示证书正常。
	OCSPCertGood = 0
	// OCSPCertRevoked 表示证书已吊销。
	OCSPCertRevoked = 1
	// OCSPCertUnknown 表示响应者不认识该证书。
	OCSPCertUnknown = 2
)

// OCSPRequestResult 是 OCSP 请求的生成结果。
type OCSPRequestResult struct {
	// Hash 是实际使用的哈希算法名。
	Hash string `json:"hash"`
	// DER 是请求的 DER 编码（RFC 6960 网络格式）。
	DER []byte `json:"-"`
}

// CreateOCSPRequestDER 生成 OCSP 请求。
//
// hash 必须是 OCSPHashes 之一，空值按 SHA-1 处理（与上游一致）。
func CreateOCSPRequestDER(cert, issuer *Certificate, hash string) (*OCSPRequestResult, error) {
	if cert == nil || cert.cert == nil {
		return nil, toolerrInvalidCert("certificate unavailable")
	}
	if issuer == nil || issuer.cert == nil {
		return nil, toolerrInvalidCert("issuer certificate unavailable")
	}
	if hash == "" {
		hash = OCSPHashSHA1
	}
	if !containsString(OCSPHashes(), hash) {
		return nil, unsupportedAlgorithm(hash, OCSPHashes())
	}
	if !issuerMatchesCertificate(cert, issuer) {
		return nil, toolerr.Newf(toolerr.CodeKeyMismatch, map[string]any{
			"field": "issuer",
		})
	}

	der, err := x509.CreateOCSPRequest(cert.cert, issuer.cert, hash)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidCert, map[string]any{
			"reason": "cannot build ocsp request",
		})
	}
	return &OCSPRequestResult{Hash: hash, DER: der}, nil
}

// OCSPResponseInfo 是 OCSP 响应的公开快照。
type OCSPResponseInfo struct {
	// Status 是响应级状态码（0 表示 successful）。
	Status int `json:"status"`
	// StatusText 是响应级状态说明。
	StatusText string `json:"statusText"`
	// ProducedAt 是响应生成时间。
	ProducedAt time.Time `json:"producedAt"`
	// CertStatus 是目标证书状态（0 正常 / 1 已吊销 / 2 未知）。
	CertStatus int `json:"certStatus"`
	// CertStatusText 是目标证书状态说明。
	CertStatusText string `json:"certStatusText"`
	// RevocationTime 是吊销时间；未吊销时为零值。
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
	//
	// 只有调用方显式传入信任根时才会验证；
	// 未验证时**不代表**响应可信（见 CheckOCSPResponse 的说明）。
	SignatureVerified bool `json:"signatureVerified"`
}

// OCSPResponse 是 OCSP 响应句柄。
//
// 它持有原生句柄，**必须**显式 Close。
type OCSPResponse struct {
	// resp 是底层响应；Close 之后置为 nil。
	resp *x509.Response
	// info 是解析出的公开快照。
	info OCSPResponseInfo
}

// Close 释放响应句柄。
//
// **幂等**：重复调用只返回首次的结果。
func (r *OCSPResponse) Close() error {
	if r == nil || r.resp == nil {
		return nil
	}
	err := r.resp.Close()
	r.resp = nil
	return err
}

// Info 返回响应的公开快照。
func (r *OCSPResponse) Info() OCSPResponseInfo {
	if r == nil {
		return OCSPResponseInfo{}
	}
	return r.info
}

// ParseOCSPResponseDER 解析 OCSP 响应。
//
// cert 与 issuer 都必填：上游按 CertID 定位目标证书状态，
// 缺少任一都无法确定「这条响应说的是哪张证书」。
//
// 返回的句柄由调用方负责 Close。
func ParseOCSPResponseDER(der []byte, cert, issuer *Certificate) (*OCSPResponse, error) {
	if len(der) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "response"})
	}
	if cert == nil || cert.cert == nil {
		return nil, toolerrInvalidCert("certificate unavailable")
	}
	if issuer == nil || issuer.cert == nil {
		return nil, toolerrInvalidCert("issuer certificate unavailable")
	}

	resp, err := x509.ParseOCSPResponse(der, cert.cert, issuer.cert)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidDER, map[string]any{
			"field":    "response",
			"expected": []string{"OCSP response DER"},
		})
	}

	info := OCSPResponseInfo{
		Status:         resp.Status,
		StatusText:     resp.StatusText,
		ProducedAt:     resp.ProducedAt,
		CertStatus:     resp.CertStatus,
		CertStatusText: resp.CertStatusText,
		RevocationTime: resp.RevocationTime,
		ReasonCode:     resp.RevocationReason,
		ReasonText:     resp.ReasonText,
		ThisUpdate:     resp.ThisUpdate,
		NextUpdate:     resp.NextUpdate,
		ResponderCount: len(resp.ResponderCerts),
	}
	return &OCSPResponse{resp: resp, info: info}, nil
}

// Verify 用信任根校验 OCSP 响应的签名。
//
// roots 为空时**不做**验证并返回 nil：语义是「用户没要求验证」，
// 调用方据 SignatureVerified 区分「已验证通过」与「未验证」。
// 这**不**意味着未验证的响应可信。
func (r *OCSPResponse) Verify(roots []*Certificate) error {
	if r == nil || r.resp == nil {
		return toolerr.Newf(toolerr.CodeInvalidDER, map[string]any{
			"reason": "ocsp response already closed",
		})
	}
	if len(roots) == 0 {
		return nil
	}

	store := NewStore()
	defer func() { _ = store.Close() }()
	if err := store.addAll(roots); err != nil {
		return err
	}

	if err := r.resp.Verify(store.store, nil); err != nil {
		return toolerr.Newf(toolerr.CodeSignatureInvalid, map[string]any{
			"field": "response",
		})
	}
	r.info.SignatureVerified = true
	return nil
}

// issuerMatchesCertificate 判断 leaf 是否确实由 issuer 签发。
//
// 用信号校验而不是只比对名字：同名 CA 的旧密钥签出的证书不应被当成同一签发者。
func issuerMatchesCertificate(leaf, issuer *Certificate) bool {
	if leaf == nil || leaf.cert == nil || issuer == nil || issuer.cert == nil {
		return false
	}
	pub, err := issuer.cert.PublicKey()
	if err != nil {
		return false
	}
	defer func() { _ = asym.Close(pub) }()

	return leaf.cert.Verify(pub) == nil
}
