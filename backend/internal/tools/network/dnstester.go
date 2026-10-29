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

// 动作名。
const (
	// ActionWhois 查询注册信息。
	ActionWhois = "whois"
	// ActionDNSSEC 检查 DNSSEC 部署情况。
	ActionDNSSEC = "dnssec"
	// ActionPTR 做反向解析。
	ActionPTR = "ptr"
)

// maxDNSSECLevels 是向上追溯 DS 的层数上限。
const maxDNSSECLevels = 3

// DNSTesterActions 返回受支持的动作（顺序稳定）。
func DNSTesterActions() []string {
	return []string{ActionWhois, ActionDNSSEC, ActionPTR}
}

// DNSSECInfo 是 DNSSEC 检查结果。
type DNSSECInfo struct {
	// Supported 报告该域名是否可以启用 DNSSEC。
	Supported bool `json:"supported"`
	// DNSSECEnabled 报告当前是否已启用（存在 DS 或 DNSKEY）。
	DNSSECEnabled bool `json:"dnssecEnabled"`
	// DSPresent 报告是否存在 DS 记录。
	DSPresent bool `json:"dsPresent"`
	// SigPresent 报告响应中是否带 RRSIG。
	SigPresent bool `json:"sigPresent"`
	// ChainOfTrust 是逐级 DS 的结果，形如 `example.com: DS 存在`。
	ChainOfTrust []string `json:"chainOfTrust"`
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

// DNSTesterRequest 是 DNS 综合检测工具的请求。
type DNSTesterRequest struct {
	// Domain 是域名；ptr 动作可省略。
	Domain string `json:"domain"`
	// IP 是地址；ptr 与 whois 动作可省略。
	IP string `json:"ip"`
	// Action 是动作。
	Action string `json:"action"`
	// TimeoutMs 是超时。
	TimeoutMs int `json:"timeoutMs"`
}

// DNSTesterResult 是 DNS 综合检测工具的响应。
type DNSTesterResult struct {
	// Action 是执行的动作。
	Action string `json:"action"`
	// Whois 是 whois 结果。
	Whois *WhoisInfo `json:"whois"`
	// DNSSEC 是 dnssec 结果。
	DNSSEC *DNSSECInfo `json:"dnssec"`
	// PTR 是 ptr 结果。
	PTR *PTRInfo `json:"ptr"`
}

// Validate 做字段级校验。
func (r *DNSTesterRequest) Validate() error {
	action := strings.ToLower(strings.TrimSpace(r.Action))
	if action == "" {
		return tools.MissingField("action")
	}
	if !containsString(DNSTesterActions(), action) {
		return tools.InvalidField("action", DNSTesterActions(), action)
	}

	hasDomain := strings.TrimSpace(r.Domain) != ""
	hasIP := strings.TrimSpace(r.IP) != ""

	switch action {
	case ActionPTR:
		if !hasIP && !hasDomain {
			return tools.MissingField("ip")
		}
	case ActionWhois:
		if !hasDomain && !hasIP {
			return tools.MissingField("domain")
		}
	default:
		if !hasDomain {
			return tools.MissingField("domain")
		}
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

// DNSTester 执行 DNS 综合检测。
func DNSTester(ctx context.Context, g *netguard.Guard, req *DNSTesterRequest) (tools.Result, error) {
	if err := requireNetwork(g); err != nil {
		return tools.Result{}, err
	}
	timeout, err := resolveTimeout(req.TimeoutMs, 20000)
	if err != nil {
		return tools.Result{}, err
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	result := DNSTesterResult{Action: action}

	switch action {
	case ActionWhois:
		query := strings.TrimSpace(req.Domain)
		if strings.TrimSpace(req.IP) != "" {
			query = strings.TrimSpace(req.IP)
		}
		info, lookupErr := lookupWhois(ctx, g, query, timeout)
		if lookupErr != nil {
			return tools.Result{}, mapNetError(lookupErr)
		}
		result.Whois = info
		return tools.Result{Data: result}, nil

	case ActionDNSSEC:
		domain, domainErr := normalizeHost(req.Domain)
		if domainErr != nil {
			return tools.Result{}, domainErr
		}
		info, dnssecErr := checkDNSSEC(ctx, g, domain, timeout)
		if dnssecErr != nil {
			return tools.Result{}, dnssecErr
		}
		result.DNSSEC = info
		return tools.Result{Data: result}, nil

	default:
		info, ptrErr := checkPTR(ctx, g, req, timeout)
		if ptrErr != nil {
			return tools.Result{}, ptrErr
		}
		result.PTR = info
		return tools.Result{Data: result}, nil
	}
}

// checkDNSSEC 检查 DNSSEC 部署情况。
func checkDNSSEC(ctx context.Context, g *netguard.Guard, domain string, timeout time.Duration) (*DNSSECInfo, error) {
	servers, err := systemDNSServers()
	if err != nil {
		return nil, err
	}
	server := servers[0]

	dsAnswer := queryOneServer(ctx, g, server, domain, dns.TypeDS, "DS", timeout)
	if dsAnswer.Error != nil {
		return nil, toolerr.New(toolerr.CodeDNSError)
	}
	keyAnswer := queryOneServer(ctx, g, server, domain, dns.TypeDNSKEY, "DNSKEY", timeout)

	info := &DNSSECInfo{
		DSPresent:    len(dsAnswer.Answers) > 0,
		SigPresent:   false,
		ChainOfTrust: []string{},
	}
	if keyAnswer.Error == nil && len(keyAnswer.Answers) > 0 {
		info.SigPresent = true
	}
	info.DNSSECEnabled = info.DSPresent || len(keyAnswer.Answers) > 0
	info.Supported = info.DNSSECEnabled

	// 逐级向父域看 DS：有 DS 才算「链上有锚点」。
	labels := strings.Split(domain, ".")
	for level := 0; level < maxDNSSECLevels; level++ {
		if len(labels) <= 1 {
			break
		}
		current := strings.Join(labels, ".")
		answer := queryOneServer(ctx, g, server, current, dns.TypeDS, "DS", timeout)
		if answer.Error != nil {
			info.ChainOfTrust = append(info.ChainOfTrust, current+": 查询失败")
		} else if len(answer.Answers) > 0 {
			info.ChainOfTrust = append(info.ChainOfTrust, current+": 存在 DS")
		} else {
			info.ChainOfTrust = append(info.ChainOfTrust, current+": 无 DS")
		}
		labels = labels[1:]
	}
	return info, nil
}

// checkPTR 做反向解析。
func checkPTR(ctx context.Context, g *netguard.Guard, req *DNSTesterRequest, timeout time.Duration) (*PTRInfo, error) {
	servers, err := systemDNSServers()
	if err != nil {
		return nil, err
	}
	server := servers[0]

	var target string
	if ip := strings.TrimSpace(req.IP); ip != "" {
		target = ip
	} else {
		// 只给了域名：先正向解析，取第一个地址做反向。
		answer := queryOneServer(ctx, g, server, strings.TrimSpace(req.Domain), dns.TypeA, "A", timeout)
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
