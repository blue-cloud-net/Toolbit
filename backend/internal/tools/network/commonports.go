package network

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 常用端口的服务分类。
const (
	// PortCategoryWeb 是 Web 服务。
	PortCategoryWeb = "web"
	// PortCategoryMail 是邮件服务。
	PortCategoryMail = "mail"
	// PortCategoryFile 是文件传输与共享。
	PortCategoryFile = "file"
	// PortCategoryRemote 是远程登录与桌面。
	PortCategoryRemote = "remote"
	// PortCategoryDatabase 是数据库与缓存。
	PortCategoryDatabase = "database"
	// PortCategoryMessaging 是消息队列。
	PortCategoryMessaging = "messaging"
	// PortCategoryDirectory 是目录与认证服务。
	PortCategoryDirectory = "directory"
	// PortCategoryNetwork 是网络基础设施。
	PortCategoryNetwork = "network"
	// PortCategoryDevops 是容器与运维平台。
	PortCategoryDevops = "devops"
)

// 查询方式。
const (
	// PortQueryAll 表示未给出查询条件，返回全量列表。
	PortQueryAll = "all"
	// PortQueryPort 表示按端口号查询。
	PortQueryPort = "port"
	// PortQueryKeyword 表示按关键字查询。
	PortQueryKeyword = "keyword"
)

// CommonPortsVersion 是常用端口数据集的快照版本。
//
// 端口条目调整后**必须**更新版本号，否则响应里的版本号会误导用户。
const CommonPortsVersion = "2026-11"

// CommonPortCategories 返回常用端口的服务分类（顺序稳定）。
func CommonPortCategories() []string {
	return []string{
		PortCategoryWeb,
		PortCategoryMail,
		PortCategoryFile,
		PortCategoryRemote,
		PortCategoryDatabase,
		PortCategoryMessaging,
		PortCategoryDirectory,
		PortCategoryNetwork,
		PortCategoryDevops,
	}
}

// CommonPortEntry 是一个常用端口条目。
type CommonPortEntry struct {
	// Port 是端口号。
	Port int `json:"port"`
	// Label 是服务名称。
	Label string `json:"label"`
	// Description 是中文说明。
	Description string `json:"description"`
	// Protocols 是该端口使用的传输协议。
	Protocols []string `json:"protocols"`
	// Category 是服务分类。
	Category string `json:"category"`
}

// CommonPortRequest 是 common-ports 工具的请求。
//
// 字段都可以省略：都不给时返回**全量**端口列表，
// 便于前端默认展示完整列表、再按输入本地过滤。
type CommonPortRequest struct {
	// Query 是查询词：端口号（如 443）或关键字（如 https）。
	Query string `json:"query"`
	// Category 是服务分类过滤；为空表示不过滤。
	Category string `json:"category"`
}

// CommonPortResult 是 common-ports 工具的响应。
type CommonPortResult struct {
	// Query 是规范化后的查询词。
	Query string `json:"query"`
	// Kind 是查询方式：all / port / keyword。
	Kind string `json:"kind"`
	// Matched 是否命中端口条目。
	Matched bool `json:"matched"`
	// Port 是命中的端口号；未命中为 `0`。
	Port int `json:"port"`
	// Label 是命中的服务名称。
	Label string `json:"label"`
	// Description 是命中的中文说明。
	Description string `json:"description"`
	// Protocols 是命中的传输协议。
	Protocols []string `json:"protocols"`
	// Category 是命中的服务分类。
	Category string `json:"category"`
	// Matches 是命中的条目列表。
	Matches []CommonPortEntry `json:"matches"`
	// Total 是命中条数。
	Total int `json:"total"`
	// DatasetTotal 是数据集条目数。
	DatasetTotal int `json:"datasetTotal"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// 常用端口数据集缓存。
var (
	commonPortsOnce   sync.Once
	commonPortEntries []CommonPortEntry
	commonPortsByPort map[int]*CommonPortEntry
)

// loadCommonPorts 解析并缓存常用端口数据集。
func loadCommonPorts() {
	commonPortsOnce.Do(func() {
		commonPortEntries = parseCommonPorts(readDataset("dataset/common_ports.csv"))
		commonPortsByPort = make(map[int]*CommonPortEntry, len(commonPortEntries))
		for i := range commonPortEntries {
			entry := &commonPortEntries[i]
			commonPortsByPort[entry.Port] = entry
		}
	})
}

// parseCommonPorts 解析常用端口数据集。
func parseCommonPorts(r io.Reader) []CommonPortEntry {
	out := []CommonPortEntry{}
	for _, rec := range readCSV(r) {
		port, err := strconv.Atoi(strings.TrimSpace(csvField(rec, 0)))
		if err != nil {
			continue
		}
		out = append(out, CommonPortEntry{
			Port:        port,
			Label:       strings.TrimSpace(csvField(rec, 1)),
			Description: strings.TrimSpace(csvField(rec, 2)),
			Protocols:   splitList(csvField(rec, 3)),
			Category:    strings.TrimSpace(csvField(rec, 4)),
		})
	}
	return out
}

// CommonPorts 查询常用端口的服务名、协议与用途。
//
// 数据集是**离线内置**的快照，**不**出网。
func CommonPorts(_ context.Context, req *CommonPortRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	loadCommonPorts()

	category := strings.TrimSpace(req.Category)
	result := CommonPortResult{
		Query:        strings.TrimSpace(req.Query),
		Kind:         PortQueryAll,
		Protocols:    []string{},
		Matches:      []CommonPortEntry{},
		DatasetTotal: len(commonPortEntries),
		DataVersion:  CommonPortsVersion,
	}

	query := strings.TrimSpace(req.Query)
	if query == "" && category == "" {
		result.Matches = append([]CommonPortEntry{}, commonPortEntries...)
		result.Total = len(result.Matches)
		return tools.Result{Data: result}, nil
	}

	if query != "" && isAllDigits(query) {
		port, err := strconv.Atoi(query)
		if err != nil {
			return tools.Result{}, tools.InvalidField("query", "端口号或关键字", query)
		}
		result.Kind = PortQueryPort
		result.Query = strconv.Itoa(port)
		if entry, ok := commonPortsByPort[port]; ok && portInCategory(*entry, category) {
			result.Matches = []CommonPortEntry{*entry}
			result.Total = 1
			fillCommonPort(&result, *entry)
		}
		return tools.Result{Data: result}, nil
	}

	result.Kind = PortQueryKeyword
	lowered := strings.ToLower(query)
	for _, entry := range commonPortEntries {
		if !portInCategory(entry, category) {
			continue
		}
		if query != "" && !matchesCommonPort(entry, lowered) {
			continue
		}
		result.Matches = append(result.Matches, entry)
	}
	result.Total = len(result.Matches)
	if len(result.Matches) > 0 {
		fillCommonPort(&result, result.Matches[0])
	}
	return tools.Result{Data: result}, nil
}

// portInCategory 判断条目是否落在指定分类；空分类一律放行。
func portInCategory(entry CommonPortEntry, category string) bool {
	return category == "" || entry.Category == category
}

// fillCommonPort 用命中的首条填充摘要字段。
func fillCommonPort(result *CommonPortResult, entry CommonPortEntry) {
	result.Matched = true
	result.Port = entry.Port
	result.Label = entry.Label
	result.Description = entry.Description
	result.Protocols = entry.Protocols
	result.Category = entry.Category
}

// matchesCommonPort 判断条目是否命中关键字。
//
// 匹配范围：端口号、服务名、中文说明与传输协议。
func matchesCommonPort(entry CommonPortEntry, lowered string) bool {
	if lowered == "" {
		return false
	}
	if strings.Contains(strings.ToLower(entry.Label), lowered) {
		return true
	}
	if strings.Contains(entry.Description, lowered) {
		return true
	}
	for _, protocol := range entry.Protocols {
		if strings.Contains(strings.ToLower(protocol), lowered) {
			return true
		}
	}
	return false
}

// Validate 做字段级校验。
func (r *CommonPortRequest) Validate() error {
	if len(r.Query) > 64 {
		return tools.InvalidField("query", []int{0, 64}, len(r.Query))
	}
	if r.Category != "" && !containsString(CommonPortCategories(), r.Category) {
		return tools.InvalidField("category", CommonPortCategories(), r.Category)
	}
	return nil
}
