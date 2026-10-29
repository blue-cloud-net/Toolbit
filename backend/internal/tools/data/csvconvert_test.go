package data

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCSVConvertCSVToJSON 断言 CSV 转 JSON。
func TestCSVConvertCSVToJSON(t *testing.T) {
	result, err := CSVConvert(t.Context(), &CSVConvertRequest{
		Operation: CSVToJSON,
		Input:     "name,age\nalice,30\nbob,25\n",
		HasHeader: true,
	})
	data := result.Data.(CSVConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.RowCount != 2 {
		t.Errorf("行数 = %d，期望 2", data.RowCount)
	}
	if data.ColumnCount != 2 {
		t.Errorf("列数 = %d，期望 2", data.ColumnCount)
	}
	if len(data.Columns) != 2 || data.Columns[0] != "name" {
		t.Errorf("列名 = %v", data.Columns)
	}
}

// TestCSVConvertJSONToCSV 断言 JSON 转 CSV。
func TestCSVConvertJSONToCSV(t *testing.T) {
	result, err := CSVConvert(t.Context(), &CSVConvertRequest{
		Operation: JSONToCSV,
		Input:     `[{"name":"alice","age":30}]`,
		HasHeader: true,
	})
	data := result.Data.(CSVConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.ColumnCount != 2 {
		t.Errorf("列数 = %d，期望 2", data.ColumnCount)
	}
}

// TestCSVConvertUnsupportedOperation 断言未知方向返回 INVALID_FIELD。
func TestCSVConvertUnsupportedOperation(t *testing.T) {
	_, err := CSVConvert(t.Context(), &CSVConvertRequest{
		Operation: "xmlToJson",
		Input:     "a",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestCSVConvertMissingInput 断言缺少输入返回 MISSING_FIELD。
func TestCSVConvertMissingInput(t *testing.T) {
	_, err := CSVConvert(t.Context(), &CSVConvertRequest{Operation: CSVToJSON})
	assertCode(t, err, toolerr.CodeMissingField)
}
