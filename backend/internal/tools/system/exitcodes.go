package system

import (
	"context"
	"embed"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// exitCodeFS 内嵌退出码数据集。
//
// 这里是**唯一**的数据来源：退出码是静态表，
// **不**在运行时执行命令，也**不**读取系统状态。
//
//go:embed dataset/exit_codes.csv
var exitCodeFS embed.FS

// ExitCodesVersion 是退出码数据集的快照版本。
const ExitCodesVersion = "2026-11"

// 退出码类别。
const (
	// ExitCodeTypeSuccess 是正常退出。
	ExitCodeTypeSuccess = "success"
	// ExitCodeTypeFailure 是失败退出。
	ExitCodeTypeFailure = "failure"
	// ExitCodeTypeSignal 是被信号终止。
	ExitCodeTypeSignal = "signal"
)

// 查询方式。
const (
	// ExitCodeQueryAll 表示未给出查询条件，返回全量列表。
	ExitCodeQueryAll = "all"
	// ExitCodeQueryCode 表示按退出码查询。
	ExitCodeQueryCode = "code"
	// ExitCodeQueryKeyword 表示按关键字查询。
	ExitCodeQueryKeyword = "keyword"
)

// ExitCodeTypes 返回退出码类别（顺序稳定）。
func ExitCodeTypes() []string {
	return []string{ExitCodeTypeSuccess, ExitCodeTypeFailure, ExitCodeTypeSignal}
}

// ExitCodeEntry 是一个退出码条目。
type ExitCodeEntry struct {
	// Code 是退出码。
	Code int `json:"code"`
	// Label 是简短名称。
	Label string `json:"label"`
	// Description 是中文说明。
	Description string `json:"description"`
	// Type 是类别：success / failure / signal。
	Type string `json:"type"`
}

// ExitCodeRequest 是 exit-code 工具的请求。
//
// 字段都可以省略：都不给时返回**全量**退出码列表。
type ExitCodeRequest struct {
	// Query 是查询词：退出码（如 127）或关键字（如 段错误）。
	Query string `json:"query"`
	// Type 是类别过滤：success / failure / signal；为空表示不过滤。
	Type string `json:"type"`
}

// ExitCodeResult 是 exit-code 工具的响应。
type ExitCodeResult struct {
	// Query 是规范化后的查询词。
	Query string `json:"query"`
	// Kind 是查询方式：all / code / keyword。
	Kind string `json:"kind"`
	// Matched 是否命中退出码条目。
	Matched bool `json:"matched"`
	// Code 是命中的退出码；未命中为 `0`。
	Code int `json:"code"`
	// Label 是命中的简短名称。
	Label string `json:"label"`
	// Description 是命中的中文说明。
	Description string `json:"description"`
	// Type 是命中的类别。
	Type string `json:"type"`
	// Matches 是命中的条目列表。
	Matches []ExitCodeEntry `json:"matches"`
	// Total 是命中条数。
	Total int `json:"total"`
	// DatasetTotal 是数据集条目数。
	DatasetTotal int `json:"datasetTotal"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// 退出码数据集缓存。
var (
	exitCodesOnce   sync.Once
	exitCodeEntries []ExitCodeEntry
	exitCodesByCode map[int]*ExitCodeEntry
)

// loadExitCodes 解析并缓存退出码数据集。
func loadExitCodes() {
	exitCodesOnce.Do(func() {
		raw, err := exitCodeFS.ReadFile("dataset/exit_codes.csv")
		if err != nil {
			return
		}
		exitCodeEntries = parseExitCodes(strings.NewReader(string(raw)))
		exitCodesByCode = make(map[int]*ExitCodeEntry, len(exitCodeEntries))
		for i := range exitCodeEntries {
			entry := &exitCodeEntries[i]
			exitCodesByCode[entry.Code] = entry
		}
	})
}

// parseExitCodes 解析退出码数据集。
func parseExitCodes(r io.Reader) []ExitCodeEntry {
	out := []ExitCodeEntry{}
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return out
	}
	for _, rec := range records[1:] {
		if len(rec) < 4 {
			continue
		}
		code, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			continue
		}
		out = append(out, ExitCodeEntry{
			Code:        code,
			Label:       strings.TrimSpace(rec[1]),
			Description: strings.TrimSpace(rec[2]),
			Type:        strings.TrimSpace(rec[3]),
		})
	}
	return out
}

// ExitCodes 查询 Linux 退出码的含义。
//
// 数据集是**离线内置**的快照，**不**出网。
func ExitCodes(_ context.Context, req *ExitCodeRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadExitCodes()

	result := ExitCodeResult{
		Query:        strings.TrimSpace(req.Query),
		Kind:         ExitCodeQueryAll,
		Matches:      []ExitCodeEntry{},
		DatasetTotal: len(exitCodeEntries),
		DataVersion:  ExitCodesVersion,
	}

	query := result.Query
	if query == "" && req.Type == "" {
		result.Matches = append([]ExitCodeEntry{}, exitCodeEntries...)
		result.Total = len(result.Matches)
		return tools.Result{Data: result}, nil
	}

	if query != "" && isAllDigits(query) {
		code, err := strconv.Atoi(query)
		if err != nil {
			return tools.Result{}, tools.InvalidField("query", "退出码或关键字", query)
		}
		result.Kind = ExitCodeQueryCode
		result.Query = strconv.Itoa(code)
		if entry, ok := exitCodesByCode[code]; ok && exitCodeInType(*entry, req.Type) {
			result.Matches = []ExitCodeEntry{*entry}
			result.Total = 1
			fillExitCode(&result, *entry)
		}
		return tools.Result{Data: result}, nil
	}

	result.Kind = ExitCodeQueryKeyword
	lowered := strings.ToLower(query)
	for _, entry := range exitCodeEntries {
		if !exitCodeInType(entry, req.Type) {
			continue
		}
		if query != "" && !matchesExitCode(entry, lowered) {
			continue
		}
		result.Matches = append(result.Matches, entry)
	}
	result.Total = len(result.Matches)
	if len(result.Matches) > 0 {
		fillExitCode(&result, result.Matches[0])
	}
	return tools.Result{Data: result}, nil
}

// exitCodeInType 判断条目是否落在指定类别；空类别一律放行。
func exitCodeInType(entry ExitCodeEntry, kind string) bool {
	return kind == "" || entry.Type == kind
}

// fillExitCode 用命中的首条填充摘要字段。
func fillExitCode(result *ExitCodeResult, entry ExitCodeEntry) {
	result.Matched = true
	result.Code = entry.Code
	result.Label = entry.Label
	result.Description = entry.Description
	result.Type = entry.Type
}

// matchesExitCode 判断条目是否命中关键字。
//
// 匹配范围：简短名称与中文说明。
func matchesExitCode(entry ExitCodeEntry, lowered string) bool {
	if lowered == "" {
		return false
	}
	if strings.Contains(strings.ToLower(entry.Label), lowered) {
		return true
	}
	return strings.Contains(entry.Description, lowered)
}

// Validate 做字段级校验。
func (r *ExitCodeRequest) Validate() error {
	if len(r.Query) > 64 {
		return tools.InvalidField("query", []int{0, 64}, len(r.Query))
	}
	if r.Type != "" && !containsString(ExitCodeTypes(), r.Type) {
		return tools.InvalidField("type", ExitCodeTypes(), r.Type)
	}
	return nil
}
