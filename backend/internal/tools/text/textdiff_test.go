package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestTextDiffIdentical 断言两侧一致时不产生差异。
func TestTextDiffIdentical(t *testing.T) {
	result, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode:  DiffModeText,
		Left:  DiffInput{Data: "a\nb\nc\n"},
		Right: DiffInput{Data: "a\nb\nc\n"},
	})
	data := result.Data.(TextDiffResult)
	if err != nil {
		t.Fatalf("对比失败：%v", err)
	}
	if !data.Identical {
		t.Error("两侧一致时 identical 应为 true")
	}
	if data.Stat.Changes != 0 {
		t.Errorf("差异数 = %d，期望 0", data.Stat.Changes)
	}
}

// TestTextDiffDetectsChange 断言行级差异被标出。
func TestTextDiffDetectsChange(t *testing.T) {
	result, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode:  DiffModeText,
		Left:  DiffInput{Data: "a\nb\nc\n"},
		Right: DiffInput{Data: "a\nx\nc\n"},
	})
	data := result.Data.(TextDiffResult)
	if err != nil {
		t.Fatalf("对比失败：%v", err)
	}
	if data.Identical {
		t.Error("两侧不同��� identical 应为 false")
	}
	if data.Stat.Added != 1 || data.Stat.Removed != 1 {
		t.Errorf("增删行数 = %d/%d，期望 1/1", data.Stat.Added, data.Stat.Removed)
	}
}

// TestTextDiffJSONPaths 断言 JSON 模式给出结构化路径。
func TestTextDiffJSONPaths(t *testing.T) {
	result, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode:  DiffModeJSON,
		Left:  DiffInput{Data: `{"a":1,"b":{"c":2}}`},
		Right: DiffInput{Data: `{"a":1,"b":{"c":3}}`},
	})
	data := result.Data.(TextDiffResult)
	if err != nil {
		t.Fatalf("对比失败：%v", err)
	}
	if len(data.Changes) == 0 {
		t.Fatal("JSON 模式应给出结构化差异")
	}
	if data.Changes[0].Path == "" {
		t.Error("结构化差异应带路径")
	}
}

// TestTextDiffMissingSide 断言缺少一侧返回 MISSING_FIELD。
func TestTextDiffMissingSide(t *testing.T) {
	_, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode:  DiffModeText,
		Right: DiffInput{Data: "a"},
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestTextDiffUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestTextDiffUnsupportedMode(t *testing.T) {
	_, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode:  "xml",
		Left:  DiffInput{Data: "a"},
		Right: DiffInput{Data: "b"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
