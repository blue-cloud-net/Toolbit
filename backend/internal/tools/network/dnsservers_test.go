package network

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// dnsServers 执行一次目录查询并取出结果。
func dnsServers(t *testing.T, req *DNSServersRequest) DNSServersResult {
	t.Helper()
	result, err := DNSServers(context.Background(), req)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	out, ok := result.Data.(DNSServersResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

// dnsServersExpectError 执行一次预期失败的目录查询。
func dnsServersExpectError(t *testing.T, req *DNSServersRequest, code toolerr.Code) {
	t.Helper()
	_, err := DNSServers(context.Background(), req)
	if err == nil {
		t.Fatal("期望出错，但没有")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型不符：%T", err)
	}
	if toolErr.Code != code {
		t.Errorf("错误码 = %s，期望 %s", toolErr.Code, code)
	}
}

func TestDNSServersDataset(t *testing.T) {
	t.Parallel()

	out := dnsServers(t, &DNSServersRequest{})
	if out.Total == 0 {
		t.Fatal("目录不应为空")
	}
	if out.Providers < 20 {
		t.Errorf("服务商数 = %d，期望至少 20 家", out.Providers)
	}
	if out.DataVersion != DNSServersVersion {
		t.Errorf("dataVersion = %s，期望 %s", out.DataVersion, DNSServersVersion)
	}

	// 四种协议必须齐全 —— 这是本工具存在的意义。
	protos := map[string]int{}
	families := map[string]int{}
	regions := map[string]int{}
	for _, entry := range out.Servers {
		protos[entry.Proto]++
		families[entry.Family]++
		regions[entry.Region]++
	}
	for _, proto := range DNSProtos() {
		if protos[proto] == 0 {
			t.Errorf("协议 %s 没有任何端点", proto)
		}
	}
	for _, family := range DNSServerFamilies() {
		if families[family] == 0 {
			t.Errorf("地址族 %s 没有任何端点", family)
		}
	}
	for _, region := range DNSServerRegions() {
		if regions[region] == 0 {
			t.Errorf("地区 %s 没有任何端点", region)
		}
	}
}

// TestDNSServersEndpointsAreUsable 断言目录里的端点都能直接作为
// `dns-lookup` 的自定义服务器使用 —— 这条不变量让「实测」跳转成立。
func TestDNSServersEndpointsAreUsable(t *testing.T) {
	t.Parallel()

	out := dnsServers(t, &DNSServersRequest{})
	for _, entry := range out.Servers {
		if err := validateDNSServer(DNSServerInput{Proto: entry.Proto, Addr: entry.Addr}); err != nil {
			t.Errorf("%s 的端点 %s（%s）不可用：%v", entry.Provider, entry.Addr, entry.Proto, err)
		}
	}
}

// TestDNSServersProtocolPairs 断言每个提供明文查询的地址都同时给出 udp 与 tcp。
//
// 只看 `udp` / `tcp` 端点：DoH 的 IP 字面量 URL 同样记在 `ipv4` 族里，
// 但它是加密通道，**不**参与明文协议配对。
func TestDNSServersProtocolPairs(t *testing.T) {
	t.Parallel()

	out := dnsServers(t, &DNSServersRequest{})
	pairs := make(map[string]map[string]bool)
	for _, entry := range out.Servers {
		if entry.Proto != DNSProtoUDP && entry.Proto != DNSProtoTCP {
			continue
		}
		if pairs[entry.Addr] == nil {
			pairs[entry.Addr] = map[string]bool{}
		}
		pairs[entry.Addr][entry.Proto] = true
	}
	for addr, protos := range pairs {
		if !protos[DNSProtoUDP] || !protos[DNSProtoTCP] {
			t.Errorf("%s 缺少 udp 或 tcp 端点", addr)
		}
	}
}

func TestDNSServersFilters(t *testing.T) {
	t.Parallel()

	byRegion := dnsServers(t, &DNSServersRequest{Region: DNSServerRegionChina})
	if byRegion.Total == 0 {
		t.Fatal("中国大陆应有端点")
	}
	for _, entry := range byRegion.Servers {
		if entry.Region != DNSServerRegionChina {
			t.Fatalf("地区过滤泄漏：%s", entry.Region)
		}
	}

	byProto := dnsServers(t, &DNSServersRequest{Proto: DNSProtoDoH})
	if byProto.Total == 0 {
		t.Fatal("DoH 应有端点")
	}
	for _, entry := range byProto.Servers {
		if entry.Proto != DNSProtoDoH {
			t.Fatalf("协议过滤泄漏：%s", entry.Proto)
		}
	}

	byFamily := dnsServers(t, &DNSServersRequest{Family: DNSServerFamilyIPv6})
	if byFamily.Total == 0 {
		t.Fatal("IPv6 应有端点")
	}
	for _, entry := range byFamily.Servers {
		if entry.Family != DNSServerFamilyIPv6 {
			t.Fatalf("地址族过滤泄漏：%s", entry.Family)
		}
	}
}

func TestDNSServersQuery(t *testing.T) {
	t.Parallel()

	// 中文服务商名。
	out := dnsServers(t, &DNSServersRequest{Query: "阿里"})
	if out.Total == 0 {
		t.Fatal("「阿里」应命中 AliDNS")
	}

	// 英文服务商名。
	out = dnsServers(t, &DNSServersRequest{Query: "cloudflare"})
	if out.Total == 0 {
		t.Fatal("「cloudflare」应命中")
	}

	// 地址片段。
	out = dnsServers(t, &DNSServersRequest{Query: "1.1.1.1"})
	if out.Total == 0 {
		t.Fatal("「1.1.1.1」应命中")
	}

	// 未命中时返回空列表而不是错误。
	out = dnsServers(t, &DNSServersRequest{Query: "zzzz-not-a-dns"})
	if out.Total != 0 || len(out.Servers) != 0 {
		t.Errorf("未命中时 total = %d，期望 0", out.Total)
	}
}

// TestDNSServersStableOrder 断言排序稳定：同一服务商内按地址族与协议排序。
func TestDNSServersStableOrder(t *testing.T) {
	t.Parallel()

	out := dnsServers(t, &DNSServersRequest{})
	lastProvider := ""
	for _, entry := range out.Servers {
		if entry.Provider != lastProvider {
			lastProvider = entry.Provider
			continue
		}
	}
	// 断言目录里每个服务商只出现一次连续区块。
	seen := make(map[string]bool)
	previous := ""
	for _, entry := range out.Servers {
		if entry.Provider == previous {
			continue
		}
		if seen[entry.Provider] {
			t.Fatalf("服务商 %s 的端点不连续", entry.Provider)
		}
		seen[entry.Provider] = true
		previous = entry.Provider
	}
}

func TestDNSServersValidation(t *testing.T) {
	t.Parallel()

	dnsServersExpectError(t, &DNSServersRequest{Region: "mars"}, toolerr.CodeInvalidField)
	dnsServersExpectError(t, &DNSServersRequest{Proto: "gopher"}, toolerr.CodeInvalidField)
	dnsServersExpectError(t, &DNSServersRequest{Family: "ipv7"}, toolerr.CodeInvalidField)
	dnsServersExpectError(t, &DNSServersRequest{Query: strings.Repeat("x", 300)}, toolerr.CodeInvalidField)
}

// TestDNSServersNoLeakOfInput 断言错误信息不回显用户输入。
func TestDNSServersNoLeakOfInput(t *testing.T) {
	t.Parallel()

	secret := "secret-region-value"
	_, err := DNSServers(context.Background(), &DNSServersRequest{Region: secret})
	if err == nil {
		t.Fatal("期望出错，但没有")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("错误信息回显了用户输入：%s", err.Error())
	}
}
