package data

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestMockDataGenerates 断言按类型生成假数据。
func TestMockDataGenerates(t *testing.T) {
	result, err := MockData(t.Context(), &MockDataRequest{
		Types: []string{MockTypeString, MockTypeInt},
		Count: 3,
	})
	data := result.Data.(MockDataResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if data.Count != 3 {
		t.Errorf("每类数量 = %d，期望 3", data.Count)
	}
	// 两种类型各生成 3 条，共 6 条。
	if len(data.Fields) != 6 {
		t.Errorf("字段数 = %d，期望 6", len(data.Fields))
	}
}

// TestMockDataEmail 断言邮箱类型产出合法邮箱。
func TestMockDataEmail(t *testing.T) {
	result, err := MockData(t.Context(), &MockDataRequest{
		Types: []string{MockTypeEmail},
		Count: 1,
	})
	data := result.Data.(MockDataResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.Fields) != 1 {
		t.Fatalf("字段数 = %d，期望 1", len(data.Fields))
	}
	if data.Fields[0].Type != MockTypeEmail {
		t.Errorf("字段类型 = %q", data.Fields[0].Type)
	}
}

// TestMockDataUnsupportedLocale 断言未知语言返回 INVALID_FIELD。
func TestMockDataUnsupportedLocale(t *testing.T) {
	_, err := MockData(t.Context(), &MockDataRequest{
		Types:  []string{MockTypeString},
		Locale: "fr-FR",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestMockDataTooManyCount 断言数量超限返回 INVALID_FIELD。
func TestMockDataTooManyCount(t *testing.T) {
	_, err := MockData(t.Context(), &MockDataRequest{
		Types: []string{MockTypeString},
		Count: 100000,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
