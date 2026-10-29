package documents

import (
	"context"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestMarkdownPreviewBasic 断言基本渲染。
func TestMarkdownPreviewBasic(t *testing.T) {
	result, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "# 标题\n\n正文 **加粗**。",
	})
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	if !strings.Contains(result.HTML, "<h1") {
		t.Errorf("HTML 缺少标题：%s", result.HTML)
	}
	if !strings.Contains(result.HTML, "<strong>加粗</strong>") {
		t.Errorf("HTML 缺少加粗：%s", result.HTML)
	}
}

// TestMarkdownPreviewTOC 断言目录生成。
func TestMarkdownPreviewTOC(t *testing.T) {
	result, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "# 一级\n\n## 二级\n\n### 三级",
	})
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	if len(result.TOC) != 3 {
		t.Fatalf("目录条目数 = %d，期望 3", len(result.TOC))
	}
	if result.TOC[0].Level != 1 || result.TOC[0].Text != "一级" {
		t.Errorf("一级目录不正确：%+v", result.TOC[0])
	}
	if result.TOC[0].ID == "" {
		t.Error("目录缺少锚点")
	}
	if !strings.Contains(result.HTML, `id="`+result.TOC[0].ID+`"`) {
		t.Errorf("HTML 锚点与目录不一致：%s", result.HTML)
	}
}

// TestMarkdownPreviewSanitizesScript 断言脚本被剥离。
func TestMarkdownPreviewSanitizesScript(t *testing.T) {
	result, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "<script>alert(1)</script>\n\n正常内容",
	})
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	if strings.Contains(result.HTML, "<script") {
		t.Errorf("脚本未被剥离：%s", result.HTML)
	}
	if strings.Contains(result.HTML, "alert(1)") {
		t.Errorf("脚本内容仍在：%s", result.HTML)
	}
}

// TestMarkdownPreviewDisallowsRawHTML 断言默认转义原始 HTML。
func TestMarkdownPreviewDisallowsRawHTML(t *testing.T) {
	result, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "<b onclick=\"evil()\">粗体</b>",
	})
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	if strings.Contains(result.HTML, "onclick") {
		t.Errorf("事件属性未被剥离：%s", result.HTML)
	}
}

// TestMarkdownPreviewStripsJavascriptURL 断言 javascript: URL 被剥离。
func TestMarkdownPreviewStripsJavascriptURL(t *testing.T) {
	result, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "[点击](javascript:alert(1))",
	})
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	if strings.Contains(result.HTML, "javascript:") {
		t.Errorf("javascript: URL 未被剥离：%s", result.HTML)
	}
}

// TestMarkdownPreviewGFMTable 断言 GFM 表格。
func TestMarkdownPreviewGFMTable(t *testing.T) {
	result, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "| a | b |\n| --- | --- |\n| 1 | 2 |",
	})
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	if !strings.Contains(result.HTML, "<table>") {
		t.Errorf("缺少表格：%s", result.HTML)
	}
}

// TestMarkdownPreviewCommonMark 断言 CommonMark 不渲染表格。
func TestMarkdownPreviewCommonMark(t *testing.T) {
	result, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "| a | b |\n| --- | --- |\n| 1 | 2 |",
		Flavor:   FlavorCommonMark,
	})
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	if strings.Contains(result.HTML, "<table>") {
		t.Errorf("CommonMark 不应渲染表格：%s", result.HTML)
	}
}

// TestMarkdownPreviewMissingSource 断言缺少源文时报 MISSING_FIELD。
func TestMarkdownPreviewMissingSource(t *testing.T) {
	_, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestMarkdownPreviewInvalidFlavor 断言未知方言报 INVALID_FIELD。
func TestMarkdownPreviewInvalidFlavor(t *testing.T) {
	_, err := MarkdownPreview(context.Background(), &MarkdownPreviewRequest{
		Markdown: "# x",
		Flavor:   "markdown",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
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
