package pkisvc

import (
	"encoding/hex"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/digest"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// CSRInfo 是证书签名请求的公开元信息快照。
//
// ⚠️ 上游的 `x509.CertificateRequest` **没有公开的 Close 方法**，
// 其原生句柄只能靠 runtime finalizer 兜底释放。这是
// docs/standards/03-crypto.md §3「句柄必须显式释放」的**唯一书面豁免**，
// 已登记于 docs/upstream/tongsuo-go-issues.md I-15。
type CSRInfo struct {
	// Subject 是主体的 RDN 条目。
	Subject []NameEntryInfo `json:"subject"`
	// SubjectText 是主体的单行文本。
	SubjectText string `json:"subjectText"`
	// SAN 是主体备用名条目。
	SAN []string `json:"sans"`
	// PublicKey 是公钥元信息。
	PublicKey PublicKeyInfo `json:"publicKey"`
	// SignatureAlgorithm 是签名算法短名。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SignatureAlgorithmOID 是签名算法的点分 OID。
	SignatureAlgorithmOID string `json:"signatureAlgorithmOid"`
	// Extensions 是请求扩展。
	Extensions []ExtInfo `json:"extensions"`
	// FingerprintSHA256 是 CSR 的 SHA-256 指纹（算在 DER 上，小写十六进制）。
	FingerprintSHA256 string `json:"fingerprintSha256"`
	// ChallengePasswordPresent 报告是否携带挑战口令。
	//
	// **只报存在性**：口令值**绝不**回显（见 AGENTS.md §4.5）。
	ChallengePasswordPresent bool `json:"challengePasswordPresent"`
	// SignatureValid 报告 CSR 自签名是否校验通过。
	SignatureValid bool `json:"signatureValid"`
}

// ParseCSR 解析证书签名请求。
//
// source 是 PEM 文本或 DER 二进制；isPEM 报告它是哪一种。
func ParseCSR(source []byte, isPEM bool) (*CSRInfo, error) {
	if len(source) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "input.data"})
	}

	var (
		csr *x509.CertificateRequest
		err error
	)
	if isPEM {
		csr, err = x509.LoadCertificateRequestPEM(source)
	} else {
		csr, err = x509.LoadCertificateRequestDER(source)
	}
	if err != nil {
		return nil, toolerrInvalidCSR("cannot parse certificate request")
	}

	info := &CSRInfo{
		Subject:                  nameEntries(csr.SubjectEntries()),
		SubjectText:              csr.SubjectText(),
		SAN:                      sanFromExtensions(csr.Extensions()),
		SignatureAlgorithm:       csr.SignatureAlgorithm(),
		SignatureAlgorithmOID:    csr.SignatureAlgorithmOID(),
		Extensions:               extensionInfos(csr.Extensions()),
		ChallengePasswordPresent: csr.ChallengePassword() != "",
	}

	// 自签名校验失败**不**让整个解析失败：它是「结论」，不是「输入非法」。
	if verifyErr := csr.Verify(); verifyErr == nil {
		info.SignatureValid = true
	}

	der, err := csr.MarshalDER()
	if err != nil {
		return nil, toolerrInvalidCSR("cannot export certificate request")
	}
	sum, err := digest.Sum("SHA256", der)
	if err != nil {
		return nil, toolerrInvalidCSR("cannot compute fingerprint")
	}
	info.FingerprintSHA256 = hex.EncodeToString(sum)

	pub, err := csr.PublicKey()
	if err != nil {
		return nil, toolerrInvalidCSR("cannot read public key")
	}
	defer func() { _ = asym.Close(pub) }()

	keyInfo := PublicKeyInfo{
		Algorithm: algorithmFromUpstream(pub.Algorithm().String()),
	}
	params, err := asym.Params(pub)
	if err == nil && params != nil {
		keyInfo.Bits = rsaBits(params)
		keyInfo.Curve = curveFromUpstream(params.Curve)
	}
	info.PublicKey = keyInfo

	return info, nil
}

// VerifyCSR 校验证书签名请求的自签名。
//
// 通过返回 nil；不通过返回 SIGNATURE_INVALID。
func VerifyCSR(source []byte, isPEM bool) error {
	if len(source) == 0 {
		return toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "input.data"})
	}

	var (
		csr *x509.CertificateRequest
		err error
	)
	if isPEM {
		csr, err = x509.LoadCertificateRequestPEM(source)
	} else {
		csr, err = x509.LoadCertificateRequestDER(source)
	}
	if err != nil {
		return toolerrInvalidCSR("cannot parse certificate request")
	}

	if verifyErr := csr.Verify(); verifyErr != nil {
		return toolerr.Newf(toolerr.CodeSignatureInvalid, map[string]any{
			"reason": "csr signature verification failed",
		})
	}
	return nil
}
