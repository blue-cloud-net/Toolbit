package netguard

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// 建连与握手的固定上限。
const (
	// dialTimeout 是单次 TCP 建连超时。
	dialTimeout = 10 * time.Second
	// handshakeTimeout 是 TLS 握手超时。
	handshakeTimeout = 15 * time.Second
	// maxRedirects 是允许的最大重定向跳数。
	maxRedirects = 3
	// maxHeaderBytes 是出网响应头的读取上限。
	maxHeaderBytes = 64 * 1024
)

// DialIP 与**已校验**的地址建连。
//
// 调用方必须先经 Resolve 取得地址，再交给本方法 —— 这样连接用的就是校验过的 IP，
// 攻击者无法靠二次解析把连接引导到内网（DNS rebinding）。
func (g *Guard) DialIP(ctx context.Context, network string, addr netip.Addr, port int) (net.Conn, error) {
	if err := g.requireEnabled(); err != nil {
		return nil, err
	}
	if !g.allowPort(port) {
		return nil, &Error{Reason: ReasonPort, Port: port}
	}
	if err := ValidateAddr(addr); err != nil && !g.allowsHost(addr.String()) {
		return nil, err
	}
	return g.dialer.DialContext(ctx, network, net.JoinHostPort(addr.Unmap().String(), strconv.Itoa(port)))
}

// DialContext 实现 net.Dialer 的签名，供 http.Transport 使用。
//
// 内部完成「解析 → 全部校验 → 按顺序尝试已校验的 IP」。
func (g *Guard) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if err := g.requireEnabled(); err != nil {
		return nil, err
	}
	host, port, err := g.hostPort(address)
	if err != nil {
		return nil, err
	}
	addrs, err := g.Resolve(ctx, host)
	if err != nil {
		return nil, err
	}

	var lastErr error = &Error{Reason: ReasonHost, Host: host}
	for _, addr := range addrs {
		conn, dialErr := g.dialer.DialContext(ctx, network, net.JoinHostPort(addr.String(), strconv.Itoa(port)))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, lastErr
}

// TLSOptions 是 DialTLS 的参数。
type TLSOptions struct {
	// Host 是目标主机名（用于解析）。
	Host string
	// Port 是目标端口，必须在白名单内。
	Port int
	// ServerName 是 SNI；为空时用 Host。
	ServerName string
	// Versions 是允许的 TLS 版本下限 / 上限；零值表示用 Go 默认。
	MinVersion uint16
	// MaxVersion 是允许的 TLS 版本上限。
	MaxVersion uint16
	// CipherSuites 是可选的密码套件 ID 列表；为空表示用默认。
	CipherSuites []uint16
	// InsecureSkipVerify 仅用于**探测**：我们要自己读取并展示证书链，
	// 因此不依赖握手期的链验证结论。
	InsecureSkipVerify bool
	// HandshakeTimeout 是握手超时；0 表示用默认值。
	HandshakeTimeout time.Duration
}

// TLSConn 是一次已完成的 TLS 连接。
type TLSConn struct {
	// Conn 是底层连接；调用方负责关闭。
	Conn *tls.Conn
	// Addr 是实际连接的地址。
	Addr netip.Addr
	// ServerName 是实际使用的 SNI。
	ServerName string
}

// DialTLS 在已校验的地址上完成 TLS 握手。
//
// 返回的 Conn 已完成握手，调用方可直接 ConnectionState()。
func (g *Guard) DialTLS(ctx context.Context, opts TLSOptions) (*TLSConn, error) {
	if err := g.requireEnabled(); err != nil {
		return nil, err
	}
	if !g.allowPort(opts.Port) {
		return nil, &Error{Reason: ReasonPort, Host: opts.Host, Port: opts.Port}
	}
	addrs, err := g.Resolve(ctx, opts.Host)
	if err != nil {
		return nil, err
	}

	sni := opts.ServerName
	if sni == "" {
		sni = opts.Host
	}

	timeout := opts.HandshakeTimeout
	if timeout <= 0 {
		timeout = handshakeTimeout
	}

	var lastErr error
	for _, addr := range addrs {
		dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
		conn, dialErr := g.dialer.DialContext(dialCtx, "tcp", net.JoinHostPort(addr.String(), strconv.Itoa(opts.Port)))
		cancel()
		if dialErr != nil {
			lastErr = dialErr
			continue
		}

		cfg := &tls.Config{
			ServerName:         sni,
			MinVersion:         opts.MinVersion,
			MaxVersion:         opts.MaxVersion,
			CipherSuites:       opts.CipherSuites,
			InsecureSkipVerify: opts.InsecureSkipVerify, //nolint:gosec // 探测工具需要读取未验证的链
		}
		tlsConn := tls.Client(conn, cfg)

		hsCtx, hsCancel := context.WithTimeout(ctx, timeout)
		hsErr := tlsConn.HandshakeContext(hsCtx)
		hsCancel()
		if hsErr != nil {
			_ = conn.Close()
			lastErr = hsErr
			continue
		}
		return &TLSConn{Conn: tlsConn, Addr: addr, ServerName: sni}, nil
	}

	if lastErr == nil {
		lastErr = &Error{Reason: ReasonHost, Host: opts.Host}
	}
	return nil, lastErr
}

// HTTPOptions 是 HTTPClient 的参数。
type HTTPOptions struct {
	// Timeout 是整体请求超时；0 表示由调用方 ctx 控制。
	Timeout time.Duration
	// MaxRedirects 覆盖默认跳数上限；0 表示用默认值。
	MaxRedirects int
}

// HTTPClient 返回受控的 http.Client。
//
// 关键点：
//   - Transport 的 DialContext 指向本包，因此每一跳都重新解析并校验；
//   - 禁用环境代理（`HTTP_PROXY` 可被用来绕过校验）；
//   - 重定向逐跳校验，最多 3 跳；
//   - 只允许 http / https。
func (g *Guard) HTTPClient(opts HTTPOptions) *http.Client {
	maxHops := opts.MaxRedirects
	if maxHops <= 0 {
		maxHops = maxRedirects
	}
	transport := &http.Transport{
		Proxy:                  nil,
		DialContext:            g.DialContext,
		ForceAttemptHTTP2:      true,
		MaxIdleConns:           8,
		IdleConnTimeout:        30 * time.Second,
		TLSHandshakeTimeout:    handshakeTimeout,
		ExpectContinueTimeout:  time.Second,
		ResponseHeaderTimeout:  0,
		MaxResponseHeaderBytes: maxHeaderBytes,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxHops {
				return &Error{Reason: ReasonRedirect, Host: req.URL.Hostname()}
			}
			return g.validateURL(req.URL)
		},
	}
}

// validateURL 校验 URL 的协议、端口与主机。
func (g *Guard) validateURL(u *url.URL) error {
	if u == nil {
		return &Error{Reason: ReasonHost}
	}
	if err := CheckScheme(u.Scheme); err != nil {
		return err
	}
	port := portOf(u)
	if !g.allowPort(port) {
		return &Error{Reason: ReasonPort, Host: u.Hostname(), Port: port}
	}
	_, err := g.Resolve(context.Background(), u.Hostname())
	return err
}

// portOf 返回 URL 的显式端口；缺省按协议补全。
func portOf(u *url.URL) int {
	if p := u.Port(); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil {
			return -1
		}
		return v
	}
	switch strings.ToLower(u.Scheme) {
	case "http":
		return 80
	case "https":
		return 443
	default:
		return -1
	}
}
