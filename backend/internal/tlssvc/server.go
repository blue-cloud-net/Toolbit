package tlssvc

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	tsTLS "github.com/blue-cloud-net/tongsuo-go/tls"
	tsX509 "github.com/blue-cloud-net/tongsuo-go/x509"
)

// serveConnTimeout 是测试服务端单连接的存活上限。
const serveConnTimeout = 500 * time.Millisecond

// ServerOptions 是本地 TLS 服务端的参数。
//
// 它只服务于**测试与自检**（回环上验证探测逻辑），不对外提供业务能力。
type ServerOptions struct {
	// CertPEM 是服务端证书的 PEM。
	CertPEM []byte
	// KeyPEM 是服务端私钥的 PEM。
	KeyPEM []byte
	// MinVersion / MaxVersion 是协议版本范围；0 表示不限制。
	MinVersion uint16
	MaxVersion uint16
	// CipherSuites 是套件白名单；为空表示用库默认。
	CipherSuites []string
	// NextProtos 是 ALPN 协议列表。
	NextProtos []string
}

// Listen 在已有监听器上包一层 TLS。
//
// 返回的监听器每次 Accept 都完成握手；握手失败会关闭该连接并返回错误，
// 监听器本身仍可继续使用。
func Listen(ln net.Listener, opts ServerOptions) (net.Listener, error) {
	if ln == nil {
		return nil, fmt.Errorf("tlssvc: nil listener")
	}
	cfg, release, err := serverConfig(opts)
	if err != nil {
		return nil, err
	}
	wrapped, err := tsTLS.NewListener(cfg, ln)
	if err != nil {
		release()
		return nil, fmt.Errorf("tlssvc: new listener: %w", err)
	}
	return &ownedListener{Listener: wrapped, release: release}, nil
}

// Serve 在监听器上循环接受连接、完成握手后立即关闭。
//
// 单条连接的握手失败（例如探测方故意用了服务端不支持的协议版本）
// **不会**终止服务；只有监听器被关闭时本函数才返回。
func Serve(ln net.Listener, opts ServerOptions) error {
	tlsLn, err := Listen(ln, opts)
	if err != nil {
		return err
	}
	defer func() { _ = tlsLn.Close() }()

	for {
		conn, acceptErr := tlsLn.Accept()
		if acceptErr != nil {
			if errors.Is(acceptErr, net.ErrClosed) {
				return nil
			}
			// 握手失败：连接已被上游关闭，继续服务下一条。
			continue
		}
		go func(c net.Conn) {
			defer func() { _ = c.Close() }()
			// 探测只做握手、不发应用层数据：给一个短期限即可。
			_ = c.SetDeadline(time.Now().Add(serveConnTimeout))
			buf := make([]byte, 1)
			_, _ = c.Read(buf)
		}(conn)
	}
}

// serverConfig 构造服务端配置，并返回释放凭据的函数。
func serverConfig(opts ServerOptions) (*tsTLS.Config, func(), error) {
	cert, err := tsX509.LoadCertificatePEM(opts.CertPEM)
	if err != nil {
		return nil, func() {}, fmt.Errorf("tlssvc: load server certificate: %w", err)
	}
	key, err := asym.LoadPrivateKeyPEM(opts.KeyPEM)
	if err != nil {
		_ = cert.Close()
		return nil, func() {}, fmt.Errorf("tlssvc: load server key: %w", err)
	}

	cfg := &tsTLS.Config{
		Protocol:     tsTLS.ProtocolTLS,
		Cert:         cert,
		Key:          key,
		MinVersion:   opts.MinVersion,
		MaxVersion:   opts.MaxVersion,
		CipherSuites: opts.CipherSuites,
		NextProtos:   opts.NextProtos,
		Verification: tsTLS.VerifyNone,
	}
	release := func() {
		_ = cert.Close()
		_ = asym.Close(key)
	}
	return cfg, release, nil
}

// ownedListener 关闭时一并释放证书与私钥句柄。
type ownedListener struct {
	// Listener 是上游监听器。
	net.Listener
	// release 释放凭据句柄。
	release func()
}

// Close 关闭监听器并释放凭据句柄。幂等。
func (l *ownedListener) Close() error {
	err := l.Listener.Close()
	if l.release != nil {
		l.release()
		l.release = nil
	}
	return err
}
