package network

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// whois 相关常量。
const (
	// whoisPort 是 whois 协议端口。
	whoisPort = 43
	// whoisBootstrap 是 IANA 的 whois 引导服务器。
	whoisBootstrap = "whois.iana.org"
	// whoisMaxBytes 是单次 whois 响应的读取上限。
	whoisMaxBytes = 128 * 1024
	// rdapBase 是公共 RDAP 引导服务。
	rdapBase = "https://rdap.org/"
	// rdapMaxBytes 是 RDAP 响应的读取上限。
	rdapMaxBytes = 1024 * 1024
)

// WhoisInfo 是域名 / 地址的注册信息。
//
// 字段可能为空：部分注册局限制查询或对字段脱敏。
type WhoisInfo struct {
	// Source 是数据来源：rdap 或 whois。
	Source string `json:"source"`
	// Registrar 是注册商。
	Registrar string `json:"registrar"`
	// CreationDate 是注册时间。
	CreationDate string `json:"creationDate"`
	// ExpirationDate 是到期时间。
	ExpirationDate string `json:"expirationDate"`
	// UpdatedDate 是最后更新时间。
	UpdatedDate string `json:"updatedDate"`
	// NameServers 是名称服务器。
	NameServers []string `json:"nameServers"`
	// Status 是状态码列表。
	Status []string `json:"status"`
	// RegistrantCountry 是注册人国家 / 地区。
	RegistrantCountry string `json:"registrantCountry"`
	// Registries 是查询过的注册局 / 注册商主机。
	Registries []string `json:"registries"`
}

// lookupWhois 查询域名或 IP 的注册信息。
//
// 优先走 RDAP（HTTPS，结构化）；失败时回退到 whois 协议（TCP 43）。
// 两条路径都经 netguard，端口固定不可由用户指定。
func lookupWhois(ctx context.Context, g *netguard.Guard, query string, timeout time.Duration) (*WhoisInfo, error) {
	if info, err := lookupRDAP(ctx, g, query, timeout); err == nil {
		return info, nil
	}
	return lookupWhoisText(ctx, g, query, timeout)
}

// lookupRDAP 通过 RDAP 查询。
func lookupRDAP(ctx context.Context, g *netguard.Guard, query string, timeout time.Duration) (*WhoisInfo, error) {
	kind := "domain"
	if addr, err := netip.ParseAddr(query); err == nil && addr.IsValid() {
		kind = "ip"
	}
	rawURL := rdapBase + kind + "/" + query

	client := g.HTTPClient(netguard.HTTPOptions{Timeout: timeout})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/rdap+json, application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, toolerr.New(toolerr.CodeUpstreamError)
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, rdapMaxBytes))
	if readErr != nil {
		return nil, readErr
	}

	var doc rdapDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return doc.toInfo("rdap", rdapBase), nil
}

// rdapDocument 是 RDAP 响应的最小子集。
type rdapDocument struct {
	Handle      string `json:"handle"`
	LDHName     string `json:"ldhName"`
	Status      []string
	Nameservers []struct {
		LDHName string `json:"ldhName"`
	} `json:"nameservers"`
	Events []struct {
		Action string `json:"eventAction"`
		Date   string `json:"eventDate"`
	} `json:"events"`
	Entities []struct {
		Roles      []string `json:"roles"`
		VCardArray []any    `json:"vcardArray"`
	} `json:"entities"`
}

// toInfo 把 RDAP 文档转成统一结构。
func (d *rdapDocument) toInfo(source, registry string) *WhoisInfo {
	info := &WhoisInfo{
		Source:      source,
		NameServers: []string{},
		Status:      d.Status,
		Registries:  []string{registry},
	}
	for _, ns := range d.Nameservers {
		if ns.LDHName != "" {
			info.NameServers = append(info.NameServers, strings.ToLower(ns.LDHName))
		}
	}
	for _, event := range d.Events {
		switch strings.ToLower(event.Action) {
		case "registration":
			info.CreationDate = event.Date
		case "expiration":
			info.ExpirationDate = event.Date
		case "last changed", "last update of rdap database":
			if info.UpdatedDate == "" {
				info.UpdatedDate = event.Date
			}
		}
	}
	for _, entity := range d.Entities {
		if !containsString(lowerAll(entity.Roles), "registrar") {
			continue
		}
		info.Registrar = vcardValue(entity.VCardArray, "fn")
		if country := vcardValue(entity.VCardArray, "adr"); country != "" {
			info.RegistrantCountry = country
		}
		break
	}
	return info
}

// vcardValue 从 vCard 数组中取指定属性的文本值。
func vcardValue(vcard []any, name string) string {
	if len(vcard) < 2 {
		return ""
	}
	entries, ok := vcard[1].([]any)
	if !ok {
		return ""
	}
	for _, raw := range entries {
		item, ok := raw.([]any)
		if !ok || len(item) < 4 {
			continue
		}
		key, _ := item[0].(string)
		if !strings.EqualFold(key, name) {
			continue
		}
		switch value := item[3].(type) {
		case string:
			return value
		case []any:
			parts := make([]string, 0, len(value))
			for _, part := range value {
				if text, okText := part.(string); okText && text != "" {
					parts = append(parts, text)
				}
			}
			return strings.Join(parts, " ")
		}
	}
	return ""
}

// lookupWhoisText 通过 whois 协议查询。
func lookupWhoisText(ctx context.Context, g *netguard.Guard, query string, timeout time.Duration) (*WhoisInfo, error) {
	host := whoisBootstrap
	registries := make([]string, 0, 2)
	var lastBody string

	for hop := 0; hop < 2; hop++ {
		registries = append(registries, host)
		body, err := whoisExchange(ctx, g, host, query, timeout)
		if err != nil {
			return nil, err
		}
		lastBody = body
		next := referServerOf(body)
		if next == "" || next == host {
			break
		}
		host = next
	}

	info := parseWhoisText(lastBody)
	info.Source = "whois"
	info.Registries = registries
	return info, nil
}

// whoisExchange 向一台 whois 服务器发起查询。
func whoisExchange(ctx context.Context, g *netguard.Guard, host, query string, timeout time.Duration) (string, error) {
	conn, err := g.DialContext(ctx, "tcp", net.JoinHostPort(host, itoa(whoisPort)))
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(timeout))
	}

	if _, writeErr := conn.Write([]byte(query + "\r\n")); writeErr != nil {
		return "", writeErr
	}
	buf, readErr := io.ReadAll(io.LimitReader(conn, whoisMaxBytes))
	if readErr != nil {
		return "", readErr
	}
	return string(buf), nil
}

// referServerOf 从 whois 响应里取下一个服务器。
func referServerOf(body string) string {
	for _, line := range strings.Split(body, "\n") {
		text := strings.TrimSpace(line)
		lower := strings.ToLower(text)
		for _, key := range []string{"refer:", "whois:", "referralserver:"} {
			if strings.HasPrefix(lower, key) {
				value := strings.TrimSpace(text[len(key):])
				value = strings.TrimPrefix(value, "//")
				if host, _, err := net.SplitHostPort(value); err == nil {
					return host
				}
				if value != "" && !strings.ContainsAny(value, " /") {
					return value
				}
			}
		}
	}
	return ""
}

// parseWhoisText 解析 whois 文本响应。
func parseWhoisText(body string) *WhoisInfo {
	info := &WhoisInfo{
		NameServers: []string{},
		Status:      []string{},
		Registries:  []string{},
	}
	seenNS := make(map[string]bool, 4)

	for _, line := range strings.Split(body, "\n") {
		text := strings.TrimSpace(line)
		if text == "" || strings.HasPrefix(text, "%") || strings.HasPrefix(text, "#") {
			continue
		}
		key, value, ok := strings.Cut(text, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		switch key {
		case "registrar", "sponsoring registrar", "registrar name":
			if info.Registrar == "" {
				info.Registrar = value
			}
		case "creation date", "created", "registered on", "registration time":
			if info.CreationDate == "" {
				info.CreationDate = value
			}
		case "registry expiry date", "expiry date", "expiration time", "expiration date":
			if info.ExpirationDate == "" {
				info.ExpirationDate = value
			}
		case "updated date", "last updated", "last modified":
			if info.UpdatedDate == "" {
				info.UpdatedDate = value
			}
		case "name server", "nameserver", "nserver":
			name := strings.ToLower(strings.Fields(value)[0])
			if !seenNS[name] {
				seenNS[name] = true
				info.NameServers = append(info.NameServers, name)
			}
		case "domain status", "status":
			state := strings.Fields(value)[0]
			if !containsString(info.Status, state) {
				info.Status = append(info.Status, state)
			}
		case "registrant country", "country":
			if info.RegistrantCountry == "" {
				info.RegistrantCountry = value
			}
		}
	}
	return info
}

// lowerAll 返回全小写副本。
func lowerAll(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, strings.ToLower(item))
	}
	return out
}
