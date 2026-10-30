# 工具候选清单

> 状态：Stable ｜ 适用版本：v0.11.0+ ｜ 最后更新：2026-10-30

本文记录**候选工具**的能力定义、目标分类、与既有工具的关系、依赖、实现边界与优先级，并标注**落地状态**。

与 [roadmap.md](roadmap.md) §16.1 的关系：§16.1 记录 v0.8.0 评审时识别的候选，本文记录其后的全部批次并给出优先级与边界，是候选工具的**唯一记录处**。

**约定**

- 候选工具**未进入** [`backend/catalog/tools.yaml`](../backend/catalog/tools.yaml) 前不得实现；确定落地后按 [../AGENTS.md](../AGENTS.md) §8 的完整流程登记，并递增工具总数。
- slug 一经登记**禁止**改名（[roadmap.md](roadmap.md) §1.3）；候选阶段可调整。
- 「与既有工具的关系」列用于避免重复造能力；既有工具的增强**不新增 slug**、**不递增工具总数**。
- 本表**不记录**任何外部项目名称（[../AGENTS.md](../AGENTS.md) §5 禁止清单）。
- 成本口径：**S** ≈ 半个工作日，**M** ≈ 1–2 天，**L** ≈ 3 天以上。

> **当前状态**：本文列出的候选已于 v0.8.0 **全部交付**，slug 均已在 catalog 登记。
> 下表保留能力定义与取舍理由，作为交付记录与后续增强的输入；**待办候选为零**，
> 新一批候选评审后再补充。

---

## 1. 批次 A —— 已交付

### 1.1 新增工具（✅ v0.8.0 已交付）

| slug | 分类 | 能力 | 与既有工具的关系 | 依赖 | 成本 | 优先级 |
| --- | --- | --- | --- | --- | --- | --- |
| `text-diff` | 文本 | 文本与 JSON 差异对比：JSON 模式按结构化路径标注增删改，文本模式按行级 LCS 标注 | 无；[roadmap.md](roadmap.md) §16.1 的 `json-diff` 并入本项，不单独登记 | 零新依赖 | M | P0 |
| `nginx-config` | 系统 | Nginx 可视化配置生成：全局 / 站点双模式，覆盖反向代理、负载均衡、SSL、Gzip、CORS、WebSocket | 无；与 `curl-generator` 同属「配置生成」族，共享参数区组件 | 零新依赖 | M | P0 |
| `firewall-rules` | 系统 | 防火墙规则生成：按场景拼装 iptables / UFW / nftables 规则，支持批量导出 | 无；与 `nginx-config` 同族 | 零新依赖 | M | P0 |
| `regex-validate` | 文本 | 正则表达式语法校验：只校验不匹配，输出是否合法、错误位置与方言提示 | 与 `regex-tester` 共用同一条 RE2 编译路径与方言判定，不重复实现方言逻辑 | 零新依赖 | S | P0 |
| `regex-build` | 文本 | 正则表达式生成：由结构化选项（字符类、量词、分组、前后缀、常用模板）拼装表达式并给出释义 | 无；生成结果仍按 RE2 方言校验，非法组合即时报错 | 零新依赖 | M | P1 |
| `device-headers` | 网络 | 客户端请求头生成与 User-Agent（UA）解析：按平台与浏览器版本生成完整请求头，反向解析 UA 为结构化字段 | 无；`http-status` / `mime-lookup` 同属速查族，本项属「构造 + 解析」 | 零新依赖 | M | P0 |
| `ssh-keygen` | 密钥 | SSH 私钥生成：产出 OpenSSH 格式（`openssh-key-v1`）密钥对、`authorized_keys` 公钥行与 SHA256 指纹 | 与 `keygen` 互补：`keygen` 产出 PKCS#8 / SPKI PEM，本项产出 OpenSSH 格式；密钥材料复用既有算法能力 | 零新依赖（受 §3 边界约束） | M | P1 |

### 1.2 既有工具增强（✅ v0.8.0 已交付）

| 既有 slug | 分类 | 增强内容 | 依赖 | 成本 | 优先级 |
| --- | --- | --- | --- | --- | --- |
| `password-gen` | 加密 | 在既有 `customCharset` / `requireEachClass` 之上补充：可发音口令生成、易混淆字符排除集（`I` `l` `O` `0` `1`）、批量生成、策略模板（长度 / 字符类 / 排除项） | 零新依赖；随机源沿用 Tongsuo CSPRNG 与拒绝采样 | S–M | P0 |
| `jwt-codec` | 编码 | 非对称签名验证：现仅支持 `HS256` / `HS384` / `HS512`，补 `RS256` / `RS384` / `RS512` / `PS256` / `ES256` / `ES384`，并输出签名算法与密钥匹配结论；不重复实现验签 | 零新依赖；复用 `signature` 的 RSA / ECDSA 能力 | M | P0 |

> `password-gen` 与 `jwt-codec` 均为既有工具增强，**不新增 slug**、**不递增工具总数**；仅扩展请求字段与响应。两项已于 v0.8.0 交付，请求与响应字段见 [api/tools/crypto.md](api/tools/crypto.md#password-gen) 与 [api/tools/encoding.md](api/tools/encoding.md#jwt-codec)。

### 1.3 工具总数影响

批次 A 含 **7** 个新增 slug；批次 B 的 19 个备选全部并入本版交付，另补充 `mime-lookup` 与 `http-status` 两个速查类工具，合计 **28** 个新增 slug。工具总数由 **70** 变为 **98**（`password-gen` 与 `jwt-codec` 为既有增强，不计入）。

---

## 2. 批次 B —— 已并入 v0.8.0 交付

以下原为备选池，**已于 v0.8.0 全部交付**，slug 均在 catalog 登记。保留能力定义与取舍理由，供后续增强参考。

| slug | 分类 | 能力 | 与既有工具的关系 | 依赖 | 成本 | 优先级 |
| --- | --- | --- | --- | --- | --- | --- |
| `id-validator` | 数据 | 身份证 / 统一社会信用代码 / 手机号 / 银行卡 / 邮编 的校验位计算与规范化 | 无；Toolbit 目前无校验类工具 | 零新依赖 | S | P0 |
| `data-mask` | 文本 | 敏感数据脱敏：手机号、身份证、银行卡、姓名、邮箱 | 无；契合「无存储、隐私优先」定位 | 零新依赖 | S–M | P0 |
| `cert-lifecycle` | 证书 | 证书生命周期计算：有效期区间、到期倒计时、续期建议 | 扩充 `cert-parse` 的时间维度，不改其契约 | 零新依赖 | S | P1 |
| `password-policy` | 加密 | 口令策略合规检查：长度、字符类、历史、有效期 | 与 `password-gen` 的策略模板成对 | 零新依赖 | S | P1 |
| `totp` | 加密 | TOTP / HOTP 生成与校验 | [roadmap.md](roadmap.md) §16.1 已列；需 RFC 6238 / 4226 标准向量 | 复用 `hmac-gen` 的 HMAC 能力 | M | P1 |
| `csv-convert` | 数据 | JSON ↔ CSV 转换，CSV 格式化与合法性校验 | 补 `format-convert` 的 CSV 缺口（现覆盖 JSON / YAML / TOML / XML） | 零新依赖 | M | P1 |
| `json-to-types` | 数据 | JSON 转语言类型定义（TypeScript / Go / Rust / Java） | [roadmap.md](roadmap.md) §16.1 已列 | 零新依赖 | M | P1 |
| `docker-compose-gen` | 系统 | Dockerfile / compose 可视化生成，`docker run` 命令转 compose | 无 | 零新依赖 | M | P1 |
| `color-convert` | 格式化 | 颜色格式转换（HEX / RGB / HSL / OKLCH）与对比度检查 | [roadmap.md](roadmap.md) §16.1 已列 | 零新依赖 | S | P1 |
| `archive-inspect` | 数据 | ZIP / TAR / GZIP 内容查看与条目清单 | 扩充 `gzip-deflate`；**必须**限制条目数与解压体积 | 零新依赖 | M | P2 |
| `file-analyze` | 数据 | 文件类型识别与元数据提取（魔数、MIME、大小、摘要） | `mimetype` 已在依赖树，提升为直接依赖 | 零新依赖 | S | P2 |
| `text-stats` | 文本 | 字数 / 行数 / 字符统计，中英文分别计数 | 无 | 零新依赖 | S | P2 |
| `wifi-qr` | 数据 | 访客 WiFi 批量二维码生成 | 复用 `qr-code` 的生成能力；**已移除**（能力与 `qr-code` 重叠，且会明文回显含网络口令的载荷，改由 `qr-code` 承担二维码生成） | 零新依赖 | S | P2 |
| `ip-range-expand` | 网络 | IP 段与 CIDR 互转、批量展开 | 扩充 `cidr-calc` | 零新依赖 | S–M | P2 |
| `id-generator` | 数据 | UUID / ULID / 雪花 ID / NanoID 统一生成 | 扩充 `uuid-gen` 优于新增 slug | 零新依赖 | S | P2 |
| `mock-data` | 数据 | 按字段类型生成假数据 | 无 | 零新依赖 | M | P2 |
| `lookup-table` | 系统 | 速查表：端口号、Linux 退出码、时间戳格式 | [roadmap.md](roadmap.md) §16.1 已列 | 零新依赖 | S | P2 |
| `markdown-convert` | 文档 | Markdown 与 HTML 互转 | 复用既有 `goldmark` 依赖 | 零新依赖 | S | P2 |
| `html-escape` | 编码 | HTML 实体编解码 | 补 `html-format` 的实体方向 | 零新依赖 | S | P2 |

---

## 3. 实现边界与受阻项

候选中触及密码学能力缺口的项，一律按 [../AGENTS.md](../AGENTS.md) §15 处理：**上游未提供的能力跳过，不自研兜底**；缺口登记在 [upstream/tongsuo-go-issues.md](upstream/tongsuo-go-issues.md)。

| 事项 | 结论 | 依据 |
| --- | --- | --- |
| `ssh-keygen` 未加密私钥 | **可以实现** | `openssh-key-v1` 明文变体只做序列化；校验和为 SHA-256，须经 `cryptosvc` 取摘要，不自行实现 |
| `ssh-keygen` 口令加密私钥 | **受阻，必须走上游** | 该变体需要 bcrypt KDF 与密文格式转换；AGENTS §15 禁止自研密码学运算与密文格式转换 |
| `bcrypt` / `Argon2` / `scrypt` 口令哈希 | **受阻** | Tongsuo 无对应入口；改用第三方库等同自研兜底，需先向上游提 issue |
| ECIES / DES / 3DES | **受阻** | 已是缺口 I-02 与 I-05，上游补齐后再入契约 |
| `regex-build` | **可以** | 纯字符串拼装，不涉及密码学运算；产出结果按 RE2 方言校验 |
| `nginx-config` / `firewall-rules` | **可以** | 纯模板拼装，不出网、不涉及密码学 |
| `device-headers` 的 UA 库 | **需边界** | UA 数据需要更新机制；首版只支持内置常见平台与自定义填写，**不引入**在线 UA 查询 |
| 一切出网类需求 | **需单独评审** | SSRF 校验、限流与出网总开关要求见 [standards/06-security.md](standards/06-security.md) |
| 任何新增依赖 | **需说明理由** | 变更 `backend/go.mod` 必须在提交信息写明 `adds: <pkg> <version>`（AGENTS §12） |
| `hash-calc` 扩充摘要算法（SHA3 / BLAKE2 / CRC32 等） | **需先确认上游** | 现支持 SM3 / MD5 / SHA1 / SHA224 / SHA256 / SHA384 / SHA512；新增算法须经 `cryptosvc` 转调 Tongsuo，上游无入口则走上游 issue，**不得**用标准库 `crypto/*` 兜底 |

---

## 4. 落地检查清单

每项候选从「记录」转为「已交付」时，必须逐项完成：

1. 在 [`backend/catalog/tools.yaml`](../backend/catalog/tools.yaml) 登记 slug、分类、顺序、`transport`、`rateClass`、`maxBodyBytes`、`timeoutMs`、`sensitive`、`requiresNetwork`、`autoRun`、`wide`、`i18nKey`；
2. 后端实现：请求 / 响应结构体与 `Validate()`，错误统一走 `toolerr`；涉及密码学时经 `cryptosvc` / `pkisvc` 转调；
3. 前端实现：`frontend/src/tools/<category>/<slug>/{index.ts,Tool.vue,example.ts}`，并在 `frontend/src/tools/index.ts` 登记；优先复用共享组件；
4. 双语词条：`frontend/src/locales/zh-CN/tools/<slug>.json` 与 `frontend/src/locales/en-US/tools/<slug>.json` 字段齐平；
5. 契约测试：`backend/testdata/contract/<slug>.json`；
6. 文档同步：[tools.md](tools.md)、对应 [api/tools/](api/tools/) 分类文件、[roadmap.md](roadmap.md) §2 工具总数；
7. `make check-docs` 通过。

---

## 5. 相关文档

| 主题 | 文档 |
| --- | --- |
| 候选工具的来源与历史批次 | [roadmap.md](roadmap.md) §16.1 |
| 新增工具的完整流程 | [../AGENTS.md](../AGENTS.md) §8 |
| 已交付工具的产品规格 | [tools.md](tools.md) |
| 工具元数据唯一事实来源 | [`backend/catalog/tools.yaml`](../backend/catalog/tools.yaml) |
| 上游依赖缺口 | [upstream/tongsuo-go-issues.md](upstream/tongsuo-go-issues.md) |
| 密码学实现约束 | [standards/03-crypto.md](standards/03-crypto.md) |
| 文档写作约定 | [standards/09-docs.md](standards/09-docs.md) |
