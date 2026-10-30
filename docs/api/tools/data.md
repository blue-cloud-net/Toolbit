# 数据 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **13** 个数据类工具。端点均为 `POST /api/v1/tools/{slug}`。

| slug | 名称 | 档位 |
| --- | --- | --- |
| [timestamp](#timestamp) | 时间戳转换 | light |
| [uuid-gen](#uuid-gen) | UUID 生成 | light |
| [date-calc](#date-calc) | 日期计算 | light |
| [country-code](#country-code) | 国家代码 | light |
| [cn-regions](#cn-regions) | 中国行政区划 | light |
| [qr-code](#qr-code) | 二维码生成 | light |
| [id-validator](#id-validator) | 号码校验 | light |
| [csv-convert](#csv-convert) | CSV 转换 | light |
| [json-to-types](#json-to-types) | JSON 转类型 | light |
| [mock-data](#mock-data) | 模拟数据 | light |
| [archive-inspect](#archive-inspect) | 归档查看 | light |
| [file-analyze](#file-analyze) | 文件识别 | light |
| [id-generator](#id-generator) | ID 生成 | light |

---

## timestamp

时间戳与日期互转。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `value` | string \| number | 否 | 待转换值；缺省取当前时间 |
| `unit` | string | 否 | `s`（秒）\| `ms`（毫秒）\| `us` \| `ns` |
| `timezone` | string | 否 | IANA 时区名，如 `Asia/Shanghai`；缺省用 UTC |
| `operation` | string | 否 | `toDate`（默认）\| `toTimestamp` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `timestamp` | object | `{ s, ms }` |
| `iso8601` | string | ISO 8601 格式 |
| `rfc3339` | string | RFC 3339 格式 |
| `local` | string | 指定时区的本地时间 |
| `utc` | string | UTC 时间 |

### 错误

`INVALID_TIME`、`INVALID_FIELD`。

### 警告

`TIMEZONE_ASSUMED` —— 未指定时区时按 UTC 处理。

> Unix 时间戳自动按位数识别秒 / 毫秒（10 位 / 13 位）。

---

## uuid-gen

UUID 生成。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `version` | int | 否 | `4`（默认）\| `1` \| `7` |
| `count` | int | 否 | 生成数量，默认 1（≤ 1000） |
| `uppercase` | bool | 否 | 是否大写 |
| `hyphens` | bool | 否 | 是否带连字符（默认 `true`） |
| `braces` | bool | 否 | 是否带花括号 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `uuids` | array | 生成的 UUID 列表 |
| `count` | number | 生成条数 |

### 错误

`INVALID_FIELD`、`TOO_MANY_ITEMS`。

> UUID v1 与 v7 **必须**使用单调时钟保证同毫秒内不重复。

---

## date-calc

日期计算。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `diff` \| `add` \| `sub` \| `workdays` |
| `dateA` | string | 是 | 日期（RFC3339 或 `YYYY-MM-DD`） |
| `dateB` | string | `diff` 时是 | 第二个日期 |
| `days` | int | `add` / `sub` 时是 | 增减天数 |
| `unit` | string | 否 | `day`（默认）\| `week` \| `month` \| `year` |
| `holidays` | array | 否 | 节假日（`YYYY-MM-DD`），`workdays` 时排除 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `resultDate` | string | 计算结果（`add` / `sub`） |
| `diffDays` | number | 相差天数（`diff`） |
| `workdays` | number | 工作日数（`workdays`） |

### 错误

`INVALID_TIME`、`INVALID_FIELD`、`MISSING_FIELD`。

> `diff` 按自然日计算（忽略时分秒），端点**不**判断节假日（`holidays` 仅在 `workdays` 中生效）。

---

## country-code

国家与地区代码查询。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `query` | string | 否 | 名称或代码关键词；缺省返回全部 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `countries` | array | `[{ alpha2, alpha3, numeric, nameZh, nameEn }]` |
| `total` | number | 匹配条数 |
| `dataVersion` | string | 数据集版本 |

### 错误

`INVALID_FIELD`。

> 数据源为 ISO 3166-1（离线内置，**不**出网）。
>
> 缺省请求返回**全量**国家与地区（249 条），供前端一次性取回后在本地筛选。

---

## cn-regions

中国行政区划查询。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `query` | string | 否 | 名称或编码关键词 |
| `province` | string | 否 | 指定省级行政区，缩小城市查询范围 |
| `level` | string | 否 | 层级过滤：`province` \| `city` \| `district` \| `all`；为空表示缺省只返回省级 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `provinces` | array | 省级行政区 |
| `cities` | array | 地级 / 县级行政区 |
| `total` | number | 匹配条数 |
| `dataVersion` | string | 数据集版本 |

### 错误

`INVALID_FIELD`。

> 数据为**离线内置快照**，含行政区划代码（6 位）。数据版本**必须**在响应元信息中标注。
>
> `level=all` 返回全部层级（约 3400 条），供前端一次性取回后在本地筛选；
> `level` 与 `query` / `province` 可叠加。

---

## qr-code

二维码生成与识别。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `generate` \| `decode` |
| `text` | string | `generate` 时是 | 二维码内容 |
| `options` | object | 否 | `{ size, margin, ecLevel, fg, bg }` |
| `imageBase64` | string | `decode` 时是 | 待识别的二维码图片（base64 的 PNG / JPEG） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `svg` | string | 矢量图（`generate`） |
| `pngBase64` | string | 位图（`generate`） |
| `dimension` | number | 模块边长，含静默区（`generate`） |
| `moduleSize` | number | 模块的像素边长（`generate`） |
| `content` | string | 识别内容（`decode`） |
| `format` | string | 图片格式（`decode`） |

### 错误

`INVALID_FIELD`、`MISSING_FIELD`、`TOO_LARGE`。

> `ecLevel` ∈ `L` `M` `Q` `H`（默认 `M`）。生成内容长度上限 4 KB。

---

## id-validator

身份证、信用代码、手机号、银行卡校验位计算（ID Validator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `type` | string | 是 | `idCard` `uscc` `phone` `bankCard` `postcode` |
| `value` | string | 是 | 待校验的号码 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `type` | string | 校验类型 |
| `value` | string | 规范化后的值 |
| `valid` | bool | 是否通过校验 |
| `normalized` | string | 规范化结果 |
| `checkDigit` | string | 校验位 |
| `region` | string | 归属地区 |
| `birthday` | string | 出生日期 |
| `gender` | string | 性别 |
| `issue` | string | 失败原因；通过时为空串 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 只做**校验位计算与格式规范化**，不查询任何外部数据库。

---

## csv-convert

CSV 与 JSON 互转（CSV Converter）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 否 | `csvToJson` / `jsonToCsv`，默认 `csvToJson` |
| `input` | string | 是 | 源数据文本 |
| `delimiter` | string | 否 | CSV 分隔符，单字符，默认 `,` |
| `hasHeader` | bool | 否 | 首行是否为表头，默认 `false` |
| `pretty` | bool | 否 | 输出是否缩进，默认 `false` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `operation` | string | 实际使用的转换方向 |
| `output` | string | 转换结果 |
| `rowCount` | int | 数据行数 |
| `columnCount` | int | 列数 |
| `columns` | array | 列名列表 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> CSV 解析允许宽松引号（`LazyQuotes`）；无表头时列名按 `column1`、`column2` 生成。

---

## json-to-types

由 JSON 样本生成语言类型定义（JSON to Types）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | string | 是 | JSON 样本 |
| `language` | string | 否 | `typescript` `go` `rust` `java`，默认 `typescript` |
| `rootName` | string | 否 | 根类型名，默认按结构推断 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `language` | string | 目标语言 |
| `output` | string | 生成的类型定义 |
| `typeCount` | int | 类型数量 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 类型由样本**推断**得出，不保证与真实结构一致。

---

## mock-data

按字段类型生成假数据（Mock Data）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `types` | array | 否 | 字段类型列表，见下表；留空表示全部类型 |
| `count` | int | 否 | 每种类型的生成条数，默认 3 |
| `locale` | string | 否 | `zh-CN` / `en-US`，默认 `zh-CN` |
| `prefix` | string | 否 | 生成值前缀 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `types` | array | 生成的类型列表 |
| `fields` | array | 生成的字段值 `[{ type, value }]` |
| `count` | int | 每种类型的条数 |
| `locale` | string | 语言 |
| `seed` | string | 本次随机种子 |

### 错误

`INVALID_FIELD`。

> 使用**非密码学**随机源，仅用于演示与占位；**不可**用作真实凭据。

---

## archive-inspect

查看 ZIP / TAR / GZIP 的条目清单（Archive Inspector）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `data` | string | 是 | 归档内容的 Base64 |
| `format` | string | 否 | `zip` `tar` `tarGz` `gzip`；留空则自动识别 |
| `includeContent` | bool | 否 | 是否给出条目内容预览，默认 `false` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `format` | string | 归档格式 |
| `entryCount` | int | 条目数 |
| `totalSize` | int | 解压后总大小 |
| `compressedSize` | int | 归档大小 |
| `entries` | array | 条目清单 `[{ name, size, compressedSize, isDir, mode, modified, contentPreview }]` |
| `truncated` | bool | 条目清单是否被截断 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 条目数上限 5000、单条目解压上限 64 MiB、总解压上限 256 MiB；归档输入上限 32 MiB。

---

## file-analyze

按魔数识别文件类型并给出摘要（File Analyzer）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `data` | string | 是 | 文件内容的 Base64 |
| `fileName` | string | 否 | 文件名，用于推断扩展名与 MIME |
| `digestAlgorithm` | string | 否 | 摘要算法，默认 `SHA256` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `fileName` | string | 文件名 |
| `size` | int | 字节数 |
| `mimeType` | string | 媒体类型 |
| `extension` | string | 扩展名 |
| `description` | string | 类型说明 |
| `isTextual` | bool | 是否文本类 |
| `magicHex` | string | 文件头十六进制 |
| `digest` | string | 摘要值 |
| `digestAlgorithm` | string | 摘要算法 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 类型识别基于**魔数表**，未收录的类型按扩展名与 MIME 回落。

---

## id-generator

生成 UUID / ULID / 雪花 ID / NanoID（ID Generator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `format` | string | 否 | `uuid` `ulid` `snowflake` `nanoid`，默认 `uuid` |
| `count` | int | 否 | 生成条数，默认 1 |
| `length` | int | 否 | NanoID 长度，默认 21 |
| `epoch` | int | 否 | 雪花 ID 起始纪元（毫秒） |
| `nodeId` | int | 否 | 雪花 ID 节点号，取值 0–1023 |
| `uppercase` | bool | 否 | 是否转大写，默认 `false` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `format` | string | ID 格式 |
| `ids` | array | 生成的 ID 列表 |
| `count` | int | 生成条数 |
| `length` | int | ID 长度 |

### 错误

`INVALID_FIELD`。

> 雪花 ID 的纪元与节点号由请求指定，**不保证**全局唯一。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
