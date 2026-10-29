package data

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CSV 转换方向。
const (
	// CSVToJSON 是 CSV 转 JSON。
	CSVToJSON = "csvToJson"
	// JSONToCSV 是 JSON 转 CSV。
	JSONToCSV = "jsonToCsv"
)

// CSVOperation 返回受支持的转换方向（顺序稳定）。
func CSVOperation() []string {
	return []string{CSVToJSON, JSONToCSV}
}

// CSVConvertRequest 是 csv-convert 工具的请求。
type CSVConvertRequest struct {
	// Operation 是转换方向：csvToJson / jsonToCsv；默认 csvToJson。
	Operation string `json:"operation"`
	// Input 是待转换的内容。
	Input string `json:"input"`
	// Delimiter 是 CSV 分隔符；默认逗号。
	Delimiter string `json:"delimiter"`
	// HasHeader 是否首行为表头；默认 true。
	HasHeader bool `json:"hasHeader"`
	// Pretty 是否输出缩进 JSON。
	Pretty bool `json:"pretty"`
}

// CSVConvertResult 是 csv-convert 工具的响应。
type CSVConvertResult struct {
	// Operation 是实际使用的转换方向。
	Operation string `json:"operation"`
	// Output 是转换结果。
	Output string `json:"output"`
	// RowCount 是数据行数（不含表头）。
	RowCount int `json:"rowCount"`
	// ColumnCount 是列数。
	ColumnCount int `json:"columnCount"`
	// Columns 是列名；无表头时为空。
	Columns []string `json:"columns"`
}

// maxCSVBytes 是待转换内容的字节上限。
const maxCSVBytes = 1024 * 1024

// CSVConvert 在 CSV 与 JSON 之间转换。
func CSVConvert(_ context.Context, req *CSVConvertRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	operation := req.Operation
	if operation == "" {
		operation = CSVToJSON
	}

	if operation == JSONToCSV {
		return jsonToCSV(req)
	}
	return csvToJSON(req)
}

// csvToJSON 把 CSV 转成 JSON 数组。
func csvToJSON(req *CSVConvertRequest) (tools.Result, error) {
	delimiter := req.Delimiter
	if delimiter == "" {
		delimiter = ","
	}

	reader := csv.NewReader(strings.NewReader(req.Input))
	reader.Comma = rune(delimiter[0])
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return tools.Result{}, tools.InvalidField("input", "valid CSV", "parse error")
	}
	if len(records) == 0 {
		return tools.Result{Data: CSVConvertResult{
			Operation:   CSVToJSON,
			Output:      "[]",
			RowCount:    0,
			ColumnCount: 0,
			Columns:     []string{},
		}}, nil
	}

	columns := []string{}
	rows := []map[string]any{}
	dataStart := 0

	if req.HasHeader {
		columns = records[0]
		dataStart = 1
	} else {
		// 无表头时按列序号生成列名，保证 JSON 键稳定。
		for i := range records[0] {
			columns = append(columns, "column"+itoa(i+1))
		}
	}

	for _, record := range records[dataStart:] {
		row := map[string]any{}
		for i, cell := range record {
			key := "column" + itoa(i+1)
			if i < len(columns) {
				key = columns[i]
			}
			row[key] = cell
		}
		rows = append(rows, row)
	}

	output, err := marshalJSONValue(rows, req.Pretty)
	if err != nil {
		return tools.Result{}, tools.InvalidField("input", "valid CSV", "encode error")
	}

	return tools.Result{Data: CSVConvertResult{
		Operation:   CSVToJSON,
		Output:      output,
		RowCount:    len(rows),
		ColumnCount: len(columns),
		Columns:     columns,
	}}, nil
}

// jsonToCSV 把 JSON 数组转成 CSV。
func jsonToCSV(req *CSVConvertRequest) (tools.Result, error) {
	delimiter := req.Delimiter
	if delimiter == "" {
		delimiter = ","
	}

	decoder := json.NewDecoder(strings.NewReader(req.Input))
	var value any
	if err := decoder.Decode(&value); err != nil {
		return tools.Result{}, tools.InvalidField("input", "valid JSON", "parse error")
	}

	rows, ok := value.([]any)
	if !ok {
		return tools.Result{}, tools.InvalidField("input", "JSON array", "not an array")
	}

	columns := []string{}
	seen := map[string]bool{}
	for _, row := range rows {
		object, ok := row.(map[string]any)
		if !ok {
			continue
		}
		for key := range object {
			if !seen[key] {
				seen[key] = true
				columns = append(columns, key)
			}
		}
	}
	// 列名排序让输出稳定。
	sortStrings(columns)

	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	writer.Comma = rune(delimiter[0])

	if req.HasHeader {
		if err := writer.Write(columns); err != nil {
			return tools.Result{}, tools.InvalidField("input", "valid JSON", "encode error")
		}
	}

	for _, row := range rows {
		object, ok := row.(map[string]any)
		if !ok {
			continue
		}
		record := make([]string, len(columns))
		for i, column := range columns {
			record[i] = formatJSONScalar(object[column])
		}
		if err := writer.Write(record); err != nil {
			return tools.Result{}, tools.InvalidField("input", "valid JSON", "encode error")
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return tools.Result{}, tools.InvalidField("input", "valid JSON", "encode error")
	}

	return tools.Result{Data: CSVConvertResult{
		Operation:   JSONToCSV,
		Output:      builder.String(),
		RowCount:    len(rows),
		ColumnCount: len(columns),
		Columns:     columns,
	}}, nil
}

// formatJSONScalar 把 JSON 值转成 CSV 单元格文本。
func formatJSONScalar(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case float64:
		return formatJSONNumber(typed)
	case json.Number:
		return typed.String()
	default:
		raw, err := json.Marshal(value)
		if err != nil {
			return ""
		}
		return string(raw)
	}
}

// formatJSONNumber 把浮点数转成不带多余小数的文本。
func formatJSONNumber(value float64) string {
	if value == float64(int64(value)) {
		return itoa(int(int64(value)))
	}
	return strings.TrimRight(strings.TrimRight(formatFloat(value), "0"), ".")
}

// formatFloat 渲染浮点数。
func formatFloat(value float64) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "0"
	}
	return string(raw)
}

// marshalJSONValue 序列化 JSON 值。
func marshalJSONValue(value any, pretty bool) (string, error) {
	if pretty {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// sortStrings 原地排序字符串切片。
func sortStrings(items []string) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j] < items[j-1]; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

// itoa 把整数转成字符串。
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// Validate 做字段级校验。
func (r *CSVConvertRequest) Validate() error {
	operation := r.Operation
	if operation == "" {
		operation = CSVToJSON
	}
	if !containsString(CSVOperation(), operation) {
		return tools.InvalidField("operation", CSVOperation(), operation)
	}
	if strings.TrimSpace(r.Input) == "" {
		return tools.MissingField("input")
	}
	if len(r.Input) > maxCSVBytes {
		return tools.InvalidField("input", []int{0, maxCSVBytes}, len(r.Input))
	}
	if r.Delimiter != "" {
		length := len([]rune(r.Delimiter))
		if length != 1 {
			return tools.InvalidField("delimiter", "single character", length)
		}
	}
	return nil
}
