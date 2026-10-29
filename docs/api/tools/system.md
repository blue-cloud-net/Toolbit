# 系统 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **7** 个系统类工具。端点均为 `POST /api/v1/tools/{slug}`，全部为 `light` 档位。

| slug | 名称 |
| --- | --- |
| [curl-generator](#curl-generator) | cURL 命令生成 |
| [crontab-generator](#crontab-generator) | Crontab 表达式生成 |
| [chmod-calc](#chmod-calc) | 权限计算 |
| [nginx-config](#nginx-config) | Nginx 配置生成 |
| [firewall-rules](#firewall-rules) | 防火墙规则生成 |
| [docker-compose-gen](#docker-compose-gen) | Compose 配置生成 |
| [lookup-table](#lookup-table) | 速查表 |

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

## nginx-config

生成 Nginx 全局与站点配置（Nginx Config Generator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 否 | `global` 全局段 / `site` 站点段，默认 `site` |
| `serverName` | string | 否 | 站点域名 |
| `listenPort` | int | 否 | 监听端口，默认 80 |
| `root` | string | 否 | 站点根目录 |
| `indexFiles` | array | 否 | 索引文件列表 |
| `enableSsl` | bool | 否 | 是否启用 HTTPS，默认 `false` |
| `sslCertificate` | string | 否 | 证书路径 |
| `sslPrivateKey` | string | 否 | 私钥路径 |
| `enableProxy` | bool | 否 | 是否启用反向代理，默认 `false` |
| `proxyPass` | string | 否 | 代理目标 |
| `proxyPath` | string | 否 | 代理路径前缀 |
| `enableGzip` | bool | 否 | 是否启用 Gzip，默认 `false` |
| `enableCORS` | bool | 否 | 是否输出 CORS 段，默认 `false` |
| `allowOrigin` | string | 否 | 允许的来源，默认 `*` |
| `enableWebSocket` | bool | 否 | 是否启用 WebSocket 升级，默认 `false` |
| `enableLoadBalance` | bool | 否 | 是否输出 upstream 段，默认 `false` |
| `upstreams` | array | 否 | 上游地址列表 |
| `clientMaxBodySize` | string | 否 | 请求体上限，如 `10m` |
| `workerProcesses` | int | 否 | `global` 模式下的 worker 进程数 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 配置模式 |
| `config` | string | 配置文本 |
| `sections` | array | 配置段标题 |
| `warnings` | array | 配置提示 |

### 错误

`INVALID_FIELD`。

> 生成的是**配置片段**，需自行合并到 `nginx.conf`；证书与私钥只输出路径，不输出内容。

---

## firewall-rules

生成 ufw / firewalld / iptables / nftables 规则（Firewall Rule Generator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `tool` | string | 否 | `ufw` `firewalld` `iptables` `nftables`，默认 `ufw` |
| `scenario` | string | 否 | `http` `https` `ssh` `database` `denyAll` `allowAll` |
| `ports` | array | 否 | 端口列表，支持 `8000-8100` 区间写法 |
| `protocol` | string | 否 | `tcp` `udp` `any`，默认 `tcp` |
| `source` | string | 否 | 来源地址或网段 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `tool` | string | 防火墙工具 |
| `scenario` | string | 场景 |
| `rules` | array | 规则清单 `[{ action, port, protocol, source, description }]` |
| `commands` | array | 可执行命令 |
| `notes` | array | 使用提示 |

### 错误

`INVALID_FIELD`。

> 输出的是**待执行命令**，工具**不会**代为执行；上线前请先在测试环境验证。

---

## docker-compose-gen

由服务定义生成 Compose 配置（Docker Compose Generator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `projectName` | string | 是 | 项目名，用作服务名前缀 |
| `services` | array | 是 | 服务定义列表，见下表 |
| `networkName` | string | 否 | 自定义网络名 |
| `includeVolumes` | bool | 否 | 是否输出具名卷，默认 `false` |
| `includeNetwork` | bool | 否 | 是否输出自定义网络，默认 `true` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `compose` | string | Compose 配置文本 |
| `serviceCount` | int | 服务数量 |
| `warnings` | array | 配置提示 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 生成的是 Compose 文本，不含 `build` 段与密钥。

---

## lookup-table

查询端口、退出码与时间戳格式（Lookup Table）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `kind` | string | 否 | `port` `exitcode` `timestamp`，默认 `port` |
| `key` | string | 是 | 查询词：端口号、退出码或关键字 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `kind` | string | 类别 |
| `query` | string | 规范化后的查询词 |
| `matched` | bool | 是否精确命中 |
| `entry` | object | 命中的条目 `{ kind, key, label, description, extra[] }`；未命中时为 `null` |
| `candidates` | array | 模糊检索的候选条目 |
| `dataVersion` | string | 数据集版本 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 内置数据集离线查询，**不出网**；数据集版本随发布更新。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
