package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestRegexTesterMatch 断言基本匹配行为。
func TestRegexTesterMatch(t *testing.T) {
	result, err := RegexTester(&RegexTesterRequest{
		Pattern: `(\w+)@(\w+)\.com`,
		Input:   TextInput{Data: "mail: alice@example.com and bob@test.com"},
	})
	if err != nil {
		t.Fatalf("匹配失败：%v", err)
	}
	if result.Count != 2 {
		t.Fatalf("匹配数 = %d，期望 2", result.Count)
	}
	if result.Matches[0].Text != "alice@example.com" {
		t.Errorf("首条匹配 = %q", result.Matches[0].Text)
	}
	if len(result.Matches[0].Groups) != 2 {
		t.Fatalf("捕获组数 = %d，期望 2", len(result.Matches[0].Groups))
	}
	if result.Matches[0].Groups[1] != "example" {
		t.Errorf("第二捕获组 = %q，期望 example", result.Matches[0].Groups[1])
	}
	if result.Matches[0].Line != 1 || result.Matches[0].Column != 7 {
		t.Errorf("行列 = %d:%d，期望 1:7", result.Matches[0].Line, result.Matches[0].Column)
	}
}

// TestRegexTesterFlags 断言标志生效。
func TestRegexTesterFlags(t *testing.T) {
	result, err := RegexTester(&RegexTesterRequest{
		Pattern: `abc`,
		Flags:   []string{RegexFlagIgnoreCase},
		Input:   TextInput{Data: "ABC"},
	})
	if err != nil {
		t.Fatalf("匹配失败：%v", err)
	}
	if result.Count != 1 {
		t.Fatalf("匹配数 = %d，期望 1", result.Count)
	}
}

// TestRegexTesterInvalidPattern 断言非法表达式返回 INVALID_REGEX。
func TestRegexTesterInvalidPattern(t *testing.T) {
	_, err := RegexTester(&RegexTesterRequest{
		Pattern: `a(b`,
		Input:   TextInput{Data: "abc"},
	})
	assertCode(t, err, toolerr.CodeInvalidRegex)
}

// TestRegexTesterLookaheadRejected 断言前后向断言被拒绝。
//
// RE2 不支持前向断言，编译期即失败。
func TestRegexTesterLookaheadRejected(t *testing.T) {
	_, err := RegexTester(&RegexTesterRequest{
		Pattern: `a(?=b)`,
		Input:   TextInput{Data: "ab"},
	})
	assertCode(t, err, toolerr.CodeInvalidRegex)
}

// TestRegexTesterTruncate 断言结果按上限截断。
func TestRegexTesterTruncate(t *testing.T) {
	result, err := RegexTester(&RegexTesterRequest{
		Pattern:    "a",
		Input:      TextInput{Data: "aaaa"},
		MaxMatches: 2,
	})
	if err != nil {
		t.Fatalf("匹配失败：%v", err)
	}
	if result.Count != 2 {
		t.Fatalf("匹配数 = %d，期望 2", result.Count)
	}
	if !result.Truncated {
		t.Error("期望 truncated 为真")
	}
}

// TestRegexTesterMissingInput 断言缺少输入时报 MISSING_FIELD。
func TestRegexTesterMissingInput(t *testing.T) {
	_, err := RegexTester(&RegexTesterRequest{
		Pattern: "a",
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestRegexTesterUnknownFlag 断言未知标志报 INVALID_FIELD。
func TestRegexTesterUnknownFlag(t *testing.T) {
	_, err := RegexTester(&RegexTesterRequest{
		Pattern: "a",
		Flags:   []string{"z"},
		Input:   TextInput{Data: "abc"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestHasDialectDifference 断言方言差异检测。
func TestHasDialectDifference(t *testing.T) {
	cases := []struct {
		pattern string
		want    bool
	}{
		{`a(?=b)`, true},
		{`a(?!b)`, true},
		{`(?<=a)b`, true},
		{`(?<!a)b`, true},
		{`\d+`, false},
		{`^abc$`, false},
	}
	for _, tc := range cases {
		t.Run(tc.pattern, func(t *testing.T) {
			if got := hasDialectDifference(tc.pattern); got != tc.want {
				t.Errorf("hasDialectDifference(%q) = %v，期望 %v", tc.pattern, got, tc.want)
			}
		})
	}
}

// assertCode 断言错误的错误码。
func assertCode(t *testing.T, err error, want toolerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望错误 %s，实际无错误", want)
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("期望 *toolerr.Error，实际 %T", err)
	}
	if toolErr.Code != want {
		t.Errorf("错误码 = %s，期望 %s", toolErr.Code, want)
	}
}
