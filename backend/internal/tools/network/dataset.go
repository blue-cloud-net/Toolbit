package network

import (
	"bytes"
	"embed"
	"encoding/csv"
	"io"
	"strings"
)

// datasetFS 内嵌离线数据集。
//
// 这里是**唯一**的数据来源：MIME 类型与 HTTP 状态码都是静态表，
// **不**在运行时出网，也不从系统 MIME 数据库读取（各发行版不一致）。
//
//go:embed dataset/*.csv
var datasetFS embed.FS

// 数据集版本。
//
// 快照日期决定数据集内容：条目调整后**必须**更新快照与版本号，
// 否则响应里的版本号会误导用户。
const (
	// MimeTypesVersion 是 MIME 类型数据集的快照版本。
	MimeTypesVersion = "2026-10"
	// HTTPStatusVersion 是 HTTP 状态码数据集的快照版本。
	HTTPStatusVersion = "2026-10"
	// DNSServersVersion 是 DNS 服务器目录的快照版本。
	DNSServersVersion = "2026-11"
)

// readDataset 读取内嵌数据集文件。
//
// 内嵌文件在编译期就已确定存在；读取失败按空数据集处理，
// 让查询返回空结果而不是让整个进程崩溃。
func readDataset(name string) io.Reader {
	raw, err := datasetFS.ReadFile(name)
	if err != nil {
		return strings.NewReader("")
	}
	return bytes.NewReader(raw)
}

// readCSV 读取 CSV 并跳过表头。
func readCSV(r io.Reader) [][]string {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return nil
	}
	return records[1:]
}

// csvField 读取 CSV 记录的第 i 列；越界时返回空串。
func csvField(rec []string, i int) string {
	if i >= len(rec) {
		return ""
	}
	return rec[i]
}

// splitList 按分号拆分列表项，忽略空白项。
func splitList(raw string) []string {
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
