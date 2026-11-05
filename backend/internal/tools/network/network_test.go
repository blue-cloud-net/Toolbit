package network

import (
	"context"
	"errors"
	"testing"

	"github.com/miekg/dns"

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
	if err := (&CIDRCalcRequest{CIDR: "10.0.0.0/24", Limit: maxExpandAddresses + 1}).Validate(); err == nil {
		t.Fatal("超出上限的 limit 应当报错")
	}
	if err := (&CIDRCalcRequest{CIDR: "10.0.0.0/24", Limit: -1}).Validate(); err == nil {
		t.Fatal("负数的 limit 应当报错")
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

func TestBuildDNSQueryAdvertisesEDNS0(t *testing.T) {
	t.Parallel()

	msg := buildDNSQuery("example.com", dns.TypeA)
	if len(msg.Question) != 1 {
		t.Fatalf("问题区条目数 = %d，期望 1", len(msg.Question))
	}
	if msg.Question[0].Name != "example.com." {
		t.Fatalf("问题名 = %q，期望 example.com.", msg.Question[0].Name)
	}
	if msg.Question[0].Qtype != dns.TypeA {
		t.Fatalf("查询类型 = %d，期望 %d", msg.Question[0].Qtype, dns.TypeA)
	}
	if !msg.RecursionDesired {
		t.Fatal("应当请求递归解析")
	}
	// 不通告 EDNS0 时上游按 512 字节截断应答，
	// 记录会被切掉 → 跨服务器比对出现「假差异」。
	opt := msg.IsEdns0()
	if opt == nil {
		t.Fatal("必须通告 EDNS0 缓冲")
	}
	if opt.UDPSize() != dnsUDPBufferSize {
		t.Fatalf("EDNS0 缓冲 = %d，期望 %d", opt.UDPSize(), dnsUDPBufferSize)
	}
}

func TestDNSWhoisValidate(t *testing.T) {
	t.Parallel()

	if err := (&DNSWhoisRequest{Domain: "example.com"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&DNSWhoisRequest{}).Validate(); err == nil {
		t.Fatal("缺少 domain 与 ip 应当报错")
	}
	if err := (&DNSWhoisRequest{IP: "not-an-ip"}).Validate(); err == nil {
		t.Fatal("非法 IP 应当报错")
	}
}

func TestDNSSECCheckValidate(t *testing.T) {
	t.Parallel()

	if err := (&DNSSECCheckRequest{Domain: "example.com"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&DNSSECCheckRequest{}).Validate(); err == nil {
		t.Fatal("缺少 domain 应当报错")
	}
	if err := (&DNSSECCheckRequest{Domain: "bad domain"}).Validate(); err == nil {
		t.Fatal("非法域名应当报错")
	}
}

func TestDNSPTRValidate(t *testing.T) {
	t.Parallel()

	if err := (&DNSPTRRequest{IP: "1.1.1.1"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&DNSPTRRequest{Domain: "example.com"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&DNSPTRRequest{}).Validate(); err == nil {
		t.Fatal("缺少 ip 与 domain 应当报错")
	}
	if err := (&DNSPTRRequest{IP: "not-an-ip"}).Validate(); err == nil {
		t.Fatal("非法 IP 应当报错")
	}
}

func TestDNSToolsRequireNetwork(t *testing.T) {
	t.Parallel()

	_, err := DNSSECCheck(context.Background(), offlineGuard(), &DNSSECCheckRequest{
		Domain: "example.com",
	})
	assertCode(t, err, toolerr.CodeNetworkDisabled)

	_, err = DNSWhois(context.Background(), offlineGuard(), &DNSWhoisRequest{
		Domain: "example.com",
	})
	assertCode(t, err, toolerr.CodeNetworkDisabled)

	_, err = DNSPTR(context.Background(), offlineGuard(), &DNSPTRRequest{
		IP: "1.1.1.1",
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

// TestCIDRCalcAddressFamilyConsistency 断言首末地址与网段地址同属一个地址族。
//
// 4-in-6 前缀（::ffff:a.b.c.d/120）曾让 network 走 16 字节形式、
// last 走 4 字节形式，首末地址族不一致。
func TestCIDRCalcAddressFamilyConsistency(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{CIDR: "::ffff:192.168.1.0/120"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	if out.FirstAddress != "::ffff:192.168.1.0" {
		t.Errorf("firstAddress = %q，期望 ::ffff:192.168.1.0", out.FirstAddress)
	}
	if out.LastAddress != "::ffff:192.168.1.255" {
		t.Errorf("lastAddress = %q，期望 ::ffff:192.168.1.255", out.LastAddress)
	}
}

// TestCIDRCalcExpand 断言网段被逐个展开成地址列表。
func TestCIDRCalcExpand(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{
		CIDR:   "192.168.1.0/30",
		Expand: true,
	})
	if err != nil {
		t.Fatalf("展开失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	want := []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"}
	if len(out.Addresses) != len(want) {
		t.Fatalf("addresses = %v，期望 %v", out.Addresses, want)
	}
	for i := range want {
		if out.Addresses[i] != want[i] {
			t.Fatalf("addresses = %v，期望 %v", out.Addresses, want)
		}
	}
	if out.AddressCount != "4" {
		t.Errorf("addressCount = %s，期望 4", out.AddressCount)
	}
	if out.Truncated {
		t.Error("未超限时不应标记截断")
	}
}

// TestCIDRCalcExpandRange 断言范围模式按起止地址展开。
func TestCIDRCalcExpandRange(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{
		Start:  "10.0.0.2",
		End:    "10.0.0.5",
		Expand: true,
	})
	if err != nil {
		t.Fatalf("展开失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	want := []string{"10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5"}
	if len(out.Addresses) != len(want) {
		t.Fatalf("addresses = %v，期望 %v", out.Addresses, want)
	}
	for i := range want {
		if out.Addresses[i] != want[i] {
			t.Fatalf("addresses = %v，期望 %v", out.Addresses, want)
		}
	}
	if out.AddressCount != "4" {
		t.Errorf("addressCount = %s，期望 4", out.AddressCount)
	}
}

// TestCIDRCalcExpandTruncated 断言超出上限时截断并标记。
func TestCIDRCalcExpandTruncated(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{
		CIDR:   "10.0.0.0/24",
		Expand: true,
		Limit:  4,
	})
	if err != nil {
		t.Fatalf("展开失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	if len(out.Addresses) != 4 {
		t.Fatalf("addresses 数量 = %d，期望 4", len(out.Addresses))
	}
	if !out.Truncated {
		t.Error("超出上限时应当标记截断")
	}
	// 总数仍按实际区间给出，不受上限影响。
	if out.AddressCount != "256" {
		t.Errorf("addressCount = %s，期望 256", out.AddressCount)
	}
}

// TestCIDRCalcExpandIPv6 断言 IPv6 网段也能展开，且首末地址成对。
func TestCIDRCalcExpandIPv6(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{
		CIDR:   "2001:db8::/126",
		Expand: true,
	})
	if err != nil {
		t.Fatalf("展开失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	want := []string{"2001:db8::", "2001:db8::1", "2001:db8::2", "2001:db8::3"}
	if len(out.Addresses) != len(want) {
		t.Fatalf("addresses = %v，期望 %v", out.Addresses, want)
	}
	for i := range want {
		if out.Addresses[i] != want[i] {
			t.Fatalf("addresses = %v，期望 %v", out.Addresses, want)
		}
	}
}

// TestCIDRCalcExpandAddressFamilyConsistency 断言首末地址与展开列表同属一个地址族。
//
// 4-in-6 前缀（::ffff:a.b.c.d/120）曾让 network 走 16 字节形式、
// last 走 4 字节形式，首末地址族不一致。
func TestCIDRCalcExpandAddressFamilyConsistency(t *testing.T) {
	t.Parallel()

	result, err := CIDRCalc(context.Background(), &CIDRCalcRequest{
		CIDR:   "::ffff:192.168.1.0/120",
		Expand: true,
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	out := result.Data.(CIDRCalcResult)
	if out.FirstAddress != out.Addresses[0] {
		t.Errorf("firstAddress = %q，与展开首项 %q 不一致", out.FirstAddress, out.Addresses[0])
	}
	if out.LastAddress != out.Addresses[len(out.Addresses)-1] {
		t.Errorf("lastAddress = %q，与展开末项 %q 不一致", out.LastAddress, out.Addresses[len(out.Addresses)-1])
	}
	if out.AddressCount != "256" {
		t.Errorf("addressCount = %s，期望 256", out.AddressCount)
	}
}
