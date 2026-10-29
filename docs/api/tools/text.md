# 文本 API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

覆盖 **2** 个文本类工具。端点均为 `POST /api/v1/tools/{slug}`。

| slug | 名称 | 档位 |
| --- | --- | --- |
| [regex-tester](#regex-tester) | 正则测试 | standard |
| [text-convert](#text-convert) | 文本转换 | light |

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

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
