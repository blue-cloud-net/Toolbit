package pki

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ChainRepairRequest 是证书链修复工具的请求。
type ChainRepairRequest struct {
	// Cert 是叶证书。
	Cert PemBundle `json:"cert"`
	// Intermediates 是可用于补链的中间证书。
	//
	// 本字段是 v0.5.0 新增的**可选**字段：没有它，
	// 「只从输入集合补链」就没有候选可用。
	Intermediates []PemBundle `json:"intermediates"`
	// TrustedRoots 是受信任根证书；同样参与补链。
	TrustedRoots []PemBundle `json:"trustedRoots"`
}

// ChainRepairResult 是证书链修复工具的响应。
type ChainRepairResult struct {
	// ChainPem 是补全后的链（叶 → 根）。
	ChainPem []string `json:"chainPem"`
	// AddedCount 是补充的证书数。
	AddedCount int `json:"addedCount"`
	// AddedFroms 是每一张补充证书的来源（intermediates / trustedRoots）。
	AddedFroms []string `json:"addedFroms"`
	// ReachedRoot 报告是否到达了自签名根或受信任根。
	ReachedRoot bool `json:"reachedRoot"`
	// Note 是结论枚举。
	Note string `json:"note"`
}

// Validate 做字段级校验。
func (r *ChainRepairRequest) Validate() error {
	if err := r.Cert.Validate("cert"); err != nil {
		return err
	}
	for i := range r.Intermediates {
		if err := r.Intermediates[i].Validate("intermediates"); err != nil {
			return err
		}
	}
	for i := range r.TrustedRoots {
		if err := r.TrustedRoots[i].Validate("trustedRoots"); err != nil {
			return err
		}
	}
	return nil
}

// ChainRepair 用给定集合补全证书链。
//
// 本工具**纯本地**：只从 `intermediates` 与 `trustedRoots` 里挑，
// **不**联网下载（出网能力在 v0.7.0 之后）。
func ChainRepair(_ context.Context, req *ChainRepairRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Cert.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	leaf, err := pkisvc.LoadCertificate(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = leaf.Close() }()

	intermediates, err := loadCertificates(req.Intermediates)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(intermediates)

	roots, err := loadCertificates(req.TrustedRoots)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(roots)

	out, err := pkisvc.RepairChain(leaf, intermediates, roots)
	if err != nil {
		return tools.Result{}, err
	}

	result := &ChainRepairResult{
		ChainPem:    out.ChainPEM,
		AddedCount:  out.AddedCount,
		AddedFroms:  out.AddedFroms,
		ReachedRoot: out.ReachedRoot,
		Note:        out.Note,
	}

	return tools.Result{Data: result}, nil
}
