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
| `method` | string | 否 | HTTP 方法；缺省 `GET` |
| `url` | string | 是 | 目标 URL |
| `headers` | array | 否 | `[{ name, value, sensitive? }]`；`sensitive` 为真时值以占位符呈现 |
| `body` | string | 否 | 请求体 |
| `bodyFormat` | string | 否 | `json` \| `form` \| `raw` |
| `auth` | string | 否 | `none`（默认）\| `basic` \| `bearer` |
| `username` | string | `basic` 时是 | Basic 认证用户名 |
| `password` | string | 否 | Basic 认证口令（输出为 `<PASSWORD>`） |
| `token` | string | `bearer` 时是 | Bearer Token（输出为 `<TOKEN>`） |
| `insecure` | bool | 否 | 是否跳过 TLS 证书校验 |
| `followRedirects` | bool | 否 | 是否跟随重定向 |
| `timeout` | int | 否 | 超时秒数（0–3600） |
| `shell` | string | 否 | `bash`（默认）\| `sh` \| `powershell` \| `cmd` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `command` | string | 生成的命令 |
| `shell` | string | 目标 Shell |
| `placeholders` | array | 被占位符替代的凭据字段名 |

### 错误

`INVALID_FIELD`、`MISSING_FIELD`。

> 生成的是**文本**，服务端**不**执行命令。口令与 Token **必须**用可见占位符（如 `<PASSWORD>`）提示用户替换，**禁止**直接内联明文凭据。

---

## crontab-generator

生成与解释 Crontab 表达式。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `preset` | string | 二选一 | 预设名：`every-minute` `every-5-minutes` `hourly` `daily` `weekly` `monthly` `yearly` |
| `minute` | string | 二选一 | 分钟字段 |
| `hour` | string | 二选一 | 小时字段 |
| `dayOfMonth` | string | 二选一 | 日字段 |
| `month` | string | 二选一 | 月字段 |
| `dayOfWeek` | string | 二选一 | 星期字段 |
| `timezone` | string | 否 | IANA 时区名；缺省 UTC |
| `count` | int | 否 | 预览的触发次数；默认 5（≤ 10） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `expression` | string | 五段式表达式 |
| `description` | string | 人类可读描述 |
| `occurrences` | array | 未来若干次触发时间 `[{ time }]`（RFC 3339） |
| `timezone` | string | 使用的时区 |

### 错误

`INVALID_CRON`、`INVALID_FIELD`、`MISSING_FIELD`。

> 表达式为**五段**（分 时 日 月 周），**不**支持秒级与 `@reboot` 等扩展语法（`@daily` 等预设**会**被转换展开）。周字段 `0` 与 `7` 均表示周日。

---

## chmod-calc

Unix 权限计算与解释。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 二选一 | 权限值：符号形式（`rwxr-xr-x`）、八进制形式（`755`）或长形式（`u=rwx,g=rx,o=rx`） |
| `user` | array | 否 | 属主权限位（`read` / `write` / `execute`）；给出时按位**设置** |
| `group` | array | 否 | 属组权限位（`read` / `write` / `execute`） |
| `other` | array | 否 | 其他人权限位（`read` / `write` / `execute`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `octal` | string | 八进制权限（如 `755`） |
| `symbolic` | string | 符号权限（如 `rwxr-xr-x`） |
| `classes` | array | `[{ class, read, write, execute, symbol, numeric, meaning }]` |

### 错误

`INVALID_CHMOD`、`INVALID_FIELD`、`MISSING_FIELD`。

> 只解释**低 9 位**（`rwx`）；`setuid` / `setgid` / `sticky`（`4xxx` / `2xxx` / `1xxx`）**不在**本工具范围内，输入时会被**忽略**。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
