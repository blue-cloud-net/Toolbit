package pkisvc

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 签发模式。
const (
	// SignModeSelf 表示自签（签发者是证书本身）。
	SignModeSelf = "self"
	// SignModeCA 表示由 CA 签发。
	SignModeCA = "ca"
)

// SignModes 返回受支持的签发模式（顺序稳定）。
func SignModes() []string {
	return []string{SignModeSelf, SignModeCA}
}

// DNFields 返回允许出现在主体 / 签发者里的字段短名（顺序稳定）。
//
// 该白名单与 docs/standards/03-crypto.md §8.1 一致：白名单之外的字段名一律
// 返回 INVALID_FIELD，**不**交给上游的 `Add`（它会 panic）。
func DNFields() []string {
	return []string{
		"CN", "C", "O", "OU", "L", "ST",
		"serialNumber", "emailAddress", "title", "SN", "givenName",
		"initials", "generationQualifier", "pseudonym", "businessCategory", "DC",
	}
}

// BasicConstraintsOption 是 basicConstraints 扩展的参数。
type BasicConstraintsOption struct {
	// IsCA 报告是否为 CA 证书。
	IsCA bool
	// PathLen 是路径长度约束；负数表示不指定。
	PathLen int
}

// CertIssueOptions 是签发证书的参数。
//
// 本结构**不**包含签发者与签名者：两者由 IssueCertificate 的参数传入，
// 以免调用方把「主体」与「签发者」混为一谈。
type CertIssueOptions struct {
	// Subject 是主体的 RDN 条目。
	Subject []NameEntryInfo
	// Serial 是序列号（正整数）。
	Serial int64
	// NotBefore / NotAfter 是有效期窗口。
	NotBefore time.Time
	NotAfter  time.Time
	// SAN 是主体备用名条目（如 `DNS:a.example.com`、`IP:1.2.3.4`）。
	SAN []string
	// KeyUsage 是密钥用途位名。
	KeyUsage []string
	// ExtendedKeyUsage 是扩展密钥用途。
	ExtendedKeyUsage []string
	// BasicConstraints 是基本约束；nil 表示不添加该扩展。
	BasicConstraints *BasicConstraintsOption
	// AutoKeyID 报告是否自动补 SKID / AKID。
	//
	// CA 模式且为 true 时，AKID 取自签发者证书的 SKID；
	// 自签模式且为 true 时 AKID 与 SKID 相同（与 `openssl req -x509` 一致）。
	AutoKeyID bool
}

// BuildName 把 RDN 条目构造为 DN。
//
// 一律走 `x509.Name.TryAdd`：用户可控输入**禁止**用 `Add`
// （对非法字段会 panic，见 AGENTS.md §4.9）。
//
// 参数 field 是调用方给这组条目起的字段路径，只用于错误回显**字段名**，
// 不回显字段值（见 AGENTS.md §4.5）。
func BuildName(entries []NameEntryInfo, field string) (*x509.Name, error) {
	if len(entries) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": field})
	}

	name := x509.NewName()
	for i, entry := range entries {
		if err := name.TryAdd(entry.Field, entry.Value); err != nil {
			// 不把上游的英文错误与字段值透出：只回显**位置**与**白名单**。
			return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"field":    fmt.Sprintf("%s.%d", field, i),
				"expected": DNFields(),
			})
		}
	}
	return name, nil
}

// RandomSerial 生成一个正整数序列号。
//
// 取 8 字节随机数并清掉最高位，保证结果落在 int64 的正区间：
// 上游只用 `int64` 承载序列号（见 I-21），超过 8 字节的序列号无法表达。
func RandomSerial() (int64, error) {
	raw, err := cryptosvc.RandomBytes(8)
	if err != nil {
		return 0, toolerr.Newf(toolerr.CodeCryptoOperationFailed, map[string]any{
			"reason": "random source unavailable",
		})
	}
	raw[0] &= 0x7f
	serial := int64(binary.BigEndian.Uint64(raw))
	if serial == 0 {
		serial = 1
	}
	return serial, nil
}

// IssueCertificate 签发一张证书。
//
// issuer 为 nil 时按**自签**处理：主体与签发者取同一份 DN，签名者必须与
// pub 配对。issuer 非 nil 时按 **CA 签发**处理：签发者 DN 取自 issuer 证书。
//
// 实现走 `x509.NewCertificate()` 的分步构建（而不是 `CreateSelfSigned` /
// `CreateCertificate`）：后两者**立即签名**，签名之后的 `Add*` 调用会被
// 「already signed」锁拒绝，无法再写扩展。
//
// 返回的证书句柄由调用方负责 `Close`。
func IssueCertificate(opts CertIssueOptions, pub *PublicKey, signer *PrivateKey,
	issuer *Certificate) (*Certificate, error) {
	if pub == nil || pub.key == nil {
		return nil, toolerrInvalidKey("public key unavailable")
	}
	if signer == nil || signer.key == nil {
		return nil, toolerrInvalidKey("signer key unavailable")
	}
	if !opts.NotAfter.After(opts.NotBefore) {
		return nil, toolerr.Newf(toolerr.CodeNotAfterBeforeNow, map[string]any{
			"field": "notAfter",
		})
	}
	if opts.Serial <= 0 {
		return nil, toolerr.Newf(toolerr.CodeSerialInvalid, map[string]any{
			"field": "serial",
		})
	}

	subject, err := BuildName(opts.Subject, "subject")
	if err != nil {
		return nil, err
	}

	issuerName := subject
	if issuer != nil {
		if issuer.cert == nil {
			return nil, toolerrInvalidCert("issuer certificate already closed")
		}
		issuerName = issuer.cert.SubjectName()
	}

	cert := x509.NewCertificate()
	closed := false
	defer func() {
		if !closed {
			_ = cert.Close()
		}
	}()

	steps := []func() error{
		func() error { return cert.SetVersion(2) },
		func() error { return cert.SetSerial(opts.Serial) },
		func() error { return cert.SetSubject(subject) },
		func() error { return cert.SetIssuer(issuerName) },
		func() error { return cert.SetValidity(opts.NotBefore, opts.NotAfter) },
		func() error { return cert.SetPublicKey(pub.key) },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, toolerrInvalidCert("cannot build certificate")
		}
	}

	if err := addCertificateExtensions(cert, opts, issuer); err != nil {
		return nil, err
	}

	// 配对校验按模式区分：
	//   - 自签：签名私钥必须与证书公钥配对，否则产物无法验证；
	//   - CA 签发：签名私钥必须与**签发者证书**配对（叶证书公钥是独立的）。
	if issuer == nil {
		if !keyMatchesPublic(signer, pub) {
			return nil, toolerr.Newf(toolerr.CodeKeyMismatch, map[string]any{
				"field": "signer",
			})
		}
	} else if !signerMatchesCertificate(signer, issuer) {
		return nil, toolerr.Newf(toolerr.CodeKeyMismatch, map[string]any{
			"field": "issuerKey",
		})
	}

	if err := cert.Sign(signer.key); err != nil {
		return nil, toolerrInvalidCert("cannot sign certificate")
	}

	closed = true
	return &Certificate{cert: cert}, nil
}

// addCertificateExtensions 按顺序写入扩展。
//
// 必须在 `Sign` **之前**调用：上游证书在签名后锁定，
// 再调 `Add*` 会返回 "already signed"。
func addCertificateExtensions(cert *x509.Certificate, opts CertIssueOptions,
	issuer *Certificate) error {
	fail := func() error {
		return toolerrInvalidCert("cannot apply certificate extension")
	}

	if len(opts.SAN) > 0 {
		if err := cert.AddSubjectAltName(strings.Join(opts.SAN, ",")); err != nil {
			return fail()
		}
	}
	if len(opts.KeyUsage) > 0 {
		if err := cert.AddKeyUsage(strings.Join(opts.KeyUsage, ",")); err != nil {
			return fail()
		}
	}
	if len(opts.ExtendedKeyUsage) > 0 {
		if err := cert.AddExtendedKeyUsage(strings.Join(opts.ExtendedKeyUsage, ",")); err != nil {
			return fail()
		}
	}
	if opts.BasicConstraints != nil {
		if err := cert.AddBasicConstraints(opts.BasicConstraints.IsCA); err != nil {
			return fail()
		}
	}
	if !opts.AutoKeyID {
		return nil
	}
	if err := cert.AddSubjectKeyID(); err != nil {
		return fail()
	}
	if issuer != nil && issuer.cert != nil {
		if err := cert.AddAuthorityKeyID(issuer.cert); err != nil {
			return fail()
		}
		return nil
	}
	// 自签场景：AKID 与 SKID 同源，直接以自身作为签发者引用。
	if err := cert.AddAuthorityKeyID(cert); err != nil {
		return fail()
	}
	return nil
}

// keyMatchesPublic 判断私钥与公钥是否配对。
//
// 依据两侧 SPKI PEM 的逐字节比较：`asym.Match` 只接受同层句柄，
// 而公钥可能来自本层包装或上层传入的独立句柄。
func keyMatchesPublic(signer *PrivateKey, pub *PublicKey) bool {
	if signer == nil || signer.key == nil || pub == nil || pub.key == nil {
		return false
	}
	derived, err := signer.MarshalPublicKeyPEM()
	if err != nil || len(derived) == 0 {
		return false
	}
	actual, err := pub.MarshalPEM()
	if err != nil || len(actual) == 0 {
		return false
	}
	return string(derived) == string(actual)
}

// signerPublicKey 返回与私钥配对的公钥句柄。
//
// ⚠️ 返回的是**别名**：它与私钥共享同一个底层句柄，
// **不得**单独 Close（见 AGENTS.md §4.8）。
func signerPublicKey(signer *PrivateKey) asym.PublicKey {
	if signer == nil || signer.key == nil {
		return nil
	}
	return signer.key.Public()
}
