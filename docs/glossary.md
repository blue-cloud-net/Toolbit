# 术语表

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

本文统一本项目文档与代码中的术语。

---

## 1. 通用

| 术语 | 英文 / 标识 | 含义 |
| --- | --- | --- |
| 工具 | tool | 一个可独立调用的功能单元，标识为 `slug` |
| 目录 | catalog | `catalog/tools.yaml`，工具元数据的唯一事实来源 |
| 标识 | slug | 工具的小写 kebab-case 唯一标识，如 `sm4-file-crypt` |
| 分类 | category | 工具的归属分组，共 10 个 |
| 档位 | rateClass | 限流档位：`light` / `standard` / `heavy` / `file` / `network` |
| 传输 | transport | 交互方式：`json` / `multipart` / `websocket` / `client` |
| 信封 | envelope | 统一的 JSON 响应结构（`code` / `message` / `data` / `error` / `meta` / `requestId`） |
| 错误码 | error code | 机器可读的错误标识，如 `INVALID_FIELD` |
| 警告 | warning | 不阻断执行、仅提示的附加信息，如 `INSECURE_ECB` |
| 契约测试 | contract test | 校验「请求 → 响应」固定形状的测试 |
| 黄金样本 | golden fixture | 固定的期望输入 / 输出数据 |

---

## 2. 密码学

| 术语 | 英文 / 标识 | 含义 |
| --- | --- | --- |
| 国密 | — | 中国商用密码算法（SM2 / SM3 / SM4 / SM9 / ZUC） |
| SM2 | — | 非对称算法：密钥交换、签名、加密 |
| SM3 | — | 杂凑（摘要）算法，输出 256 bit |
| SM4 | — | 分组对称算法，分组 128 bit，密钥 128 bit |
| 摘要 | digest / hash | 任意长度数据 → 定长比特串 |
| 杂凑 | — | 「摘要」的同义表述（国标用词） |
| 签名 | signature | 用私钥对消息生成的认证数据 |
| 验签 | verify signature | 校验签名是否有效 |
| userId | `sm2UserId` | SM2 签名时参与运算的标识串，默认 `1234567812345678` |
| 密文顺序 | C1C3C2 / C1C2C3 | SM2 密文中 C1 / C2 / C3 的排列顺序 |
| 裸格式 | raw | 不带 ASN.1 包装的 SM2 密文，C1 以 `04` 开头 |
| DER | — | ASN.1  Distinguished Encoding Rules，二进制编码 |
| 认证标签 | tag | GCM 等 AEAD 模式的完整性校验值 |
| 附加认证数据 | AAD | AEAD 中不加密但参与认证的数据 |
| 填充 | padding | 分组密码中不足一整组的补齐方式 |
| PBES | — | 基于口令的加密方案（口令 → 密钥） |
| KDF | — | 密钥派生函数 |
| 密钥对 | key pair | 公钥 + 私钥 |

---

## 3. X.509 / PKI

| 术语 | 英文 / 标识 | 含义 |
| --- | --- | --- |
| 证书 | certificate | 由 CA 签发的公钥凭证 |
| 证书签名请求 | CSR | 向 CA 申请证书的请求文件 |
| 吊销列表 | CRL | 列出已吊销证书的签名列表 |
| 在线证书状态 | OCSP | 单证书吊销状态查询协议 |
| 证书链 | certificate chain | 叶证书 → 中间 CA → 根 CA |
| 根证书 | root certificate | 自签、作为信任锚点的证书 |
| 信任锚 | trust anchor | 验证链的起点（受信任根） |
| 主体 | subject | 证书描述的持有者 |
| 签发者 | issuer | 签发该证书的 CA |
| 序列号 | serial number | 证书的唯一编号 |
| 主体备用名 | SAN | 证书中的域名 / IP 等名称列表 |
| 密钥用途 | key usage | 证书允许的密钥用途（签名、加密…） |
| 扩展密钥用途 | EKU | 证书允许的 TLS 等用途 |
| 基本约束 | basic constraints | 是否为 CA 及路径长度限制 |
| 证书透明度 | CT | 公开记录证书签发的日志体系 |
| CAA | — | 限制哪些 CA 可为某域名签发证书的 DNS 记录 |
| PKCS#7 | — | 证书 / 签名容器格式（`.p7b`） |
| PKCS#10 | — | CSR 的格式 |
| PKCS#12 | — | 私钥 + 证书容器格式（`.pfx` / `.p12`） |
| PKCS#8 | — | 通用私钥格式 |
| SPKI | — | 公钥格式（SubjectPublicKeyInfo） |
| 吊销原因码 | reason code | CRL 中说明吊销原因的编号（RFC 5280 §5.3.1） |

---

## 4. 后端

| 术语 | 英文 / 标识 | 含义 |
| --- | --- | --- |
| 分发器 | dispatcher | 把 `/api/v1/tools/{slug}` 路由到对应 handler 的组件 |
| 会话 | session | 一次请求内的密码学句柄与资源容器（`cryptosvc.Session`） |
| 句柄 | handle | Tongsuo 对象的原生资源，需显式释放 |
| 服务 | service | `internal/cryptosvc` 等不含 HTTP 语义的业务层 |
| 限额 | limit | 单次请求允许的最大体积、条目数、时长等 |
| 限流 | rate limit | 按 IP 限制请求速率 |
| 幂等 | idempotent | 重复执行结果相同（如 `Close()`） |
| 信封 | envelope | 见 §1 |
| 请求 ID | requestId | 贯穿日志与响应的请求标识 |

---

## 5. 前端

| 术语 | 英文 / 标识 | 含义 |
| --- | --- | --- |
| 组合式函数 | composable | `use*` 前缀的逻辑复用单元 |
| 共享组件 | shared component | 跨工具复用的 UI 组件 |
| 工具组件 | tool component | 单个工具的页面组件 |
| 分栏 | wide | ≥1280px 下左右分栏的布局标记 |
| 预设 | preset | 固定选项集合（如主色、字符集） |
| 主色 | accent | 界面强调色，共 8 组预设 |
| 收藏 | favorites | 存在 `localStorage` 的工具 slug 列表 |
| 最近使用 | recent | 存在 `localStorage` 的最近工具 slug 列表 |

---

## 6. 部署与运维

| 术语 | 英文 / 标识 | 含义 |
| --- | --- | --- |
| 镜像 | image | 容器镜像 |
| 数据卷 | volume | 持久化存储（统计库） |
| 反向代理 | reverse proxy | 终止 TLS 并转发到本服务的组件 |
| 探针 | probe | `/healthz`（存活）与 `/readyz`（就绪） |
| 优雅退出 | graceful shutdown | 停止收新请求、等待进行中请求后退出 |
| SSRF | — | 服务端请求伪造 |
| 私网段 | private range | `10/8`、`172.16/12`、`192.168/16` 等内网地址范围 |
| 归一化 | normalization | 把不同形态的输入统一到标准形态（如 PEM 规范化） |
| 指纹 | fingerprint | 公钥或证书的哈希摘要 |

---

## 7. 相关文档

- [architecture.md](architecture.md) —— 架构与关键决策
- [api/README.md](api/README.md) —— API 总览
- [api/errors.md](api/errors.md) —— 错误码全集
- [i18n.md](i18n.md) —— 国际化规范
