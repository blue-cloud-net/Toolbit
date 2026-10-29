# 数据 API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

覆盖 **6** 个数据类工具。端点均为 `POST /api/v1/tools/{slug}`。

| slug | 名称 | 档位 |
| --- | --- | --- |
| [timestamp](#timestamp) | 时间戳转换 | light |
| [uuid-gen](#uuid-gen) | UUID 生成 | light |
| [date-calc](#date-calc) | 日期计算 | light |
| [country-code](#country-code) | 国家代码 | light |
| [cn-regions](#cn-regions) | 中国行政区划 | light |
| [qr-code](#qr-code) | 二维码生成 | light |

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

---

## cn-regions

中国行政区划查询。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `query` | string | 否 | 名称或编码关键词 |
| `province` | string | 否 | 指定省级行政区，缩小城市查询范围 |

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

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
