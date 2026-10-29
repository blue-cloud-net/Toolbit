package pkisvc

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 证书指纹算法。
const (
	// FingerprintMD5 是 MD5 指纹（弱，仅用于与旧系统对齐）。
	FingerprintMD5 = "md5"
	// FingerprintSHA1 是 SHA-1 指纹（弱，仅用于与旧系统对齐）。
	FingerprintSHA1 = "sha1"
	// FingerprintSHA256 是 SHA-256 指纹。
	FingerprintSHA256 = "sha256"
	// FingerprintSHA384 是 SHA-384 指纹。
	FingerprintSHA384 = "sha384"
	// FingerprintSHA512 是 SHA-512 指纹。
	FingerprintSHA512 = "sha512"
	// FingerprintSM3 是 SM3 指纹。
	FingerprintSM3 = "sm3"
)

// 证书有效性状态。
const (
	// ValidityValid 表示在有效期内。
	ValidityValid = "valid"
	// ValidityExpired 表示已过期。
	ValidityExpired = "expired"
	// ValidityNotYetValid 表示尚未生效。
	ValidityNotYetValid = "not_yet_valid"
)

// FingerprintAlgorithms 返回受支持的证书指纹算法（顺序稳定）。
//
// 全部算法都**允许**计算；md5 / sha1 属弱哈希，
// 由上层附一条 WEAK_HASH 警告（见 docs/standards/04-errors.md §6）。
func FingerprintAlgorithms() []string {
	return []string{
		FingerprintMD5,
		FingerprintSHA1,
		FingerprintSHA256,
		FingerprintSHA384,
		FingerprintSHA512,
		FingerprintSM3,
	}
}

// WeakFingerprintAlgorithms 返回需要附带弱哈希警告的算法（顺序稳定）。
func WeakFingerprintAlgorithms() []string {
	return []string{FingerprintMD5, FingerprintSHA1}
}

// Certificate 是 X.509 证书句柄。
//
// 它持有原生句柄，**必须**显式 Close（见 docs/standards/03-crypto.md §3）。
type Certificate struct {
	// cert 是底层证书；Close 之后置为 nil。
	cert *x509.Certificate
}

// Close 释放证书句柄。
//
// **幂等**：重复调用只返回首次的结果。
func (c *Certificate) Close() error {
	if c == nil || c.cert == nil {
		return nil
	}
	err := c.cert.Close()
	c.cert = nil
	return err
}

// MarshalPEM 导出证书 PEM。
func (c *Certificate) MarshalPEM() ([]byte, error) {
	if c == nil || c.cert == nil {
		return nil, toolerrInvalidCert("certificate already closed")
	}
	out, err := c.cert.MarshalPEM()
	if err != nil {
		return nil, toolerrInvalidCert("cannot export certificate")
	}
	return out, nil
}

// MarshalDER 导出证书 DER。
func (c *Certificate) MarshalDER() ([]byte, error) {
	if c == nil || c.cert == nil {
		return nil, toolerrInvalidCert("certificate already closed")
	}
	out, err := c.cert.MarshalDER()
	if err != nil {
		return nil, toolerrInvalidCert("cannot export certificate")
	}
	return out, nil
}

// NameEntryInfo 是一个 RDN 条目。
type NameEntryInfo struct {
	// Field 是字段短名（如 CN、O）。
	Field string `json:"field"`
	// Value 是字段值。
	Value string `json:"value"`
}

// PublicKeyInfo 是公钥的元信息。
type PublicKeyInfo struct {
	// Algorithm 是算法标识（RSA / SM2 / EC …）。
	Algorithm string `json:"algorithm"`
	// Bits 是 RSA 密钥位数；非 RSA 为 0。
	Bits int `json:"bits"`
	// Curve 是曲线名（P-256 / P-384 / P-521 / SM2）；非椭圆曲线为空。
	Curve string `json:"curve"`
}

// ExtInfo 是一个 X.509 扩展的元信息。
type ExtInfo struct {
	// Nid 是扩展的 OpenSSL NID。
	Nid int `json:"nid"`
	// OID 是扩展的点分 OID。
	OID string `json:"oid"`
	// Name 是扩展的短名（如 subjectAltName）；未知扩展为空。
	Name string `json:"name"`
	// Critical 报告扩展是否标记为 critical。
	Critical bool `json:"critical"`
	// Value 是扩展值的十六进制。
	Value string `json:"value"`
}

// ValidityInfo 是有效性的判定结果。
type ValidityInfo struct {
	// Status 取值见 ValidityValid / ValidityExpired / ValidityNotYetValid。
	Status string `json:"status"`
	// DaysRemaining 是距离过期的剩余天数；已过期为负数。
	DaysRemaining int `json:"daysRemaining"`
}

// FingerprintInfo 是一组指纹中的一项。
type FingerprintInfo struct {
	// Algorithm 是算法名（小写）。
	Algorithm string `json:"algorithm"`
	// Value 是指纹值（小写十六进制）。
	Value string `json:"value"`
}

// CertInfo 是证书的公开元信息快照。
//
// 它是 pkisvc 对 X.509 的**唯一**元信息出口：公开签名不出现
// tongsuo-go 的类型，因此上层拿不到底层库的句柄。
type CertInfo struct {
	// Version 是版本号（2 表示 v3）。
	Version int `json:"version"`
	// Subject 是主体的 RDN 条目（按证书中的顺序）。
	Subject []NameEntryInfo `json:"subject"`
	// SubjectText 是主体的单行文本。
	SubjectText string `json:"subjectText"`
	// Issuer 是签发者的 RDN 条目。
	Issuer []NameEntryInfo `json:"issuer"`
	// IssuerText 是签发者的单行文本。
	IssuerText string `json:"issuerText"`
	// Serial 是序列号（大写十六进制，不带 0x 前缀）。
	Serial string `json:"serial"`
	// NotBefore 是生效时间。
	NotBefore time.Time `json:"notBefore"`
	// NotAfter 是过期时间。
	NotAfter time.Time `json:"notAfter"`
	// PublicKey 是公钥元信息。
	PublicKey PublicKeyInfo `json:"publicKey"`
	// SignatureAlgorithm 是签名算法短名。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SignatureAlgorithmOID 是签名算法的点分 OID。
	SignatureAlgorithmOID string `json:"signatureAlgorithmOid"`
	// IsCA 报告是否为 CA 证书。
	IsCA bool `json:"isCa"`
	// PathLen 是 CA 路径长度约束；非 CA 或无该扩展为 -1。
	PathLen int64 `json:"pathLen"`
	// SAN 是主体备用名条目。
	SAN []string `json:"san"`
	// KeyUsage 是密钥用途位名。
	KeyUsage []string `json:"keyUsage"`
	// ExtendedKeyUsage 是扩展密钥用途。
	ExtendedKeyUsage []string `json:"extendedKeyUsage"`
	// SubjectKeyID 是主题密钥标识符（大写十六进制）；无该扩展为空。
	SubjectKeyID string `json:"subjectKeyId"`
	// AuthorityKeyID 是签发者密钥标识符（大写十六进制）；无该扩展为空。
	AuthorityKeyID string `json:"authorityKeyId"`
	// SubjectHash 是主体名的 X509_NAME_hash（8 位大写十六进制）。
	SubjectHash string `json:"subjectHash"`
	// IssuerHash 是签发者名的 X509_NAME_hash（8 位大写十六进制）。
	IssuerHash string `json:"issuerHash"`
	// Extensions 是全部扩展。
	Extensions []ExtInfo `json:"extensions"`
	// SelfSigned 报告是否自签名。
	SelfSigned bool `json:"selfSigned"`
}

// LoadCertificate 加载一张证书。
//
// source 是 PEM 文本或 DER 二进制；isPEM 报告它是哪一种。
// 输入是 PEM bundle 时只取**第一块**。
func LoadCertificate(source []byte, isPEM bool) (*Certificate, error) {
	if len(source) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "input.data"})
	}

	if isPEM {
		cert, err := x509.LoadCertificatePEM(source)
		if err != nil {
			return nil, toolerrInvalidCert("cannot parse certificate")
		}
		return &Certificate{cert: cert}, nil
	}

	cert, err := x509.LoadCertificateDER(source)
	if err != nil {
		return nil, toolerrInvalidCert("cannot parse certificate")
	}
	return &Certificate{cert: cert}, nil
}

// ParseCertificates 加载 PEM bundle 中的全部证书。
//
// 块数超过 MaxPEMBlocks 时返回 TOO_MANY_ITEMS；
// 其中任一块不是证书时返回 INVALID_CERT。
func ParseCertificates(source []byte, isPEM bool) ([]*Certificate, error) {
	if len(source) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "input.data"})
	}

	if !isPEM {
		cert, err := LoadCertificate(source, false)
		if err != nil {
			return nil, err
		}
		return []*Certificate{cert}, nil
	}

	blocks, err := SplitPEM(source)
	if err != nil {
		return nil, err
	}

	certs := make([]*Certificate, 0, len(blocks))
	for _, block := range blocks {
		if block.Label != normalizePEMLabel(PEMBlockCertificate) {
			// 非证书块：跳过（bundle 里混有密钥是常见的）。
			continue
		}
		cert, err := x509.LoadCertificatePEM(pemEncode(block.Type, block.Bytes))
		if err != nil {
			closeCertificates(certs)
			return nil, toolerrInvalidCert("cannot parse certificate")
		}
		certs = append(certs, &Certificate{cert: cert})
	}

	if len(certs) == 0 {
		return nil, toolerrInvalidCert("no certificate in bundle")
	}
	return certs, nil
}

// Describe 读取证书的元信息。
//
// at 为零值时按当前时间判定有效性；fpAlgs 为空时只算 SHA-256。
func Describe(c *Certificate, at time.Time, fpAlgs []string) (*CertInfo, error) {
	if c == nil || c.cert == nil {
		return nil, toolerrInvalidCert("certificate already closed")
	}

	info := &CertInfo{
		Version:               c.cert.Version(),
		Subject:               nameEntries(c.cert.SubjectEntries()),
		SubjectText:           c.cert.SubjectText(),
		Issuer:                nameEntries(c.cert.IssuerEntries()),
		IssuerText:            c.cert.IssuerText(),
		Serial:                c.cert.SerialHex(),
		NotBefore:             c.cert.NotBefore(),
		NotAfter:              c.cert.NotAfter(),
		SignatureAlgorithm:    c.cert.SignatureAlgorithm(),
		SignatureAlgorithmOID: c.cert.SignatureAlgorithmOID(),
		IsCA:                  c.cert.IsCA(),
		PathLen:               c.cert.PathLen(),
		SAN:                   c.cert.SAN(),
		KeyUsage:              c.cert.KeyUsage(),
		ExtendedKeyUsage:      c.cert.ExtendedKeyUsage(),
		SubjectKeyID:          hexUpper(c.cert.SubjectKeyID()),
		AuthorityKeyID:        hexUpper(c.cert.AuthorityKeyID()),
		SubjectHash:           fmt.Sprintf("%08X", c.cert.SubjectHash()),
		IssuerHash:            fmt.Sprintf("%08X", c.cert.IssuerHash()),
		Extensions:            extensionInfos(c.cert.Extensions()),
	}

	selfSigned, err := c.cert.SelfSigned()
	if err != nil {
		// 自签名判定失败**不**影响其它字段：按「不是自签名」处理。
		selfSigned = false
	}
	info.SelfSigned = selfSigned

	pub, err := describeCertificatePublicKey(c)
	if err != nil {
		return nil, err
	}
	info.PublicKey = pub

	return info, nil
}

// describeCertificatePublicKey 读取证书公钥的元信息。
func describeCertificatePublicKey(c *Certificate) (PublicKeyInfo, error) {
	pub, err := c.cert.PublicKey()
	if err != nil {
		return PublicKeyInfo{}, toolerrInvalidCert("cannot read public key")
	}
	defer func() { _ = asym.Close(pub) }()

	return describeUpstreamPublicKey(pub), nil
}

// describeUpstreamPublicKey 读取底层公钥句柄的元信息。
//
// 句柄的释放由调用方负责：公钥可能来自 `Certificate.PublicKey()`
// （独立句柄，需 `asym.Close`）或 `PrivateKey.Public()`（与私钥共享的**别名**，
// 不得单独关闭）。
func describeUpstreamPublicKey(pub asym.PublicKey) PublicKeyInfo {
	out := PublicKeyInfo{
		Algorithm: algorithmFromUpstream(pub.Algorithm().String()),
	}
	params, err := asym.Params(pub)
	if err != nil {
		// 读不到参数时只保留算法名，不视为致命错误。
		return out
	}
	if params != nil {
		out.Bits = rsaBits(params)
		out.Curve = curveFromUpstream(params.Curve)
	}
	return out
}

// Fingerprints 计算证书的指纹。
//
// algs 为空时只算 SHA-256；算法名大小写不敏感，
// 返回值按请求顺序排列。
func Fingerprints(c *Certificate, algs []string) ([]FingerprintInfo, error) {
	if c == nil || c.cert == nil {
		return nil, toolerrInvalidCert("certificate already closed")
	}
	if len(algs) == 0 {
		algs = []string{FingerprintSHA256}
	}

	out := make([]FingerprintInfo, 0, len(algs))
	for _, alg := range algs {
		name := normalizeFingerprintAlgorithm(alg)
		if !containsString(FingerprintAlgorithms(), name) {
			return nil, toolerr.Newf(toolerr.CodeUnsupportedAlgorithm, map[string]any{
				"algorithm": name,
				"supported": FingerprintAlgorithms(),
			})
		}
		value, err := c.cert.Fingerprint(name)
		if err != nil {
			return nil, toolerrInvalidCert("cannot compute fingerprint")
		}
		out = append(out, FingerprintInfo{Algorithm: name, Value: value})
	}
	return out, nil
}

// normalizeFingerprintAlgorithm 归一化指纹算法名（小写）。
func normalizeFingerprintAlgorithm(alg string) string {
	return strings.ToLower(strings.TrimSpace(alg))
}

// CheckValidity 判定证书在 at 时刻的有效性。
//
// at 为零值时按当前时间判定。
func CheckValidity(c *Certificate, at time.Time) ValidityInfo {
	if c == nil || c.cert == nil {
		return ValidityInfo{Status: ValidityNotYetValid}
	}
	now := at
	if now.IsZero() {
		now = time.Now()
	}

	notBefore := c.cert.NotBefore()
	notAfter := c.cert.NotAfter()

	switch {
	case now.Before(notBefore):
		return ValidityInfo{
			Status:        ValidityNotYetValid,
			DaysRemaining: int(math.Floor(notAfter.Sub(now).Hours() / 24)),
		}
	case now.After(notAfter):
		return ValidityInfo{
			Status:        ValidityExpired,
			DaysRemaining: int(math.Floor(notAfter.Sub(now).Hours() / 24)),
		}
	default:
		return ValidityInfo{
			Status:        ValidityValid,
			DaysRemaining: int(math.Floor(notAfter.Sub(now).Hours() / 24)),
		}
	}
}

// nameEntries 把 RDN 条目转成对外视图。
func nameEntries(entries []x509.NameEntry) []NameEntryInfo {
	out := make([]NameEntryInfo, 0, len(entries))
	for _, entry := range entries {
		out = append(out, NameEntryInfo{Field: entry.Field, Value: entry.Value})
	}
	return out
}

// extensionInfos 把扩展转成对外视图。
func extensionInfos(extensions []x509.Extension) []ExtInfo {
	out := make([]ExtInfo, 0, len(extensions))
	for _, ext := range extensions {
		out = append(out, ExtInfo{
			Nid:      ext.Nid,
			OID:      ext.OID,
			Name:     ext.Field,
			Critical: ext.Critical,
			Value:    hex.EncodeToString(ext.Data),
		})
	}
	return out
}

// hexUpper 把字节渲染为大写十六进制；空输入返回空串。
func hexUpper(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return hex.EncodeToString(data)
}

// closeCertificates 关闭一组证书句柄（忽略错误）。
func closeCertificates(certs []*Certificate) {
	for _, cert := range certs {
		_ = cert.Close()
	}
}
