package network

import (
	"context"
	"net/netip"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// DNSWhoisRequest 是 dns-whois 工具的请求。
type DNSWhoisRequest struct {
	// Domain 是域名；可省略（此时用 IP 查询）。
	Domain string `json:"domain"`
	// IP 是地址；可省略（此时用域名查询）。
	IP string `json:"ip"`
	// TimeoutMs 是超时。
	TimeoutMs int `json:"timeoutMs"`
}

// DNSWhoisResult 是 dns-whois 工具的响应。
type DNSWhoisResult struct {
	// Query 是实际查询的目标：域名或 IP。
	Query string `json:"query"`
	// Whois 是注册信息。
	Whois *WhoisInfo `json:"whois"`
}

// DNSWhois 查询域名或 IP 的注册信息。
//
// 优先走 RDAP（HTTPS，结构化）；失败时回退到 whois 协议（TCP 43）。
// 两条路径都经 netguard，端口固定不可由用户指定。
func DNSWhois(ctx context.Context, g *netguard.Guard, req *DNSWhoisRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}
	if err := requireNetwork(g); err != nil {
		return tools.Result{}, err
	}
	timeout, err := resolveTimeout(req.TimeoutMs, 20000)
	if err != nil {
		return tools.Result{}, err
	}

	query := strings.TrimSpace(req.Domain)
	if ip := strings.TrimSpace(req.IP); ip != "" {
		query = ip
	}

	info, lookupErr := lookupWhois(ctx, g, query, timeout)
	if lookupErr != nil {
		return tools.Result{}, mapNetError(lookupErr)
	}
	return tools.Result{Data: DNSWhoisResult{Query: query, Whois: info}}, nil
}

// Validate 做字段级校验。
func (r *DNSWhoisRequest) Validate() error {
	hasDomain := strings.TrimSpace(r.Domain) != ""
	hasIP := strings.TrimSpace(r.IP) != ""

	if !hasDomain && !hasIP {
		return tools.MissingField("domain")
	}
	if hasDomain {
		if _, err := normalizeHost(r.Domain); err != nil {
			return err
		}
	}
	if hasIP {
		if _, err := netip.ParseAddr(strings.TrimSpace(r.IP)); err != nil {
			return tools.InvalidField("ip", "IPv4 或 IPv6 地址", r.IP)
		}
	}
	return nil
}
