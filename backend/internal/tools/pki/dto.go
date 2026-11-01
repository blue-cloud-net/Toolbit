// Package pki 实现证书类工具。
//
// 本包**不**实现算法：只做「解码输入 → 校验字段 → 调用 pkisvc → 编码 DTO」。
//
// 所有 X.509 处理都经 pkisvc 转调 Tongsuo；本包**不得**直接 import tongsuo-go，
// 也**不得**用标准库 crypto/x509 解析证书。
package pki

import (
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// PemBundle 是证书族材料输入（复用 tools 包的共享类型）。
type PemBundle = tools.PemBundle

// KeyInput 是密钥输入（复用 tools 包的共享类型）。
type KeyInput = tools.KeyInput

// 输入格式（复用 tools 包的共享定义）。
const (
	// FormatAuto 自动识别。
	FormatAuto = tools.FormatAuto
	// FormatPEM 是 PEM 文本。
	FormatPEM = tools.FormatPEM
	// FormatDER 是 DER 二进制的 base64 表示。
	FormatDER = tools.FormatDER
)

// PEM 类型过滤值。
const (
	// TypeCert 只保留证书。
	TypeCert = pkisvc.PEMTypeCert
	// TypeKey 只保留密钥。
	TypeKey = pkisvc.PEMTypeKey
	// TypeCSR 只保留证书签名请求。
	TypeCSR = pkisvc.PEMTypeCSR
	// TypeCRL 只保留吊销列表。
	TypeCRL = pkisvc.PEMTypeCRL
	// TypeAny 保留全部。
	TypeAny = pkisvc.PEMTypeAny
)

// 转换目标格式。
const (
	// ToPEM 输出 PEM 文本。
	ToPEM = pkisvc.ConvertToPEM
	// ToDER 输出 DER 二进制（base64 承载）。
	ToDER = pkisvc.ConvertToDER
)

// NameEntryOutput 是一个 RDN 条目的输出形态。
type NameEntryOutput struct {
	// Field 是字段短名（如 CN、O）。
	Field string `json:"field"`
	// Value 是字段值。
	Value string `json:"value"`
}

// PublicKeyOutput 是公钥元信息的输出形态。
type PublicKeyOutput struct {
	// Algorithm 是算法标识（RSA / SM2 / EC …）。
	Algorithm string `json:"algorithm"`
	// Bits 是 RSA 密钥位数；非 RSA 为 0。
	Bits int `json:"bits"`
	// Curve 是曲线名；非椭圆曲线为空。
	Curve string `json:"curve"`
}

// ExtensionOutput 是一个扩展的输出形态。
type ExtensionOutput struct {
	// Nid 是扩展的 OpenSSL NID。
	Nid int `json:"nid"`
	// OID 是扩展的点分 OID。
	OID string `json:"oid"`
	// Name 是扩展的短名；未知扩展为空。
	Name string `json:"name"`
	// Critical 报告扩展是否标记为 critical。
	Critical bool `json:"critical"`
	// Value 是扩展值的十六进制。
	Value string `json:"value"`
	// Text 是扩展的可读文本；未知扩展为空。
	Text string `json:"text"`
}

// ValidityOutput 是有效性判定的输出形态。
type ValidityOutput struct {
	// Status 取值：valid / expired / not_yet_valid。
	Status string `json:"status"`
	// DaysRemaining 是距离过期的剩余天数；已过期为负数。
	DaysRemaining int `json:"daysRemaining"`
}

// FingerprintOutput 是一组指纹中的一项。
type FingerprintOutput struct {
	// Algorithm 是算法名（小写）。
	Algorithm string `json:"algorithm"`
	// Value 是指纹值（小写十六进制）。
	Value string `json:"value"`
}

// CertInfoOutput 是证书元信息的输出形态。
type CertInfoOutput struct {
	// Version 是版本号（2 表示 v3）。
	Version int `json:"version"`
	// Subject 是主体的 RDN 条目。
	Subject []NameEntryOutput `json:"subject"`
	// SubjectText 是主体的单行文本。
	SubjectText string `json:"subjectText"`
	// Issuer 是签发者的 RDN 条目。
	Issuer []NameEntryOutput `json:"issuer"`
	// IssuerText 是签发者的单行文本。
	IssuerText string `json:"issuerText"`
	// Serial 是序列号（大写十六进制）。
	Serial string `json:"serial"`
	// NotBefore 是生效时间。
	NotBefore time.Time `json:"notBefore"`
	// NotAfter 是过期时间。
	NotAfter time.Time `json:"notAfter"`
	// PublicKey 是公钥元信息。
	PublicKey PublicKeyOutput `json:"publicKey"`
	// SignatureAlgorithm 是签名算法短名。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SignatureAlgorithmOid 是签名算法的点分 OID。
	SignatureAlgorithmOid string `json:"signatureAlgorithmOid"`
	// IsCa 报告是否为 CA 证书。
	IsCa bool `json:"isCa"`
	// PathLen 是 CA 路径长度约束；非 CA 或无该扩展为 -1。
	PathLen int64 `json:"pathLen"`
	// San 是主体备用名条目。
	San []string `json:"san"`
	// KeyUsage 是密钥用途位名。
	KeyUsage []string `json:"keyUsage"`
	// ExtendedKeyUsage 是扩展密钥用途。
	ExtendedKeyUsage []string `json:"extendedKeyUsage"`
	// SubjectKeyId 是主题密钥标识符（大写十六进制）。
	SubjectKeyId string `json:"subjectKeyId"`
	// AuthorityKeyId 是签发者密钥标识符（大写十六进制）。
	AuthorityKeyId string `json:"authorityKeyId"`
	// SubjectHash 是主体名的 X509_NAME_hash。
	SubjectHash string `json:"subjectHash"`
	// IssuerHash 是签发者名的 X509_NAME_hash。
	IssuerHash string `json:"issuerHash"`
	// Extensions 是全部扩展。
	Extensions []ExtensionOutput `json:"extensions"`
	// SelfSigned 报告是否自签名。
	SelfSigned bool `json:"selfSigned"`
	// SignatureHex 是签名字节的十六进制。
	SignatureHex string `json:"signatureHex"`
	// Policies 是证书策略列表；无该扩展时为空数组。
	Policies []PolicyOutput `json:"policies"`
}

// PublicKeyDetailOutput 是证书公钥展开视图的输出形态。
type PublicKeyDetailOutput struct {
	// Algorithm 是算法标识（RSA / SM2 / EC …）。
	Algorithm string `json:"algorithm"`
	// Bits 是 RSA 密钥位数；非 RSA 为 0。
	Bits int `json:"bits"`
	// Curve 是曲线名；非椭圆曲线为空。
	Curve string `json:"curve"`
	// Pem 是公钥的 SPKI PEM 文本。
	Pem string `json:"pem"`
	// SpkiHex 是 SPKI DER 的十六进制。
	SpkiHex string `json:"spkiHex"`
	// Fingerprint 是 SPKI DER 的 SHA-256 指纹。
	Fingerprint string `json:"fingerprint"`
	// Modulus 是 RSA 模数（大写十六进制）；非 RSA 为空。
	Modulus string `json:"modulus"`
	// Exponent 是 RSA 公钥指数（大写十六进制）；非 RSA 为空。
	Exponent string `json:"exponent"`
	// X 是椭圆曲线公钥点 X 分量（大写十六进制）。
	X string `json:"x"`
	// Y 是椭圆曲线公钥点 Y 分量（大写十六进制）。
	Y string `json:"y"`
}

// PolicyOutput 是证书策略的输出形态。
type PolicyOutput struct {
	// Oid 是策略标识符（点分形式）。
	Oid string `json:"oid"`
	// CpsUris 是策略声明 URL 列表。
	CpsUris []string `json:"cpsUris"`
	// Organization 是通知机构名。
	Organization string `json:"organization"`
	// NoticeNumbers 是通知编号列表。
	NoticeNumbers []int `json:"noticeNumbers"`
	// ExplicitText 是通知正文。
	ExplicitText string `json:"explicitText"`
}

// ChainLinkOutput 是链上连接诊断的输出形态。
type ChainLinkOutput struct {
	// Method 是连接方式（leaf / keyid / name / self）。
	Method string `json:"method"`
	// Ok 报告连接是否成立。
	Ok bool `json:"ok"`
	// Reason 是断链原因；正常时为空。
	Reason string `json:"reason"`
}

// certInfoOutput 把服务层的证书视图转成输出形态。
func certInfoOutput(info *pkisvc.CertInfo) CertInfoOutput {
	if info == nil {
		return CertInfoOutput{}
	}
	return CertInfoOutput{
		Version:               info.Version,
		Subject:               nameEntriesOutput(info.Subject),
		SubjectText:           info.SubjectText,
		Issuer:                nameEntriesOutput(info.Issuer),
		IssuerText:            info.IssuerText,
		Serial:                info.Serial,
		NotBefore:             info.NotBefore,
		NotAfter:              info.NotAfter,
		PublicKey:             publicKeyOutput(info.PublicKey),
		SignatureAlgorithm:    info.SignatureAlgorithm,
		SignatureAlgorithmOid: info.SignatureAlgorithmOID,
		IsCa:                  info.IsCA,
		PathLen:               info.PathLen,
		San:                   info.SAN,
		KeyUsage:              info.KeyUsage,
		ExtendedKeyUsage:      info.ExtendedKeyUsage,
		SubjectKeyId:          info.SubjectKeyID,
		AuthorityKeyId:        info.AuthorityKeyID,
		SubjectHash:           info.SubjectHash,
		IssuerHash:            info.IssuerHash,
		Extensions:            extensionsOutput(info.Extensions),
		SelfSigned:            info.SelfSigned,
		SignatureHex:          info.SignatureHex,
		Policies:              policiesOutput(info.Policies),
	}
}

// publicKeyDetailOutput 把公钥展开视图转成输出形态。
func publicKeyDetailOutput(detail *pkisvc.PublicKeyDetail) *PublicKeyDetailOutput {
	if detail == nil {
		return nil
	}
	return &PublicKeyDetailOutput{
		Algorithm:   detail.Algorithm,
		Bits:        detail.Bits,
		Curve:       detail.Curve,
		Pem:         detail.PEM,
		SpkiHex:     detail.SPKIHex,
		Fingerprint: detail.Fingerprint,
		Modulus:     detail.Modulus,
		Exponent:    detail.Exponent,
		X:           detail.X,
		Y:           detail.Y,
	}
}

// policiesOutput 把策略列表转成输出形态。
func policiesOutput(policies []pkisvc.PolicyInfo) []PolicyOutput {
	out := make([]PolicyOutput, 0, len(policies))
	for _, item := range policies {
		out = append(out, PolicyOutput{
			Oid:           item.OID,
			CpsUris:       item.CPSURIs,
			Organization:  item.Organization,
			NoticeNumbers: item.NoticeNumbers,
			ExplicitText:  item.ExplicitText,
		})
	}
	return out
}

// chainLinkOutput 把连接诊断转成输出形态。
func chainLinkOutput(link pkisvc.ChainLink) ChainLinkOutput {
	return ChainLinkOutput{
		Method: link.Method,
		Ok:     link.OK,
		Reason: link.Reason,
	}
}

// nameEntriesOutput 把 RDN 条目转成输出形态。
func nameEntriesOutput(entries []pkisvc.NameEntryInfo) []NameEntryOutput {
	out := make([]NameEntryOutput, 0, len(entries))
	for _, entry := range entries {
		out = append(out, NameEntryOutput{Field: entry.Field, Value: entry.Value})
	}
	return out
}

// publicKeyOutput 把公钥元信息转成输出形态。
func publicKeyOutput(info pkisvc.PublicKeyInfo) PublicKeyOutput {
	return PublicKeyOutput{
		Algorithm: info.Algorithm,
		Bits:      info.Bits,
		Curve:     info.Curve,
	}
}

// extensionsOutput 把扩展列表转成输出形态。
func extensionsOutput(extensions []pkisvc.ExtInfo) []ExtensionOutput {
	out := make([]ExtensionOutput, 0, len(extensions))
	for _, ext := range extensions {
		out = append(out, ExtensionOutput{
			Nid:      ext.Nid,
			OID:      ext.OID,
			Name:     ext.Name,
			Critical: ext.Critical,
			Value:    ext.Value,
			Text:     ext.Text,
		})
	}
	return out
}
