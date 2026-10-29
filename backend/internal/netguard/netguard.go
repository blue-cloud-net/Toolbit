// Package netguard 提供统一的出网防护。
//
// 全部出网工具**必须**经本包建连：工具里**禁止**直接 `http.Get` / `net.Dial`。
// 本包负责四件事：
//
//  1. **解析并校验**：目标主机的**全部**解析结果都必须通过私网 / 保留网段校验；
//  2. **钉住地址**：建连时使用已校验的 IP，而不是再解析一次（防 DNS rebinding）；
//  3. **端口白名单**：只允许固定端口，用户**不能**指定任意端口；
//  4. **重定向逐跳校验**：每一跳都重新校验，且最多 3 跳。
//
// 本包**不做**密码学运算，也不认识任何工具语义。
package netguard

import (
	"context"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

// 固定端口白名单。
//
// 端口由**服务端**按工具语义固定选择，用户输入无法影响此处取值：
// 43 仅用于 whois，53 仅用于自定义 DNS 服务器，80 仅用于 domain-verify 的
// http-file，443 用于 HTTPS / DoH / TLS 探测，853 用于 DoT。
var allowedPorts = map[uint16]struct{}{
	43:  {},
	53:  {},
	80:  {},
	443: {},
	853: {},
}

// blockedPrefixes 是禁止访问的网段。
//
// 顺序无关；匹配前会先把 IPv4-mapped IPv6 还原为 IPv4。
var blockedPrefixes = []netip.Prefix{
	// IPv4
	netip.MustParsePrefix("0.0.0.0/8"),       // 未指定
	netip.MustParsePrefix("10.0.0.0/8"),      // 私有
	netip.MustParsePrefix("100.64.0.0/10"),   // 运营商级 NAT
	netip.MustParsePrefix("127.0.0.0/8"),     // 环回
	netip.MustParsePrefix("169.254.0.0/16"),  // 链路本地
	netip.MustParsePrefix("172.16.0.0/12"),   // 私有
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF 协议分配
	netip.MustParsePrefix("192.0.2.0/24"),    // 文档
	netip.MustParsePrefix("192.168.0.0/16"),  // 私有
	netip.MustParsePrefix("198.18.0.0/15"),   // 基准测试
	netip.MustParsePrefix("198.51.100.0/24"), // 文档
	netip.MustParsePrefix("203.0.113.0/24"),  // 文档
	netip.MustParsePrefix("224.0.0.0/4"),     // 组播
	netip.MustParsePrefix("240.0.0.0/4"),     // 保留
	// IPv6
	netip.MustParsePrefix("::/128"),        // 未指定
	netip.MustParsePrefix("::1/128"),       // 环回
	netip.MustParsePrefix("64:ff9b::/96"),  // NAT64
	netip.MustParsePrefix("100::/64"),      // 丢弃
	netip.MustParsePrefix("2001:db8::/32"), // 文档
	netip.MustParsePrefix("fc00::/7"),      // ULA
	netip.MustParsePrefix("fe80::/10"),     // 链路本地
	netip.MustParsePrefix("ff00::/8"),      // 组播
}

// Config 是出网守卫的配置。
type Config struct {
	// Enabled 是出网总开关；false 时任何出网请求都返回 ErrDisabled。
	Enabled bool
	// Allowlist 是额外放行的主机。
	//
	// 支持精确主机名（`dns.internal`）与通配后缀（`*.internal`）。
	// 命中的主机**跳过私网校验**，用于自建内网服务；
	// 端口与协议白名单**仍然生效**。
	Allowlist []string
	// ExtraPorts 是额外放行的端口。
	//
	// **仅供测试**：生产配置不从环境变量读取该值，因此无法在部署时放宽端口白名单。
	// 唯一用途是让测试能把工具指向本地 httptest 服务器的随机端口。
	ExtraPorts []int
}

// Guard 是出网守卫。
//
// 零值不可用；必须用 New 构造。
// Guard 是并发安全的：内部无共享可变状态。
type Guard struct {
	enabled   bool
	allowlist []string
	resolver  *net.Resolver
	dialer    *net.Dialer
	ports     map[uint16]struct{}
}

// New 构造出网守卫。
func New(cfg Config) *Guard {
	allow := make([]string, 0, len(cfg.Allowlist))
	for _, entry := range cfg.Allowlist {
		if v := strings.ToLower(strings.TrimSpace(entry)); v != "" {
			allow = append(allow, v)
		}
	}
	ports := make(map[uint16]struct{}, len(allowedPorts)+len(cfg.ExtraPorts))
	for p := range allowedPorts {
		ports[p] = struct{}{}
	}
	for _, p := range cfg.ExtraPorts {
		if p > 0 && p <= 65535 {
			ports[uint16(p)] = struct{}{}
		}
	}
	return &Guard{
		enabled:   cfg.Enabled,
		allowlist: allow,
		resolver:  net.DefaultResolver,
		dialer:    &net.Dialer{Timeout: dialTimeout},
		ports:     ports,
	}
}

// Enabled 报告出网总开关状态。
func (g *Guard) Enabled() bool {
	return g != nil && g.enabled
}

// Allowlist 返回实际生效的放行主机列表（小写）。
func (g *Guard) Allowlist() []string {
	if g == nil {
		return nil
	}
	return append([]string(nil), g.allowlist...)
}

// AllowedPort 报告端口是否在默认白名单内。
func AllowedPort(port int) bool {
	if port <= 0 || port > 65535 {
		return false
	}
	_, ok := allowedPorts[uint16(port)]
	return ok
}

// allowPort 报告该守卫是否允许此端口。
//
// 生产环境与 AllowedPort 等价；测试可临时放宽以指向本地 httptest 服务器。
func (g *Guard) allowPort(port int) bool {
	if port <= 0 || port > 65535 {
		return false
	}
	if g == nil || g.ports == nil {
		return AllowedPort(port)
	}
	_, ok := g.ports[uint16(port)]
	return ok
}

// CheckScheme 校验协议是否受支持。
func CheckScheme(scheme string) error {
	switch strings.ToLower(scheme) {
	case "https", "http":
		return nil
	default:
		return &Error{Reason: ReasonScheme, Host: scheme}
	}
}

// ValidateAddr 校验单个地址。
//
// IPv4-mapped IPv6（如 `::ffff:127.0.0.1`）会先还原为 IPv4 再判断，
// 因此**无法**用它绕过私网校验。
func ValidateAddr(addr netip.Addr) error {
	if !addr.IsValid() {
		return &Error{Reason: ReasonHost}
	}
	unmapped := addr.Unmap()
	for _, prefix := range blockedPrefixes {
		if prefix.Contains(unmapped) {
			return &Error{Reason: ReasonAddress, Addr: unmapped.String()}
		}
	}
	return nil
}

// ValidateAddrString 校验字符串形式的地址。
func ValidateAddrString(raw string) error {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return &Error{Reason: ReasonHost}
	}
	return ValidateAddr(addr)
}

// allowsHost 报告主机是否命中放行列表。
func (g *Guard) allowsHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, entry := range g.allowlist {
		if entry == host {
			return true
		}
		if suffix, ok := strings.CutPrefix(entry, "*"); ok && strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

// Resolve 解析主机并校验**全部**结果。
//
// 返回按解析顺序排列的地址；只要有一个地址落在禁止网段就整体拒绝。
// 主机本身是 IP 字面量时不查 DNS，直接校验。
func (g *Guard) Resolve(ctx context.Context, host string) ([]netip.Addr, error) {
	if g == nil || !g.enabled {
		return nil, ErrDisabled
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, &Error{Reason: ReasonHost}
	}

	if addr, err := netip.ParseAddr(strings.Trim(host, "[]")); err == nil {
		if err := ValidateAddr(addr); err != nil {
			if g.allowsHost(host) {
				return []netip.Addr{addr.Unmap()}, nil
			}
			return nil, err
		}
		return []netip.Addr{addr.Unmap()}, nil
	}

	addrs, err := g.resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, &Error{Reason: ReasonHost, Host: host}
	}

	out := make([]netip.Addr, 0, len(addrs))
	for _, addr := range addrs {
		if err := ValidateAddr(addr); err != nil {
			if g.allowsHost(host) {
				out = append(out, addr.Unmap())
				continue
			}
			return nil, err
		}
		out = append(out, addr.Unmap())
	}
	return out, nil
}

// hostPort 拆分并校验 host:port。
func (g *Guard) hostPort(address string) (string, int, error) {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, &Error{Reason: ReasonHost}
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return "", 0, &Error{Reason: ReasonPort}
	}
	if !g.allowPort(port) {
		return "", 0, &Error{Reason: ReasonPort, Host: host, Port: port}
	}
	return host, port, nil
}

// requireEnabled 是全部出口的统一前置检查。
func (g *Guard) requireEnabled() error {
	if g == nil || !g.enabled {
		return ErrDisabled
	}
	return nil
}
