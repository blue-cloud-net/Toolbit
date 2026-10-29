package toolset

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// TestDefinitionsNoDuplicate 断言全部工具定义的 slug 唯一。
//
// slug 是 API 契约的一部分：重复定义会让分发器的查表结果
// 取决于注册顺序，属于**静默**的缺陷，必须在编译期就被拦住。
func TestDefinitionsNoDuplicate(t *testing.T) {
	seen := make(map[string]bool)
	for _, d := range Definitions() {
		if d.Slug == "" {
			t.Fatal("存在空 slug 的工具定义")
		}
		if seen[d.Slug] {
			t.Errorf("slug 重复定义：%s", d.Slug)
		}
		seen[d.Slug] = true
	}
}

// TestDefinitionsInCatalog 断言全部实现都在 catalog 中登记。
//
// 反向检查（catalog 有、实现无）是正常的——未实现的工具返回 501；
// 这里只保证**实现**不会脱离目录。
func TestDefinitionsInCatalog(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("加载目录失败：%v", err)
	}

	registry := tools.NewRegistry(Definitions())
	_, orphan := tools.CrossCheck(cat, registry)
	if len(orphan) > 0 {
		t.Errorf("存在未登记的工具实现：%v", orphan)
	}
}
