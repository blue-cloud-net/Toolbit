package pkisvc

import (
	"encoding/hex"
	"strings"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/digest"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// CSRGenOptions 是生成证书签名请求的参数。
type CSRGenOptions struct {
	// Subject 是主体的 RDN 条目。
	Subject []NameEntryInfo
	// SAN 是主体备用名条目（如 `DNS:a.example.com`）。
	SAN []string
	// KeyUsage 是密钥用途位名。
	KeyUsage []string
	// ExtendedKeyUsage 是扩展密钥用途。
	ExtendedKeyUsage []string
	// BasicConstraints 是基本约束；nil 表示不添加该扩展。
	BasicConstraints *BasicConstraintsOption
	// ChallengePassword 是挑战口令；为空表示不添加。
	//
	// 口令值**不会**出现在响应里（只用 ChallengePasswordPresent 报告存在性）。
	ChallengePassword string
}

// CSRGenResult 是生成证书签名请求的结果。
type CSRGenResult struct {
	// PEM 是 CSR 的 PEM 文本。
	PEM []byte
	// Info 是刚生成的 CSR 的公开元信息（与 csr-parse 的字段一致）。
	Info *CSRInfo
}

// CreateCSR 用给定私钥生成证书签名请求。
//
// 走 `NewEmptyCertificateRequest` 的分步构建：上游的 `NewCertificateRequest`
// 是「立即签名」的便捷版，签名之后再加扩展会让请求内容与签名不匹配。
//
// 返回的 PEM 已自签名，可通过 ParseCSR 的 `signatureValid` 复核。
func CreateCSR(opts CSRGenOptions, signer *PrivateKey) (*CSRGenResult, error) {
	if signer == nil || signer.key == nil {
		return nil, toolerrInvalidKey("signer key unavailable")
	}

	subject, err := BuildName(opts.Subject, "subject")
	if err != nil {
		return nil, err
	}

	pub := signerPublicKey(signer)
	if pub == nil {
		return nil, toolerrInvalidKey("public key unavailable")
	}

	req := x509.NewEmptyCertificateRequest()
	if err := req.SetSubject(subject); err != nil {
		return nil, toolerrInvalidCSR("cannot set subject")
	}
	if err := req.SetPublicKey(pub); err != nil {
		return nil, toolerrInvalidCSR("cannot set public key")
	}
	if opts.ChallengePassword != "" {
		if err := req.SetChallengePassword(opts.ChallengePassword); err != nil {
			return nil, toolerrInvalidCSR("cannot set challenge password")
		}
	}
	if err := addCSRExtensions(req, opts); err != nil {
		return nil, err
	}
	if err := req.Sign(signer.key); err != nil {
		return nil, toolerrInvalidCSR("cannot sign certificate request")
	}

	pemBytes, err := req.MarshalPEM()
	if err != nil {
		return nil, toolerrInvalidCSR("cannot export certificate request")
	}

	info, err := describeCSR(req, pub)
	if err != nil {
		return nil, err
	}

	return &CSRGenResult{PEM: pemBytes, Info: info}, nil
}

// addCSRExtensions 按顺序写入 CSR 扩展（须在 Sign 之前调用）。
func addCSRExtensions(req *x509.CertificateRequest, opts CSRGenOptions) error {
	fail := func() error {
		return toolerrInvalidCSR("cannot apply request extension")
	}

	if len(opts.SAN) > 0 {
		if err := req.AddSubjectAltName(strings.Join(opts.SAN, ",")); err != nil {
			return fail()
		}
	}
	if len(opts.KeyUsage) > 0 {
		if err := req.AddKeyUsage(strings.Join(opts.KeyUsage, ",")); err != nil {
			return fail()
		}
	}
	if len(opts.ExtendedKeyUsage) > 0 {
		if err := req.AddExtendedKeyUsage(strings.Join(opts.ExtendedKeyUsage, ",")); err != nil {
			return fail()
		}
	}
	if opts.BasicConstraints != nil {
		if err := req.AddBasicConstraints(opts.BasicConstraints.IsCA); err != nil {
			return fail()
		}
	}
	return nil
}

// describeCSR 直接从已构建的请求生成元信息。
//
// 这里**不**把 PEM 重新解析一遍：每次解析都会多持有一个上游 CSR 句柄，
// 而该句柄**没有**公开的 Close（见 I-15），重复解析等于放大泄漏面。
func describeCSR(req *x509.CertificateRequest, pub asym.PublicKey) (*CSRInfo, error) {
	info := &CSRInfo{
		Subject:                  nameEntries(req.SubjectEntries()),
		SubjectText:              req.SubjectText(),
		SAN:                      sanFromExtensions(req.Extensions()),
		SignatureAlgorithm:       req.SignatureAlgorithm(),
		SignatureAlgorithmOID:    req.SignatureAlgorithmOID(),
		Extensions:               extensionInfos(req.Extensions()),
		ChallengePasswordPresent: req.ChallengePassword() != "",
	}

	// 刚由本函数签名，校验通过才算产物可用。
	if err := req.Verify(); err == nil {
		info.SignatureValid = true
	}

	der, err := req.MarshalDER()
	if err != nil {
		return nil, toolerrInvalidCSR("cannot export certificate request")
	}
	sum, err := digest.Sum("SHA256", der)
	if err != nil {
		return nil, toolerrInvalidCSR("cannot compute fingerprint")
	}
	info.FingerprintSHA256 = hex.EncodeToString(sum)

	info.PublicKey = describeUpstreamPublicKey(pub)
	return info, nil
}

// PublicKeyFromCSR 取出证书签名请求里的公钥。
//
// 返回的是**独立持有**的句柄（上游 `CertificateRequest.PublicKey()` 已做 Dup），
// 调用方必须显式 Close。
func PublicKeyFromCSR(source []byte, isPEM bool) (*PublicKey, error) {
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
	pub, err := csr.PublicKey()
	if err != nil {
		return nil, toolerrInvalidCSR("cannot read public key")
	}
	return &PublicKey{key: pub}, nil
}
