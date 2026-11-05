package network

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// DNSPTRRequest 是 dns-ptr 工具的请求。
type DNSPTRRequest struct {
	// Domain 是域名；可省略（此时用 IP 反查）。
	Domain string `json:"domain"`
	// IP 是地址；可省略（此时用域名正向解析后再反查）。
	IP string `json:"ip"`
	// TimeoutMs 是超时。
	TimeoutMs int `json:"timeoutMs"`
}

// DNSPTRResult 是 dns-ptr 工具的响应。
type DNSPTRResult struct {
	// Query 是实际查询的目标：域名或 IP。
	Query string `json:"query"`
	// PTR 是反向解析结果。
	PTR *PTRInfo `json:"ptr"`
}

// PTRInfo 是反向解析结果。
type PTRInfo struct {
	// Reverse 是查询的反向域名。
	Reverse string `json:"reverse"`
	// Addresses 是正向解析到的地址。
	Addresses []string `json:"addresses"`
	// Names 是反向解析到的名字。
	Names []string `json:"names"`
}

// DNSPTR 做反向解析（PTR）。
//
// 只给域名时先正向解析取第一个地址，再反查；
// 反查出的名字会再做一次正向解析，用于暴露「反向与正向不一致」。
func DNSPTR(ctx context.Context, g *netguard.Guard, req *DNSPTRRequest) (tools.Result, error) {
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

	info, ptrErr := checkPTR(ctx, g, req.Domain, req.IP, timeout)
	if ptrErr != nil {
		return tools.Result{}, mapNetError(ptrErr)
	}

	query := strings.TrimSpace(req.IP)
	if query == "" {
		query = strings.TrimSpace(req.Domain)
	}
	return tools.Result{Data: DNSPTRResult{Query: query, PTR: info}}, nil
}

// Validate 做字段级校验。
func (r *DNSPTRRequest) Validate() error {
	hasDomain := strings.TrimSpace(r.Domain) != ""
	hasIP := strings.TrimSpace(r.IP) != ""

	if !hasDomain && !hasIP {
		return tools.MissingField("ip")
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

// checkPTR 做反向解析。
//
// 只给域名时先正向解析取第一个地址再反查；
// 反查出的名字会再做一次正向解析，用于暴露「反向与正向不一致」。
func checkPTR(ctx context.Context, g *netguard.Guard, domain, ip string, timeout time.Duration) (*PTRInfo, error) {
	servers, err := systemDNSServers()
	if err != nil {
		return nil, err
	}
	server := servers[0]

	var target string
	if trimmed := strings.TrimSpace(ip); trimmed != "" {
		target = trimmed
	} else {
		// 只给了域名：先正向解析，取第一个地址做反向。
		answer := queryOneServer(ctx, g, server, strings.TrimSpace(domain), dns.TypeA, "A", timeout)
		if answer.Error != nil || len(answer.Answers) == 0 {
			return nil, toolerr.New(toolerr.CodeDNSError)
		}
		target = answer.Answers[0].Value
	}

	addr, err := netip.ParseAddr(target)
	if err != nil {
		return nil, tools.InvalidField("ip", "IPv4 或 IPv6 地址", target)
	}
	reverse, err := dns.ReverseAddr(addr.String())
	if err != nil {
		return nil, toolerr.New(toolerr.CodeInvalidDomain)
	}

	info := &PTRInfo{Reverse: reverse, Addresses: []string{}, Names: []string{}}
	ptrAnswer := queryOneServer(ctx, g, server, reverse, dns.TypePTR, "PTR", timeout)
	if ptrAnswer.Error != nil {
		return nil, toolerr.New(toolerr.CodeDNSError)
	}
	for _, item := range ptrAnswer.Answers {
		info.Names = append(info.Names, strings.TrimSuffix(item.Value, "."))
	}

	// 对每个名字再做正向解析，暴露「反向与正向不一致」的情况。
	for _, name := range info.Names {
		qtype := dns.TypeA
		if addr.Is6() {
			qtype = dns.TypeAAAA
		}
		forward := queryOneServer(ctx, g, server, name, qtype, dns.TypeToString[qtype], timeout)
		if forward.Error != nil {
			continue
		}
		for _, item := range forward.Answers {
			if !containsString(info.Addresses, item.Value) {
				info.Addresses = append(info.Addresses, item.Value)
			}
		}
	}
	return info, nil
}
