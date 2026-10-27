# 证书 API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

覆盖 **23** 个证书类工具 —— 本项目最大的分类。端点均为 `POST /api/v1/tools/{slug}`，除 `cert-chain-fetch`（multipart + 出网）外均为 JSON。

| slug | 名称 | 档位 | 出网 |
| --- | --- | --- | --- |
| [cert-parse](#cert-parse) | 证书解析 | standard | 否 |
| [csr-gen](#csr-gen) | CSR 生成 | standard | 否 |
| [csr-parse](#csr-parse) | CSR 解析 | standard | 否 |
| [pem-der](#pem-der) | PEM/DER 转换 | light | 否 |
| [asn1-view](#asn1-view) | ASN.1 视图 | standard | 否 |
| [cert-chain-parse](#cert-chain-parse) | 证书链解析 | standard | 否 |
| [cert-chain-verify](#cert-chain-verify) | 证书链验证 | standard | 否 |
| [cert-chain-repair](#cert-chain-repair) | 证书链修复 | standard | 否 |
| [revocation-check](#revocation-check) | 吊销状态检查 | network | 否 |
| [crl-parse](#crl-parse) | CRL 解析 | standard | 否 |
| [crl-gen](#crl-gen) | CRL 生成 | heavy | 否 |
| [ocsp-request](#ocsp-request) | OCSP 请求生成 | standard | 否 |
| [ocsp-response](#ocsp-response) | OCSP 响应解析 | standard | 否 |
| [cert-gen](#cert-gen) | 证书生成 | heavy | 否 |
| [pfx-pack](#pfx-pack) | PKCS#12 打包 | heavy | 否 |
| [pfx-parse](#pfx-parse) | PKCS#12 解析 | heavy | 否 |
| [pfx-change-password](#pfx-change-password) | PKCS#12 改密 | heavy | 否 |
| [p7b-pack](#p7b-pack) | PKCS#7 打包 | standard | 否 |
| [p7b-extract](#p7b-extract) | PKCS#7 提取 | standard | 否 |
| [caa-gen](#caa-gen) | CAA 生成 | light | 否 |
| [cert-chain-fetch](#cert-chain-fetch) | 证书链获取 | network | 是 |
| [ct-search](#ct-search) | 证书透明度查询 | network | 是 |
| [domain-verify](#domain-verify) | 域名验证 | network | 是 |

---

## cert-parse

解析 X.509 证书。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `PemBundle`（自动识别 PEM / DER / base64 / hex） |
| `atTime` | string | 否 | 以该时刻判断有效性（RFC3339） |
| `fingerprints` | array | 否 | 需要计算的指纹算法，如 `["SHA256","SHA1"]` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certificate` | object | 证书主体信息（subject、issuer、serial、validity、extensions、publicKey、signatureAlgorithm…） |
| `validity` | object | `{ status, daysRemaining }`；`status` ∈ `valid` \| `expired` \| `not_yet_valid` |
| `fingerprints` | array | `[{ algorithm, value }]` |
| `selfSigned` | bool | 是否自签名 |
| `rawDerBase64` | string | 原始 DER（base64） |

### 错误

`INVALID_CERT`、`INVALID_PEM`、`INVALID_DER`、`INVALID_TIME`。

### 警告

`EXPIRED_CERT`、`NOT_YET_VALID`、`SELF_SIGNED_CERT`。

---

## csr-gen

生成 PKCS#10 证书签名请求。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `certType` | string | 否 | `server` \| `client` \| `email` \| `code-signing` |
| `subject` | object | 是 | DN 字段（`CN` `C` `O` `OU` `L` `ST` `emailAddress` 等） |
| `key` | object | 否 | `{ type, bits, curve }`；缺省生成新密钥 |
| `existingPrivateKey` | object | 否 | 使用已有私钥（`KeyInput`） |
| `sans` | array | 否 | 主体备用名 |
| `challengePassword` | string | 否 | 挑战口令 |
| `keyUsages` | array | 否 | 密钥用途 |
| `extKeyUsages` | array | 否 | 扩展密钥用途 |
| `basicConstraints` | object | 否 | `{ isCa, pathLen }` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `csrPem` | string | CSR 的 PEM |
| `privateKeyPem` | string | 新生成时返回私钥 PEM |
| `keyType` / `keyBits` / `keyCurve` | string / int / string | 密钥信息 |

### 错误

`INVALID_FIELD`、`UNSUPPORTED_ALGORITHM`、`INVALID_KEY`、`KEY_SIZE_MISMATCH`。

> 提供 `existingPrivateKey` 时**不**返回 `privateKeyPem`。DN 字段非法时**不** panic，返回 `INVALID_FIELD`。

---

## csr-parse

解析 PKCS#10 CSR。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `PemBundle` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `subject` | object | 主体 |
| `sans` | array | 主体备用名 |
| `publicKey` | object | `{ algorithm, bits, curve }` |
| `signatureAlgorithm` | string | 签名算法 |
| `extensions` | array | 请求扩展 |
| `fingerprintSha256` | string | 指纹 |

### 错误

`INVALID_CSR`、`INVALID_PEM`、`INVALID_DER`。

---

## pem-der

PEM ↔ DER 互转。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data, format }`；`format` ∈ `pem` \| `der` \| `base64` \| `hex` |
| `type` | string | 否 | `cert` \| `key` \| `csr` \| `crl` \| `any`（默认） |
| `to` | string | 是 | `pem` \| `der` |
| `label` | string | 否 | 输出 PEM 的标签 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | object | `{ data, format, encoding }` |
| `labels` | array | 识别到的 PEM 标签列表 |

### 错误

`INVALID_PEM`、`INVALID_DER`、`UNSUPPORTED_FORMAT`。

---

## asn1-view

ASN.1 结构树可视化。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ der \| base64 \| pem }` |
| `maxDepth` | int | 否 | 最大展开深度，默认 32 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `length` | int | 字节长度 |
| `tree` | object | `{ nodes: [...] }` |
| `hexDump` | string | 十六进制转储 |

### 错误

`INVALID_DER`、`TOO_LARGE`、`TOO_MANY_ITEMS`。

> 解析深度上限 64，节点数上限 10000，超出返回 `TOO_LARGE` / `TOO_MANY_ITEMS`。

---

## cert-chain-parse

解析并排序证书链（PEM bundle）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `bundle` | object | 是 | `PemBundle`（多张证书） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `count` | int | 证书数量 |
| `chain` | array | `[{ pem, info }]`（已排序，叶 → 根） |
| `ordered` | bool | 输入是否已按序排列 |
| `missingRoot` | bool | 是否缺少根证书 |
| `selfSigned` | bool | 根证书是否自签名 |

### 错误

`INVALID_CERT`、`INVALID_PEM`、`TOO_MANY_ITEMS`。

> 证书链上限 64 张，超出返回 `TOO_MANY_ITEMS`。

---

## cert-chain-verify

验证证书链。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `bundle` | object | 是 | `PemBundle` |
| `trustedRoots` | array | 否 | 受信任根证书 PEM 列表 |
| `atTime` | string | 否 | 验证时刻（RFC3339） |
| `host` | string | 否 | 校验主机名匹配 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `valid` | bool | 整链是否可信 |
| `errors` | array | `[{ code, message }]` |
| `details` | array | `[{ subject, valid, reason }]` |
| `hostnameMatch` | bool | 主机名是否匹配 |
| `hostnameError` | string | 主机名不匹配原因 |

### 错误

`INVALID_CERT`、`INVALID_TIME`、`TOO_MANY_ITEMS`。

### 警告

`EXPIRED_CERT`、`NOT_YET_VALID`、`SELF_SIGNED_CERT`、`UNTRUSTED_CHAIN`。

> 验证**不通过**是**业务结果**，返回 `200` + `valid: false`。

---

## cert-chain-repair

按缺失的中间证书补全证书链。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cert` | object | 是 | 叶证书 |
| `trustedRoots` | array | 否 | 受信任根证书 PEM 列表 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `chainPem` | array | 补全后的链（叶 → 根） |
| `addedCount` | int | 补充的证书数 |
| `addedFroms` | array | 补充来源说明 |
| `reachedRoot` | bool | 是否到达受信任根 |
| `note` | string | 备注 |

### 错误

`INVALID_CERT`、`TOO_MANY_ITEMS`。

---

## revocation-check

用 CRL 检查证书吊销状态。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cert` | object | 是 | 待检查证书 |
| `crl` | object | 是 | CRL（PEM） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | string | `good` \| `revoked` \| `unknown` |
| `revokedAt` | string | 吊销时间 |
| `reasonCode` | int | 吊销原因码 |
| `reasonText` | string | 原因文本 |

### 错误

`INVALID_CERT`、`INVALID_CRL`。

> 服务端**必须**先校验 CRL 签名与有效期，再判断吊销状态。

---

## crl-parse

解析证书吊销列表。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `PemBundle` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `issuer` | object | 签发者 |
| `thisUpdate` / `nextUpdate` | string | 有效期 |
| `signatureAlgorithm` | string | 签名算法 |
| `revoked` | array | `[{ serial, revokedAt, reasonCode, reasonText }]` |

### 错误

`INVALID_CRL`、`INVALID_PEM`、`TOO_MANY_ITEMS`。

> 吊销条目上限 10000。

---

## crl-gen

生成证书吊销列表。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `issuerCert` | object | 是 | 签发者证书（须为 CA） |
| `issuerKey` | object | 是 | `{ data, passphrase? }` |
| `number` | string | 否 | CRL 编号，默认自动生成 |
| `thisUpdate` / `nextUpdate` | string | 否 | 有效期 |
| `revoked` | array | 否 | `[{ cert \| serial, revokedAt?, reasonCode }]` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `crlPem` | string | CRL 的 PEM |
| `crlDerBase64` | string | CRL 的 DER（base64） |
| `issuer` | object | 签发者 |
| `number` | string | CRL 编号 |
| `counts` | object | `{ total, revoked }` |

### 错误

`INVALID_CERT`、`INVALID_KEY`、`KEY_MISMATCH`、`INVALID_FIELD`。

> 原因码取 RFC 5280 §5.3.1，码值 `7` 为保留值，**禁止**使用。

---

## ocsp-request

生成 OCSP 请求（DER）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cert` | object | 是 | 待查询证书 |
| `issuer` | object | 是 | 签发者证书 |
| `hash` | string | 否 | `SHA1`（默认）\| `SHA256` \| `SM3` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `hash` | string | 实际哈希算法 |
| `derBase64` | string | 请求 DER（base64） |
| `derHex` | string | 请求 DER（hex） |
| `length` | int | 字节长度 |

### 错误

`INVALID_CERT`、`KEY_MISMATCH`、`UNSUPPORTED_ALGORITHM`。

---

## ocsp-response

解析 OCSP 响应。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `response` | object | 是 | `{ auto }`（自动识别 DER / base64） |
| `cert` | object | 是 | 被查询证书 |
| `issuer` | object | 是 | 签发者证书 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | string | 响应级状态（`successful` / `malformed` / …） |
| `certStatus` | string | `good` \| `revoked` \| `unknown` |
| `producedAt` / `thisUpdate` / `nextUpdate` | string | 时间 |
| `revocationTime` | string | 吊销时间 |
| `reasonCode` | int | 吊销原因码 |
| `responderCerts` | array | 响应者证书 |

### 错误

`INVALID_CERT`、`INVALID_DER`、`INVALID_FIELD`。

---

## cert-gen

签发证书。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `signMode` | string | 是 | `self`（自签）\| `ca`（由 CA 签发） |
| `issuer` | object | `ca` 时是 | `{ cert, key, passphrase? }` |
| `subject` | object | `self` 时是 | 主体 DN |
| `csr` | object | 二选一 | 待签发的 CSR |
| `templateCert` | object | 二选一 | 模板证书（复制其扩展） |
| `subjectOverride` | object | 否 | 覆盖主体字段 |
| `key` | object | 否 | `{ type, bits, curve }` |
| `existingPrivateKey` | object | 否 | 使用已有私钥 |
| `notAfterDays` | int | 否 | 有效期天数 |
| `notAfter` | string | 否 | 绝对失效时间 |
| `serial` | string | 否 | 序列号（十六进制） |
| `sans` | array | 否 | 主体备用名 |
| `keyUsages` / `extKeyUsages` | array | 否 | 用途 |
| `basicConstraints` | object | 否 | `{ isCa, pathLen }` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certPem` | string | 证书 PEM |
| `privateKeyPem` | string | 新生成密钥时返回 |
| `chainPem` | array | CA 模式下的完整链 |
| `serial` | string | 序列号 |
| `subject` / `issuer` | object | 主体与签发者 |
| `notBefore` / `notAfter` | string | 有效期 |

### 错误

`INVALID_FIELD`、`INVALID_CERT`、`INVALID_KEY`、`KEY_MISMATCH`、`NOT_AFTER_BEFORE_NOW`、`SERIAL_INVALID`。

> `ca` 模式**必须**校验签发者私钥与证书是否配对（`KEY_MISMATCH`）。

---

## pfx-pack

打包 PKCS#12。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cert` | object | 是 | 叶证书 |
| `privateKey` | object | 否 | 私钥 |
| `chain` | array | 否 | 附加证书链 |
| `password` | string | 是 | 口令 |
| `alias` | string | 否 | 条目别名 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pfxBase64` | string | PKCS#12 内容（base64） |

### 错误

`INVALID_CERT`、`INVALID_KEY`、`KEY_MISMATCH`、`INVALID_FIELD`。

---

## pfx-parse

解析 PKCS#12。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `pfx` | object | 是 | `{ base64 }` |
| `password` | string | 是 | 口令 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certPem` | string | 主证书 |
| `chainPem` | array | 附加证书 |
| `privateKeyPem` | string | 私钥（如存在） |

### 错误

`INVALID_PKCS12`、`PASSPHRASE_INCORRECT`、`INVALID_CERT`。

---

## pfx-change-password

修改 PKCS#12 口令。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `pfx` | object | 是 | `{ base64 }` |
| `oldPassword` | string | 是 | 原口令 |
| `newPassword` | string | 是 | 新口令 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pfxBase64` | string | 改密后的内容（base64） |

### 错误

`INVALID_PKCS12`、`PASSPHRASE_INCORRECT`、`MISSING_FIELD`。

---

## p7b-pack

打包 PKCS#7（DER / PEM）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `chainPem` | array | 是 | 证书链 PEM 列表 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `p7bPem` | string | PKCS#7 PEM |

### 错误

`INVALID_CERT`、`INVALID_PKCS7`、`MISSING_FIELD`。

---

## p7b-extract

从 PKCS#7 中提取证书。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `p7b` | object | 是 | `{ auto }` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `chainPem` | array | 提取出的证书 PEM |

### 错误

`INVALID_PKCS7`、`INVALID_PEM`、`INVALID_DER`。

---

## caa-gen

生成 CAA（证书颁发机构授权）记录。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain` | string | 是 | 域名 |
| `cas` | array | 否 | `[{ flag, tag, value }]` |
| `provider` | string | 否 | `bind` \| `route53` \| `dnsmasq` \| `cloudflare` |
| `ttl` | int | 否 | TTL 秒数 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `records` | array | `[{ provider, text }]` |

### 错误

`INVALID_DOMAIN`、`INVALID_FIELD`。

---

## cert-chain-fetch

从目标主机获取证书链（需出网）。

### 请求（multipart 或 JSON）

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `host` | string | 是 | 目标主机 |
| `port` | number | 否 | 默认 443 |
| `serverName` | string | 否 | SNI |
| `timeoutMs` | int | 否 | 超时，默认 15000 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `chainPem` | string | 完整链 PEM |
| `chain` | array | `[{ pem, info }]` |
| `chainSummary` | object | 链的摘要信息 |

### 错误

`SSRF_BLOCKED`、`DNS_ERROR`、`TLS_ERROR`、`HOST_UNREACHABLE`、`UPSTREAM_TIMEOUT`、`NETWORK_DISABLED`。

> 私网 / 环回 / 链路本地地址**一律**被拦截（`SSRF_BLOCKED`）。

---

## ct-search

查询证书透明度日志（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain` | string | 是 | 域名 |
| `matchType` | string | 否 | `exact`（默认）\| `contains` \| `wildcard` |
| `limit` | int | 否 | 返回条数，默认 20 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certs` | array | `[{ id, notBefore, notAfter, cn, issuer, serial }]` |

### 错误

`UPSTREAM_ERROR`、`UPSTREAM_TIMEOUT`、`NETWORK_DISABLED`、`INVALID_DOMAIN`。

---

## domain-verify

域名所有权验证（需出网）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `type` | string | 是 | `dns-txt` \| `dns-cname` \| `http-file` |
| `domain` | string | 是 | 域名 |
| `expected` | string | 是 | 期望值 |
| `filePath` | string | `http-file` 时是 | 文件路径（相对站点根） |
| `resolvers` | array | 否 | 自定义 DNS 服务器 |
| `timeoutMs` | int | 否 | 超时 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `matched` | bool | 是否匹配 |
| `perResolver` | array | 每个解析器的结果 |
| `httpStatus` | number | HTTP 方式的状态码 |
| `bodyPreview` | string | 响应体预览（截断） |

### 错误

`INVALID_DOMAIN`、`DNS_ERROR`、`UPSTREAM_ERROR`、`UPSTREAM_TIMEOUT`、`NETWORK_DISABLED`。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/03-crypto.md](../../standards/03-crypto.md) —— 密码学与 PKI 规范
