// Package catalog 把工具目录 YAML 内嵌进二进制。
//
// tools.yaml 是工具元数据的**唯一事实来源**（见 docs/architecture.md §6）。
// 它同时驱动后端行为、前端渲染与文档校验。
//
// 本包只负责「把 YAML 原文交出去」，**不**解析、**不**校验：
// 解析与校验由 internal/catalog 负责，避免两处逻辑漂移。
package catalog

import _ "embed"

// ToolsYAML 是 catalog/tools.yaml 的原始内容。
//
// //go:embed 不能引用上级目录，因此 tools.yaml 与本文件同目录。
//
//go:embed tools.yaml
var ToolsYAML []byte
