package pkisvc

import (
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
)

// 证书详情视图的限额。
const (
	// CertTextMaxLen 是证书文本视图的长度上限（字节）。
	//
	// 超长时截断并追加提示行：文本视图是给人看的，
	// 不需要完整承载极端输入（AGENTS.md §4.6 的「不写敏感数据」同理，
	// 输出也不应无限膨胀）。
	CertTextMaxLen = 64 * 1024
	// maxExtTextItems 是单个扩展文本的最大条目数。
	//
	// SAN / AIA / CRLDP 都可能带大量条目，必须在**解析前**设限，
	// 避免极端输入把响应体撑爆（docs/standards/03-crypto.md §8.6）。
	maxExtTextItems = 64
	// maxPolicyItems 是证书策略的最大条数。
	maxPolicyItems = 32
)

// 常见扩展的点分 OID。
const (
	// oidSubjectKeyID 是 subjectKeyIdentifier 扩展。
	oidSubjectKeyID = "2.5.29.14"
	// oidKeyUsage 是 keyUsage 扩展。
	oidKeyUsage = "2.5.29.15"
	// oidSubjectAltName 是 subjectAltName 扩展。
	oidSubjectAltName = "2.5.29.17"
	// oidIssuerAltName 是 issuerAltName 扩展。
	oidIssuerAltName = "2.5.29.18"
	// oidBasicConstraints 是 basicConstraints 扩展。
	oidBasicConstraints = "2.5.29.19"
	// oidCertificatePolicies 是 certificatePolicies 扩展。
	oidCertificatePolicies = "2.5.29.32"
	// oidCRLDistributionPoints 是 cRLDistributionPoints 扩展。
	oidCRLDistributionPoints = "2.5.29.31"
	// oidAuthorityKeyID 是 authorityKeyIdentifier 扩展。
	oidAuthorityKeyID = "2.5.29.35"
	// oidExtKeyUsage 是 extendedKeyUsage 扩展。
	oidExtKeyUsage = "2.5.29.37"
	// oidAuthorityInfoAccess 是 authorityInfoAccess 扩展。
	oidAuthorityInfoAccess = "1.3.6.1.5.5.7.1.1"
)

// 证书链上的角色。
const (
	// ChainRoleLeaf 是端证书（链的起点）。
	ChainRoleLeaf = "leaf"
	// ChainRoleIntermediate 是中间 CA。
	ChainRoleIntermediate = "intermediate"
	// ChainRoleRoot 是自签名的根证书。
	ChainRoleRoot = "root"
	// ChainRoleUnlinked 是无法接入主链的证书。
	ChainRoleUnlinked = "unlinked"
)

// 链的连接方式与断链原因。
const (
	// ChainLinkLeaf 表示该证书是链的起点。
	ChainLinkLeaf = "leaf"
	// ChainLinkKeyID 表示该证书与上一级通过 AKID ↔ SKID 精确匹配。
	ChainLinkKeyID = "keyid"
	// ChainLinkName 表示该证书与上一级仅通过主体 / 签发者文本匹配。
	ChainLinkName = "name"
	// ChainLinkSelf 表示该证书自签名（链的终点）。
	ChainLinkSelf = "self"

	// ChainReasonOK 表示连接正常。
	ChainReasonOK = ""
	// ChainReasonKeyIDMismatch 表示签发者存在但密钥标识符对不上。
	ChainReasonKeyIDMismatch = "keyid_mismatch"
	// ChainReasonMissingIntermediate 表示签发者不在输入里（缺中间 CA）。
	ChainReasonMissingIntermediate = "missing_intermediate"
	// ChainReasonUnlinked 表示该证书无法接入主链（多余证书）。
	ChainReasonUnlinked = "unlinked"
)

// PublicKeyDetail 是证书公钥的展开视图。
//
// 它比 PublicKeyInfo 更细：额外给出 SPKI PEM / DER、指纹与算法分量。
// RSA 给出模数与指数，EC / SM2 给出公钥点坐标。
type PublicKeyDetail struct {
	// Algorithm 是算法标识（RSA / SM2 / EC …）。
	Algorithm string `json:"algorithm"`
	// Bits 是 RSA 密钥位数；非 RSA 为 0。
	Bits int `json:"bits"`
	// Curve 是曲线名（P-256 / P-384 / P-521 / SM2）；非椭圆曲线为空。
	Curve string `json:"curve"`
	// PEM 是公钥的 SubjectPublicKeyInfo PEM 文本。
	PEM string `json:"pem"`
	// SPKIHex 是 SubjectPublicKeyInfo DER 的十六进制。
	SPKIHex string `json:"spkiHex"`
	// Fingerprint 是 SPKI DER 的 SHA-256 指纹（小写十六进制）。
	//
	// 它只依赖公钥：同一把钥匙的私钥与公钥指纹相同。
	Fingerprint string `json:"fingerprint"`
	// Modulus 是 RSA 模数（大写十六进制）；非 RSA 为空。
	Modulus string `json:"modulus"`
	// Exponent 是 RSA 公钥指数（大写十六进制）；非 RSA 为空。
	Exponent string `json:"exponent"`
	// X 是椭圆曲线公钥点 X 分量（大写十六进制）；非椭圆曲线为空。
	X string `json:"x"`
	// Y 是椭圆曲线公钥点 Y 分量（大写十六进制）；非椭圆曲线为空。
	Y string `json:"y"`
}

// PolicyInfo 是 certificatePolicies 扩展里的一条策略。
type PolicyInfo struct {
	// OID 是策略标识符（点分形式）。
	OID string `json:"oid"`
	// CPSURIs 是 id-qt-cps 限定符的策略声明 URL 列表。
	CPSURIs []string `json:"cpsUris"`
	// Organization 是 id-qt-unotice 限定符的通知机构名。
	Organization string `json:"organization"`
	// NoticeNumbers 是 id-qt-unotice 限定符的通知编号列表。
	NoticeNumbers []int `json:"noticeNumbers"`
	// ExplicitText 是 id-qt-unotice 限定符的通知正文。
	ExplicitText string `json:"explicitText"`
}

// ChainLink 描述证书与其上一级之间的连接方式。
type ChainLink struct {
	// Method 是连接方式（见 ChainLink* 常量）。
	Method string `json:"method"`
	// OK 报告连接是否成立。
	OK bool `json:"ok"`
	// Reason 是断链原因（见 ChainReason* 常量）；正常时为空。
	Reason string `json:"reason"`
}

// PublicKeyDetailOf 读取证书公钥的展开视图。
//
// 失败时返回 INVALID_CERT：公钥读不出意味着证书本身不可用，
// 没有降级展示的意义。
func PublicKeyDetailOf(c *Certificate) (*PublicKeyDetail, error) {
	if c == nil || c.cert == nil {
		return nil, toolerrInvalidCert("certificate already closed")
	}

	pub, err := c.cert.PublicKey()
	if err != nil {
		return nil, toolerrInvalidCert("cannot read public key")
	}
	defer func() { _ = asym.Close(pub) }()

	pemBytes, err := pub.MarshalPublicKeyPEM()
	if err != nil {
		return nil, toolerrInvalidCert("cannot export public key")
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, toolerrInvalidCert("cannot decode public key")
	}

	sum, err := cryptosvc.Digest("SHA256", block.Bytes)
	if err != nil {
		return nil, err
	}

	out := &PublicKeyDetail{
		Algorithm:   algorithmFromUpstream(pub.Algorithm().String()),
		PEM:         string(pemBytes),
		SPKIHex:     hex.EncodeToString(block.Bytes),
		Fingerprint: hex.EncodeToString(sum),
	}

	params, err := asym.Params(pub)
	if err == nil && params != nil {
		out.Bits = rsaBits(params)
		out.Curve = curveFromUpstream(params.Curve)
		out.Modulus = bigHex(params.N)
		out.Exponent = bigHex(params.E)
		out.X = bigHex(params.X)
		out.Y = bigHex(params.Y)
	}
	return out, nil
}

// PoliciesOf 读取证书的策略列表。
//
// 证书没有该扩展或解析失败时返回 nil（不是错误）：
// 策略是可选信息，不影响证书本身的可用性。
func PoliciesOf(c *Certificate) []PolicyInfo {
	if c == nil || c.cert == nil {
		return nil
	}
	policies, err := c.cert.CertificatePolicies()
	if err != nil || len(policies) == 0 {
		return nil
	}

	out := make([]PolicyInfo, 0, len(policies))
	for _, item := range policies {
		if len(out) >= maxPolicyItems {
			break
		}
		info := PolicyInfo{
			OID:     item.OID,
			CPSURIs: append([]string{}, item.CPSURIs...),
		}
		if item.UserNotice != nil {
			info.Organization = item.UserNotice.Organization
			info.NoticeNumbers = append([]int{}, item.UserNotice.NoticeNumbers...)
			info.ExplicitText = item.UserNotice.ExplicitText
		}
		out = append(out, info)
	}
	return out
}

// fillExtensionTexts 回填扩展的可读文本。
//
// 文本按 OpenSSL `x509 -text` 的体例生成，但**不**依赖外部 openssl：
// 已解码的字段（SAN / KU / EKU / BasicConstraints / SKID / AKID / 策略）
// 直接取自上游；只有 SAN / AIA / CRLDP 的**条目类型**需要额外结构解析，
// 由 parseGeneralNames / parseAIA / parseCRLDistributionPoints 完成
// （只读结构解析，见 docs/standards/03-crypto.md §8.6）。
//
// raw 与 info.Extensions 顺序一致（都来自 c.cert.Extensions()）。
func fillExtensionTexts(info *CertInfo, raw []x509.Extension) {
	for i := range info.Extensions {
		if i >= len(raw) {
			break
		}
		info.Extensions[i].Text = extensionText(info, raw[i])
	}
}

// extensionText 生成单个扩展的可读文本。
//
// 未知扩展返回空串：调用方回落到十六进制原值。
func extensionText(info *CertInfo, raw x509.Extension) string {
	switch raw.OID {
	case oidBasicConstraints:
		return basicConstraintsText(info)
	case oidKeyUsage:
		return strings.Join(keyUsageDisplayNames(info.KeyUsage), ", ")
	case oidExtKeyUsage:
		return strings.Join(extKeyUsageDisplayNames(info.ExtendedKeyUsage), ", ")
	case oidSubjectAltName:
		// 主体备用名直接用上游已解码的值（带类型前缀，与 SAN 区一致）。
		return strings.Join(info.SAN, ", ")
	case oidIssuerAltName:
		// 签发者备用名上游未单独暴露，需要自行解析。
		return strings.Join(generalNameTexts(raw.Data, nil), ", ")
	case oidSubjectKeyID:
		return keyIDText(info.SubjectKeyID)
	case oidAuthorityKeyID:
		return keyIDText(info.AuthorityKeyID)
	case oidCertificatePolicies:
		return policiesText(info.Policies)
	case oidAuthorityInfoAccess:
		return strings.Join(aiaTexts(raw.Data), ", ")
	case oidCRLDistributionPoints:
		return strings.Join(crlDistributionTexts(raw.Data), ", ")
	default:
		return ""
	}
}

// basicConstraintsText 生成 BasicConstraints 的可读文本。
func basicConstraintsText(info *CertInfo) string {
	if info.IsCA {
		if info.PathLen >= 0 {
			return fmt.Sprintf("CA:TRUE, pathlen:%d", info.PathLen)
		}
		return "CA:TRUE"
	}
	return "CA:FALSE"
}

// keyUsageDisplayNames 把 KU 短名映射为 OpenSSL 的显示名。
//
// 上游返回 RFC 5280 的短名（digitalSignature 等），
// 文本视图按 OpenSSL 惯例展示（Digital Signature 等）。
func keyUsageDisplayNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if display, ok := keyUsageDisplayNamesMap[name]; ok {
			out = append(out, display)
			continue
		}
		out = append(out, name)
	}
	return out
}

// keyUsageDisplayNamesMap 是 KU 短名到 OpenSSL 显示名的映射。
var keyUsageDisplayNamesMap = map[string]string{
	"digitalSignature": "Digital Signature",
	"nonRepudiation":   "Non Repudiation",
	"keyEncipherment":  "Key Encipherment",
	"dataEncipherment": "Data Encipherment",
	"keyAgreement":     "Key Agreement",
	"keyCertSign":      "Certificate Sign",
	"cRLSign":          "CRL Sign",
	"encipherOnly":     "Encipher Only",
	"decipherOnly":     "Decipher Only",
}

// extKeyUsageDisplayNames 把 EKU 短名映射为 OpenSSL 的显示名。
func extKeyUsageDisplayNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if display, ok := extKeyUsageDisplayNamesMap[name]; ok {
			out = append(out, display)
			continue
		}
		out = append(out, name)
	}
	return out
}

// extKeyUsageDisplayNamesMap 是 EKU 短名到 OpenSSL 显示名的映射。
var extKeyUsageDisplayNamesMap = map[string]string{
	"serverAuth":          "TLS Web Server Authentication",
	"clientAuth":          "TLS Web Client Authentication",
	"codeSigning":         "Code Signing",
	"emailProtection":     "E-mail Protection",
	"timeStamping":        "Time Stamping",
	"OCSPSigning":         "OCSP Signing",
	"anyExtendedKeyUsage": "Any Extended Key Usage",
}

// keyIDText 生成密钥标识符的可读文本（keyid:AB:CD）。
func keyIDText(keyID string) string {
	if keyID == "" {
		return ""
	}
	return "keyid:" + strings.ToUpper(keyID)
}

// policiesText 生成证书策略的可读文本。
func policiesText(policies []PolicyInfo) string {
	parts := make([]string, 0, len(policies))
	for _, item := range policies {
		part := "Policy: " + item.OID
		if len(item.CPSURIs) > 0 {
			part += ", CPS: \"" + strings.Join(item.CPSURIs, ", ") + "\""
		}
		if item.Organization != "" {
			part += ", Organization: \"" + item.Organization + "\""
		}
		if len(item.NoticeNumbers) > 0 {
			numbers := make([]string, 0, len(item.NoticeNumbers))
			for _, number := range item.NoticeNumbers {
				numbers = append(numbers, fmt.Sprintf("%d", number))
			}
			part += ", Notice Numbers: " + strings.Join(numbers, ", ")
		}
		if item.ExplicitText != "" {
			part += ", Explicit Text: \"" + item.ExplicitText + "\""
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

// generalNameTexts 生成 GeneralNames 的可读文本。
//
// 主体备用名（SAN）**不**走这里：上游的 Certificate.SAN() 已经给出
// 带类型前缀的条目（"DNS:…" / "IP:…" / "email:…"，与 OpenSSL 体例一致），
// 直接用上游的值即可。
//
// 只有上游未暴露的场景（签发者备用名、AIA 的 accessLocation）才自行解析。
func generalNameTexts(data []byte, fallback []string) []string {
	names, ok := parseGeneralNames(data)
	if !ok || len(names) == 0 {
		return fallback
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, name.Kind+":"+name.Value)
	}
	return out
}

// aiaTexts 生成 authorityInfoAccess 的可读文本。
func aiaTexts(data []byte) []string {
	out, ok := parseAIA(data)
	if !ok {
		return nil
	}
	return out
}

// crlDistributionTexts 生成 cRLDistributionPoints 的可读文本。
func crlDistributionTexts(data []byte) []string {
	out, ok := parseCRLDistributionPoints(data)
	if !ok {
		return nil
	}
	return out
}

// generalName 是 GeneralName 的一条条目。
type generalName struct {
	// Kind 是类型短名（DNS / IP Address / email / URI …）。
	Kind string
	// Value 是条目值。
	Value string
}

// parseGeneralNames 解析 GeneralNames（AIA 的 accessLocation 共用）。
//
// data 是 GeneralNames 的完整 DER（外层是 SEQUENCE）。
// 解析失败时返回 false，调用方回落到其它来源。
func parseGeneralNames(data []byte) ([]generalName, bool) {
	content, ok := derSequenceContent(data)
	if !ok {
		return nil, false
	}
	items, ok := derItems(content)
	if !ok {
		return nil, false
	}

	out := make([]generalName, 0, len(items))
	for _, item := range items {
		if item.Class != asn1.ClassContextSpecific {
			continue
		}
		if len(out) >= maxExtTextItems {
			break
		}
		switch item.Tag {
		case 1:
			out = append(out, generalName{Kind: "email", Value: string(item.Bytes)})
		case 2:
			out = append(out, generalName{Kind: "DNS", Value: string(item.Bytes)})
		case 4:
			out = append(out, generalName{Kind: "DirName", Value: hex.EncodeToString(item.FullBytes)})
		case 6:
			out = append(out, generalName{Kind: "URI", Value: string(item.Bytes)})
		case 7:
			out = append(out, generalName{Kind: "IP Address", Value: ipText(item.Bytes)})
		case 8:
			out = append(out, generalName{Kind: "Registered ID", Value: oidText(item.Bytes)})
		default:
			out = append(out, generalName{Kind: "othername", Value: hex.EncodeToString(item.FullBytes)})
		}
	}
	return out, true
}

// parseAIA 解析 authorityInfoAccess 扩展。
//
// 结构：SEQUENCE OF AccessDescription{ accessMethod OID, accessLocation GeneralName }。
func parseAIA(data []byte) ([]string, bool) {
	content, ok := derSequenceContent(data)
	if !ok {
		return nil, false
	}
	descriptions, ok := derItems(content)
	if !ok {
		return nil, false
	}

	out := make([]string, 0, len(descriptions))
	for _, description := range descriptions {
		if len(out) >= maxExtTextItems {
			break
		}
		parts, ok := derItems(description.Bytes)
		if !ok || len(parts) < 2 {
			continue
		}
		method, ok := asn1ObjectIdentifier(parts[0].FullBytes)
		if !ok {
			continue
		}
		location := parts[1]
		if location.Class != asn1.ClassContextSpecific || location.Tag != 6 {
			continue
		}
		out = append(out, aiaMethodName(method)+": "+string(location.Bytes))
	}
	return out, true
}

// aiaMethodName 把 AIA 的访问方法 OID 映射为 OpenSSL 显示名。
func aiaMethodName(oid string) string {
	switch oid {
	case "1.3.6.1.5.5.7.48.1":
		return "OCSP - URI:"
	case "1.3.6.1.5.5.7.48.2":
		return "CA Issuers - URI:"
	default:
		return oid
	}
}

// parseCRLDistributionPoints 解析 cRLDistributionPoints 扩展。
//
// 结构：SEQUENCE OF DistributionPoint{ distributionPoint [0] DistributionPointName }，
// 其中 DistributionPointName 的 fullName [0] 是 GeneralNames。
func parseCRLDistributionPoints(data []byte) ([]string, bool) {
	content, ok := derSequenceContent(data)
	if !ok {
		return nil, false
	}
	points, ok := derItems(content)
	if !ok {
		return nil, false
	}

	out := make([]string, 0, len(points))
	for _, point := range points {
		if len(out) >= maxExtTextItems {
			break
		}
		fields, ok := derItems(point.Bytes)
		if !ok || len(fields) == 0 {
			continue
		}
		name := fields[0]
		if name.Class != asn1.ClassContextSpecific || name.Tag != 0 {
			continue
		}
		// fullName 是 [0] 内的 GeneralNames（再嵌一层 [0]）。
		inner, ok := derItems(name.Bytes)
		if !ok || len(inner) == 0 {
			continue
		}
		names, ok := parseGeneralNames(inner[0].FullBytes)
		if !ok {
			continue
		}
		parts := make([]string, 0, len(names))
		for _, item := range names {
			parts = append(parts, item.Kind+":"+item.Value)
		}
		out = append(out, strings.Join(parts, ", "))
	}
	return out, true
}

// derSequenceContent 取出一个 SEQUENCE 的内容八位组。
//
// data 必须是 SEQUENCE 的完整 DER；结构不符或解析失败时返回 false。
func derSequenceContent(data []byte) ([]byte, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var seq asn1.RawValue
	if _, err := asn1.Unmarshal(data, &seq); err != nil {
		return nil, false
	}
	if seq.Class != asn1.ClassUniversal || seq.Tag != asn1.TagSequence {
		return nil, false
	}
	return seq.Bytes, true
}

// derItems 把一段 DER 拆成若干 TLV 元素。
//
// 与 `asn1.Unmarshal(data, &[]RawValue)` 不同，这里**不**做标签匹配：
// 逐个读取 TLV，天然支持上下文特定标签开头的元素
// （如 GeneralNames 的 [2] / [7]）。
//
// 解析失败时返回 false：调用方回落到其它来源。
func derItems(data []byte) ([]asn1.RawValue, bool) {
	out := make([]asn1.RawValue, 0, 8)
	rest := data
	for len(rest) > 0 {
		var item asn1.RawValue
		var err error
		rest, err = asn1.Unmarshal(rest, &item)
		if err != nil {
			return nil, false
		}
		if len(item.FullBytes) == 0 {
			return nil, false
		}
		out = append(out, item)
	}
	return out, true
}

// asn1ObjectIdentifier 解析一个 OID 的点分形式。
func asn1ObjectIdentifier(der []byte) (string, bool) {
	var oid asn1.ObjectIdentifier
	if _, err := asn1.Unmarshal(der, &oid); err != nil {
		return "", false
	}
	return oid.String(), true
}

// oidText 解析**隐式标记**的 OID 内容八位组。
//
// 隐式标记的上下文标签里装的是 OID 的内容（不含 tag / 长度），
// 需要补回 OID 头才能交给标准库解析。
func oidText(content []byte) string {
	if len(content) == 0 || len(content) > 127 {
		return ""
	}
	full := make([]byte, 0, len(content)+2)
	full = append(full, 0x06, byte(len(content)))
	full = append(full, content...)

	oid, ok := asn1ObjectIdentifier(full)
	if !ok {
		return ""
	}
	return oid
}

// ipText 把 iPAddress 的原始字节渲染为文本。
//
// 4 字节是 IPv4，16 字节是 IPv6；其它长度视为未知。
func ipText(raw []byte) string {
	switch len(raw) {
	case net.IPv4len, net.IPv6len:
		return net.IP(raw).String()
	default:
		return hex.EncodeToString(raw)
	}
}

// bigHex 把大整数渲染为大写十六进制。
//
// nil 或零值返回空串；位数补齐到偶数，便于按字节阅读。
func bigHex(value *big.Int) string {
	if value == nil || value.Sign() == 0 {
		return ""
	}
	text := strings.ToUpper(value.Text(16))
	if len(text)%2 == 1 {
		text = "0" + text
	}
	return text
}

// hexColon 把十六进制串渲染成冒号分隔的大写字节。
func hexColon(text string) string {
	if text == "" {
		return ""
	}
	upper := strings.ToUpper(text)
	parts := make([]string, 0, len(upper)/2)
	for i := 0; i+1 < len(upper); i += 2 {
		parts = append(parts, upper[i:i+2])
	}
	return strings.Join(parts, ":")
}

// CertTextView 生成证书的文本视图。
//
// 体例对齐 OpenSSL `x509 -text`（英文），便于与命令行输出对照；
// 但**不**调用外部 openssl：全部内容由本包已解码的字段拼装，
// 因此容器内无需额外二进制。
//
// pemText 是证书的 PEM 文本（用于「原始证书」页签的下载内容），
// 为空时不影响文本视图。
func CertTextView(info *CertInfo, detail *PublicKeyDetail, pemText string) string {
	if info == nil {
		return ""
	}

	var b strings.Builder
	write := func(format string, args ...any) {
		fmt.Fprintf(&b, format, args...)
	}

	write("Certificate:\n")
	write("    Data:\n")
	write("        Version: %d (0x%x)\n", info.Version+1, info.Version)
	write("        Serial Number:\n")
	write("            %s\n", hexColon(info.Serial))
	write("        Signature Algorithm: %s\n", info.SignatureAlgorithm)
	write("        Issuer: %s\n", info.IssuerText)
	write("        Validity\n")
	write("            Not Before: %s\n", opensslTime(info.NotBefore))
	write("            Not After : %s\n", opensslTime(info.NotAfter))
	write("        Subject: %s\n", info.SubjectText)

	if detail != nil {
		write("        Subject Public Key Info:\n")
		write("            Public Key Algorithm: %s\n", publicKeyAlgorithmName(detail))
		if detail.Bits > 0 {
			write("                Public-Key: (%d bit)\n", detail.Bits)
		}
		if detail.Modulus != "" {
			write("                Modulus:\n")
			for _, line := range wrapHex(detail.Modulus, 32) {
				write("                    %s\n", line)
			}
			write("                Exponent: %s\n", detail.Exponent)
		}
		if detail.X != "" && detail.Y != "" {
			write("                pub:\n")
			for _, line := range wrapHex(detail.X+detail.Y, 32) {
				write("                    %s\n", line)
			}
		}
	}

	if len(info.Extensions) > 0 {
		write("        X509v3 extensions:\n")
		for _, ext := range info.Extensions {
			name := ext.Name
			if name == "" {
				name = ext.OID
			}
			write("            %s", name)
			if ext.Critical {
				write(": critical")
			}
			write("\n")
			text := ext.Text
			if text == "" {
				text = hexColon(ext.Value)
			}
			for _, line := range wrapText(text, 64) {
				write("                %s\n", line)
			}
		}
	}

	write("    Signature Algorithm: %s\n", info.SignatureAlgorithm)
	if info.SignatureHex != "" {
		write("    Signature Value:\n")
		for _, line := range wrapHex(info.SignatureHex, 32) {
			write("        %s\n", line)
		}
	}

	out := b.String()
	if len(out) > CertTextMaxLen {
		out = out[:CertTextMaxLen] + "\n... (truncated)\n"
	}
	_ = pemText
	return out
}

// publicKeyAlgorithmName 返回公钥算法的 OpenSSL 显示名。
func publicKeyAlgorithmName(detail *PublicKeyDetail) string {
	switch detail.Algorithm {
	case AlgRSA:
		return "rsaEncryption"
	case AlgEC, AlgSM2:
		return "id-ecPublicKey"
	default:
		return detail.Algorithm
	}
}

// opensslTime 按 OpenSSL 的时间格式渲染（Jan  2 15:04:05 2006 MST）。
func opensslTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("Jan _2 15:04:05 2006 MST")
}

// wrapHex 把十六进制串按每行 n 个字节拆成冒号分隔的行。
func wrapHex(text string, perLine int) []string {
	if text == "" {
		return nil
	}
	upper := strings.ToUpper(text)
	bytes := make([]string, 0, len(upper)/2)
	for i := 0; i+1 < len(upper); i += 2 {
		bytes = append(bytes, upper[i:i+2])
	}

	lines := make([]string, 0, len(bytes)/perLine+1)
	for start := 0; start < len(bytes); start += perLine {
		end := start + perLine
		if end > len(bytes) {
			end = len(bytes)
		}
		lines = append(lines, strings.Join(bytes[start:end], ":"))
	}
	return lines
}

// wrapText 把长文本按每行 n 个字符拆行。
func wrapText(text string, perLine int) []string {
	if text == "" {
		return nil
	}
	lines := make([]string, 0, len(text)/perLine+1)
	for len(text) > perLine {
		lines = append(lines, text[:perLine])
		text = text[perLine:]
	}
	if text != "" {
		lines = append(lines, text)
	}
	return lines
}
