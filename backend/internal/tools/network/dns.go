package network

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// DNS 查询协议取值。
const (
	// DNSProtoUDP 是明文 UDP。
	DNSProtoUDP = "udp"
	// DNSProtoTCP 是明文 TCP。
	DNSProtoTCP = "tcp"
	// DNSProtoDoT 是 DNS over TLS。
	DNSProtoDoT = "dot"
	// DNSProtoDoH 是 DNS over HTTPS。
	DNSProtoDoH = "doh"
)

// DNSProtos 返回受支持的查询协议（顺序稳定）。
func DNSProtos() []string {
	return []string{DNSProtoUDP, DNSProtoTCP, DNSProtoDoT, DNSProtoDoH}
}

// DNSRecordTypes 返回受支持的记录类型（顺序稳定）。
func DNSRecordTypes() []string {
	return []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "SRV", "CAA", "PTR", "ANY"}
}

// dnsRecordTypes 是记录类型到查询类型的映射。
var dnsRecordTypes = map[string]uint16{
	"A":     dns.TypeA,
	"AAAA":  dns.TypeAAAA,
	"CNAME": dns.TypeCNAME,
	"MX":    dns.TypeMX,
	"TXT":   dns.TypeTXT,
	"NS":    dns.TypeNS,
	"SOA":   dns.TypeSOA,
	"SRV":   dns.TypeSRV,
	"CAA":   dns.TypeCAA,
	"PTR":   dns.TypePTR,
	"ANY":   dns.TypeANY,
}

// DNSServerInput 是一台自定义 DNS 服务器。
type DNSServerInput struct {
	// Proto 是查询协议：udp / tcp / dot / doh。
	Proto string `json:"proto"`
	// Addr 是服务器地址；DoH 时是完整 URL。
	Addr string `json:"addr"`
}

// DNSAnswer 是一条解析结果。
type DNSAnswer struct {
	// Name 是记录名。
	Name string `json:"name"`
	// Type 是记录类型。
	Type string `json:"type"`
	// TTL 是生存时间（秒）。
	TTL uint32 `json:"ttl"`
	// Value 是记录值的文本形式。
	Value string `json:"value"`
}

// DNSServerError 是单台服务器的查询错误。
type DNSServerError struct {
	// Code 是错误码。
	Code string `json:"code"`
	// Message 是错误描述（不含用户输入）。
	Message string `json:"message"`
}

// DNSServerResult 是单台服务器的查询结果。
type DNSServerResult struct {
	// Server 是服务器标识。
	Server string `json:"server"`
	// Proto 是查询协议。
	Proto string `json:"proto"`
	// LatencyMs 是查询耗时（毫秒）。
	LatencyMs int64 `json:"latencyMs"`
	// Answers 是解析结果。
	Answers []DNSAnswer `json:"answers"`
	// Error 是查询错误；成功时为 null。
	Error *DNSServerError `json:"error"`
}

// DNSLookupResult 是 DNS 查询工具的响应。
type DNSLookupResult struct {
	// Host 是查询的域名。
	Host string `json:"host"`
	// Type 是记录类型。
	Type string `json:"type"`
	// PerServer 是逐服务器的结果。
	PerServer []DNSServerResult `json:"perServer"`
}

// DNSLookupRequest 是 DNS 查询工具的请求。
type DNSLookupRequest struct {
	// Host 是域名。
	Host string `json:"host"`
	// Type 是记录类型。
	Type string `json:"type"`
	// Servers 是自定义服务器；为空时使用系统配置的解析器。
	Servers []DNSServerInput `json:"servers"`
	// TimeoutMs 是单服务器超时。
	TimeoutMs int `json:"timeoutMs"`
}

// Validate 做字段级校验。
func (r *DNSLookupRequest) Validate() error {
	if _, err := normalizeHost(r.Host); err != nil {
		return err
	}
	typ := strings.ToUpper(strings.TrimSpace(r.Type))
	if typ == "" {
		return tools.MissingField("type")
	}
	if _, ok := dnsRecordTypes[typ]; !ok {
		return tools.InvalidField("type", DNSRecordTypes(), typ)
	}
	if len(r.Servers) > maxServers {
		return toolerrTooMany("servers", len(r.Servers), maxServers)
	}
	for _, srv := range r.Servers {
		if err := validateDNSServer(srv); err != nil {
			return err
		}
	}
	return nil
}

// validateDNSServer 校验一台自定义服务器。
func validateDNSServer(srv DNSServerInput) error {
	proto := strings.ToLower(strings.TrimSpace(srv.Proto))
	if proto == "" {
		return tools.MissingField("servers.proto")
	}
	if !containsString(DNSProtos(), proto) {
		return tools.InvalidField("servers.proto", DNSProtos(), proto)
	}
	addr := strings.TrimSpace(srv.Addr)
	if addr == "" {
		return tools.MissingField("servers.addr")
	}
	if proto == DNSProtoDoH {
		u, err := url.Parse(addr)
		if err != nil || u.Host == "" {
			return tools.InvalidField("servers.addr", "完整的 https URL", addr)
		}
		if strings.ToLower(u.Scheme) != "https" {
			return tools.InvalidField("servers.addr", "https", u.Scheme)
		}
		return nil
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return tools.InvalidField("servers.addr", "host:port", addr)
	}
	return nil
}

// DNSLookup 执行 DNS 查询。
func DNSLookup(ctx context.Context, g *netguard.Guard, req *DNSLookupRequest) (tools.Result, error) {
	if err := requireNetwork(g); err != nil {
		return tools.Result{}, err
	}
	host, err := normalizeHost(req.Host)
	if err != nil {
		return tools.Result{}, err
	}
	typ := strings.ToUpper(strings.TrimSpace(req.Type))
	qtype := dnsRecordTypes[typ]

	timeout, err := resolveTimeout(req.TimeoutMs, defaultTimeoutMs)
	if err != nil {
		return tools.Result{}, err
	}

	servers := req.Servers
	if len(servers) == 0 {
		system, sysErr := systemDNSServers()
		if sysErr != nil {
			return tools.Result{}, sysErr
		}
		servers = system
	}

	out := DNSLookupResult{Host: host, Type: typ, PerServer: make([]DNSServerResult, len(servers))}

	// 逐台并发查询：8 台 × 5s 的顺序查询会超出工具 20s 的预算，
	// 并发上限取 4，兼顾总时限与对上游的礼貌程度。
	// 结果按下标回填，保持与请求顺序一致。
	sem := make(chan struct{}, maxParallelQueries)
	var wg sync.WaitGroup
	for index, srv := range servers {
		wg.Add(1)
		go func(index int, server DNSServerInput) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out.PerServer[index] = queryOneServer(ctx, g, server, host, qtype, typ, timeout)
		}(index, srv)
	}
	wg.Wait()

	return tools.Result{Data: out}, nil
}

// queryOneServer 查询单台服务器，错误记录在结果里而不是整体失败。
func queryOneServer(
	ctx context.Context, g *netguard.Guard, srv DNSServerInput,
	host string, qtype uint16, typ string, timeout time.Duration,
) DNSServerResult {
	start := time.Now()
	result := DNSServerResult{
		Server:  srv.Addr,
		Proto:   srv.Proto,
		Answers: []DNSAnswer{},
	}

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), qtype)
	msg.RecursionDesired = true

	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := exchangeDNS(queryCtx, g, srv, msg)
	result.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		mapped := mapNetError(err)
		result.Error = &DNSServerError{Code: string(mapped.Code), Message: mapped.Error()}
		return result
	}
	if resp == nil {
		result.Error = &DNSServerError{Code: "DNS_ERROR", Message: "empty response"}
		return result
	}
	if resp.Rcode != dns.RcodeSuccess {
		result.Error = &DNSServerError{
			Code:    "DNS_ERROR",
			Message: dns.RcodeToString[resp.Rcode],
		}
		return result
	}
	result.Answers = renderAnswers(resp.Answer, typ)
	return result
}

// exchangeDNS 按协议发起一次查询。
//
// 所有建连都经 netguard；用户指定的服务器同样要做私网校验。
func exchangeDNS(ctx context.Context, g *netguard.Guard, srv DNSServerInput, msg *dns.Msg) (*dns.Msg, error) {
	switch srv.Proto {
	case DNSProtoDoH:
		return exchangeDoH(ctx, g, srv.Addr, msg)
	case DNSProtoDoT:
		return exchangeStream(ctx, g, srv.Addr, 853, true, msg)
	case DNSProtoTCP:
		return exchangeStream(ctx, g, srv.Addr, 53, false, msg)
	default:
		return exchangePacket(ctx, g, srv.Addr, msg)
	}
}

// exchangePacket 走 UDP。
func exchangePacket(ctx context.Context, g *netguard.Guard, addr string, msg *dns.Msg) (*dns.Msg, error) {
	conn, err := dialDNS(ctx, g, addr, 53, "udp")
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	co := &dns.Conn{Conn: conn}
	co.UDPSize = dns.MinMsgSize
	if deadline, ok := ctx.Deadline(); ok {
		_ = co.SetDeadline(deadline)
	}
	if writeErr := co.WriteMsg(msg); writeErr != nil {
		return nil, writeErr
	}
	return co.ReadMsg()
}

// exchangeStream 走 TCP 或 TLS。
func exchangeStream(ctx context.Context, g *netguard.Guard, addr string, defPort int, useTLS bool, msg *dns.Msg) (*dns.Msg, error) {
	conn, err := dialDNS(ctx, g, addr, defPort, "tcp")
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	if useTLS {
		host, _, _ := net.SplitHostPort(addrWithDefaultPort(addr, defPort))
		cfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		tlsConn := tls.Client(conn, cfg)
		defer func() { _ = tlsConn.Close() }()
		if hsErr := tlsConn.HandshakeContext(ctx); hsErr != nil {
			return nil, hsErr
		}
		conn = tlsConn
	}

	co := &dns.Conn{Conn: conn}
	if deadline, ok := ctx.Deadline(); ok {
		_ = co.SetDeadline(deadline)
	}
	if writeErr := co.WriteMsg(msg); writeErr != nil {
		return nil, writeErr
	}
	return co.ReadMsg()
}

// exchangeDoH 走 DNS over HTTPS。
func exchangeDoH(ctx context.Context, g *netguard.Guard, rawURL string, msg *dns.Msg) (*dns.Msg, error) {
	packed, err := msg.Pack()
	if err != nil {
		return nil, err
	}
	client := g.HTTPClient(netguard.HTTPOptions{})
	req, err := newDoHRequest(ctx, rawURL, packed)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("doh status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, dns.MaxMsgSize))
	if err != nil {
		return nil, err
	}
	out := new(dns.Msg)
	if err := out.Unpack(body); err != nil {
		return nil, err
	}
	return out, nil
}

// dialDNS 经 netguard 建立到 DNS 服务器的连接。
//
// 自定义服务器由用户提供，因此**必须**做私网与端口校验。
func dialDNS(ctx context.Context, g *netguard.Guard, addr string, defPort int, network string) (net.Conn, error) {
	full := addrWithDefaultPort(addr, defPort)
	return g.DialContext(ctx, network, full)
}

// addrWithDefaultPort 在缺少端口时补默认端口。
func addrWithDefaultPort(addr string, defPort int) string {
	addr = strings.TrimSpace(addr)
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}
	return net.JoinHostPort(strings.Trim(addr, "[]"), strconv.Itoa(defPort))
}

// systemDNSServers 读取系统配置的解析器。
//
// 这些服务器来自 `/etc/resolv.conf`（运维配置，非用户输入），
// 因此**不**做私网校验 —— 容器与桌面环境里它们常常就是回环上的桩解析器。
func systemDNSServers() ([]DNSServerInput, error) {
	cfg, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil || len(cfg.Servers) == 0 {
		return nil, toolerr.New(toolerr.CodeDNSError)
	}
	out := make([]DNSServerInput, 0, len(cfg.Servers))
	for _, s := range cfg.Servers {
		out = append(out, DNSServerInput{
			Proto: DNSProtoUDP,
			Addr:  net.JoinHostPort(s, cfg.Port),
		})
	}
	return out, nil
}

// renderAnswers 把资源记录渲染为对外结构。
func renderAnswers(rrs []dns.RR, fallbackType string) []DNSAnswer {
	out := make([]DNSAnswer, 0, len(rrs))
	for _, rr := range rrs {
		typ, value := renderRR(rr)
		if typ == "" {
			continue
		}
		if fallbackType != "ANY" && typ != fallbackType {
			continue
		}
		header := rr.Header()
		out = append(out, DNSAnswer{
			Name:  strings.TrimSuffix(header.Name, "."),
			Type:  typ,
			TTL:   header.Ttl,
			Value: value,
		})
	}
	return out
}

// renderRR 渲染单条记录。
func renderRR(rr dns.RR) (string, string) {
	switch v := rr.(type) {
	case *dns.A:
		return "A", v.A.String()
	case *dns.AAAA:
		return "AAAA", v.AAAA.String()
	case *dns.CNAME:
		return "CNAME", strings.TrimSuffix(v.Target, ".")
	case *dns.PTR:
		return "PTR", strings.TrimSuffix(v.Ptr, ".")
	case *dns.NS:
		return "NS", strings.TrimSuffix(v.Ns, ".")
	case *dns.MX:
		return "MX", fmt.Sprintf("%d %s", v.Preference, strings.TrimSuffix(v.Mx, "."))
	case *dns.TXT:
		return "TXT", strings.Join(v.Txt, "")
	case *dns.SOA:
		return "SOA", fmt.Sprintf("%s %s %d %d %d %d %d",
			strings.TrimSuffix(v.Ns, "."), strings.TrimSuffix(v.Mbox, "."),
			v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl)
	case *dns.SRV:
		return "SRV", fmt.Sprintf("%d %d %d %s", v.Priority, v.Weight, v.Port, strings.TrimSuffix(v.Target, "."))
	case *dns.CAA:
		return "CAA", fmt.Sprintf("%d %s %q", v.Flag, v.Tag, v.Value)
	case *dns.DS:
		return "DS", fmt.Sprintf("%d %d %d %s", v.KeyTag, v.Algorithm, v.DigestType, strings.ToUpper(v.Digest))
	case *dns.DNSKEY:
		return "DNSKEY", fmt.Sprintf("%d %d %d", v.Flags, v.Protocol, v.Algorithm)
	case *dns.RRSIG:
		return "RRSIG", fmt.Sprintf("%s %d %d %d", dns.TypeToString[v.TypeCovered], v.Algorithm, v.KeyTag, v.OrigTtl)
	default:
		return "", ""
	}
}

// newDoHRequest 构造 DoH 的 POST 请求。
func newDoHRequest(ctx context.Context, rawURL string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")
	return req, nil
}
