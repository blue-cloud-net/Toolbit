package pkisvc

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// MaxChainLength 是证书链的最大张数。
const MaxChainLength = 64

// ChainCertificate 是链中的一张证书。
type ChainCertificate struct {
	// PEM 是证书的 PEM 文本。
	PEM string `json:"pem"`
	// Info 是证书的元信息。
	Info *CertInfo `json:"info"`

	// cert 是底层证书句柄。
	//
	// 它**不**对外暴露：上层拿不到 tongsuo-go 的类型，
	// 只能用本包提供的 SplitChain 拆链。
	cert *Certificate
}

// ChainInfo 是证书链的解析结果。
type ChainInfo struct {
	// Count 是证书总数。
	Count int `json:"count"`
	// Chain 是排好序的证书（叶 → 根）；无法接入链的证书排在末尾。
	Chain []ChainCertificate `json:"chain"`
	// Ordered 报告输入是否已经是有序的。
	//
	// 它只表示「输入需要重排」，**不**表示排序失败：
	// 交叉签名的中间 CA 会让顺序不唯一，此时仍按输入顺序稳定输出。
	Ordered bool `json:"ordered"`
	// MissingRoot 报告链尾是否缺少根证书。
	MissingRoot bool `json:"missingRoot"`
	// SelfSigned 报告链尾证书是否自签名。
	SelfSigned bool `json:"selfSigned"`
}

// chainNode 是排序用的证书索引项。
type chainNode struct {
	// cert 是证书句柄。
	cert *Certificate
	// subject 是主体文本。
	subject string
	// issuer 是签发者文本。
	issuer string
	// skid 是主体密钥标识符（大写十六进制）。
	skid string
	// akid 是签发者密钥标识符（大写十六进制）。
	akid string
	// index 是输入序号，用于保持顺序稳定。
	index int
}

// OrderChain 把散乱的证书排成叶 → 根的链。
//
// 排序依据：先用 **AKID ↔ SKID** 匹配，再用 subject / issuer 文本匹配；
// 同一层有多个候选时按**输入顺序**取第一个（稳定排序）。
// 这样交叉签名的中间 CA 不会导致随机顺序。
//
// 无法接入链的证书（与主链无父子关系）排在链尾，顺序保持输入顺序。
func OrderChain(certs []*Certificate) (*ChainInfo, error) {
	if len(certs) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "bundle.data"})
	}
	if len(certs) > MaxChainLength {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "bundle.data",
			"limit":  MaxChainLength,
			"actual": len(certs),
		})
	}

	nodes := make([]*chainNode, 0, len(certs))
	for i, cert := range certs {
		info, err := Describe(cert, zeroTime, nil)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &chainNode{
			cert:    cert,
			subject: info.SubjectText,
			issuer:  info.IssuerText,
			skid:    info.SubjectKeyID,
			akid:    info.AuthorityKeyID,
			index:   i,
		})
	}

	leaf := pickLeaf(nodes)
	ordered := buildChain(nodes, leaf)

	result := &ChainInfo{
		Count: len(nodes),
		Chain: make([]ChainCertificate, 0, len(nodes)),
	}

	for _, node := range ordered {
		info, err := Describe(node.cert, zeroTime, nil)
		if err != nil {
			return nil, err
		}
		pemBytes, err := node.cert.MarshalPEM()
		if err != nil {
			return nil, err
		}
		result.Chain = append(result.Chain, ChainCertificate{
			PEM:  string(pemBytes),
			Info: info,
			cert: node.cert,
		})
	}

	result.Ordered = sameOrder(nodes, ordered)
	result.SelfSigned = len(ordered) > 0 && ordered[len(ordered)-1].subject == ordered[len(ordered)-1].issuer
	result.MissingRoot = !result.SelfSigned

	return result, nil
}

// pickLeaf 选出链的起点（叶证书）。
//
// 判据：主体没有被任何其它证书作为签发者引用。
// 有多个候选时取输入顺序的第一个；一个都没有（成环）时取第一个。
func pickLeaf(nodes []*chainNode) *chainNode {
	referenced := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		referenced[node.issuer] = true
	}
	for _, node := range nodes {
		if !referenced[node.subject] {
			return node
		}
	}
	return nodes[0]
}

// buildChain 从叶子出发逐级向上拼链。
//
// 返回的切片包含**全部**证书：主链在前，无法接入的排在末尾。
func buildChain(nodes []*chainNode, leaf *chainNode) []*chainNode {
	visited := make(map[int]bool, len(nodes))
	out := make([]*chainNode, 0, len(nodes))

	current := leaf
	for current != nil && !visited[current.index] {
		visited[current.index] = true
		out = append(out, current)
		current = findParent(nodes, current, visited)
	}

	for _, node := range nodes {
		if !visited[node.index] {
			visited[node.index] = true
			out = append(out, node)
		}
	}
	return out
}

// findParent 找出当前证书的上一级（签发者）。
//
// 先按 AKID ↔ SKID 精确匹配；匹配不到再退回 subject / issuer 文本匹配。
// 多个候选时取输入顺序的第一个。
func findParent(nodes []*chainNode, current *chainNode, visited map[int]bool) *chainNode {
	var byKeyID *chainNode
	var byName *chainNode

	for _, node := range nodes {
		if visited[node.index] || node == current {
			continue
		}
		if node.subject != current.issuer {
			continue
		}
		if current.akid != "" && node.skid != "" {
			if current.akid == node.skid {
				if byKeyID == nil {
					byKeyID = node
				}
				continue
			}
			// keyid 明确不匹配：不是这一张。
			continue
		}
		if byName == nil {
			byName = node
		}
	}

	if byKeyID != nil {
		return byKeyID
	}
	return byName
}

// SplitChain 把排序后的链拆成「叶证书 + 中间证书」。
//
// 只有**主链**（从叶到根的那一段）会被拆出；
// 无法接入链的证书（排在末尾的那些）不参与验证：
// 把无关证书塞进链里只会让验证结果变得不可解释。
//
// 链为空时返回两个 nil。
func SplitChain(info *ChainInfo) (*Certificate, []*Certificate) {
	if info == nil || len(info.Chain) == 0 {
		return nil, nil
	}
	leaf := info.Chain[0].cert
	intermediates := make([]*Certificate, 0, len(info.Chain)-1)
	for _, item := range info.Chain[1:] {
		if item.cert == nil {
			continue
		}
		intermediates = append(intermediates, item.cert)
	}
	return leaf, intermediates
}

// sameOrder 判断结果顺序是否与输入顺序一致。
func sameOrder(nodes []*chainNode, ordered []*chainNode) bool {
	if len(nodes) != len(ordered) {
		return false
	}
	for i, node := range ordered {
		if node.index != i {
			return false
		}
	}
	return true
}
