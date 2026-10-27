# 文档 API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

覆盖 **3** 个文档类工具。端点均为 `POST /api/v1/tools/{slug}`。

| slug | 名称 | 传输 | 档位 |
| --- | --- | --- | --- |
| [pdf-sign-verify](#pdf-sign-verify) | PDF 签名校验 | multipart | file |
| [markdown-preview](#markdown-preview) | Markdown 预览 | JSON | light |
| [eml-parser](#eml-parser) | EML 邮件解析 | JSON / multipart | light |

---

## pdf-sign-verify

校验 PDF 文档中的数字签名（需出网与否取决于是否拉取信任根）。

### 请求（multipart）

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | file | 是 | PDF 文件（单文件） |
| `trustedRoots` | string | 否 | 受信任根证书 PEM（JSON 数组串） |
| `checkRevocation` | bool | 否 | 是否检查吊销状态（默认 `false`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `signatures` | array | `[{ field, subFilter, valid, reason, signer, chain[], time }]` |
| `fileInfo` | object | `{ name, size, pages?, version?, encrypted }` |

`signer` 结构：`{ subject, issuer, serial, notBefore, notAfter, algorithm }`；
`chain` 结构：`[{ subject, issuer, valid, reason }]`。

### 错误

`INVALID_FIELD`、`TOO_LARGE`、`INVALID_CERT`、`MISSING_FIELD`。

### 警告

`EXPIRED_CERT`、`NOT_YET_VALID`、`SELF_SIGNED_CERT`、`UNTRUSTED_CHAIN`。

> PDF 签名验证依赖底层库；**若**运行环境缺少该能力，端点**必须**返回 `NOT_IMPLEMENTED`（501）而**不是**静默返回空结果。

---

## markdown-preview

Markdown 渲染为 HTML，并生成目录。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `markdown` | string | 是 | Markdown 源文 |
| `flavor` | string | 否 | `gfm`（默认）\| `commonmark` |
| `allowHtml` | bool | 否 | 是否允许内嵌 HTML（默认 `false`） |
| `breaks` | bool | 否 | 单换行是否视为换行（默认 `false`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `html` | string | 渲染后的 HTML |
| `toc` | array | 目录 `[{ level, text, id }]` |

### 错误

`INVALID_FIELD`、`TOO_LARGE`。

### 安全

- `allowHtml: false`（默认）时，原始 HTML **必须**被转义；
- 输出 HTML **必须**经过 HTML 净化（禁止 `script`、事件属性、`javascript:` URL）；
- 前端渲染时**必须**在隔离容器 / 沙箱中插入，**禁止**直接 `v-html` 到主文档。

---

## eml-parser

解析 EML 邮件（RFC 5322）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 二选一 | `{ text, encoding }` |
| `file` | file | 二选一 | `.eml` 文件 |
| `maxAttachmentBytes` | int | 否 | 单附件最大字节数，默认 10 MB |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `headers` | array | 原始头字段 `[{ name, value }]` |
| `headerMap` | object | 头字段映射（小写键） |
| `bodies` | array | `[{ contentType, charset, text, html? }]` |
| `attachments` | array | `[{ filename, mime, size, rawBase64 }]` |

### 错误

`INVALID_FIELD`、`MISSING_FIELD`、`TOO_LARGE`。

> 附件**必须**按 `maxAttachmentBytes` 截断并在结果中标记；超限附件**不**返回内容，只返回元信息。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/06-security.md](../../standards/06-security.md) —— 安全规范（XSS 防护）
