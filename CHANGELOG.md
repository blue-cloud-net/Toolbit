# Changelog

本文件记录 Blue Cloud Toolbit 的所有显著变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

部署产物与容器探活子命令：仓库从「文档声明了部署形态」变为「真的能构建、能跑」。

### Added

- 品牌标识（Toolbit 徽标）：矢量重绘，贴合品牌名语义
  - 图形为「T 型手柄 + 六角批头」：`T` 既是首字母，也构成 T 型螺丝刀轮廓；
    底端尖头六角是批头（bit），合起来正好是「Tool + Bit」
  - `frontend/public/favicon.svg` 为唯一矢量源；`favicon-32.png` / `favicon.ico` /
    `apple-touch-icon.png` 均由它生成，不手工维护第二套图形
  - `frontend/src/components/common/BrandMark.vue`：顶栏品牌标识组件，内联 SVG
  - 配色走新增的 `--tb-brand-from` / `--tb-brand-to` / `--tb-brand-fg` 令牌，
    **固定靛蓝**，不随主色切换 —— 徽标是品牌锚点
  - 顶栏原有的 `◆` 占位字符与 `.header__logo` 样式一并移除
- `deployments/Dockerfile`：四阶段构建（`tongsuo` → `frontend` → `backend` → `runtime`）
  - Tongsuo 8.4.0 在镜像内编译安装到 `/opt/tongsuo`（`--libdir=lib64 shared enable-ntls`），
    安装路径与 `tongsuo-go` 的 cgo 指令、rpath 严格一致
  - 源码包先按 sha256 校验再解包；基础镜像固定 digest
  - 运行镜像基于 `debian:bookworm-slim`，非 root（UID 10001），含 CA 证书与时区数据，
    只带二进制 + `/app/web` + Tongsuo 运行库，**不**含 Go / Node / 编译工具链
- `deployments/docker-compose.yml`：单容器编排，命名卷 `toolbit-data`、
  `cap_drop: ALL`、`no-new-privileges`、内存上限、日志轮转、`STOPSIGNAL SIGTERM`
- `deployments/.env.example`：16 个 `TOOLBIT_*` 与 6 个编排变量的默认值模板（无任何密钥）
- `.dockerignore`：排除依赖、构建产物、文档、测试与密钥
- `toolbit healthcheck` 子命令（`backend/cmd/toolbit/healthcheck.go`）：
  探 `http://127.0.0.1:<TOOLBIT_ADDR 端口>/readyz`，3 秒超时，非 200 退出 1、未知子命令退出 2，
  不输出响应体与配置值 —— 兑现 `docs/api/admin.md` §9.1 早已写明的契约
- Makefile 新增 `docker-build` / `docker-up` / `docker-smoke` / `docker-logs` / `docker-ps` /
  `docker-shell` / `docker-down` / `docker-clean`

### Changed

- 部署产物目录由文档中的 `deploy/` 统一为 **`deployments/`**（`AGENTS.md`、
  `docs/architecture.md`、`README.md`、`.gitignore` 同步）
- 容器内路径口径统一：前端产物 `/app/web`、统计库 `/app/data/toolbit.db`
  （原先 `deploy.md` 写 `/app/web` + `stats.db`，`architecture.md` 写 `/app/wwwroot` + `toolbit.db`）
- `docs/standards/08-deploy.md`：阶段表、基础镜像选型、Tongsuo `--libdir=lib64` 约定、
  digest 固定口径、探针实现方式、compose 位置与禁止清单更新
- `docs/deploy.md`：Docker 章节改写为 `make docker-up` 为主线，补编排变量表与容器故障排查项
- `docs/api/admin.md` §9.1：补充探针地址来源、退出码与输出约束

## [0.7.0] - 2026-10-29

网络探测：出网能力就继，SSRF 防护全覆盖。
**新增 7 个工具**（累计 43 个已实现），`ct-search` 本版未实现。

本版引入统一的出网防护层 `netguard`，并把 `ssl-check` 做成 WebSocket 流式端点。

### Added

- 出网防护层 `backend/internal/netguard/`
  - 私网 / 回环 / 链路本地 / `0.0.0.0/8` / 运营商级 NAT / 文档与保留网段全部拦截
  - IPv4-mapped IPv6（`::ffff:127.0.0.1`）先还原再判定，无法绕过
  - 建连使用**已校验的 IP**（防 DNS rebinding），不二次解析
  - 端口白名单固定为 43 / 53 / 80 / 443 / 853，用户无法指定任意端口
  - 重定向逐跳校验，最多 3 跳；禁用环境代理
  - `TOOLBIT_NETWORK_ALLOWLIST` 支持精确主机名与 `*.suffix`
- 工具 `cidr-calc`：网段计算、地址范围反推、按前缀拆分；IPv6 大数以十进制字符串表达
- 工具 `dns-lookup`：`udp` / `tcp` / `dot` / `doh` 四种协议，逐服务器结果与耗时
- 工具 `dns-tester`：RDAP 优先、whois（TCP 43）回退；DNSSEC 逐级 DS 检查；PTR + 正向复核
- 工具 `ssl-check`：WebSocket 流式 TLS 探测，四个阶段的进度帧
- 工具 `cert-chain-fetch`：只呈现服务端实发链，附链摘要与验证结论
- 工具 `domain-verify`：`dns-txt` / `dns-cname` / `http-file` 三种方式
- 工具 `websocket-test`：浏览器原生直连，**无**后端端点
- WebSocket 端点 `GET /api/v1/tools/ssl-check/stream`
  - 首帧 5 秒超时、心跳 30 秒、空闲 60 秒兜底、总预算 30 秒
  - 同 IP 同时最多 1 条连接，超出在升级前返回 `429`
  - 关闭码：`1000` / `1008` / `1011`
- 前端 8 个工具页与 16 份双语词条；流式页面自带四态状态机
- 契约测试数据 5 份；`netguard` 与网络工具层单元测试

### Changed

- catalog 传输方式校正
  - `ssl-check` → `transport: websocket`（此前误标为 `json`）
  - `websocket-test` → `transport: client`（与「纯前端」的定义对齐）
  - `cert-chain-fetch` → `transport: json`，请求体上限降到 1 MiB
- 分发器按 `transport` 分流：`client` 工具返回 `404`，`websocket` 工具的 JSON 通道返回 `405`
- 端口**白名单**判定从字段校验下移到 `netguard`：字段层只校验范围，越界端口返回
  `SSRF_BLOCKED`，策略只有一个事实来源
- `toolset.Definitions` 需要注入出网守卫，进程入口与 API 测试共用同一份装配

### Fixed

- `cidr-calc` 的地址族字段此前返回位宽（32/128），改为 `4` / `6`

### Known limitations

- `ct-search` **未实现**：公开证书透明度日志查询服务没有稳定 SLA，
  catalog 保留登记，调用返回 `NOT_IMPLEMENTED`
- TLS 客户端使用 Go 的 `crypto/tls`：仅限出网探测，链验证仍走 `pkisvc.VerifyChain`
  （条款见 `docs/standards/03-crypto.md`）
- 本地信任库可能缺失：`ssl-check` 届时只能判定链内自洽，
  响应里用 `verified` / `implicitTrust` 显式区分
- 系统解析器（`/etc/resolv.conf`）属运维配置，不做私网校验；
  用户自定义的 DNS 服务器仍要过出网防护
- TLS 1.3 的密码套件无法通过配置裁剪

---

## [0.6.0] - 2026-10-29

PKI 生成、吊销与容器：能签发证书、处理吊销，并读写 PKCS#12 / PKCS#7 容器。
**新增 12 个工具**（累计 36 个已实现）。

本版把 `pkisvc` 从「X.509 只读侧」扩展到**写侧**：CSR / 证书 / CRL 签发、
离线吊销判定、OCSP 请求与响应、PKCS#12 与 PKCS#7 容器。

### Added

- PKI 服务层新增 8 个文件（`backend/internal/pkisvc/`）
  - `certgen.go`：DN 构造（`TryAdd` 白名单）、随机序列号、
    **分步构建**的证书签发（`NewCertificate` → `Set*` → `Add*` → `Sign`）
  - `csrgen.go`：CSR 生成（`NewEmptyCertificateRequest` 分步构建）
  - `crlgen.go`：CRL 签发与吊销原因码白名单（拒绝保留值 7）
  - `revocation.go`：`RevocationList` 句柄与 `CheckRevocation`（**先验签后判定**）
  - `ocsp.go`：OCSP 请求生成与响应解析（含可选验签）
  - `pkcs12.go`：打包 / 解析 / 改密；口令错误按**操作名**判别
  - `pkcs7.go`：证书袋打包与提取
  - `store.go`：信任库句柄（每请求新建）
- 工具 `csr-gen`：PKCS#10 请求，按证书类型预设 KU / EKU
- 工具 `cert-gen`：自签或 CA 签发，SKID / AKID 自动补齐
- 工具 `crl-gen`：CA 签发 CRL，原因码白名单
- 工具 `revocation-check`：离线吊销判定（未验签时只报 `unverified`）
- 工具 `ocsp-request` / `ocsp-response`：OCSP 请求与响应（可选 `trustedRoots` 验签）
- 工具 `pfx-pack` / `pfx-parse` / `pfx-change-password`：PKCS#12 容器
- 工具 `p7b-pack` / `p7b-extract`：PKCS#7 证书袋
- 工具 `caa-gen`：CAA 记录生成（纯字符串拼装，不进 `pkisvc`）
- 前端 12 个工具页与 24 份双语词条；夹具 `frontend/src/tools/certificates/fixtures-v060.ts`
  （一次性生成的**测试密钥**，非真实密钥）
- 契约测试数据 12 份；服务层与工具层单元测试各新增一批用例
- 生成脚本 `scripts/gen-v060-fixtures.sh` / `gen-v060-contracts.py` / `gen-v060-web-fixtures.py`

### Changed

- **契约扩展**（只加**可选**字段）
  - `revocation-check` 新增 `issuer`（提供时先验签）、`atTime`
  - `ocsp-response` 新增 `trustedRoots`，响应新增 `signatureVerified`
- **契约收窄**
  - `cert-gen.templateCert` **只**复用公钥与有效期，**不**拷贝扩展
  - `crl-gen.revoked[]` **只**接受证书，去掉「按序列号吊销」选项
  - `revocation-check` 的 `status` 新增 `unverified` 取值；未验签时**绝不**返回 `good`
- 新增警告码 `CRL_UNVERIFIED`
- 路线图：`docs/roadmap.md` §9 全部任务勾选，§9.3 补 M-T9 – M-T12

### Fixed

- 口令错误判别不再依赖 OpenSSL 错误栈文本：该栈在成功调用后不会清空，
  同进程内前一次失败的条目会覆盖本次原因，导致「容器损坏」被误报成「口令错误」

### Known limitations

- `cert-gen.templateCert` **不拷贝扩展**：上游读证书只能拿到扩展 DER，
  没有「按 DER 添加扩展」的入口（见 `docs/upstream/tongsuo-go-issues.md` I-19）
- `crl-gen` **无法按序列号吊销**：上游 `CRLBuilder.Revoke` 只接受证书对象（见 I-20）
- 上游的便捷签发入口会**立即签名**，签名后 `Add*` 被「已签名锁」拒绝（见 I-21）；
  Toolbit 因此固定走分步构建
- `pfx-change-password` **只**支持本库打包的容器：上游 `PKCS12_newpass`
  无法对外部实现产出的容器改密（见 I-22）
- `p7b-pack` / `p7b-extract` **只做证书袋**：上游 PKCS#7 尚无签名 / 加密（见 I-11）
- 序列号上限 8 字节：上游用 `int64` 承载（与 I-18 同源），超出返回 `SERIAL_INVALID`
- `ocsp-response` 未提供 `trustedRoots` 时**不验签**，只解读响应内容

---

## [0.5.0] - 2026-10-29

PKI 解析与链验证：能读懂证书，并能判定一条链是否可信。
**新增 8 个工具**（累计 24 个已实现）。

本版把 `pkisvc` 从「密钥与格式」扩展到完整的 **X.509 只读侧**：
证书 / CSR / CRL 解析、链排序、链验证、链修复与 ASN.1 展开。

### Added

- PKI 服务层新增 8 个文件（`backend/internal/pkisvc/`）
  - `cert.go`：证书加载与解析、PEM / DER / base64 / hex 四种输入；
    主体 / 签发者 / 有效期 / 扩展 / SAN / KU / EKU / SKID / AKID；
    指纹支持 `MD5` / `SHA1` / `SHA256` / `SHA384` / `SHA512` / `SM3`
  - `pemder.go`：PEM 块拆分、类型过滤（cert / key / csr / crl）、PEM ↔ DER 互转
  - `csr.go`：CSR 解析与自签名校验；SAN 由扩展手工解析
  - `asn1.go`：DER 展开为 TLV 树 + 十六进制转储；调用上游**之前**自测深度 / 节点 / 体积
  - `chain.go`：AKID → SKID 优先的链排序与链拆分
  - `chainverify.go`：逐层验证、域名匹配、`X509_V_ERR_*` 结构化错误项
  - `chainrepair.go`：从中间证书与受信任根里补链
  - `crl.go`：CRL 解析（签发者 / 时间窗 / 编号 / 吊销条目）
- 工具 `cert-parse`：证书详情、有效性状态、指纹、原始 DER
- 工具 `csr-parse`：主体、SAN、公钥、扩展、挑战口令**存在性**、自签名校验结果
- 工具 `pem-der`：PEM ↔ DER 互转、标签识别、块数
- 工具 `asn1-view`：DER 结构树 + 十六进制转储（catalog 标记 `wide`）
- 工具 `cert-chain-parse`：链排序、是否已排序、是否缺根、是否自签名
- 工具 `cert-chain-verify`：链可信性、错误列表、逐张证书结论、域名匹配、隐含信任锚标志
- 工具 `cert-chain-repair`：补全缺失的中间证书（纯本地）
- 工具 `crl-parse`：CRL 签发者、时间窗、签名算法、编号、吊销条目
- 共享输入类型 `backend/internal/tools/pembundle.go`：`tools.PemBundle`，
  与 `KeyInput` 共用同一套传输层解码逻辑
- 前端 8 个工具页与 16 份双语词条；共享组件
  `CertTextInput.vue`（证书族输入）与 `CodeEditor.vue`（**自渲染**只读树 / hex 视图）
- 测试夹具 `backend/testdata/fixtures/pki/`（自签根 CA + 中间 CA + 叶证书 + CSR + CRL）
  与生成脚本 `scripts/gen-pki-fixtures.sh`；**不含任何私钥**
- 契约测试数据 8 份；服务层与工具层单元测试各 1 个文件

### Changed

- **契约扩展**（只加**可选**字段，不改已有字段语义）
  - `cert-chain-repair` 新增 `intermediates`
  - `cert-chain-verify` 的 `trustedRoots` 语义明确为「受信任根列表」
  - `cert-parse` 的 `fingerprints` 扩展为**全部**哈希算法
- `cert-chain-parse` 的 `ordered: false` **只**表示输入需要重排，**不是**错误
- `asn1-view` 的请求体上限由 8 MB 降为 **1 MB**（DER 展开的内存放大）
- 契约测试与 501 用例改为**从 `toolset.Definitions()` 推导**，
  新增工具时不再需要手改测试清单
- 路线图：`docs/roadmap.md` §8 全部任务勾选；§8.3 补 L-T9 – L-T11 三条技术决策

### Fixed

- `backend/internal/api` 的「未实现工具返回 501」用例改用**推导**出来的未实现 slug，
  避免新工具落地后用例失效

### Known limitations

- `cert-chain-verify` **没有系统信任根可用**：上游未暴露默认信任库
  （见 `docs/upstream/tongsuo-go-issues.md` I-17）。
  未传入受信任根时，链内的自签名根被当作**隐含信任锚**，
  此时 `valid: true` **只**说明链内自洽；响应里的 `implicitTrust` 标出这种情况
- `csr-parse` 的 CSR 句柄**无法显式释放**：上游 `CertificateRequest` 没有公开的
  `Close()`（见 I-15），只能依赖 `finalizer` —— v0.5.0 唯一一处句柄释放例外
- `crl-parse` 的吊销序列号**超过 64 位时输出空字符串**：上游用 64 位整数承载序列号，
  超长时溢出为 `-1`（见 I-18），工具**不**输出溢出的错误值
- `asn1-view` 的深度 / 节点上限由 **Toolbit 侧**测量实现（上游 `asn1.Parse`
  无上限参数，见 I-16）；契约约定的 64 层 / 10000 节点不变
- CSR 的 SAN 由扩展**手工**解析（上游未暴露 `SAN()` 方法），只覆盖
  email / DNS / URI / IP / RegisteredID 五类
- 证书签发、吊销判定、OCSP、PKCS 容器均在 v0.6.0 范围，本版**不含**

## [0.4.0] - 2026-10-29

密钥管理：交付密钥全生命周期（生成 → 转换 → 解析 → 运算 → 配对）。
**新增 6 个工具**（累计 16 个已实现）。

本版同时建立 v0.4.0 – v0.6.0 共用的底座：新增 `pkisvc` 密码学层与
`toolset` 工具装配层，把「密码学原语」与「密钥 / PKI 结构」分开。

### Added

- PKI 服务层 `backend/internal/pkisvc/`：密钥与 **PKI 领域**
  - `Generate`：RSA / SM2 / EC / Ed25519 / Ed448 / X25519 / X448；
    公钥是**独立持有**的句柄，可与私钥分别释放
  - `Convert`：PKCS#1 / PKCS#8 / SPKI / DER / PEM 互转；
    DER 源靠**试加载**识别编码，加密私钥转换后**保持加密**
  - `Inspect`：算法、类型、位数、曲线、公钥指纹（SPKI DER 的 SHA-256）与 RSA 私钥参数
  - `EncryptPrivateKey` / `DecryptPrivateKey` / `ChangePassphrase` / `ExtractPublicKey`
  - `Match`：私钥 ↔ 公钥 / 证书，依据公钥指纹判定
  - `JWKFromPEM` / `PEMFromJWK` / `XMLFromPEM` / `PEMFromXML`：RFC 7517 与 .NET `RSAKeyValue`
- 工具 `keygen`：7 种算法；`privateFormat` 支持 `pkcs8` / `pkcs1`；
  可选用口令加密私钥（AES-256-CBC / DES-EDE3-CBC）；生成走重操作信号量
- 工具 `key-convert`：格式互转 + 源密钥识别（算法 / 编码 / 是否加密 / 位数 / 曲线）
- 工具 `key-parse`：密钥元信息、RSA 私钥参数与规范化 PEM
- 工具 `key-ops`：加密、解密、提取公钥、修改口令
- 工具 `key-match`：私钥 ↔ 公钥 / 证书配对检测，给出双方公钥指纹
- 工具 `jwk-xml-convert`：JWK / PEM / XML 三向互转
- 工具装配层 `backend/internal/toolset/`：汇总各分类 `Definitions()`，
  进程入口与 API 测试共用；测试守住 slug 唯一与 catalog 登记
- 共享输入类型 `backend/internal/tools/bytes.go` `keyinput.go`：
  `tools.Bytes` 与 `tools.KeyInput`，各工具包用类型别名复用
- 前端 6 个工具页与 12 份双语词条；共享夹具 `frontend/src/tools/keys/fixtures.ts`
  （一次性生成的**测试密钥**，非真实密钥）
- 契约测试数据 6 份；服务层与工具层单元测试各 1 个文件

### Changed

- **架构**：新增 `backend/internal/pkisvc` 作为**第二个**允许 `import tongsuo-go` 的层；
  `AGENTS.md` §4.1 与 `docs/architecture.md` 的依赖图同步更新
- **依赖守卫**：`.golangci.yml` 的 `depguard` 规则由「仅 cryptosvc」改为
  「cryptosvc + pkisvc」双白名单；`no-stdlib-crypto` 新增 `crypto/x509` 禁令
- **catalog**：`revocation-check` 由 `rateClass: network` + `requiresNetwork: true`
  改为 `standard` + `false`——其契约自 v0.1.0 起就是**离线**的「证书 + CRL」判定
- **路线图**：`docs/roadmap.md` 新增 §6.8「v0.4.0 起的跨版本公共基座」，
  并把 §7 / §8 / §9 细化为与 v0.3.0 同构的可执行计划（交付物 / 技术决策 /
  分阶段任务表 / 完成标准 / 不在本版范围 / 风险）

### Fixed

- `backend/internal/api` 的两处「未实现工具返回 501」用例改用仍未实现的工具
  （`cert-parse` / `cert-chain-fetch`），避免 `keygen` 落地后用例失效

### Known limitations

- `keygen` 的 `count` 字段**当前只支持 1**：响应结构只能表达一把密钥，
  批量生成需要新增响应字段（属破坏性契约变更），已在
  `docs/api/tools/keys.md` 与 `docs/roadmap.md` §7.7 双处登记
- `keygen` 的 `publicFormat: pkcs1` 与 `key-convert` 的公钥 → `pkcs1`
  **暂不支持**：上游 `tongsuo-go` 缺 PKCS#1 公钥的导出 / 加载入口
  （见 `docs/upstream/tongsuo-go-issues.md` I-14）
- `jwk-xml-convert` 的 JWK 侧**只支持 RSA 与 EC**（上游无 OKP，见 I-12），
  XML 侧**只支持 RSA**（见 I-13）
- `p7b-pack` / `p7b-extract`（v0.6.0）**只做证书袋**：上游 PKCS#7 尚无签名 / 加密（见 I-11）

## [0.3.1] - 2026-10-28

密码学文件工具：交付 `sm4-file-crypt`，让大文件加解密可用。
**新增 1 个工具**（累计 10 个已实现）。

本版打通 multipart 通道：此前分发器只认 JSON，
既无法接收表单，也无法返回二进制流。

### Added

- 文件容器 `backend/internal/cryptosvc/tbfc.go`：`TBFC` 分块 GCM 容器
  - 头（22 字节）：magic `TBFC`(4) + version(1) + flags(1) + chunkSize(4) + baseNonce(12)
  - 块：`length`(4) + ciphertext + `tag`(16)；`length = 0` 表示结束
  - 分块 nonce：`baseNonce[0:8] || uint32be(chunkIndex)`
  - AAD：`容器头 || uint32be(chunkIndex)`，同时绑定容器头与块索引
  - 流式读写，内存占用约 2×chunkSize，与文件大小无关
- multipart 通道 `backend/internal/tools/multipart.go`
  - `MultipartRequest` / `MultipartResult` / `MultipartHandler`
  - `MultipartForm`：把 `multipart.Reader` 的顺序读取包装成「先收小字段、再收文件」
  - `SanitizeFilename` / `ContentDisposition`：文件名来自客户端，**不可信**，必须清洗
- 分发器分流 `backend/internal/api/dispatch_multipart.go`
  - 按 catalog 的 `transport` 分流：`multipart` 走表单 + 二进制流通道
  - 二进制响应不套信封，警告改走 `X-Warnings` 响应头
  - 失败时**仍返回 JSON 信封**：错误在响应头发出之前就能确定
- 工具 `sm4-file-crypt`：TBFC 容器，分块 GCM，100 MB 上限
  - 解密**先同步**校验容器头与第一块：密钥不对、容器被篡改、不是容器
    这三类错误因此返回干净的 JSON 错误，而不是「200 + 被截断的下载」
- 前端 `callToolBinary`：multipart 请求 + 二进制响应，带上传进度
- 共享组件 `FileDropZone`：点击选择 + 拖放，单文件
- 工具页 `sm4-file-crypt`：操作、密钥、基础 nonce、分块大小与拖放区；结果以「下载」给出
- 组合式函数 `useFileToolRun`：与 `useToolRun` 同构，但结果是二进制文件，额外提供上传进度
- 契约测试执行器支持 multipart 分支；契约数据落在 `backend/testdata/contract/sm4-file-crypt.json`

### Changed

- 密码学规范 `docs/standards/03-crypto.md` §11 重写：容器格式逐字节落档，
  补「首块预校验」「multipart 顺序约束」两节
- 加密 API 契约 `docs/api/tools/crypto.md`：`sm4-file-crypt` 章节按实现重写，
  含容器格式与错误码表
- 通用 API 契约 `docs/api/README.md`：`X-Warnings` 的格式明确为 JSON 数组；
  §8 补 `transport` 分流；§4.5 与 §12.4 补「文件必须是最后一个 part」
- 架构文档 `docs/architecture.md`：§4.1 双通道、§4.2 两种工具形态、§4.3 补文件容器
- 路线图 `docs/roadmap.md`：v0.3.1 补齐 8 项任务清单与完成标准

### 破坏性变更

- multipart 工具的**文件部分必须是最后一个 part**：客户端必须先发所有普通字段，
  最后再发文件。服务端读到文件后就无法回头取后面的字段
- multipart 工具的**失败响应仍是 JSON 信封**，不是二进制流；
  客户端必须按 `Content-Type` 分流

## [0.3.0] - 2026-10-28

密码学核心：国密能力就绪，建立 `cryptosvc` 服务层与句柄生命周期规范。
**新增 7 个工具**（累计 9 个已实现）。

### Added

- 密码学服务层 `backend/internal/cryptosvc/`：**唯一**允许 import `tongsuo-go` 的层，所有密码学运算在此转调 Tongsuo
  - `Session` 句柄作用域：LIFO 释放、`Close` 幂等、警告累积、重操作信号量（≤ 2）、orphan 记录
  - Session 经 `ctx` 传递，**不**改变 `tools.Handler` 签名
  - `Digest` / `MAC` / `SymEncrypt` / `SymDecrypt` / `Sign` / `Verify` / `AsymEncrypt` / `AsymDecrypt` / `ECDHSharedSecret` / `ECDHDerive` / `RandomBytes` 服务方法
  - 底层错误统一映射为 `toolerr` 错误码，不回显底层英文文本
  - 密钥封装 `PrivateKey` / `PublicKey`：显式 `Close`，源缓冲用后清零
- 编码服务层 `backend/internal/codecsvc/`：utf8 / hex / base64 / base64url 归一化与 PEM / DER 识别
- 加密类工具 `backend/internal/tools/crypto/`：
  - `hash-calc` —— SM3 / MD5 / SHA1 / SHA224 / SHA256 / SHA384 / SHA512 多算法摘要，弱哈希告警
  - `hmac-gen` —— HMAC-SM3 与 HMAC-SHA 系列
  - `password-gen` —— CSPRNG 密码生成，拒绝采样避免模偏差，附带熵估算
  - `symmetric-crypt` —— SM4 / AES，ECB / CBC / CTR / OFB / CFB / GCM，GCM 自动生成 nonce 并回传
  - `signature` —— SM2 / RSA / ECDSA / Ed25519 / Ed448 签名与验签
  - `asymmetric-crypt` —— SM2 / RSA 公私钥加解密
  - `ecdh` —— P-256 / P-384 / P-521 / X25519 / X448 协商 + HKDF 派生
- 前端 7 个加密工具页 `frontend/src/tools/crypto/`，含共享 `KeyInput` 组件（密钥只存于组件局部状态，**不**写浏览器存储）
- 双语词条：7 个工具的 `zh-CN` / `en-US` 十项字段全覆盖；词条一致性测试改为用 `import.meta.glob` 自动收集，新增工具无需再改测试
- 契约测试执行器 `backend/internal/api/contract_test.go`：遍历 `backend/testdata/contract/*.json` 校验状态码、错误码、响应数据与警告
- 上游能力缺口登记 `docs/upstream/tongsuo-go-issues.md`：10 条缺口的现象 / 期望 / 验收标准 / 下游影响
- CGO 构建基线：`Makefile` 新增 `check-tongsuo`，`CGO_CFLAGS` / `CGO_LDFLAGS` 指向 `/opt/tongsuo`，以 rpath 固定运行期库路径

### Changed

- 密码学规范 `docs/standards/03-crypto.md`：新增 §1.1 运算归属、对称加解密的「算法 × 模式 × 填充」支持矩阵、签名结构预校验规则、GCM nonce 回传约定
- 架构文档 `docs/architecture.md`：§4.3 服务层改为与实现一致的包级函数清单；新增关键决策 D11「密码学运算只转调，不自研」
- 加密 API 契约 `docs/api/tools/crypto.md`：逐字段对齐实现，新增算法支持矩阵与实现状态列
- 产品文档 `docs/tools.md`：加密章节新增状态列，移除未实现算法的承诺
- 路线图 `docs/roadmap.md`：v0.3.0 补齐 6 个阶段的任务清单与验收标准；新增 v0.3.1（`sm4-file-crypt`）
- 工程门禁 `.golangci.yml`：新增 `depguard` 规则 `no-stdlib-crypto`，禁止在 `cryptosvc` 之外导入标准库密码学包
- 协作契约 `AGENTS.md`：新增不变量 §4.15（密码学运算一律经 `cryptosvc` 转调）、两条禁止项与上游 issue 登记流程

### 破坏性变更

- 加密类 API 契约按实际实现收敛，**删除**以下尚未实现的承诺：
  - `symmetric-crypt`：`DES` / `3DES` 算法；AES 的 `OFB` / `CFB` 模式；AES 的 `zero` / `none` 填充
  - `asymmetric-crypt`：`ECIES` 算法；请求参数 `sm2Order` / `sm2Prefix04` / `sm2Format`（SM2 输出恒为 `C1C3C2` + `04` 前缀 + 裸格式）
  - `signature`：请求参数 `format` / `saltLength`；响应中的 `format` 字段（签名值改为 Base64 字符串）
  - `hmac-gen`：算法名改为 `HMAC-SM3` / `HMAC-SHA256` 等带前缀形式；移除文件输入
- `sm4-file-crypt` **未在本版交付**，端点返回 `501 NOT_IMPLEMENTED`，顺延至 v0.3.1
- `tongsuo-go` 能力缺口不做兜底：`DES` / `3DES` / `ECIES` / SM2 压缩点密文 / RSA-PSS 的 DER 形式**均不可用**，缺口登记在上游 issue 文档

### 依赖

- adds: `github.com/blue-cloud-net/tongsuo-go` v0.5.0

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

[未发布]: https://github.com/blue-cloud-net/Toolbit/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/blue-cloud-net/Toolbit/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/blue-cloud-net/Toolbit/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/blue-cloud-net/Toolbit/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/blue-cloud-net/Toolbit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/blue-cloud-net/Toolbit/releases/tag/v0.1.0
