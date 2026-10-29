package pkisvc

import (
	"time"

	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 吊销状态。
const (
	// RevocationGood 表示 CRL 未收录该证书，且 CRL 签名已验证。
	RevocationGood = "good"
	// RevocationRevoked 表示证书已被吊销。
	RevocationRevoked = "revoked"
	// RevocationUnknown 表示 CRL 与证书不匹配，无法判定。
	RevocationUnknown = "unknown"
	// RevocationUnverified 表示 CRL 签名**未**验证，只确认了「未被吊销」。
	RevocationUnverified = "unverified"
)

// RevocationStatuses 返回受支持的吊销状态（顺序稳定）。
func RevocationStatuses() []string {
	return []string{RevocationGood, RevocationRevoked, RevocationUnknown, RevocationUnverified}
}

// RevocationList 是 CRL 句柄。
//
// 它持有原生句柄，**必须**显式 Close（见 docs/standards/03-crypto.md §3）。
type RevocationList struct {
	// crl 是底层吊销列表；Close 之后置为 nil。
	crl *x509.CRL
}

// LoadRevocationList 加载一张吊销列表。
//
// source 是 PEM 文本或 DER 二进制；isPEM 报告它是哪一种。
func LoadRevocationList(source []byte, isPEM bool) (*RevocationList, error) {
	if len(source) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "input.data"})
	}

	var (
		crl *x509.CRL
		err error
	)
	if isPEM {
		crl, err = x509.LoadCRLPEM(source)
	} else {
		crl, err = x509.LoadCRLDER(source)
	}
	if err != nil {
		return nil, toolerrInvalidCRL("cannot parse revocation list")
	}
	return &RevocationList{crl: crl}, nil
}

// Close 释放吊销列表句柄。
//
// **幂等**：重复调用只返回首次的结果。
func (r *RevocationList) Close() error {
	if r == nil || r.crl == nil {
		return nil
	}
	err := r.crl.Close()
	r.crl = nil
	return err
}

// MarshalPEM 导出吊销列表 PEM。
func (r *RevocationList) MarshalPEM() ([]byte, error) {
	if r == nil || r.crl == nil {
		return nil, toolerrInvalidCRL("revocation list already closed")
	}
	out, err := r.crl.MarshalPEM()
	if err != nil {
		return nil, toolerrInvalidCRL("cannot export revocation list")
	}
	return out, nil
}

// MarshalDER 导出吊销列表 DER。
func (r *RevocationList) MarshalDER() ([]byte, error) {
	if r == nil || r.crl == nil {
		return nil, toolerrInvalidCRL("revocation list already closed")
	}
	out, err := r.crl.MarshalDER()
	if err != nil {
		return nil, toolerrInvalidCRL("cannot export revocation list")
	}
	return out, nil
}

// Info 读取吊销列表的元信息。
func (r *RevocationList) Info() (*CRLInfo, error) {
	if r == nil || r.crl == nil {
		return nil, toolerrInvalidCRL("revocation list already closed")
	}
	entries := r.crl.RevokedEntries()
	if len(entries) > MaxCRLEntries {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "input.data",
			"limit":  MaxCRLEntries,
			"actual": len(entries),
		})
	}
	return &CRLInfo{
		Issuer:                nameEntries(r.crl.IssuerEntries()),
		IssuerText:            r.crl.IssuerText(),
		ThisUpdate:            r.crl.LastUpdate(),
		NextUpdate:            r.crl.NextUpdate(),
		SignatureAlgorithm:    r.crl.SignatureAlgorithm(),
		SignatureAlgorithmOID: r.crl.SignatureAlgorithmOID(),
		Number:                r.crl.Number(),
		Revoked:               revokedItems(entries),
	}, nil
}

// Verify 用签发者证书校验 CRL 的签名。
func (r *RevocationList) Verify(issuer *Certificate) error {
	if r == nil || r.crl == nil {
		return toolerrInvalidCRL("revocation list already closed")
	}
	if issuer == nil || issuer.cert == nil {
		return toolerrInvalidCert("issuer certificate unavailable")
	}
	if err := r.crl.Verify(issuer.cert); err != nil {
		return toolerrInvalidCRL("revocation list signature mismatch")
	}
	return nil
}

// RevocationResult 是吊销检查的结论。
type RevocationResult struct {
	// Status 取值见 RevocationGood 等常量。
	Status string `json:"status"`
	// SignatureVerified 报告 CRL 签名是否已验证。
	SignatureVerified bool `json:"signatureVerified"`
	// RevokedAt 是吊销生效时间；未吊销时为零值。
	RevokedAt time.Time `json:"revokedAt"`
	// ReasonCode 是原因码；未吊销或未收录时为 -1。
	ReasonCode int `json:"reasonCode"`
	// ReasonText 是原因名；未吊销或未收录时为空。
	ReasonText string `json:"reasonText"`
	// Serial 是证书序列号（大写十六进制）。
	Serial string `json:"serial"`
	// CRLNumber 是 CRL Number；无该扩展为 -1。
	CRLNumber int64 `json:"crlNumber"`
	// CRLThisUpdate / CRLNextUpdate 是 CRL 的时间窗。
	CRLThisUpdate time.Time `json:"crlThisUpdate"`
	CRLNextUpdate time.Time `json:"crlNextUpdate"`
}

// CheckRevocation 判定证书是否被 CRL 吊销。
//
// **顺序是契约的一部分**：给了 issuer 时先 `CRL.Verify`（签名 + 与签发者
// 匹配），验签失败直接返回 INVALID_CRL，**不**再给出任何吊销结论
// （见 docs/standards/03-crypto.md §8.4）。
//
// 未提供 issuer 时不验签：此时**绝不**返回 good，
// 最多给出 revoked（有正面证据）或 unverified（无结论）。
//
// at 为零值时按当前时间判断 CRL 是否在有效窗口内。
func CheckRevocation(cert *Certificate, crl *RevocationList, issuer *Certificate,
	at time.Time) (*RevocationResult, error) {
	if cert == nil || cert.cert == nil {
		return nil, toolerrInvalidCert("certificate unavailable")
	}
	if crl == nil || crl.crl == nil {
		return nil, toolerrInvalidCRL("revocation list unavailable")
	}

	if at.IsZero() {
		at = time.Now()
	}

	thisUpdate := crl.crl.LastUpdate()
	nextUpdate := crl.crl.NextUpdate()
	if !thisUpdate.IsZero() && at.Before(thisUpdate) {
		return nil, toolerrInvalidCRL("revocation list is not yet valid")
	}
	// 过期的 CRL 不能作为吊销依据（无论结论是正是负）。
	if !nextUpdate.IsZero() && at.After(nextUpdate) {
		return nil, toolerrInvalidCRL("revocation list is stale")
	}

	result := &RevocationResult{
		Status:        RevocationUnknown,
		Serial:        cert.cert.SerialHex(),
		ReasonCode:    -1,
		CRLNumber:     crl.crl.Number(),
		CRLThisUpdate: thisUpdate,
		CRLNextUpdate: nextUpdate,
	}

	if issuer != nil {
		if err := crl.Verify(issuer); err != nil {
			return nil, err
		}
		result.SignatureVerified = true
	}

	revoked := crl.crl.IsRevoked(cert.cert)
	if revoked {
		result.Status = RevocationRevoked
		if entry, ok := findRevokedEntry(crl, cert); ok {
			result.RevokedAt = entry.RevocationDate
			result.ReasonCode = entry.ReasonCode
			result.ReasonText = entry.Reason
		}
		return result, nil
	}

	if !result.SignatureVerified {
		result.Status = RevocationUnverified
		return result, nil
	}
	result.Status = RevocationGood
	return result, nil
}

// findRevokedEntry 在 CRL 中定位某张证书的吊销条目。
//
// 上游用 `int64` 承载序列号，超过 8 字节时读到 -1（见 I-18）：
// 此时无法按序列号比对，只返回 false —— 吊销结论仍由
// `CRL.IsRevoked` 给出，不因为读不到时间stamp就改判。
func findRevokedEntry(crl *RevocationList, cert *Certificate) (x509.RevokedEntry, bool) {
	if crl == nil || crl.crl == nil || cert == nil || cert.cert == nil {
		return x509.RevokedEntry{}, false
	}
	serial := cert.cert.Serial()
	if serial <= 0 {
		return x509.RevokedEntry{}, false
	}
	for _, entry := range crl.crl.RevokedEntries() {
		if entry.Serial == serial {
			return entry, true
		}
	}
	return x509.RevokedEntry{}, false
}
