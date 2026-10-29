package pkisvc

import (
	"fmt"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// MaxCRLEntries 是吊销条目数的上限。
const MaxCRLEntries = 10000

// RevokedItem 是 CRL 中的一条吊销记录。
type RevokedItem struct {
	// Serial 是被吊销证书的序列号（大写十六进制）。
	Serial string `json:"serial"`
	// RevokedAt 是吊销生效时间。
	RevokedAt time.Time `json:"revokedAt"`
	// ReasonCode 是原因码；-1 表示未指定。
	ReasonCode int `json:"reasonCode"`
	// ReasonText 是原因名（如 keyCompromise）；未指定时为空。
	ReasonText string `json:"reasonText"`
}

// CRLInfo 是证书吊销列表的公开元信息快照。
type CRLInfo struct {
	// Issuer 是签发者的 RDN 条目。
	Issuer []NameEntryInfo `json:"issuer"`
	// IssuerText 是签发者的单行文本。
	IssuerText string `json:"issuerText"`
	// ThisUpdate 是生效时间。
	ThisUpdate time.Time `json:"thisUpdate"`
	// NextUpdate 是下次更新时间；CRL 未给出时为零值。
	NextUpdate time.Time `json:"nextUpdate"`
	// SignatureAlgorithm 是签名算法短名。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SignatureAlgorithmOID 是签名算法的点分 OID。
	SignatureAlgorithmOID string `json:"signatureAlgorithmOid"`
	// Number 是 CRL Number 扩展值；无该扩展为 -1。
	Number int64 `json:"number"`
	// Revoked 是吊销条目列表。
	Revoked []RevokedItem `json:"revoked"`
}

// ParseCRLInfo 解析证书吊销列表。
//
// source 是 PEM 文本或 DER 二进制；isPEM 报告它是哪一种。
// 吊销条目超过 MaxCRLEntries 时返回 TOO_MANY_ITEMS。
func ParseCRLInfo(source []byte, isPEM bool) (*CRLInfo, error) {
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
	defer func() { _ = crl.Close() }()

	entries := crl.RevokedEntries()
	if len(entries) > MaxCRLEntries {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "input.data",
			"limit":  MaxCRLEntries,
			"actual": len(entries),
		})
	}

	info := &CRLInfo{
		Issuer:                nameEntries(crl.IssuerEntries()),
		IssuerText:            crl.IssuerText(),
		ThisUpdate:            crl.LastUpdate(),
		NextUpdate:            crl.NextUpdate(),
		SignatureAlgorithm:    crl.SignatureAlgorithm(),
		SignatureAlgorithmOID: crl.SignatureAlgorithmOID(),
		Number:                crl.Number(),
		Revoked:               make([]RevokedItem, 0, len(entries)),
	}

	for _, entry := range entries {
		info.Revoked = append(info.Revoked, RevokedItem{
			Serial:     formatSerial(entry.Serial),
			RevokedAt:  entry.RevocationDate,
			ReasonCode: entry.ReasonCode,
			ReasonText: entry.Reason,
		})
	}

	return info, nil
}

// formatSerial 把序列号渲染为大写十六进制。
//
// ⚠️ 上游用 `int64` 表达吊销条目的序列号
// （`ASN1_INTEGER_get` 在长度超过 8 字节时返回 -1）。
// 因此超过 64 位的序列号**读不出来**：
// 此时返回空串，**不**把 -1 当成真实序列号展示
// （见 docs/upstream/tongsuo-go-issues.md I-18）。
func formatSerial(serial int64) string {
	if serial < 0 {
		return ""
	}
	return fmt.Sprintf("%X", serial)
}
