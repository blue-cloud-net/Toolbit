# 编码 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **7** 个编码类工具。端点均为 `POST /api/v1/tools/{slug}`，全部为 `light` 档位。

| slug | 名称 |
| --- | --- |
| [base64](#base64) | Base64 编解码 |
| [url-codec](#url-codec) | URL 编解码 |
| [jwt-codec](#jwt-codec) | JWT 编解码 |
| [number-base](#number-base) | 进制转换 |
| [encoding-convert](#encoding-convert) | 编码转换 |
| [gzip-deflate](#gzip-deflate) | GZip/Deflate |
| [html-escape](#html-escape) | HTML 转义 |

---

## base64

Base64 编解码。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `encode` \| `decode` |
| `input` | object | 是 | `{ data, encoding }`；`encoding` ∈ `utf8` \| `hex` |
| `variant` | string | 否 | `standard`（默认）\| `url` |
| `padding` | bool | 否 | 是否保留 `=` 填充（默认 `true`） |
| `outputEncoding` | string | 否 | 默认 `utf8`（解码时） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |

### 错误

`INVALID_BASE64`、`INVALID_ENCODING`、`INVALID_FIELD`。

> 解码时忽略换行与空白；`variant: url` 使用 `-` / `_` 字符集。

---

## url-codec

URL 编解码与解析。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `encode` \| `decode` \| `parse` |
| `input` | object | 是 | `{ data, encoding }` |
| `component` | string | 否 | `full`（默认）\| `query` \| `path` \| `fragment` |
| `plusAsSpace` | bool | 否 | `+` 是否视为空格（默认 `true`，仅 `decode`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `parsed` | object | `parse` 时返回 `{ scheme, host, port, path, query, params }` |

### 错误

`INVALID_ENCODING`、`INVALID_FIELD`。

> `parsed.params` 为查询参数的键值对数组（同名参数保留多个条目）。

---

## jwt-codec

JWT 解码、编码与校验。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `decode` \| `encode` \| `verify` |
| `token` | string | `decode` / `verify` 时是 | JWT 字符串 |
| `header` | object | `encode` 时是 | JWT 头部 |
| `payload` | object | `encode` 时是 | JWT 载荷 |
| `secret` | string | `encode` / `verify`（HS 系列）时是 | 密钥（HMAC 系列） |
| `publicKey` | string | `verify`（RS / PS / ES 系列）时是 | 公钥（PEM，SPKI 或 PKCS#1） |
| `algorithm` | string | 否 | 生成时用于 `encode`，默认 `HS256`；见下方支持矩阵 |

**签名算法支持矩阵**

| 算法 | `decode` | `encode` | `verify` |
| --- | --- | --- | --- |
| `HS256` `HS384` `HS512` | ✅ | ✅ | ✅ |
| `RS256` `RS384` `RS512` | ✅ | ❌ | ✅ |
| `PS256` `PS384` `PS512` | ✅ | ❌ | ✅ |
| `ES256` | ✅ | ❌ | ✅ |
| `ES384` | ✅ | ❌ | ❌ |

> `ES384` 受上游限制暂不支持：`cryptosvc` 的 ECDSA 摘要固定为 SHA-256；请求中传入会返回 `UNSUPPORTED_ALGORITHM`。
> `verify` 用哪个算法由**令牌头部的 `alg`** 决定：HS 系列取 `secret`，RS / PS / ES 系列取 `publicKey`。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `header` | object | 头部 |
| `payload` | object | 载荷 |
| `signature` | string | 签名（base64url） |
| `token` | string | `encode` 时返回完整 token |
| `valid` | bool | `verify` 时返回 |
| `status` | object | 载荷含 `exp` / `nbf` 时返回 `{ expired, notYetValid }` |

### 错误

`INVALID_JWT`、`INVALID_FIELD`、`MISSING_FIELD`、`UNSUPPORTED_ALGORITHM`。

> 非对称验签**只做验签**，不复用 `signature` 之外的路径；缺公钥返回 `MISSING_FIELD`。
> 解码 `exp` / `nbf` 时返回**提示性** `expired` 标记，**不**作为错误。

---

## number-base

进制转换（2–36）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `value` | string | 是 | 待转换的值（字符串，支持大整数） |
| `from` | int | 是 | 源进制（2–36） |
| `to` | int | 是 | 目标进制（2–36） |
| `prefix` | bool | 否 | 是否加进制前缀（`0x` / `0o` / `0b`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `result` | string | 转换结果 |
| `from` / `to` | int | 进制 |

### 错误

`INVALID_FIELD`。

> 使用任意精度整数，**不**受 64 位限制。

---

## encoding-convert

字符编码转换（UTF-8 / UTF-16 / GBK / Big5 / Latin-1 / Shift_JIS 等）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data, encoding }`；`encoding` 是**传输编码**（`utf8` \| `hex` \| `base64` \| `base64url`），不是字符集 |
| `from` | string | 是 | 源字符集；`auto` 表示自动探测 |
| `to` | string | 是 | 目标字符集（不支持 `auto`） |
| `bom` | bool | 否 | 是否写入 BOM（默认否） |
| `outputEncoding` | string | 否 | 输出传输编码，默认 `base64` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 转换结果（按 `outputEncoding` 编码的文本） |
| `encoding` | string | 实际输出编码 |
| `detected` | string | `from: auto` 时的探测结果 |

### 错误

`INVALID_ENCODING`、`INVALID_FIELD`。

> 非法字节序列**必须**替换为 U+FFFD 并在 `meta.warnings` 中提示，**不**直接失败。

---

## gzip-deflate

压缩与解压。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `compress` \| `decompress` |
| `format` | string | 否 | `gzip`（默认）\| `deflate`（zlib）\| `raw`（无头） |
| `input` | object | 是 | `{ data, encoding }` |
| `level` | int | 否 | 压缩级别（0–9） |
| `outputEncoding` | string | 否 | 默认 `base64` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 压缩 / 解压结果 |
| `encoding` | string | 实际使用的输出编码 |
| `ratio` | number | 压缩比 |

### 错误

`INVALID_FIELD`、`TOO_LARGE`、`DECRYPT_FAILED`（解压失败统一映射）。

> 解压时对输出有上限保护（默认 400 MB），超出返回 `TOO_LARGE`，防止**解压炸弹**。

---

## html-escape

HTML 实体编解码（HTML Escape）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 否 | `encode` 转义 / `decode` 反转义，默认 `encode` |
| `input` | string | 是 | 待处理文本 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 模式 |
| `output` | string | 转义 / 反转义结果 |
| `changedCount` | int | 发生变化的字符数 |

### 错误

`INVALID_FIELD`。

> 只处理命名实体与十进制 / 十六进制数字实体，不做 CSS / JS 上下文转义。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/01-frontend.md](../../standards/01-frontend.md) —— 前端规范
