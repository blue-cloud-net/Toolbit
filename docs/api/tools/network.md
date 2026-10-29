# 网络 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **9** 个网络类工具。其中 4 个**需要出网**，`websocket-test` 为**纯前端**工具（无后端端点），
`mime-lookup` 与 `http-status` 为**纯查表**工具（不出网）。

| slug | 名称 | 传输 | 端点 |
| --- | --- | --- | --- |
| [cidr-calc](#cidr-calc) | CIDR 计算 | JSON | `POST /api/v1/tools/cidr-calc` |
| [ssl-check](#ssl-check) | SSL 证书检测 | WebSocket | `GET /api/v1/tools/ssl-check/stream` |
| [dns-lookup](#dns-lookup) | DNS 查询 | JSON | `POST /api/v1/tools/dns-lookup` |
| [dns-tester](#dns-tester) | DNS 多记录测试 | JSON | `POST /api/v1/tools/dns-tester` |
| [websocket-test](#websocket-test) | WebSocket 测试 | 浏览器原生 | 无 |
| [mime-lookup](#mime-lookup) | MIME 类型查询 | JSON | `POST /api/v1/tools/mime-lookup` |
| [http-status](#http-status) | HTTP 状态码查询 | JSON | `POST /api/v1/tools/http-status` |
| [device-headers](#device-headers) | 设备请求头 | JSON | `POST /api/v1/tools/device-headers` |
| [ip-range-expand](#ip-range-expand) | IP 段展开 | JSON | `POST /api/v1/tools/ip-range-expand` |

---

## 出网安全（全部出网工具适用）

实现位于 `backend/internal/netguard`，全部出网工具**必须**经它建连，工具内**禁止**直接
`http.Get` / `net.Dial`。

- 目标主机**先解析 DNS**，全部解析结果**必须**通过私网校验；只要有一个地址落在禁止网段就整体拒绝；
- **禁止**访问下列网段：

  | 范围 | 说明 |
  | --- | --- |
  | `0.0.0.0/8` | 未指定 |
  | `10.0.0.0/8` `172.16.0.0/12` `192.168.0.0/16` | 私有 |
  | `100.64.0.0/10` | 运营商级 NAT |
  | `127.0.0.0/8` | 环回 |
  | `169.254.0.0/16` | 链路本地 |
  | `192.0.0.0/24` `192.0.2.0/24` `198.51.100.0/24` `203.0.113.0/24` | 协议分配 / 文档 |
  | `198.18.0.0/15` | 基准测试 |
  | `224.0.0.0/4` `240.0.0.0/4` | 组播 / 保留 |
  | `::/128` `::1/128` | 未指定 / 环回 v6 |
  | `64:ff9b::/96` `100::/64` `2001:db8::/32` | NAT64 / 丢弃 / 文档 |
  | `fc00::/7` `fe80::/10` `ff00::/8` | ULA / 链路本地 / 组播 |

- **必须**校验 IPv4-mapped IPv6：`::ffff:127.0.0.1` 会先还原为 `127.0.0.1` 再判定；
- 连接时使用**已校验的 IP**（防 DNS rebinding），**不**二次解析；
- 重定向**逐跳校验**，最多 3 跳；**禁止**自动跟随到内网；
- **只允许固定端口**：`43`（whois）、`53`（自定义 DNS 服务器）、`80`（`domain-verify` 的
  `http-file`）、`443`（HTTPS / DoH / TLS 探测）、`853`（DoT）。端口由**服务端**按工具语义选择，
  用户**不能**指定任意端口。

### 出网开关与例外

- `TOOLBIT_NETWORK_ENABLED=false` 时**全部**出网工具返回 `NETWORK_DISABLED`（`403`）；
- `TOOLBIT_NETWORK_ALLOWLIST` 可放行指定的内网主机（精确主机名或 `*.suffix`），
  命中后**跳过私网校验**，但端口与协议白名单仍然生效；
- **系统解析器例外**：`dns-lookup` / `dns-tester` 在用默认解析器时读取 `/etc/resolv.conf`
  并直接查询，**不**做私网校验 —— 这些服务器来自运维配置，容器与桌面环境里它们常常就是
  回环上的桩解析器。用户**自定义**的服务器仍然要过出网防护。

---

## cidr-calc

CIDR 子网计算（**不出网**）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cidr` | string | 二选一 | 如 `192.168.1.0/24` |
| `start` / `end` | string | 二选一 | 地址范围（反推 CIDR） |
| `split` | number | 否 | 按指定前缀长度拆分；仅 `cidr` 模式支持 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `version` | number | 地址族：`4` 或 `6` |
| `network` | string | 网络地址（含前缀长度） |
| `prefix` | number | 前缀长度 |
| `mask` | string | 子网掩码（IPv4 点分十进制，IPv6 压缩十六进制） |
| `firstAddress` / `lastAddress` | string | 网段首末地址 |
| `broadcast` | string | 广播地址；IPv6 为空串 |
| `firstUsable` / `lastUsable` | string | 可用地址区间 |
| `hostCount` | string | 可用主机数（**十进制字符串**） |
| `totalAddresses` | string | 地址总数（十进制字符串） |
| `cidrs` | array | 按 `split` 拆分后的子网 |
| `range` | array | 范围反推得到的最小 CIDR 集合 |

### 错误

`INVALID_CIDR`、`INVALID_FIELD`、`MISSING_FIELD`、`TOO_MANY_ITEMS`。

> IPv4 与 IPv6 **分别**计算，**不**混合。
> `hostCount` 与 `totalAddresses` 用字符串表达：IPv6 的地址数量可达 2¹²⁸，超出 64 位整数。
> IPv4 的 `/31` 计 2 个可用地址、`/32` 计 1 个；IPv6 无广播概念，全部地址可用。
> 一次拆分或反推最多产出 **256** 个子网，超出返回 `TOO_MANY_ITEMS`。

---

## ssl-check

TLS 握手与证书链检测（**流式**，需出网）。

### 传输

WebSocket：`GET /api/v1/tools/ssl-check/stream`。协议细节见 [../README.md §11](../README.md)。

> 该 slug 在 catalog 中为 `transport: websocket`：`POST /api/v1/tools/ssl-check` 返回
> `405` + `METHOD_NOT_ALLOWED`，JSON 通道**不**存在。

### 客户端首帧

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `host` | string | 是 | 目标主机 |
| `port` | int | 否 | 默认 `443` |
| `sni` | string | 否 | SNI 主机名（缺省用 `host`） |
| `depth` | int | 否 | 证书链深度上限，默认 `5`，最大 `10` |
| `versions` | array | 否 | 如 `["TLS1.3","TLS1.2"]`，缺省用 Go 默认范围 |
| `ciphers` | array | 否 | 密码套件名列表；TLS 1.3 套件**无法**裁剪 |
| `timeoutMs` | int | 否 | 超时，默认 `30000` |

### 服务端帧

| `type` | 载荷 |
| --- | --- |
| `progress` | `{ phase, done, total }`；`phase` ∈ `resolve` \| `connect` \| `handshake` \| `chain`，`total` 恒为 `4` |
| `result` | `{ data }`，`data` 见下表 |
| `error` | `{ error: { code, message, params } }` |

`result.data` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `host` / `port` | string / number | 目标 |
| `addr` | string | 实际连接的 IP |
| `protocol` | string | 协议族名，固定 `TLS` |
| `negotiatedVersion` | string | 协商出的版本，如 `TLS1.3` |
| `cipherSuite` | string | 协商出的套件名 |
| `certificates` | array | `[{ pem, info }]`，`info` 为证书结构（同 `cert-parse`） |
| `chainValid` | bool | 链是否通过验证（含主机名匹配） |
| `chainErrors` | array | 验证错误码列表 |
| `verified` | bool | 是否用**系统信任根**完成验证 |
| `implicitTrust` | bool | 是否退化为「链内自洽」判定 |
| `ocspStapled` | bool | 服务端是否携带 OCSP 装订响应 |
| `grade` | string | 综合评级：`A+` / `A` / `B` / `C` / `T` |
| `checks` | array | 逐项检查结论，`[{ id, label, passed, detail? }]` |

#### 评级规则

评级只描述**本次握手与证书链**的观测结果，不代表站点整体安全评分。

每项检查按「通过 / 不通过」计分，不通过项按严重度扣分（满分 100）：

| 检查项 `id` | 含义 | 不通过扣分 |
| --- | --- | --- |
| `chain` | 证书链完整（含主机名匹配） | 100（致命） |
| `trust` | 受信任颁发机构签发 | 40（未受信任）/ 20（隐式信任） |
| `tlsVersion` | 支持 TLS 1.2 及以上 | 100（致命） |
| `tls13` | 协商出 TLS 1.3 | 5 |
| `ocsp` | 携带 OCSP 装订响应 | 5 |
| `cipher` | 协商出密码套件 | 100（致命） |

剩余分映射到等级：`100 → A+`、`≥90 → A`、`≥75 → B`、`>0 → C`、`0 → T`。

> 隐式信任（无法访问系统信任根）按 20 分计，而非 40 分 —— 它反映的是**探测端**
> 的信任库缺失，不是服务端的缺陷。

### 错误

`SSRF_BLOCKED`、`DNS_ERROR`、`TLS_ERROR`、`HOST_UNREACHABLE`、`UPSTREAM_TIMEOUT`、
`NETWORK_DISABLED`、`INVALID_FIELD`、`INVALID_DOMAIN`。

### 关闭码

`1000` 正常、`1008` 输入非法或被策略拒绝、`1011` 服务端内部错误。

### 警告

链不通过或缺少系统信任根时给 `UNTRUSTED_CHAIN`；叶证书自签名给 `SELF_SIGNED_CERT`；
过期 / 未生效分别给 `EXPIRED_CERT` / `NOT_YET_VALID`。

> 探测**不**在握手期校验链（否则拿不到无效证书的细节），而是拿到链之后另行验证。
> 验证时优先加载本地信任库（`/etc/ssl/certs/ca-certificates.crt` 等）；信任库不可用时
> `verified` 为 `false`、`implicitTrust` 为 `true`。
> 空闲 60 秒关闭；单次任务总预算 30 秒；同一 IP **同时最多 1 条**连接。

---

## dns-lookup

DNS 记录查询（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `host` | string | 是 | 域名 |
| `type` | string | 是 | `A` `AAAA` `CNAME` `MX` `TXT` `NS` `SOA` `SRV` `CAA` `PTR` `ANY` |
| `servers` | array | 否 | 自定义服务器 `[{ proto, addr }]`；`proto` ∈ `udp` \| `tcp` \| `dot` \| `doh`，最多 8 台 |
| `timeoutMs` | int | 否 | 单服务器超时，默认 `5000` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `host` / `type` | string | 回显请求 |
| `perServer` | array | `[{ server, proto, latencyMs, answers[], error }]` |

`answers[]` 为 `{ name, type, ttl, value }`；`error` 成功时为 `null`，
形如 `{ code, message }`。

### 错误

`INVALID_FIELD`、`MISSING_FIELD`、`TOO_MANY_ITEMS`、`NETWORK_DISABLED`。

> **单台服务器失败不返回错误**：错误记在对应 `perServer[].error` 里，整体仍是 `200`。
> 默认使用系统解析器；指定 `servers` 时**只**查询指定服务器。
> DoH 需要完整的 `https://` 地址；DoT 固定 `853` 端口。
> `ANY` 查询可能被上游服务器拒绝，部分服务商已不再支持。

---

## dns-tester

DNS 综合检测（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `action` | string | 是 | `whois` \| `dnssec` \| `ptr` |
| `domain` | string | 视 action | 域名 |
| `ip` | string | 视 action | IP 地址（`ptr` 可用它代替域名） |
| `timeoutMs` | int | 否 | 超时，默认 `20000` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `action` | string | 回显动作 |
| `whois` | object \| null | 仅 `whois` |
| `dnssec` | object \| null | 仅 `dnssec` |
| `ptr` | object \| null | 仅 `ptr` |

各动作的载荷：

| `action` | 载荷结构 |
| --- | --- |
| `whois` | `{ source, registrar, creationDate, expirationDate, updatedDate, nameServers[], status[], registrantCountry, registries[] }` |
| `dnssec` | `{ supported, dnssecEnabled, dsPresent, sigPresent, chainOfTrust[] }` |
| `ptr` | `{ reverse, addresses[], names[] }` |

### 错误

`INVALID_FIELD`、`MISSING_FIELD`、`DNS_ERROR`、`UPSTREAM_ERROR`、`UPSTREAM_TIMEOUT`、
`SSRF_BLOCKED`、`NETWORK_DISABLED`。

> `whois` 先走 RDAP（`https://rdap.org`，HTTPS 443），失败后回退到 whois 协议
> （`whois.iana.org:43`，按 `refer:` 最多再跳 1 次）。端口由服务端固定，用户无法指定。
> 部分注册局限制查询或脱敏，**缺失字段以空串表示**。
> `dnssec` 逐级向上查 DS（最多 3 级），并结合 DNSKEY 判断是否已启用。
> `ptr` 在完成反向解析后会对每个名字再做一次正向解析，用于暴露正反不一致。

---

## websocket-test

WebSocket 连接测试（**纯前端**，无后端端点）。

### 说明

本工具在**浏览器内**使用原生 `WebSocket` API 直连目标服务器，**不经过**后端。

- catalog 中 `transport: client`；`POST /api/v1/tools/websocket-test` 返回 `404` + `TOOL_NOT_FOUND`；
- **不**受服务端出网策略约束 —— 由**用户浏览器**发起连接；
- 结果在浏览器端渲染，**不**写入服务端统计（仅记录工具被打开）。

### 前端参数

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `url` | string | `ws://` 或 `wss://` 地址 |
| `protocols` | array | 子协议 |
| `message` | string | 连接后发送的消息 |
| `headers` | object | 浏览器**不支持**自定义请求头，此项仅作提示 |

### 结果

| 字段 | 说明 |
| --- | --- |
| `connected` | 是否连接成功 |
| `openMs` | 握手耗时 |
| `protocol` | 协商出的子协议 |
| `extension` | 协商出的扩展 |
| `sent` / `received` | 收发内容 |
| `closed` | 关闭码与原因 |

### 注意事项

> 浏览器**无法**自定义请求头、**无法**使用自定义 TLS 证书、**无法**指定源端口；这些限制**必须**在 UI 上明确说明。
> 连接**跨域**时遵循浏览器同源策略；本页**不**使用服务端 CORS 配置。

---

## mime-lookup

MIME 媒体类型查询（**不出网**）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `extension` | string | 二选一 | 扩展名；可写 `png`、`.png` 或 `report.png` |
| `mimeType` | string | 二选一 | 媒体类型；可写 `image/png`，带参数时自动去掉参数 |

> 两个字段**互斥**：同时给出返回 `INVALID_FIELD`，都不给出返回 `MISSING_FIELD`。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 规范化后的查询词（扩展名小写去点；媒体类型小写去参数） |
| `kind` | string | 查询方式：`extension` \| `mimeType` |
| `matched` | bool | 是否命中数据集 |
| `mimeType` | string | 规范媒体类型；未命中时为空串 |
| `extensions` | array | 该媒体类型的常用扩展名（小写，不含点） |
| `category` | string | 顶层类型：`application` \| `audio` \| `font` \| `image` \| `message` \| `model` \| `multipart` \| `text` \| `video` |
| `isTextual` | bool | 是否是文本类媒体类型 |
| `charset` | string | 默认字符集；非文本类为空串 |
| `description` | string | 用途说明 |
| `aliases` | array | 历史别名与常见变体 |
| `dataVersion` | string | 数据集版本 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 数据来自**内置快照**，**不**读取系统 MIME 数据库（各发行版映射表不一致），**不**联网。
> 扩展名查询按「先到先得」命中：同一扩展名被多个类型描述时，返回数据集中靠前的那一个。
> 历史别名（如 `image/x-png`）会归一到规范类型，响应里的 `mimeType` 是规范值。
> 未收录**不是**错误：返回 `200` 与 `matched: false`。

---

## http-status

HTTP 状态码查询（**不出网**）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `code` | number | 二选一 | 状态码，`100`–`599` 的整数 |
| `query` | string | 二选一 | 关键字；纯数字时按状态码处理 |

> 两个字段**互斥**：同时给出返回 `INVALID_FIELD`，都不给出返回 `MISSING_FIELD`。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 规范化后的查询词 |
| `kind` | string | 查询方式：`code` \| `keyword` |
| `matched` | bool | 是否命中标准状态码 |
| `code` | number | 状态码；关键字查询取首条命中，未命中为 `0` |
| `reason` | string | 标准原因短语 |
| `category` | string | 类别：`informational` \| `success` \| `redirect` \| `clientError` \| `serverError` |
| `description` | string | 中文含义说明 |
| `deprecated` | bool | 是否已废弃（如 `305`、`306`） |
| `matches` | array | 命中的条目 `[{ code, reason, category, description, deprecated }]` |
| `total` | number | 命中条数 |
| `dataVersion` | string | 数据集版本 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 关键字匹配**原因短语**（不区分大小写）、**中文说明**与**类别中文名**，
> 因此 `not found`、`未找到`、`重定向` 都能命中。
> 落在 `100`–`599` 但未收录的状态码（如 `299`）**不是**错误：返回 `200`、`matched: false`，
> 并给出所属 `category`。
> 越界的状态码（如 `999`）返回 `INVALID_FIELD`，错误**不回显**用户输入。

---

## device-headers

生成客户端请求头并解析 User-Agent（Device Headers）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 否 | `generate` 生成 / `parseUa` 解析，默认 `generate` |
| `platform` | string | 否 | `chrome` `firefox` `safari` `edge` `android` `ios` `curl` |
| `version` | string | 否 | 浏览器版本号 |
| `os` | string | 否 | 操作系统名称与版本 |
| `userAgent` | string | 否 | `parseUa` 模式下的 User-Agent 串 |
| `acceptLanguage` | string | 否 | Accept-Language 头 |
| `referer` | string | 否 | Referer 头 |
| `customHeaders` | object | 否 | 额外请求头，键值对 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 模式 |
| `headers` | array | 请求头列表 `[{ name, value }]` |
| `userAgent` | string | User-Agent 串 |
| `parsed` | object | 解析结果 `{ browser, browserVersion, os, osVersion, deviceType, isBot, engine }` |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 生成的 UA 是**常见组合**，不代表任何真实浏览器；解析只做结构化识别，不做指纹追踪。

---

## ip-range-expand

IP 段与 CIDR 互转及批量展开（IP Range Expander）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `mode` | string | 否 | `expand` 展开 / `toCidr` 转 CIDR / `toRange` 转区间 |
| `cidr` | string | 否 | 网段；`expand` 与 `toRange` 模式使用 |
| `start` | string | 否 | 起始地址；`toCidr` 模式使用 |
| `end` | string | 否 | 结束地址；`toCidr` 模式使用 |
| `limit` | int | 否 | 展开条数上限，默认 1024 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 模式 |
| `cidrs` | array | 规范化后的 CIDR 列表 |
| `start` | string | 起始地址 |
| `end` | string | 结束地址 |
| `count` | int | 地址总数 |
| `addresses` | array | 展开的地址列表 |
| `truncated` | bool | 地址列表是否被上限截断 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 展开条数上限受 `limit` 约束，超出时 `truncated` 为 `true`。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封、限流、WebSocket 协议
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/06-security.md](../../standards/06-security.md) —— 出网与 SSRF 防护
