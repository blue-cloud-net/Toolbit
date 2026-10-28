# 文档规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

---

## 1. 文档分类

| 分类 | 目录 | 受众 | 变更频率 |
| --- | --- | --- | --- |
| 规范 | `docs/standards/` | 开发者 | 中 |
| 架构 | `docs/architecture.md` | 开发者 | 低 |
| 接口 | `docs/api/` | 前后端开发者 | 中 |
| 产品 | `docs/tools.md` | 用户 | 中 |
| 运维 | `docs/deploy.md` | 运维 | 低 |
| 路线 | `docs/roadmap.md` | 所有人 | 低 |
| 政策 | `docs/privacy.md` | 所有人 | 低 |
| 索引 | `docs/README.md` | 所有人 | 低 |

---

## 2. 唯一事实来源（SSOT）

| 主题 | 唯一来源 | 其他文档的角色 |
| --- | --- | --- |
| 工具清单与元数据 | `backend/catalog/tools.yaml` | 文档**引用**，不重复定义 |
| 错误码 | `backend/internal/toolerr/registry.go` | `docs/api/errors.md` 为镜像 |
| 内部错误码 | `backend/internal/toolerr/registry.go` | `docs/api/errors.md` 为镜像 |
| 配置项 | `backend/internal/config` | `docs/deploy.md` 为镜像 |
| 国际化词条 | `frontend/src/locales/` | `docs/i18n.md` 描述规则 |
| 架构决策 | `docs/architecture.md` D1–D10 | 其他文档引用编号 |
| 隐私口径 | `docs/privacy.md` | README 引用 |

**必须**：一份信息只在一处**定义**，其他处**链接**。

**禁止**在两处维护同一份表格而不同步。

---

## 3. 文件命名与路径

- **必须**使用 kebab-case：`api-contract.md`、`sm4-file-crypt.md`；
- 目录**必须**与内容类型对应；
- **禁止**使用中文文件名；
- **禁止**在文件名中使用日期（变更历史在 git 里）。

---

## 4. 文档结构

**必须**以如下开头：

```markdown
# 标题

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27
```

| 状态 | 含义 |
| --- | --- |
| `Draft` | 草稿，可能变动 |
| `Stable` | 稳定，可依赖 |
| `Deprecated` | 已废弃，保留历史 |

**必须**：

- 一级标题**唯一**且是文档名；
- 章节从 `##` 开始，**不**跳级；
- 章节**必须**编号（`## 1. 目标`），便于引用；
- 表格**必须**有表头分隔行；
- 代码块**必须**标注语言；
- 链接**必须**相对路径（**禁止**裸 URL 作为引用形式，可用裸 URL 作为参考链接定义）。

---

## 5. 规范用语

遵循 RFC 2119：

| 用语 | 含义 |
| --- | --- |
| **必须** / 禁止 | 绝对要求 |
| **应当** / 不应当 | 推荐，偏离需说明理由 |
| **可以** | 可选 |

**必须**在规范类文档中使用上述中文术语，不用「要」「应该」「尽量」。

---

## 6. 写给谁

| 文档 | 主要读者 | 写法 |
| --- | --- | --- |
| `docs/standards/*` | 开发者 | 命令式、明确、可判定 |
| `docs/api/*` | 前后端开发者 | 精确到字段与类型 |
| `docs/tools.md` | 最终用户 | 讲用途与注意事项，不讲实现 |
| `docs/deploy.md` | 运维 | 命令可复制粘贴 |
| `docs/privacy.md` | 所有人 | 清楚、无歧义 |

**禁止**在面向用户的文档中暴露内部实现细节（包名、函数名）。

**禁止**在面向开发者的文档中省略必要的错误码与边界。

---

## 7. 双语策略

- **文档正文一律中文**；
- 标识符、代码、命令、协议字段用英文；
- README **必须**附英文摘要（`## English Summary`）；
- 面向用户的界面文案由 i18n 提供，**不在**文档里重复维护全部词条。

**禁止**整篇英文文档。

---

## 8. 图表

**必须**使用 Mermaid：

| 场景 | 类型 |
| --- | --- |
| 系统结构 | `flowchart` |
| 请求流程 | `sequenceDiagram` |
| 数据模型 | `erDiagram` |
| 状态机 | `stateDiagram-v2` |

**必须**

- Mermaid 节点 ID 用英文，中文只出现在标签里；
- **禁止**在 Mermaid 中使用 HTML 标签（`<br/>` 除外）；
- 复杂图**应当**拆成多张，而不是画成一坨。

---

## 9. 代码示例

**必须**

- 标注语言；
- 示例**必须**可运行或明确标注为「示意」；
- 展示**推荐**用法，**禁止**展示反例作为正面示例；
- 涉及密钥的示例**必须**使用明显的占位值。

---

## 10. 表格

**必须**

- 表头**必须**有语义；
- 列**应当**少于 6 个；
- 单元格内容**应当**简短（一句话以内）；
- 长内容**应当**拆到正文或列表。

---

## 11. 链接

**必须**

- 相对路径链接到仓库内文件；
- 链接文字**必须**描述目标（**禁止**「点击这里」「更多」）；
- 链接到文件**应当**带章节锚点。

**禁止**链接到已删除或计划中的文件。

---

## 12. 校验

CI **必须**执行：

| 检查 | 工具 | 失败条件 |
| --- | --- | --- |
| Markdown 风格 | `markdownlint-cli2` | 任何 lint 错误 |
| 内部链接可达 | `lychee` | 404 / 断链 |
| 锚点有效 | `lychee` | 锚点不存在 |
| catalog 覆盖 | `scripts/check-docs.mjs` | 工具缺少文档 |
| 错误码覆盖 | `scripts/check-docs.mjs` | 错误码缺文档 |
| 配置项覆盖 | `scripts/check-docs.mjs` | 变量缺文档 |
| 禁词 | `scripts/check-docs.mjs` | 出现禁词 |

`scripts/check-docs.mjs` **必须**：

- 以 `backend/catalog/tools.yaml` 为准，校验 `docs/api/tools/*.md` 覆盖全部 70 个工具；
- 以 `backend/internal/toolerr/registry.go` 为准，校验 `docs/api/errors.md` 覆盖全部错误码；
- 以 `backend/internal/config` 为准，校验 `docs/deploy.md` 覆盖全部环境变量；
- 输出**明确的差异清单**（缺什么），**禁止**只输出「失败」。

---

## 13. 模板

**必须**使用 `docs/_templates/` 下的模板：

| 模板 | 用途 |
| --- | --- |
| `tool-spec.md` | 单工具产品规格 |
| `api-category.md` | 单分类 API |
| `standard.md` | 规范文档 |

---

## 14. 变更同步矩阵

| 变更类型 | **必须**同步更新的文档 |
| --- | --- |
| 新增工具 | `backend/catalog/tools.yaml`、`docs/api/tools/<分类>.md`、`docs/tools.md`、i18n 词条 |
| 新增错误码 | `backend/internal/toolerr/registry.go`、`docs/api/errors.md`、i18n 词条 |
| 新增配置项 | `backend/internal/config`、`docs/deploy.md` |
| 架构决策变化 | `docs/architecture.md`、受影响的 `docs/standards/*` |
| 破坏性 API 变化 | `docs/api/*`、`CHANGELOG.md` |
| 新增依赖 | `docs/architecture.md` 技术栈、`CHANGELOG.md` |

---

## 15. 禁止清单

- ❌ 一份信息在多处定义。
- ❌ 文档中包含密钥、口令、真实用户数据。
- ❌ 面向用户的文档暴露内部包名与函数名。
- ❌ 规范文档用「尽量」「建议」代替「必须/应当」。
- ❌ 提交断链文档。
- ❌ 文档中直接粘贴大段实现代码代替设计说明。
