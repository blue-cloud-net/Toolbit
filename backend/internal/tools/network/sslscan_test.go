package network

import (
	"context"
	"encoding/hex"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tlssvc"
)

// selfSignedServer 在回环上起一个自签名 TLS 服务端。
//
// 返回监听地址、端口与关停函数。私钥在测试内生成，**不**落盘。
func selfSignedServer(t *testing.T, minVersion uint16) (string, int, func()) {
	t.Helper()

	pair, err := pkisvc.Generate("EC", 0, "P-256")
	if err != nil {
		t.Fatalf("生成密钥失败：%v", err)
	}

	serial, err := pkisvc.RandomSerial()
	if err != nil {
		t.Fatalf("生成序列号失败：%v", err)
	}
	cert, err := pkisvc.IssueCertificate(pkisvc.CertIssueOptions{
		Subject: []pkisvc.NameEntryInfo{
			{Field: "C", Value: "CN"},
			{Field: "O", Value: "Toolbit Test"},
			{Field: "CN", Value: "localhost"},
		},
		Serial:           serial,
		NotBefore:        time.Now().Add(-time.Hour),
		NotAfter:         time.Now().Add(30 * 24 * time.Hour),
		SAN:              []string{"DNS:localhost", "IP:127.0.0.1"},
		KeyUsage:         []string{"digitalSignature", "keyEncipherment"},
		ExtendedKeyUsage: []string{"serverAuth"},
		AutoKeyID:        true,
	}, pair.Public, pair.Private, nil)
	if err != nil {
		t.Fatalf("签发证书失败：%v", err)
	}

	certPEM, err := cert.MarshalPEM()
	if err != nil {
		t.Fatalf("导出证书失败：%v", err)
	}
	keyPEM, err := pair.Private.MarshalPKCS8PEM()
	if err != nil {
		t.Fatalf("导出私钥失败：%v", err)
	}

	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败：%v", err)
	}
	serveOpts := tlssvc.ServerOptions{
		CertPEM:    certPEM,
		KeyPEM:     keyPEM,
		MinVersion: minVersion,
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = tlssvc.Serve(raw, serveOpts)
	}()

	stopped := false
	cleanup := func() {
		if stopped {
			return
		}
		stopped = true
		_ = raw.Close()
		<-done
		_ = cert.Close()
		_ = pair.Close()
	}

	addr := raw.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port, cleanup
}

// newScanGuard 构造只放行回环与目标端口的出网防护。
func newScanGuard(t *testing.T, port int) *netguard.Guard {
	t.Helper()
	return netguard.New(netguard.Config{
		Enabled:    true,
		Allowlist:  []string{"127.0.0.1"},
		ExtraPorts: []int{port},
	})
}

// TestScanTLSProtocolsAndGrade 覆盖协议矩阵与自签名证书的评级。
func TestScanTLSProtocolsAndGrade(t *testing.T) {
	host, port, cleanup := selfSignedServer(t, tlssvc.VersionTLS12)
	defer cleanup()

	guard := newScanGuard(t, port)
	result, _, err := ScanTLS(context.Background(), guard, ScanOptions{
		Host:     host,
		Port:     port,
		SNI:      "localhost",
		Mode:     ScanModeFull,
		Checks:   []string{CheckProtocols, CheckCert},
		BudgetMs: 30000,
	}, nil)
	if err != nil {
		t.Fatalf("扫描失败：%v", err)
	}

	supported := map[string]bool{}
	for _, item := range result.Protocols {
		supported[item.Version] = item.Supported
	}
	if !supported["TLS1.2"] {
		t.Errorf("期望支持 TLS 1.2，实际 %v", result.Protocols)
	}
	if supported["TLS1.0"] {
		t.Errorf("期望不支持 TLS 1.0，实际 %v", result.Protocols)
	}

	if result.Chain == nil || result.Chain.Count == 0 {
		t.Fatalf("未解析出证书链：%+v", result.Chain)
	}
	if result.Certificate.SanMatchesHost != true {
		t.Errorf("SAN 应覆盖 localhost：%+v", result.Certificate)
	}
	// 自签名证书无法追溯到信任根，评级必须被 T 封顶。
	if result.Grade.Letter != "T" {
		t.Errorf("自签名证书的等级 = %s，期望 T", result.Grade.Letter)
	}
	if result.Chain.Chain[0].Role != "leaf" && result.Chain.Chain[0].Role != "root" {
		t.Errorf("链首角色 = %s", result.Chain.Chain[0].Role)
	}
}

// TestScanTLSCipherMatrix 覆盖套件矩阵与客户端模拟。
func TestScanTLSCipherMatrix(t *testing.T) {
	host, port, cleanup := selfSignedServer(t, tlssvc.VersionTLS12)
	defer cleanup()

	guard := newScanGuard(t, port)
	result, _, err := ScanTLS(context.Background(), guard, ScanOptions{
		Host:     host,
		Port:     port,
		SNI:      "localhost",
		Mode:     ScanModeFull,
		Checks:   []string{CheckProtocols, CheckCiphers, CheckClients},
		Clients:  []string{"go", "java8"},
		BudgetMs: 60000,
	}, nil)
	if err != nil {
		t.Fatalf("扫描失败：%v", err)
	}

	if len(result.Ciphers) == 0 {
		t.Fatal("套件矩阵为空")
	}
	accepted := 0
	for _, item := range result.Ciphers {
		if item.Verdict == SuiteSupported {
			accepted++
		}
		if item.Family == "" || item.Kex == "" {
			t.Errorf("套件描述不完整：%+v", item)
		}
	}
	if accepted == 0 {
		t.Error("本应至少有一个套件被接受")
	}

	if len(result.Clients) != 2 {
		t.Fatalf("客户端模拟数量 = %d，期望 2", len(result.Clients))
	}
	for _, item := range result.Clients {
		if item.Id == "go" && !item.Supported {
			t.Errorf("Go 客户端画像应支持，实际 %+v", item)
		}
		if item.Id == "java8" && item.Supported && item.NegotiatedVersion == "" {
			t.Errorf("成功握手应带协商版本：%+v", item)
		}
	}
}

// TestScanTLSRejectsUnknownCheck 覆盖检查项白名单。
func TestScanTLSRejectsUnknownCheck(t *testing.T) {
	err := ValidateScanOptions(&ScanOptions{Host: "example.com", Checks: []string{"bogus"}})
	if err == nil {
		t.Fatal("未知检查项应被拒绝")
	}
	err = ValidateScanOptions(&ScanOptions{Host: "example.com", Clients: []string{"nope"}})
	if err == nil {
		t.Fatal("未知客户端画像应被拒绝")
	}
	if err = ValidateScanOptions(&ScanOptions{Host: "example.com", BudgetMs: 999999}); err == nil {
		t.Fatal("超预算应被拒绝")
	}
}

// TestDescribeSuite 覆盖套件名解析。
func TestDescribeSuite(t *testing.T) {
	cases := []struct {
		name   string
		family string
		bits   int
		kex    string
	}{
		{"TLS_AES_128_GCM_SHA256", "AEAD", 128, "ECDHE"},
		{"ECDHE-RSA-AES256-GCM-SHA384", "AEAD", 256, "ECDHE"},
		{"ECDHE-RSA-AES128-SHA", "CBC", 128, "ECDHE"},
		{"DHE-RSA-AES256-SHA256", "CBC", 256, "DHE"},
		{"AES128-SHA", "CBC", 128, "其它"},
	}
	for _, item := range cases {
		family, bits, kex := describeSuite(item.name)
		if family != item.family || bits != item.bits || kex != item.kex {
			t.Errorf("%s → (%s, %d, %s)，期望 (%s, %d, %s)",
				item.name, family, bits, kex, item.family, item.bits, item.kex)
		}
	}
}

// TestSanMatchesHost 覆盖主机名匹配（含单层通配符）。
func TestSanMatchesHost(t *testing.T) {
	san := []string{"DNS:example.com", "DNS:*.example.com", "IP Address:1.2.3.4"}
	if !sanMatchesHost(san, "example.com") {
		t.Error("精确匹配失败")
	}
	if !sanMatchesHost(san, "a.example.com") {
		t.Error("通配符匹配失败")
	}
	if sanMatchesHost(san, "a.b.example.com") {
		t.Error("通配符不应跨层匹配")
	}
	if !sanMatchesHost(san, "1.2.3.4") {
		t.Error("IP 匹配失败")
	}
	if sanMatchesHost(san, "other.com") {
		t.Error("不应匹配无关域名")
	}
}

// TestGradeScanCaps 覆盖档位上限。
func TestGradeScanCaps(t *testing.T) {
	result := &ScanResult{
		Certificate: ScanCertificate{
			ChainValid:     true,
			Verified:       true,
			SanMatchesHost: true,
			DaysRemaining:  200,
			KeyAlgorithm:   "RSA",
			KeyBits:        2048,
		},
		Negotiated: ScanNegotiated{Cipher: "ECDHE-RSA-AES128-GCM-SHA256", KeyExchangeGroup: "X25519"},
		Protocols: []ProtocolResult{
			{Version: "TLS1.2", Supported: true},
		},
	}
	grade := gradeScan(result, map[string]bool{
		CheckProtocols: true, CheckCert: true, CheckCiphers: true,
		CheckFeatures: true, CheckClients: true, CheckRevocation: true,
		CheckCAA: true, CheckHSTS: true,
	})
	if grade.Letter != "A" {
		t.Errorf("仅支持 TLS 1.2 应封顶 A，实际 %s", grade.Letter)
	}
	if len(grade.Caps) == 0 || grade.Caps[0] != "noTLS13" {
		t.Errorf("caps = %v", grade.Caps)
	}

	result.Certificate.Verified = false
	grade = gradeScan(result, map[string]bool{CheckCert: true, CheckProtocols: true})
	if grade.Letter != "T" {
		t.Errorf("不可信证书应为 T，实际 %s", grade.Letter)
	}
}

// TestScanPortHelper 确保端口字符串化在测试里可用。
func TestScanPortHelper(t *testing.T) {
	if got := strconv.Itoa(443); got != "443" {
		t.Fatalf("strconv 异常：%s", got)
	}
}

// TestStripURLPrefix 覆盖扩展文本里的双冒号。
func TestStripURLPrefix(t *testing.T) {
	cases := []struct {
		item  string
		label string
		want  string
	}{
		{"OCSP - URI:: http://ocsp.example.com", "OCSP - URI", "http://ocsp.example.com"},
		{"OCSP - URI: http://ocsp.example.com", "OCSP - URI", "http://ocsp.example.com"},
		{
			"CA Issuers - URI:: http://ca.example.com/ca.crt",
			"CA Issuers - URI",
			"http://ca.example.com/ca.crt",
		},
	}
	for _, item := range cases {
		if got := stripURLPrefix(item.item, item.label); got != item.want {
			t.Errorf("%q → %q，期望 %q", item.item, got, item.want)
		}
	}
}

// TestMustStapleOf 覆盖 TLS Feature 扩展解析。
func TestMustStapleOf(t *testing.T) {
	// SEQUENCE { INTEGER 5 } —— status_request
	if !mustStapleOf("3003020105") {
		t.Error("status_request 应识别为 must-staple")
	}
	// SEQUENCE { INTEGER 17 } —— status_request_v2
	if !mustStapleOf("3003020111") {
		t.Error("status_request_v2 应识别为 must-staple")
	}
	// SEQUENCE { INTEGER 5, INTEGER 17 }
	if !mustStapleOf("3006020105020111") {
		t.Error("含 status_request 的序列应识别为 must-staple")
	}
	// SEQUENCE { INTEGER 41 } —— 其它特性
	if mustStapleOf("3003020129") {
		t.Error("无关特性不应识别为 must-staple")
	}
	if mustStapleOf("not-hex") || mustStapleOf("") {
		t.Error("无法解析时应返回 false")
	}
}

// TestSCTCountOf 覆盖 SCT 列表计数。
func TestSCTCountOf(t *testing.T) {
	// SignedCertificateTimestampList：总长度 10，内含两个 3 字节条目。
	list := []byte{0x00, 0x0a, 0x00, 0x03, 0x01, 0x02, 0x03, 0x00, 0x03, 0x04, 0x05, 0x06}
	if got := sctCountOf(hex.EncodeToString(list)); got != 2 {
		t.Errorf("SCT 条目数 = %d，期望 2", got)
	}
	if got := sctCountOf(""); got != 0 {
		t.Errorf("空输入 = %d，期望 0", got)
	}
	if got := sctCountOf("aabbcc"); got != 1 {
		t.Errorf("无法解析时应回落到 1，实际 %d", got)
	}
}

// TestExtractLeafURLs 覆盖 AIA 与 CRL 地址提取。
func TestExtractLeafURLs(t *testing.T) {
	info := &pkisvc.CertInfo{
		Extensions: []pkisvc.ExtInfo{
			{
				OID:  "1.3.6.1.5.5.7.1.1",
				Text: "CA Issuers - URI:: http://ca.example.com/ca.crt, OCSP - URI:: http://ocsp.example.com",
			},
			{
				OID:  "2.5.29.31",
				Text: "URI:http://crl.example.com/root.crl",
			},
		},
	}
	urls := extractLeafURLs(info)
	if len(urls.ocsp) != 1 || urls.ocsp[0] != "http://ocsp.example.com" {
		t.Errorf("ocsp = %v", urls.ocsp)
	}
	if len(urls.caIssuers) != 1 || urls.caIssuers[0] != "http://ca.example.com/ca.crt" {
		t.Errorf("caIssuers = %v", urls.caIssuers)
	}
	if len(urls.crl) != 1 || urls.crl[0] != "http://crl.example.com/root.crl" {
		t.Errorf("crl = %v", urls.crl)
	}
}
