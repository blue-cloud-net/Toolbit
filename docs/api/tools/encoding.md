# 编码 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-11-02

覆盖 **9** 个编码类工具。端点均为 `POST /api/v1/tools/{slug}`，全部为 `light` 档位。

| slug | 名称 |
| --- | --- |
| [base64](#base64) | Base64 编解码 |
| [url-codec](#url-codec) | URL 编解码 |
| [jwt-codec](#jwt-codec) | JWT 编解码 |
| [number-base](#number-base) | 进制转换 |
| [encoding-convert](#encoding-convert) | 编码转换 |
| [gzip-deflate](#gzip-deflate) | GZip/Deflate |
| [html-escape](#html-escape) | HTML 转义 |
| [protobuf-decoder](#protobuf-decoder) | Protobuf 解码器 |
| [protobuf-encoder](#protobuf-encoder) | Protobuf 编码器 |

---

## base64

Base64 编解码，以及标准变体与 URL 安全变体之间的互转。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `encode` \| `decode` \| `convert` |
| `input` | object | 是 | `{ data, encoding }`；`encoding` ∈ `utf8` \| `hex`（`convert` 时仅 `utf8`） |
| `variant` | string | 否 | `standard`（默认）\| `url`；`convert` 时可取 `auto` |
| `padding` | bool | 否 | 是否保留 `=` 填充（默认 `true`） |
| `outputEncoding` | string | 否 | 默认 `utf8`（解码时） |
| `targetVariant` | string | 否 | `convert` 的输出变体：`standard` \| `url`（默认 `url`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `encoding` | string | 输出编码（`base64` \| `base64url` \| `utf8` \| `hex`） |

### 错误

`INVALID_BASE64`、`INVALID_ENCODING`、`INVALID_FIELD`。

> 解码时忽略换行与空白；`variant: url` 使用 `-` / `_` 字符集。
>
> `convert` 在字节层面完成：先按 `variant` 解出字节，再按 `targetVariant` 重新编码。
> `variant: auto` 按输入用到的字符判断来源变体；两组字母表混用（既含 `+`/`/` 又含 `-`/`_`）返回 `INVALID_BASE64`。

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

## protobuf-decoder

解析 protobuf 二进制数据。支持两种模式：`wire`（按线格式解析，不需要描述文件）与
`schema`（按 `.proto` 描述解释字段名与声明类型）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 否 | `wire`（默认）\| `schema` |
| `input` | object | 是 | `{ data, encoding }`；`encoding` ∈ `utf8` \| `hex` \| `base64` \| `base64url` |
| `schema` | string | `schema` 模式必填 | `.proto` 描述文本；只支持单文件 |
| `messageType` | string | 否 | 顶层消息全名或简名；`schema` 中顶层消息唯一时可省略 |
| `maxDepth` | int | 否 | 嵌套解析深度上限（0–128），默认 `32` |
| `maxFields` | int | 否 | 字段总数上限（1–100000），默认 `10000` |
| `inferStrings` | bool | 否 | 是否**优先**把长度分隔字段识别为 UTF-8 文本，默认 `true` |
| `includeHexDump` | bool | 否 | 是否返回十六进制转储，默认 `false` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 实际使用的解析模式 |
| `messageType` | string | 解析所用的消息名；`wire` 模式为空 |
| `byteLength` | int | 输入字节数 |
| `fieldCount` | int | 解析出的字段总数（含嵌套字段） |
| `depth` | int | 实际达到的最大嵌套深度 |
| `maxDepth` | int | 生效的深度上限 |
| `fields` | array | 顶层字段列表 |
| `json` | string | 解析结果的 JSON 文本 |
| `unknownFieldCount` | int | `schema` 模式下无法识别的字段数 |
| `hexDump` | string | 十六进制转储；未请求时为空 |
| `truncated` | bool | 结果是否因上限被截断 |

`fields[]` 的元素结构：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `path` | string | 字段路径：根为 `root`，子字段逐级追加 `/<字段号>`，重复字段的元素追加 `[<下标>]` |
| `offset` | int | 该字段在线格式字节流中的起始偏移 |
| `fieldNumber` | int | 字段号 |
| `name` | string | 字段名；仅 `schema` 模式有值 |
| `wireType` | string | 线类型：`varint` \| `fixed32` \| `fixed64` \| `bytes` \| `group` |
| `declaredType` | string | 声明类型；仅 `schema` 模式有值 |
| `value` | string | 字段值的文本表示 |
| `length` | int | 长度分隔字段的字节数 |
| `text` | bool | 该值是否被识别为 UTF-8 文本 |
| `message` | bool | 该值是否被识别为嵌套消息 |
| `hex` | string | 二进制值的十六进制预览（最多 64 字节） |
| `float` | number | 32 / 64 位浮点字段的浮点解释 |
| `children` | array | 嵌套消息的子字段，结构同上 |

### 错误

`INVALID_PROTOBUF`、`MISSING_FIELD`、`INVALID_FIELD`、`UNSUPPORTED_MODE`、
`INVALID_ENCODING`、`INVALID_HEX`、`INVALID_BASE64`。

> `wire` 模式只依据字节结构判断语义：长度分隔字段先看是不是可打印的 UTF-8 文本，
> 不是才尝试按嵌套消息解析，两者都不成立则给出十六进制预览。
>
> `schema` 模式支持 Google 标准类型（如 `google/protobuf/timestamp.proto`）；
> 其它 import 需要先把定义合并进同一份描述文件。
>
> 超出 `maxFields` / `maxDepth` 时结果被**截断**并置 `truncated: true`，不报错。

---

## protobuf-encoder

把字段规格或 protojson 编码成 protobuf 二进制数据。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 否 | `wire`（默认）\| `schema` |
| `schema` | string | `schema` 模式必填 | `.proto` 描述文本；只支持单文件 |
| `messageType` | string | 否 | 顶层消息全名或简名 |
| `message` | string | `schema` 模式必填 | protojson 文本 |
| `fields` | array | `wire` 模式必填 | 字段规格列表 |
| `outputEncoding` | string | 否 | 默认 `base64` |

`fields[]` 的元素结构：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `field` | int | 是 | 字段号（1–536870911） |
| `type` | string | 是 | 语义类型名，见下表 |
| `value` | string | 是 | 值的文本表示；**一律用字符串**，避免 64 位整数丢精度 |
| `encoding` | string | 否 | `bytes` 类型值的编码（`utf8` \| `hex` \| `base64` \| `base64url`） |
| `fields` | array | 否 | `message` 类型的子字段规格 |

`type` 的取值与对应的线类型：

| `type` | 线类型 | `value` 的含义 |
| --- | --- | --- |
| `int32` / `int64` | varint | 十进制有符号整数 |
| `uint32` / `uint64` | varint | 十进制无符号整数（也支持 `0x` 前缀的十六进制） |
| `sint32` / `sint64` | varint | 十进制有符号整数，按 ZigZag 编码 |
| `bool` | varint | `true` / `false` |
| `enum` | varint | 十进制有符号整数（无线格式时只支持数值） |
| `fixed32` / `sfixed32` | fixed32 | 32 位整数 |
| `float` | fixed32 | 32 位浮点数 |
| `fixed64` / `sfixed64` | fixed64 | 64 位整数 |
| `double` | fixed64 | 64 位浮点数 |
| `string` | bytes | UTF-8 文本 |
| `bytes` | bytes | 按 `encoding` 解释的字节 |
| `message` | bytes | 由子字段编码出的嵌套消息 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 编码结果 |
| `encoding` | string | 实际使用的输出编码 |
| `byteLength` | int | 结果字节数 |
| `fieldCount` | int | 写入的字段总数（含嵌套字段） |

### 错误

`INVALID_PROTOBUF`、`MISSING_FIELD`、`INVALID_FIELD`、`UNSUPPORTED_MODE`、
`INVALID_ENCODING`、`INVALID_HEX`、`INVALID_BASE64`。

> `wire` 模式按字段在列表中的**先后顺序**依次写出，不做字段号排序；
> 同一个字段号出现多次即为 protobuf 的重复字段语义。
>
> `schema` 模式按 protojson 规范解析 `message`：64 位整数用字符串表示，
> `bytes` 用 Base64 表示，枚举可以用枚举名；编码结果按**字段号**排序，输出稳定可复现。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/01-frontend.md](../../standards/01-frontend.md) —— 前端规范
