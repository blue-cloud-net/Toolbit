package network

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// httpStatus 执行一次状态码查询并取出结果。
func httpStatus(t *testing.T, req *HttpStatusRequest) HttpStatusResult {
	t.Helper()
	result, err := HttpStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	out, ok := result.Data.(HttpStatusResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

func TestHttpStatusByCode(t *testing.T) {
	t.Parallel()

	out := httpStatus(t, &HttpStatusRequest{Code: 404})
	if !out.Matched {
		t.Fatal("404 应当命中")
	}
	if out.Kind != StatusQueryCode {
		t.Errorf("kind = %s，期望 %s", out.Kind, StatusQueryCode)
	}
	if out.Code != 404 {
		t.Errorf("code = %d，期望 404", out.Code)
	}
	if out.Reason != "Not Found" {
		t.Errorf("reason = %s，期望 Not Found", out.Reason)
	}
	if out.Category != StatusCategoryClientError {
		t.Errorf("category = %s，期望 %s", out.Category, StatusCategoryClientError)
	}
	if out.Description == "" {
		t.Error("应给出中文说明")
	}
	if out.Total != 1 {
		t.Errorf("total = %d，期望 1", out.Total)
	}
}

func TestHttpStatusCategories(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code int
		want string
	}{
		{100, StatusCategoryInformational},
		{204, StatusCategorySuccess},
		{301, StatusCategoryRedirect},
		{404, StatusCategoryClientError},
		{503, StatusCategoryServerError},
	}
	for _, c := range cases {
		out := httpStatus(t, &HttpStatusRequest{Code: c.code})
		if out.Category != c.want {
			t.Errorf("%d 的类别 = %s，期望 %s", c.code, out.Category, c.want)
		}
		if !out.Matched {
			t.Errorf("%d 应当命中", c.code)
		}
	}
}

func TestHttpStatusUnknownCodeKeepsCategory(t *testing.T) {
	t.Parallel()

	// 299 落在 2xx 区间但不在表内：仍应给出类别，只是不命中。
	out := httpStatus(t, &HttpStatusRequest{Code: 299})
	if out.Matched {
		t.Error("299 不应命中")
	}
	if out.Code != 299 {
		t.Errorf("code = %d，期望 299", out.Code)
	}
	if out.Category != StatusCategorySuccess {
		t.Errorf("category = %s，期望 %s", out.Category, StatusCategorySuccess)
	}
	if out.Reason != "" {
		t.Errorf("未命中时 reason 应为空，实际 %q", out.Reason)
	}
}

func TestHttpStatusNumericKeyword(t *testing.T) {
	t.Parallel()

	// 纯数字的关键字按状态码处理。
	out := httpStatus(t, &HttpStatusRequest{Query: "500"})
	if !out.Matched {
		t.Fatal("500 应当命中")
	}
	if out.Kind != StatusQueryCode {
		t.Errorf("kind = %s，期望 %s", out.Kind, StatusQueryCode)
	}
	if out.Reason != "Internal Server Error" {
		t.Errorf("reason = %s", out.Reason)
	}
}

func TestHttpStatusKeyword(t *testing.T) {
	t.Parallel()

	out := httpStatus(t, &HttpStatusRequest{Query: "redirect"})
	if !out.Matched {
		t.Fatal("redirect 应当命中")
	}
	if out.Kind != StatusQueryKeyword {
		t.Errorf("kind = %s，期望 %s", out.Kind, StatusQueryKeyword)
	}
	if out.Total < 2 {
		t.Errorf("total = %d，期望至少 2 条重定向", out.Total)
	}
	// 原因短语里带 Redirect 的最早一条是 307 Temporary Redirect。
	if out.Code != 307 {
		t.Errorf("code = %d，期望首条为 307", out.Code)
	}
}

func TestHttpStatusChineseKeyword(t *testing.T) {
	t.Parallel()

	out := httpStatus(t, &HttpStatusRequest{Query: "重定向"})
	if out.Total == 0 {
		t.Fatal("中文关键字应当命中")
	}
	for _, entry := range out.Matches {
		if entry.Category != StatusCategoryRedirect {
			t.Errorf("命中的 %d 类别 = %s，期望 %s", entry.Code, entry.Category, StatusCategoryRedirect)
		}
	}
}

func TestHttpStatusKeywordNotMatched(t *testing.T) {
	t.Parallel()

	out := httpStatus(t, &HttpStatusRequest{Query: "不存在的关键字"})
	if out.Matched {
		t.Error("不应命中")
	}
	if out.Total != 0 {
		t.Errorf("total = %d，期望 0", out.Total)
	}
	if out.Matches == nil {
		t.Error("未命中时也应返回空数组而不是 null")
	}
}

func TestHttpStatusValidate(t *testing.T) {
	t.Parallel()

	if err := (&HttpStatusRequest{}).Validate(); err == nil {
		t.Fatal("空请求应当报缺少字段")
	}
	if err := (&HttpStatusRequest{Code: 404, Query: "404"}).Validate(); err == nil {
		t.Fatal("两个字段同时给出应当报错")
	}
	if err := (&HttpStatusRequest{Code: 99}).Validate(); err == nil {
		t.Fatal("越界的状态码应当报错")
	}
	if err := (&HttpStatusRequest{Code: 600}).Validate(); err == nil {
		t.Fatal("越界的状态码应当报错")
	}
	if err := (&HttpStatusRequest{Query: "999"}).Validate(); err == nil {
		t.Fatal("越界的数字关键字应当报错")
	}
	if err := (&HttpStatusRequest{Code: 404}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
	if err := (&HttpStatusRequest{Query: "not found"}).Validate(); err != nil {
		t.Fatalf("合法请求不应报错：%v", err)
	}
}

func TestHttpStatusInvalidCode(t *testing.T) {
	t.Parallel()

	_, err := HttpStatus(context.Background(), &HttpStatusRequest{Code: 999})
	assertCode(t, err, toolerr.CodeInvalidField)
}

func TestStatusDatasetIntegrity(t *testing.T) {
	t.Parallel()

	loadStatusDataset()

	if len(statusEntries) == 0 {
		t.Fatal("状态码数据集为空")
	}
	seen := map[int]bool{}
	for _, entry := range statusEntries {
		if entry.Code < 100 || entry.Code > 599 {
			t.Errorf("状态码越界：%d", entry.Code)
		}
		if seen[entry.Code] {
			t.Errorf("状态码重复：%d", entry.Code)
		}
		seen[entry.Code] = true
		if entry.Reason == "" {
			t.Errorf("%d 缺少原因短语", entry.Code)
		}
		if entry.Description == "" {
			t.Errorf("%d 缺少中文说明", entry.Code)
		}
		if entry.Category == "" {
			t.Errorf("%d 缺少类别", entry.Code)
		}
		if entry.Category != statusCategory(entry.Code) {
			t.Errorf("%d 的类别与状态Category函数不一致", entry.Code)
		}
	}
}
