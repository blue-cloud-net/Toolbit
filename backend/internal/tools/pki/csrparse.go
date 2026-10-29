package pki

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CSRParseRequest 是证书签名请求解析工具的请求。
type CSRParseRequest struct {
	// Input 是待解析的 CSR。
	Input PemBundle `json:"input"`
}

// CSRParseResult 是证书签名请求解析工具的响应。
type CSRParseResult struct {
	// Subject 是主体的 RDN 条目。
	Subject []NameEntryOutput `json:"subject"`
	// SubjectText 是主体的单行文本。
	SubjectText string `json:"subjectText"`
	// Sans 是主体备用名条目。
	Sans []string `json:"sans"`
	// PublicKey 是公钥元信息。
	PublicKey PublicKeyOutput `json:"publicKey"`
	// SignatureAlgorithm 是签名算法短名。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SignatureAlgorithmOid 是签名算法的点分 OID。
	SignatureAlgorithmOid string `json:"signatureAlgorithmOid"`
	// Extensions 是请求扩展。
	Extensions []ExtensionOutput `json:"extensions"`
	// FingerprintSha256 是 CSR 的 SHA-256 指纹。
	FingerprintSha256 string `json:"fingerprintSha256"`
	// ChallengePasswordPresent 报告是否携带挑战口令。
	//
	// **只报存在性**：口令值绝不回显（见 AGENTS.md §4.5）。
	ChallengePasswordPresent bool `json:"challengePasswordPresent"`
	// SignatureValid 报告 CSR 自签名是否校验通过。
	SignatureValid bool `json:"signatureValid"`
}

// Validate 做字段级校验。
func (r *CSRParseRequest) Validate() error {
	return r.Input.Validate("input")
}

// CSRParse 解析一张证书签名请求。
func CSRParse(_ context.Context, req *CSRParseRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Input.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	info, err := pkisvc.ParseCSR(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}

	result := &CSRParseResult{
		Subject:                  nameEntriesOutput(info.Subject),
		SubjectText:              info.SubjectText,
		Sans:                     info.SAN,
		PublicKey:                publicKeyOutput(info.PublicKey),
		SignatureAlgorithm:       info.SignatureAlgorithm,
		SignatureAlgorithmOid:    info.SignatureAlgorithmOID,
		Extensions:               extensionsOutput(info.Extensions),
		FingerprintSha256:        info.FingerprintSHA256,
		ChallengePasswordPresent: info.ChallengePasswordPresent,
		SignatureValid:           info.SignatureValid,
	}

	return tools.Result{Data: result}, nil
}
