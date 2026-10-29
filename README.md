# 蓝云工具箱（Blue Cloud Toolbit）

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

一个**前后端一体的 IT 工具箱网站**：前端 Vue 3 + TypeScript + Element Plus，后端 Go 1.25 + Gin，密码学能力全部经 [Tongsuo（铜锁）](https://www.tongsuo.net/) 原生库（通过 `github.com/blue-cloud-net/tongsuo-go` 封装）实现。

v1 规划 **98 个工具**，覆盖国密算法（SM2 / SM3 / SM4）、密钥管理、X.509 与 PKI 全链路、编码格式化、文本处理、数据生成、系统与网络辅助、文档解析。

所有工具在**服务端**计算（唯一例外：`websocket-test` 使用浏览器原生 `WebSocket`）。前端构建产物由 Go 进程直接以静态目录方式提供，**不部署 Nginx**。

---

## 特性

| 能力 | 说明 |
| --- | --- |
| 国密算法 | SM2 签名/验签与公钥加解密、SM3 摘要、SM4 多种模式、HMAC-SM3；SM2 密文支持 `C1C3C2` / `C1C2C3` 与 DER 互转 |
| 密钥管理 | SM2 / RSA / EC / Ed25519 / Ed448 / X25519 / X448 密钥生成、PEM ↔ DER 转换、密钥解析、口令加解密、公私钥匹配、JWK 互转 |
| PKI 全链路 | 证书解析、CSR 生成与解析、证书链解析/验证/补全、CRL 生成与解析、OCSP 请求构造与响应解析、PKCS#7 / PKCS#12 打包解析 |
| 通用工具 | Base64、URL、JWT、进制、编码转换、GZip、JSON / JS / SQL / HTML / XML / TOML / YAML 格式化与互转、正则、时间戳、UUID、二维码、CIDR、cron、chmod、curl |
| 在线探测 | SSL/TLS 版本与套件矩阵（含国密双证书）、DNS 多协议多服务商查询、CT 日志检索、在线证书链获取、域名所有权验证 |
| 文档处理 | Markdown 预览、EML 解析（含 MIME 与编码探测）、PDF 数字签名校验 |
| 双语界面 | 简体中文 / English 即时切换 |
| 主题 | 跟随系统的浅色 / 深色主题，8 组预设主色 |
| 隐私 | 无账号；不存储工具输入；调用统计仅记录匿名聚合数据 |
| 交付 | 多阶段 Dockerfile（`deployments/`）：镜像内编译 Tongsuo，运行期只需 glibc |

---

## 技术栈

**前端**

- Vue 3.5 / TypeScript / Vite 6
- Element Plus 2.14（按需自动导入）
- Pinia 3 / vue-router 4 / vue-i18n 11
- UnoCSS 66

**后端**

- Go 1.25 / Gin
- `github.com/blue-cloud-net/tongsuo-go` v0.5.0（cgo → Tongsuo 8.4.0）
- `modernc.org/sqlite`（纯 Go 驱动，匿名调用统计）
- `github.com/miekg/dns`（DNS / DoT / DoH）、`github.com/coder/websocket`（流式探测）

**工具库**

- `gopkg.in/yaml.v3`、`github.com/pelletier/go-toml/v2`
- `github.com/yuin/goldmark` + `github.com/microcosm-cc/bluemonday`（Markdown）
- `github.com/dop251/goja`（JS / SQL 格式化）
- `github.com/skip2/go-qrcode`（二维码）

---

## 快速开始

### Docker（推荐）

镜像尚未发布到公共仓库，从源码构建即可（需要本机有 Docker，**不需要**预装 Tongsuo）：

```bash
# 构建并启动（首次会编译 Tongsuo，约 5-10 分钟）
make docker-up

# 探活 + 目录 + 前端首页
make docker-smoke
```

浏览器打开 <http://127.0.0.1:30234>。数据（匿名统计库）放在命名卷 `toolbit-data`。

> ⚠️ **部署形态说明**：本项目默认只监听 `30234` 端口的 **HTTP**，不内置 TLS。
> 由于工具会在服务端处理密钥与明文，**公网暴露时务必在其前方加一层 TLS 反向代理**（Nginx / Caddy），示例见 [docs/deploy.md](docs/deploy.md)。

### 本地开发

前置条件：Go 1.25+、Node 22+、pnpm 10+，以及已安装的 Tongsuo 8.4.0（默认路径 `/opt/tongsuo`）。

```bash
# 1. 准备 Tongsuo（若尚未安装）
git clone https://github.com/Tongsuo-Project/Tongsuo.git
cd Tongsuo && git checkout 8.4.0
./config --prefix=/opt/tongsuo -Wl,-rpath,/opt/tongsuo/lib64 enable-ntls
make -j"$(nproc)" && sudo make install

# 2. 后端
export TONGSUO_HOME=/opt/tongsuo
export CGO_CFLAGS="-I${TONGSUO_HOME}/include -Wno-deprecated-declarations"
export CGO_LDFLAGS="-L${TONGSUO_HOME}/lib64"
go run ./backend/cmd/toolbit

# 3. 前端（另开终端）
cd web && pnpm install && pnpm dev
```

前端开发服务器默认 <http://127.0.0.1:5173>，并把 `/api` 代理到后端 `30234`。

---

## 常用命令

| 命令 | 说明 |
| --- | --- |
| `make dev` | 同时启动后端与前端（开发） |
| `make build` | 构建后端二进制与前端产物 |
| `make test` | 后端单测 + 前端单测 |
| `make lint` | Go / TS / Markdown / 文档覆盖检查 |
| `make check-docs` | 校验 catalog、错误码、环境变量在文档中的覆盖率 |
| `make docker-up` | 构建镜像并启动（首次会编译 Tongsuo） |
| `make docker-smoke` | 容器冒烟检查（healthz / readyz / 目录 / 首页） |
| `make catalog-validate` | 校验 `backend/catalog/tools.yaml` 结构与唯一性 |

---

## 目录结构

仓库分为**两个独立构建单元**：`backend/`（Go 模块）与 `frontend/`（pnpm 项目）。

```text
.
├── backend/                 # 后端：独立 Go 模块
│   ├── cmd/toolbit/         # 程序入口
│   ├── catalog/             # 工具目录（内嵌 backend/catalog/tools.yaml）
│   └── internal/
│       ├── api/             # HTTP 分发器（dispatcher）
│       ├── buildinfo/       # 版本信息（ldflags 注入）
│       ├── catalog/         # 工具目录解析与校验
│       ├── codecsvc/        # 编码 / PEM 归一化
│       ├── cryptosvc/       # 唯一允许 import tongsuo-go 的层
│       ├── httpx/           # 统一响应信封、错误码、中间件
│       ├── netguard/        # 出网工具的 SSRF 防护
│       ├── stats/           # SQLite 匿名调用统计
│       ├── toolerr/         # 错误码注册表
│       ├── tools/           # 各领域工具实现
│       └── webui/           # 静态资源与 SPA 回退
├── frontend/                # Vue 前端
├── deployments/             # Dockerfile / compose / .env.example
├── docs/                    # 架构、规范、API、工具规格
├── scripts/                 # 文档与一致性校验脚本
└── .github/workflows/       # CI
```

---

## 文档

| 文档 | 内容 |
| --- | --- |
| [docs/README.md](docs/README.md) | 文档索引与阅读路径 |
| [docs/architecture.md](docs/architecture.md) | 架构总览与关键决策 |
| [docs/standards/](docs/standards/00-intro.md) | 开发规范（前端 / 后端 / 密码学 / 错误 / 测试 / 安全 / 可观测 / 部署 / 文档 / Git） |
| [docs/api/README.md](docs/api/README.md) | API 契约总览（信封、错误、通用类型、限流） |
| [docs/api/errors.md](docs/api/errors.md) | 错误码与 warnings 全表 |
| [docs/api/tools/](docs/api/tools/crypto.md) | 98 个工具的请求 / 响应契约 |
| [docs/tools.md](docs/tools.md) | 98 个工具的产品规格 |
| [docs/roadmap.md](docs/roadmap.md) | 路线图与发布清单 |
| [docs/deploy.md](docs/deploy.md) | 部署与运维指南 |
| [docs/privacy.md](docs/privacy.md) | 隐私说明 |
| [docs/glossary.md](docs/glossary.md) | 术语表 |
| [AGENTS.md](AGENTS.md) | Agent / 贡献者工作指南 |

---

## 安全与隐私

- 工具输入在服务端内存中处理，**不写数据库、不写日志、不写本地存储**。
- 调用统计只记录工具标识、耗时、字节数、是否成功与一个**按日轮换盐值哈希后的访客指纹**，不记录原始 IP、User-Agent 原文或任何输入内容。
- 访问日志只记录 `slug`、状态码、耗时、字节数、错误码与 `requestId`。
- 密钥类工具页会明确提示「输入会发送到服务端处理」。
- 出网类工具统一经过 SSRF 防护：解析 → 私网校验 → 固定 IP → 限制重定向 → 限制响应体积。
- 详见 [docs/standards/06-security.md](docs/standards/06-security.md) 与 [docs/privacy.md](docs/privacy.md)。

---

## English Summary

**Blue Cloud Toolbit** is a full-stack IT toolbox website.
The frontend is Vue 3 + TypeScript + Element Plus (bilingual zh-CN / en-US);
the backend is Go 1.25 + Gin.
All cryptographic operations are performed server-side through **Tongsuo**
(OpenAtom 铜锁) via the `github.com/blue-cloud-net/tongsuo-go` wrapper —
SM2, SM3, SM4, X.509, CSR, CRL, OCSP, PKCS#7/#12 and more.

The v1 release targets **98 tools** across 10 categories.
The frontend build output is served directly by the Go process from a static
directory (no Nginx), and the project ships as a multi-stage container image
published to GHCR for `linux/amd64` and `linux/arm64`.

There is **no user account system**. Tool inputs are never persisted; usage statistics are anonymous aggregates only.

---

## License

[Apache-2.0](LICENSE) © blue-cloud-net
