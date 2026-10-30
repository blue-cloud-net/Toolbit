package network

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 通用上限与默认值。
const (
	// defaultTimeoutMs 是单次出网动作的默认超时。
	defaultTimeoutMs = 5000
	// maxTimeoutMs 是允许的最大超时。
	maxTimeoutMs = 60000
	// maxHostLength 是主机名长度上限。
	maxHostLength = 253
	// maxServers 是一次请求允许的自定义服务器数量。
	maxServers = 8
	// maxParallelQueries 是多服务器查询的并发上限。
	//
	// 顺序查询在 8 台 × 5s 时会超出工具 20s 的预算；并发取 4
	// 既能把总时长压到 10s 以内，也不会同时压满上游。
	maxParallelQueries = 4
)

// normalizeHost 规范化并校验主机名。
//
// 允许 IP 字面量与域名；**不**允许带协议前缀、路径或端口 —— 端口另有字段。
func normalizeHost(host string) (string, error) {
	h := strings.TrimSpace(host)
	if h == "" {
		return "", &toolerr.Error{Code: toolerr.CodeMissingField, Params: map[string]any{"field": "host"}}
	}
	if len(h) > maxHostLength {
		return "", invalidHost()
	}
	if strings.ContainsAny(h, "/:@?#\\ \t\r\n") {
		return "", invalidHost()
	}
	h = strings.ToLower(strings.TrimSuffix(h, "."))
	if h == "" {
		return "", invalidHost()
	}
	return h, nil
}

// invalidHost 构造主机名非法错误。
func invalidHost() *toolerr.Error {
	return toolerr.New(toolerr.CodeInvalidDomain)
}

// normalizePort 校验端口并套用默认值。
func normalizePort(port, def int) (int, error) {
	if port == 0 {
		return def, nil
	}
	if port < 1 || port > 65535 {
		return 0, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "port",
			"expected": []string{"1..65535"},
			"actual":   port,
		})
	}
	// 端口**白名单**由 netguard 在出网时裁决（返回 SSRF_BLOCKED）；
	// 这里只保证取值是合法端口，避免把策略判断混进字段校验。
	return port, nil
}

// resolveTimeout 把毫秒超时转成 time.Duration。
func resolveTimeout(ms, def int) (time.Duration, error) {
	if ms == 0 {
		ms = def
	}
	if ms < 1 || ms > maxTimeoutMs {
		return 0, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "timeoutMs",
			"expected": fmt.Sprintf("1..%d", maxTimeoutMs),
			"actual":   ms,
		})
	}
	return time.Duration(ms) * time.Millisecond, nil
}

// mapNetError 把出网错误映射为工具错误码。
//
// 映射顺序很重要：策略拒绝优先于网络错误，超时优先于通用不可达。
func mapNetError(err error) *toolerr.Error {
	if err == nil {
		return nil
	}
	var toolErr *toolerr.Error
	if errors.As(err, &toolErr) {
		return toolErr
	}
	if errors.Is(err, netguard.ErrDisabled) {
		return toolerr.New(toolerr.CodeNetworkDisabled)
	}

	var ngErr *netguard.Error
	if errors.As(err, &ngErr) {
		switch ngErr.Reason {
		case netguard.ReasonAddress, netguard.ReasonPort,
			netguard.ReasonScheme, netguard.ReasonRedirect:
			return toolerr.New(toolerr.CodeSSRFBlocked)
		default:
			return invalidHost()
		}
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return toolerr.New(toolerr.CodeUpstreamTimeout)
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return toolerr.New(toolerr.CodeUpstreamTimeout)
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return toolerr.New(toolerr.CodeDNSError)
	}
	var recordErr *tls.CertificateVerificationError
	if errors.As(err, &recordErr) {
		return toolerr.New(toolerr.CodeTLSError)
	}
	var alertErr tls.AlertError
	if errors.As(err, &alertErr) {
		return toolerr.New(toolerr.CodeTLSError)
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "tls:"), strings.Contains(msg, "handshake failure"):
		return toolerr.New(toolerr.CodeTLSError)
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "no route to host"),
		strings.Contains(msg, "network is unreachable"),
		strings.Contains(msg, "i/o timeout"):
		return toolerr.New(toolerr.CodeHostUnreachable)
	default:
		return toolerr.New(toolerr.CodeUpstreamError)
	}
}

// toolerrTooMany 构造条目数超限的错误。
func toolerrTooMany(field string, actual, max int) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
		"field":  field,
		"max":    max,
		"actual": actual,
	})
}

// containsString 报告列表中是否包含目标值。
func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// requireNetwork 在出网开关关闭时提前返回，避免无谓的解析。
func requireNetwork(g *netguard.Guard) error {
	if g == nil || !g.Enabled() {
		return toolerr.New(toolerr.CodeNetworkDisabled)
	}
	return nil
}
