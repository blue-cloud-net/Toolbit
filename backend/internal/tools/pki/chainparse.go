package pki

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ChainParseRequest 是证书链解析工具的请求。
type ChainParseRequest struct {
	// Bundle 是证书链材料（可含多张证书）。
	Bundle PemBundle `json:"bundle"`
}

// ChainParseResult 是证书链解析工具的响应。
type ChainParseResult struct {
	// Count 是证书总数。
	Count int `json:"count"`
	// Chain 是排好序的证书（叶 → 根）。
	Chain []ChainCertOutput `json:"chain"`
	// Ordered 报告输入是否已经是有序的。
	Ordered bool `json:"ordered"`
	// MissingRoot 报告链尾是否缺少根证书。
	MissingRoot bool `json:"missingRoot"`
	// SelfSigned 报告链尾证书是否自签名。
	SelfSigned bool `json:"selfSigned"`
}

// ChainCertOutput 是链上一张证书的输出形态。
type ChainCertOutput struct {
	// Pem 是证书的 PEM 文本。
	Pem string `json:"pem"`
	// Info 是证书元信息。
	Info CertInfoOutput `json:"info"`
	// Role 是证书在链中的角色（leaf / intermediate / root / unlinked）。
	Role string `json:"role"`
	// Link 是该证书与其上一级的连接诊断。
	Link ChainLinkOutput `json:"link"`
	// Validity 是该证书的有效性判定。
	Validity ValidityOutput `json:"validity"`
	// FingerprintSha256 是该证书的 SHA-256 指纹。
	FingerprintSha256 string `json:"fingerprintSha256"`
	// PublicKey 是该证书公钥的展开视图。
	PublicKey *PublicKeyDetailOutput `json:"publicKey"`
}

// Validate 做字段级校验。
func (r *ChainParseRequest) Validate() error {
	return r.Bundle.Validate("bundle")
}

// ChainParse 解析并排序证书链。
//
// 排序依据是 **AKID ↔ SKID**，退回 subject / issuer 文本；
// 同一层有多个候选时按输入顺序取第一个（稳定排序）。
func ChainParse(_ context.Context, req *ChainParseRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Bundle.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	certs, err := pkisvc.ParseCertificates(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(certs)

	info, err := pkisvc.OrderChain(certs)
	if err != nil {
		return tools.Result{}, err
	}

	result := &ChainParseResult{
		Count:       info.Count,
		Chain:       chainOutput(info.Chain),
		Ordered:     info.Ordered,
		MissingRoot: info.MissingRoot,
		SelfSigned:  info.SelfSigned,
	}

	return tools.Result{Data: result}, nil
}

// chainOutput 把链转成输出形态。
func chainOutput(chain []pkisvc.ChainCertificate) []ChainCertOutput {
	out := make([]ChainCertOutput, 0, len(chain))
	for _, item := range chain {
		out = append(out, ChainCertOutput{
			Pem:               item.PEM,
			Info:              certInfoOutput(item.Info),
			Role:              item.Role,
			Link:              chainLinkOutput(item.Link),
			Validity:          validityOutput(item.Validity),
			FingerprintSha256: item.FingerprintSHA256,
			PublicKey:         publicKeyDetailOutput(item.PublicKey),
		})
	}
	return out
}

// closeAll 关闭一组证书句柄（忽略错误）。
func closeAll(certs []*pkisvc.Certificate) {
	for _, cert := range certs {
		_ = cert.Close()
	}
}
