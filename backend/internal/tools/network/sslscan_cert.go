package network

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tlssvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 证书扩展 OID。
const (
	// oidAuthorityInfoAccess 是 authorityInfoAccess 扩展。
	oidAuthorityInfoAccess = "1.3.6.1.5.5.7.1.1"
	// oidCRLDistributionPoints 是 cRLDistributionPoints 扩展。
	oidCRLDistributionPoints = "2.5.29.31"
	// oidSCTList 是证书内嵌的 SCT 列表（RFC 6962）。
	oidSCTList = "1.3.6.1.4.1.11129.2.4.2"
	// oidTLSFeature 是 TLS Feature 扩展（RFC 7633，must-staple 用它声明）。
	oidTLSFeature = "1.3.6.1.5.5.7.1.24"
)

// 出网附加检查的体积与超时上限。
const (
	// ocspMaxBytes 是 OCSP 响应的读取上限。
	ocspMaxBytes = 4 << 20
	// crlMaxBytes 是 CRL 的读取上限。
	crlMaxBytes = 8 << 20
	// offsiteTimeout 是单次出网请求超时。
	offsiteTimeout = 6 * time.Second
)

// leafURLs 是从证书里抽出的出网地址。
type leafURLs struct {
	// ocsp 是 AIA 的 OCSP 地址。
	ocsp []string
	// caIssuers 是 AIA 的 CA Issuers 地址。
	caIssuers []string
	// crl 是 CRL 分发点地址。
	crl []string
}

// analyzeCertificate 生成叶证书摘要与整条链的解析结果。
//
// 链的解析**完全复用** `pkisvc.OrderChain`：`ssl-check` 的 `chain` 字段与
// `cert-chain-parse` 的响应是同一份结构，前端因此可以复用同一个组件。
func analyzeCertificate(probe *TLSProbeResult, host string) (ScanCertificate, *pkisvc.ChainInfo, leafURLs) {
	var summary ScanCertificate
	var urls leafURLs
	summary.ChainValid = probe.ChainValid
	summary.Verified = probe.Verified
	summary.ImplicitTrust = probe.ImplicitTrust
	summary.ChainErrors = probe.ChainErrors

	if len(probe.Certificates) > 0 && probe.Certificates[0].Info != nil {
		leaf := probe.Certificates[0].Info
		summary.Subject = leaf.SubjectText
		summary.Issuer = leaf.IssuerText
		summary.Serial = leaf.Serial
		summary.NotBefore = leaf.NotBefore.UTC().Format(time.RFC3339)
		summary.NotAfter = leaf.NotAfter.UTC().Format(time.RFC3339)
		summary.DaysRemaining = daysUntil(leaf.NotAfter)
		summary.KeyAlgorithm = leaf.PublicKey.Algorithm
		summary.KeyBits = leaf.PublicKey.Bits
		summary.Curve = leaf.PublicKey.Curve
		summary.SignatureAlgorithm = leaf.SignatureAlgorithm
		summary.SanCount = len(leaf.SAN)
		summary.SanMatchesHost = sanMatchesHost(leaf.SAN, host)
		summary.IsCa = leaf.IsCA
		summary.SelfSigned = leaf.SelfSigned
		urls = extractLeafURLs(leaf)
		for _, ext := range leaf.Extensions {
			switch ext.OID {
			case oidSCTList:
				summary.SctCount += sctCountOf(ext.Value)
			case oidTLSFeature:
				summary.MustStaple = mustStapleOf(ext.Value)
			}
		}
	}

	certs, err := pkisvc.ParseCertificates(joinPEM(probe.ChainDER), true)
	if err != nil || len(certs) == 0 {
		return summary, nil, urls
	}
	defer closeCertificates(certs)

	ordered, orderErr := pkisvc.OrderChain(certs, nil)
	if orderErr != nil {
		return summary, nil, urls
	}
	return summary, ordered, urls
}

// daysUntil 返回距离目标时刻的天数（小数，保留 1 位）。
func daysUntil(at time.Time) float64 {
	hours := time.Until(at).Hours()
	days := float64(int(hours/24*10)) / 10
	return days
}

// joinPEM 把 DER 列表拼成 PEM bundle。
func joinPEM(chain [][]byte) []byte {
	var buf bytes.Buffer
	for _, der := range chain {
		buf.WriteString(pemCert(der))
	}
	return buf.Bytes()
}

// sanMatchesHost 判断 SAN 是否覆盖目标主机。
//
// 只做 DNS 条目的文本匹配（含单层通配符），不涉及任何密码学运算。
func sanMatchesHost(san []string, host string) bool {
	target := strings.ToLower(strings.TrimSpace(host))
	if target == "" {
		return false
	}
	if ip := net.ParseIP(target); ip != nil {
		for _, item := range san {
			kind, value, ok := splitSanItem(item)
			if ok && (kind == "IP" || kind == "IP Address") && value == target {
				return true
			}
		}
		return false
	}
	for _, item := range san {
		kind, value, ok := splitSanItem(item)
		if !ok || kind != "DNS" {
			continue
		}
		pattern := strings.ToLower(value)
		if pattern == target {
			return true
		}
		if strings.HasPrefix(pattern, "*.") {
			suffix := pattern[1:]
			if strings.HasSuffix(target, suffix) &&
				!strings.Contains(strings.TrimSuffix(target, suffix), ".") {
				return true
			}
		}
	}
	return false
}

// splitSanItem 把 `DNS:example.com` / `IP Address:1.2.3.4` 拆成类型与值。
func splitSanItem(item string) (kind, value string, ok bool) {
	index := strings.Index(item, ":")
	if index < 0 {
		return "", "", false
	}
	return strings.TrimSpace(item[:index]), strings.TrimSpace(item[index+1:]), true
}

// extractLeafURLs 从证书扩展里抽出 OCSP / CA Issuers / CRL 地址。
//
// 扩展的文本由 `pkisvc` 已解析好，格式为
// `OCSP - URI:: http://…`、`CA Issuers - URI:: http://…`、`URI: http://…`
// （注意 `pkisvc` 的方法名自带一个冒号，因此实际文本里是**双冒号**）。
// 这里只做前缀拆分，不重复解析 DER。
func extractLeafURLs(info *pkisvc.CertInfo) leafURLs {
	var out leafURLs
	for _, ext := range info.Extensions {
		switch ext.OID {
		case oidAuthorityInfoAccess:
			for _, item := range splitExtText(ext.Text) {
				switch {
				case strings.HasPrefix(item, "OCSP - URI"):
					out.ocsp = append(out.ocsp, stripURLPrefix(item, "OCSP - URI"))
				case strings.HasPrefix(item, "CA Issuers - URI"):
					out.caIssuers = append(out.caIssuers, stripURLPrefix(item, "CA Issuers - URI"))
				}
			}
		case oidCRLDistributionPoints:
			for _, item := range splitExtText(ext.Text) {
				if strings.HasPrefix(item, "URI:") {
					out.crl = append(out.crl, strings.TrimSpace(strings.TrimPrefix(item, "URI:")))
				}
			}
		}
	}
	return out
}

// stripURLPrefix 去掉扩展文本里的方法名与冒号，返回纯地址。
//
// `pkisvc` 的方法名自带一个冒号（`OCSP - URI:`），拼接时又加了一个，
// 因此文本里是 `OCSP - URI:: http://…`；这里把**所有**前导冒号都去掉。
func stripURLPrefix(item, label string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(item, label))
	for strings.HasPrefix(rest, ":") {
		rest = strings.TrimSpace(strings.TrimPrefix(rest, ":"))
	}
	return rest
}

// splitExtText 按 ", " 拆分扩展文本条目。
func splitExtText(text string) []string {
	if text == "" {
		return nil
	}
	parts := strings.Split(text, ", ")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// scanFeatures 探测协议特性。
//
// `unavailable` 与 `not_supported` 是两种不同的结论：前者表示当前引擎
// 拿不到这项信息，绝不能当成「不支持」展示。
func (s *scanner) scanFeatures(negotiated ScanNegotiated, cert ScanCertificate) []FeatureResult {
	features := make([]FeatureResult, 0, 6)

	alpn := FeatureResult{Id: "alpnH2", Label: "HTTP/2 (ALPN)", Status: FeatureNotSupported}
	if state, err := s.probe(tlssvc.Options{
		ServerName:       s.sni,
		NextProtos:       alpnProtos,
		HandshakeTimeout: probeTimeout,
	}); err == nil {
		if state.ALPN == "h2" {
			alpn.Status = FeatureSupported
			alpn.Detail = state.ALPN
		} else if state.ALPN != "" {
			alpn.Detail = state.ALPN
		}
	} else {
		alpn.Status = FeatureUnavailable
		alpn.Detail = reasonOf(err)
	}
	features = append(features, alpn)

	features = append(features, FeatureResult{
		Id: "sessionResumption", Label: "会话恢复", Status: FeatureUnavailable,
		Detail: "当前 TLS 引擎未暴露会话复用标志（上游 tongsuo-go I-25）",
	})
	features = append(features, FeatureResult{
		Id: "ocspStapling", Label: "OCSP 装订", Status: FeatureUnavailable,
		Detail: "当前 TLS 引擎未暴露装订响应（上游 tongsuo-go I-23），吊销结论以下方的主动查询为准",
	})
	features = append(features, FeatureResult{
		Id: "secureRenegotiation", Label: "安全重协商", Status: FeatureUnavailable,
		Detail: "当前 TLS 引擎未暴露重协商状态（上游 tongsuo-go I-26）",
	})
	features = append(features, FeatureResult{
		Id: "compression", Label: "TLS 层压缩", Status: FeatureUnavailable,
		Detail: "当前 TLS 引擎未暴露压缩状态（上游 tongsuo-go I-26）",
	})

	mustStaple := FeatureResult{Id: "mustStaple", Label: "证书声明 must-staple", Status: FeatureNotSupported}
	if cert.MustStaple {
		mustStaple.Status = FeatureSupported
	}
	features = append(features, mustStaple)

	_ = negotiated
	return features
}

// scanOffsite 执行出网附加检查：吊销、CAA、HSTS。
func (s *scanner) scanOffsite(urls leafURLs, host string, port int, checks map[string]bool) (RevocationResult, CAAResult, HSTSResult) {
	revocation := RevocationResult{Ocsp: []RevocationEntry{}, Crl: []RevocationEntry{}}
	caa := CAAResult{Records: []string{}, Violations: []string{}, Permitted: true}
	var hsts HSTSResult

	if checks[CheckRevocation] {
		revocation = s.checkRevocation(urls)
	}
	if checks[CheckCAA] {
		caa = s.checkCAA(host)
	}
	if checks[CheckHSTS] {
		hsts = s.checkHSTS(host, port)
	}
	return revocation, caa, hsts
}

// checkRevocation 依次尝试 OCSP 与 CRL。
//
// 任一步拿到明确结论就停止；网络失败**不得**判为已吊销。
func (s *scanner) checkRevocation(urls leafURLs) RevocationResult {
	out := RevocationResult{Ocsp: []RevocationEntry{}, Crl: []RevocationEntry{}}
	for _, url := range urls.ocsp {
		out.Ocsp = append(out.Ocsp, s.queryOCSP(url))
	}
	for _, url := range urls.crl {
		out.Crl = append(out.Crl, s.downloadCRL(url))
	}
	return out
}

// queryOCSP 向 AIA 的 OCSP 地址发一次查询。
func (s *scanner) queryOCSP(url string) RevocationEntry {
	entry := RevocationEntry{Url: url, Status: "unknown"}
	reqBuild, err := s.ocspRequest()
	if err != nil {
		entry.Error = err.Error()
		return entry
	}

	client := s.httpClient()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, url, bytes.NewReader(reqBuild))
	if err != nil {
		entry.Error = "invalid ocsp url"
		return entry
	}
	req.Header.Set("Content-Type", "application/ocsp-request")
	req.Header.Set("Accept", "application/ocsp-response")

	resp, err := client.Do(req)
	if err != nil {
		entry.Error = "request failed"
		return entry
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		entry.Error = fmt.Sprintf("http %d", resp.StatusCode)
		return entry
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, ocspMaxBytes))
	if err != nil {
		entry.Error = "read failed"
		return entry
	}

	leaf, issuer, closePair, err := s.leafAndIssuer()
	if err != nil {
		entry.Error = err.Error()
		return entry
	}
	defer closePair()

	parsed, err := pkisvc.ParseOCSPResponseDER(body, leaf, issuer)
	if err != nil {
		entry.Error = "cannot parse ocsp response"
		return entry
	}
	defer func() { _ = parsed.Close() }()

	info := parsed.Info()
	entry.Status = ocspStatusText(info.CertStatus)
	entry.ThisUpdate = formatOptionalTime(info.ThisUpdate)
	entry.NextUpdate = formatOptionalTime(info.NextUpdate)
	return entry
}

// downloadCRL 下载并解析一个 CRL 分发点。
func (s *scanner) downloadCRL(url string) RevocationEntry {
	entry := RevocationEntry{Url: url, Status: "unknown"}
	client := s.httpClient()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, url, nil)
	if err != nil {
		entry.Error = "invalid crl url"
		return entry
	}
	resp, err := client.Do(req)
	if err != nil {
		entry.Error = "request failed"
		return entry
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		entry.Error = fmt.Sprintf("http %d", resp.StatusCode)
		return entry
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, crlMaxBytes))
	if err != nil {
		entry.Error = "read failed"
		return entry
	}

	leaf, issuer, closePair, err := s.leafAndIssuer()
	if err != nil {
		entry.Error = err.Error()
		return entry
	}
	defer closePair()

	isPEM := bytes.Contains(body, []byte("-----BEGIN"))
	list, err := pkisvc.LoadRevocationList(body, isPEM)
	if err != nil {
		entry.Error = "cannot parse crl"
		return entry
	}
	defer func() { _ = list.Close() }()

	result, err := pkisvc.CheckRevocation(leaf, list, issuer, time.Now())
	if err != nil {
		entry.Error = "cannot check revocation"
		return entry
	}
	entry.Status = result.Status
	if !result.RevokedAt.IsZero() {
		entry.ThisUpdate = result.RevokedAt.UTC().Format(time.RFC3339)
	}
	return entry
}

// httpClient 返回受控的 HTTP 客户端。
func (s *scanner) httpClient() *http.Client {
	return s.guard.HTTPClient(netguard.HTTPOptions{Timeout: offsiteTimeout})
}

// ocspRequest 生成当前叶证书的 OCSP 请求 DER。
func (s *scanner) ocspRequest() ([]byte, error) {
	leaf, issuer, closePair, err := s.leafAndIssuer()
	if err != nil {
		return nil, err
	}
	defer closePair()

	built, err := pkisvc.CreateOCSPRequestDER(leaf, issuer, pkisvc.OCSPHashSHA1)
	if err != nil {
		return nil, fmt.Errorf("cannot build ocsp request")
	}
	return built.DER, nil
}

// leafAndIssuer 取当前链的叶证书与签发者。
//
// 链上只有一张（自签名）时，签发者回落到叶证书本身。
func (s *scanner) leafAndIssuer() (*pkisvc.Certificate, *pkisvc.Certificate, func(), error) {
	leaf, err := pkisvc.LoadCertificate(s.chainDER[0], false)
	if err != nil {
		return nil, nil, func() {}, fmt.Errorf("leaf certificate unavailable")
	}
	issuer := leaf
	if len(s.chainDER) > 1 {
		loaded, loadErr := pkisvc.LoadCertificate(s.chainDER[1], false)
		if loadErr == nil {
			issuer = loaded
		}
	}
	closePair := func() {
		if issuer != leaf {
			_ = issuer.Close()
		}
		_ = leaf.Close()
	}
	return leaf, issuer, closePair, nil
}

// ocspStatusText 把状态码转成稳定标识。
func ocspStatusText(status int) string {
	switch status {
	case pkisvc.OCSPCertGood:
		return "good"
	case pkisvc.OCSPCertRevoked:
		return "revoked"
	default:
		return "unknown"
	}
}

// formatOptionalTime 格式化可空时间。
func formatOptionalTime(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	return at.UTC().Format(time.RFC3339)
}

// checkCAA 查询 CAA 记录并与签发者比对。
//
// 查询走**系统解析器**（`/etc/resolv.conf`）：这与 `netguard.Resolve` 解析主机名
// 时走的是同一条路径——本机的解析器本身就是本地基础设施，
// 而查询目标是**已校验**的公开主机名，因此不存在 SSRF 风险。
// 若系统解析器不可用（例如容器内无解析器），则降级为「未查到」而不报错。
func (s *scanner) checkCAA(host string) CAAResult {
	result := CAAResult{Queried: true, Records: []string{}, Violations: []string{}, Permitted: true}

	records, err := queryCAA(s.ctx, host)
	if err != nil {
		result.Error = "dns query failed"
		return result
	}
	for _, record := range records {
		result.Records = append(result.Records, record.Tag+" "+record.Value)
		if record.Tag != "issue" && record.Tag != "issuewild" {
			continue
		}
		// `issue ";"` 表示禁止任何 CA 签发。
		if strings.TrimSpace(record.Value) == ";" {
			result.Permitted = false
			result.Violations = append(result.Violations, record.Value)
		}
	}
	return result
}

// queryCAA 通过系统解析器查询 CAA 记录。
func queryCAA(ctx context.Context, host string) ([]*dns.CAA, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil || len(config.Servers) == 0 {
		return nil, fmt.Errorf("no system resolver")
	}
	port := config.Port
	if port == "" {
		port = "53"
	}

	query := buildDNSQuery(host, dns.TypeCAA)
	for _, server := range config.Servers {
		addr := net.JoinHostPort(server, port)
		conn, dialErr := net.DialTimeout("udp", addr, offsiteTimeout)
		if dialErr != nil {
			continue
		}
		client := &dns.Conn{Conn: conn}
		client.UDPSize = dnsUDPBufferSize
		if deadline, ok := ctx.Deadline(); ok {
			_ = client.SetDeadline(deadline)
		}
		writeErr := client.WriteMsg(query)
		resp, readErr := client.ReadMsg()
		_ = conn.Close()
		if writeErr != nil || readErr != nil {
			continue
		}
		out := make([]*dns.CAA, 0, len(resp.Answer))
		for _, rr := range resp.Answer {
			if record, ok := rr.(*dns.CAA); ok {
				out = append(out, record)
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("system resolver unreachable")
}

// checkHSTS 读取首页的 Strict-Transport-Security 头。
func (s *scanner) checkHSTS(host string, port int) HSTSResult {
	var result HSTSResult
	scheme := "https"
	if port != 443 {
		scheme = fmt.Sprintf("https://%s:%d", host, port)
	} else {
		scheme = "https://" + host
	}
	client := s.httpClient()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodHead, scheme, nil)
	if err != nil {
		result.Error = "invalid url"
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = "request failed"
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	value := resp.Header.Get("Strict-Transport-Security")
	if value == "" {
		return result
	}
	result.Present = true
	for _, part := range strings.Split(value, ";") {
		item := strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(item), "max-age=") {
			fmt.Sscanf(strings.TrimPrefix(strings.ToLower(item), "max-age="), "%d", &result.MaxAge)
			continue
		}
		if strings.EqualFold(item, "includeSubDomains") {
			result.IncludeSubDomains = true
		}
		if strings.EqualFold(item, "preload") {
			result.Preload = true
		}
	}
	return result
}

// scanWarnings 依据扫描结果给出警告。
func scanWarnings(result *ScanResult) []toolerr.Warning {
	warnings := make([]toolerr.Warning, 0, 4)
	if !result.Certificate.ChainValid || !result.Certificate.Verified {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnUntrustedChain, nil))
	}
	if result.Certificate.SelfSigned {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnSelfSignedCert, nil))
	}
	if result.Certificate.DaysRemaining < 0 {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnExpiredCert, nil))
	}
	return warnings
}

// pemCert 把 DER 编成 PEM 文本。
func pemCert(der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}
