package pki

import (
	"context"
	"sort"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// OID 查询的请求上限。
const (
	// OIDMaxLength 是 OID 字符串的最大长度。
	OIDMaxLength = 256
	// OIDMaxArcs 是 OID 最大的弧数。
	OIDMaxArcs = 128
	// OIDMaxQueryLength 是搜索词的最大长度。
	OIDMaxQueryLength = 128
	// OIDDefaultLimit 是子节点与搜索结果的默认条数。
	OIDDefaultLimit = 50
	// OIDMaxLimit 是单次返回的最大条数。
	OIDMaxLimit = 200
)

// OID 查询模式。
const (
	// OIDModeRoot 是根节点浏览模式。
	OIDModeRoot = "root"
	// OIDModeNode 是精确节点查询模式。
	OIDModeNode = "node"
	// OIDModeSearch 是搜索模式。
	OIDModeSearch = "search"
)

// OIDNode 是一个 OID 节点。
type OIDNode struct {
	// OID 是点分十进制表示的对象标识符，如 `1.3.6.1.5.5.7.3.36`。
	OID string `json:"oid"`
	// Arc 是末弧号，如 `36`。
	Arc string `json:"arc"`
	// Name 是官方英文名，如 `id-kp-documentSigning`。
	Name string `json:"name"`
	// NameZh 是中文名。
	NameZh string `json:"nameZh"`
	// Description 是官方英文说明。
	Description string `json:"description"`
	// DescriptionZh 是中文说明。
	DescriptionZh string `json:"descriptionZh"`
	// Kind 是节点种类：`root` / `branch` / `leaf`。
	Kind string `json:"kind"`
	// Status 是状态：`current` / `reserved` / `deprecated` / `obsolete`。
	Status string `json:"status"`
	// Source 是来源标注，如 `RFC 5280`。
	Source string `json:"source"`
	// SourceURL 是来源链接；无链接时为空串。
	SourceURL string `json:"sourceUrl"`
	// Aliases 是历史别名与常见写法。
	Aliases []string `json:"aliases"`
	// ChildCount 是直接子节点数量。
	ChildCount int `json:"childCount"`
}

// clone 复制节点。
//
// 数据集是**只读**的共享状态；返回副本可避免调用方意外改写缓存。
func (n *OIDNode) clone() *OIDNode {
	if n == nil {
		return nil
	}
	out := *n
	if len(n.Aliases) > 0 {
		out.Aliases = append([]string(nil), n.Aliases...)
	}
	return &out
}

// OIDLookupRequest 是 oid-lookup 工具的请求。
//
// `oid` 与 `query` **互斥**：同时给出返回 `INVALID_FIELD`。
// 两者都不给时返回**根节点**（`0` / `1` / `2`）。
type OIDLookupRequest struct {
	// OID 是待查询的对象标识符；空表示浏览根节点。
	OID string `json:"oid"`
	// Query 是搜索词；按 OID、名称与说明做子串匹配。
	Query string `json:"query"`
	// Offset 是分页偏移量。
	Offset int `json:"offset"`
	// Limit 是单页条数；0 表示使用默认值。
	Limit int `json:"limit"`
}

// OIDMatch 是一条搜索命中。
type OIDMatch struct {
	// Node 是命中的节点。
	Node *OIDNode `json:"node"`
	// Path 是该节点的祖先链（从根到父节点）。
	Path []*OIDNode `json:"path"`
	// Rank 是命中档位；数字越小越靠前。不序列化。
	Rank int `json:"-"`
}

// OIDLookupResult 是 oid-lookup 工具的响应。
type OIDLookupResult struct {
	// Mode 是查询模式：`root` / `node` / `search`。
	Mode string `json:"mode"`
	// OID 是规范化后的查询 OID；搜索模式为空串。
	OID string `json:"oid"`
	// Query 是规范化后的搜索词；非搜索模式为空串。
	Query string `json:"query"`
	// Matched 是否命中精确节点。
	Matched bool `json:"matched"`
	// Node 是命中的节点；未命中时为 nil。
	Node *OIDNode `json:"node"`
	// Path 是命中节点的祖先链（从根到父节点）。
	//
	// 未命中时是最近祖先的路径（含最近祖先自身），供前端铺开该分支。
	Path []*OIDNode `json:"path"`
	// PathChildren 是祖先链各层的子节点，供前端直接铺开树。
	//
	// 键是父节点的 OID，值是该层的子节点。
	PathChildren map[string][]*OIDNode `json:"pathChildren"`
	// Children 是命中节点的直接子节点（已分页）。
	Children []*OIDNode `json:"children"`
	// Total 是子节点总数；搜索模式下是命中总数。
	Total int `json:"total"`
	// Offset 是分页偏移量。
	Offset int `json:"offset"`
	// Limit 是单页条数。
	Limit int `json:"limit"`
	// HasMore 是否还有下一页。
	HasMore bool `json:"hasMore"`
	// Matches 是搜索结果；非搜索模式为空数组。
	Matches []OIDMatch `json:"matches"`
	// NearestOid 是未命中时的最近祖先 OID。
	NearestOid string `json:"nearestOid"`
	// MissingSuffix 是相对最近祖先缺失的弧后缀。
	MissingSuffix string `json:"missingSuffix"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// OIDLookup 查询 OID 树。
//
// 三种用法：
//   - 空请求：返回**根节点**（`0` / `1` / `2`）；
//   - `oid`：返回该节点的详情、祖先链与直接子节点；
//   - `query`：按 OID、名称与说明搜索。
//
// 数据集是**离线内置**的官方快照，**不**出网。
func OIDLookup(_ context.Context, req *OIDLookupRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	query := strings.TrimSpace(req.Query)
	oid := normalizeOID(req.OID)

	// 把分页参数收敛到有效区间；响应里的 limit 必须是**实际生效**的值，
	// 否则客户端看到 limit=0 会误以为没有分页。
	offset, limit := normalizeOidPaging(req.Offset, req.Limit)

	switch {
	case query != "":
		return oidSearch(query, offset, limit)
	case oid != "":
		return oidExactLookup(oid, offset, limit)
	default:
		return oidRootListing(offset, limit)
	}
}

// normalizeOidPaging 把分页参数收敛到有效区间。
func normalizeOidPaging(offset, limit int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = OIDDefaultLimit
	}
	if limit > OIDMaxLimit {
		limit = OIDMaxLimit
	}
	return offset, limit
}

// oidRootListing 返回根节点列表。
func oidRootListing(offset, limit int) (tools.Result, error) {
	roots := oidRootNodes()
	page, total, hasMore := paginate(len(roots), offset, limit)
	return tools.Result{Data: OIDLookupResult{
		Mode:         OIDModeRoot,
		Path:         []*OIDNode{},
		Children:     roots[page.start:page.end],
		Total:        total,
		Offset:       offset,
		Limit:        limit,
		HasMore:      hasMore,
		Matches:      []OIDMatch{},
		PathChildren: map[string][]*OIDNode{},
		DataVersion:  OIDDatasetVersion,
	}}, nil
}

// oidExactLookup 查询精确节点。
//
// OID 语法合法但未收录时**不是错误**：返回 `200` 与最近祖先，
// 由用户据此判断该 OID 是否属于已知体系。
func oidExactLookup(oid string, offset, limit int) (tools.Result, error) {
	result := OIDLookupResult{
		Mode:         OIDModeNode,
		OID:          oid,
		Path:         []*OIDNode{},
		Children:     []*OIDNode{},
		Matches:      []OIDMatch{},
		PathChildren: map[string][]*OIDNode{},
		Offset:       offset,
		Limit:        limit,
		DataVersion:  OIDDatasetVersion,
	}

	node := oidLookupNode(oid)
	if node == nil {
		nearest := oidNearestAncestor(oid)
		result.Matched = false
		result.NearestOid = nearest
		result.MissingSuffix = oidMissingSuffix(oid, nearest)
		if nearest != "" {
			// 未命中时也返回最近祖先的路径，前端据此把该分支铺进树里，
			// 用户可以直接点开祖先继续浏览，而不是只能看到一行提示。
			result.Path = append(oidAncestors(nearest), oidLookupNode(nearest))
			result.PathChildren = oidPathChildren(result.Path)
		}
		return tools.Result{Data: result}, nil
	}

	result.Matched = true
	result.Node = node
	result.Path = oidAncestors(oid)
	result.PathChildren = oidPathChildren(result.Path)

	children := oidChildNodes(oid)
	result.Total = len(children)
	page, _, hasMore := paginate(len(children), offset, limit)
	result.Children = children[page.start:page.end]
	result.HasMore = hasMore
	return tools.Result{Data: result}, nil
}

// oidPathChildren 收集祖先链各层的子节点。
//
// 前端据此一次性铺开从根到命中节点的整条路径，无需逐层请求。
func oidPathChildren(path []*OIDNode) map[string][]*OIDNode {
	out := make(map[string][]*OIDNode, len(path))
	for _, ancestor := range path {
		out[ancestor.OID] = oidChildNodes(ancestor.OID)
	}
	return out
}

// oidSearch 按关键词搜索 OID 树。
func oidSearch(query string, offset, limit int) (tools.Result, error) {
	term := strings.ToLower(query)
	nodes := oidAllNodes()

	scored := make([]OIDMatch, 0, len(nodes))
	for _, node := range nodes {
		rank, ok := matchOIDNode(node, term)
		if !ok {
			continue
		}
		scored = append(scored, OIDMatch{Node: node, Path: oidAncestors(node.OID), Rank: rank})
	}

	// 精确命中排前面；同档位保持 OID 字典序，保证结果稳定。
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Rank != scored[j].Rank {
			return scored[i].Rank < scored[j].Rank
		}
		return scored[i].Node.OID < scored[j].Node.OID
	})

	page, total, hasMore := paginate(len(scored), offset, limit)
	matches := make([]OIDMatch, 0, page.end-page.start)
	for _, item := range scored[page.start:page.end] {
		matches = append(matches, OIDMatch{Node: item.Node, Path: item.Path})
	}

	return tools.Result{Data: OIDLookupResult{
		Mode:         OIDModeSearch,
		Query:        query,
		Path:         []*OIDNode{},
		Children:     []*OIDNode{},
		Matches:      matches,
		Total:        total,
		Offset:       offset,
		Limit:        limit,
		HasMore:      hasMore,
		PathChildren: map[string][]*OIDNode{},
		DataVersion:  OIDDatasetVersion,
	}}, nil
}

// 命中档位：数字越小越靠前。
const (
	// oidRankExact 是 OID 或名称完全相同。
	oidRankExact = 0
	// oidRankAlias 是别名完全相同。
	oidRankAlias = 1
	// oidRankPrefix 是名称前缀相同。
	oidRankPrefix = 2
	// oidRankSubstring 是子串命中。
	oidRankSubstring = 3
)

// matchOIDNode 报告节点是否命中关键词，并给出命中档位。
//
// 匹配范围：OID、英文名、中文名、英文说明、中文说明与别名。
func matchOIDNode(node *OIDNode, term string) (int, bool) {
	oid := strings.ToLower(node.OID)
	if oid == term {
		return oidRankExact, true
	}
	for _, alias := range node.Aliases {
		if strings.ToLower(alias) == term {
			return oidRankAlias, true
		}
	}
	for _, candidate := range []string{node.Name, node.NameZh} {
		if candidate == "" {
			continue
		}
		lower := strings.ToLower(candidate)
		if lower == term {
			return oidRankExact, true
		}
		if strings.HasPrefix(lower, term) {
			return oidRankPrefix, true
		}
		if strings.Contains(lower, term) {
			return oidRankSubstring, true
		}
	}
	for _, candidate := range []string{node.Description, node.DescriptionZh, node.Source} {
		if candidate == "" {
			continue
		}
		if strings.Contains(strings.ToLower(candidate), term) {
			return oidRankSubstring, true
		}
	}
	return 0, false
}

// oidPage 是一次分页的结果。
type oidPage struct {
	// start 是起始下标（含）。
	start int
	// end 是结束下标（不含）。
	end int
}

// paginate 计算分页区间。
//
// 越界的 offset 返回空区间而不是报错 —— 分页越界不是输入错误。
func paginate(total, offset, limit int) (oidPage, int, bool) {
	if limit <= 0 {
		limit = OIDDefaultLimit
	}
	if limit > OIDMaxLimit {
		limit = OIDMaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return oidPage{start: offset, end: end}, total, end < total
}

// normalizeOID 归一化用户输入的 OID。
//
// 去掉空白与首尾的点；空串表示「未指定」。
func normalizeOID(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "OID:")
	value = strings.TrimPrefix(value, "oid:")
	value = strings.TrimSpace(value)
	value = strings.Trim(value, ".")
	return value
}

// Validate 做字段级校验。
//
// 错误**不回显**用户输入，只报告期望形式。
func (r *OIDLookupRequest) Validate() error {
	oid := normalizeOID(r.OID)
	query := strings.TrimSpace(r.Query)

	if oid != "" && query != "" {
		return tools.InvalidField("query", "与 oid 互斥", "同时给出 oid 与 query")
	}
	if len(oid) > OIDMaxLength {
		return tools.InvalidField("oid", []int{0, OIDMaxLength}, len(oid))
	}
	if oid != "" && !validOIDSyntax(oid) {
		return tools.InvalidField("oid", "点分十进制弧序列，如 1.3.6.1.5.5.7.3.36", "syntax error")
	}
	if len(query) > OIDMaxQueryLength {
		return tools.InvalidField("query", []int{0, OIDMaxQueryLength}, len(query))
	}
	if r.Offset < 0 {
		return tools.InvalidField("offset", "非负整数", r.Offset)
	}
	if r.Limit < 0 || r.Limit > OIDMaxLimit {
		return tools.InvalidField("limit", []int{0, OIDMaxLimit}, r.Limit)
	}
	return nil
}

// validOIDSyntax 校验 OID 语法：一串非负整数弧，用点分隔。
func validOIDSyntax(oid string) bool {
	if oid == "" {
		return false
	}
	arcs := strings.Split(oid, ".")
	if len(arcs) > OIDMaxArcs {
		return false
	}
	for _, arc := range arcs {
		if arc == "" {
			return false
		}
		for _, ch := range arc {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}
