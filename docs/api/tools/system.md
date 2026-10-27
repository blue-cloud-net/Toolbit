# 系统 API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

覆盖 **3** 个系统类工具。端点均为 `POST /api/v1/tools/{slug}`，全部为 `light` 档位。

| slug | 名称 |
| --- | --- |
| [curl-generator](#curl-generator) | cURL 命令生成 |
| [crontab-generator](#crontab-generator) | Crontab 表达式生成 |
| [chmod-calc](#chmod-calc) | 权限计算 |

---

## curl-generator

生成可复制的 cURL 命令。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `method` | string | 是 | `GET` `POST` `PUT` `PATCH` `DELETE` `HEAD` `OPTIONS` |
| `url` | string | 是 | 目标 URL |
| `headers` | array | 否 | `[{ name, value }]` |
| `body` | string | 否 | 请求体 |
| `bodyType` | string | 否 | `json` \| `form` \| `raw` \| `binary` |
| `auth` | object | 否 | `{ type, username, password, token, bearer }` |
| `options` | object | 否 | `{ insecure, followRedirects, timeout, compressed, ipv4, verbose, proxy }` |
| `shell` | string | 否 | `bash`（默认）\| `cmd` \| `powershell` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `command` | string | 生成的命令 |

### 错误

`INVALID_FIELD`、`MISSING_FIELD`。

> 生成的是**文本**，服务端**不**执行命令。口令与 Token **必须**用可见占位符（如 `<PASSWORD>`）提示用户替换，**禁止**直接内联明文凭据。

---

## crontab-generator

生成与解释 Crontab 表达式。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `fields` | object | 二选一 | `{ minute, hour, dom, month, dow }` |
| `preset` | string | 二选一 | 预设名，如 `every-minute` `hourly` `daily` `weekly` `monthly` `yearly` |
| `timezone` | string | 否 | IANA 时区名 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `expression` | string | 五段式表达式 |
| `nextRuns` | array | 未来若干次触发时间（RFC3339） |
| `description` | string | 人类可读描述 |

### 错误

`INVALID_CRON`、`INVALID_FIELD`、`MISSING_FIELD`。

> 表达式为**五段**（分 时 日 月 周），**不**支持秒级与 `@reboot` 等扩展语法（`@daily` 等预设**会**被转换展开）。周字段 `0` 与 `7` 均表示周日。

---

## chmod-calc

Unix 权限计算与解释。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 二选一 | `symbolic`（如 `rwxr-xr-x`）\| `numeric`（如 `755`） |
| `owner` / `group` / `others` | string | 否 | 直接指定三类权限（`rwx` 组合） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `numeric` | string | 八进制权限（如 `755`） |
| `symbolic` | string | 符号权限（如 `rwxr-xr-x`） |
| `bits` | array | `[{ who, read, write, execute }]` |
| `meaning` | string | 说明（如「属主可读写执行，属组与其他人可读可执行」） |

### 错误

`INVALID_CHMOD`、`INVALID_FIELD`、`MISSING_FIELD`。

> 只解释**低 9 位**（`rwx`）；`setuid` / `setgid` / `sticky`（`4xxx` / `2xxx` / `1xxx`）**不在**本工具范围内，输入时**必须**返回 `INVALID_CHMOD`。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
