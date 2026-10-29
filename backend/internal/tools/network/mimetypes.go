package network

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// MIME 查询方式。
const (
	// MimeQueryExtension 表示按文件扩展名查询。
	MimeQueryExtension = "extension"
	// MimeQueryType 表示按媒体类型查询。
	MimeQueryType = "mimeType"
)

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
type MimeLookupRequest struct {
	// Extension 是扩展名查询：可写 `png`、`.png` 或 `report.png`。
	Extension string `json:"extension"`
	// MimeType 是媒体类型查询：可写 `image/png` 或带参数的
	// `text/plain; charset=utf-8`。
	MimeType string `json:"mimeType"`
}

// MimeLookupResult 是 mime-lookup 工具的响应。
type MimeLookupResult struct {
	// Query 是规范化后的查询词。
	Query string `json:"query"`
	// Kind 是查询方式：extension / mimeType。
	Kind string `json:"kind"`
	// Matched 是否命中数据集。
	Matched bool `json:"matched"`
	// MimeType 是规范媒体类型；未命中时为空串。
	MimeType string `json:"mimeType"`
	// Extensions 是该媒体类型的常用扩展名。
	Extensions []string `json:"extensions"`
	// Category 是顶层类型。
	Category string `json:"category"`
	// IsTextual 是否是文本类媒体类型。
	IsTextual bool `json:"isTextual"`
	// Charset 是默认字符集。
	Charset string `json:"charset"`
	// Description 是用途说明。
	Description string `json:"description"`
	// Aliases 是历史别名与常见变体。
	Aliases []string `json:"aliases"`
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

// MimeLookup 查询文件扩展名或媒体类型对应的 MIME 信息。
//
// 数据集是**离线内置**的快照，**不**出网；也**不**读取系统 MIME 数据库，
// 因为各发行版的映射表并不一致。
func MimeLookup(_ context.Context, req *MimeLookupRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadMimeDataset()

	result := MimeLookupResult{
		Kind:        MimeQueryExtension,
		Extensions:  []string{},
		Aliases:     []string{},
		DataVersion: MimeTypesVersion,
	}

	if strings.TrimSpace(req.Extension) != "" {
		ext := normalizeExtension(req.Extension)
		result.Kind = MimeQueryExtension
		result.Query = ext
		if entry, ok := mimeByExtension[ext]; ok {
			fillMimeResult(&result, entry)
		}
		return tools.Result{Data: result}, nil
	}

	mimeType := normalizeMimeType(req.MimeType)
	result.Kind = MimeQueryType
	result.Query = mimeType
	if entry, ok := lookupMimeType(mimeType); ok {
		fillMimeResult(&result, entry)
	}
	return tools.Result{Data: result}, nil
}

// lookupMimeType 按媒体类型查条目：先查规范名，再查历史别名。
func lookupMimeType(mimeType string) (*MimeEntry, bool) {
	if entry, ok := mimeByType[mimeType]; ok {
		return entry, true
	}
	entry, ok := mimeByAlias[mimeType]
	return entry, ok
}

// fillMimeResult 把条目内容写进响应。
func fillMimeResult(result *MimeLookupResult, entry *MimeEntry) {
	result.Matched = true
	result.MimeType = entry.MimeType
	result.Category = entry.Category
	result.IsTextual = entry.IsTextual
	result.Charset = entry.Charset
	result.Description = entry.Description
	result.Extensions = append([]string{}, entry.Extensions...)
	result.Aliases = append([]string{}, entry.Aliases...)
}

// Validate 做字段级校验。
func (r *MimeLookupRequest) Validate() error {
	ext := strings.TrimSpace(r.Extension)
	mime := strings.TrimSpace(r.MimeType)

	if ext == "" && mime == "" {
		return tools.MissingField("extension")
	}
	if ext != "" && mime != "" {
		return tools.InvalidField("mimeType", "only one of extension / mimeType", "both provided")
	}
	if len(r.Extension) > 128 {
		return tools.InvalidField("extension", []int{0, 128}, len(r.Extension))
	}
	if len(r.MimeType) > 128 {
		return tools.InvalidField("mimeType", []int{0, 128}, len(r.MimeType))
	}

	if ext != "" && !isValidExtension(normalizeExtension(r.Extension)) {
		return tools.InvalidField("extension", "alphanumeric extension", "invalid format")
	}
	if mime != "" && !isValidMimeType(normalizeMimeType(r.MimeType)) {
		return tools.InvalidField("mimeType", "type/subtype", "invalid format")
	}
	return nil
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

// isValidExtension 判断规整后的扩展名是否合法。
func isValidExtension(ext string) bool {
	if ext == "" || len(ext) > 32 {
		return false
	}
	for _, c := range ext {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '+' || c == '-' || c == '_':
		default:
			return false
		}
	}
	return true
}

// isValidMimeType 判断规整后的媒体类型是否形如 `type/subtype`。
func isValidMimeType(mimeType string) bool {
	if mimeType == "" || len(mimeType) > 128 {
		return false
	}
	parts := strings.Split(mimeType, "/")
	if len(parts) != 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, c := range part {
			switch {
			case c >= 'a' && c <= 'z':
			case c >= '0' && c <= '9':
			case c == '+' || c == '-' || c == '.' || c == '_':
			default:
				return false
			}
		}
	}
	return true
}

// mimeCategory 取媒体类型的顶层类型。
func mimeCategory(mimeType string) string {
	if idx := strings.Index(mimeType, "/"); idx > 0 {
		return mimeType[:idx]
	}
	return "application"
}
