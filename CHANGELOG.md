# Changelog

本文件记录 Blue Cloud Toolbit 的所有显著变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

尚无变更。

## [0.2.1] - 2026-10-28

前端骨架：让 v0.2.0 的后端链路在浏览器里跑起来。**不新增工具**。

### Added

- 前端工程脚手架（Vite 6 + Vue 3.5 + TypeScript 5.7 + Pinia + vue-router + vue-i18n + UnoCSS），Element Plus **按组件路径**引入
- 设计令牌与布局：`tokens.css` / `base.css` / `element-overrides.css`，`header` / `nav` / `main` / `footer` 地标齐备，「跳到主内容」链接可用
- API 客户端 `frontend/src/api/client.ts`：统一拆信封、`AbortController` 取消、超时归一化（`TIMEOUT`）、非 JSON 响应归为 `INVALID_RESPONSE`
- 目录 store 与发现页：首页 70 个入口 + 10 个分类、中英文名搜索、分类页、404
- 工具页骨架 `ToolPageLayout` / `ToolPanel` / `RunBar` / `ResultPanel` / `ErrorAlert` / `CopyButton` / `EncodingSelect`，四态齐备，`aria-live` 播报，`Ctrl` / `⌘ + Enter` 运行
- `base64` 与 `number-base` 两个工具页，含示例数据与「填入示例 → 运行 → 复制」全链路
- 双语词条：`zh-CN` 与 `en-US` 全覆盖（含 81 个错误码与警告码）
- 侧栏导航：`全部工具` / `最近使用` / `我的收藏` 三个特殊快捷入口，分类可折叠并展开工具列表
- 侧栏可折叠成图标轨道（仅图标 + 分类首两字），状态持久化；分类全部展开后菜单超出视口时**侧栏内部滚动且隐藏滚动条**
- 顶栏：主题切换改为图标按钮；语言与「跟随系统」收进设置菜单（Esc 关闭、点击外部关闭、关闭后焦点回到触发器）
- 工具页标题行新增收藏星标；新增 `/recent` 与 `/favorites` 两个页面
- 布局：侧栏为整列直通到底部（贴住视口底、带右边框），页脚只占内容列；页脚文案居中，含数据说明、版权、Apache-2.0 许可证、版本号（链接到发布记录）与 GitHub 仓库链接
- 站点常量集中在 `frontend/src/config/site.ts`（仓库地址、发布地址、版本号、版权年份），组件内不再写死
- 主题（`system` / `light` / `dark`）与语言切换，持久化并同步 `<html lang>`
- 体积预算脚本 `scripts/check-budget.mjs`：初始 JS < 250 KB（gzip）、单工具 chunk < 40 KB（gzip）
- `Makefile` 新增 `lint-web` / `test-web` / `build-web` / `check-budget`，并纳入 `verify`

### 依赖

- adds: `vue` 3.5.44
- adds: `vue-router` 4.6.4
- adds: `pinia` 3.0.4
- adds: `vue-i18n` 11.4.16
- adds: `element-plus` 2.14.7
- adds: `vite` 6.4.1
- adds: `vitest` 3.2.7
- adds: `unocss` 66.12.0

## [0.2.0] - 2026-10-28

后端骨架与基础设施。**不含任何前端产物**，前端骨架见 `v0.2.1`。

### Added

- 仓库拆分为两个独立构建单元：`backend/`（独立 Go 模块 `github.com/blue-cloud-net/Toolbit/backend`）与 `frontend/`（pnpm 项目）
- 工具目录内嵌：`backend/catalog/embed.go` 以 `//go:embed` 携带 `tools.yaml`，`internal/catalog` 负责解析与六类一致性校验
- 错误框架 `internal/toolerr`：67 个错误码 + 14 个警告码的注册表、统一 `Error` 类型与 `toolerr.From` 归一化
- 配置 `internal/config`：16 个 `TOOLBIT_*` 环境变量，非法值 fail-fast 且错误信息不含变量值
- 响应信封与中间件 `internal/httpx`：统一信封、requestId 透传与校验、安全响应头、访问日志、Body 限制、跨站校验、panic 恢复
- HTTP 分发器 `internal/api`：按 10 步顺序校验的统一端点、目录端点（`ETag` 协商）、限流（按 `rateClass` 五档分桶）、运维端点
- 统计落库 `internal/stats`：版本化迁移、异步批量写入、队列满丢弃计数、每日保留期清理
- 静态资源服务 `internal/webui`：缓存头、SPA 回退、路径穿越拦截、`/api/*` JSON 404
- 工具实现：`base64` 与 `number-base` 两个编码类工具，含契约测试数据
- 进程入口 `backend/cmd/toolbit/main.go`：配置装配、`SIGTERM` 优雅退出
- 工程门禁：`Makefile`（`make verify`）、`.golangci.yml`、`.editorconfig`

### Changed

- 重构路线图：[docs/roadmap.md](docs/roadmap.md) 拆分为 19 节，版本序列新增 `v0.2.1`（前端骨架）与 `v0.9.0`（统计体系），原「质量与硬化」顺延为 `v0.10.0+`
- `v0.2.0` 收敛为纯后端骨架（17 个 Issue 级任务），前端工作整体移入 `v0.2.1`
- 新增 `v0.9.0` 统计体系章节（9 个 Issue 级任务），涵盖访客标识、日聚合、公开统计端点与统计页

### 破坏性变更

- 目录布局：后端代码由仓库根移入 `backend/`，`catalog/tools.yaml` 移入 `backend/catalog/`，前端目录由 `web/` 改为 `frontend/`
- Go module 路径改为 `github.com/blue-cloud-net/Toolbit/backend`
- 68 个尚未实现的工具返回 `501 NOT_IMPLEMENTED`（catalog 仍返回全部 70 条）

### 依赖

- adds: `github.com/gin-gonic/gin` v1.12.0
- adds: `gopkg.in/yaml.v3` v3.0.1
- adds: `modernc.org/sqlite` v1.39.0
- adds: `golang.org/x/time` v0.12.0

## [0.1.0] - 2026-10-27

首个开发版本，仅包含工程地基与文档体系，工具尚未上线。

### Added

- 仓库骨架：`backend/cmd/toolbit`、`backend/internal/{api,catalog,codecsvc,cryptosvc,httpx,netguard,stats,toolerr,tools,webui}`、`frontend/`、`deploy/`、`catalog/`
- 文档体系：[README](README.md)、[AGENTS](AGENTS.md)、[架构总览](docs/architecture.md)、[开发规范](docs/standards/00-intro.md)、[API 契约](docs/api/README.md)、[工具规格](docs/tools.md)、[路线图](docs/roadmap.md)
- 工具目录 `backend/catalog/tools.yaml`：登记 70 个工具的分类、传输方式、限流档位与限额
- CI 工作流骨架：`ci`（lint / test）、`builder`（构建 Tongsuo 基础镜像）、`docker`（多架构镜像推 GHCR）
- 文档校验：Markdown lint、相对链接检查、catalog / 错误码 / 环境变量覆盖率检查

[未发布]: https://github.com/blue-cloud-net/Toolbit/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/blue-cloud-net/Toolbit/releases/tag/v0.1.0
