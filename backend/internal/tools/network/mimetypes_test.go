package network

import (
	"context"
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

	out := mimeLookup(t, &MimeLookupRequest{Extension: "png"})
	if !out.Matched {
		t.Fatal("png 应当命中")
	}
	if out.MimeType != "image/png" {
		t.Errorf("mimeType = %s，期望 image/png", out.MimeType)
	}
	if out.Category != "image" {
		t.Errorf("category = %s，期望 image", out.Category)
	}
	if out.IsTextual {
		t.Error("图片不应是文本类")
	}
	if out.Charset != "" {
		t.Errorf("图片不应有字符集，实际 %q", out.Charset)
	}
	if len(out.Extensions) == 0 {
		t.Error("应给出常用扩展名")
	}
}

func TestMimeLookupExtensionForms(t *testing.T) {
	t.Parallel()

	// `png`、`.png` 与 `report.png` 应当等价。
	for _, input := range []string{"png", ".png", "report.png", "dir/report.PNG"} {
		out := mimeLookup(t, &MimeLookupRequest{Extension: input})
		if out.MimeType != "image/png" {
			t.Errorf("%s → mimeType = %s，期望 image/png", input, out.MimeType)
		}
		if out.Query != "png" {
			t.Errorf("%s → query = %s，期望 png", input, out.Query)
		}
	}
}

func TestMimeLookupByType(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{MimeType: "application/json"})
	if !out.Matched {
		t.Fatal("application/json 应当命中")
	}
	if out.Kind != MimeQueryType {
		t.Errorf("kind = %s，期望 %s", out.Kind, MimeQueryType)
	}
	if !out.IsTextual || out.Charset != "utf-8" {
		t.Errorf("JSON 应是文本类且字符集为 utf-8，实际 textual=%v charset=%q", out.IsTextual, out.Charset)
	}
}

func TestMimeLookupStripsParameters(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{MimeType: "text/plain; charset=utf-8"})
	if !out.Matched {
		t.Fatal("带参数的 text/plain 应当命中")
	}
	if out.MimeType != "text/plain" {
		t.Errorf("mimeType = %s，期望 text/plain", out.MimeType)
	}
	if out.Query != "text/plain" {
		t.Errorf("query = %s，期望 text/plain", out.Query)
	}
}

func TestMimeLookupAlias(t *testing.T) {
	t.Parallel()

	// `image/x-png` 是历史别名，应指向规范类型 image/png。
	out := mimeLookup(t, &MimeLookupRequest{MimeType: "image/x-png"})
	if !out.Matched {
		t.Fatal("别名应当命中")
	}
	if out.MimeType != "image/png" {
		t.Errorf("mimeType = %s，期望 image/png", out.MimeType)
	}
}

func TestMimeLookupNotMatched(t *testing.T) {
	t.Parallel()

	out := mimeLookup(t, &MimeLookupRequest{Extension: "zzzz"})
	if out.Matched {
		t.Error("未知扩展名不应命中")
	}
	if out.MimeType != "" {
		t.Errorf("未命中时 mimeType 应为空，实际 %q", out.MimeType)
	}
	if out.Extensions == nil || out.Aliases == nil {
		t.Error("未命中时也应返回空数组而不是 null")
	}
}

func TestMimeLookupValidate(t *testing.T) {
	t.Parallel()

	if err := (&MimeLookupRequest{}).Validate(); err == nil {
		t.Fatal("空请求应当报缺少字段")
	}
	if err := (&MimeLookupRequest{Extension: "png", MimeType: "image/png"}).Validate(); err == nil {
		t.Fatal("两个字段同时给出应当报错")
	}
	if err := (&MimeLookupRequest{Extension: "a b"}).Validate(); err == nil {
		t.Fatal("非法扩展名应当报错")
	}
	if err := (&MimeLookupRequest{MimeType: "not-a-mime"}).Validate(); err == nil {
		t.Fatal("缺少斜杠的媒体类型应当报错")
	}
	if err := (&MimeLookupRequest{MimeType: "image/"}).Validate(); err == nil {
		t.Fatal("缺少子类型的媒体类型应当报错")
	}
	if err := (&MimeLookupRequest{Extension: "png"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
}

func TestMimeLookupBothFieldsRejected(t *testing.T) {
	t.Parallel()

	_, err := MimeLookup(context.Background(), &MimeLookupRequest{Extension: "png", MimeType: "image/png"})
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
