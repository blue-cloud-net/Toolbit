package pki

import (
	"context"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CRLParseRequest 是吊销列表解析工具的请求。
type CRLParseRequest struct {
	// Input 是待解析的吊销列表。
	Input PemBundle `json:"input"`
}

// CRLParseResult 是吊销列表解析工具的响应。
type CRLParseResult struct {
	// Issuer 是签发者的 RDN 条目。
	Issuer []NameEntryOutput `json:"issuer"`
	// IssuerText 是签发者的单行文本。
	IssuerText string `json:"issuerText"`
	// ThisUpdate 是生效时间。
	ThisUpdate time.Time `json:"thisUpdate"`
	// NextUpdate 是下次更新时间；CRL 未给出时为零值。
	NextUpdate time.Time `json:"nextUpdate"`
	// SignatureAlgorithm 是签名算法短名。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SignatureAlgorithmOid 是签名算法的点分 OID。
	SignatureAlgorithmOid string `json:"signatureAlgorithmOid"`
	// Number 是 CRL Number 扩展值；无该扩展为 -1。
	Number int64 `json:"number"`
	// Revoked 是吊销条目列表。
	Revoked []RevokedItemOutput `json:"revoked"`
}

// RevokedItemOutput 是一条吊销记录的输出形态。
type RevokedItemOutput struct {
	// Serial 是被吊销证书的序列号（大写十六进制）。
	Serial string `json:"serial"`
	// RevokedAt 是吊销生效时间。
	RevokedAt time.Time `json:"revokedAt"`
	// ReasonCode 是原因码；-1 表示未指定。
	ReasonCode int `json:"reasonCode"`
	// ReasonText 是原因名（如 keyCompromise）。
	ReasonText string `json:"reasonText"`
}

// Validate 做字段级校验。
func (r *CRLParseRequest) Validate() error {
	return r.Input.Validate("input")
}

// CRLParse 解析证书吊销列表。
func CRLParse(_ context.Context, req *CRLParseRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Input.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	info, err := pkisvc.ParseCRLInfo(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}

	result := &CRLParseResult{
		Issuer:                nameEntriesOutput(info.Issuer),
		IssuerText:            info.IssuerText,
		ThisUpdate:            info.ThisUpdate,
		NextUpdate:            info.NextUpdate,
		SignatureAlgorithm:    info.SignatureAlgorithm,
		SignatureAlgorithmOid: info.SignatureAlgorithmOID,
		Number:                info.Number,
		Revoked:               revokedOutput(info.Revoked),
	}

	return tools.Result{Data: result}, nil
}

// revokedOutput 把吊销条目转成输出形态。
func revokedOutput(items []pkisvc.RevokedItem) []RevokedItemOutput {
	out := make([]RevokedItemOutput, 0, len(items))
	for _, item := range items {
		out = append(out, RevokedItemOutput{
			Serial:     item.Serial,
			RevokedAt:  item.RevokedAt,
			ReasonCode: item.ReasonCode,
			ReasonText: item.ReasonText,
		})
	}
	return out
}
