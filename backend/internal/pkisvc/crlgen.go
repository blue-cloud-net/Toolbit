package pkisvc

import (
	"time"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// CRL 吊销原因码（RFC 5280 §5.3.1）。
//
// 码值 **7 是保留值**，上游与 Toolbit 均拒绝使用
// （见 docs/standards/03-crypto.md §8.4）。
const (
	// ReasonUnspecified 表示未指明原因（码值 0）。
	ReasonUnspecified = 0
	// ReasonKeyCompromise 表示密钥泄露（码值 1）。
	ReasonKeyCompromise = 1
	// ReasonCACompromise 表示 CA 密钥泄露（码值 2）。
	ReasonCACompromise = 2
	// ReasonAffiliationChanged 表示归属变更（码值 3）。
	ReasonAffiliationChanged = 3
	// ReasonSuperseded 表示已被取代（码值 4）。
	ReasonSuperseded = 4
	// ReasonCessationOfOperation 表示停止运营（码值 5）。
	ReasonCessationOfOperation = 5
	// ReasonCertificateHold 表示临时挂起（码值 6）。
	ReasonCertificateHold = 6
	// ReasonRemoveFromCRL 表示从 CRL 中移除（码值 8）。
	ReasonRemoveFromCRL = 8
	// ReasonPrivilegeWithdrawn 表示权限被撤销（码值 9）。
	ReasonPrivilegeWithdrawn = 9
	// ReasonAACompromise 表示属性权威密钥泄露（码值 10）。
	ReasonAACompromise = 10
)

// RevocationReasons 返回受支持的吊销原因码（顺序稳定，**不含**保留值 7）。
func RevocationReasons() []int {
	return []int{
		ReasonUnspecified,
		ReasonKeyCompromise,
		ReasonCACompromise,
		ReasonAffiliationChanged,
		ReasonSuperseded,
		ReasonCessationOfOperation,
		ReasonCertificateHold,
		ReasonRemoveFromCRL,
		ReasonPrivilegeWithdrawn,
		ReasonAACompromise,
	}
}

// ReasonNames 返回原因码到 OpenSSL 短名的映射。
//
// 未登记的码值在 map 中不存在，调用方据此判断「未知原因」。
func ReasonNames() map[int]string {
	return map[int]string{
		ReasonUnspecified:          "unspecified",
		ReasonKeyCompromise:        "keyCompromise",
		ReasonCACompromise:         "cACompromise",
		ReasonAffiliationChanged:   "affiliationChanged",
		ReasonSuperseded:           "superseded",
		ReasonCessationOfOperation: "cessationOfOperation",
		ReasonCertificateHold:      "certificateHold",
		ReasonRemoveFromCRL:        "removeFromCRL",
		ReasonPrivilegeWithdrawn:   "privilegeWithdrawn",
		ReasonAACompromise:         "aACompromise",
	}
}

// CRLRevokeItem 是一条待写入 CRL 的吊销记录。
//
// 上游 `CRLBuilder.Revoke` **只接受证书对象**，没有「按序列号吊销」的入口，
// 因此本结构不提供 `Serial` 字段（见 I-20）。
type CRLRevokeItem struct {
	// Cert 是被吊销的证书；必填。
	Cert *Certificate
	// RevokedAt 是吊销生效时间；零值时用 CRL 的 thisUpdate。
	RevokedAt time.Time
	// Reason 是原因码；必须属于 RevocationReasons。
	Reason int
}

// CRLGenOptions 是签发 CRL 的参数。
type CRLGenOptions struct {
	// Number 是 CRL Number；0 表示自动生成。
	Number int64
	// ThisUpdate 是本次更新时间；零值时用当前时间。
	ThisUpdate time.Time
	// NextUpdate 是下次更新时间；零值表示不写该字段。
	NextUpdate time.Time
}

// CRLGenResult 是签发 CRL 的结果。
type CRLGenResult struct {
	// PEM 是 CRL 的 PEM 文本。
	PEM []byte
	// DER 是 CRL 的 DER 二进制。
	DER []byte
	// Info 是签发结果的公开元信息（与 crl-parse 的字段一致）。
	Info *CRLInfo
}

// CreateCRL 用 CA 证书与私钥签发一张吊销列表。
//
// issuer 必须是 CA 证书且带 SKID（CRL 的 AKID 取自该扩展）；
// 签发者私钥必须与证书配对。
func CreateCRL(issuer *Certificate, signer *PrivateKey, opts CRLGenOptions,
	items []CRLRevokeItem) (*CRLGenResult, error) {
	if issuer == nil || issuer.cert == nil {
		return nil, toolerrInvalidCert("issuer certificate unavailable")
	}
	if signer == nil || signer.key == nil {
		return nil, toolerrInvalidKey("signer key unavailable")
	}
	if len(items) > MaxCRLEntries {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "revoked",
			"limit":  MaxCRLEntries,
			"actual": len(items),
		})
	}
	if !issuerHasKeyID(issuer) {
		return nil, toolerrInvalidCert("issuer certificate has no subject key identifier")
	}
	if !signerMatchesCertificate(signer, issuer) {
		return nil, toolerr.Newf(toolerr.CodeKeyMismatch, map[string]any{
			"field": "issuerKey",
		})
	}

	thisUpdate := opts.ThisUpdate
	if thisUpdate.IsZero() {
		thisUpdate = time.Now()
	}
	if !opts.NextUpdate.IsZero() && !opts.NextUpdate.After(thisUpdate) {
		return nil, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
			"field":    "nextUpdate",
			"expected": "later than thisUpdate",
		})
	}

	builder, err := x509.NewCRLBuilder(issuer.cert)
	if err != nil {
		return nil, toolerrInvalidCRL("cannot create revocation list builder")
	}
	signed := false
	defer func() {
		if !signed {
			_ = builder.Close()
		}
	}()

	number := opts.Number
	if number == 0 {
		generated, genErr := RandomSerial()
		if genErr != nil {
			return nil, genErr
		}
		number = generated
	}
	if err := builder.SetNumber(number); err != nil {
		return nil, toolerrInvalidCRL("cannot set revocation list number")
	}
	if err := builder.SetThisUpdate(thisUpdate); err != nil {
		return nil, toolerrInvalidCRL("cannot set this update")
	}
	if !opts.NextUpdate.IsZero() {
		if err := builder.SetNextUpdate(opts.NextUpdate); err != nil {
			return nil, toolerrInvalidCRL("cannot set next update")
		}
	}

	for i := range items {
		item := items[i]
		if item.Cert == nil || item.Cert.cert == nil {
			return nil, toolerrInvalidCert("revoked certificate unavailable")
		}
		if !containsInt(RevocationReasons(), item.Reason) {
			return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"field":    "revoked.reasonCode",
				"expected": RevocationReasons(),
			})
		}
		at := item.RevokedAt
		if at.IsZero() {
			at = thisUpdate
		}
		if err := builder.Revoke(item.Cert.cert, at, x509.RevocationReason(item.Reason)); err != nil {
			return nil, toolerrInvalidCRL("cannot add revoked certificate")
		}
	}

	crl, err := builder.Sign(signer.key)
	if err != nil {
		return nil, toolerrInvalidCRL("cannot sign revocation list")
	}
	signed = true
	defer func() { _ = crl.Close() }()

	pemBytes, err := crl.MarshalPEM()
	if err != nil {
		return nil, toolerrInvalidCRL("cannot export revocation list")
	}
	derBytes, err := crl.MarshalDER()
	if err != nil {
		return nil, toolerrInvalidCRL("cannot export revocation list")
	}

	return &CRLGenResult{
		PEM: pemBytes,
		DER: derBytes,
		Info: &CRLInfo{
			Issuer:                nameEntries(crl.IssuerEntries()),
			IssuerText:            crl.IssuerText(),
			ThisUpdate:            crl.LastUpdate(),
			NextUpdate:            crl.NextUpdate(),
			SignatureAlgorithm:    crl.SignatureAlgorithm(),
			SignatureAlgorithmOID: crl.SignatureAlgorithmOID(),
			Number:                crl.Number(),
			Revoked:               revokedItems(crl.RevokedEntries()),
		},
	}, nil
}

// revokedItems 把上游的吊销条目转成本层的公开形态。
func revokedItems(entries []x509.RevokedEntry) []RevokedItem {
	out := make([]RevokedItem, 0, len(entries))
	for _, entry := range entries {
		out = append(out, RevokedItem{
			Serial:     formatSerial(entry.Serial),
			RevokedAt:  entry.RevocationDate,
			ReasonCode: entry.ReasonCode,
			ReasonText: entry.Reason,
		})
	}
	return out
}

// issuerHasKeyID 判断 CA 证书是否带 SKID 扩展。
//
// 上游 CRL 构建器要求签发者带 SKID（CRL 的 AKID 取自该扩展），
// 缺失时提前给出明确的错误，而不是让底层在签发阶段失败。
func issuerHasKeyID(issuer *Certificate) bool {
	if issuer == nil || issuer.cert == nil {
		return false
	}
	return len(issuer.cert.SubjectKeyID()) > 0
}

// signerMatchesCertificate 判断私钥与证书公钥是否配对。
func signerMatchesCertificate(signer *PrivateKey, cert *Certificate) bool {
	if signer == nil || signer.key == nil || cert == nil || cert.cert == nil {
		return false
	}
	pub, err := cert.cert.PublicKey()
	if err != nil {
		return false
	}
	defer func() { _ = asym.Close(pub) }()

	derived, err := signer.MarshalPublicKeyPEM()
	if err != nil || len(derived) == 0 {
		return false
	}
	actual, err := pub.MarshalPublicKeyPEM()
	if err != nil || len(actual) == 0 {
		return false
	}
	return string(derived) == string(actual)
}
