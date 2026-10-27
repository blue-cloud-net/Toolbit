# 网络 API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

覆盖 **5** 个网络类工具。其中 4 个**需要出网**，`websocket-test` 为**纯前端**工具（无后端端点）。

| slug | 名称 | 传输 | 端点 |
| --- | --- | --- | --- |
| [cidr-calc](#cidr-calc) | CIDR 计算 | JSON | `POST /api/v1/tools/cidr-calc` |
| [ssl-check](#ssl-check) | SSL 证书检测 | WebSocket | `GET /api/v1/tools/ssl-check/stream` |
| [dns-lookup](#dns-lookup) | DNS 查询 | JSON | `POST /api/v1/tools/dns-lookup` |
| [dns-tester](#dns-tester) | DNS 多记录测试 | JSON | `POST /api/v1/tools/dns-tester` |
| [websocket-test](#websocket-test) | WebSocket 测试 | 浏览器原生 | 无 |

---

## 出网安全（全部出网工具适用）

- 目标主机**先解析 DNS**，全部解析结果**必须**通过私网校验；
- **禁止**访问 `127.0.0.0/8`、`10.0.0.0/8`、`172.16.0.0/12`、`192.168.0.0/16`、`169.254.0.0/16`、`::1/128`、`fc00::/7`、`fe80::/10`、`0.0.0.0/8`；
- **必须**校验 IPv4-mapped IPv6（如 `::ffff:127.0.0.1`）；
- 连接时使用**已校验的 IP**（防 DNS rebinding）；
- 重定向**逐跳校验**，最多 3 跳；
- 出网总开关 `TOOLBIT_NETWORK_ENABLED=false` 时**全部**出网工具返回 `NETWORK_DISABLED`。

---

## cidr-calc

CIDR 子网计算（**不出网**）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cidr` | string | 二选一 | 如 `192.168.1.0/24` |
| `start` / `end` | string | 二选一 | 地址范围（反推 CIDR） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `network` | string | 网络地址 |
| `broadcast` | string | 广播地址 |
| `firstUsable` | string | 首个可用地址 |
| `lastUsable` | string | 末个可用地址 |
| `hostCount` | number | 可用主机数 |
| `mask` | string | 子网掩码 |
| `prefix` | int | 前缀长度 |
| `cidrs` | array | 拆分子网列表（按需） |

### 错误

`INVALID_CIDR`、`INVALID_FIELD`、`MISSING_FIELD`。

> IPv4 与 IPv6 **分别**计算，**不**混合。`hostCount` 对 `/31` 与 `/32` 有特殊处理（2 与 1）。

---

## ssl-check

TLS 握手与证书链检测（**流式**，需出网）。

### 传输

WebSocket：`GET /api/v1/tools/ssl-check/stream`。协议细节见 [../README.md §11](../README.md)。

### 客户端首帧

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `host` | string | 是 | 目标主机 |
| `port` | int | 否 | 默认 443 |
| `sni` | string | 否 | SNI 主机名（缺省用 `host`） |
| `depth` | int | 否 | 证书链深度上限，默认 5 |
| `versions` | array | 否 | 如 `["TLS1.3","TLS1.2"]` |
| `ciphers` | array | 否 | 密码套件名列表 |
| `timeoutMs` | int | 否 | 超时，默认 30000 |

### 服务端帧

| `type` | 载荷 |
| --- | --- |
| `progress` | `{ phase, done, total }`；`phase` ∈ `resolve` \| `connect` \| `handshake` \| `chain` |
| `result` | `{ data }`，`data` 含 `host`、`port`、`protocol`、`cipherSuite`、`negotiatedVersion`、`certificates[]`、`chainValid`、`chainErrors[]`、`verified`、`ocspStapled` |
| `error` | `{ error: { code, message, params } }` |

### 错误

`SSRF_BLOCKED`、`DNS_ERROR`、`TLS_ERROR`、`HOST_UNREACHABLE`、`UPSTREAM_TIMEOUT`、`NETWORK_DISABLED`。

### 关闭码

`1000` 正常、`1008` 输入非法或被策略拒绝、`1011` 服务端内部错误。

> 空闲 60 秒关闭；单次任务总预算 30 秒；同一 IP **同时最多 1 条**连接。

---

## dns-lookup

DNS 记录查询（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `host` | string | 是 | 域名 |
| `type` | string | 是 | `A` `AAAA` `CNAME` `MX` `TXT` `NS` `SOA` `SRV` `CAA` `PTR` `ANY` |
| `servers` | array | 否 | 自定义服务器 `[{ proto, addr }]`；`proto` ∈ `udp` \| `tcp` \| `dot` \| `doh` |
| `timeoutMs` | int | 否 | 单服务器超时，默认 5000 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `perServer` | array | `[{ server, proto, latencyMs, answers[], error? }]` |

### 错误

`DNS_ERROR`、`UPSTREAM_TIMEOUT`、`NETWORK_DISABLED`、`INVALID_FIELD`、`TOO_MANY_ITEMS`。

> 默认使用系统 resolver；指定 `servers` 时**只**查询指定服务器。DoH 服务器需提供完整 URL。

---

## dns-tester

DNS 综合检测（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain` | string | 二选一 | 域名 |
| `ip` | string | 二选一 | IP 地址（用于 PTR / whois） |
| `action` | string | 是 | `whois` \| `dnssec` \| `ptr` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `result` | object | 随 `action` 变化 |

`action` 对应的 `result`：

| `action` | `result` 结构 |
| --- | --- |
| `whois` | `{ registrar, creationDate, expirationDate, updatedDate, nameServers[], status[], registrantCountry, registries[] }` |
| `dnssec` | `{ supported, dnssecEnabled, dsPresent, sigPresent, chainOfTrust[] }` |
| `ptr` | `{ reverse, addresses[], names[] }` |

### 错误

`UPSTREAM_ERROR`、`UPSTREAM_TIMEOUT`、`DNS_ERROR`、`NETWORK_DISABLED`、`INVALID_FIELD`。

> `whois` 结果**可能不完整**（部分注册局限制查询或限流），缺失字段以 `null` 表示。

---

## websocket-test

WebSocket 连接测试（**纯前端**，无后端端点）。

### 说明

本工具在**浏览器内**使用原生 `WebSocket` API 直连目标服务器，**不经过**后端。

- catalog 中 `transport: client`；
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

## 相关文档

- [../README.md](../README.md) —— 通用信封、限流、WebSocket 协议
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/06-security.md](../../standards/06-security.md) —— 出网与 SSRF 防护
