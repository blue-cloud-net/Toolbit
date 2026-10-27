# 路线图

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

本文记录 Toolbit 的开发计划与版本目标。

---

## 1. 版本策略

采用语义化版本（SemVer）：

| 版本位 | 含义 |
| --- | --- |
| MAJOR | 破坏性变更（API 不兼容、镜像行为变化） |
| MINOR | 新功能（新工具、新端点） |
| PATCH | 缺陷修复与文档修订 |

---

## 2. 阶段总览

| 阶段 | 目标 | 产出 |
| --- | --- | --- |
| **P0** 文档与契约 | 冻结设计 | 全部规范、API 契约、catalog |
| **P1** 基础设施 | 可运行骨架 | Go 工程、catalog 加载、分发器、前端脚手架 |
| **P2** 密码学核心 | 国密能力 | `cryptosvc`、签名 / 加解密 / 摘要 / HMAC / 密钥 |
| **P3** PKI 全链路 | 证书能力 | 解析 / 签发 / 链验证 / CRL / OCSP / PKCS#12 / #7 |
| **P4** 通用工具 | 编码 / 格式化 / 文本 / 数据 / 系统 | 26 个工具 |
| **P5** 网络与文档 | 出网与文档处理 | 5 个网络工具 + 3 个文档工具 |
| **P6** 打磨与发布 | 质量与交付 | 视觉打磨、E2E、镜像发布 |

---

## 3. P0 —— 文档与契约（已完成）

| 交付物 | 位置 |
| --- | --- |
| 项目说明 | [../README.md](../README.md) |
| 架构与关键决策 | [architecture.md](architecture.md) |
| 规范（11 篇） | [standards/](standards/) |
| 工具目录 | [../catalog/tools.yaml](../catalog/tools.yaml) |
| API 契约 | [api/](api/) |
| 工具产品规格 | [tools.md](tools.md) |
| 部署 / 隐私 / 国际化 / 术语 | [deploy.md](deploy.md) [privacy.md](privacy.md) [i18n.md](i18n.md) [glossary.md](glossary.md) |

**完成标准**：目录、错误码、配置项、契约四者互相一致，且由脚本校验。

---

## 4. P1 —— 基础设施

| 任务 | 说明 |
| --- | --- |
| Go 工程骨架 | `cmd/toolbit`、`internal/*` 分层 |
| catalog 加载 | YAML 解析 + 校验 + 嵌入二进制 |
| 工具注册表 | slug → handler 映射，与 catalog 交叉校验 |
| 分发器 | 路由、Content-Type 分支、Body 限制、超时 |
| 错误框架 | `internal/toolerr` 注册表 + i18n key |
| 限流中间件 | 按 `rateClass` 实施 |
| 跨站校验 | `Origin` / `Sec-Fetch-Site` |
| 静态资源服务 | SPA 回退 + 缓存头 |
| 前端脚手架 | Vite + Vue 3 + TS + Element Plus + Pinia + i18n + UnoCSS |
| 首页与分类页 | 目录渲染、搜索 |

**完成标准**：能启动、能显示 70 个工具的目录、`/readyz` 通过。

---

## 5. P2 —— 密码学核心

| 工具 | 备注 |
| --- | --- |
| `hash-calc` | SM3 / SHA 系列 / MD5 |
| `hmac-gen` | 含 SM3 |
| `password-gen` | CSPRNG |
| `signature` | SM2 / RSA / ECDSA / Ed25519 / Ed448 |
| `symmetric-crypt` | SM4 / AES，含 GCM |
| `asymmetric-crypt` | SM2（`C1C3C2` 默认）/ RSA / ECIES |
| `ecdh` | P-256/384/521、X25519/X448 |
| `sm4-file-crypt` | TBFC 容器，分块流式 |
| `keygen` / `key-convert` / `key-parse` / `key-ops` / `key-match` / `jwk-xml-convert` | 密钥类 6 个 |

**完成标准**：国密标准向量全部通过；`Session` 生命周期无泄漏；`-race` 通过。

---

## 6. P3 —— PKI 全链路

| 工具 | 备注 |
| --- | --- |
| `cert-parse` / `csr-parse` / `pem-der` / `asn1-view` | 解析类 |
| `cert-chain-parse` / `cert-chain-verify` / `cert-chain-repair` | 链类 |
| `csr-gen` / `cert-gen` | 生成类 |
| `crl-parse` / `crl-gen` / `revocation-check` | CRL 类 |
| `ocsp-request` / `ocsp-response` | OCSP 类 |
| `pfx-pack` / `pfx-parse` / `pfx-change-password` | PKCS#12 |
| `p7b-pack` / `p7b-extract` | PKCS#7 |
| `caa-gen` | CAA |

**完成标准**：自签 CA → 叶证书 → 链验证 → CRL → 吊销检查全链路可用。

---

## 7. P4 —— 通用工具

| 分类 | 工具 |
| --- | --- |
| 编码 | `base64` `url-codec` `jwt-codec` `number-base` `encoding-convert` `gzip-deflate` |
| 格式化 | `json-format` `js-format` `sql-format` `html-format` `xml-format` `toml-format` `yaml-format` `format-convert` |
| 文本 | `regex-tester` `text-convert` |
| 数据 | `timestamp` `uuid-gen` `date-calc` `country-code` `cn-regions` `qr-code` |
| 系统 | `curl-generator` `crontab-generator` `chmod-calc` |

**完成标准**：26 个工具全部有契约测试与 i18n 词条。

---

## 8. P5 —— 网络与文档

| 工具 | 备注 |
| --- | --- |
| `cidr-calc` | 纯计算 |
| `ssl-check` | WebSocket 流式 |
| `dns-lookup` | UDP / TCP / DoT / DoH |
| `dns-tester` | whois / DNSSEC / PTR |
| `websocket-test` | 纯前端 |
| `cert-chain-fetch` / `ct-search` / `domain-verify` | 出网 |
| `pdf-sign-verify` | 依赖底层库，**不可用时降级为 501** |
| `markdown-preview` | 需 XSS 净化 |
| `eml-parser` | MIME 解析 |

**完成标准**：出网工具全部通过 SSRF 校验测试；`pdf-sign-verify` 的降级路径已冻结。

---

## 9. P6 —— 打磨与发布

| 任务 | 说明 |
| --- | --- |
| 视觉打磨 | 8 组主色预设、深浅色、响应式 |
| E2E | Playwright 覆盖 70 个工具路由 |
| 可访问性 | axe 扫描无 serious / critical 问题 |
| 性能预算 | 初始 JS < 250 KB（gzip）；单工具 chunk < 40 KB |
| 容器 | multi-stage、非 root、digest 固定、Trivy 扫描 |
| CI | lint / test / scan / build / docs / publish 全链路 |
| 发布 | v0.1.0 镜像与 tag |

---

## 10. v1.1 展望

| 项 | 说明 |
| --- | --- |
| JSON Schema | `GET /api/v1/schema/{slug}` 自动生成 |
| OpenAPI | 由反射生成 OpenAPI 3.1 文档 |
| 工具收藏同步 | 可选的加密导出 / 导入 |
| 更多语言 | 繁体中文、日语 |
| 批量工具 | 多文件 / 多条输入的并行处理 |
| SM9 | 支持基于标识的密码算法 |
| 工具链 | 证书链自动下载（需显式授权） |

---

## 11. 明确不做

| 项 | 原因 |
| --- | --- |
| 用户账号体系 | 违背「无账号、隐私优先」的设计 |
| 服务端存储用户数据 | 会引入数据泄露风险 |
| 内置 TLS | 交由反向代理，减少实现复杂度与配置面 |
| 自研密码算法 | 安全风险；一律使用 Tongsuo |
| 移动端原生 App | Web 覆盖足够，且维护成本高 |
| 插件系统 | v1 工具集合固定，避免不可信扩展 |

---

## 12. 风险与对策

| 风险 | 影响 | 对策 |
| --- | --- | --- |
| `tongsuo-go` 版本漂移 | API 不兼容 | CI 用 `go doc` 快照比对；锁 `go.sum` |
| PDF 签名库不可用 | `pdf-sign-verify` 缺失 | 降级返回 501；契约冻结 |
| 容器内 cgo 兼容性 | 镜像无法启动 | 固定基础镜像 digest；`healthcheck` 兜底 |
| 70 个工具范围过大 | 交付延期 | 分阶段；每阶段有独立完成标准 |
| 出网工具被滥用 | 资源消耗 / 被用于攻击 | SSRF 校验 + 限流 + 出网总开关 |
| 单进程吞吐不足 | 高并发下体验差 | 重操作信号量；后续可横向扩展（统计库为共享瓶颈） |

---

## 13. 相关文档

- [architecture.md](architecture.md) —— 架构与关键决策
- [standards/09-docs.md](standards/09-docs.md) —— 文档规范与同步矩阵
- [../AGENTS.md](../AGENTS.md) —— 协作约定
- [../CHANGELOG.md](../CHANGELOG.md) —— 变更记录
