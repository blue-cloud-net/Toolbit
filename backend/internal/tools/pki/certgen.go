package pki

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 默认有效期（天）。
const defaultCertDays = 365

// CertGenRequest 是证书签发工具的请求。
type CertGenRequest struct {
	// SignMode 是签发模式（self / ca）。
	SignMode string `json:"signMode"`
	// Issuer 是签发者证书与私钥；ca 模式必填。
	Issuer *IssuerInput `json:"issuer"`
	// Subject 是主体 DN；self 模式必填（用 CSR 时忽略）。
	Subject SubjectInput `json:"subject"`
	// Csr 是待签发的证书签名请求。
	Csr *PemBundle `json:"csr"`
	// TemplateCert 是模板证书。
	//
	// ⚠️ 只复用它的**公钥与有效期**：上游没有「按 DER 添加扩展」的入口
	// （读证书只拿得到扩展 DER，没有 OpenSSL 配置串），
	// 因此扩展**不会**被拷贝（见 docs/upstream/tongsuo-go-issues.md I-19）。
	TemplateCert *PemBundle `json:"templateCert"`
	// SubjectOverride 覆盖主体字段（逐项覆盖，不是整体替换）。
	SubjectOverride *SubjectInput `json:"subjectOverride"`
	// Key 是密钥生成参数。
	Key KeyGenInput `json:"key"`
	// ExistingPrivateKey 是已有私钥。
	ExistingPrivateKey *KeyInput `json:"existingPrivateKey"`
	// NotAfterDays 是有效期天数。
	NotAfterDays int `json:"notAfterDays"`
	// NotAfter 是绝对失效时间（RFC3339），优先于 NotAfterDays。
	NotAfter string `json:"notAfter"`
	// Serial 是序列号（十六进制）。
	Serial string `json:"serial"`
	// Sans / KeyUsages / ExtKeyUsages / BasicConstraints 是扩展。
	Sans             []string               `json:"sans"`
	KeyUsages        []string               `json:"keyUsages"`
	ExtKeyUsages     []string               `json:"extKeyUsages"`
	BasicConstraints *BasicConstraintsInput `json:"basicConstraints"`
}

// IssuerInput 是签发者材料的输入形态。
type IssuerInput struct {
	// Cert 是签发者证书。
	Cert PemBundle `json:"cert"`
	// Key 是签发者私钥。
	Key KeyInput `json:"key"`
}

// CertGenResult 是证书签发工具的响应。
type CertGenResult struct {
	// CertPem 是新证书的 PEM。
	CertPem string `json:"certPem"`
	// PrivateKeyPem 是新生成密钥时的私钥 PEM。
	PrivateKeyPem string `json:"privateKeyPem"`
	// ChainPem 是 ca 模式下的完整链（叶 → 根）。
	ChainPem []string `json:"chainPem"`
	// Serial 是序列号（大写十六进制）。
	Serial string `json:"serial"`
	// SubjectText / IssuerText 是主体与签发者的单行文本。
	SubjectText string `json:"subjectText"`
	IssuerText  string `json:"issuerText"`
	// NotBefore / NotAfter 是有效期。
	NotBefore time.Time `json:"notBefore"`
	NotAfter  time.Time `json:"notAfter"`
	// KeyType / KeyBits / KeyCurve 是证书公钥参数。
	KeyType  string `json:"keyType"`
	KeyBits  int    `json:"keyBits"`
	KeyCurve string `json:"keyCurve"`
}

// Validate 做字段级校验。
func (r *CertGenRequest) Validate() error {
	if r.SignMode == "" {
		return tools.MissingField("signMode")
	}
	if !containsString(pkisvc.SignModes(), r.SignMode) {
		return tools.InvalidField("signMode", pkisvc.SignModes(), r.SignMode)
	}
	if r.SignMode == pkisvc.SignModeCA && r.Issuer == nil {
		return tools.MissingField("issuer")
	}
	if r.Issuer != nil {
		if err := r.Issuer.Cert.Validate("issuer.cert"); err != nil {
			return err
		}
		if err := r.Issuer.Key.Validate("issuer.key"); err != nil {
			return err
		}
	}
	if r.Csr != nil {
		if err := r.Csr.Validate("csr"); err != nil {
			return err
		}
	}
	if r.TemplateCert != nil {
		if err := r.TemplateCert.Validate("templateCert"); err != nil {
			return err
		}
	}
	if r.SignMode == pkisvc.SignModeSelf && r.Subject.isEmpty() &&
		(r.Csr == nil || r.Csr.Data == "") {
		return tools.MissingField("subject")
	}
	if err := r.Key.Validate(); err != nil {
		return err
	}
	if r.ExistingPrivateKey != nil {
		if err := r.ExistingPrivateKey.Validate("existingPrivateKey"); err != nil {
			return err
		}
	}
	if r.NotAfterDays < 0 {
		return tools.InvalidField("notAfterDays", []int{1, 3650}, r.NotAfterDays)
	}
	return nil
}

// CertGen 签发证书（自签或由 CA 签发）。
func CertGen(_ context.Context, req *CertGenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	notBefore := time.Now()
	notAfter, err := resolveNotAfter(req, notBefore)
	if err != nil {
		return tools.Result{}, err
	}

	serial, err := resolveSerial(req.Serial)
	if err != nil {
		return tools.Result{}, err
	}

	issuerCert, issuerKey, err := loadIssuer(req)
	if err != nil {
		return tools.Result{}, err
	}
	if issuerCert != nil {
		defer func() { _ = issuerCert.Close() }()
	}
	if issuerKey != nil {
		defer func() { _ = issuerKey.Close() }()
	}

	subject, templateValidity, err := resolveSubject(req)
	if err != nil {
		return tools.Result{}, err
	}
	// 模板证书提供了有效期窗口时，以模板为准（用户没显式指定时）。
	if !templateValidity.isZero() && req.NotAfter == "" && req.NotAfterDays == 0 {
		notBefore, notAfter = templateValidity.notBefore, templateValidity.notAfter
	}

	material, err := resolveCertMaterial(req, issuerKey)
	if err != nil {
		return tools.Result{}, err
	}
	defer material.cleanup()

	cert, err := pkisvc.IssueCertificate(pkisvc.CertIssueOptions{
		Subject:          subject,
		Serial:           serial,
		NotBefore:        notBefore,
		NotAfter:         notAfter,
		SAN:              req.Sans,
		KeyUsage:         req.KeyUsages,
		ExtendedKeyUsage: req.ExtKeyUsages,
		BasicConstraints: req.BasicConstraints.option(),
		AutoKeyID:        true,
	}, material.public, material.signer, issuerCert)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = cert.Close() }()

	certPEM, err := cert.MarshalPEM()
	if err != nil {
		return tools.Result{}, err
	}

	info, err := pkisvc.Describe(cert, notBefore, nil)
	if err != nil {
		return tools.Result{}, err
	}

	result := &CertGenResult{
		CertPem:       string(certPEM),
		PrivateKeyPem: material.privatePem,
		ChainPem:      chainPEM(req.SignMode, certPEM, issuerCert),
		Serial:        info.Serial,
		SubjectText:   info.SubjectText,
		IssuerText:    info.IssuerText,
		NotBefore:     info.NotBefore,
		NotAfter:      info.NotAfter,
		KeyType:       material.info.Algorithm,
		KeyBits:       material.info.Bits,
		KeyCurve:      material.info.Curve,
	}
	return tools.Result{Data: result}, nil
}

// certMaterial 是新证书的公钥与签名者。
type certMaterial struct {
	// public 是写入新证书的公钥。
	public *pkisvc.PublicKey
	// signer 是签名用的私钥。
	signer *pkisvc.PrivateKey
	// privatePem 是新生成私钥时的 PEM 文本。
	privatePem string
	// info 是公钥参数。
	info pkisvc.PublicKeyInfo
	// cleanup 释放本次解析中新建的全部句柄。
	cleanup func()
}

// resolveCertMaterial 解析公钥与签名者。
//
// 公钥优先级：模板证书 > CSR > 自有私钥。
// 签名者：ca 模式用签发者私钥；self 模式用自有私钥。
//
// self 模式配合模板证书 / CSR 时，**必须**提供 existingPrivateKey：
// 否则新证书的公钥与签名私钥不配对，产物无法通过验签。
func resolveCertMaterial(req *CertGenRequest,
	issuerKey *pkisvc.PrivateKey) (*certMaterial, error) {
	out := &certMaterial{cleanup: func() {}}
	closers := make([]func(), 0, 3)
	out.cleanup = func() {
		for _, close := range closers {
			close()
		}
	}
	fail := func(err error) (*certMaterial, error) {
		out.cleanup()
		return nil, err
	}

	needOwnKey := req.SignMode == pkisvc.SignModeSelf
	foreignKey := hasForeignPublicKey(req)
	if needOwnKey && foreignKey && (req.ExistingPrivateKey == nil ||
		req.ExistingPrivateKey.Data == "") {
		return fail(tools.MissingField("existingPrivateKey"))
	}

	// 自签模式必须有自有私钥：模板 / CSR 只提供公钥，没有私钥无法签名。
	if needOwnKey || !foreignKey {
		existing := req.ExistingPrivateKey
		if foreignKey && !needOwnKey {
			// ca 模式用模板 / CSR 的公钥，不需要自有私钥。
			existing = nil
		}
		if existing != nil || !foreignKey {
			key, pemText, err := loadOrCreateKey(req.Key, existing)
			if err != nil {
				return fail(err)
			}
			closers = append(closers, func() { _ = key.Close() })
			out.signer = key
			out.privatePem = pemText
			if !foreignKey {
				pub, err := pkisvc.PublicKeyFromPrivateKey(key)
				if err != nil {
					return fail(err)
				}
				closers = append(closers, func() { _ = pub.Close() })
				out.public = pub
			}
		}
	}

	if needsIssuerSignature(req) {
		out.signer = issuerKey
	}

	if out.public == nil {
		pub, info, err := publicFromTemplateOrCSR(req)
		if err != nil {
			return fail(err)
		}
		closers = append(closers, func() { _ = pub.Close() })
		out.public = pub
		out.info = info
	} else {
		// 公钥参数取自**新证书的公钥**，而不是签名私钥：
		// ca 模式下两者分别属于叶证书与签发者。
		info, err := pkisvc.DescribePublicKey(out.public)
		if err != nil {
			return fail(err)
		}
		out.info = publicKeyInfoOf(info)
	}

	return out, nil
}

// needsIssuerSignature 报告是否由签发者私钥签名。
func needsIssuerSignature(req *CertGenRequest) bool {
	return req.SignMode == pkisvc.SignModeCA
}

// hasForeignPublicKey 报告请求里是否带了「非自有私钥」的公钥来源。
func hasForeignPublicKey(req *CertGenRequest) bool {
	if req.TemplateCert != nil && req.TemplateCert.Data != "" {
		return true
	}
	return req.Csr != nil && req.Csr.Data != ""
}

// publicFromTemplateOrCSR 从模板证书或 CSR 取公钥。
func publicFromTemplateOrCSR(req *CertGenRequest) (*pkisvc.PublicKey, pkisvc.PublicKeyInfo, error) {
	if req.TemplateCert != nil && req.TemplateCert.Data != "" {
		source, isPEM, err := req.TemplateCert.Resolve()
		if err != nil {
			return nil, pkisvc.PublicKeyInfo{}, err
		}
		template, err := pkisvc.LoadCertificate(source, isPEM)
		if err != nil {
			return nil, pkisvc.PublicKeyInfo{}, err
		}
		defer func() { _ = template.Close() }()

		pub, err := pkisvc.PublicKeyFromCertificate(template)
		if err != nil {
			return nil, pkisvc.PublicKeyInfo{}, err
		}
		info, err := pkisvc.Describe(template, time.Time{}, nil)
		if err != nil {
			_ = pub.Close()
			return nil, pkisvc.PublicKeyInfo{}, err
		}
		return pub, info.PublicKey, nil
	}

	source, isPEM, err := req.Csr.Resolve()
	if err != nil {
		return nil, pkisvc.PublicKeyInfo{}, err
	}
	pub, err := pkisvc.PublicKeyFromCSR(source, isPEM)
	if err != nil {
		return nil, pkisvc.PublicKeyInfo{}, err
	}
	info, err := pkisvc.ParseCSR(source, isPEM)
	if err != nil {
		_ = pub.Close()
		return nil, pkisvc.PublicKeyInfo{}, err
	}
	return pub, info.PublicKey, nil
}

// loadOrCreateKey 按「已有私钥优先，否则新建」返回私钥。
//
// 返回的私钥由调用方负责 Close；第二个返回值是新生成时的 PKCS#8 PEM。
func loadOrCreateKey(gen KeyGenInput, existing *KeyInput) (*pkisvc.PrivateKey, string, error) {
	if existing != nil && existing.Data != "" {
		source, isPEM, err := existing.Resolve()
		if err != nil {
			return nil, "", err
		}
		if !isPEM {
			return nil, "", tools.UnsupportedFormat("der", []string{"pem"})
		}
		key, err := pkisvc.LoadPrivateKey(source, existing.Passphrase)
		if err != nil {
			return nil, "", err
		}
		return key, "", nil
	}

	pair, err := pkisvc.Generate(gen.algorithm(), gen.Bits, gen.Curve)
	if err != nil {
		return nil, "", err
	}
	// 公钥句柄在这里用不到（需要时另行派生独立句柄）。
	_ = pair.Public.Close()

	pemBytes, err := pair.Private.MarshalPKCS8PEM()
	if err != nil {
		_ = pair.Private.Close()
		return nil, "", err
	}
	return pair.Private, string(pemBytes), nil
}

// chainPEM 组装响应里的证书链。
//
// ca 模式返回「叶 → 签发者」；self 模式只返回叶证书本身。
func chainPEM(mode string, leafPEM []byte, issuer *pkisvc.Certificate) []string {
	chain := []string{string(leafPEM)}
	if mode != pkisvc.SignModeCA || issuer == nil {
		return chain
	}
	if issuerPEM, err := issuer.MarshalPEM(); err == nil {
		chain = append(chain, string(issuerPEM))
	}
	return chain
}

// resolveNotAfter 计算失效时间。
func resolveNotAfter(req *CertGenRequest, notBefore time.Time) (time.Time, error) {
	if req.NotAfter != "" {
		at, err := time.Parse(time.RFC3339, req.NotAfter)
		if err != nil {
			return time.Time{}, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
				"field":    "notAfter",
				"expected": "RFC3339",
			})
		}
		if !at.After(notBefore) {
			return time.Time{}, toolerr.Newf(toolerr.CodeNotAfterBeforeNow, map[string]any{
				"field": "notAfter",
			})
		}
		return at, nil
	}

	days := req.NotAfterDays
	if days == 0 {
		days = defaultCertDays
	}
	return notBefore.AddDate(0, 0, days), nil
}

// resolveSerial 解析十六进制序列号；为空时随机生成。
func resolveSerial(value string) (int64, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return pkisvc.RandomSerial()
	}
	text = strings.TrimPrefix(strings.TrimPrefix(text, "0x"), "0X")
	serial, err := strconv.ParseInt(text, 16, 64)
	if err != nil || serial <= 0 {
		return 0, toolerr.Newf(toolerr.CodeSerialInvalid, map[string]any{
			"field":    "serial",
			"expected": "positive hex integer up to 8 bytes",
		})
	}
	return serial, nil
}

// loadIssuer 加载签发者证书与私钥。
func loadIssuer(req *CertGenRequest) (*pkisvc.Certificate, *pkisvc.PrivateKey, error) {
	if req.Issuer == nil {
		return nil, nil, nil
	}

	source, isPEM, err := req.Issuer.Cert.Resolve()
	if err != nil {
		return nil, nil, err
	}
	cert, err := pkisvc.LoadCertificate(source, isPEM)
	if err != nil {
		return nil, nil, err
	}

	keySource, keyIsPEM, err := req.Issuer.Key.Resolve()
	if err != nil {
		_ = cert.Close()
		return nil, nil, err
	}
	if !keyIsPEM {
		_ = cert.Close()
		return nil, nil, tools.UnsupportedFormat("der", []string{"pem"})
	}
	key, err := pkisvc.LoadPrivateKey(keySource, req.Issuer.Key.Passphrase)
	if err != nil {
		_ = cert.Close()
		return nil, nil, err
	}
	return cert, key, nil
}

// validityWindow 是一个有效期窗口。
type validityWindow struct {
	notBefore time.Time
	notAfter  time.Time
}

// isZero 报告窗口是否未设置。
func (w validityWindow) isZero() bool {
	return w.notBefore.IsZero() && w.notAfter.IsZero()
}

// resolveSubject 解析主体 DN。
//
// 优先级：subjectOverride 逐项覆盖 > CSR 的主体 > subject 字段。
func resolveSubject(req *CertGenRequest) ([]pkisvc.NameEntryInfo, validityWindow, error) {
	var (
		entries  []pkisvc.NameEntryInfo
		validity validityWindow
	)

	if req.Csr != nil && req.Csr.Data != "" {
		source, isPEM, err := req.Csr.Resolve()
		if err != nil {
			return nil, validityWindow{}, err
		}
		info, err := pkisvc.ParseCSR(source, isPEM)
		if err != nil {
			return nil, validityWindow{}, err
		}
		entries = info.Subject
		// CSR 没有有效期概念，窗口留空。
	} else if req.TemplateCert != nil && req.TemplateCert.Data != "" {
		source, isPEM, err := req.TemplateCert.Resolve()
		if err != nil {
			return nil, validityWindow{}, err
		}
		template, err := pkisvc.LoadCertificate(source, isPEM)
		if err != nil {
			return nil, validityWindow{}, err
		}
		info, describeErr := pkisvc.Describe(template, time.Time{}, nil)
		_ = template.Close()
		if describeErr != nil {
			return nil, validityWindow{}, describeErr
		}
		entries = info.Subject
		validity = validityWindow{notBefore: info.NotBefore, notAfter: info.NotAfter}
	}

	if len(entries) == 0 {
		entries = req.Subject.entries()
	}
	if req.SubjectOverride != nil {
		entries = mergeSubject(entries, req.SubjectOverride.entries())
	}
	return entries, validity, nil
}

// mergeSubject 用覆盖项替换同名 RDN，未提到的条目保持原样。
func mergeSubject(base, override []pkisvc.NameEntryInfo) []pkisvc.NameEntryInfo {
	overrides := make(map[string]string, len(override))
	for _, entry := range override {
		overrides[entry.Field] = entry.Value
	}

	out := make([]pkisvc.NameEntryInfo, 0, len(base)+len(override))
	for _, entry := range base {
		if value, ok := overrides[entry.Field]; ok {
			out = append(out, pkisvc.NameEntryInfo{Field: entry.Field, Value: value})
			delete(overrides, entry.Field)
			continue
		}
		out = append(out, entry)
	}
	// 覆盖项里新增的字段按输入顺序追加。
	for _, entry := range override {
		if value, ok := overrides[entry.Field]; ok {
			out = append(out, pkisvc.NameEntryInfo{Field: entry.Field, Value: value})
			delete(overrides, entry.Field)
		}
	}
	return out
}
