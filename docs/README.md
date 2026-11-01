# 文档索引

> 状态：Draft ｜ 适用版本：v0.3.0 ｜ 最后更新：2026-11-01

本目录是 Toolbit 的完整文档集。

---

## 1. 按角色选择

| 你是 | 从这里开始 |
| --- | --- |
| **使用者** | [tools.md](tools.md) → [../README.md](../README.md) |
| **运维** | [deploy.md](deploy.md) → [api/admin.md](api/admin.md) |
| **前端开发者** | [standards/01-frontend.md](standards/01-frontend.md) → [api/README.md](api/README.md) → [i18n.md](i18n.md) |
| **后端开发者** | [standards/02-backend.md](standards/02-backend.md) → [architecture.md](architecture.md) → [api/errors.md](api/errors.md) |
| **密码学相关** | [standards/03-crypto.md](standards/03-crypto.md) → [api/tools/crypto.md](api/tools/crypto.md) |
| **贡献者** | [../AGENTS.md](../AGENTS.md) → [standards/10-git.md](standards/10-git.md) |

---

## 2. 全部文档

### 2.1 根目录

| 文档 | 说明 |
| --- | --- |
| [../README.md](../README.md) | 项目总览、快速开始、技术栈 |
| [../AGENTS.md](../AGENTS.md) | 架构不变量、禁止清单、新增工具流程 |
| [../CHANGELOG.md](../CHANGELOG.md) | 变更记录 |
| [../LICENSE](../LICENSE) | Apache-2.0 |

### 2.2 架构与产品

| 文档 | 说明 |
| --- | --- |
| [architecture.md](architecture.md) | 系统上下文、分层、关键决策 D1–D11 |
| [tools.md](tools.md) | 98 个工具的产品规格与使用说明 |
| [tool-candidates.md](tool-candidates.md) | 候选工具清单、边界与落地状态 |
| [roadmap.md](roadmap.md) | 阶段计划、v1.1 展望、风险对策 |
| [glossary.md](glossary.md) | 术语表 |

### 2.2.1 上游依赖

| 文档 | 说明 |
| --- | --- |
| [upstream/tongsuo-go-issues.md](upstream/tongsuo-go-issues.md) | `tongsuo-go` 的能力缺口登记与验收标准 |

### 2.3 规范（`standards/`）

| 文档 | 说明 |
| --- | --- |
| [00-intro.md](standards/00-intro.md) | 规范总则、用语与文档状态 |
| [01-frontend.md](standards/01-frontend.md) | 前端：技术栈、目录、组件、主题、i18n、性能 |
| [02-backend.md](standards/02-backend.md) | 后端：分层、handler 形态、解码、错误、日志、配置 |
| [03-crypto.md](standards/03-crypto.md) | 密码学：句柄生命周期、算法默认值、PKI、限额、TBFC |
| [04-errors.md](standards/04-errors.md) | 错误：两层结构、注册表、params、映射、校验顺序 |
| [05-testing.md](standards/05-testing.md) | 测试：分层、契约测试、模糊测试、覆盖率目标 |
| [06-security.md](standards/06-security.md) | 安全：威胁模型、SSRF、限流、密钥处理、容器安全 |
| [07-observability.md](standards/07-observability.md) | 可观测性：requestId、日志、指标、健康检查、统计 |
| [08-deploy.md](standards/08-deploy.md) | 部署：镜像、配置、优雅退出、Compose、CI/CD、升级 |
| [09-docs.md](standards/09-docs.md) | 文档：SSOT、写法、双语、图表、校验、同步矩阵 |
| [10-git.md](standards/10-git.md) | Git：分支、提交信息、粒度、PR、自检 |

### 2.4 API（`api/`）

| 文档 | 说明 |
| --- | --- |
| [api/README.md](api/README.md) | 总览：端点、信封、通用类型、限流、WebSocket、示例 |
| [api/errors.md](api/errors.md) | 错误码与警告全集 |
| [api/admin.md](api/admin.md) | 健康、元信息、统计、探针、内部端点 |
| [api/tools/crypto.md](api/tools/crypto.md) | 加密（8） |
| [api/tools/keys.md](api/tools/keys.md) | 密钥（6） |
| [api/tools/certificates.md](api/tools/certificates.md) | 证书（23） |
| [api/tools/encoding.md](api/tools/encoding.md) | 编码（6） |
| [api/tools/formatting.md](api/tools/formatting.md) | 格式化（8） |
| [api/tools/text.md](api/tools/text.md) | 文本（2） |
| [api/tools/data.md](api/tools/data.md) | 数据（6） |
| [api/tools/system.md](api/tools/system.md) | 系统（3） |
| [api/tools/network.md](api/tools/network.md) | 网络（5） |
| [api/tools/documents.md](api/tools/documents.md) | 文档（3） |

### 2.5 政策与运维

| 文档 | 说明 |
| --- | --- |
| [deploy.md](deploy.md) | 部署指南：Docker、源码构建、配置、反代、故障排查 |
| [privacy.md](privacy.md) | 隐私说明：收集什么、不收集什么、保留期 |
| [i18n.md](i18n.md) | 国际化：词条结构、命名空间、校验 |

---

## 3. 唯一事实来源

| 主题 | 位置 |
| --- | --- |
| 工具元数据 | [`../backend/catalog/tools.yaml`](../backend/catalog/tools.yaml) |
| 错误码与 HTTP 状态 | `backend/internal/toolerr/registry.go` |
| 配置项 | `backend/internal/config` |
| i18n 词条 | `frontend/src/locales/` |
| 架构决策 | [architecture.md](architecture.md) §关键决策 |

**其它文档只引用，不重复定义。**

---

## 4. 文档约定

- 中文为主，标识符与代码用英文；README 附英文摘要；
- 每篇文档第二行为状态行：`> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27`；
- 规范类文档使用「必须 / 应当 / 可以 / 禁止」；
- 章节编号从 `## 1.` 开始；
- 图表使用 Mermaid。

详见 [standards/09-docs.md](standards/09-docs.md)。

---

## 5. 文档校验

CI 会执行：

| 检查 | 工具 |
| --- | --- |
| Markdown 风格 | `markdownlint-cli2` |
| 链接与锚点可达 | `lychee` |
| catalog ↔ 文档覆盖 | `scripts/check-docs.mjs` |
| 错误码 ↔ 文档覆盖 | `scripts/check-docs.mjs` |
| 配置项 ↔ 文档覆盖 | `scripts/check-docs.mjs` |
| i18n 词条一致性 | `frontend/src/i18n/parity.test.ts` |
| 禁词 | `scripts/check-docs.mjs` |
