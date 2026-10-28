package tools

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
)

// stubRequest 是测试用的最小请求类型。
type stubRequest struct {
	Value string `json:"value"`
}

// Validate 实现 Request 接口。
func (r stubRequest) Validate() error { return nil }

// stubDefinition 构造一个测试用实现。
func stubDefinition(slug string) Definition {
	return Definition{
		Slug:       slug,
		NewRequest: func() Request { return &stubRequest{} },
		Handle: func(_ context.Context, _ Request) (Result, error) {
			return Result{Data: map[string]string{"ok": "1"}}, nil
		},
	}
}

func TestRegistryLookupIsCaseSensitive(t *testing.T) {
	r := NewRegistry([]Definition{stubDefinition("base64")})
	if _, ok := r.Lookup("base64"); !ok {
		t.Error("应能按 slug 找到 base64")
	}
	if _, ok := r.Lookup("Base64"); ok {
		t.Error("slug 查询必须大小写敏感")
	}
}

func TestRegistrySlugsSorted(t *testing.T) {
	r := NewRegistry([]Definition{
		stubDefinition("zeta"),
		stubDefinition("alpha"),
		stubDefinition("mid"),
	})
	got := r.Slugs()
	want := []string{"alpha", "mid", "zeta"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Slugs() = %v，期望 %v", got, want)
		}
	}
}

func TestCrossCheckReportsMissingAndOrphan(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}

	// 只实现 base64：其余 67 个已启用工具都算缺失。
	r := NewRegistry([]Definition{stubDefinition("base64")})
	missing, orphan := CrossCheck(cat, r)
	if len(missing) == 0 {
		t.Error("应当报出未实现的工具")
	}
	if len(orphan) != 0 {
		t.Errorf("不应有孤儿实现，实际 %v", orphan)
	}
	for _, slug := range missing {
		if slug == "base64" {
			t.Error("已实现的工具不应出现在缺失列表中")
		}
	}

	// 注册一个 catalog 中不存在的 slug：应当报为孤儿。
	r2 := NewRegistry([]Definition{stubDefinition("base64"), stubDefinition("not-in-catalog")})
	_, orphan2 := CrossCheck(cat, r2)
	if len(orphan2) != 1 || orphan2[0] != "not-in-catalog" {
		t.Errorf("孤儿实现 = %v，期望 [not-in-catalog]", orphan2)
	}
}

func TestCrossCheckIgnoresDisabledTools(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}

	// 构造一个只含已禁用工具的目录：不应报缺失。
	disabled := *cat
	disabled.Tools = []catalog.Tool{{
		Slug:      "off-tool",
		Category:  "encoding",
		Enabled:   false,
		Transport: catalog.TransportJSON,
		RateClass: catalog.RateLight,
	}}

	r := NewRegistry(nil)
	missing, _ := CrossCheck(&disabled, r)
	if len(missing) != 0 {
		t.Errorf("已禁用的工具不应算缺失，实际 %v", missing)
	}
}

func TestReportMentionsCounts(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}
	r := NewRegistry([]Definition{stubDefinition("base64")})
	report := Report(cat, r)
	if report == "" {
		t.Error("报告不得为空")
	}
}
