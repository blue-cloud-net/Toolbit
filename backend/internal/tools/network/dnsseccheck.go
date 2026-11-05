package network

import (
	"context"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// maxDNSSECLevels 是向上追溯 DS 的层数上限。
const maxDNSSECLevels = 3

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

// DNSSECCheckRequest 是 dnssec-check 工具的请求。
type DNSSECCheckRequest struct {
	// Domain 是要检查的域名。
	Domain string `json:"domain"`
	// TimeoutMs 是超时。
	TimeoutMs int `json:"timeoutMs"`
}

// DNSSECCheckResult 是 dnssec-check 工具的响应。
type DNSSECCheckResult struct {
	// Domain 是被检查的域名。
	Domain string `json:"domain"`
	// DNSSEC 是检查结果。
	DNSSEC *DNSSECInfo `json:"dnssec"`
}

// DNSSECCheck 检查域名的 DNSSEC 部署情况。
//
// 检查项：是否存在 DS 记录、响应是否带 RRSIG，以及逐级向上追溯的信任链。
func DNSSECCheck(ctx context.Context, g *netguard.Guard, req *DNSSECCheckRequest) (tools.Result, error) {
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

	domain, domainErr := normalizeHost(req.Domain)
	if domainErr != nil {
		return tools.Result{}, domainErr
	}

	info, dnssecErr := checkDNSSEC(ctx, g, domain, timeout)
	if dnssecErr != nil {
		return tools.Result{}, mapNetError(dnssecErr)
	}
	return tools.Result{Data: DNSSECCheckResult{Domain: domain, DNSSEC: info}}, nil
}

// Validate 做字段级校验。
func (r *DNSSECCheckRequest) Validate() error {
	if strings.TrimSpace(r.Domain) == "" {
		return tools.MissingField("domain")
	}
	if _, err := normalizeHost(r.Domain); err != nil {
		return err
	}
	return nil
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
