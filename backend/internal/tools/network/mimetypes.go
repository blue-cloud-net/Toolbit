package network

import (
	"context"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// MimeCategories 返回受支持的顶层类型（顺序稳定）。
func MimeCategories() []string {
	return []string{
		"application",
		"audio",
		"font",
		"image",
		"message",
		"model",
		"multipart",
		"text",
		"video",
	}
}

// MimeEntry 是一个媒体类型条目。
type MimeEntry struct {
	// MimeType 是规范媒体类型，如 `image/png`。
	MimeType string `json:"mimeType"`
	// Extensions 是常用扩展名（小写，不含点）。
	Extensions []string `json:"extensions"`
	// Category 是顶层类型：application / audio / font / image / message /
	// model / multipart / text / video。
	Category string `json:"category"`
	// IsTextual 是否是文本类媒体类型。
	IsTextual bool `json:"isTextual"`
	// Charset 是默认字符集；非文本类为空串。
	Charset string `json:"charset"`
	// Description 是用途说明。
	Description string `json:"description"`
	// Aliases 是历史别名与常见变体。
	Aliases []string `json:"aliases"`
}

// MimeLookupRequest 是 mime-lookup 工具的请求。
//
// 两个字段都可以省略：都不给时返回**全量**数据集，
// 便于前端默认展示完整列表、再按输入本地过滤。
type MimeLookupRequest struct {
	// Query 是查询词，可写媒体类型（`image/png`）、扩展名（`png`、`.png`、
	// `report.png`）或用途说明里的关键词（如 `图片`）。
	Query string `json:"query"`
	// Category 是顶层类型过滤；为空表示不过滤。
	Category string `json:"category"`
}

// MimeLookupResult 是 mime-lookup 工具的响应。
type MimeLookupResult struct {
	// Query 是规范化后的查询词；全量列表时为空串。
	Query string `json:"query"`
	// Category 是生效的顶层类型过滤；未过滤时为空串。
	Category string `json:"category"`
	// Entries 是命中的条目；未命中时为空数组。
	Entries []MimeEntry `json:"entries"`
	// Total 是命中条数。
	Total int `json:"total"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// MIME 数据集缓存。
var (
	mimeOnce    sync.Once
	mimeEntries []MimeEntry
	// mimeByType 按规范媒体类型索引。
	mimeByType map[string]*MimeEntry
	// mimeByAlias 按历史别名索引。
	mimeByAlias map[string]*MimeEntry
	// mimeByExtension 按扩展名索引。
	mimeByExtension map[string]*MimeEntry
)

// loadMimeDataset 解析并缓存 MIME 数据集。
//
// 解析只做一次：CSV 是内嵌的静态数据，重复解析纯属浪费。
func loadMimeDataset() {
	mimeOnce.Do(func() {
		mimeEntries = parseMimeEntries(readDataset("dataset/mime_types.csv"))
		mimeByType = make(map[string]*MimeEntry, len(mimeEntries))
		mimeByAlias = make(map[string]*MimeEntry, len(mimeEntries))
		mimeByExtension = make(map[string]*MimeEntry, len(mimeEntries))
		for i := range mimeEntries {
			entry := &mimeEntries[i]
			mimeByType[entry.MimeType] = entry
			for _, alias := range entry.Aliases {
				if _, ok := mimeByAlias[alias]; !ok {
					mimeByAlias[alias] = entry
				}
			}
			for _, ext := range entry.Extensions {
				if _, ok := mimeByExtension[ext]; !ok {
					mimeByExtension[ext] = entry
				}
			}
		}
	})
}

// parseMimeEntries 解析 MIME 数据集。
func parseMimeEntries(r io.Reader) []MimeEntry {
	out := []MimeEntry{}
	for _, rec := range readCSV(r) {
		mimeType := strings.ToLower(strings.TrimSpace(csvField(rec, 0)))
		if mimeType == "" {
			continue
		}
		entry := MimeEntry{
			MimeType:    mimeType,
			Extensions:  splitList(csvField(rec, 1)),
			Charset:     strings.TrimSpace(csvField(rec, 2)),
			Description: strings.TrimSpace(csvField(rec, 3)),
			Aliases:     splitList(csvField(rec, 4)),
			Category:    mimeCategory(mimeType),
		}
		// 文本类媒体类型的判定依据是「有默认字符集」。
		entry.IsTextual = entry.Charset != ""
		out = append(out, entry)
	}
	return out
}

// MimeLookup 查询媒体类型数据集。
//
// 数据集是**离线内置**的快照，**不**出网；也**不**读取系统 MIME 数据库，
// 因为各发行版的映射表并不一致。
//
// 三种用法：
//   - 空请求：返回全量条目（列表浏览）；
//   - 精确词（媒体类型 / 别名 / 扩展名）：命中项排在最前；
//   - 其它关键词：在媒体类型、扩展名、别名、用途说明与顶层类型里做子串匹配。
//
// `category` 可与任一方式叠加，用于收窄到某一顶层类型。
func MimeLookup(_ context.Context, req *MimeLookupRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadMimeDataset()

	query := normalizeMimeQuery(req.Query)
	category := strings.ToLower(strings.TrimSpace(req.Category))

	type scored struct {
		entry MimeEntry
		rank  int
	}
	hits := make([]scored, 0, len(mimeEntries))
	for i := range mimeEntries {
		entry := &mimeEntries[i]
		if category != "" && entry.Category != category {
			continue
		}
		rank, ok := matchMimeEntry(entry, query)
		if !ok {
			continue
		}
		hits = append(hits, scored{entry: *entry, rank: rank})
	}

	// 精确命中排前面，其余保持数据集顺序（稳定排序）。
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].rank < hits[j].rank })

	matched := make([]MimeEntry, 0, len(hits))
	for _, hit := range hits {
		matched = append(matched, hit.entry)
	}

	return tools.Result{Data: MimeLookupResult{
		Query:       query,
		Category:    category,
		Entries:     matched,
		Total:       len(matched),
		DataVersion: MimeTypesVersion,
	}}, nil
}

// 命中档位：数字越小越靠前。
const (
	// mimeRankExact 是媒体类型或别名完全相同。
	mimeRankExact = 0
	// mimeRankExtension 是扩展名完全相同。
	mimeRankExtension = 1
	// mimeRankKeyword 是子串命中。
	mimeRankKeyword = 2
)

// matchMimeEntry 报告条目是否命中查询词，并给出命中档位。
//
// 匹配范围：媒体类型、扩展名、历史别名、用途说明与顶层类型。
//
// 精确命中（媒体类型 / 别名 / 扩展名）通过数据集索引判定，
// 保证「同一个扩展名被多个类型描述」时结果稳定（先到先得）。
func matchMimeEntry(entry *MimeEntry, query string) (int, bool) {
	if query == "" {
		return mimeRankKeyword, true
	}
	if found, ok := mimeByType[query]; ok && found == entry {
		return mimeRankExact, true
	}
	if found, ok := mimeByAlias[query]; ok && found == entry {
		return mimeRankExact, true
	}
	if found, ok := mimeByExtension[query]; ok && found == entry {
		return mimeRankExtension, true
	}
	candidates := make([]string, 0, len(entry.Extensions)+len(entry.Aliases)+3)
	candidates = append(candidates, entry.MimeType, entry.Category, entry.Description)
	candidates = append(candidates, entry.Extensions...)
	candidates = append(candidates, entry.Aliases...)
	for _, candidate := range candidates {
		if candidate != "" && strings.Contains(strings.ToLower(candidate), query) {
			return mimeRankKeyword, true
		}
	}
	return 0, false
}

// Validate 做字段级校验。
//
// 查询词是**自由文本**（允许中文关键词），因此只校验长度与顶层类型枚举；
// 空请求是合法的 —— 表示「列出全部条目」。
func (r *MimeLookupRequest) Validate() error {
	if len(r.Query) > 128 {
		return tools.InvalidField("query", []int{0, 128}, len(r.Query))
	}
	category := strings.ToLower(strings.TrimSpace(r.Category))
	if category != "" && !containsString(MimeCategories(), category) {
		return tools.InvalidField("category", MimeCategories(), category)
	}
	return nil
}

// normalizeMimeQuery 归一化查询词。
//
//   - 媒体类型：去掉 `;` 之后的参数，统一小写；
//   - 扩展名：去掉目录前缀与前导点，统一小写（`report.PNG` → `png`）。
func normalizeMimeQuery(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, ";"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	// 形如 `type/subtype` 且子类型里不含点的，按媒体类型处理；
	// 否则按文件名 / 路径取扩展名（`dir/report.pdf` → `pdf`）。
	if strings.Contains(value, "/") {
		rest := value[strings.LastIndex(value, "/")+1:]
		if !strings.Contains(rest, ".") {
			return normalizeMimeType(value)
		}
	}
	return normalizeExtension(value)
}

// normalizeExtension 把用户输入规整成扩展名。
//
// 接受 `png`、`.png`、`report.png` 与 `dir/report.png`；统一小写并去掉前导点。
func normalizeExtension(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if idx := strings.LastIndexAny(value, "/\\"); idx >= 0 {
		value = value[idx+1:]
	}
	if idx := strings.LastIndex(value, "."); idx >= 0 {
		value = value[idx+1:]
	}
	return strings.TrimPrefix(value, ".")
}

// normalizeMimeType 规整媒体类型：去掉参数、转小写。
func normalizeMimeType(raw string) string {
	value := strings.TrimSpace(raw)
	if idx := strings.Index(value, ";"); idx >= 0 {
		value = value[:idx]
	}
	return strings.ToLower(strings.TrimSpace(value))
}

// mimeCategory 取媒体类型的顶层类型。
func mimeCategory(mimeType string) string {
	if idx := strings.Index(mimeType, "/"); idx > 0 {
		return mimeType[:idx]
	}
	return "application"
}
