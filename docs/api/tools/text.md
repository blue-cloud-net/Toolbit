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
| `flags` | string | 否 | `g` `i` `m` `s` 可组合（默认 `g`） |
| `input` | object | 是 | `{ data, encoding }` |
| `maxMatches` | int | 否 | 最多返回匹配数，默认 1000 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `matches` | array | `[{ index, value, groups[], start, end }]` |
| `count` | int | 匹配总数 |
| `truncated` | bool | 是否因上限截断 |

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
| `input` | object | 是 | `{ data, encoding }` |
| `action` | string | 是 | 见下表 |

| `action` | 结果 |
| --- | --- |
| `lower` | 全部小写 |
| `upper` | 全部大写 |
| `title` | 单词首字母大写 |
| `sentence` | 句首大写 |
| `camel` | 驼峰（`camelCase`） |
| `pascal` | 帕斯卡（`PascalCase`） |
| `snake` | 下划线（`snake_case`） |
| `kebab` | 短横线（`kebab-case`） |
| `constant` | 常量（`CONSTANT_CASE`） |
| `dot` | 点分（`dot.case`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 转换结果 |

### 错误

`INVALID_FIELD`、`UNSUPPORTED_MODE`。

> 单词切分规则：非字母数字字符视为分隔符；连续大写按驼峰规则处理（`HTTPServer` → `http_server`）。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
