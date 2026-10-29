package network

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// HTTP 状态码类别。
const (
	// StatusCategoryInformational 是 1xx：信息。
	StatusCategoryInformational = "informational"
	// StatusCategorySuccess 是 2xx：成功。
	StatusCategorySuccess = "success"
	// StatusCategoryRedirect 是 3xx：重定向。
	StatusCategoryRedirect = "redirect"
	// StatusCategoryClientError 是 4xx：客户端错误。
	StatusCategoryClientError = "clientError"
	// StatusCategoryServerError 是 5xx：服务端错误。
	StatusCategoryServerError = "serverError"
)

// 查询方式。
const (
	// StatusQueryCode 表示按状态码查询。
	StatusQueryCode = "code"
	// StatusQueryKeyword 表示按关键字查询。
	StatusQueryKeyword = "keyword"
)

// HttpStatusEntry 是一个 HTTP 状态码条目。
type HttpStatusEntry struct {
	// Code 是状态码。
	Code int `json:"code"`
	// Reason 是标准原因短语。
	Reason string `json:"reason"`
	// Category 是类别：informational / success / redirect / clientError / serverError。
	Category string `json:"category"`
	// Description 是中文说明。
	Description string `json:"description"`
	// Deprecated 是否已废弃。
	Deprecated bool `json:"deprecated"`
}

// HttpStatusRequest 是 http-status 工具的请求。
type HttpStatusRequest struct {
	// Code 是状态码查询（如 404）；`0` 表示未给出。
	Code int `json:"code"`
	// Query 是关键字查询（如 `not found`、`重定向`）。
	Query string `json:"query"`
}

// HttpStatusResult 是 http-status 工具的响应。
type HttpStatusResult struct {
	// Query 是规范化后的查询词。
	Query string `json:"query"`
	// Kind 是查询方式：code / keyword。
	Kind string `json:"kind"`
	// Matched 是否命中标准状态码。
	Matched bool `json:"matched"`
	// Code 是状态码；关键字查询时取第一条命中，未命中为 `0`。
	Code int `json:"code"`
	// Reason 是标准原因短语。
	Reason string `json:"reason"`
	// Category 是类别。
	Category string `json:"category"`
	// Description 是中文说明。
	Description string `json:"description"`
	// Deprecated 是否已废弃。
	Deprecated bool `json:"deprecated"`
	// Matches 是命中的条目；状态码查询至多一条。
	Matches []HttpStatusEntry `json:"matches"`
	// Total 是命中条数。
	Total int `json:"total"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// HTTP 状态码数据集缓存。
var (
	statusOnce    sync.Once
	statusEntries []HttpStatusEntry
	// statusByCode 按状态码索引。
	statusByCode map[int]*HttpStatusEntry
)

// loadStatusDataset 解析并缓存状态码数据集。
func loadStatusDataset() {
	statusOnce.Do(func() {
		statusEntries = parseStatusEntries(readDataset("dataset/http_status.csv"))
		statusByCode = make(map[int]*HttpStatusEntry, len(statusEntries))
		for i := range statusEntries {
			entry := &statusEntries[i]
			statusByCode[entry.Code] = entry
		}
	})
}

// parseStatusEntries 解析状态码数据集。
func parseStatusEntries(r io.Reader) []HttpStatusEntry {
	out := []HttpStatusEntry{}
	for _, rec := range readCSV(r) {
		code, err := strconv.Atoi(strings.TrimSpace(csvField(rec, 0)))
		if err != nil {
			continue
		}
		out = append(out, HttpStatusEntry{
			Code:        code,
			Reason:      strings.TrimSpace(csvField(rec, 1)),
			Description: strings.TrimSpace(csvField(rec, 2)),
			Deprecated:  strings.TrimSpace(csvField(rec, 3)) == "1",
			Category:    statusCategory(code),
		})
	}
	return out
}

// HttpStatus 查询 HTTP 状态码的类别、原因短语与说明。
//
// 数据集是**离线内置**的快照，**不**出网。
//
// 两种查询方式：
//   - 按状态码：给 `code`，或给一个纯数字的 `query`；
//   - 按关键字：给 `query`，匹配原因短语与中文说明。
func HttpStatus(_ context.Context, req *HttpStatusRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadStatusDataset()

	result := HttpStatusResult{
		Matches:     []HttpStatusEntry{},
		DataVersion: HTTPStatusVersion,
	}

	// 纯数字的关键字按状态码处理。
	code := req.Code
	if code == 0 {
		if raw := strings.TrimSpace(req.Query); isAllDigits(raw) {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				return tools.Result{}, invalidStatusCode()
			}
			code = parsed
		}
	}

	if code != 0 {
		result.Kind = StatusQueryCode
		result.Query = strconv.Itoa(code)
		fillStatusByCode(&result, code)
		if entry, ok := statusByCode[code]; ok {
			result.Matches = []HttpStatusEntry{*entry}
			result.Total = 1
		}
		return tools.Result{Data: result}, nil
	}

	query := strings.TrimSpace(req.Query)
	result.Kind = StatusQueryKeyword
	result.Query = query

	matched := []HttpStatusEntry{}
	lowered := strings.ToLower(query)
	for _, entry := range statusEntries {
		if matchesStatus(entry, lowered) {
			matched = append(matched, entry)
		}
	}

	result.Matches = matched
	result.Total = len(matched)
	if len(matched) > 0 {
		first := matched[0]
		result.Matched = true
		result.Code = first.Code
		result.Reason = first.Reason
		result.Category = first.Category
		result.Description = first.Description
		result.Deprecated = first.Deprecated
	}
	return tools.Result{Data: result}, nil
}

// fillStatusByCode 按状态码填充响应。
//
// 落在 100–599 但不在表内的状态码（如 299）仍给出类别，
// 便于用户判断它属于哪一档。
func fillStatusByCode(result *HttpStatusResult, code int) {
	result.Code = code
	result.Category = statusCategory(code)
	if entry, ok := statusByCode[code]; ok {
		result.Matched = true
		result.Reason = entry.Reason
		result.Description = entry.Description
		result.Deprecated = entry.Deprecated
	}
}

// matchesStatus 判断条目是否命中关键字。
//
// 匹配范围：原因短语（不区分大小写）、中文说明与类别中文名，
// 因此「重定向」「客户端错误」这类中文关键字同样可用。
func matchesStatus(entry HttpStatusEntry, lowered string) bool {
	if lowered == "" {
		return false
	}
	if strings.Contains(strings.ToLower(entry.Reason), lowered) {
		return true
	}
	if strings.Contains(entry.Description, lowered) {
		return true
	}
	return strings.Contains(statusCategoryLabel(entry.Category), lowered)
}

// statusCategoryLabel 给类别一个中文名，用于关键字检索。
//
// 前端另有本地化词条；这里的中文名只服务于服务端检索。
func statusCategoryLabel(category string) string {
	switch category {
	case StatusCategoryInformational:
		return "信息"
	case StatusCategorySuccess:
		return "成功"
	case StatusCategoryRedirect:
		return "重定向"
	case StatusCategoryClientError:
		return "客户端错误"
	case StatusCategoryServerError:
		return "服务端错误"
	default:
		return ""
	}
}

// Validate 做字段级校验。
func (r *HttpStatusRequest) Validate() error {
	query := strings.TrimSpace(r.Query)

	if r.Code == 0 && query == "" {
		return tools.MissingField("code")
	}
	if r.Code != 0 && query != "" {
		return tools.InvalidField("query", "only one of code / query", "both provided")
	}
	if len(r.Query) > 128 {
		return tools.InvalidField("query", []int{0, 128}, len(r.Query))
	}

	if r.Code != 0 {
		return validateStatusCode(r.Code)
	}
	if isAllDigits(query) {
		parsed, err := strconv.Atoi(query)
		if err != nil {
			return invalidStatusCode()
		}
		return validateStatusCode(parsed)
	}
	return nil
}

// validateStatusCode 判断状态码是否落在合法区间。
func validateStatusCode(code int) error {
	if code < 100 || code > 599 {
		return invalidStatusCode()
	}
	return nil
}

// invalidStatusCode 构造状态码非法的错误。
//
// **不回显**用户输入：只声明合法区间。
func invalidStatusCode() error {
	return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
		"field":    "code",
		"expected": "100-599",
	})
}

// isAllDigits 判断字符串是否只含数字。
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// statusCategory 按状态码首位数字给出类别。
func statusCategory(code int) string {
	switch {
	case code >= 100 && code < 200:
		return StatusCategoryInformational
	case code >= 200 && code < 300:
		return StatusCategorySuccess
	case code >= 300 && code < 400:
		return StatusCategoryRedirect
	case code >= 400 && code < 500:
		return StatusCategoryClientError
	case code >= 500 && code < 600:
		return StatusCategoryServerError
	default:
		return ""
	}
}
