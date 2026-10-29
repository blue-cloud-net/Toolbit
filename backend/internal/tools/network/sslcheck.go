package network

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// TLS 探测阶段名。
const (
	// PhaseResolve 表示域名解析阶段。
	PhaseResolve = "resolve"
	// PhaseConnect 表示 TCP 建连阶段。
	PhaseConnect = "connect"
	// PhaseHandshake 表示 TLS 握手阶段。
	PhaseHandshake = "handshake"
	// PhaseChain 表示证书链整理与验证阶段。
	PhaseChain = "chain"
)

// TLSTotalPhases 是探测的阶段总数。
const TLSTotalPhases = 4

// 证书链深度默认值与上限。
const (
	// defaultChainDepth 是默认链深度。
	defaultChainDepth = 5
	// maxChainDepth 是允许的最大链深度。
	maxChainDepth = 10
)

// tlsVersionNames 是版本名到协议版本号的映射。
var tlsVersionNames = map[string]uint16{
	"TLS1.0": tls.VersionTLS10,
	"TLS1.1": tls.VersionTLS11,
	"TLS1.2": tls.VersionTLS12,
	"TLS1.3": tls.VersionTLS13,
}

// TLSVersions 返回受支持的协议版本名（顺序稳定）。
func TLSVersions() []string {
	return []string{"TLS1.3", "TLS1.2", "TLS1.1", "TLS1.0"}
}

// TLSProbeOptions 是 TLS 探测参数。
type TLSProbeOptions struct {
	// Host 是目标主机。
	Host string
	// Port 是目标端口。
	Port int
	// SNI 是服务器名字指示；为空时用 Host。
	SNI string
	// Depth 是链深度上限。
	Depth int
	// Versions 是允许的协议版本名。
	Versions []string
	// Ciphers 是允许的密码套件名。
	Ciphers []string
	// TimeoutMs 是总超时。
	TimeoutMs int
}

// TLSProgress 是一次进度通知。
type TLSProgress struct {
	// Phase 是阶段名。
	Phase string
	// Done 是已完成阶段数。
	Done int
	// Total 是阶段总数。
	Total int
}

// TLSChainCert 是链上的一张证书。
type TLSChainCert struct {
	// Pem 是证书的 PEM 文本。
	Pem string `json:"pem"`
	// Info 是证书的结构化信息；解析失败时为空对象。
	Info *pkisvc.CertInfo `json:"info"`
	// Trusted 报告该证书是否来自本地信任库。
	Trusted bool `json:"trusted"`
}

// TLSProbeResult 是一次 TLS 探测的结果。
type TLSProbeResult struct {
	// Host 是目标主机。
	Host string `json:"host"`
	// Port 是目标端口。
	Port int `json:"port"`
	// Addr 是实际连接的 IP。
	Addr string `json:"addr"`
	// Protocol 是协议族名。
	Protocol string `json:"protocol"`
	// NegotiatedVersion 是协商出的协议版本。
	NegotiatedVersion string `json:"negotiatedVersion"`
	// CipherSuite 是协商出的密码套件。
	CipherSuite string `json:"cipherSuite"`
	// Certificates 是服务端发送的证书链。
	Certificates []TLSChainCert `json:"certificates"`
	// ChainValid 报告链是否通过验证（含主机名匹配）。
	ChainValid bool `json:"chainValid"`
	// ChainErrors 是链验证的错误列表。
	ChainErrors []string `json:"chainErrors"`
	// Verified 报告是否用**系统信任根**完成了验证。
	Verified bool `json:"verified"`
	// ImplicitTrust 报告是否退化为「链内自洽」判定。
	ImplicitTrust bool `json:"implicitTrust"`
	// OcspStapled 报告服务端是否携带 OCSP 装订响应。
	OcspStapled bool `json:"ocspStapled"`
	// Grade 是本次探测的综合评级（A+ / A / B / C / T）。
	Grade string `json:"grade"`
	// Checks 是逐项检查结论。
	Checks []TLSCheck `json:"checks"`
}

// ValidateTLSProbe 校验探测参数。
func ValidateTLSProbe(opts *TLSProbeOptions) error {
	if _, err := normalizeHost(opts.Host); err != nil {
		return err
	}
	if _, err := normalizePort(opts.Port, 443); err != nil {
		return err
	}
	if opts.Depth < 0 || opts.Depth > maxChainDepth {
		return tools.InvalidField("depth", []string{"0..10"}, opts.Depth)
	}
	for _, v := range opts.Versions {
		if _, ok := tlsVersionNames[strings.ToUpper(strings.TrimSpace(v))]; !ok {
			return tools.InvalidField("versions", TLSVersions(), v)
		}
	}
	if len(opts.Ciphers) > 0 {
		known := cipherSuiteIDs()
		for _, c := range opts.Ciphers {
			if _, ok := known[strings.ToUpper(strings.TrimSpace(c))]; !ok {
				return tools.InvalidField("ciphers", "受支持的密码套件名", c)
			}
		}
	}
	return nil
}

// TLSProbe 完成一次 TLS 探测。
//
// onProgress 可为 nil；非 nil 时按阶段回调，用于 WebSocket 流式反馈。
func TLSProbe(
	ctx context.Context, g *netguard.Guard, opts TLSProbeOptions,
	onProgress func(TLSProgress),
) (*TLSProbeResult, error) {
	if err := requireNetwork(g); err != nil {
		return nil, err
	}
	host, err := normalizeHost(opts.Host)
	if err != nil {
		return nil, err
	}
	port, err := normalizePort(opts.Port, 443)
	if err != nil {
		return nil, err
	}
	timeout, err := resolveTimeout(opts.TimeoutMs, 30000)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	report := func(phase string, done int) {
		if onProgress != nil {
			onProgress(TLSProgress{Phase: phase, Done: done, Total: TLSTotalPhases})
		}
	}

	report(PhaseResolve, 0)

	minVersion, maxVersion := versionRange(opts.Versions)
	suites, err := cipherSuiteIDsOf(opts.Ciphers)
	if err != nil {
		return nil, err
	}

	report(PhaseConnect, 1)
	conn, err := g.DialTLS(ctx, netguard.TLSOptions{
		Host:               host,
		Port:               port,
		ServerName:         strings.TrimSpace(opts.SNI),
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
		CipherSuites:       suites,
		InsecureSkipVerify: true,
		HandshakeTimeout:   timeout,
	})
	if err != nil {
		return nil, mapNetError(err)
	}
	defer func() { _ = conn.Conn.Close() }()

	report(PhaseHandshake, 2)

	state := conn.Conn.ConnectionState()
	result := &TLSProbeResult{
		Host:              host,
		Port:              port,
		Addr:              conn.Addr.String(),
		Protocol:          "TLS",
		NegotiatedVersion: tls.VersionName(state.Version),
		CipherSuite:       tls.CipherSuiteName(state.CipherSuite),
		Certificates:      []TLSChainCert{},
		ChainErrors:       []string{},
		OcspStapled:       len(state.OCSPResponse) > 0,
	}

	depth := opts.Depth
	if depth <= 0 {
		depth = defaultChainDepth
	}
	peers := state.PeerCertificates
	if len(peers) > depth {
		peers = peers[:depth]
	}
	for _, peer := range peers {
		result.Certificates = append(result.Certificates, chainCertOf(peer.Raw))
	}

	report(PhaseChain, 3)
	rawChain := make([][]byte, 0, len(peers))
	for _, peer := range peers {
		rawChain = append(rawChain, peer.Raw)
	}
	result.ChainValid, result.Verified, result.ImplicitTrust, result.ChainErrors =
		verifyPeerChain(rawChain, host)
	result.Grade, result.Checks = evaluateTLSProbe(result)
	report(PhaseChain, TLSTotalPhases)
	return result, nil
}

// chainCertOf 把 DER 证书转成对外结构。
func chainCertOf(der []byte) TLSChainCert {
	buf := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	cert := TLSChainCert{Pem: string(buf)}
	parsed, err := pkisvc.LoadCertificate(der, false)
	if err != nil {
		return cert
	}
	defer func() { _ = parsed.Close() }()
	if info, infoErr := pkisvc.Describe(parsed, time.Time{}, nil); infoErr == nil {
		cert.Info = info
	}
	return cert
}

// verifyPeerChain 验证服务端证书链。
//
// 优先用本地信任库真实验签；信任库不可用时退化为链内自洽判定，
// 并置 ImplicitTrust（上游未暴露默认信任库，见 I-17）。
func verifyPeerChain(peers [][]byte, host string) (bool, bool, bool, []string) {
	if len(peers) == 0 {
		return false, false, false, []string{"no certificate presented"}
	}
	leaf, err := pkisvc.LoadCertificate(peers[0], false)
	if err != nil {
		return false, false, false, []string{"invalid leaf certificate"}
	}
	defer func() { _ = leaf.Close() }()

	intermediates := make([]*pkisvc.Certificate, 0, len(peers))
	for _, der := range peers[1:] {
		cert, loadErr := pkisvc.LoadCertificate(der, false)
		if loadErr != nil {
			continue
		}
		intermediates = append(intermediates, cert)
	}
	defer closeCertificates(intermediates)

	roots, closeRoots := systemRoots()
	defer closeRoots()

	implicit := len(roots) == 0
	verify, verifyErr := pkisvc.VerifyChain(leaf, intermediates, roots, time.Time{}, host)
	if verifyErr != nil {
		return false, false, implicit, []string{verifyErr.Error()}
	}

	errs := make([]string, 0, len(verify.Errors))
	for _, item := range verify.Errors {
		errs = append(errs, item.Code)
	}
	if verify.HostnameError != "" {
		errs = append(errs, verify.HostnameError)
	}
	return verify.Valid, !implicit, implicit, errs
}

// closeCertificates 释放证书句柄。
func closeCertificates(certs []*pkisvc.Certificate) {
	for _, cert := range certs {
		if cert != nil {
			_ = cert.Close()
		}
	}
}

// systemRoots 读取本地信任库。
//
// 只做**读取与解析**，不做任何密码学运算。
// 返回的句柄由调用方负责释放；返回空切片表示信任库不可用。
func systemRoots() ([]*pkisvc.Certificate, func()) {
	noop := func() {}
	paths := []string{
		"/etc/ssl/certs/ca-certificates.crt",
		"/etc/pki/tls/certs/ca-bundle.crt",
		"/etc/ssl/ca-bundle.pem",
		"/etc/ssl/cert.pem",
	}
	var data []byte
	for _, path := range paths {
		if buf, err := os.ReadFile(path); err == nil && len(buf) > 0 {
			data = buf
			break
		}
	}
	if len(data) == 0 {
		data = readCertDir("/etc/ssl/certs")
	}
	if len(data) == 0 {
		return nil, noop
	}

	parsed, err := pkisvc.ParseCertificates(data, true)
	if err == nil && len(parsed) > 0 {
		return parsed, func() { closeCertificates(parsed) }
	}
	// 单文件解析失败时逐张尝试，容忍个别条目的格式差异。
	certs := make([]*pkisvc.Certificate, 0, 128)
	for _, block := range splitPEMBlocks(data) {
		cert, loadErr := pkisvc.LoadCertificate(block, true)
		if loadErr != nil {
			continue
		}
		certs = append(certs, cert)
	}
	if len(certs) == 0 {
		return nil, noop
	}
	return certs, func() { closeCertificates(certs) }
}

// readCertDir 把目录下的 PEM 证书拼接起来。
func readCertDir(dir string) []byte {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []byte
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".pem") && !strings.HasSuffix(name, ".crt") {
			continue
		}
		buf, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			continue
		}
		out = append(out, buf...)
		out = append(out, '\n')
	}
	return out
}

// splitPEMBlocks 拆出全部 PEM 块。
func splitPEMBlocks(data []byte) [][]byte {
	var out [][]byte
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return out
		}
		if block.Type == "CERTIFICATE" {
			out = append(out, pem.EncodeToMemory(block))
		}
	}
}

// versionRange 把版本名集合折叠成闭区间。
//
// 未指定时返回零值，交给 Go 的默认值（TLS1.2 起）。
func versionRange(names []string) (uint16, uint16) {
	if len(names) == 0 {
		return 0, 0
	}
	minVersion := uint16(0xffff)
	maxVersion := uint16(0)
	for _, name := range names {
		v, ok := tlsVersionNames[strings.ToUpper(strings.TrimSpace(name))]
		if !ok {
			continue
		}
		if v < minVersion {
			minVersion = v
		}
		if v > maxVersion {
			maxVersion = v
		}
	}
	if maxVersion == 0 {
		return 0, 0
	}
	return minVersion, maxVersion
}

// cipherSuiteIDs 返回全部可配置密码套件的名字到 ID 映射。
func cipherSuiteIDs() map[string]uint16 {
	out := make(map[string]uint16)
	for _, suite := range tls.CipherSuites() {
		out[strings.ToUpper(suite.Name)] = suite.ID
	}
	for _, suite := range tls.InsecureCipherSuites() {
		out[strings.ToUpper(suite.Name)] = suite.ID
	}
	return out
}

// cipherSuiteIDsOf 把套件名解析为 ID 列表。
//
// 为空表示使用默认套件。注意 TLS 1.3 的套件无法通过配置裁剪，
// 这一点会在 API 文档中说明。
func cipherSuiteIDsOf(names []string) ([]uint16, error) {
	if len(names) == 0 {
		return nil, nil
	}
	known := cipherSuiteIDs()
	out := make([]uint16, 0, len(names))
	for _, name := range names {
		id, ok := known[strings.ToUpper(strings.TrimSpace(name))]
		if !ok {
			return nil, toolerr.New(toolerr.CodeUnsupportedAlgorithm)
		}
		out = append(out, id)
	}
	return out, nil
}
