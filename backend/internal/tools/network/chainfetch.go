package network

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ChainFetchRequest 是证书链获取工具的请求。
type ChainFetchRequest struct {
	// Host 是目标主机。
	Host string `json:"host"`
	// Port 是目标端口，默认 443。
	Port int `json:"port"`
	// ServerName 是 SNI；为空时用 Host。
	ServerName string `json:"serverName"`
	// TimeoutMs 是超时。
	TimeoutMs int `json:"timeoutMs"`
}

// ChainSummary 是证书链的摘要。
type ChainSummary struct {
	// Count 是链上证书数量。
	Count int `json:"count"`
	// Subject 是叶证书主体。
	Subject string `json:"subject"`
	// Issuer 是叶证书签发者。
	Issuer string `json:"issuer"`
	// NotBefore 是叶证书生效时间（RFC3339）。
	NotBefore string `json:"notBefore"`
	// NotAfter 是叶证书失效时间（RFC3339）。
	NotAfter string `json:"notAfter"`
	// SelfSigned 报告叶证书是否自签名。
	SelfSigned bool `json:"selfSigned"`
	// ChainValid 报告链是否通过验证。
	ChainValid bool `json:"chainValid"`
	// ChainErrors 是链验证错误。
	ChainErrors []string `json:"chainErrors"`
	// Verified 报告是否用系统信任根完成验证。
	Verified bool `json:"verified"`
}

// ChainFetchResult 是证书链获取工具的响应。
type ChainFetchResult struct {
	// Host 是目标主机。
	Host string `json:"host"`
	// Port 是目标端口。
	Port int `json:"port"`
	// ChainPem 是整条链的 PEM（叶证书在前）。
	ChainPem string `json:"chainPem"`
	// Chain 是逐张证书的明细。
	Chain []TLSChainCert `json:"chain"`
	// ChainSummary 是链摘要。
	ChainSummary ChainSummary `json:"chainSummary"`
}

// Validate 做字段级校验。
func (r *ChainFetchRequest) Validate() error {
	return ValidateTLSProbe(&TLSProbeOptions{
		Host: r.Host, Port: r.Port, SNI: r.ServerName, TimeoutMs: r.TimeoutMs,
	})
}

// ChainFetch 从目标主机抓取证书链。
func ChainFetch(ctx context.Context, g *netguard.Guard, req *ChainFetchRequest) (tools.Result, error) {
	probe, err := TLSProbe(ctx, g, TLSProbeOptions{
		Host:      req.Host,
		Port:      req.Port,
		SNI:       req.ServerName,
		TimeoutMs: req.TimeoutMs,
	}, nil)
	if err != nil {
		return tools.Result{}, err
	}

	var builder strings.Builder
	for _, cert := range probe.Certificates {
		builder.WriteString(strings.TrimRight(cert.Pem, "\n"))
		builder.WriteString("\n")
	}

	summary := ChainSummary{
		Count:       len(probe.Certificates),
		ChainValid:  probe.ChainValid,
		ChainErrors: probe.ChainErrors,
		Verified:    probe.Verified,
	}
	if len(probe.Certificates) > 0 && probe.Certificates[0].Info != nil {
		leaf := probe.Certificates[0].Info
		summary.Subject = leaf.SubjectText
		summary.Issuer = leaf.IssuerText
		summary.NotBefore = leaf.NotBefore.UTC().Format("2006-01-02T15:04:05Z")
		summary.NotAfter = leaf.NotAfter.UTC().Format("2006-01-02T15:04:05Z")
		summary.SelfSigned = leaf.SelfSigned
	}

	result := ChainFetchResult{
		Host:         probe.Host,
		Port:         probe.Port,
		ChainPem:     builder.String(),
		Chain:        probe.Certificates,
		ChainSummary: summary,
	}
	warnings := chainWarnings(probe, summary)
	return tools.Result{Data: result, Warnings: warnings}, nil
}
