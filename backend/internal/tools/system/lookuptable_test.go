package system

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestLookupTablePort 断言端口查询命中。
func TestLookupTablePort(t *testing.T) {
	result, err := LookupTable(t.Context(), &LookupTableRequest{
		Kind: LookupKindPort,
		Key:  "443",
	})
	data := result.Data.(LookupTableResult)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if !data.Matched {
		t.Fatal("443 应命中")
	}
	if data.Entry == nil {
		t.Fatal("应给出条目")
	}
	if data.DataVersion == "" {
		t.Error("应给出数据集版本")
	}
}

// TestLookupTableFuzzy 断言模糊检索给出候选。
func TestLookupTableFuzzy(t *testing.T) {
	result, err := LookupTable(t.Context(), &LookupTableRequest{
		Kind: LookupKindExitCode,
		Key:  "错误",
	})
	data := result.Data.(LookupTableResult)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(data.Candidates) == 0 {
		t.Error("应给出候选")
	}
}

// TestLookupTableMissingKey 断言缺少查询词返回 MISSING_FIELD。
func TestLookupTableMissingKey(t *testing.T) {
	_, err := LookupTable(t.Context(), &LookupTableRequest{Kind: LookupKindPort})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestLookupTableUnknownKind 断言未知类别返回 INVALID_FIELD。
func TestLookupTableUnknownKind(t *testing.T) {
	_, err := LookupTable(t.Context(), &LookupTableRequest{Kind: "unknown", Key: "1"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
