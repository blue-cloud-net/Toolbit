package data

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// TimeFormatsVersion 是时间格式数据集的快照版本。
const TimeFormatsVersion = "2026-11"

// 时间格式分类。
const (
	// TimeFormatCategoryEpoch 是时间戳与计数类格式。
	TimeFormatCategoryEpoch = "epoch"
	// TimeFormatCategoryStandard 是标准日期格式。
	TimeFormatCategoryStandard = "standard"
	// TimeFormatCategoryNumeric 是数值纪年格式。
	TimeFormatCategoryNumeric = "numeric"
)

// 查询方式。
const (
	// TimeFormatQueryAll 表示未给出查询条件，返回全量列表。
	TimeFormatQueryAll = "all"
	// TimeFormatQueryKeyword 表示按关键字查询。
	TimeFormatQueryKeyword = "keyword"
)

// TimeFormatCategories 返回时间格式分类（顺序稳定）。
func TimeFormatCategories() []string {
	return []string{
		TimeFormatCategoryEpoch,
		TimeFormatCategoryStandard,
		TimeFormatCategoryNumeric,
	}
}

// TimeFormatEntry 是一个时间格式条目。
type TimeFormatEntry struct {
	// Key 是格式标识。
	Key string `json:"key"`
	// Label 是格式名称。
	Label string `json:"label"`
	// Description 是中文说明。
	Description string `json:"description"`
	// Example 是同一时刻在该格式下的示例值。
	Example string `json:"example"`
	// Category 是格式分类。
	Category string `json:"category"`
}

// TimeFormatRequest 是 time-format 工具的请求。
//
// 字段可以省略：不给时返回**全量**格式列表。
type TimeFormatRequest struct {
	// Query 是查询词：格式标识（如 unix-ms）或关键字（如 RFC）。
	Query string `json:"query"`
	// Category 是分类过滤；为空表示不过滤。
	Category string `json:"category"`
}

// TimeFormatResult 是 time-format 工具的响应。
type TimeFormatResult struct {
	// Query 是规范化后的查询词。
	Query string `json:"query"`
	// Kind 是查询方式：all / keyword。
	Kind string `json:"kind"`
	// Matched 是否命中格式条目。
	Matched bool `json:"matched"`
	// Key 是命中的格式标识。
	Key string `json:"key"`
	// Label 是命中的格式名称。
	Label string `json:"label"`
	// Description 是命中的中文说明。
	Description string `json:"description"`
	// Example 是命中的示例值。
	Example string `json:"example"`
	// Category 是命中的分类。
	Category string `json:"category"`
	// Matches 是命中的条目列表。
	Matches []TimeFormatEntry `json:"matches"`
	// Total 是命中条数。
	Total int `json:"total"`
	// DatasetTotal 是数据集条目数。
	DatasetTotal int `json:"datasetTotal"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// 时间格式数据集缓存。
var (
	timeFormatsOnce   sync.Once
	timeFormatEntries []TimeFormatEntry
)

// loadTimeFormats 解析并缓存时间格式数据集。
func loadTimeFormats() {
	timeFormatsOnce.Do(func() {
		timeFormatEntries = parseTimeFormats(readDataset("dataset/time_formats.csv"))
	})
}

// parseTimeFormats 解析时间格式数据集。
func parseTimeFormats(r io.Reader) []TimeFormatEntry {
	out := []TimeFormatEntry{}
	for _, rec := range readCSV(r) {
		if len(rec) < 5 {
			continue
		}
		out = append(out, TimeFormatEntry{
			Key:         strings.TrimSpace(rec[0]),
			Label:       strings.TrimSpace(rec[1]),
			Description: strings.TrimSpace(rec[2]),
			Example:     strings.TrimSpace(rec[3]),
			Category:    strings.TrimSpace(rec[4]),
		})
	}
	return out
}

// TimeFormats 查询常见日期时间格式的定义与示例。
//
// 数据集是**离线内置**的快照，**不**出网。
func TimeFormats(_ context.Context, req *TimeFormatRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadTimeFormats()

	query := strings.TrimSpace(req.Query)
	result := TimeFormatResult{
		Query:        query,
		Kind:         TimeFormatQueryAll,
		Matches:      []TimeFormatEntry{},
		DatasetTotal: len(timeFormatEntries),
		DataVersion:  TimeFormatsVersion,
	}

	if query == "" && req.Category == "" {
		result.Matches = append([]TimeFormatEntry{}, timeFormatEntries...)
		result.Total = len(result.Matches)
		return tools.Result{Data: result}, nil
	}

	result.Kind = TimeFormatQueryKeyword
	lowered := strings.ToLower(query)
	for _, entry := range timeFormatEntries {
		if req.Category != "" && entry.Category != req.Category {
			continue
		}
		if query != "" && !matchesTimeFormat(entry, lowered) {
			continue
		}
		result.Matches = append(result.Matches, entry)
	}
	result.Total = len(result.Matches)
	if len(result.Matches) > 0 {
		first := result.Matches[0]
		result.Matched = true
		result.Key = first.Key
		result.Label = first.Label
		result.Description = first.Description
		result.Example = first.Example
		result.Category = first.Category
	}
	return tools.Result{Data: result}, nil
}

// matchesTimeFormat 判断条目是否命中关键字。
//
// 匹配范围：格式标识、格式名称与中文说明。
func matchesTimeFormat(entry TimeFormatEntry, lowered string) bool {
	if lowered == "" {
		return false
	}
	if strings.Contains(strings.ToLower(entry.Key), lowered) {
		return true
	}
	if strings.Contains(strings.ToLower(entry.Label), lowered) {
		return true
	}
	return strings.Contains(entry.Description, lowered)
}

// Validate 做字段级校验。
func (r *TimeFormatRequest) Validate() error {
	if len(r.Query) > 64 {
		return tools.InvalidField("query", []int{0, 64}, len(r.Query))
	}
	if r.Category != "" && !containsString(TimeFormatCategories(), r.Category) {
		return tools.InvalidField("category", TimeFormatCategories(), r.Category)
	}
	return nil
}
