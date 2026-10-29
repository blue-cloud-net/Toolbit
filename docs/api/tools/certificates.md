# 证书 API

> 状态：Draft ｜ 适用版本：v0.5.0 ｜ 最后更新：2026-10-29

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
| `atTime` | string | 否 | 以该时刻判断有效性（RFC3339）；为空时按当前时间 |
| `fingerprints` | array | 否 | 需要计算的指纹算法；为空时只算 `["SHA256"]` |

`fingerprints` 的合法取值：`MD5`、`SHA1`、`SHA256`、`SHA384`、`SHA512`、`SM3`（大小写不敏感，输出统一小写算法名）。
弱算法（`MD5`、`SHA1`）**不拒绝**，只附 `WEAK_HASH` 警告。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certificate` | object | 证书主体信息（见下表） |
| `validity` | object | `{ status, daysRemaining }`；`status` ∈ `valid` \| `expired` \| `not_yet_valid` |
| `fingerprints` | array | `[{ algorithm, value }]` |
| `selfSigned` | bool | 是否自签名 |
| `rawDerBase64` | string | 原始 DER（base64） |

`certificate` 的字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `version` | int | 版本号（2 表示 v3） |
| `subject` / `issuer` | array | RDN 条目 `[{ field, value }]`（`field` 短名如 `CN`、`O`） |
| `subjectText` / `issuerText` | string | OpenSSL 风格单行文本，如 `/C=CN/O=Toolbit Test/CN=leaf.example.com` |
| `serial` | string | 序列号（大写十六进制） |
| `notBefore` / `notAfter` | string | 生效 / 过期时间（RFC3339） |
| `publicKey` | object | `{ algorithm, bits, curve }` |
| `signatureAlgorithm` / `signatureAlgorithmOid` | string | 签名算法短名 / 点分 OID |
| `isCa` | bool | 是否为 CA 证书 |
| `pathLen` | int | 路径长度约束；非 CA 或无该扩展为 `-1` |
| `san` | array | 主体备用名条目 |
| `keyUsage` / `extendedKeyUsage` | array | 密钥用途 / 扩展密钥用途 |
| `subjectKeyId` / `authorityKeyId` | string | 主题 / 签发者密钥标识符（大写十六进制） |
| `subjectHash` / `issuerHash` | string | 名称的 `X509_NAME_hash` |
| `extensions` | array | `[{ nid, oid, name, critical, value }]` |
| `selfSigned` | bool | 是否自签名 |

### 错误

`INVALID_CERT`、`INVALID_PEM`、`INVALID_DER`、`INVALID_TIME`、`UNSUPPORTED_ALGORITHM`、`TOO_LARGE`。

### 警告

`EXPIRED_CERT`、`NOT_YET_VALID`、`SELF_SIGNED_CERT`、`WEAK_HASH`。

---

## csr-gen

生成 PKCS#10 证书签名请求。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `certType` | string | 否 | `server`（默认）\| `client` \| `email` \| `code-signing`；决定默认的 KU / EKU |
| `subject` | object | 是 | DN 字段（`CN` `C` `ST` `L` `O` `OU` `serialNumber` `emailAddress` `title` `DC`）；空字段被跳过 |
| `key` | object | 否 | `{ type, bits, curve }`；缺省 `RSA` 2048 |
| `existingPrivateKey` | object | 否 | 使用已有私钥（`KeyInput`）；提供时**不**返回私钥 |
| `sans` | array | 否 | 主体备用名（`DNS:…` / `IP:…`） |
| `challengePassword` | string | 否 | 挑战口令 |
| `keyUsages` / `extKeyUsages` | array | 否 | 覆盖按 `certType` 推出的默认值 |
| `basicConstraints` | object | 否 | `{ isCa, pathLen }` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `csrPem` | string | CSR 的 PEM |
| `privateKeyPem` | string | 新生成时的 PKCS#8 私钥；使用已有私钥时为空 |
| `keyType` / `keyBits` / `keyCurve` | string / int / string | 密钥参数 |
| `info` | object | `{ subjectText, sans, publicKey, signatureAlgorithm, challengePasswordPresent, signatureValid, fingerprintSha256 }` |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`、`UNSUPPORTED_ALGORITHM`、`UNSUPPORTED_FORMAT`、`INVALID_KEY`。

> `challengePasswordPresent` **只报存在性**，口令值绝不回显（见 AGENTS.md §4.5）。
> 已有私钥仅接受 PEM：DER 形式返回 `UNSUPPORTED_FORMAT`。

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
| `subject` | array | 主体 RDN 条目 `[{ field, value }]` |
| `subjectText` | string | 主体单行文本（OpenSSL 风格） |
| `sans` | array | 主体备用名条目 |
| `publicKey` | object | `{ algorithm, bits, curve }` |
| `signatureAlgorithm` / `signatureAlgorithmOid` | string | 签名算法短名 / 点分 OID |
| `extensions` | array | 请求扩展 `[{ nid, oid, name, critical, value }]` |
| `fingerprintSha256` | string | CSR 的 SHA-256 指纹 |
| `challengePasswordPresent` | bool | 是否携带挑战口令（**只报存在性，绝不回显口令值**） |
| `signatureValid` | bool | CSR 自签名是否校验通过 |

### 错误

`INVALID_CSR`、`INVALID_PEM`、`INVALID_DER`。

> CSR 的 SAN 由扩展手工解析（上游未暴露 `SAN()` 方法，见上游缺口登记）。

---

## pem-der

PEM ↔ DER 互转。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ data, format }`；`format` ∈ `auto` \| `pem` \| `der` \| `base64` \| `hex` |
| `type` | string | 否 | `cert` \| `key` \| `csr` \| `crl` \| `any`（默认 `any`） |
| `to` | string | 是 | `pem` \| `der` |
| `label` | string | 否 | 输出 PEM 的标签；为空时沿用源标签 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | object | `{ data, format, encoding }`；`data` 在 `to: der` 时是 base64，`encoding` 为 `base64` / `utf8` |
| `labels` | array | 输入中识别到的 PEM 标签列表 |
| `blockCount` | int | 输出块数 |

### 错误

`INVALID_PEM`、`INVALID_DER`、`UNSUPPORTED_FORMAT`、`MISSING_FIELD`。

> `to: der` 时**只允许一个块**：DER 装不下多个对象。单个输入的 PEM 块上限 64。

---

## asn1-view

ASN.1 结构树可视化。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `PemBundle`（PEM 会先剥掉外壳） |
| `maxDepth` | int | 否 | 最大展开深度；`0` 或省略表示默认 32，上限 64 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `length` | int | 输入字节长度 |
| `tree` | object | 根节点（见下表） |
| `nodeCount` | int | 解析出的节点总数 |
| `depth` | int | 实际展开到的深度 |
| `maxDepth` | int | 生效的深度上限 |
| `hexDump` | string | 十六进制转储 |

`tree` 的节点字段：`offset`（tag 字节的绝对偏移）、`tag`、`class`（0 通用 / 1 应用 / 2 上下文 / 3 私有）、`number`（tag 编号）、`constructed`、`length`、`value`（十六进制）、`truncated`（因超出深度未展开）、`children`（子节点数组）。

### 错误

`INVALID_DER`、`INVALID_PEM`、`TOO_LARGE`、`TOO_MANY_ITEMS`。

> 限额：请求体 **1 MB**；解析深度上限 64、节点数上限 10000，超出返回 `TOO_LARGE` / `TOO_MANY_ITEMS`。

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

`info` 的字段与 [cert-parse](#cert-parse) 的 `certificate` 相同。

### 错误

`INVALID_CERT`、`INVALID_PEM`、`TOO_MANY_ITEMS`。

> 证书链上限 64 张，超出返回 `TOO_MANY_ITEMS`。
> 排序优先按 **AKID ↔ SKID** 匹配，退回主体 / 签发者文本；同一层有多个候选时按输入顺序取第一个（稳定排序）。
> `ordered: false` **只**表示输入需要重排，**不是**错误。

---

## cert-chain-verify

验证证书链。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `bundle` | object | 是 | `PemBundle` |
| `trustedRoots` | array | 否 | 受信任根证书列表（每项一个 `PemBundle`）；为空时用链内的自签名根作隐含信任锚 |
| `atTime` | string | 否 | 验证时刻（RFC3339）；为空时按当前时间 |
| `host` | string | 否 | 校验主机名匹配；为空时不做名字匹配 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `valid` | bool | 整链是否可信 |
| `errors` | array | `[{ code, message, depth }]`；`code` 是 `X509_V_ERR_*` 名，`depth` 为出错深度（0 为待验证证书本身） |
| `details` | array | `[{ subject, issuer, valid, reason }]` |
| `hostnameMatch` | bool | 主机名是否匹配 |
| `hostnameError` | string | 主机名不匹配原因 |
| `implicitTrust` | bool | 是否使用了「链内自签名根」作为隐含信任锚 |

### 错误

`INVALID_CERT`、`INVALID_PEM`、`INVALID_TIME`、`TOO_MANY_ITEMS`。

### 警告

`EXPIRED_CERT`、`NOT_YET_VALID`、`SELF_SIGNED_CERT`、`UNTRUSTED_CHAIN`。

> 验证**不通过**是**业务结果**，返回 `200` + `valid: false`。
> **没有系统信任库可用**（上游未暴露默认信任库，见上游缺口登记）：`trustedRoots` 为空时把链内的自签名根当作隐含信任锚，此时 `valid: true` **只**说明链内自洽，不代表被外部信任；`implicitTrust` 为 `true` 即表示这种情况。
> 主机名匹配**只**比对名字，不参与链可信判定。

---

## cert-chain-repair

按缺失的中间证书补全证书链。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cert` | object | 是 | 叶证书 |
| `intermediates` | array | 否 | 可用于补链的中间证书列表（每项一个 `PemBundle`）；**v0.5.0 新增的可选字段** |
| `trustedRoots` | array | 否 | 受信任根证书列表（每项一个 `PemBundle`）；同样参与补链 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `chainPem` | array | 补全后的链（叶 → 根） |
| `addedCount` | int | 补充的证书数 |
| `addedFroms` | array | 每张补充证书的来源（`intermediates` / `trustedRoots`） |
| `reachedRoot` | bool | 是否到达自签名根或受信任根 |
| `note` | string | 结论枚举：`reached_root` \| `already_complete` \| `missing_intermediate` |

### 错误

`INVALID_CERT`、`INVALID_PEM`、`TOO_MANY_ITEMS`。

> **纯本地**：只从 `intermediates` 与 `trustedRoots` 里挑，**不**联网下载（出网能力在 v0.7.0 之后）。

---

## revocation-check

用 CRL 检查证书吊销状态。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cert` | object | 是 | 待检查证书 |
| `crl` | object | 是 | CRL（`PemBundle`） |
| `issuer` | object | 否 | 签发者证书；提供时**先验签再判定** |
| `atTime` | string | 否 | 判定时刻（RFC3339）；为空时按当前时间 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | string | `good` \| `revoked` \| `unknown` \| `unverified` |
| `signatureVerified` | bool | CRL 签名是否已验证 |
| `serial` | string | 证书序列号（大写十六进制） |
| `revokedAt` | string | 吊销时间；未吊销时为零值 |
| `reasonCode` | int | 吊销原因码；未吊销为 `-1` |
| `reasonText` | string | 原因名（如 `keyCompromise`） |
| `crlNumber` | int | CRL Number 扩展值 |
| `crlThisUpdate` / `crlNextUpdate` | string | CRL 的时间窗 |

### 错误

`INVALID_CERT`、`INVALID_CRL`、`INVALID_PEM`、`INVALID_TIME`。

### 警告

`CRL_UNVERIFIED`（未提供 `issuer` 时）。

> 服务端**必须**先校验 CRL 签名与有效期，再判断吊销状态；验签失败**直接**返回 `INVALID_CRL`，不给任何吊销结论。
> 未提供 `issuer` 时**绝不**返回 `good`：最多给出 `revoked`（有正面证据）或 `unverified`（无结论）。
> 过期的 CRL 一律拒绝（`INVALID_CRL`），无论结论是正是负。

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
| `issuer` | array | 签发者 RDN 条目 `[{ field, value }]` |
| `issuerText` | string | 签发者单行文本（OpenSSL 风格） |
| `thisUpdate` / `nextUpdate` | string | 本次 / 下次更新时间（RFC3339）；`nextUpdate` 缺失时为零值 |
| `signatureAlgorithm` / `signatureAlgorithmOid` | string | 签名算法短名 / 点分 OID |
| `number` | int | CRL Number 扩展值；无该扩展为 `-1` |
| `revoked` | array | `[{ serial, revokedAt, reasonCode, reasonText }]` |

### 错误

`INVALID_CRL`、`INVALID_PEM`、`TOO_MANY_ITEMS`。

> 吊销条目上限 10000。
> `serial` 为**空字符串**表示序列号超过 64 位：上游用 64 位整数承载序列号，超长时无法表达（见上游缺口登记），此时**不**输出错误值。

---

## crl-gen

生成证书吊销列表。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `issuerCert` | object | 是 | 签发者证书（须为带 SKID 的 CA 证书） |
| `issuerKey` | object | 是 | `{ data, format, passphrase? }` |
| `number` | string | 否 | CRL 编号（十六进制）；为空时自动生成 |
| `thisUpdate` / `nextUpdate` | string | 否 | 有效期（RFC3339）；`thisUpdate` 为空时用当前时间 |
| `revoked` | array | 否 | `[{ cert, revokedAt?, reasonCode }]` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `crlPem` | string | CRL 的 PEM |
| `crlDerBase64` | string | CRL 的 DER（base64） |
| `issuerText` | string | 签发者单行文本 |
| `number` | int | CRL 编号 |
| `thisUpdate` / `nextUpdate` | string | 实际生效的时间窗 |
| `revokedCount` | int | 写入的吊销条目数 |

### 错误

`INVALID_CERT`、`INVALID_KEY`、`KEY_MISMATCH`、`INVALID_FIELD`、`INVALID_TIME`、`TOO_MANY_ITEMS`。

> 原因码取 RFC 5280 §5.3.1，码值 `7` 为保留值，**禁止**使用（返回 `INVALID_FIELD`）。
> 吊销项**只接受证书**：上游 `CRLBuilder.Revoke` 没有「按序列号吊销」的入口（见 I-20）。
> 签发者证书必须带 SKID，否则返回 `INVALID_CERT`；`nextUpdate` 必须晚于 `thisUpdate`。
> 吊销条目上限 10000。

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
| `response` | object | 是 | `PemBundle`（自动识别 DER / base64 / hex） |
| `cert` | object | 是 | 被查询证书 |
| `issuer` | object | 是 | 签发者证书 |
| `trustedRoots` | array | 否 | 响应签名的信任根；为空时**不验签** |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | int | 响应级状态码（`0` 表示 successful） |
| `statusText` | string | 响应级状态说明 |
| `certStatus` | string | `good` \| `revoked` \| `unknown` |
| `signatureVerified` | bool | 响应签名是否已验证 |
| `producedAt` / `thisUpdate` / `nextUpdate` | string | 时间 |
| `revocationTime` | string | 吊销时间 |
| `reasonCode` | int | 吊销原因码；无原因时为 `-1` |
| `reasonText` | string | 原因名 |
| `responderCount` | int | 响应内嵌的响应者证书数 |

### 错误

`INVALID_CERT`、`INVALID_DER`、`MISSING_FIELD`、`SIGNATURE_INVALID`。

### 警告

`CRL_UNVERIFIED`（未提供 `trustedRoots` 时）。

> 响应签名校验需要 `trustedRoots`：未提供时**不**验签，`signatureVerified` 为 `false`，
> 此时响应内容只被解读，**不能**据此相信结论。
> `cert` 与 `issuer` 必填：响应里的 CertID 需要它们才能定位目标证书。

---

## cert-gen

签发证书（自签或由 CA 签发）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `signMode` | string | 是 | `self`（自签）\| `ca`（由 CA 签发） |
| `issuer` | object | `ca` 时是 | `{ cert, key: { data, format, passphrase? } }` |
| `subject` | object | `self` 时是 | 主体 DN（用 `csr` / `templateCert` 时可省） |
| `csr` | object | 二选一 | `PemBundle`；公钥取自 CSR |
| `templateCert` | object | 二选一 | `PemBundle`；**只**复用其公钥与有效期 |
| `subjectOverride` | object | 否 | 逐项覆盖主体字段 |
| `key` | object | 否 | `{ type, bits, curve }` |
| `existingPrivateKey` | object | 否 | 使用已有私钥 |
| `notAfterDays` | int | 否 | 有效期天数，默认 365 |
| `notAfter` | string | 否 | 绝对失效时间（RFC3339），优先于 `notAfterDays` |
| `serial` | string | 否 | 序列号（十六进制）；缺省随机生成 |
| `sans` / `keyUsages` / `extKeyUsages` | array | 否 | 扩展 |
| `basicConstraints` | object | 否 | `{ isCa, pathLen }` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certPem` | string | 新证书的 PEM |
| `privateKeyPem` | string | 新生成密钥时的 PKCS#8 私钥 |
| `chainPem` | array | `ca` 模式下为「叶 → 签发者」；`self` 模式只有叶证书 |
| `serial` | string | 序列号（大写十六进制） |
| `subjectText` / `issuerText` | string | 主体 / 签发者单行文本 |
| `notBefore` / `notAfter` | string | 有效期（RFC3339） |
| `keyType` / `keyBits` / `keyCurve` | string / int / string | 新证书公钥参数 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`、`INVALID_CERT`、`INVALID_KEY`、`KEY_MISMATCH`、
`NOT_AFTER_BEFORE_NOW`、`SERIAL_INVALID`、`INVALID_TIME`、`UNSUPPORTED_FORMAT`。

> `ca` 模式**必须**校验签发者私钥与签发者证书是否配对；`self` 模式校验公钥与签名私钥是否配对。两者不匹配均返回 `KEY_MISMATCH`。
> SKID / AKID **默认自动补齐**：CA 模式下叶证书的 AKID 取自签发者证书的 SKID。
> `templateCert` **不**拷贝扩展：上游没有「按 DER 添加扩展」的入口（见 I-19）。
> `serial` 上限 8 字节：上游用 64 位整数承载序列号，超出返回 `SERIAL_INVALID`（见 I-21）。

---

## pfx-pack

打包 PKCS#12。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cert` | object | 是 | 叶证书 |
| `key` | object | 是 | 与证书配对的私钥（`KeyInput`） |
| `caCerts` | array | 否 | 随包的 CA 链（每个一项 `PemBundle`） |
| `passphrase` | string | 否 | 容器口令；为空表示不加密 |
| `friendlyName` | string | 否 | 友好名称 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `base64` | string | PKCS#12 容器的 DER（base64） |
| `length` | int | 字节长度 |
| `encoding` | string | 固定 `base64` |

### 错误

`INVALID_CERT`、`INVALID_KEY`、`KEY_MISMATCH`、`INVALID_PKCS12`、`TOO_MANY_ITEMS`。

### 警告

`LOSSY_FORMAT`（口令为空时，容器不受口令保护）。

---

## pfx-parse

解析 PKCS#12。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `container` | object | 是 | `PemBundle`（DER / base64 / hex） |
| `passphrase` | string | 否 | 容器口令 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `hasPrivateKey` | bool | 容器是否含私钥 |
| `certificate` | object | 主证书的元信息（无证书时为 `null`） |
| `privateKeyPem` | string | 容器内私钥的 PKCS#8 PEM |
| `caCertificates` | array | 随包 CA 链的元信息 |
| `caCount` | int | CA 证书数量 |

### 错误

`INVALID_PKCS12`、`PASSPHRASE_INCORRECT`、`INVALID_CERT`、`MISSING_FIELD`。

---

## pfx-change-password

修改 PKCS#12 口令。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `container` | object | 是 | `PemBundle` |
| `oldPassphrase` | string | 是 | 原口令 |
| `newPassphrase` | string | 是 | 新口令 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `base64` | string | 改密后的容器（base64） |
| `length` | int | 字节长度 |
| `encoding` | string | 固定 `base64` |

### 错误

`INVALID_PKCS12`、`PASSPHRASE_INCORRECT`、`MISSING_FIELD`。

> ⚠️ 只支持**本库打包**的容器：上游 `PKCS12_newpass` 无法对外部实现产出的容器改密（见 I-22）。
> 两个口令都必填：缺任一返回 `MISSING_FIELD`。

---

## p7b-pack

打包 PKCS#7（DER / PEM）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `certs` | array | 是 | 待打包证书（每个一项 `PemBundle`），上限 64 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `pem` | string | 证书袋的 PEM |
| `derBase64` | string | 证书袋的 DER（base64） |
| `length` | int | DER 字节长度 |
| `certCount` | int | 打包的证书数量 |

### 错误

`INVALID_CERT`、`INVALID_PKCS7`、`MISSING_FIELD`、`TOO_MANY_ITEMS`。

> 只支持**证书袋**：签名 / 加密容器不在本版范围（见 I-11）。

---

## p7b-extract

从 PKCS#7 中提取证书。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `container` | object | 是 | `PemBundle`（自动识别 DER / base64 / hex） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certificates` | array | 提取出的证书元信息 |
| `count` | int | 证书数量 |
| `pems` | array | 提取出的证书 PEM |

### 错误

`INVALID_PKCS7`、`INVALID_PEM`、`INVALID_DER`、`TOO_MANY_ITEMS`。

---

## caa-gen

生成 CAA（证书颁发机构授权）记录。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `domain` | string | 是 | 域名 |
| `records` | array | 是 | `[{ flags?, tag, value }]`；`tag` ∈ `issue` \| `issuewild` \| `iodef` |
| `ttl` | int | 否 | TTL 秒数；`0` 表示不输出 TTL |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `domain` | string | 域名 |
| `records` | array | `[{ flags, tag, value, zone }]` |
| `zoneLines` | array | 可直接粘贴到区域文件的行（含 TTL） |
| `count` | int | 记录数量 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> 纯字符串拼装，**不**涉及密码学运算；记录数上限 32。

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
