package pkisvc

import (
	"errors"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/x509"
)

// 主机名匹配的结论。
const (
	// HostnameMismatch 表示主机名与证书不匹配。
	HostnameMismatch = "hostname_mismatch"
	// HostnameInvalidCertificate 表示证书已失效，无法做名字匹配。
	HostnameInvalidCertificate = "invalid_certificate"
)

// VerifyErrorItem 是链验证的一条错误。
type VerifyErrorItem struct {
	// Code 是 X509_V_ERR_* 错误码，形如 "X509_V_ERR_CERT_HAS_EXPIRED"。
	Code string `json:"code"`
	// Message 是错误说明（**不含**用户输入）。
	Message string `json:"message"`
	// Depth 是出错深度（0 为待验证证书本身）。
	Depth int `json:"depth"`
}

// CertDetail 是链上一张证书的验证结论。
type CertDetail struct {
	// Subject 是主体文本。
	Subject string `json:"subject"`
	// Issuer 是签发者文本。
	Issuer string `json:"issuer"`
	// Valid 报告这张证书本身是否通过验证。
	Valid bool `json:"valid"`
	// Reason 是失败原因（通过时为空）。
	Reason string `json:"reason"`
}

// VerifyResult 是证书链验证的结果。
//
// **验证不通过是业务结果**，不是错误：调用方应返回 200 + valid: false。
type VerifyResult struct {
	// Valid 报告整链是否可信。
	Valid bool `json:"valid"`
	// Errors 是失败原因列表。
	Errors []VerifyErrorItem `json:"errors"`
	// Details 是逐张证书的结论。
	Details []CertDetail `json:"details"`
	// HostnameMatch 报告主机名是否匹配。
	//
	// 未指定 host 时恒为 false（无从匹配）。
	HostnameMatch bool `json:"hostnameMatch"`
	// HostnameError 是主机名不匹配的原因；未指定 host 或匹配成功时为空。
	HostnameError string `json:"hostnameError"`
	// ImplicitTrust 报告是否使用了「链内自签名根」作为隐含信任锚。
	//
	// 为真表示 trustedRoots 为空：服务端**没有**系统信任根可用
	// （上游未暴露默认信任库，见 I-17），
	// 此时「验证通过」只说明**链内自洽**，不代表被外部信任。
	ImplicitTrust bool `json:"implicitTrust"`
}

// VerifyChain 验证证书链。
//
// leaf 是待验证的叶证书；intermediates 是中间证书；roots 是受信任根。
//
// 规则：
//
//   - **每请求新建** x509.Store（SetTime 会影响该 store 的后续所有验证）；
//   - roots 为空时，把链内的**自签名证书**作为隐含信任锚，
//     并置 ImplicitTrust 为真（没有系统信任根可用，见 I-17）；
//   - at 非零时才 SetTime（回拨验证时刻会让已过期证书通过，只在必要时使用）；
//   - host 非空时做名字匹配（**只**匹配名字，不验证链）。
func VerifyChain(
	leaf *Certificate,
	intermediates []*Certificate,
	roots []*Certificate,
	at time.Time,
	host string,
) (*VerifyResult, error) {
	if leaf == nil || leaf.cert == nil {
		return nil, toolerrInvalidCert("certificate already closed")
	}

	store := x509.NewStore()
	defer func() { _ = store.Close() }()

	result := &VerifyResult{
		Errors:  []VerifyErrorItem{},
		Details: []CertDetail{},
	}

	implicitRoot := false
	if len(roots) > 0 {
		for _, root := range roots {
			if root == nil || root.cert == nil {
				continue
			}
			if err := store.AddCert(root.cert); err != nil {
				return nil, toolerrInvalidCert("cannot add trusted root")
			}
		}
	} else {
		implicitRoot = true
		for _, candidate := range append([]*Certificate{leaf}, intermediates...) {
			if candidate == nil || candidate.cert == nil {
				continue
			}
			if candidate.cert.SubjectText() != candidate.cert.IssuerText() {
				continue
			}
			if err := store.AddCert(candidate.cert); err != nil {
				return nil, toolerrInvalidCert("cannot add trusted root")
			}
			break
		}
	}
	result.ImplicitTrust = implicitRoot

	if !at.IsZero() {
		if err := store.SetTime(at); err != nil {
			return nil, toolerrInvalidCert("cannot set verification time")
		}
	}

	interCerts := make([]*x509.Certificate, 0, len(intermediates))
	for _, inter := range intermediates {
		if inter == nil || inter.cert == nil {
			continue
		}
		interCerts = append(interCerts, inter.cert)
	}

	chain, err := x509.ChainVerify(leaf.cert, store, interCerts)
	if err != nil {
		var verifyErr *x509.VerifyError
		if !errors.As(err, &verifyErr) {
			return nil, toolerrInvalidCert("chain verification failed")
		}
		item := VerifyErrorItem{
			Code:    verifyErrorCode(verifyErr.Code),
			Message: verifyErrorMessage(verifyErr.Code),
			Depth:   verifyErr.Depth,
		}
		result.Errors = append(result.Errors, item)
		result.Details = buildDetails(chain, leaf, intermediates, verifyErr.Depth, item.Message)
	} else {
		result.Valid = true
		result.Details = buildDetails(chain, leaf, intermediates, -1, "")
	}

	if host != "" {
		if hostnameErr := leaf.cert.VerifyHostname(host); hostnameErr != nil {
			result.HostnameError = HostnameMismatch
		} else {
			result.HostnameMatch = true
		}
	}

	return result, nil
}

// buildDetails 构造逐张证书的验证结论。
//
// failedDepth 为 -1 表示整体通过。
func buildDetails(
	chain []*x509.Certificate,
	leaf *Certificate,
	intermediates []*Certificate,
	failedDepth int,
	reason string,
) []CertDetail {
	source := chain
	if len(source) == 0 {
		source = nil
		if leaf != nil && leaf.cert != nil {
			source = append(source, leaf.cert)
		}
		for _, inter := range intermediates {
			if inter != nil && inter.cert != nil {
				source = append(source, inter.cert)
			}
		}
	}

	details := make([]CertDetail, 0, len(source))
	for i, cert := range source {
		detail := CertDetail{
			Subject: cert.SubjectText(),
			Issuer:  cert.IssuerText(),
			Valid:   true,
		}
		if i == failedDepth {
			detail.Valid = false
			detail.Reason = reason
		}
		details = append(details, detail)
	}
	return details
}

// verifyErrorCode 把 X509_V_ERR_* 数值码转成可读名称。
func verifyErrorCode(code int) string {
	if name, ok := verifyErrorNames[code]; ok {
		return name
	}
	return "X509_V_ERR_UNSPECIFIED"
}

// verifyErrorMessage 把 X509_V_ERR_* 数值码转成固定说明。
//
// 说明文本是**固定枚举**：不取底层库的英文串，
// 避免把用户输入带进响应（见 AGENTS.md §4.5）。
func verifyErrorMessage(code int) string {
	if msg, ok := verifyErrorMessages[code]; ok {
		return msg
	}
	return "certificate chain verification failed"
}

// verifyErrorNames 是 X509_V_ERR_* 数值码到名称的映射。
//
// 数值取自 Tongsuo 的 x509_vfy.h（与 OpenSSL 3 一致）。
var verifyErrorNames = map[int]string{
	1:  "X509_V_ERR_UNSPECIFIED",
	2:  "X509_V_ERR_UNABLE_TO_GET_ISSUER_CERT",
	3:  "X509_V_ERR_UNABLE_TO_GET_CRL",
	4:  "X509_V_ERR_UNABLE_TO_DECRYPT_CERT_SIGNATURE",
	5:  "X509_V_ERR_UNABLE_TO_DECRYPT_CRL_SIGNATURE",
	6:  "X509_V_ERR_UNABLE_TO_DECODE_ISSUER_PUBLIC_KEY",
	7:  "X509_V_ERR_CERT_SIGNATURE_FAILURE",
	8:  "X509_V_ERR_CRL_SIGNATURE_FAILURE",
	9:  "X509_V_ERR_CERT_NOT_YET_VALID",
	10: "X509_V_ERR_CERT_HAS_EXPIRED",
	11: "X509_V_ERR_CRL_NOT_YET_VALID",
	12: "X509_V_ERR_CRL_HAS_EXPIRED",
	13: "X509_V_ERR_ERROR_IN_CERT_NOT_BEFORE_FIELD",
	14: "X509_V_ERR_ERROR_IN_CERT_NOT_AFTER_FIELD",
	15: "X509_V_ERR_ERROR_IN_CRL_LAST_UPDATE_FIELD",
	16: "X509_V_ERR_ERROR_IN_CRL_NEXT_UPDATE_FIELD",
	18: "X509_V_ERR_DEPTH_ZERO_SELF_SIGNED_CERT",
	19: "X509_V_ERR_SELF_SIGNED_CERT_IN_CHAIN",
	20: "X509_V_ERR_UNABLE_TO_GET_ISSUER_CERT_LOCALLY",
	21: "X509_V_ERR_UNABLE_TO_VERIFY_LEAF_SIGNATURE",
	22: "X509_V_ERR_CERT_CHAIN_TOO_LONG",
	23: "X509_V_ERR_CERT_REVOKED",
	24: "X509_V_ERR_NO_ISSUER_PUBLIC_KEY",
	25: "X509_V_ERR_PATH_LENGTH_EXCEEDED",
	26: "X509_V_ERR_INVALID_PURPOSE",
	27: "X509_V_ERR_CERT_UNTRUSTED",
	28: "X509_V_ERR_CERT_REJECTED",
	29: "X509_V_ERR_SUBJECT_ISSUER_MISMATCH",
	30: "X509_V_ERR_AKID_SKID_MISMATCH",
	31: "X509_V_ERR_AKID_ISSUER_SERIAL_MISMATCH",
	32: "X509_V_ERR_KEYUSAGE_NO_CERTSIGN",
	33: "X509_V_ERR_UNABLE_TO_GET_CRL_ISSUER",
	34: "X509_V_ERR_UNHANDLED_CRITICAL_EXTENSION",
	35: "X509_V_ERR_KEYUSAGE_NO_CRL_SIGN",
	36: "X509_V_ERR_UNHANDLED_CRITICAL_CRL_EXTENSION",
	37: "X509_V_ERR_INVALID_NON_CA",
	39: "X509_V_ERR_KEYUSAGE_NO_DIGITAL_SIGNATURE",
	41: "X509_V_ERR_INVALID_EXTENSION",
	42: "X509_V_ERR_INVALID_POLICY_EXTENSION",
	62: "X509_V_ERR_HOSTNAME_MISMATCH",
	79: "X509_V_ERR_INVALID_CA",
}

// verifyErrorMessages 是 X509_V_ERR_* 数值码到固定说明的映射。
//
// 说明文本是**固定枚举**：不取底层库的英文串，
// 避免把用户输入带进响应（见 AGENTS.md §4.5）。
var verifyErrorMessages = map[int]string{
	1:  "unspecified verification error",
	2:  "unable to get issuer certificate",
	3:  "unable to get certificate revocation list",
	4:  "unable to decrypt certificate signature",
	5:  "unable to decrypt revocation list signature",
	6:  "unable to decode issuer public key",
	7:  "certificate signature failure",
	8:  "revocation list signature failure",
	9:  "certificate is not yet valid",
	10: "certificate has expired",
	11: "revocation list is not yet valid",
	12: "revocation list has expired",
	13: "error in certificate notBefore field",
	14: "error in certificate notAfter field",
	15: "error in revocation list lastUpdate field",
	16: "error in revocation list nextUpdate field",
	18: "self-signed certificate at depth zero",
	19: "self-signed certificate in chain",
	20: "unable to get local issuer certificate",
	21: "unable to verify the first certificate",
	22: "certificate chain is too long",
	23: "certificate has been revoked",
	24: "no issuer public key",
	25: "path length constraint exceeded",
	26: "unsupported certificate purpose",
	27: "certificate is not trusted",
	28: "certificate is rejected",
	29: "subject and issuer do not match",
	30: "authority key identifier and subject key identifier do not match",
	31: "authority key identifier and issuer serial number do not match",
	32: "key usage does not permit certificate signing",
	33: "unable to get revocation list issuer certificate",
	34: "unhandled critical extension",
	35: "key usage does not permit revocation list signing",
	36: "unhandled critical revocation list extension",
	37: "invalid non-CA certificate",
	39: "key usage does not permit digital signature",
	41: "invalid extension",
	42: "invalid policy extension",
	62: "hostname mismatch",
	79: "invalid certificate authority",
}
