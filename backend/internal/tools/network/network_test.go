package network

import (
	"context"
	"errors"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// offlineGuard 是一个**关闭出网**的守卫，用于验证 NETWORK_DISABLED 路径。
func offlineGuard() *netguard.Guard {
	return netguard.New(netguard.Config{Enabled: false})
}

// assertCode 断言错误携带指定错误码。
func assertCode(t *testing.T, err error, code toolerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望错误码 %s，实际无错误", code)
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("期望 *toolerr.Error，实际 %T", err)
	}
	if toolErr.Code != code {
		t.Fatalf("错误码 = %s，期望 %s", toolErr.Code, code)
	}
}

func TestCIDRCalcIPv4(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{CIDR: "192.168.1.0/24"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	out, ok := result.Data.(CIDRCalcResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	if out.Version != 4 {
		t.Errorf("version = %d，期望 4", out.Version)
	}
	if out.HostCount != "254" {
		t.Errorf("hostCount = %s，期望 254", out.HostCount)
	}
	if out.Broadcast != "192.168.1.255" {
		t.Errorf("broadcast = %s", out.Broadcast)
	}
	if out.FirstUsable != "192.168.1.1" || out.LastUsable != "192.168.1.254" {
		t.Errorf("可用地址区间不符：%s - %s", out.FirstUsable, out.LastUsable)
	}
	if out.Mask != "255.255.255.0" {
		t.Errorf("mask = %s", out.Mask)
	}
}

func TestCIDRCalcEdgePrefixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		cidr string
		host string
	}{
		{"10.0.0.0/32", "1"},
		{"10.0.0.0/31", "2"},
		{"10.0.0.0/30", "2"},
		{"10.0.0.0/0", "4294967294"},
	}
	for _, c := range cases {
		result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{CIDR: c.cidr})
		if err != nil {
			t.Fatalf("%s 计算失败：%v", c.cidr, err)
		}
		out := result.Data.(CIDRCalcResult)
		if out.HostCount != c.host {
			t.Errorf("%s hostCount = %s，期望 %s", c.cidr, out.HostCount, c.host)
		}
	}
}

func TestCIDRCalcIPv6(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{CIDR: "2001:db8::/64"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	if out.Version != 6 {
		t.Errorf("version = %d，期望 6", out.Version)
	}
	if out.Broadcast != "" {
		t.Errorf("IPv6 不应有广播地址，实际 %q", out.Broadcast)
	}
	// 2^64 超出 64 位整数，必须以十进制字符串精确表示。
	if out.HostCount != "18446744073709551616" {
		t.Errorf("hostCount = %s", out.HostCount)
	}
}

func TestCIDRCalcRangeToPrefixes(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{Start: "10.0.0.1", End: "10.0.0.6"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	want := []string{"10.0.0.1/32", "10.0.0.2/31", "10.0.0.4/31", "10.0.0.6/32"}
	if len(out.Range) != len(want) {
		t.Fatalf("range = %v，期望 %v", out.Range, want)
	}
	for i := range want {
		if out.Range[i] != want[i] {
			t.Fatalf("range = %v，期望 %v", out.Range, want)
		}
	}
}

func TestCIDRCalcSplit(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{CIDR: "10.0.0.0/24", Split: 26})
	if err != nil {
		t.Fatalf("拆分失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	if len(out.CIDRs) != 4 {
		t.Fatalf("拆分结果数量 = %d，期望 4", len(out.CIDRs))
	}
	if out.CIDRs[0] != "10.0.0.0/26" || out.CIDRs[3] != "10.0.0.192/26" {
		t.Fatalf("拆分结果不符：%v", out.CIDRs)
	}
}

func TestCIDRCalcSplitTooMany(t *testing.T) {
	t.Parallel()

	_, err := CIDRCalc(context.Background(), &CIDRCalcRequest{CIDR: "10.0.0.0/8", Split: 24})
	assertCode(t, err, toolerr.CodeTooManyItems)
}

func TestCIDRCalcInvalid(t *testing.T) {
	t.Parallel()

	req := &CIDRCalcRequest{CIDR: "999.1.1.0/24"}
	if err := req.Validate(); err != nil {
		t.Fatalf("格式错误应在执行阶段报错，Validate 只做字段组合校验：%v", err)
	}
	_, err := CIDRCalc(context.Background(), req)
	assertCode(t, err, toolerr.CodeInvalidCIDR)
}

func TestCIDRCalcValidate(t *testing.T) {
	t.Parallel()

	if err := (&CIDRCalcRequest{}).Validate(); err == nil {
		t.Fatal("空请求应当报缺少字段")
	}
	if err := (&CIDRCalcRequest{CIDR: "10.0.0.0/24", Start: "10.0.0.1"}).Validate(); err == nil {
		t.Fatal("cidr 与 start 混用应当报错")
	}
	if err := (&CIDRCalcRequest{Start: "10.0.0.1"}).Validate(); err == nil {
		t.Fatal("只给 start 应当报缺少 end")
	}
	if err := (&CIDRCalcRequest{Start: "10.0.0.1", End: "10.0.0.9", Split: 24}).Validate(); err == nil {
		t.Fatal("范围模式下不允许 split")
	}
}

func TestDNSLookupValidate(t *testing.T) {
	t.Parallel()

	valid := &DNSLookupRequest{Host: "example.com", Type: "a"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&DNSLookupRequest{Host: "example.com"}).Validate(); err == nil {
		t.Fatal("缺少 type 应当报错")
	}
	if err := (&DNSLookupRequest{Host: "example.com", Type: "BOGUS"}).Validate(); err == nil {
		t.Fatal("不支持的 type 应当报错")
	}
	if err := (&DNSLookupRequest{
		Host: "example.com", Type: "A",
		Servers: []DNSServerInput{{Proto: "gopher", Addr: "1.1.1.1:53"}},
	}).Validate(); err == nil {
		t.Fatal("不支持的 proto 应当报错")
	}
	if err := (&DNSLookupRequest{
		Host: "example.com", Type: "A",
		Servers: []DNSServerInput{{Proto: "doh", Addr: "http://1.1.1.1/dns-query"}},
	}).Validate(); err == nil {
		t.Fatal("DoH 必须用 https")
	}
	if err := (&DNSLookupRequest{
		Host: "example.com", Type: "A",
		Servers: []DNSServerInput{{Proto: "udp", Addr: "1.1.1.1"}},
	}).Validate(); err == nil {
		t.Fatal("缺少端口应当报错")
	}
}

func TestDNSLookupRequiresNetwork(t *testing.T) {
	t.Parallel()

	_, err := DNSLookup(context.Background(), offlineGuard(), &DNSLookupRequest{Host: "example.com", Type: "A"})
	assertCode(t, err, toolerr.CodeNetworkDisabled)
}

func TestDNSTesterValidate(t *testing.T) {
	t.Parallel()

	if err := (&DNSTesterRequest{Action: "whois", Domain: "example.com"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&DNSTesterRequest{Action: "bogus"}).Validate(); err == nil {
		t.Fatal("不支持的 action 应当报错")
	}
	if err := (&DNSTesterRequest{Action: "dnssec"}).Validate(); err == nil {
		t.Fatal("dnssec 缺少 domain 应当报错")
	}
	if err := (&DNSTesterRequest{Action: "ptr"}).Validate(); err == nil {
		t.Fatal("ptr 缺少 ip 应当报错")
	}
	if err := (&DNSTesterRequest{Action: "ptr", IP: "not-an-ip"}).Validate(); err == nil {
		t.Fatal("非法 IP 应当报错")
	}
}

func TestDNSTesterRequiresNetwork(t *testing.T) {
	t.Parallel()

	_, err := DNSTester(context.Background(), offlineGuard(), &DNSTesterRequest{
		Action: "dnssec", Domain: "example.com",
	})
	assertCode(t, err, toolerr.CodeNetworkDisabled)
}

func TestDomainVerifyValidate(t *testing.T) {
	t.Parallel()

	if err := (&DomainVerifyRequest{
		Type: "dns-txt", Domain: "example.com", Expected: "abc",
	}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&DomainVerifyRequest{
		Type: "http-file", Domain: "example.com", Expected: "abc",
	}).Validate(); err == nil {
		t.Fatal("http-file 缺少 filePath 应当报错")
	}
	if err := (&DomainVerifyRequest{
		Type: "dns-txt", Domain: "example.com", Expected: "abc",
		Resolvers: []string{"1.1.1.1"},
	}).Validate(); err == nil {
		t.Fatal("解析器缺少端口应当报错")
	}
	if err := (&DomainVerifyRequest{
		Type: "bogus", Domain: "example.com", Expected: "abc",
	}).Validate(); err == nil {
		t.Fatal("不支持的验证方式应当报错")
	}
}

func TestChainFetchValidate(t *testing.T) {
	t.Parallel()

	if err := (&ChainFetchRequest{Host: "example.com"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&ChainFetchRequest{Host: "example.com", Port: 8443}).Validate(); err != nil {
		t.Fatalf("端口白名单由 netguard 裁决，字段校验不应拦截：%v", err)
	}
	if err := (&ChainFetchRequest{Host: "example.com", Port: 70000}).Validate(); err == nil {
		t.Fatal("超出范围的端口应当报错")
	}
	if err := ValidateTLSProbe(&TLSProbeOptions{Host: "example.com", Versions: []string{"SSLv3"}}); err == nil {
		t.Fatal("不支持的 TLS 版本应当报错")
	}
	if err := (&ChainFetchRequest{Host: "127.0.0.1"}).Validate(); err != nil {
		t.Fatalf("回环地址在参数层是合法的（由 netguard 拦截）：%v", err)
	}
}

func TestChainFetchBlocksLoopback(t *testing.T) {
	t.Parallel()

	guard := netguard.New(netguard.Config{Enabled: true})
	_, err := ChainFetch(context.Background(), guard, &ChainFetchRequest{Host: "127.0.0.1"})
	assertCode(t, err, toolerr.CodeSSRFBlocked)
}

func TestChainFetchBlocksNonWhitelistedPort(t *testing.T) {
	t.Parallel()

	guard := netguard.New(netguard.Config{Enabled: true})
	_, err := ChainFetch(context.Background(), guard, &ChainFetchRequest{Host: "example.com", Port: 8443})
	assertCode(t, err, toolerr.CodeSSRFBlocked)
}

func TestValidateTLSProbe(t *testing.T) {
	t.Parallel()

	if err := ValidateTLSProbe(&TLSProbeOptions{Host: "example.com"}); err != nil {
		t.Fatalf("合法参数不应报错：%v", err)
	}
	if err := ValidateTLSProbe(&TLSProbeOptions{Host: "example.com", Depth: 99}); err == nil {
		t.Fatal("链深度超限应当报错")
	}
	if err := ValidateTLSProbe(&TLSProbeOptions{Host: "example.com", Ciphers: []string{"BOGUS"}}); err == nil {
		t.Fatal("不支持的套件应当报错")
	}
}

func TestSystemRootsLoadable(t *testing.T) {
	t.Parallel()

	// 容器里可能没有信任库；此时返回空切片且不 panic 即为通过。
	roots, closeRoots := systemRoots()
	defer closeRoots()
	t.Logf("本地信任库证书数量：%d", len(roots))
}

func TestParseWhoisText(t *testing.T) {
	t.Parallel()

	body := `Domain Name: EXAMPLE.COM
Registrar: Example Registrar, Inc.
Creation Date: 1995-08-14T04:00:00Z
Registry Expiry Date: 2025-08-13T04:00:00Z
Name Server: A.IANA-SERVERS.NET
Name Server: B.IANA-SERVERS.NET
Domain Status: clientDeleteProhibited https://icann.org/epp#clientDeleteProhibited
`
	info := parseWhoisText(body)
	if info.Registrar != "Example Registrar, Inc." {
		t.Errorf("registrar = %q", info.Registrar)
	}
	if len(info.NameServers) != 2 {
		t.Errorf("nameServers = %v", info.NameServers)
	}
	if info.ExpirationDate == "" {
		t.Error("未解析出到期时间")
	}
	if len(info.Status) == 0 {
		t.Error("未解析出状态")
	}
}

func TestReferServerOf(t *testing.T) {
	t.Parallel()

	if got := referServerOf("refer: whois.verisign-grs.com\n"); got != "whois.verisign-grs.com" {
		t.Errorf("refer = %q", got)
	}
	if got := referServerOf("no referral here"); got != "" {
		t.Errorf("无引导时应返回空，实际 %q", got)
	}
}

func TestVcardValue(t *testing.T) {
	t.Parallel()

	vcard := []any{
		"vcard",
		[]any{
			[]any{"version", map[string]any{}, "text", "4.0"},
			[]any{"fn", map[string]any{}, "text", "Example Registrar"},
		},
	}
	if got := vcardValue(vcard, "fn"); got != "Example Registrar" {
		t.Errorf("fn = %q", got)
	}
	if got := vcardValue(vcard, "adr"); got != "" {
		t.Errorf("缺失属性应返回空，实际 %q", got)
	}
}
