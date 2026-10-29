package data

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestJSONToTypesTypeScript 断言生成 TypeScript 类型。
func TestJSONToTypesTypeScript(t *testing.T) {
	result, err := JSONToTypes(t.Context(), &JSONToTypesRequest{
		Input:    `{"id":1,"name":"a"}`,
		Language: LangTypeScript,
		RootName: "User",
	})
	data := result.Data.(JSONToTypesResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Output, "User") {
		t.Errorf("输出应含根类型名：%s", data.Output)
	}
	if data.TypeCount == 0 {
		t.Error("类型数应大于 0")
	}
}

// TestJSONToTypesGo 断言生成 Go 类型。
func TestJSONToTypesGo(t *testing.T) {
	result, err := JSONToTypes(t.Context(), &JSONToTypesRequest{
		Input:    `{"id":1}`,
		Language: LangGo,
		RootName: "Item",
	})
	data := result.Data.(JSONToTypesResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Output, "type Item struct") {
		t.Errorf("输出应含结构体定义：%s", data.Output)
	}
}

// TestJSONToTypesUnsupportedLanguage 断言未知语言返回 INVALID_FIELD。
func TestJSONToTypesUnsupportedLanguage(t *testing.T) {
	_, err := JSONToTypes(t.Context(), &JSONToTypesRequest{
		Input:    "{}",
		Language: "cobol",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestJSONToTypesMissingInput 断言缺少输入返回 MISSING_FIELD。
func TestJSONToTypesMissingInput(t *testing.T) {
	_, err := JSONToTypes(t.Context(), &JSONToTypesRequest{Language: LangGo})
	assertCode(t, err, toolerr.CodeMissingField)
}
