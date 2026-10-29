package network

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 域名验证方式。
const (
	// DomainVerifyDNS 通过 TXT 记录验证。
	DomainVerifyDNS = "dns-txt"
	// DomainVerifyCNAME 通过 CNAME 记录验证。
	DomainVerifyCNAME = "dns-cname"
	// DomainVerifyHTTP 通过站点上的文件验证。
	DomainVerifyHTTP = "http-file"
)

// maxBodyPreview 是响应体预览的截断长度。
const maxBodyPreview = 512

// DomainVerifyMethods 返回受支持的验证方式（顺序稳定）。
func DomainVerifyMethods() []string {
	return []string{DomainVerifyDNS, DomainVerifyCNAME, DomainVerifyHTTP}
}

// DomainVerifyRequest 是域名验证工具的请求。
type DomainVerifyRequest struct {
	// Type 是验证方式。
	Type string `json:"type"`
	// Domain 是域名。
	Domain string `json:"domain"`
	// Expected 是期望值。
	Expected string `json:"expected"`
	// FilePath 是 http-file 方式下的相对路径。
	FilePath string `json:"filePath"`
	// Resolvers 是自定义 DNS 服务器（host:port）；为空时用系统配置。
	Resolvers []string `json:"resolvers"`
	// TimeoutMs 是超时。
	TimeoutMs int `json:"timeoutMs"`
}

// ResolverOutcome 是单个解析器的结论。
type ResolverOutcome struct {
	// Resolver 是解析器标识。
	Resolver string `json:"resolver"`
	// Values 是查到的记录值。
	Values []string `json:"values"`
	// Matched 报告该解析器下是否匹配。
	Matched bool `json:"matched"`
	// Error 是查询错误；成功时为 null。
	Error *DNSServerError `json:"error"`
}

// DomainVerifyResult 是域名验证工具的响应。
type DomainVerifyResult struct {
	// Type 是验证方式。
	Type string `json:"type"`
	// Domain 是域名。
	Domain string `json:"domain"`
	// Expected 是期望值。
	Expected string `json:"expected"`
	// Matched 是总体结论。
	Matched bool `json:"matched"`
	// PerResolver 是逐解析器的结论（DNS 方式）。
	PerResolver []ResolverOutcome `json:"perResolver"`
	// HTTPStatus 是 http-file 方式的状态码。
	HTTPStatus int `json:"httpStatus"`
	// BodyPreview 是 http-file 方式的响应体预览。
	BodyPreview string `json:"bodyPreview"`
}

// Validate 做字段级校验。
func (r *DomainVerifyRequest) Validate() error {
	typ := strings.ToLower(strings.TrimSpace(r.Type))
	if typ == "" {
		return tools.MissingField("type")
	}
	if !containsString(DomainVerifyMethods(), typ) {
		return tools.InvalidField("type", DomainVerifyMethods(), typ)
	}
	if _, err := normalizeHost(r.Domain); err != nil {
		return err
	}
	if strings.TrimSpace(r.Expected) == "" {
		return tools.MissingField("expected")
	}
	if typ == DomainVerifyHTTP && strings.TrimSpace(r.FilePath) == "" {
		return tools.MissingField("filePath")
	}
	if len(r.Resolvers) > maxServers {
		return toolerrTooMany("resolvers", len(r.Resolvers), maxServers)
	}
	for _, resolver := range r.Resolvers {
		if _, _, err := net.SplitHostPort(strings.TrimSpace(resolver)); err != nil {
			return tools.InvalidField("resolvers", "host:port", resolver)
		}
	}
	return nil
}

// DomainVerify 执行域名所有权验证。
func DomainVerify(ctx context.Context, g *netguard.Guard, req *DomainVerifyRequest) (tools.Result, error) {
	if err := requireNetwork(g); err != nil {
		return tools.Result{}, err
	}
	domain, err := normalizeHost(req.Domain)
	if err != nil {
		return tools.Result{}, err
	}
	timeout, err := resolveTimeout(req.TimeoutMs, defaultTimeoutMs)
	if err != nil {
		return tools.Result{}, err
	}

	typ := strings.ToLower(strings.TrimSpace(req.Type))
	result := DomainVerifyResult{
		Type:        typ,
		Domain:      domain,
		Expected:    strings.TrimSpace(req.Expected),
		PerResolver: []ResolverOutcome{},
	}

	if typ == DomainVerifyHTTP {
		return verifyHTTPFile(ctx, g, &result, req, timeout)
	}
	return verifyDNSToken(ctx, g, &result, req, typ, timeout)
}

// verifyDNSToken 用 DNS 记录比对期望值。
func verifyDNSToken(
	ctx context.Context, g *netguard.Guard, result *DomainVerifyResult,
	req *DomainVerifyRequest, typ string, timeout time.Duration,
) (tools.Result, error) {
	servers, err := dnsServersOf(req.Resolvers)
	if err != nil {
		return tools.Result{}, err
	}
	qtype := dnsRecordTypes["TXT"]
	if typ == DomainVerifyCNAME {
		qtype = dnsRecordTypes["CNAME"]
	}
	expect := strings.ToLower(strings.TrimSuffix(result.Expected, "."))

	for _, srv := range servers {
		answer := queryOneServer(ctx, g, srv, result.Domain, qtype, dns.TypeToString[qtype], timeout)
		outcome := ResolverOutcome{
			Resolver: srv.Addr,
			Values:   []string{},
			Error:    answer.Error,
		}
		for _, item := range answer.Answers {
			outcome.Values = append(outcome.Values, item.Value)
			if strings.ToLower(strings.TrimSuffix(item.Value, ".")) == expect {
				outcome.Matched = true
				result.Matched = true
			}
		}
		result.PerResolver = append(result.PerResolver, outcome)
	}
	return tools.Result{Data: result}, nil
}

// verifyHTTPFile 拉取站点文件比对内容。
func verifyHTTPFile(
	ctx context.Context, g *netguard.Guard, result *DomainVerifyResult,
	req *DomainVerifyRequest, timeout time.Duration,
) (tools.Result, error) {
	path := strings.TrimSpace(req.FilePath)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	rawURL := "https://" + result.Domain + path

	client := g.HTTPClient(netguard.HTTPOptions{Timeout: timeout})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return tools.Result{}, toolerr.New(toolerr.CodeInvalidDomain)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return tools.Result{}, mapNetError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodyPreview*4))
	if readErr != nil {
		return tools.Result{}, mapNetError(readErr)
	}

	result.HTTPStatus = resp.StatusCode
	result.BodyPreview = truncate(string(body), maxBodyPreview)
	if resp.StatusCode == http.StatusOK {
		result.Matched = strings.TrimSpace(string(body)) == result.Expected
	}
	return tools.Result{Data: result}, nil
}

// dnsServersOf 把解析器列表转成查询目标。
func dnsServersOf(resolvers []string) ([]DNSServerInput, error) {
	if len(resolvers) == 0 {
		return systemDNSServers()
	}
	out := make([]DNSServerInput, 0, len(resolvers))
	for _, resolver := range resolvers {
		out = append(out, DNSServerInput{Proto: DNSProtoUDP, Addr: strings.TrimSpace(resolver)})
	}
	return out, nil
}

// truncate 按字符数截断字符串。
func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}

// itoa 是整数转字符串的简写。
func itoa(v int) string {
	return strconv.Itoa(v)
}
