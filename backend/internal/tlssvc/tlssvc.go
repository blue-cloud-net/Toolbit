// Package tlssvc 是**传输层**（TLS / NTLS）能力层。
//
// 职责边界：
//
//   - 本包是 `tongsuo-go` 的第三个、也是唯一的传输层封装点；
//     其余包（除 `cryptosvc` / `pkisvc`）一律不得直接 import `tongsuo-go/tls`。
//   - 本包**不拨号**：只接受调用方已经建立并校验过的 `net.Conn`。
//     出网解析、私网校验与钉 IP 一律由 `netguard` 负责，SSRF 防护只有一处实现。
//
// 对外暴露的结构体只含 DER / 字符串 / 枚举，不出现上游类型。
package tlssvc

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/blue-cloud-net/tongsuo-go/meta"
	tsTLS "github.com/blue-cloud-net/tongsuo-go/tls"
)

// 协议版本（IANA wire 值）。
const (
	// VersionTLS10 是 TLS 1.0。
	VersionTLS10 uint16 = 0x0301
	// VersionTLS11 是 TLS 1.1。
	VersionTLS11 uint16 = 0x0302
	// VersionTLS12 是 TLS 1.2。
	VersionTLS12 uint16 = 0x0303
	// VersionTLS13 是 TLS 1.3。
	VersionTLS13 uint16 = 0x0304

	// VersionSSL30 是 SSL 3.0 的 wire 值。
	//
	// 它**仅用于对外说明**：Tongsuo 基于 OpenSSL 3.x 兼容层，协议本身已被移除，
	// 无法作为探测目标。见 docs/tools/ssl-check-refactor.md §7.1。
	VersionSSL30 uint16 = 0x0300
)

// defaultHandshakeTimeout 是单次握手的默认超时。
const defaultHandshakeTimeout = 5 * time.Second

// Version 是一个协议版本的展示形态。
type Version struct {
	// Wire 是 IANA wire 值。
	Wire uint16
	// Name 是短名，如 `TLS1.2`。
	Name string
}

// Versions 返回可探测的协议版本（由高到低，顺序稳定）。
func Versions() []Version {
	return []Version{
		{Wire: VersionTLS13, Name: "TLS1.3"},
		{Wire: VersionTLS12, Name: "TLS1.2"},
		{Wire: VersionTLS11, Name: "TLS1.1"},
		{Wire: VersionTLS10, Name: "TLS1.0"},
	}
}

// VersionName 把 wire 值转成短名；未知版本返回空串。
func VersionName(wire uint16) string {
	for _, v := range Versions() {
		if v.Wire == wire {
			return v.Name
		}
	}
	return ""
}

// Suite 是一个密码套件的展示形态。
type Suite struct {
	// Name 是 OpenSSL / IANA 名称。
	Name string
	// ID 是 16 位 wire 值。
	ID uint16
}

// EnumerateSuites 枚举指定协议版本下本机可用的全部密码套件。
//
// 返回值顺序由上游决定，调用方不应依赖；未知版本返回空切片。
func EnumerateSuites(wire uint16) []Suite {
	infos := tsTLS.CipherSuites(wire)
	out := make([]Suite, 0, len(infos))
	for _, info := range infos {
		out = append(out, Suite{Name: info.Name, ID: info.ID})
	}
	return out
}

// HasSuite 报告某个套件名是否被本机识别。
func HasSuite(name string) bool {
	_, err := tsTLS.CipherSuiteByName(name)
	return err == nil
}

// AvailableGroups 返回本机构建支持的密钥交换组名。
//
// 组名是 OpenSSL 的命名组（如 `X25519`、`P-256`、`P-384`）。
// 旧版 Tongsuo 可能不支持组枚举，此时返回空切片 —— 调用方据此把
// 依赖具体组名的画像降级为「不裁剪组」。
func AvailableGroups() []string {
	groups := meta.TLSGroups()
	out := make([]string, 0, len(groups))
	out = append(out, groups...)
	return out
}

// HasGroup 报告某个组名是否被本机支持。
func HasGroup(name string) bool {
	for _, group := range AvailableGroups() {
		if group == name {
			return true
		}
	}
	return false
}

// Options 是一次握手探测的参数。
type Options struct {
	// ServerName 是 SNI；为空时不发送 SNI（由上游从对端地址推导并丢弃 IP 字面量）。
	ServerName string
	// MinVersion 是协议版本下限；0 表示不限制。
	MinVersion uint16
	// MaxVersion 是协议版本上限；0 表示不限制。
	MaxVersion uint16
	// CipherSuites 是套件白名单（OpenSSL 名）；为空表示用库默认。
	//
	// TLS 1.3 的套件用 `TLS_` 前缀名，走 `set_ciphersuites`；
	// 经典名走 `set_cipher_list`。两者可混用。
	CipherSuites []string
	// Curves 是密钥交换组白名单；为空表示用库默认。
	Curves []string
	// SignatureAlgorithms 是签名算法白名单；为空表示用库默认。
	SignatureAlgorithms []string
	// NextProtos 是 ALPN 协议列表。
	NextProtos []string
	// HandshakeTimeout 是握手超时；0 表示用默认值。
	HandshakeTimeout time.Duration
}

// State 是一次**成功**握手的观测快照。
//
// 它不持有任何原生句柄，可以安全地长期保存与跨 goroutine 传递。
type State struct {
	// Version 是协商出的协议版本 wire 值。
	Version uint16
	// VersionName 是版本短名（如 `TLS1.2`）。
	VersionName string
	// CipherSuite 是协商出的套件 wire 值。
	CipherSuite uint16
	// CipherName 是协商出的套件名。
	CipherName string
	// ALPN 是 ALPN 协商结果；未协商时为空。
	ALPN string
	// KeyExchangeGroup 是密钥交换组名；上游未报告时为空。
	KeyExchangeGroup string
	// PeerSignatureAlgorithm 是对端握手签名算法；上游未报告时为空。
	PeerSignatureAlgorithm string
	// PeerCertificates 是对端证书链的 DER（叶 → 根）。
	PeerCertificates [][]byte
}

// FailureReason 是握手失败的语义分类。
type FailureReason string

const (
	// ReasonNoCommonVersion 表示协议版本无交集。
	ReasonNoCommonVersion FailureReason = "noCommonVersion"
	// ReasonNoCommonCipher 表示密码套件无交集。
	ReasonNoCommonCipher FailureReason = "noCommonCipher"
	// ReasonRejected 表示对端主动拒绝（alert），原因见 Detail。
	ReasonRejected FailureReason = "handshakeFailure"
	// ReasonNetwork 表示建连或传输层故障。
	ReasonNetwork FailureReason = "network"
	// ReasonTimeout 表示握手超时。
	ReasonTimeout FailureReason = "timeout"
)

// HandshakeError 是一个可分类的握手失败。
type HandshakeError struct {
	// Reason 是语义分类。
	Reason FailureReason
	// Detail 是上游给出的原因文本，**仅用于展示**，不得用于分支判断。
	Detail string
	// Err 是底层错误。
	Err error
}

// Error 实现 error。
func (e *HandshakeError) Error() string {
	if e == nil {
		return ""
	}
	if e.Detail != "" {
		return string(e.Reason) + ": " + e.Detail
	}
	if e.Err != nil {
		return string(e.Reason) + ": " + e.Err.Error()
	}
	return string(e.Reason)
}

// Unwrap 返回底层错误。
func (e *HandshakeError) Unwrap() error { return e.Err }

// Session 是一条已完成握手的连接。
type Session struct {
	// conn 是上游连接。
	conn *tsTLS.Conn
	// state 是握手快照。
	state State
}

// State 返回握手快照。
func (s *Session) State() State { return s.state }

// Close 关闭连接。幂等。
func (s *Session) Close() error {
	if s == nil || s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

// Handshake 在**已建立的连接**上完成一次 TLS 握手。
//
// 成功时返回的 Session 接管 raw 的所有权，调用方只需 Close 它。
// 失败时 raw 已被关闭，Session 为 nil。
//
// 对端证书**不在握手期校验**（`VerifyNone`）：探测工具必须能读到无效证书的细节，
// 链验证由 `pkisvc` 在拿到证书后另行完成。
func Handshake(ctx context.Context, raw net.Conn, opts Options) (*Session, error) {
	if raw == nil {
		return nil, &HandshakeError{Reason: ReasonNetwork, Err: errors.New("nil connection")}
	}

	timeout := opts.HandshakeTimeout
	if timeout <= 0 {
		timeout = defaultHandshakeTimeout
	}

	cfg := &tsTLS.Config{
		Protocol:            tsTLS.ProtocolTLS,
		MinVersion:          opts.MinVersion,
		MaxVersion:          opts.MaxVersion,
		CipherSuites:        opts.CipherSuites,
		CurvePreferences:    opts.Curves,
		SignatureAlgorithms: opts.SignatureAlgorithms,
		NextProtos:          opts.NextProtos,
		ServerName:          opts.ServerName,
		// 显式策略：只做协议层握手，不评估信任链。
		Verification: tsTLS.VerifyNone,
	}

	conn, err := tsTLS.Client(raw, cfg)
	if err != nil {
		_ = raw.Close()
		return nil, classifyConfig(err)
	}

	hsCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if hsErr := conn.HandshakeContext(hsCtx); hsErr != nil {
		_ = conn.Close()
		return nil, classify(hsErr)
	}

	return &Session{conn: conn, state: snapshot(conn.ConnectionState())}, nil
}

// snapshot 把上游状态转成本包的快照。
func snapshot(in tsTLS.ConnectionState) State {
	state := State{
		Version:                in.Version,
		VersionName:            VersionName(in.Version),
		CipherSuite:            in.CipherSuite,
		CipherName:             in.CipherName,
		ALPN:                   in.NegotiatedProtocol,
		KeyExchangeGroup:       in.NegotiatedGroup,
		PeerSignatureAlgorithm: in.PeerSignatureAlgorithm,
		PeerCertificates:       in.PeerCertificates,
	}
	if state.VersionName == "" {
		// 上游在本机未识别的版本上用 CipherName 兜底不了，保留 wire 值即可。
		state.VersionName = in.CipherName
	}
	return state
}

// classifyConfig 把**配置期**失败（尚未发包）映射为语义分类。
//
// 配置期失败意味着「本机根本发不出这个 ClientHello」：
// 套件名不被本机识别，或协议版本范围非法。这类失败与对端的拒绝无关。
func classifyConfig(err error) *HandshakeError {
	out := &HandshakeError{Detail: err.Error(), Err: err}
	switch {
	case errors.Is(err, tsTLS.ErrNoSharedCipher):
		out.Reason = ReasonNoCommonCipher
		return out
	case errors.Is(err, tsTLS.ErrVersionNotSupported):
		out.Reason = ReasonNoCommonVersion
		return out
	default:
		out.Reason = ReasonRejected
		return out
	}
}

// classify 把上游错误映射为语义分类。
func classify(err error) *HandshakeError {
	out := &HandshakeError{Detail: err.Error(), Err: err}
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		out.Reason = ReasonTimeout
		return out
	case errors.Is(err, tsTLS.ErrVersionNotSupported):
		out.Reason = ReasonNoCommonVersion
		return out
	case errors.Is(err, tsTLS.ErrNoSharedCipher):
		out.Reason = ReasonNoCommonCipher
		return out
	case errors.Is(err, tsTLS.ErrNetwork):
		out.Reason = ReasonNetwork
		return out
	}

	var hsErr *tsTLS.HandshakeError
	if errors.As(err, &hsErr) {
		switch hsErr.Kind {
		case tsTLS.HandshakeErrorVersion:
			out.Reason = ReasonNoCommonVersion
			return out
		case tsTLS.HandshakeErrorCipher:
			out.Reason = ReasonNoCommonCipher
			return out
		case tsTLS.HandshakeErrorNetwork:
			out.Reason = ReasonNetwork
			return out
		}
	}

	out.Reason = ReasonRejected
	return out
}
