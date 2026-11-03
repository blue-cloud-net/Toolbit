package pki

import (
	"bytes"
	"embed"
	"encoding/csv"
	"strings"
	"sync"
)

// oidDatasetFS 内嵌 OID 数据集。
//
// 数据来自**官方注册表的离线快照**（IANA SMI / PKIX / S-MIME、IETF RFC、
// ITU-T、GM/T），见 dataset/README.md。
//
// 运行时**不联网**：查表全在本地完成。
//
//go:embed dataset/*.csv
var oidDatasetFS embed.FS

// OIDDatasetVersion 是 OID 数据集的快照版本。
//
// 内容调整后**必须**更新，否则响应里的版本号会误导用户。
const OIDDatasetVersion = "2026-11"

// OID 数据集缓存。
var (
	oidOnce sync.Once
	// oidNodes 按 OID 索引的节点表。
	oidNodes map[string]*OIDNode
	// oidChildren 按父 OID 索引的子节点（已按弧号升序）。
	oidChildren map[string][]*OIDNode
	// oidRoots 是三个根节点（0 / 1 / 2）。
	oidRoots []*OIDNode
)

// loadOIDDataset 解析并缓存 OID 数据集。
//
// 解析只做一次：CSV 是内嵌的静态数据，重复解析纯属浪费。
func loadOIDDataset() {
	oidOnce.Do(func() {
		oidNodes = make(map[string]*OIDNode)
		oidChildren = make(map[string][]*OIDNode)

		for _, row := range readCSVRows(oidDatasetFS, "dataset/oid_nodes.csv") {
			node := &OIDNode{
				OID:         csvValue(row, 0),
				Name:        csvValue(row, 1),
				Description: csvValue(row, 2),
				Kind:        csvValue(row, 3),
				Status:      csvValue(row, 4),
				Source:      csvValue(row, 5),
				SourceURL:   csvValue(row, 6),
				Aliases:     splitOIDList(csvValue(row, 7)),
			}
			if node.OID == "" {
				continue
			}
			node.Arc = oidArc(node.OID)
			oidNodes[node.OID] = node
		}

		// 中文表覆盖英文名与英文说明；缺失时保持英文。
		for _, row := range readCSVRows(oidDatasetFS, "dataset/oid_zh.csv") {
			oid := csvValue(row, 0)
			node, ok := oidNodes[oid]
			if !ok {
				continue
			}
			if nameZh := csvValue(row, 1); nameZh != "" {
				node.NameZh = nameZh
			}
			if descZh := csvValue(row, 2); descZh != "" {
				node.DescriptionZh = descZh
			}
		}

		// 建立父子索引；子节点按弧号升序，保证树顺序稳定。
		for _, node := range oidNodes {
			parent := oidParent(node.OID)
			if parent == "" {
				oidRoots = append(oidRoots, node)
				continue
			}
			oidChildren[parent] = append(oidChildren[parent], node)
		}
		for parent := range oidChildren {
			sortOIDChildren(oidChildren[parent])
		}
		sortOIDChildren(oidRoots)

		// 回填直接子节点数量：前端用它决定是否显示展开箭头。
		for parent, children := range oidChildren {
			if node, ok := oidNodes[parent]; ok {
				node.ChildCount = len(children)
			}
		}
	})
}

// readCSVRows 读取内嵌 CSV 并跳过表头。
//
// 内嵌文件在编译期就已确定存在；读取失败按空结果处理，
// 让查询返回空结果而不是让整个进程崩溃。
func readCSVRows(fs embed.FS, name string) [][]string {
	raw, err := fs.ReadFile(name)
	if err != nil {
		return nil
	}
	reader := csv.NewReader(bytes.NewReader(raw))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return nil
	}
	return records[1:]
}

// csvValue 读取 CSV 记录的第 i 列；越界时返回空串。
func csvValue(row []string, i int) string {
	if i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

// splitOIDList 按分号拆分列表项，忽略空白项。
func splitOIDList(raw string) []string {
	out := []string{}
	for _, part := range strings.Split(raw, ";") {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

// oidParent 取父 OID；顶层节点返回空串。
func oidParent(oid string) string {
	idx := strings.LastIndex(oid, ".")
	if idx < 0 {
		return ""
	}
	return oid[:idx]
}

// oidArc 取末弧号；单段 OID（如根节点 `0`）返回自身。
func oidArc(oid string) string {
	idx := strings.LastIndex(oid, ".")
	if idx < 0 {
		return oid
	}
	return oid[idx+1:]
}

// sortOIDChildren 按弧号**数值**升序排序（不是字符串序）。
//
// 字符串序会把 `10` 排到 `2` 前面；树里必须按数值序，
// 否则 1.3.6.1.5.5.7.3 的子节点顺序会错乱。
func sortOIDChildren(nodes []*OIDNode) {
	// 插入排序：子节点数量很小（最大的分支也不到 200 个），
	// 且只在初始化时跑一次。
	for i := 1; i < len(nodes); i++ {
		item := nodes[i]
		j := i - 1
		for j >= 0 && oidArcLess(item.OID, nodes[j].OID) {
			nodes[j+1] = nodes[j]
			j--
		}
		nodes[j+1] = item
	}
}

// oidArcLess 比较两个 OID 的末弧：数值小的在前。
func oidArcLess(a, b string) bool {
	an, aok := oidArcNumber(a)
	bn, bok := oidArcNumber(b)
	if aok && bok {
		if an != bn {
			return an < bn
		}
		return a < b
	}
	return oidArc(a) < oidArc(b)
}

// oidArcNumber 解析末弧为整数。
func oidArcNumber(oid string) (int, bool) {
	arc := oidArc(oid)
	if arc == "" {
		return 0, false
	}
	n := 0
	for _, ch := range arc {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		n = n*10 + int(ch-'0')
		if n > 1<<30 {
			return 0, false
		}
	}
	return n, true
}

// oidRoots 返回三个根节点（0 / 1 / 2）。
func oidRootNodes() []*OIDNode {
	loadOIDDataset()
	out := make([]*OIDNode, 0, len(oidRoots))
	for _, node := range oidRoots {
		out = append(out, node.clone())
	}
	return out
}

// oidLookupNode 按 OID 取节点副本；未收录时返回 nil。
func oidLookupNode(oid string) *OIDNode {
	loadOIDDataset()
	node, ok := oidNodes[oid]
	if !ok {
		return nil
	}
	return node.clone()
}

// oidChildNodes 取某节点的直接子节点副本（按弧号升序）。
func oidChildNodes(oid string) []*OIDNode {
	loadOIDDataset()
	children := oidChildren[oid]
	out := make([]*OIDNode, 0, len(children))
	for _, child := range children {
		out = append(out, child.clone())
	}
	return out
}

// oidAncestors 取某节点的祖先链（从根到父节点）。
func oidAncestors(oid string) []*OIDNode {
	loadOIDDataset()
	out := []*OIDNode{}
	for parent := oidParent(oid); parent != ""; parent = oidParent(parent) {
		node, ok := oidNodes[parent]
		if !ok {
			break
		}
		out = append([]*OIDNode{node.clone()}, out...)
	}
	return out
}

// oidNearestAncestor 返回最长的**已收录**祖先 OID。
//
// 未收录时返回空串。
func oidNearestAncestor(oid string) string {
	loadOIDDataset()
	for parent := oidParent(oid); parent != ""; parent = oidParent(parent) {
		if _, ok := oidNodes[parent]; ok {
			return parent
		}
	}
	return ""
}

// oidMissingSuffix 返回未收录 OID 相对最近祖先缺失的弧后缀。
//
// 例如 `1.3.6.1.5.5.7.3.9999` 相对 `1.3.6.1.5.5.7.3` 的缺失后缀是 `.9999`。
func oidMissingSuffix(oid, nearest string) string {
	if nearest == "" {
		return oid
	}
	if len(oid) <= len(nearest) {
		return ""
	}
	return oid[len(nearest):]
}

// oidAllNodes 返回全部节点的副本（供搜索使用）。
func oidAllNodes() []*OIDNode {
	loadOIDDataset()
	out := make([]*OIDNode, 0, len(oidNodes))
	for _, node := range oidNodes {
		out = append(out, node.clone())
	}
	return out
}
