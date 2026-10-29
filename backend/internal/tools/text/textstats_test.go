package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestTextStatsCounts 断言中英文分别计数。
func TestTextStatsCounts(t *testing.T) {
	result, err := TextStats(t.Context(), &TextStatsRequest{
		Text: "Hello 世界\n第二行",
	})
	data := result.Data.(TextStatsResult)
	if err != nil {
		t.Fatalf("统计失败：%v", err)
	}
	if data.Lines != 2 {
		t.Errorf("行数 = %d，期望 2", data.Lines)
	}
	if data.ChineseChars != 5 {
		t.Errorf("中文字符数 = %d，期望 5", data.ChineseChars)
	}
	if data.EnglishWords != 1 {
		t.Errorf("英文单词数 = %d，期望 1", data.EnglishWords)
	}
}

// TestTextStatsEmptyText 断言空文本返回 MISSING_FIELD。
func TestTextStatsEmptyText(t *testing.T) {
	_, err := TextStats(t.Context(), &TextStatsRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}
