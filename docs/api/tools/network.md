# 网络 API

> 状态：Stable ｜ 适用版本：v0.7.0 ｜ 最后更新：2026-10-29

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

## 相关文档

- [../README.md](../README.md) —— 通用信封、限流、WebSocket 协议
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/06-security.md](../../standards/06-security.md) —— 出网与 SSRF 防护
