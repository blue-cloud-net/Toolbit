package system

import (
	"bytes"
	"context"
	"embed"
	"encoding/csv"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// lookupFS 内嵌速查表数据集。
//
// 这里是**唯一**的数据来源：端口、退出码与时间戳格式都是静态表，
// **不**在运行时查询系统服务，也不联网。
//
//go:embed dataset/lookup_table.csv
var lookupFS embed.FS

// LookupDataVersion 是速查表数据集的快照版本。
//
// 条目调整后**必须**更新快照与版本号，否则响应里的版本号会误导用户。
const LookupDataVersion = "2026-10"

// 速查表类别。
const (
	// LookupKindPort 是端口号速查。
	LookupKindPort = "port"
	// LookupKindExitCode 是 Linux 退出码速查。
	LookupKindExitCode = "exitcode"
	// LookupKindTimestamp 是时间戳格式速查。
	LookupKindTimestamp = "timestamp"
)

// LookupKind 返回受支持的速查表类别（顺序稳定）。
func LookupKind() []string {
	return []string{LookupKindPort, LookupKindExitCode, LookupKindTimestamp}
}

// LookupEntry 是一个速查条目。
type LookupEntry struct {
	// Kind 是条目类别。
	Kind string `json:"kind"`
	// Key 是查询键。
	Key string `json:"key"`
	// Label 是简短名称。
	Label string `json:"label"`
	// Description 是中文说明。
	Description string `json:"description"`
	// Extra 是附加信息。
	Extra []string `json:"extra"`
}

// LookupTableRequest 是 lookup-table 工具的请求。
type LookupTableRequest struct {
	// Kind 是速查表类别。
	Kind string `json:"kind"`
	// Key 是查询键。
	Key string `json:"key"`
}

// LookupTableResult 是 lookup-table 工具的响应。
type LookupTableResult struct {
	// Kind 是速查表类别。
	Kind string `json:"kind"`
	// Query 是规范化后的查询词。
	Query string `json:"query"`
	// Matched 是否精确命中。
	Matched bool `json:"matched"`
	// Entry 是命中的条目；未命中时为 nil。
	Entry *LookupEntry `json:"entry"`
	// Candidates 是模糊检索的候选；精确命中或无候选时为空数组。
	Candidates []LookupEntry `json:"candidates"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// 速查表索引。
var (
	lookupOnce   bool
	lookupByKind = map[string][]LookupEntry{}
	lookupByKey  = map[string]map[string]LookupEntry{}
)

// loadLookupTable 解析并缓存速查表数据集。
//
// 解析只做一次：CSV 是内嵌的静态数据，重复解析纯属浪费。
func loadLookupTable() {
	if lookupOnce {
		return
	}
	lookupOnce = true

	raw, err := lookupFS.ReadFile("dataset/lookup_table.csv")
	if err != nil {
		return
	}
	reader := csv.NewReader(bytes.NewReader(raw))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return
	}

	for _, rec := range records[1:] {
		if len(rec) < 4 {
			continue
		}
		entry := LookupEntry{
			Kind:        strings.TrimSpace(rec[0]),
			Key:         strings.TrimSpace(rec[1]),
			Label:       strings.TrimSpace(rec[2]),
			Description: strings.TrimSpace(rec[3]),
			Extra:       splitLookupList(rec, 4),
		}
		if entry.Kind == "" || entry.Key == "" {
			continue
		}
		lookupByKind[entry.Kind] = append(lookupByKind[entry.Kind], entry)
		if _, ok := lookupByKey[entry.Kind]; !ok {
			lookupByKey[entry.Kind] = map[string]LookupEntry{}
		}
		lookupByKey[entry.Kind][entry.Key] = entry
	}
}

// splitLookupList 读取第 index 列并按分号拆分。
func splitLookupList(rec []string, index int) []string {
	if index >= len(rec) {
		return []string{}
	}
	out := []string{}
	for _, part := range strings.Split(rec[index], ";") {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

// LookupTable 查询端口号、退出码或时间戳格式。
func LookupTable(_ context.Context, req *LookupTableRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadLookupTable()

	query := strings.ToLower(strings.TrimSpace(req.Key))
	result := LookupTableResult{
		Kind:        req.Kind,
		Query:       query,
		Matched:     false,
		Entry:       nil,
		Candidates:  []LookupEntry{},
		DataVersion: LookupDataVersion,
	}

	index, ok := lookupByKey[req.Kind]
	if !ok {
		return tools.Result{Data: result}, nil
	}

	// 精确命中：键本身，或同类的 label。
	if entry, found := index[query]; found {
		result.Matched = true
		result.Entry = &entry
		return tools.Result{Data: result}, nil
	}
	for _, entry := range lookupByKind[req.Kind] {
		if strings.ToLower(entry.Label) == query {
			result.Matched = true
			copied := entry
			result.Entry = &copied
			return tools.Result{Data: result}, nil
		}
	}

	result.Candidates = searchLookupEntries(lookupByKind[req.Kind], query)
	return tools.Result{Data: result}, nil
}

// searchLookupEntries 按匹配度检索候选条目。
//
// 匹配度依次是：label 前缀、label 包含、description 包含；
// 同级按条目在数据集中的顺序排列，保证结果稳定。
func searchLookupEntries(entries []LookupEntry, query string) []LookupEntry {
	if query == "" {
		return []LookupEntry{}
	}
	out := []LookupEntry{}
	for _, entry := range entries {
		label := strings.ToLower(entry.Label)
		description := strings.ToLower(entry.Description)
		switch {
		case strings.HasPrefix(label, query):
		case strings.Contains(label, query):
		case strings.Contains(description, query):
		default:
			continue
		}
		out = append(out, entry)
		if len(out) >= 10 {
			break
		}
	}
	return out
}

// Validate 做字段级校验。
func (r *LookupTableRequest) Validate() error {
	if r.Kind == "" {
		return tools.MissingField("kind")
	}
	if !containsLookupKind(LookupKind(), r.Kind) {
		return tools.InvalidField("kind", LookupKind(), r.Kind)
	}
	if strings.TrimSpace(r.Key) == "" {
		return tools.MissingField("key")
	}
	if len(r.Key) > 64 {
		return tools.InvalidField("key", []int{0, 64}, len(r.Key))
	}
	return nil
}

// containsLookupKind 判断切片是否包含某值。
func containsLookupKind(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
