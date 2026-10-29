package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestRegexBuildBasic 断言基础片段拼装。
func TestRegexBuildBasic(t *testing.T) {
	result, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts: []string{RegexPartDigits, RegexPartHyphen, RegexPartDigits},
	})
	data := result.Data.(RegexBuildResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !data.Valid {
		t.Error("生成的表达式应合法")
	}
	if data.Pattern == "" {
		t.Error("应给出表达式")
	}
	if len(data.Explanation) == 0 {
		t.Error("应给出逐段释义")
	}
}

// TestRegexBuildQuantifier 断言量词生效。
func TestRegexBuildQuantifier(t *testing.T) {
	result, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts:      []string{RegexPartLetters},
		Quantifier: &RegexQuantifier{Min: 2, Max: 8},
		Anchored:   true,
	})
	data := result.Data.(RegexBuildResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if data.Pattern == "" {
		t.Error("应给出表达式")
	}
}

// TestRegexBuildUnknownPart 断言未知片段返回 INVALID_FIELD。
func TestRegexBuildUnknownPart(t *testing.T) {
	_, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts: []string{"not-a-part"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestRegexBuildInvertedQuantifier 断言量词上下界颠倒被拒。
func TestRegexBuildInvertedQuantifier(t *testing.T) {
	_, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts:      []string{RegexPartLetters},
		Quantifier: &RegexQuantifier{Min: 5, Max: 2},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
