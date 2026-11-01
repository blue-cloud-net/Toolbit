package network

import (
	"context"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// DNS 服务器目录的枚举取值。
const (
	// DNSServerRegionChina 是中国大陆的公共解析器。
	DNSServerRegionChina = "china"
	// DNSServerRegionAsia 是亚太地区的公共解析器。
	DNSServerRegionAsia = "asia"
	// DNSServerRegionGlobal 是全球范围的公共解析器。
	DNSServerRegionGlobal = "global"
)

// DNS 服务器地址族取值。
const (
	// DNSServerFamilyIPv4 是 IPv4 字面量地址。
	DNSServerFamilyIPv4 = "ipv4"
	// DNSServerFamilyIPv6 是 IPv6 字面量地址。
	DNSServerFamilyIPv6 = "ipv6"
	// DNSServerFamilyDomain 是域名形式（DoT / DoH 端点）。
	DNSServerFamilyDomain = "domain"
)

// DNSServerRegions 返回受支持的地区（顺序稳定）。
func DNSServerRegions() []string {
	return []string{DNSServerRegionChina, DNSServerRegionAsia, DNSServerRegionGlobal}
}

// DNSServerFamilies 返回受支持的地址族（顺序稳定）。
func DNSServerFamilies() []string {
	return []string{DNSServerFamilyIPv4, DNSServerFamilyIPv6, DNSServerFamilyDomain}
}

// DNSServerEntry 是 DNS 服务器目录里的一条端点记录。
//
// **一行 = 一个可复制的端点**：同一个 IP 会有 `udp` 与 `tcp` 两条，
// 因为二者是独立的查询通道，必须分别列出。
type DNSServerEntry struct {
	// Provider 是服务商中文名。
	Provider string `json:"provider"`
	// ProviderEn 是服务商英文名。
	ProviderEn string `json:"providerEn"`
	// Region 是归属地区：china / asia / global。
	Region string `json:"region"`
	// Family 是地址族：ipv4 / ipv6 / domain。
	Family string `json:"family"`
	// Proto 是查询协议：udp / tcp / dot / doh。
	Proto string `json:"proto"`
	// Addr 是服务器地址；DoH 时是完整 URL。
	Addr string `json:"addr"`
	// Note 是端点说明。
	Note string `json:"note"`
}

// DNSServersRequest 是 DNS 服务器目录工具的请求。
//
// 四个字段**都可以省略**：省略时返回全量目录，
// 便于前端一次性取回后在本地做搜索与筛选。
type DNSServersRequest struct {
	// Query 是搜索词，匹配服务商名与地址；为空表示不过滤。
	Query string `json:"query"`
	// Region 是地区过滤；为空表示不过滤。
	Region string `json:"region"`
	// Proto 是协议过滤；为空表示不过滤。
	Proto string `json:"proto"`
	// Family 是地址族过滤；为空表示不过滤。
	Family string `json:"family"`
}

// DNSServersResult 是 DNS 服务器目录工具的响应。
type DNSServersResult struct {
	// Servers 是命中的端点列表。
	Servers []DNSServerEntry `json:"servers"`
	// Total 是命中的端点数。
	Total int `json:"total"`
	// Providers 是命中的服务商数。
	Providers int `json:"providers"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// Validate 做字段级校验。
func (r *DNSServersRequest) Validate() error {
	if len(r.Query) > maxHostLength {
		return tools.InvalidField("query", fmtMaxLen(maxHostLength), len(r.Query))
	}
	region := strings.ToLower(strings.TrimSpace(r.Region))
	if region != "" && !containsString(DNSServerRegions(), region) {
		return tools.InvalidField("region", DNSServerRegions(), len(region))
	}
	proto := strings.ToLower(strings.TrimSpace(r.Proto))
	if proto != "" && !containsString(DNSProtos(), proto) {
		return tools.InvalidField("proto", DNSProtos(), len(proto))
	}
	family := strings.ToLower(strings.TrimSpace(r.Family))
	if family != "" && !containsString(DNSServerFamilies(), family) {
		return tools.InvalidField("family", DNSServerFamilies(), len(family))
	}
	return nil
}

// DNS 服务器目录数据集缓存。
var (
	dnsServersOnce    sync.Once
	dnsServersEntries []DNSServerEntry
	// dnsServersProviderOrder 记录服务商的首次出现顺序，用于稳定排序。
	dnsServersProviderOrder map[string]int
)

// loadDNSServers 解析并缓存 DNS 服务器目录。
//
// 解析只做一次：CSV 是内嵌的静态数据，重复解析纯属浪费。
func loadDNSServers() {
	dnsServersOnce.Do(func() {
		entries := parseDNSServers(readDataset("dataset/dns_servers.csv"))
		dnsServersEntries = entries
		dnsServersProviderOrder = make(map[string]int, len(entries))
		for _, entry := range entries {
			if _, ok := dnsServersProviderOrder[entry.Provider]; !ok {
				dnsServersProviderOrder[entry.Provider] = len(dnsServersProviderOrder)
			}
		}
		sortDNSServers(entries)
	})
}

// sortDNSServers 按「服务商 → 地址族 → 协议 → 地址」稳定排序。
//
// 服务商之间按 CSV 里的首次出现顺序（目录是**有序**的：大陆 → 亚太 → 全球）；
// 地址族按 ipv4 → ipv6 → 域名，协议按 udp → tcp → dot → doh。
func sortDNSServers(entries []DNSServerEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Provider != b.Provider {
			return dnsServersProviderOrder[a.Provider] < dnsServersProviderOrder[b.Provider]
		}
		if a.Family != b.Family {
			return dnsServerFamilyRank(a.Family) < dnsServerFamilyRank(b.Family)
		}
		if a.Proto != b.Proto {
			return dnsServerProtoRank(a.Proto) < dnsServerProtoRank(b.Proto)
		}
		return a.Addr < b.Addr
	})
}

// dnsServerFamilyRank 返回地址族的排序权重。
func dnsServerFamilyRank(family string) int {
	switch family {
	case DNSServerFamilyIPv4:
		return 0
	case DNSServerFamilyIPv6:
		return 1
	default:
		return 2
	}
}

// dnsServerProtoRank 返回协议的排序权重。
func dnsServerProtoRank(proto string) int {
	switch proto {
	case DNSProtoUDP:
		return 0
	case DNSProtoTCP:
		return 1
	case DNSProtoDoT:
		return 2
	default:
		return 3
	}
}

// parseDNSServers 解析 DNS 服务器目录 CSV。
func parseDNSServers(r io.Reader) []DNSServerEntry {
	out := []DNSServerEntry{}
	for _, rec := range readCSV(r) {
		provider := strings.TrimSpace(csvField(rec, 0))
		proto := strings.ToLower(strings.TrimSpace(csvField(rec, 4)))
		addr := strings.TrimSpace(csvField(rec, 5))
		if provider == "" || addr == "" {
			continue
		}
		entry := DNSServerEntry{
			Provider:   provider,
			ProviderEn: strings.TrimSpace(csvField(rec, 1)),
			Region:     strings.ToLower(strings.TrimSpace(csvField(rec, 2))),
			Family:     strings.ToLower(strings.TrimSpace(csvField(rec, 3))),
			Proto:      proto,
			Addr:       addr,
			Note:       strings.TrimSpace(csvField(rec, 6)),
		}
		// 数据集里出现协议或地区之外的取值时**跳过**该行：
		// 目录是只读的快照，脏数据不应让整个工具失效。
		if !containsString(DNSProtos(), entry.Proto) {
			continue
		}
		if !containsString(DNSServerRegions(), entry.Region) {
			continue
		}
		if !containsString(DNSServerFamilies(), entry.Family) {
			continue
		}
		out = append(out, entry)
	}
	return out
}

// DNSServers 查询 DNS 服务器目录。
//
// 目录是**离线内置**的快照，**不**出网 —— 本工具只做检索，
// 真正的解析由 `dns-lookup` 负责。
//
// 三种用法：
//   - 空请求：返回全量目录（列表浏览）；
//   - 搜索词：在服务商名（中英文）与地址里做子串匹配；
//   - 枚举过滤：`region` / `proto` / `family` 可与任一方式叠加。
func DNSServers(_ context.Context, req *DNSServersRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadDNSServers()

	query := strings.ToLower(strings.TrimSpace(req.Query))
	region := strings.ToLower(strings.TrimSpace(req.Region))
	proto := strings.ToLower(strings.TrimSpace(req.Proto))
	family := strings.ToLower(strings.TrimSpace(req.Family))

	matched := make([]DNSServerEntry, 0, len(dnsServersEntries))
	for _, entry := range dnsServersEntries {
		if region != "" && entry.Region != region {
			continue
		}
		if proto != "" && entry.Proto != proto {
			continue
		}
		if family != "" && entry.Family != family {
			continue
		}
		if query != "" && !matchDNSServer(entry, query) {
			continue
		}
		matched = append(matched, entry)
	}

	return tools.Result{Data: DNSServersResult{
		Servers:     matched,
		Total:       len(matched),
		Providers:   countDNSServerProviders(matched),
		DataVersion: DNSServersVersion,
	}}, nil
}

// matchDNSServer 报告端点是否命中搜索词。
//
// 命中范围：服务商中文名、英文名、地址、协议与说明。
func matchDNSServer(entry DNSServerEntry, query string) bool {
	haystack := []string{
		entry.Provider,
		entry.ProviderEn,
		entry.Addr,
		entry.Note,
		entry.Proto,
	}
	for _, value := range haystack {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

// countDNSServerProviders 统计命中的服务商数（去重）。
func countDNSServerProviders(entries []DNSServerEntry) int {
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		seen[entry.Provider] = struct{}{}
	}
	return len(seen)
}

// fmtMaxLen 构造长度上限的期望值描述。
func fmtMaxLen(n int) string {
	return "最多 " + strconv.Itoa(n) + " 个字符"
}
