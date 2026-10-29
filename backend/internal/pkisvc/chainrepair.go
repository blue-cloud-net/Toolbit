package pkisvc

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 链修复的结论。
const (
	// RepairReachedRoot 表示已到达自签名根或受信任根。
	RepairReachedRoot = "reached_root"
	// RepairMissingIntermediate 表示缺少中间证书，链未闭合。
	RepairMissingIntermediate = "missing_intermediate"
	// RepairAlreadyComplete 表示输入本身已是完整链。
	RepairAlreadyComplete = "already_complete"
)

// RepairResult 是证书链修复的结果。
type RepairResult struct {
	// ChainPEM 是补全后的链（叶 → 根）。
	ChainPEM []string `json:"chainPem"`
	// AddedCount 是补充的证书数。
	AddedCount int `json:"addedCount"`
	// AddedFroms 是每一张补充证书的来源（intermediates / trustedRoots）。
	AddedFroms []string `json:"addedFroms"`
	// ReachedRoot 报告是否到达了自签名根或受信任根。
	ReachedRoot bool `json:"reachedRoot"`
	// Note 是结论枚举，取值见 RepairReachedRoot 等常量。
	Note string `json:"note"`
}

// RepairChain 用给定的证书集合补全一条证书链。
//
// pool 是**候选池**：intermediates 与 roots 都参与匹配；
// 匹配顺序是「先 intermediates，后 trustedRoots」。
//
// 匹配依据：先用 **AKID ↔ SKID**，再用 subject / issuer 文本。
// 本函数**不联网**（出网能力在 v0.7.0 之后，见 §10）。
func RepairChain(
	leaf *Certificate,
	intermediates []*Certificate,
	roots []*Certificate,
) (*RepairResult, error) {
	if leaf == nil || leaf.cert == nil {
		return nil, toolerrInvalidCert("certificate already closed")
	}

	total := len(intermediates) + len(roots)
	if total > MaxChainLength {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "intermediates",
			"limit":  MaxChainLength,
			"actual": total,
		})
	}

	leafPEM, err := leaf.MarshalPEM()
	if err != nil {
		return nil, err
	}

	result := &RepairResult{
		ChainPEM:   []string{string(leafPEM)},
		AddedFroms: []string{},
	}

	used := make(map[*Certificate]bool)
	current := leaf
	currentInfo, err := Describe(current, zeroTime, nil)
	if err != nil {
		return nil, err
	}

	for depth := 0; depth < MaxChainLength; depth++ {
		candidate, source, ok := findRepairCandidate(current, currentInfo, intermediates, roots, used)
		if !ok {
			break
		}

		pemBytes, err := candidate.MarshalPEM()
		if err != nil {
			return nil, err
		}
		result.ChainPEM = append(result.ChainPEM, string(pemBytes))
		result.AddedFroms = append(result.AddedFroms, source)
		result.AddedCount++
		used[candidate] = true

		current = candidate
		currentInfo, err = Describe(candidate, zeroTime, nil)
		if err != nil {
			return nil, err
		}

		if currentInfo.SelfSigned {
			result.ReachedRoot = true
			break
		}
	}

	switch {
	case result.ReachedRoot:
		result.Note = RepairReachedRoot
	case result.AddedCount == 0:
		result.Note = RepairAlreadyComplete
	default:
		result.Note = RepairMissingIntermediate
	}

	return result, nil
}

// findRepairCandidate 在候选池中找出当前证书的签发者。
//
// 先在 intermediates 里找，再在 trustedRoots 里找；
// 池内按输入顺序取第一个匹配项，保证结果稳定。
func findRepairCandidate(
	current *Certificate,
	currentInfo *CertInfo,
	intermediates []*Certificate,
	roots []*Certificate,
	used map[*Certificate]bool,
) (*Certificate, string, bool) {
	for _, pool := range []struct {
		certs  []*Certificate
		source string
	}{
		{intermediates, "intermediates"},
		{roots, "trustedRoots"},
	} {
		for _, candidate := range pool.certs {
			if candidate == nil || candidate.cert == nil || used[candidate] {
				continue
			}
			info, err := Describe(candidate, zeroTime, nil)
			if err != nil {
				continue
			}
			if !issuerMatches(currentInfo, info) {
				continue
			}
			return candidate, pool.source, true
		}
	}
	return nil, "", false
}

// issuerMatches 判断 candidate 是否是 current 的签发者。
//
// 先看 AKID ↔ SKID：两者都有且不相等时**直接否**（keyid 明确不匹配）；
// 否则退回 subject / issuer 文本匹配。
func issuerMatches(current, candidate *CertInfo) bool {
	if candidate.SubjectText != current.IssuerText {
		return false
	}
	if current.AuthorityKeyID != "" && candidate.SubjectKeyID != "" {
		return current.AuthorityKeyID == candidate.SubjectKeyID
	}
	return true
}
