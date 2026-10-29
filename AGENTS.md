# AGENTS.md — 蓝云工具箱（Blue Cloud Toolbit）工作指南

> 状态：Stable ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

本文件面向**自动化 Agent 与人类贡献者**，是本仓库的工作契约。开始任何改动前先读本文件与 [docs/architecture.md](docs/architecture.md)。

---

## 1. 项目是什么

- 前后端一体的 IT 工具箱网站；v1 交付 **70 个工具**。
- 前端：Vue 3 + TypeScript + Vite + Element Plus（zh-CN / en-US 双语）。
- 后端：Go 1.25 + Gin；**所有工具在服务端计算**。
- 密码学：唯一底层实现是 [Tongsuo（铜锁）](https://www.tongsuo.net/)，经 `github.com/blue-cloud-net/tongsuo-go` 封装。
- 无账号、无用户数据；调用统计只保留匿名聚合。

---

## 2. 项目地图

仓库分为**两个独立构建单元**：`backend/`（Go 模块）与 `frontend/`（pnpm 项目），二者只通过 HTTP 契约耦合。

| 路径 | 职责 | 允许依赖 |
| --- | --- | --- |
| `backend/cmd/toolbit/` | 进程入口、装配、优雅退出 | 全部 `backend/internal/*` |
| `backend/internal/api/` | HTTP 分发器：查目录、协商 Content-Type、限流、超时、Body 限制、Origin 校验 | `httpx` `catalog` `tools` `toolerr` `stats` |
| `backend/internal/catalog/` | 工具目录：内嵌并解析 `backend/catalog/tools.yaml` | 标准库 + `yaml.v3` |
| `backend/internal/tools/` | 各领域工具实现（按领域分子包） | `cryptosvc` `pkisvc` `codecsvc` `toolerr` `netguard` |
| `backend/internal/toolset/` | 工具装配层：汇总各分类的 `Definitions()` | `tools` `tools/<分类>` |
| `backend/internal/cryptosvc/` | 密码学**原语**层；所有摘要 / MAC / 对称 / 非对称运算在此**转调 Tongsuo** | `tongsuo-go` |
| `backend/internal/pkisvc/` | 密钥与 **PKI** 层（密钥生成与格式、X.509、JWK、PKCS 容器） | `tongsuo-go` `cryptosvc` |
| `backend/internal/codecsvc/` | 编码归一化（utf8/hex/base64/base64url）与 PEM/DER 识别 | 标准库 |
| `backend/internal/toolerr/` | 错误码注册表（唯一来源） | 标准库 |
| `backend/internal/httpx/` | 响应信封、中间件、requestId | 标准库 |
| `backend/internal/netguard/` | 出网防护（解析 → 私网校验 → 钉 IP → 重定向与体积上限） | 标准库 |
| `backend/internal/stats/` | SQLite 匿名统计：迁移、异步写入、聚合、保留期 | `modernc.org/sqlite` |
| `backend/internal/webui/` | 静态资源与 SPA 回退 | 标准库 |
| `backend/internal/buildinfo/` | 版本信息（ldflags 注入） | 标准库 |
| `backend/internal/toolset/` | 工具装配层：汇总各分类的 `Definitions()` | `tools` `tools/*` |
| `backend/catalog/tools.yaml` | **工具元数据单一事实来源** | — |
| `backend/catalog/embed.go` | 把 `tools.yaml` 内嵌进二进制（`//go:embed`） | 标准库 |
| `frontend/` | Vue 前端 | — |
| `deployments/` | Dockerfile / docker-compose.yml / .env.example | — |
| `scripts/` | 文档与一致性校验脚本（`check-docs.mjs`） | 标准库 |
| `docs/` | 文档（`standards/` `api/` `_templates/`） | — |

**依赖方向是单向的**：`api → tools → {cryptosvc, pkisvc} → tongsuo-go`。任何反向依赖都是缺陷。

> 工具目录 YAML 位于 `backend/catalog/` 而非仓库根：`//go:embed` 不能引用上级目录，
> 单一事实来源必须与 `embed.go` 同目录。

---

## 3. 常用命令

```bash
# 后端（在 backend/ 下，是独立 Go 模块）
go build ./...
go test ./... -race
go vet ./...

# 前端（在 frontend/ 下）
pnpm install
pnpm lint
pnpm type-check
pnpm test
pnpm build

# 跨前后端（在仓库根）
make lint            # golangci-lint + markdownlint + 文档覆盖 + 前端 ESLint 与类型检查
make check-docs      # catalog / 错误码 / 环境变量覆盖率
make check-budget    # 前端体积预算（初始 JS 与单工具 chunk）
make verify          # 提交前全量校验：vet + test + lint + 前端单测、构建与体积预算

# Docker（部署；构建上下文是仓库根，编排在 deployments/）
make docker-up       # 构建镜像并启动（首次会编译 Tongsuo，约 5-10 分钟）
make docker-smoke    # 容器冒烟：healthz / readyz / 目录 / 首页
make docker-down     # 停止并删除容器（保留数据卷）
```

`make verify` 是提交前的**唯一闸口**（v0.10.0+ 之前没有 CI）。

---

## 4. 架构不变量（违反即回归）

1. **只有 `backend/internal/cryptosvc` 与 `backend/internal/pkisvc` 可以 `import tongsuo-go`。** 其他包一律通过这两个包的服务接口调用。
   - `cryptosvc` —— 密码学**原语**（摘要 / MAC / 对称加解密 / 非对称运算 / 签名）；
   - `pkisvc` —— 密钥与 **PKI 领域**（密钥生成与格式、X.509、JWK、PKCS 容器）。
2. **`backend/catalog/tools.yaml` 是工具元数据的唯一来源。** 分类、顺序、传输方式、限流档位、限额只在这里定义；代码与文档都不得另立一份。
3. **所有工具走后端 API。** 唯一例外是 `websocket-test`（浏览器原生 `WebSocket`），它**没有**后端端点。
4. **工具响应必须 `Cache-Control: no-store`。** 错误响应同样。
5. **不回显输入内容。** 错误 `message` 不得包含用户输入；可回显字段名与枚举值。
6. **不写敏感数据到日志。** 只记录 `slug`、状态码、耗时、字节数、错误码、`requestId`。
7. **原生句柄必须显式释放。** 禁止依赖 `finalizer`；`cryptosvc.Session` 负责 LIFO 释放。
   唯一例外是 `x509.CertificateRequest`（上游未提供 `Close()`），见 [docs/standards/03-crypto.md](docs/standards/03-crypto.md) §8.6。
8. **`priv.Public()` 不单独关闭**：它与私钥共享底层句柄。
9. **用户输入构造 DN 一律用 `x509.Name.TryAdd`**，禁止 `Add`（对非法字段会 panic）。
10. **国际化的 key 必须成对存在。** `zh-CN` 与 `en-US` 缺一不可，由 CI 校验。
11. **工具输入不写 `localStorage` / `sessionStorage` / IndexedDB。** 敏感输入离开页面即清空。
12. **不新增依赖而不说明理由。** 变更 `backend/go.mod` / `frontend/package.json` 必须在提交信息中写明 `adds: <pkg> <version>`。
13. **`backend/` 与 `frontend/` 互不依赖。** 后端不 import 前端产物，前端不 import 后端源码；二者只通过 HTTP 契约耦合。
14. **`backend/internal/` 只允许被 `backend/` 内的包导入。** 这是 Go 的可见性边界，不得跨越。
15. **密码学运算一律经 `cryptosvc` / `pkisvc` 转调 `tongsuo-go`。** Toolbit 内**不得**用 Go 标准库 `crypto/*`
    实现或拼装密码学运算（分组模式链接、填充、KDF、签名与密文格式转换）。标准库仅可用于
    **编码**（hex / base64 / PEM）与**结构解析**（不涉及密钥运算的只读查看）。
    库缺能力时**先向上游提 issue**（登记于 [docs/upstream/tongsuo-go-issues.md](docs/upstream/tongsuo-go-issues.md)），
    **禁止**自行实现兜底。
16. **两个密码学层的句柄类型互不通用。** `cryptosvc.PrivateKey` 与 `pkisvc.PrivateKey` 各自持有原生句柄，
    都必须显式释放；跨层不得直接传递句柄，只能传 PEM / DER 等**表示层**数据。
17. **工具实现必须登记在 `backend/internal/toolset`。** 进程入口与 API 测试共用该装配层，
    不允许各自维护一份清单（漏改会导致测试与实际装配不一致）。

---

## 5. 禁止清单

- ❌ 复制外部项目的代码、文档或文案；本仓库的工具清单、命名与描述必须自有。
- ❌ 在本仓库任何文件（README、docs、注释、提交信息、配置、测试）中提及任何**参考项目**的名称或路径。
- ❌ 在日志或响应中输出私钥、口令、请求体、证书正文。
- ❌ 用 Go 标准库 `crypto/*` 在 Toolbit 内实现或拼装密码学运算（编码与结构解析除外）。
- ❌ 为 `tongsuo-go` 的能力缺口自研兜底实现；缺口必须走上游 issue。
- ❌ 把 `x509.Name.Add` 用于用户可控输入。
- ❌ 跳过 `backend/internal/api` 分发器自行注册 `/api` 路由。
- ❌ 在前端做密码学计算（唯一例外：`websocket-test` 的协议握手）。
- ❌ 用 `git commit --no-verify` 绕过检查。
- ❌ 提交 `.env`、真实密钥、真实证书到仓库。
- ❌ 在镜像级 `ENV` 设置 `LD_LIBRARY_PATH` / `LD_PRELOAD` / `GOPATH` / `GOPROXY`（见部署规范）。

---

## 6. 变更 → 文档同步矩阵

| 变更内容 | 必须同步更新 |
| --- | --- |
| 新增 / 移除 / 改名工具 | `backend/catalog/tools.yaml`、[docs/tools.md](docs/tools.md)、对应 `docs/api/tools/*.md`、前端 `frontend/src/tools/index.ts` |
| 修改工具请求 / 响应字段 | 对应 `docs/api/tools/*.md`、前端 DTO 类型 |
| 新增 / 修改错误码 | `backend/internal/toolerr/registry.go`、[docs/api/errors.md](docs/api/errors.md)、前端 `errors.*` 文案 |
| 新增 / 修改环境变量 | `backend/internal/config`、[docs/standards/08-deploy.md](docs/standards/08-deploy.md)、[docs/deploy.md](docs/deploy.md) |
| 修改 Dockerfile / 编排 / 探针 | [docs/standards/08-deploy.md](docs/standards/08-deploy.md)、[docs/deploy.md](docs/deploy.md)、[docs/api/admin.md](docs/api/admin.md) §9 |
| 修改架构或关键决策 | [docs/architecture.md](docs/architecture.md) |
| 修改工具的输入输出或默认值 | [docs/tools.md](docs/tools.md)、前端参数区默认值 |
| 调整阶段范围或退出标准 | [docs/roadmap.md](docs/roadmap.md) |
| 调整规范条款 | 对应 `docs/standards/*.md` |

---

## 7. 提交前检查（按顺序）

1. `go vet ./... && go test ./... -race`
2. `cd web && pnpm lint && pnpm type-check && pnpm test && pnpm build`
3. `make lint`（含 markdownlint、链接检查、文档覆盖、禁词检查）
4. 自查：是否触碰了第 4 节的不变量、第 5 节的禁止项
5. 确认第 6 节的文档同步已完成

---

## 8. 新增一个工具的完整流程

1. **目录**：在 `backend/catalog/tools.yaml` 登记 slug、分类、顺序、`transport`、`rateClass`、`maxBodyBytes`、`timeoutMs`、`sensitive`、`requiresNetwork`、`autoRun`、`wide`、`i18nKey`。
2. **后端**：
   - 在对应领域包（如 `backend/internal/tools/crypto/`、`backend/internal/tools/keys/`）实现 handler；
   - 需要密码学**原语**时在 `backend/internal/cryptosvc` 增加服务方法；
   - 需要密钥 / **PKI**（生成与格式、X.509、JWK、PKCS 容器）时在 `backend/internal/pkisvc` 增加服务方法；
   - **不要**在 handler 里直接调 `tongsuo-go`；
   - 若两个服务层都缺少所需能力：先确认 `tongsuo-go` 是否已提供；仍未提供则在
     [docs/upstream/tongsuo-go-issues.md](docs/upstream/tongsuo-go-issues.md) 登记并向 upstream 提 issue，
     本版**跳过**该能力，**不得**自研兜底；
   - 定义请求 / 响应结构体并实现 `Validate()`，错误统一用 `toolerr`；
   - 在 `backend/internal/toolset` 登记 slug → handler（进程入口与 API 测试共用该装配层）。
3. **前端**：
   - 新建 `frontend/src/tools/<category>/<slug>/{index.ts,Tool.vue,example.ts}`；
   - 在 `frontend/src/tools/index.ts` 登记；
   - 复用共享组件（`EncodingSelect`、`KeyInput`、`CertTextInput`、`FileDropZone`、`CodeEditor`、`ResultPanel` 等），不要重复造。
4. **文案**：在 `frontend/src/locales/zh-CN/tools/<slug>.json` 与 `frontend/src/locales/en-US/tools/<slug>.json`
   补齐 `tools.<slug>.*`，字段：`name` `description` `category` `actions` `fields` `options`
   `results` `examples` `hints` `faq`（见 [docs/i18n.md](docs/i18n.md)）。
5. **示例**：`example.ts` 提供一组**测试用**示例数据（不得使用真实密钥）。
6. **测试**：
   - handler 正例 + 反例各一；
   - 若涉及密码学，补标准向量测试（见 [docs/standards/05-testing.md](docs/standards/05-testing.md)）；
   - 契约测试数据写入 `backend/testdata/contract/<slug>.json`。
7. **文档**：更新 [docs/tools.md](docs/tools.md) 与对应 `docs/api/tools/*.md`。
8. **校验**：`make check-docs` 通过。

---

## 9. 文档写作约定

- 中文正文；术语首次出现附英文，如 `密文顺序（C1C3C2）`。
- 每篇文档第二行为状态行：`> 状态：Draft|Stable ｜ 适用版本：vX.Y ｜ 最后更新：YYYY-MM-DD`。
- 字段名、文件名、错误码一律用反引号。
- 表格用于字段、错误码、环境变量、限额；Mermaid 用于架构图、时序图、ER 图。
- 相对链接；禁止绝对本地路径。
- 详见 [docs/standards/09-docs.md](docs/standards/09-docs.md)。

---

## 10. 相关文档

| 主题 | 文档 |
| --- | --- |
| 文档总索引 | [docs/README.md](docs/README.md) |
| 架构与决策 | [docs/architecture.md](docs/architecture.md) |
| 上游依赖缺口 | [docs/upstream/tongsuo-go-issues.md](docs/upstream/tongsuo-go-issues.md) |
| 前端规范 | [docs/standards/01-frontend.md](docs/standards/01-frontend.md) |
| 后端规范 | [docs/standards/02-backend.md](docs/standards/02-backend.md) |
| 密码学规范 | [docs/standards/03-crypto.md](docs/standards/03-crypto.md) |
| 错误码规范 | [docs/standards/04-errors.md](docs/standards/04-errors.md) |
| 测试规范 | [docs/standards/05-testing.md](docs/standards/05-testing.md) |
| 安全规范 | [docs/standards/06-security.md](docs/standards/06-security.md) |
| 可观测性规范 | [docs/standards/07-observability.md](docs/standards/07-observability.md) |
| 部署规范 | [docs/standards/08-deploy.md](docs/standards/08-deploy.md) |
| 文档规范 | [docs/standards/09-docs.md](docs/standards/09-docs.md) |
| Git 规范 | [docs/standards/10-git.md](docs/standards/10-git.md) |
| API 契约 | [docs/api/README.md](docs/api/README.md) |
| 错误码全集 | [docs/api/errors.md](docs/api/errors.md) |
| 管理与运维接口 | [docs/api/admin.md](docs/api/admin.md) |
| 工具规格 | [docs/tools.md](docs/tools.md) |
| 路线图 | [docs/roadmap.md](docs/roadmap.md) |
| 部署运维 | [docs/deploy.md](docs/deploy.md) |
| 隐私说明 | [docs/privacy.md](docs/privacy.md) |
| 国际化 | [docs/i18n.md](docs/i18n.md) |
| 术语表 | [docs/glossary.md](docs/glossary.md) |
