# 格式化 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **9** 个格式化类工具。端点均为 `POST /api/v1/tools/{slug}`，全部为 `light` 档位。

| slug | 名称 |
| --- | --- |
| [json-format](#json-format) | JSON 格式化 |
| [js-format](#js-format) | JavaScript 格式化 |
| [sql-format](#sql-format) | SQL 格式化 |
| [html-format](#html-format) | HTML 格式化 |
| [xml-format](#xml-format) | XML 格式化 |
| [toml-format](#toml-format) | TOML 格式化 |
| [yaml-format](#yaml-format) | YAML 格式化 |
| [format-convert](#format-convert) | 格式转换 |
| [color-convert](#color-convert) | 颜色转换 |

---

## 通用约定

所有格式化工具共享以下字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format`（默认）\| `minify` \| `validate` |
| `indent` | int | 否 | 缩进宽度，默认 2 |

- `action: validate` **不**改写内容，只返回是否合法；
- 语法错误返回 `INVALID_JSON` / `INVALID_XML` / `INVALID_YAML` / `INVALID_TOML`，`params` 含 `line` 与 `column`；
- 服务端**必须**在解析失败时**原样保留**用户输入（前端在错误提示中展示）。

---

## json-format

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format` \| `minify` \| `validate` |
| `indent` | int | 否 | 缩进宽度（默认 2）；`minify` 时忽略 |
| `sortKeys` | bool | 否 | 是否按键排序 |
| `escapeUnicode` | bool | 否 | 是否转义非 ASCII 字符 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 格式化结果 |
| `valid` | bool | 是否合法 |
| `error` | object | `{ line, column, message }` |

### 错误

`INVALID_JSON`、`INVALID_FIELD`。

---

## js-format

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format` \| `minify` |
| `indent` | int | 否 | 缩进宽度 |
| `quotes` | string | 否 | `single` \| `double`（默认） |
| `semi` | bool | 否 | 是否保留分号 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `valid` | bool | 是否合法 |
| `error` | object | `{ line, column, message }` |

### 错误

`INVALID_FIELD`。

> 解析与打印由 JS 解析器完成（纯语法操作，**不执行**代码）。压缩时保留变量名，
> 避免改变运行时行为。

---

## sql-format

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format` \| `minify` \| `validate` |
| `dialect` | string | 否 | `standard`（默认）\| `mysql` \| `pg` \| `sqlite` |
| `keywordCase` | string | 否 | `upper` \| `lower` \| `keep`（默认） |
| `indent` | int | 否 | 缩进宽度 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `valid` | bool | 是否合法 |
| `error` | object | `{ line, column, message }` |

### 错误

`INVALID_FIELD`。

> SQL 格式化是**纯文本**处理，**不**执行 SQL：只做分词与重排空白，token 文本
> （除关键字大小写）保持不变。方言差异**可能**导致语义变化，前端**应当**提示用户复核。

---

## html-format

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format` \| `minify` \| `validate` |
| `indent` | int | 否 | 缩进宽度 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `valid` | bool | 是否合法 |
| `error` | object | `{ line, column, message }` |

### 错误

`INVALID_FIELD`、`INVALID_XML`。

> 输出**必须**经过 HTML 转义上下文处理，**禁止**直接回显未净化内容到可执行上下文。

---

## xml-format

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format` \| `minify` \| `validate` |
| `indent` | int | 否 | 缩进宽度 |
| `keepComments` | bool | 否 | 是否保留注释（默认 `true`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `valid` | bool | 是否合法 |
| `error` | object | `{ line, column, message }` |

### 错误

`INVALID_XML`、`INVALID_FIELD`。

### 警告

`COMMENTS_LOST` —— 规范化过程中注释被丢弃。

---

## toml-format

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format` \| `minify` \| `validate` |
| `indent` | int | 否 | 缩进宽度 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `valid` | bool | 是否合法 |
| `error` | object | `{ line, column, message }` |
| `commentsLost` | bool | 注释是否丢失 |

### 错误

`INVALID_TOML`、`INVALID_FIELD`。

### 警告

`COMMENTS_LOST` —— TOML 解析库不保留注释。

---

## yaml-format

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `action` | string | 否 | `format` \| `minify` \| `validate` |
| `indent` | int | 否 | 缩进宽度 |
| `sortKeys` | bool | 否 | 是否按键排序 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 结果 |
| `valid` | bool | 是否合法 |
| `error` | object | `{ line, column, message }` |

### 错误

`INVALID_YAML`、`INVALID_FIELD`。

### 警告

`COMMENTS_LOST` —— YAML 解析库不保留注释。

> YAML 支持 `anchors` / `aliases`；为避免别名展开攻击（billion laughs），**必须**限制别名展开数量，超限返回 `TOO_LARGE`。

---

## format-convert

JSON / YAML / TOML / XML 之间的结构转换。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `from` | string | 是 | `json` \| `yaml` \| `toml` \| `xml` |
| `to` | string | 是 | `json` \| `yaml` \| `toml` \| `xml` |
| `indent` | int | 否 | 缩进宽度，默认 2 |
| `pretty` | bool | 否 | 是否美化输出（默认 `true`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 转换结果 |
| `from` | string | 实际使用的源格式 |
| `to` | string | 实际使用的目标格式 |
| `lossy` | bool | 是否有信息丢失 |
| `lossyReason` | string | 丢失原因 |

### 错误

`INVALID_JSON` / `INVALID_YAML` / `INVALID_TOML` / `INVALID_XML`、`UNSUPPORTED_FORMAT`。

### 警告

`LOSSY_FORMAT` —— 结构无法完全映射（例如 XML 元素名无法表达为 JSON key，或 TOML 顶层非表结构）。

---

## color-convert

颜色格式转换与对比度检查（Color Converter）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | string | 是 | 颜色值 |
| `inputFormat` | string | 否 | `hex` `rgb` `hsl` `oklch`，默认 `hex` |
| `outputFormats` | array | 否 | 需要输出的格式列表，默认全部 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `input` | string | 输入颜色 |
| `inputFormat` | string | 输入格式 |
| `hex` | string | 十六进制表示 |
| `rgb` | string | RGB 表示 |
| `hsl` | string | HSL 表示 |
| `oklch` | string | OKLCH 表示 |
| `contrastOnWhite` | number | 相对白底的对比度 |
| `contrastOnBlack` | number | 相对黑底的对比度 |
| `isDark` | bool | 是否为深色 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 对比度按 WCAG 相对亮度公式计算；只支持 IPv4 / sRGB 色彩空间。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
