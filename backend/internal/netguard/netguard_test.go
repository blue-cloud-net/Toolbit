package netguard

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
)

// newTestGuard 构造一个放宽端口、放行环回的守卫，供 httptest 使用。
//
// 生产环境**不会**这样配置：端口白名单无法从外部放宽，环回也无法通过
// 环境变量放行（allowlist 是显式运维配置）。
func newTestGuard(t *testing.T, port int, allowLoopback bool) *Guard {
	t.Helper()
	allow := []string(nil)
	if allowLoopback {
		allow = []string{"127.0.0.1"}
	}
	g := New(Config{Enabled: true, Allowlist: allow})
	if port > 0 {
		g.ports[uint16(port)] = struct{}{}
	}
	return g
}

func TestValidateAddrBlocksReservedRanges(t *testing.T) {
	t.Parallel()

	blocked := []string{
		"127.0.0.1", "127.1.2.3", "10.0.0.1", "172.16.0.1", "172.31.255.255",
		"192.168.1.1", "169.254.169.254", "0.0.0.0", "100.64.0.1",
		"192.0.2.1", "198.18.0.1", "198.51.100.1", "203.0.113.1",
		"224.0.0.1", "240.0.0.1",
		"::", "::1", "fc00::1", "fd12:3456::1", "fe80::1", "ff02::1",
		"2001:db8::1", "64:ff9b::1",
		// IPv4-mapped IPv6 必须被还原后拦截。
		"::ffff:127.0.0.1", "::ffff:10.0.0.1", "::ffff:192.168.1.1",
	}
	for _, raw := range blocked {
		addr := netip.MustParseAddr(raw)
		if err := ValidateAddr(addr); err == nil {
			t.Errorf("ValidateAddr(%s) 应当被拒绝，实际通过", raw)
		}
	}

	allowed := []string{
		"1.1.1.1", "8.8.8.8", "93.184.216.34", "203.0.114.1",
		"2606:4700:4700::1111", "2001:4860:4860::8888",
	}
	for _, raw := range allowed {
		addr := netip.MustParseAddr(raw)
		if err := ValidateAddr(addr); err != nil {
			t.Errorf("ValidateAddr(%s) 应当通过，实际 %v", raw, err)
		}
	}
}

func TestValidateAddrRejectsInvalid(t *testing.T) {
	t.Parallel()

	if err := ValidateAddr(netip.Addr{}); err == nil {
		t.Fatal("无效地址应当被拒绝")
	}
	if err := ValidateAddrString("not-an-ip"); err == nil {
		t.Fatal("非 IP 字符串应当被拒绝")
	}
}

func TestAllowedPort(t *testing.T) {
	t.Parallel()

	for _, port := range []int{43, 53, 80, 443, 853} {
		if !AllowedPort(port) {
			t.Errorf("端口 %d 应当在白名单内", port)
		}
	}
	for _, port := range []int{0, -1, 22, 25, 3306, 6379, 8080, 65536} {
		if AllowedPort(port) {
			t.Errorf("端口 %d 不应在白名单内", port)
		}
	}
}

func TestCheckScheme(t *testing.T) {
	t.Parallel()

	for _, scheme := range []string{"http", "https", "HTTPS"} {
		if err := CheckScheme(scheme); err != nil {
			t.Errorf("协议 %s 应当被接受，实际 %v", scheme, err)
		}
	}
	for _, scheme := range []string{"file", "gopher", "ftp", "ws", ""} {
		if err := CheckScheme(scheme); err == nil {
			t.Errorf("协议 %s 应当被拒绝", scheme)
		}
	}
}

func TestResolveRejectsLoopbackLiteral(t *testing.T) {
	t.Parallel()

	g := New(Config{Enabled: true})
	if _, err := g.Resolve(context.Background(), "127.0.0.1"); err == nil {
		t.Fatal("环回字面量应当被拒绝")
	}
}

func TestResolveRejectsLocalhost(t *testing.T) {
	t.Parallel()

	g := New(Config{Enabled: true})
	_, err := g.Resolve(context.Background(), "localhost")
	if err == nil {
		t.Fatal("localhost 应当被拒绝")
	}
	var ngErr *Error
	if !errors.As(err, &ngErr) || ngErr.Reason != ReasonAddress {
		t.Fatalf("期望 private-address 拒绝，实际 %v", err)
	}
}

func TestResolveAllowlistBypassesPrivateCheck(t *testing.T) {
	t.Parallel()

	g := New(Config{Enabled: true, Allowlist: []string{"127.0.0.1"}})
	addrs, err := g.Resolve(context.Background(), "127.0.0.1")
	if err != nil {
		t.Fatalf("放行后不应报错：%v", err)
	}
	if len(addrs) != 1 || addrs[0].String() != "127.0.0.1" {
		t.Fatalf("地址不符：%v", addrs)
	}
}

func TestResolveAllowlistWildcard(t *testing.T) {
	t.Parallel()

	g := New(Config{Enabled: true, Allowlist: []string{"*.internal"}})
	if !g.allowsHost("db.internal") {
		t.Fatal("通配后缀应当命中")
	}
	if !g.allowsHost("DB.INTERNAL.") {
		t.Fatal("应当忽略大小写与末尾点")
	}
	if g.allowsHost("internal") {
		t.Fatal("缺省前导点不应命中")
	}
	if g.allowsHost("notinternal") {
		t.Fatal("后缀必须包含分隔点")
	}
}

func TestDisabledGuardRejectsEverything(t *testing.T) {
	t.Parallel()

	g := New(Config{Enabled: false})
	if g.Enabled() {
		t.Fatal("开关关闭时 Enabled 应为 false")
	}
	if _, err := g.Resolve(context.Background(), "1.1.1.1"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("期望 ErrDisabled，实际 %v", err)
	}
	client := g.HTTPClient(HTTPOptions{})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/", nil)
	if err != nil {
		t.Fatalf("构造请求失败：%v", err)
	}
	if _, doErr := client.Do(req); !errors.Is(doErr, ErrDisabled) {
		t.Fatalf("HTTP 请求期望 ErrDisabled，实际 %v", doErr)
	}
}

func TestHTTPClientServesAllowedLoopback(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	g := newTestGuard(t, portOf(u), true)

	client := g.HTTPClient(HTTPOptions{})
	resp, err := doGet(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("请求失败：%v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf("响应体不符：%q", body)
	}
}

func TestHTTPClientBlocksLoopbackWithoutAllowlist(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	g := newTestGuard(t, portOf(u), false)

	client := g.HTTPClient(HTTPOptions{})
	if _, getErr := doGet(context.Background(), client, srv.URL); getErr == nil {
		t.Fatal("未放行环回时请求应当失败")
	} else if !strings.Contains(getErr.Error(), "私有或保留网段") {
		t.Fatalf("期望私网拦截错误，实际 %v", getErr)
	}
}

func TestHTTPClientRedirectHopsLimited(t *testing.T) {
	t.Parallel()

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL+"/next", http.StatusFound)
	}))
	defer srv.Close()

	g := newTestGuard(t, portOf(mustURL(t, srv.URL)), true)
	client := g.HTTPClient(HTTPOptions{})

	_, err := doGet(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("无限重定向应当被拒绝")
	}
	var ngErr *Error
	if !errors.As(err, &ngErr) || ngErr.Reason != ReasonRedirect {
		t.Fatalf("期望 too-many-redirects，实际 %v", err)
	}
}

func TestHTTPClientRedirectToBlockedHostRejected(t *testing.T) {
	t.Parallel()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "secret")
	}))
	defer target.Close()

	// 指向 169.254.169.254（云元数据服务）：端口 443 在白名单内，
	// 但地址必须被拦截，且**不允许**真正建连。
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://169.254.169.254/latest/meta-data/", http.StatusFound)
	}))
	defer redirector.Close()

	g := newTestGuard(t, portOf(mustURL(t, redirector.URL)), true)
	client := g.HTTPClient(HTTPOptions{})

	_, err := doGet(context.Background(), client, redirector.URL)
	if err == nil {
		t.Fatal("重定向到链路本地地址应当被拒绝")
	}
	if !strings.Contains(err.Error(), "私有或保留网段") {
		t.Fatalf("期望私网拦截错误，实际 %v", err)
	}
}

func TestDialTLSCompletesHandshake(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	u := mustURL(t, srv.URL)
	g := newTestGuard(t, portOf(u), true)

	conn, err := g.DialTLS(context.Background(), TLSOptions{
		Host:               u.Hostname(),
		Port:               portOf(u),
		ServerName:         u.Hostname(),
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("TLS 握手失败：%v", err)
	}
	defer func() { _ = conn.Conn.Close() }()

	if conn.Conn.ConnectionState().Version == 0 {
		t.Fatal("握手完成但未协商出版本")
	}
	if len(conn.Conn.ConnectionState().PeerCertificates) == 0 {
		t.Fatal("未取到对端证书")
	}
}

func TestDialTLSRejectsBlockedAddress(t *testing.T) {
	t.Parallel()

	g := newTestGuard(t, 0, false)
	_, err := g.DialTLS(context.Background(), TLSOptions{
		Host: "127.0.0.1",
		Port: 443,
	})
	if err == nil {
		t.Fatal("环回地址应当被拒绝")
	}
}

func TestDialTLSRejectsPortOutsideWhitelist(t *testing.T) {
	t.Parallel()

	g := newTestGuard(t, 0, true)
	_, err := g.DialTLS(context.Background(), TLSOptions{
		Host: "127.0.0.1",
		Port: 8080,
	})
	var ngErr *Error
	if !errors.As(err, &ngErr) || ngErr.Reason != ReasonPort {
		t.Fatalf("期望 port-not-allowed，实际 %v", err)
	}
}

func TestDialContextRejectsPortOutsideWhitelist(t *testing.T) {
	t.Parallel()

	g := newTestGuard(t, 0, true)
	if _, err := g.DialContext(context.Background(), "tcp", "127.0.0.1:3306"); err == nil {
		t.Fatal("非白名单端口应当被拒绝")
	}
}

func TestDialIPRejectsBlockedAddress(t *testing.T) {
	t.Parallel()

	g := newTestGuard(t, 0, false)
	_, err := g.DialIP(context.Background(), "tcp", netip.MustParseAddr("10.0.0.1"), 80)
	if err == nil {
		t.Fatal("私有地址应当被拒绝")
	}
}

func TestPortOfDefaultsByScheme(t *testing.T) {
	t.Parallel()

	cases := map[string]int{
		"https://example.com/":     443,
		"http://example.com/":      80,
		"https://example.com:8443": 8443,
		"ftp://example.com/":       -1,
	}
	for raw, want := range cases {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("解析 %s 失败：%v", raw, err)
		}
		if got := portOf(u); got != want {
			t.Errorf("portOf(%s) = %d，期望 %d", raw, got, want)
		}
	}
}

func TestTLSConfigHasInsecureSkipVerifyForProbing(t *testing.T) {
	t.Parallel()

	// 该断言锁定「探测场景下允许跳过链验证」这一决定，
	// 避免后续有人误把它改成强制校验而让工具无法展示无效证书。
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	u := mustURL(t, srv.URL)
	g := newTestGuard(t, portOf(u), true)

	// 不跳过校验时必须失败：httptest 的证书是自签的。
	_, err := g.DialTLS(context.Background(), TLSOptions{
		Host:       u.Hostname(),
		Port:       portOf(u),
		ServerName: u.Hostname(),
	})
	if err == nil {
		t.Fatal("自签证书在强制校验下应当握手失败")
	}
	var recordErr *tls.CertificateVerificationError
	if !errors.As(err, &recordErr) {
		t.Logf("握手失败原因（非致命，仅提示）：%v", err)
	}
}

// doGet 用带 ctx 的请求发起 GET（仓库禁止无 ctx 的便捷方法）。
func doGet(ctx context.Context, client *http.Client, raw string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

// mustURL 解析 URL 并在失败时终止测试。
func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("解析 %s 失败：%v", raw, err)
	}
	return u
}
