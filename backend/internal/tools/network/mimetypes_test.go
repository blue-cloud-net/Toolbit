package network

import (
	"context"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// mimeLookup 执行一次 MIME 查询并取出结果。
func mimeLookup(t *testing.T, req *MimeLookupRequest) MimeLookupResult {
	t.Helper()
	result, err := MimeLookup(context.Background(), req)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	out, ok := result.Data.(MimeLookupResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

func TestMimeLookupByExtension(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{Query: "png"})
	if out.Total == 0 {
		t.Fatal("png 应当命中")
	}
	entry := out.Entries[0]
	if entry.MimeType != "image/png" {
		t.Errorf("首条 mimeType = %s，期望 image/png", entry.MimeType)
	}
	if entry.Category != "image" {
		t.Errorf("category = %s，期望 image", entry.Category)
	}
	if entry.IsTextual {
		t.Error("图片不应是文本类")
	}
	if entry.Charset != "" {
		t.Errorf("图片不应有字符集，实际 %q", entry.Charset)
	}
	if len(entry.Extensions) == 0 {
		t.Error("应给出常用扩展名")
	}
}

func TestMimeLookupQueryForms(t *testing.T) {
	t.Parallel()

	// `png`、`.png`、`report.png` 与 `dir/report.PNG` 应当等价。
	for _, input := range []string{"png", ".png", "report.png", "dir/report.PNG"} {
		out := mimeLookup(t, &MimeLookupRequest{Query: input})
		if out.Query != "png" {
			t.Errorf("%s → query = %s，期望 png", input, out.Query)
		}
		if out.Total == 0 || out.Entries[0].MimeType != "image/png" {
			t.Errorf("%s → 首条应为 image/png，实际 total=%d", input, out.Total)
		}
	}
}

func TestMimeLookupByType(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{Query: "application/json"})
	if out.Total != 1 {
		t.Fatalf("total = %d，期望 1", out.Total)
	}
	entry := out.Entries[0]
	if entry.MimeType != "application/json" {
		t.Errorf("mimeType = %s，期望 application/json", entry.MimeType)
	}
	if !entry.IsTextual || entry.Charset != "utf-8" {
		t.Errorf("JSON 应是文本类且字符集为 utf-8，实际 textual=%v charset=%q", entry.IsTextual, entry.Charset)
	}
}

func TestMimeLookupStripsParameters(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{Query: "text/plain; charset=utf-8"})
	if out.Query != "text/plain" {
		t.Errorf("query = %s，期望 text/plain", out.Query)
	}
	if out.Total != 1 || out.Entries[0].MimeType != "text/plain" {
		t.Fatalf("带参数的 text/plain 应唯一命中，实际 total=%d", out.Total)
	}
}

func TestMimeLookupAlias(t *testing.T) {
	t.Parallel()

	// `image/x-png` 是历史别名，应指向规范类型 image/png。
	out := mimeLookup(t, &MimeLookupRequest{Query: "image/x-png"})
	if out.Total != 1 || out.Entries[0].MimeType != "image/png" {
		t.Fatalf("别名应归一到 image/png，实际 total=%d", out.Total)
	}
}

func TestMimeLookupKeyword(t *testing.T) {
	t.Parallel()

	// 关键词命中用途说明，允许多条。
	out := mimeLookup(t, &MimeLookupRequest{Query: "图像"})
	if out.Total == 0 {
		t.Fatal("关键词「图像」应命中若干条目")
	}
	for _, entry := range out.Entries {
		if entry.MimeType == "" {
			t.Error("命中条目缺少媒体类型")
		}
	}
}

func TestMimeLookupNotMatched(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{Query: "zzzz"})
	if out.Total != 0 {
		t.Errorf("未知查询词不应命中，实际 %d 条", out.Total)
	}
	if out.Entries == nil {
		t.Error("空结果也应返回空数组而不是 null")
	}
}

func TestMimeLookupListAll(t *testing.T) {
	t.Parallel()

	loadMimeDataset()

	out := mimeLookup(t, &MimeLookupRequest{})
	if out.Total != len(mimeEntries) {
		t.Errorf("空请求应返回全量 %d 条，实际 %d", len(mimeEntries), out.Total)
	}
	if out.Query != "" || out.Category != "" {
		t.Errorf("全量列表的 query/category 应为空串，实际 %q / %q", out.Query, out.Category)
	}
}

func TestMimeLookupCategoryFilter(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{Category: "image"})
	if out.Total == 0 {
		t.Fatal("image 类别不应为空")
	}
	for _, entry := range out.Entries {
		if entry.Category != "image" {
			t.Errorf("%s 的类别是 %s，期望 image", entry.MimeType, entry.Category)
		}
	}
}

func TestMimeLookupValidate(t *testing.T) {
	t.Parallel()

	if err := (&MimeLookupRequest{}).Validate(); err != nil {
		t.Fatalf("空请求表示全量，不应报错：%v", err)
	}
	if err := (&MimeLookupRequest{Query: strings.Repeat("a", 129)}).Validate(); err == nil {
		t.Fatal("超长查询词应当报错")
	}
	if err := (&MimeLookupRequest{Category: "bogus"}).Validate(); err == nil {
		t.Fatal("非法顶层类型应当报错")
	}
	if err := (&MimeLookupRequest{Query: "png", Category: "image"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
}

func TestMimeLookupInvalidCategory(t *testing.T) {
	t.Parallel()

	_, err := MimeLookup(context.Background(), &MimeLookupRequest{Category: "bogus"})
	assertCode(t, err, toolerr.CodeInvalidField)
}

func TestMimeDatasetIntegrity(t *testing.T) {
	t.Parallel()

	loadMimeDataset()

	if len(mimeEntries) == 0 {
		t.Fatal("MIME 数据集为空")
	}
	// 每个规范类型必须能按类型查到，扩展名索引不能指向空条目。
	for i := range mimeEntries {
		entry := &mimeEntries[i]
		if entry.MimeType == "" {
			t.Error("存在空 mimeType 的条目")
		}
		if entry.Category == "" {
			t.Errorf("%s 缺少顶层类型", entry.MimeType)
		}
		if entry.Description == "" {
			t.Errorf("%s 缺少用途说明", entry.MimeType)
		}
		found, ok := mimeByType[entry.MimeType]
		if !ok || found != entry {
			t.Errorf("%s 的类型索引不一致", entry.MimeType)
		}
	}
	for ext, entry := range mimeByExtension {
		if ext == "" || entry == nil {
			t.Errorf("扩展名索引存在空项：%q", ext)
		}
	}
}
