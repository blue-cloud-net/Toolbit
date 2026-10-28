// Package toolset 汇总全部工具实现。
//
// 本包是**装配层**，不含任何工具逻辑：
// 它只把各分类包的 `Definitions()` 拼成一张表，
// 供进程入口与 API 测试共用，避免两处各维护一份清单。
//
// 依赖方向：本包 → tools/<分类> → tools。
// 因此**不能**放在 `tools` 包内（会与子包形成循环依赖）。
package toolset

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/crypto"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/data"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/documents"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/encoding"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/formatting"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/system"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/text"
)

// Definitions 返回全部已实现工具的定义。
//
// 顺序稳定：先按分类，再按分类内 `Definitions()` 的登记顺序。
// 该顺序**不**影响分发（分发按 slug 查表），但便于日志与测试比对。
func Definitions() []tools.Definition {
	defs := make([]tools.Definition, 0, 64)
	defs = append(defs, encoding.Definitions()...)
	defs = append(defs, formatting.Definitions()...)
	defs = append(defs, crypto.Definitions()...)
	defs = append(defs, text.Definitions()...)
	defs = append(defs, system.Definitions()...)
	defs = append(defs, data.Definitions()...)
	defs = append(defs, documents.Definitions()...)
	return defs
}
