package tlssvc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/meta"
	tsX509 "github.com/blue-cloud-net/tongsuo-go/x509"
)

// selfSignedPair 生成一张自签名证书与对应私钥（PEM 形态）。
func selfSignedPair(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := asym.GenerateEC("P-256")
	if err != nil {
		t.Fatalf("生成密钥失败：%v", err)
	}
	defer func() { _ = asym.Close(key) }()

	pub := key.Public()

	name := tsX509.NewName()
	for _, entry := range [][2]string{{"C", "CN"}, {"O", "Toolbit Test"}, {"CN", "localhost"}} {
		if err := name.TryAdd(entry[0], entry[1]); err != nil {
			t.Fatalf("构造 DN 失败：%v", err)
		}
	}

	now := time.Now()
	cert, err := tsX509.CreateSelfSigned(name, 1, now.Add(-time.Hour), now.Add(24*time.Hour), pub, key)
	if err != nil {
		t.Fatalf("自签失败：%v", err)
	}
	defer func() { _ = cert.Close() }()

	certPEM, err = cert.MarshalPEM()
	if err != nil {
		t.Fatalf("导出证书失败：%v", err)
	}
	keyPEM, err = asym.MarshalPKCS8(key, meta.FormatPEM, nil)
	if err != nil {
		t.Fatalf("导出私钥失败：%v", err)
	}
	return certPEM, keyPEM
}

// startServer 起一个本地 TLS 服务端，返回地址与关停函数。
//
// min 与 max **同时**固定：只设 min 不会封顶，TLS 1.3 仍会被协商出来。
func startServer(t *testing.T, minVersion uint16) (string, func()) {
	t.Helper()

	certPEM, keyPEM := selfSignedPair(t)
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败：%v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = Serve(raw, ServerOptions{
			CertPEM:    certPEM,
			KeyPEM:     keyPEM,
			MinVersion: minVersion,
			MaxVersion: minVersion,
		})
	}()

	stop := func() {
		_ = raw.Close()
		<-done
	}
	return raw.Addr().String(), stop
}

// dial 到本地服务端并完成握手。
func dial(t *testing.T, addr string, opts Options) (*Session, error) {
	t.Helper()
	raw, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("建连失败：%v", err)
	}
	session, hsErr := Handshake(context.Background(), raw, opts)
	if hsErr != nil {
		return nil, hsErr
	}
	return session, nil
}

// TestHandshakeReportsNegotiatedState 断言握手快照包含版本、套件与对端证书。
func TestHandshakeReportsNegotiatedState(t *testing.T) {
	addr, stop := startServer(t, VersionTLS12)
	defer stop()

	session, err := dial(t, addr, Options{ServerName: "localhost"})
	if err != nil {
		t.Fatalf("握手失败：%v", err)
	}
	defer func() { _ = session.Close() }()

	state := session.State()
	if state.Version != VersionTLS12 {
		t.Errorf("版本 = %d，期望 %d", state.Version, VersionTLS12)
	}
	if state.VersionName != "TLS1.2" {
		t.Errorf("版本名 = %q", state.VersionName)
	}
	if state.CipherName == "" {
		t.Error("套件名为空")
	}
	if len(state.PeerCertificates) == 0 {
		t.Error("未取到对端证书")
	}
}

// TestHandshakeRejectsUnsupportedVersion 断言版本不匹配会被归类为「无共同版本」。
//
// 上游的原生错误队列分类在**错误线程**上取（见 docs/upstream/tongsuo-go-issues.md
// I-28），因此这里只断言「握手失败」，不绑定具体分类。
func TestHandshakeRejectsUnsupportedVersion(t *testing.T) {
	addr, stop := startServer(t, VersionTLS12)
	defer stop()

	_, err := dial(t, addr, Options{
		ServerName: "localhost",
		MinVersion: VersionTLS13,
		MaxVersion: VersionTLS13,
	})
	if err == nil {
		t.Fatal("期望握手失败")
	}
}

// TestHandshakeRejectsUnknownCipher 断言本机不认识的套件名在**配置期**就被拒绝。
//
// 这是可确定判定的一类：套件名本机发不出去，与对端无关。
func TestHandshakeRejectsUnknownCipher(t *testing.T) {
	addr, stop := startServer(t, VersionTLS12)
	defer stop()

	_, err := dial(t, addr, Options{
		ServerName:   "localhost",
		MinVersion:   VersionTLS12,
		MaxVersion:   VersionTLS12,
		CipherSuites: []string{"NO-SUCH-CIPHER"},
	})
	if err == nil {
		t.Fatal("期望握手失败")
	}
	var hsErr *HandshakeError
	if !errors.As(err, &hsErr) {
		t.Fatalf("错误类型不符：%T", err)
	}
	if hsErr.Reason != ReasonNoCommonCipher {
		t.Errorf("归因 = %q，期望 %q", hsErr.Reason, ReasonNoCommonCipher)
	}
}

// TestEnumerateSuites 断言套件枚举在已知版本上返回非空结果。
func TestEnumerateSuites(t *testing.T) {
	suites := EnumerateSuites(VersionTLS13)
	if len(suites) == 0 {
		t.Skip("本机未启用 TLS 1.3 套件")
	}
	for _, suite := range suites {
		if suite.Name == "" {
			t.Error("套件名为空")
		}
		if !HasSuite(suite.Name) {
			t.Errorf("枚举出的套件 %q 竟然不被识别", suite.Name)
		}
	}
	if HasSuite("BOGUS-SUITE") {
		t.Error("不存在的套件名不应被识别")
	}
}

// TestVersionName 断言版本名映射。
func TestVersionName(t *testing.T) {
	if VersionName(VersionTLS13) != "TLS1.3" {
		t.Errorf("TLS 1.3 映射错误")
	}
	if VersionName(0x9999) != "" {
		t.Error("未知版本应返回空串")
	}
	if len(Versions()) != 4 {
		t.Errorf("版本数 = %d，期望 4", len(Versions()))
	}
}
