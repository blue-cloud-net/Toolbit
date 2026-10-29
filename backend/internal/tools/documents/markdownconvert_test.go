package documents

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestMarkdownConvertToHTML 断言 Markdown 转 HTML。
func TestMarkdownConvertToHTML(t *testing.T) {
	result, err := MarkdownConvert(t.Context(), &MarkdownConvertRequest{
		Mode:  MarkdownToHTML,
		Input: "# Title\n\nBody",
		GFM:   true,
	})
	data := result.Data.(MarkdownConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if !strings.Contains(data.Output, "<h1>") {
		t.Errorf("输出应含标题标签：%s", data.Output)
	}
	if data.HeadingCount != 1 {
		t.Errorf("标题数 = %d，期望 1", data.HeadingCount)
	}
}

// TestMarkdownConvertToMarkdown 断言 HTML 转 Markdown。
func TestMarkdownConvertToMarkdown(t *testing.T) {
	result, err := MarkdownConvert(t.Context(), &MarkdownConvertRequest{
		Mode:  HTMLToMarkdown,
		Input: "<h1>Title</h1><p>Body</p>",
	})
	data := result.Data.(MarkdownConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if !strings.Contains(data.Output, "# Title") {
		t.Errorf("输出应含 Markdown 标题：%s", data.Output)
	}
}

// TestMarkdownConvertUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestMarkdownConvertUnsupportedMode(t *testing.T) {
	_, err := MarkdownConvert(t.Context(), &MarkdownConvertRequest{Mode: "toPdf", Input: "# a"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
