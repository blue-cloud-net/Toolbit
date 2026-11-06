# 网络 API

> 状态：Stable ｜ 适用版本：v0.11.0 ｜ 最后更新：2026-11-06

覆盖 **13** 个网络类工具。其中 6 个**需要出网**，`websocket-test` 为**纯前端**工具（无后端端点），
`mime-lookup`、`http-status`、`dns-servers` 与 `common-ports` 为**纯查表**工具（不出网）。

| slug | 名称 | 传输 | 端点 |
| --- | --- | --- | --- |
| [cidr-calc](#cidr-calc) | CIDR 计算 | JSON | `POST /api/v1/tools/cidr-calc` |
| [ssl-check](#ssl-check) | SSL 证书检测 | WebSocket | `GET /api/v1/tools/ssl-check/stream` |
| [dns-lookup](#dns-lookup) | DNS 查询 | JSON | `POST /api/v1/tools/dns-lookup` |
| [dns-servers](#dns-servers) | DNS 服务器大全 | JSON | `POST /api/v1/tools/dns-servers` |
| [dns-whois](#dns-whois) | WHOIS 查询 | JSON | `POST /api/v1/tools/dns-whois` |
| [dnssec-check](#dnssec-check) | DNSSEC 检测 | JSON | `POST /api/v1/tools/dnssec-check` |
| [dns-ptr](#dns-ptr) | PTR 反向解析 | JSON | `POST /api/v1/tools/dns-ptr` |
| [common-ports](#common-ports) | 常用端口查询 | JSON | `POST /api/v1/tools/common-ports` |
| [websocket-test](#websocket-test) | WebSocket 测试 | 浏览器原生 | 无 |
| [mime-lookup](#mime-lookup) | MIME 类型查询 | JSON | `POST /api/v1/tools/mime-lookup` |
| [http-status](#http-status) | HTTP 状态码查询 | JSON | `POST /api/v1/tools/http-status` |
| [ua-header-parse](#ua-header-parse) | UA 头解析 | JSON | `POST /api/v1/tools/ua-header-parse` |
| [ua-header-gen](#ua-header-gen) | UA 头生成 | JSON | `POST /api/v1/tools/ua-header-gen` |

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
| `expand` | boolean | 否 | 是否把地址区间展开成逐个地址列表；默认 `false` |
| `limit` | number | 否 | 展开条数上限；默认 `1024`，最大 `4096` |

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
| `addresses` | array | 展开得到的逐个地址；仅 `expand` 为 `true` 时非空 |
| `addressCount` | string | 待展开范围内的地址总数（十进制字符串） |
| `truncated` | boolean | 地址列表是否因超出 `limit` 而被截断 |

### 错误

`INVALID_CIDR`、`INVALID_FIELD`、`MISSING_FIELD`、`TOO_MANY_ITEMS`。

> IPv4 与 IPv6 **分别**计算，**不**混合。
> `hostCount`、`totalAddresses` 与 `addressCount` 用字符串表达：IPv6 的地址数量可达 2¹²⁸，超出 64 位整数。
> IPv4 的 `/31` 计 2 个可用地址、`/32` 计 1 个；IPv6 无广播概念，全部地址可用。
> 一次拆分或反推最多产出 **256** 个子网，超出返回 `TOO_MANY_ITEMS`。
> `expand` 的展开区间：`cidr` 模式取网段的**全部**地址（含网络地址与广播地址），
> 范围模式取 `start`..`end` 闭区间；`addressCount` 给出该区间的地址总数，与 `limit` 无关。

---

## ssl-check

多轮 TLS 握手扫描：协议 / 套件矩阵、协议特性、客户端握手模拟、证书链全量分析与
四象限评级（**流式**，需出网）。

> 设计细节见 [../../tools/ssl-check-refactor.md](../../tools/ssl-check-refactor.md)。

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
| `mode` | string | 否 | `quick` \| `full`，默认 `full` |
| `checks` | array | 否 | 要执行的检查项，缺省按 `mode` 推导 |
| `clients` | array | 否 | 要模拟的客户端画像 id，缺省跑全部 |
| `budgetMs` | int | 否 | 总预算，默认 `60000`，上限 `180000` |

`checks` 合法取值：`protocols` `ciphers` `features` `clients` `cert` `revocation` `caa` `hsts`。

`mode: quick` 等价于 `checks: ["protocols", "cert"]`。

### 服务端帧

| `type` | 载荷 |
| --- | --- |
| `progress` | `{ phase, done, total, item? }`；`phase` 见下表；`total` 按启用的检查项动态计算 |
| `result` | `{ data, warnings? }`，`data` 见下表 |
| `error` | `{ error: { code, message, params } }` |

`phase` 取值与顺序：

```text
resolve → connect → handshake → certificate → protocols → ciphers
        → features → clients → revocation → grade
```

`item` 只在 `ciphers` / `clients` 阶段出现，表示当前正在处理的套件名或画像 id。

`result.data` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `endpoint` | object | `{ host, port, sni, addr, elapsedMs }` |
| `negotiated` | object | `{ version, versionWire, cipher, cipherId, keyExchangeGroup, peerSignatureAlgorithm, alpn }` |
| `certificate` | object | 叶证书摘要，见下 |
| `chain` | object | 证书链，**与 `cert-chain-parse` 的响应同构**（`count` / `chain[]` / `ordered` / `missingRoot` / `selfSigned`） |
| `protocols` | array | `[{ version, versionWire, supported, negotiatedCipher?, reason? }]` |
| `ciphers` | array | `[{ version, name, id, family, bits, kex, verdict }]` |
| `features` | array | `[{ id, label, status, detail? }]` |
| `clients` | array | `[{ id, vendor, os, engine, profileVersion, supported, reason?, negotiatedVersion?, cipher?, alpn? }]` |
| `caa` | object | `{ queried, records[], permitted, violations[], error? }` |
| `revocation` | object | `{ ocsp[], crl[] }`，每项 `{ url, status, thisUpdate?, nextUpdate?, error? }` |
| `hsts` | object | `{ present, maxAge, includeSubDomains, preload, error? }` |
| `grade` | object | `{ letter, score, categories, caps[], issues[] }` |

`certificate` 字段：

| 字段 | 说明 |
| --- | --- |
| `subject` / `issuer` | 单行 DN 文本 |
| `serial` | 大写十六进制 |
| `notBefore` / `notAfter` | RFC 3339 |
| `daysRemaining` | 剩余天数（小数） |
| `keyAlgorithm` / `keyBits` / `curve` | 公钥画像 |
| `signatureAlgorithm` | 证书签名算法 |
| `sanCount` / `sanMatchesHost` | SAN 覆盖情况 |
| `isCa` / `selfSigned` | 布尔 |
| `sctCount` / `mustStaple` | 透明度与装订声明 |
| `chainValid` / `verified` / `implicitTrust` | 链验证结论 |
| `chainErrors[]` | 链验证错误码 |

#### 稳定标识

`features[].id`、`clients[].id`、`grade.issues[].id` 与 `grade.caps[]` 是**稳定标识**，
服务端**不**下发本地化文案，前端按 key 本地化。

`features[].id`：`alpnH2` `sessionResumption` `ocspStapling` `secureRenegotiation`
`compression` `mustStaple`。

`features[].status` ∈ `supported` | `not_supported` | `unavailable`。
**`unavailable` 与 `not_supported` 是不同结论**：前者表示当前 TLS 引擎拿不到该信息，
不参与计分，也**不**触发档位上限。

`ciphers[].verdict` ∈ `supported` | `rejected` | `unavailable`。

`clients[].reason` ∈ `noCommonVersion` `noCommonCipher` `noCommonGroup`
`handshakeFailure` `network` `timeout`。

> 归因靠**逐步放宽**的对照握手（版本 → 套件 → 组）得出，**不**匹配错误文本，
> 也不依赖上游的握手错误分类（上游在错误线程上取原生错误队列，见
> [../../upstream/tongsuo-go-issues.md](../../upstream/tongsuo-go-issues.md) I-28）。

#### 评级规则

四象限各 100 分，**互不抵扣**：

| 象限 | 计分要点 |
| --- | --- |
| `certificate` | 链可信 / 主机名匹配为前提；剩余有效期 ≥30 天、密钥强度、签名算法强度各扣分 |
| `protocols` | 最高支持档：TLS 1.3 = 100、TLS 1.2 = 90、TLS 1.1 / 1.0 = 20 |
| `keyExchange` | ECDHE/X25519 = 100；RSA 密钥交换 = 20；HSTS 命中 +5（封顶 100） |
| `cipherStrength` | AEAD = 100；CBC = 80；RC4 / 3DES = 20 |

综合分 = 四象限平均分。档位上限（取最严的一个）：

| 条件 | 上限 |
| --- | --- |
| 证书不可信 / 已过期 | `T` |
| 主机名不匹配 | `M` |
| 证书已吊销 | `F` |
| 只支持 TLS 1.0 / 1.1 | `B` |
| 不支持 TLS 1.3 | `A` |

最终等级 = min(分数映射等级, 上限等级)。等级序列：
`A+` > `A` > `A-` > `B` > `C` > `D` > `E` > `F` > `M` > `T`。

> 评级只描述**本次扫描的观测结果**，不代表站点整体安全评分，也**不**与 SSL Labs 的分数对齐。

#### 不覆盖的能力

以下项在本工具中**不可测**，相关检查项一律返回 `unavailable`：

| 能力 | 原因 |
| --- | --- |
| SSLv2 / SSLv3 | Tongsuo 基于 OpenSSL 3.x 兼容层，协议已被移除 |
| OCSP 装订响应 | 上游未暴露装订原文（`tongsuo-go` I-23）；吊销结论改用**主动 OCSP 查询** |
| 会话恢复 | 上游未暴露会话复用标志（I-25） |
| 安全重协商 / TLS 压缩 | 上游未暴露对应状态（I-26） |
| SCT（TLS 扩展形式） | 上游未暴露扩展原文（I-24）；证书内嵌 SCT 仍可读 |
| 握手失败的具体 alert 码 | 上游在错误线程上取原生错误队列（I-28），归因改用对照握手 |
| Heartbleed / ROBOT / CCS / POODLE / Ticketbleed | 需要构造畸形握手报文；见 [../../upstream/tongsuo-go-issues.md](../../upstream/tongsuo-go-issues.md) I-27 |

> 握手模拟是**参数级近似**，不是 ClientHello 字节回放：扩展顺序、GREASE 与 TLS 指纹
> 均不一致。结论的准确含义是「以该客户端画像的参数发起握手，服务端是否接受」。

### 错误

`SSRF_BLOCKED`、`DNS_ERROR`、`TLS_ERROR`、`HOST_UNREACHABLE`、`UPSTREAM_TIMEOUT`、
`NETWORK_DISABLED`、`INVALID_FIELD`、`INVALID_DOMAIN`。

### 关闭码

`1000` 正常、`1008` 输入非法或被策略拒绝、`1011` 服务端内部错误、`1013` 预算耗尽。

### 警告

链不通过或缺少系统信任根时给 `UNTRUSTED_CHAIN`；叶证书自签名给 `SELF_SIGNED_CERT`；
过期给 `EXPIRED_CERT`。

> 探测**不**在握手期校验链（否则拿不到无效证书的细节），而是拿到链之后另行验证。
> 验证时优先加载本地信任库（`/etc/ssl/certs/ca-certificates.crt` 等）；信任库不可用时
> `verified` 为 `false`、`implicitTrust` 为 `true`。
> 所有建连统一走出网防护并**钉住**已校验的 IP；证书里的 AIA / CDP 地址同样要过防护。
> 空闲 60 秒关闭；单次任务总预算 180 秒；同一 IP **同时最多 1 条**连接。

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
> 各台服务器**并发**查询（上限 4 路），结果按下标回填、保持请求顺序；
> 8 台服务器可在工具 20 秒的预算内完成。
> 查询**通告 EDNS0 缓冲**（4096 字节）：不通告时上游按历史上的 512 字节上限截断应答，
> 多条 A/AAAA 与长 TXT 会被切掉；UDP 应答仍被截断时再**回退 TCP** 取一次完整结果。
> 前端预设服务器分五组：**主用**（6 个）、**备用**（7 个）、**TCP 通道**（5 个）、
> **DoT 加密**（6 个）与 **DoH 加密**（6 个 https 端点）。
> 结果分三类页签：**汇总**（成败台数、最快最慢、TTL 区间、一致性结论）、
> **全部服务器**（逐台列出地址、协议、耗时、状态与结果摘要）、**单台服务器**（该台的记录明细）。
> 「记录一致性」比较各台**查询成功**服务器的记录集合，结论有三态：
> **一致**、**存在差异**、**不适用**（成功服务器不足两台，没有可比对象）。
> DoH 需要完整的 `https://` 地址；DoT 固定 `853` 端口。
> `ANY` 查询可能被上游服务器拒绝，部分服务商已不再支持。

---

## dns-servers

DNS 服务器目录（**纯查表，不出网**）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `query` | string | 否 | 搜索词，匹配服务商名（中英文）、地址与说明；最多 253 字符 |
| `region` | string | 否 | 地区过滤：`china` \| `asia` \| `global` |
| `proto` | string | 否 | 协议过滤：`udp` \| `tcp` \| `dot` \| `doh` |
| `family` | string | 否 | 地址族过滤：`ipv4` \| `ipv6` \| `domain` |

四个字段都可以省略；都不给时返回**全量**目录。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `servers` | array | `[{ provider, providerEn, region, family, proto, addr, note }]` |
| `total` | number | 命中的端点数 |
| `providers` | number | 命中的服务商数 |
| `dataVersion` | string | 数据集版本（当前 `2026-11`） |

`servers[]` 的每一项是**一个可复制的端点**：同一个 IP 会有 `udp` 与 `tcp` 两条，
因为它们是两条独立的查询通道。

### 错误

`INVALID_FIELD`：`region` / `proto` / `family` 取值非法，或 `query` 超过 253 字符。

### 前端展示

接口契约与展示形态**无关**：同一份 `servers[]` 可渲染成两种视图。

| 视图 | 形态 | 用途 |
| --- | --- | --- |
| 卡片（默认） | 一家服务商一张卡，卡内按地址族列出端点，同一地址的 `udp` / `tcp` 合并成一行 | 浏览、挑选、单个复制 |
| 列表 | 一条端点一行，列与 `servers[]` 的字段一一对应 | 按列比对、整行复制 |

两种视图共用同一套筛选条件（`query` / `region` / `proto` / `family`），
切换只改变呈现方式，不重新请求接口。

> 目录是**内置快照**，运行期不联网，也不做实时可达性检测 ——
> 想验证某台服务器能否解析，请用 [dns-lookup](#dns-lookup) 实测。
> 目录里的每个端点都**可直接**作为 `dns-lookup` 的 `servers[]` 条目使用。

---

## dns-whois

域名 / IP 注册信息查询（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain` | string | 二选一 | 域名 |
| `ip` | string | 二选一 | IP 地址；给出时优先于 `domain` |
| `timeoutMs` | int | 否 | 超时，默认 `20000` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 实际查询的目标：域名或 IP |
| `whois` | object \| null | 注册信息；结构见下 |

`whois` 的结构：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `source` | string | 数据来源：`rdap` 或 `whois` |
| `registrar` | string | 注册商 |
| `creationDate` | string | 注册时间 |
| `expirationDate` | string | 到期时间 |
| `updatedDate` | string | 最后更新时间 |
| `nameServers` | array | 名称服务器 |
| `status` | array | 状态码列表 |
| `registrantCountry` | string | 注册人国家 / 地区 |
| `registries` | array | 查询过的注册局 / 注册商主机 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`、`INVALID_DOMAIN`、`UPSTREAM_ERROR`、
`UPSTREAM_TIMEOUT`、`SSRF_BLOCKED`、`NETWORK_DISABLED`。

> 先走 RDAP（`https://rdap.org`，HTTPS 443），失败后回退到 whois 协议
> （`whois.iana.org:43`，按 `refer:` 最多再跳 1 次）。端口由服务端固定，用户无法指定。
> 部分注册局限制查询或脱敏，**缺失字段以空串表示**。

---

## dnssec-check

域名 DNSSEC 部署检测（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain` | string | 是 | 要检查的域名 |
| `timeoutMs` | int | 否 | 超时，默认 `20000` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `domain` | string | 被检查的域名 |
| `dnssec` | object \| null | 检查结果；结构见下 |

`dnssec` 的结构：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `supported` | bool | 该域名是否可以启用 DNSSEC |
| `dnssecEnabled` | bool | 当前是否已启用（存在 DS 或 DNSKEY） |
| `dsPresent` | bool | 是否存在 DS 记录 |
| `sigPresent` | bool | 响应中是否带 RRSIG |
| `chainOfTrust` | array | 逐级 DS 的结果，形如 `example.com: DS 存在` |

### 错误

`MISSING_FIELD`、`INVALID_DOMAIN`、`DNS_ERROR`、`UPSTREAM_ERROR`、`UPSTREAM_TIMEOUT`、
`SSRF_BLOCKED`、`NETWORK_DISABLED`。

> 逐级向上查 DS（最多 3 级），并结合 DNSKEY 判断是否已启用。

---

## dns-ptr

PTR 反向解析（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain` | string | 二选一 | 域名；只给域名时先正向解析取第一个地址再反查 |
| `ip` | string | 二选一 | IP 地址；给出时优先于 `domain` |
| `timeoutMs` | int | 否 | 超时，默认 `20000` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 实际查询的目标：域名或 IP |
| `ptr` | object \| null | 反向解析结果；结构见下 |

`ptr` 的结构：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `reverse` | string | 查询的反向域名 |
| `addresses` | array | 正向解析到的地址 |
| `names` | array | 反向解析到的名字 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`、`INVALID_DOMAIN`、`DNS_ERROR`、`UPSTREAM_ERROR`、
`UPSTREAM_TIMEOUT`、`SSRF_BLOCKED`、`NETWORK_DISABLED`。

> 反查出的名字会再做一次正向解析，用于暴露「反向与正向不一致」。

---

## common-ports

常用端口查询（**纯查表，不出网**）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `query` | string | 否 | 查询词：端口号（如 `443`）或关键字（如 `https`）；缺省返回全量 |
| `category` | string | 否 | 服务分类过滤；为空表示不过滤 |

`category` 取值：`web` `mail` `file` `remote` `database` `messaging` `directory`
`network` `devops`。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 规范化后的查询词 |
| `kind` | string | 查询方式：`all` \| `port` \| `keyword` |
| `matched` | bool | 是否命中端口条目 |
| `port` | number | 命中的端口号；未命中为 `0` |
| `label` | string | 命中的服务名称 |
| `description` | string | 命中的中文说明 |
| `protocols` | array | 命中的传输协议 |
| `category` | string | 命中的服务分类 |
| `matches` | array | 命中的条目列表 |
| `total` | number | 命中条数 |
| `datasetTotal` | number | 数据集条目数 |
| `dataVersion` | string | 数据集版本（当前 `2026-11`） |

`matches[]` 的每一项是 `{ port, label, description, protocols[], category }`。

### 错误

`INVALID_FIELD`：`category` 取值非法，或 `query` 超过 253 字符。

> 端口条目补上了服务分类（Web、邮件、文件传输、远程登录、数据库与缓存等 9 类）；
> 页面**键入即筛选**，搜索词与分类过滤在前端本地完成，无需点击「运行」。
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
| `query` | string | 否 | 查询词：媒体类型（`image/png`）、扩展名（`png`、`.png`、`report.png`）或用途说明里的关键词（如 `图片`）；空表示**返回全量** |
| `category` | string | 否 | 顶层类型过滤：`application` \| `audio` \| `font` \| `image` \| `message` \| `model` \| `multipart` \| `text` \| `video`；空表示不过滤 |

> 两个字段都可省略：都不给时返回**全量**数据集（119 条），供前端默认展示完整列表。
> `category` 与 `query` 可叠加。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 规范化后的查询词；全量列表时为空串 |
| `category` | string | 生效的顶层类型过滤；未过滤时为空串 |
| `entries` | array | 命中的条目 `[{ mimeType, extensions, category, isTextual, charset, description, aliases }]`；未命中时为空数组 |
| `total` | number | 命中条数 |
| `dataVersion` | string | 数据集版本 |

### 错误

`INVALID_FIELD`。

> 数据来自**内置快照**，**不**读取系统 MIME 数据库（各发行版映射表不一致），**不**联网。
> 查询词归一化：媒体类型去掉 `;` 之后的参数并转小写；文件名取扩展名（`report.PNG` → `png`）。
> 匹配范围：媒体类型、扩展名、历史别名、用途说明与顶层类型；精确命中（媒体类型 / 别名 / 扩展名）排在最前。
> 扩展名按「先到先得」命中：同一扩展名被多个类型描述时，返回数据集中靠前的那一个。
> 历史别名（如 `image/x-png`）会归一到规范类型，响应里的 `mimeType` 是规范值。
> 未收录**不是**错误：返回 `200` 与空 `entries`。

---

## http-status

HTTP 状态码查询（**不出网**）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `code` | number | 否 | 状态码，`100`–`599` 的整数 |
| `query` | string | 否 | 关键字；纯数字时按状态码处理；空表示**返回全量** |
| `category` | string | 否 | 类别过滤：`informational` \| `success` \| `redirect` \| `clientError` \| `serverError`；空表示不过滤 |

> `code` 与 `query` **互斥**：同时给出返回 `INVALID_FIELD`；都不给出时返回**全量**列表（63 条）。
> `category` 可与任一方式叠加。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `query` | string | 规范化后的查询词 |
| `kind` | string | 查询方式：`code` \| `keyword` \| `all` |
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

`INVALID_FIELD`。

> 关键字匹配**原因短语**（不区分大小写）、**中文说明**与**类别中文名**，
> 因此 `not found`、`未找到`、`重定向` 都能命中。
> 落在 `100`–`599` 但未收录的状态码（如 `299`）**不是**错误：返回 `200`、`matched: false`，
> 并给出所属 `category`。
> 越界的状态码（如 `999`）返回 `INVALID_FIELD`，错误**不回显**用户输入。

---

## ua-header-parse

把 User-Agent 请求头拆成结构化字段（UA Header Parser）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `userAgent` | string | 是 | 待解析的 User-Agent 串；长度上限 1024 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `userAgent` | string | 规范化后的 UA 串（两端去空白） |
| `browser` | string | 浏览器名称；无法识别时为 `unknown` |
| `browserVersion` | string | 浏览器版本 |
| `os` | string | 操作系统名称；无法识别时为 `unknown` |
| `osVersion` | string | 操作系统版本 |
| `deviceType` | string | 设备类型：`desktop` / `mobile` / `tablet` / `bot` / `unknown` |
| `isBot` | bool | 是否是爬虫 |
| `engine` | string | 渲染引擎：`Gecko` / `Blink/WebKit` / 空串 |

### 错误

`MISSING_FIELD`（`userAgent` 为空）、`INVALID_FIELD`（`userAgent` 超过 1024 字符）。

> 识别规则是**内置的关键词与前缀匹配**：不查在线 UA 数据库，也不做指纹追踪。
> iOS 的 UA 串含 `like Mac OS X`，因此 iPhone / iPad 优先于 macOS 判定。

---

## ua-header-gen

按平台与版本生成一组客户端请求头（UA Header Generator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `platform` | string | 否 | `chrome` `firefox` `safari` `edge` `android` `ios` `curl`；默认 `chrome` |
| `version` | string | 否 | 浏览器版本号；留空用平台默认值 |
| `os` | string | 否 | 操作系统名称与版本；留空用平台默认值 |
| `acceptLanguage` | string | 否 | Accept-Language 头；留空用默认值 |
| `referer` | string | 否 | Referer 头；留空则不输出该头 |
| `customHeaders` | object | 否 | 额外请求头，键值对；至多 20 条，名称须是合法的 RFC 9110 token |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `headers` | array | 请求头列表 `[{ name, value }]` |
| `userAgent` | string | 生成的 UA 串 |

### 错误

`INVALID_FIELD`（`platform` 取值非法、`customHeaders` 名称非法或超出上限、各文本字段超长）。

> 生成的 UA 是**常见组合**，不代表任何真实浏览器。
> 自定义请求头按名称字典序输出，顺序稳定。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封、限流、WebSocket 协议
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/06-security.md](../../standards/06-security.md) —— 出网与 SSRF 防护
