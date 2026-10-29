# 文本 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **7** 个文本类工具。端点均为 `POST /api/v1/tools/{slug}`。

| slug | 名称 | 档位 |
| --- | --- | --- |
| [regex-tester](#regex-tester) | 正则测试 | standard |
| [text-convert](#text-convert) | 文本转换 | light |
| [text-diff](#text-diff) | 文本对比 | light |
| [regex-validate](#regex-validate) | 正则校验 | light |
| [regex-build](#regex-build) | 正则生成 | light |
| [data-mask](#data-mask) | 数据脱敏 | light |
| [text-stats](#text-stats) | 文本统计 | light |

---

## regex-tester

正则表达式测试与匹配。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `pattern` | string | 是 | 正则表达式（RE2 语法） |
| `flags` | array | 否 | 标志名列表：`i` `m` `s` `x` |
| `input` | object | 是 | `{ data }` |
| `maxMatches` | int | 否 | 最多返回匹配数，默认 1000 |
| `groupNames` | array | 否 | 捕获组名称（按顺序对应第 1、2 … 组） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `matches` | array | `[{ index, text, line, column, groups[], names? }]` |
| `count` | int | 匹配总数 |
| `truncated` | bool | 是否因上限截断 |
| `pattern` | string | 实际使用的正则 |

### 错误

`INVALID_REGEX`、`INVALID_FIELD`、`TOO_MANY_ITEMS`。

### 警告

`REGEX_DIALECT_DIFFERENCE` —— Go 使用 RE2 语法，**不支持**反向引用、lookahead / lookbehind；与 JavaScript / PCRE 的行为可能不同。

> 服务端使用 RE2（**线性时间**匹配，无灾难性回溯）。表达式长度上限 64 KB。

---

## text-convert

文本大小写与命名风格转换。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data }` |
| `style` | string | 否 | 见下表；**空串**表示只用 `separator` 连接 |
| `separator` | string | 否 | 自定义分隔符；仅在 `style` 为空串时生效 |

| `style` | 结果 |
| --- | --- |
| `lower` | 全部小写 |
| `upper` | 全部大写 |
| `title` | 单词首字母大写（`Http Server`） |
| `sentence` | 句首大写（`Http server`） |
| `camel` | 驼峰（`camelCase`） |
| `pascal` | 帕斯卡（`PascalCase`） |
| `snake` | 下划线（`snake_case`） |
| `kebab` | 短横线（`kebab-case`） |
| `constant` | 常量（`CONSTANT_CASE`） |
| `dot` | 点分（`dot.case`） |
| `path` | 路径（`path/case`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 转换结果 |
| `style` | string | 实际使用的风格 |
| `words` | array | 切分出的词元 |

### 错误

`INVALID_FIELD`、`MISSING_FIELD`。

> 单词切分规则：**仅**非字母数字字符视为分隔符，不做驼峰反向切分。
> 因此 `HTTPServer` 是一个词元（`snake` 风格下得到 `httpserver`），
> `http-server` 是两个词元（`http_server`）。

---

## text-diff

文本行级与 JSON 结构化差异对比（Text Diff）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 否 | `text` 行级对比 / `json` 结构化对比，默认 `text` |
| `left` | object | 是 | `{ data }`，基准侧文本 |
| `right` | object | 是 | `{ data }`，对比侧文本 |
| `ignoreWhitespace` | bool | 否 | 忽略行尾空白，默认 `false` |
| `contextLines` | int | 否 | 上下文行数，取值 0–100 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 实际使用的对比模式 |
| `identical` | bool | 两侧是否完全一致 |
| `lines` | array | `[{ op, leftLineNum, rightLineNum, text }]` |
| `changes` | array | JSON 模式的结构化差异 `[{ path, op, left, right }]` |
| `stat` | object | `{ added, removed, unchanged, changes }` |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 单侧输入上限 512 KiB、行数上限 20000。

---

## regex-validate

只校验正则语法，不做匹配（Regex Validator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `pattern` | string | 是 | 正则表达式（RE2 语法） |
| `flags` | array | 否 | 标志名列表：`i` `m` `s` `x` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `valid` | bool | 表达式是否合法 |
| `dialect` | string | 正则方言，固定 `RE2` |
| `flags` | array | 生效的标志 |
| `error` | object | 语法错误 `{ code, message, offset, position }`；合法时为 `null` |
| `suggestions` | array | 改进建议 `[{ code, message, priority }]` |
| `captureGroups` | int | 捕获组数量 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 服务端使用 RE2（**线性时间**匹配，无灾难性回溯）；只校验语法，不做匹配。

---

## regex-build

由结构化选项拼装正则表达式（Regex Builder）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `parts` | array | 否 | 片段名列表，见下表 |
| `prefix` | string | 否 | 固定前缀（自动转义） |
| `suffix` | string | 否 | 固定后缀（自动转义） |
| `quantifier` | object | 否 | `{ min, max }`，`max` 为 -1 表示无上限 |
| `anchored` | bool | 否 | 是否加 `^` 与 `$` 锚点，默认 `false` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pattern` | string | 生成的表达式 |
| `valid` | bool | 生成的表达式是否合法 |
| `explanation` | array | 逐段释义 `[{ part, fragment, description }]` |
| `error` | object | 校验错误；合法时为 `null` |
| `captureGroups` | int | 捕获组数量 |

### 错误

`INVALID_FIELD`。

> 生成是纯字符串拼装，结果仍按 RE2 方言校验；固定前后缀会自动转义。

---

## data-mask

手机号、证件号、邮箱等敏感数据脱敏（Data Masking）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `text` | string | 是 | 待脱敏文本 |
| `types` | array | 否 | 脱敏类型：`phone` `idCard` `bankCard` `email` `name` `ip` |
| `position` | string | 否 | `keepHead` `keepTail` `keepBoth` `hash`，默认 `keepHead` |
| `keepCount` | int | 否 | 保留字符数，取值 0–32 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `masked` | string | 脱敏后的文本 |
| `matches` | array | 各类命中统计 `[{ type, count, sample }]` |
| `total` | int | 脱敏总数 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> **纯文本处理**：输入不落盘、不写任何浏览器存储；命中统计里的 `sample` 同样已脱敏。

---

## text-stats

字数、行数与字符构成统计（Text Statistics）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `text` | string | 是 | 待统计文本 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `characters` | int | 字符总数 |
| `charactersNoSpace` | int | 不含空白的字符数 |
| `words` | int | 词数（中文按字、英文按词） |
| `chineseChars` | int | 中文字符数 |
| `englishWords` | int | 英文单词数 |
| `digits` | int | 数字字符数 |
| `punctuation` | int | 标点符号数 |
| `lines` | int | 行数 |
| `paragraphs` | int | 段落数 |
| `bytes` | int | UTF-8 字节数 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 中文按字计、英文按词计，`words` 为两者之和。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
