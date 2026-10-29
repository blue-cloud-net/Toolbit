# 加密 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **10** 个加密类工具。所有端点为 `POST /api/v1/tools/{slug}`，请求与响应使用[统一信封](../README.md)。

| slug | 名称 | 传输 | 档位 | 实现状态 |
| --- | --- | --- | --- | --- |
| [signature](#signature) | 数字签名 | JSON | standard | ✅ v0.3.0 |
| [asymmetric-crypt](#asymmetric-crypt) | 非对称加解密 | JSON | standard | ✅ v0.3.0 |
| [symmetric-crypt](#symmetric-crypt) | 对称加解密 | JSON | standard | ✅ v0.3.0 |
| [ecdh](#ecdh) | ECDH 密钥协商 | JSON | standard | ✅ v0.3.0 |
| [hmac-gen](#hmac-gen) | HMAC 计算 | JSON | light | ✅ v0.3.0 |
| [password-gen](#password-gen) | 密码生成 | JSON | light | ✅ v0.3.0 |
| [hash-calc](#hash-calc) | 摘要计算 | JSON | light | ✅ v0.3.0 |
| [sm4-file-crypt](#sm4-file-crypt) | SM4 文件加解密 | multipart | file | ✅ v0.3.1 |
| [password-policy](#password-policy) | 口令策略检查 | JSON | light | ✅ v0.8.0 |
| [totp](#totp) | TOTP 生成校验 | JSON | light | ✅ v0.8.0 |

> **实现状态**列标注该工具的端点是否已交付。
> 未实现的工具返回 `501 NOT_IMPLEMENTED`，catalog 仍返回全部条目。

---

## 通用约定

### 算法支持矩阵

下表是**服务端实际支持**的能力，**不是**目标矩阵。
`tongsuo-go` 尚未提供的能力（DES / 3DES / ECIES 等）在本版**一律不可用**，
缺口登记在 [../../upstream/tongsuo-go-issues.md](../../upstream/tongsuo-go-issues.md)。

| 类别 | 算法 | 支持 |
| --- | --- | --- |
| 摘要 | SM3 / MD5 / SHA1 / SHA224 / SHA256 / SHA384 / SHA512 | ✅ |
| MAC | HMAC-SM3 / HMAC-MD5 / HMAC-SHA1 / HMAC-SHA224 / HMAC-SHA256 / HMAC-SHA384 / HMAC-SHA512 | ✅ |
| 对称 | SM4 | ✅ |
| 对称 | AES-128 / AES-192 / AES-256 | ✅ |
| 对称 | DES / 3DES | ❌ 缺口 I-02 |
| 非对称加解密 | SM2 / RSA | ✅ |
| 非对称加解密 | ECIES | ❌ 缺口 I-05 |
| 签名 | SM2 / RSA / ECDSA / Ed25519 / Ed448 | ✅ |
| 密钥协商 | P-256 / P-384 / P-521 / X25519 / X448 | ✅ |

### 密钥输入

密钥字段一律用 [KeyInput](../README.md#43-keyinput)：

```json
{ "data": "-----BEGIN PRIVATE KEY-----…", "format": "auto", "passphrase": "" }
```

`format: auto` 的识别顺序：**PEM 头 → base64 → hex**（后两者解码后须是 DER），
全部失败返回 `INVALID_PEM`。

### 字节输入

字节字段一律用 [ByteInput](../README.md#42-byteinput-output)：

```json
{ "data": "48656c6c6f", "encoding": "hex" }
```

`encoding` ∈ `utf8` \| `hex` \| `base64` \| `base64url`。

---

## signature

数字签名与验签。签名与验签共用一个端点，由 `operation` 区分。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `SM2` \| `RSA` \| `ECDSA` \| `ED25519` \| `ED448` |
| `operation` | string | 是 | `sign` \| `verify` |
| `message` | object | 是 | `{ data, encoding }` |
| `key` | object | `sign` 时是 | 私钥（`KeyInput`） |
| `publicKey` | object | `verify` 时是 | 公钥（`KeyInput`） |
| `signature` | object | `verify` 时是 | `{ data, encoding }` |
| `hash` | string | 否 | **仅 RSA**：`SHA256`（默认）\| `SHA384` \| `SHA512` |
| `rsaPadding` | string | 否 | **仅 RSA**：`pss`（默认）\| `pkcs1` |
| `sm2UserId` | string | 否 | SM2 userId（默认 `1234567812345678`） |

> **摘要**：`SM2` 固定 `SM3`，`ECDSA` 固定 `SHA-256`，`ED25519` / `ED448` 是纯签名
> （RFC 8032）——这三者**不接受** `hash` 参数，传入即报 `UNSUPPORTED_ALGORITHM`。
>
> **纯签名语义**：Ed25519 / Ed448 的消息**原样**参与签名，调用方**不得**预先哈希。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `valid` | bool | **仅 `verify` 时**返回 |
| `signature` | string | **仅 `sign` 时**返回，Base64 编码的 DER |
| `algorithm` | string | 实际使用的算法 |
| `hash` | string | 实际使用的摘要；**仅 RSA** 返回（`SHA256` / `SHA384` / `SHA512`），其余算法为空 |

> 响应里的 `hash` 只在**摘要由调用方选择**时才有值。
> `SM2`（固定 SM3）、`ECDSA`（固定 SHA-256）、`ED25519` / `ED448`（纯签名）
> 的摘要由算法决定，因此该字段为空字符串。

### 错误

`UNSUPPORTED_ALGORITHM`、`KEY_SIZE_MISMATCH`、`SIGNATURE_INVALID`、`INVALID_KEY`、`INVALID_FIELD`、`INVALID_PEM`。

> **验签不通过**返回 `200` + `valid: false`，**不**是错误。
> 只有签名**结构非法**（长度不符、无法解析为 DER）才返回 `SIGNATURE_INVALID`。

```json
{
  "algorithm": "SM2",
  "operation": "sign",
  "message": { "data": "待签名内容", "encoding": "utf8" },
  "key": { "data": "-----BEGIN PRIVATE KEY-----…", "format": "auto" }
}
```

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "signature": "MEUCIQD…",
    "algorithm": "SM2",
    "hash": "SM3"
  },
  "error": null,
  "meta": { "durationMs": 8, "inputBytes": 512, "outputBytes": 384, "warnings": [] },
  "requestId": "…"
}
```

> SM2 的 `userId` **必须**签名与验签一致，否则验签不通过（`valid: false`）。
> SM2 签名输出为 ASN.1 DER，与 OpenSSL 兼容。

---

## asymmetric-crypt

非对称加解密（RSA、SM2）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `RSA` \| `SM2` |
| `operation` | string | 是 | `encrypt` \| `decrypt` |
| `key` | object | 是 | 加密用公钥，解密用私钥（`KeyInput`） |
| `input` | object | 是 | `{ data, encoding }` |
| `rsaPadding` | string | 否 | **仅 RSA**：`oaep`（默认）\| `pkcs1` |
| `oaepHash` | string | 否 | **仅 RSA**：OAEP 摘要（默认 `SHA256`） |
| `outputEncoding` | string | 否 | 输出编码（默认 `base64`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 密文或明文 |
| `algorithm` | string | 实际算法 |
| `order` | string | SM2 密文顺序（仅 SM2 返回，恒为 `C1C3C2`） |
| `format` | string | SM2 密文格式（仅 SM2 返回，恒为 `raw`） |

### 错误

`UNSUPPORTED_ALGORITHM`、`UNSUPPORTED_PADDING`、`KEY_SIZE_MISMATCH`、`DECRYPT_FAILED`、`PLAINTEXT_TOO_LONG`、`INVALID_KEY`、`INVALID_FIELD`、`INVALID_PEM`。

> SM2 输出固定为 **`C1C3C2` + `04` 前缀 + 裸格式**，兼容 BouncyCastle 等主流实现。
> 解密时服务端**自动识别** DER / `C1C3C2` / `C1C2C3` 三种输入；
> 压缩点前缀（`02` / `03`）**暂不支持**，返回 `UNSUPPORTED_FORMAT`。
>
> RSA 密文长度受限，明文超出上限返回 `PLAINTEXT_TOO_LONG`。

---

## symmetric-crypt

对称加解密（SM4、AES）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `SM4` \| `AES` |
| `operation` | string | 是 | `encrypt` \| `decrypt` |
| `mode` | string | 是 | `ECB` \| `CBC` \| `CTR` \| `OFB` \| `CFB` \| `GCM` |
| `padding` | string | 否 | `pkcs7`（默认）\| `zero` \| `none` |
| `key` | object | 是 | `{ data, encoding }` |
| `iv` | object | 视模式 | `{ data, encoding }`；GCM 为 12 字节，其余为 16 字节 |
| `aad` | object | 否 | GCM 附加认证数据 |
| `tag` | object | GCM 解密时是 | 认证标签（16 字节） |
| `input` | object | 是 | `{ data, encoding }` |
| `outputEncoding` | string | 否 | 默认 `base64` |

### 模式与填充矩阵

| 算法 | 支持的模式 | 支持的填充 |
| --- | --- | --- |
| SM4 | `ECB` `CBC` `CTR` `OFB` `CFB` `GCM` | `ECB` / `CBC`：`pkcs7` `zero` `none`；其余：`pkcs7` |
| AES | `ECB` `CBC` `CTR` `GCM` | 一律 `pkcs7` |

> 该矩阵由服务端**自行维护**：`tongsuo-go` 的按名入口对 AES 会**静默忽略**
> `padding` 参数（恒用 `pkcs7`），若不显式拦截，`padding: none` 的 AES 密文
> 将无法按预期解密。
>
> `padding: none` 要求输入长度为分组长度（16 字节）的整数倍，否则返回
> `INVALID_FIELD`。

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 密文 / 明文 |
| `tag` | string | GCM 认证标签（十六进制） |
| `iv` | string | **自动生成**的 nonce（十六进制）；仅 GCM 加密且未提供 `iv` 时返回 |
| `algorithm` / `mode` / `padding` | string | 实际参数 |
| `keyBytes` / `ivBytes` | int | 实际长度 |

### 错误

`UNSUPPORTED_MODE`、`UNSUPPORTED_PADDING`、`KEY_SIZE_MISMATCH`、`IV_SIZE_MISMATCH`、`TAG_REQUIRED`、`TAG_INVALID`、`DECRYPT_FAILED`、`INVALID_FIELD`、`INVALID_ENCODING`。

### 警告

- `INSECURE_ECB` —— 使用 ECB 模式；
- `NO_AUTHENTICATION` —— 非 GCM 模式（无认证标签）。

> GCM 加密**未提供** `iv` 时，服务端生成随机 nonce 并在响应的 `iv` 字段回传；
> 解密时**必须**复用同一个 nonce，否则 tag 校验失败。

---

## ecdh

ECDH 密钥协商 + 可选 KDF 派生。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `curve` | string | 是 | `P-256` \| `P-384` \| `P-521` \| `X25519` \| `X448` |
| `privateKey` | object | 是 | 本方私钥（`KeyInput`） |
| `peerPublicKey` | object | 是 | 对端公钥（`KeyInput`） |
| `kdf` | object | 否 | `{ type, hash, salt, info, length }` |
| `outputEncoding` | string | 否 | 默认 `base64` |

`kdf` 子字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `type` | string | 否 | `hkdf`（唯一支持值） |
| `hash` | string | 否 | `SHA256`（默认）\| `SHA384` \| `SHA512` \| `SM3` |
| `salt` | object | 否 | 盐值（`ByteInput`） |
| `info` | object | 否 | 上下文信息（`ByteInput`） |
| `length` | int | 是 | 派生密钥字节数（1 – 1048576） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `sharedSecret` | string | 共享密钥 |
| `derivedKey` | string | KDF 派生结果（提供 `kdf` 时） |
| `curve` | string | 实际曲线 |
| `length` | int | 共享密钥字节数 |

### 错误

`UNSUPPORTED_ALGORITHM`、`UNSUPPORTED_FORMAT`、`KEY_SIZE_MISMATCH`、`INVALID_KEY`、`INVALID_FIELD`、`INVALID_PEM`、`MISSING_FIELD`。

> **共享密钥必须再经 KDF 派生后使用**，不要把原始共享密钥直接当会话密钥。
> 派生长度上限 1 MB，超出返回 `TOO_LARGE`。

---

## sm4-file-crypt

SM4 文件级加解密，使用自定义容器 `TBFC`（分块 GCM）。

### 请求（multipart/form-data）

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `encrypt` \| `decrypt` |
| `mode` | string | 否 | 仅 `GCM`（默认） |
| `key` | string | 是 | JSON 串 `{"data":"…","encoding":"hex"}`；密钥固定 16 字节 |
| `iv` | string | 否 | JSON 串；基础 nonce，12 字节；缺省时随机生成 |
| `chunkSize` | int | 否 | 分块大小，默认 1048576（范围 4096 – 16777216） |
| `file` | file | 是 | 待处理文件（**单文件**，**必须**是最后一个 part） |

> **文件部分必须是最后一个 part**：服务端按顺序读表单，
> 读到文件后就无法回头取后面的字段。
>
> `key` / `iv` 的取值格式与 [ByteInput](../README.md#42-byteinput-output) 相同，
> 但以 **JSON 字符串**传递（multipart 的表单字段都是字符串）。
> `encoding` 缺省按 `hex` 处理。

### 响应

**二进制流**（不套信封）：

| 场景 | `Content-Type` | `Content-Disposition` |
| --- | --- | --- |
| 加密 | `application/octet-stream` | `attachment; filename="<原名>.sm4"` |
| 解密 | `application/octet-stream` | `attachment; filename="<原名>"` |

响应头：

| 响应头 | 说明 |
| --- | --- |
| `X-Warnings` | 警告列表（JSON 数组）；无警告时**不**出现 |
| `X-Duration-Ms` / `X-Input-Bytes` / `X-Output-Bytes` | 与信封的 `meta` 同义 |

> 文件名来自客户端，**不可信**：服务端只取基础名并剔除路径分隔符与控制字符。
> `filename*` 用 RFC 5987 的 UTF-8 编码，`filename` 是 ASCII 回退。

### 错误

错误**在响应头发出之前**就能确定，因此失败时返回 **JSON 信封**（不是二进制流）：

| 错误码 | 触发条件 |
| --- | --- |
| `INVALID_FIELD` | 操作 / 模式 / 分块大小非法；表单字段畸形；输入不是 TBFC 容器 |
| `MISSING_FIELD` | 缺少 `key` 或 `file` |
| `KEY_SIZE_MISMATCH` | 密钥长度不是 16 字节 |
| `IV_SIZE_MISMATCH` | 基础 nonce 长度不是 12 字节 |
| `UNSUPPORTED_MODE` | 模式不是 `GCM` |
| `TOO_LARGE` | 请求体超过 100 MB |

> **解密失败**（密钥不对、容器被篡改、容器被截断）返回 `400` + JSON 信封。
> 服务端会**先同步校验容器头与第一块**，因此这类错误**不会**表现为
> 「200 + 被截断的下载」。
>
> 流式响应**开始之后**的中断无法再改状态码，只能记日志与统计。

### 容器格式 `TBFC`

```text
Header (22 字节)
  magic      [4]byte  = "TBFC"
  version    uint8    = 1
  flags      uint8    = 0
  chunkSize  uint32be
  baseNonce  [12]byte

Chunk
  length     uint32be   （明文长度；0 表示结束）
  ciphertext [length]byte
  tag        [16]byte
```

- **分块 nonce**：`baseNonce[0:8] || uint32be(chunkIndex)`；
- **AAD**：`容器头 || uint32be(chunkIndex)`；
- **结束块**：`length = 0` + 一个对空明文计算的 `tag`。
  它让**截断**的容器能被检测出来，而不被误判成「文件到此为止」；
- 内存占用约 `2 × chunkSize`，与文件大小无关。

> ⚠️ 容器是**自定义格式**，`openssl` **不能**直接解密。

---

## hmac-gen

HMAC 计算。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `HMAC-SM3` \| `HMAC-SHA256` \| `HMAC-SHA384` \| `HMAC-SHA512` \| `HMAC-SHA224` \| `HMAC-SHA1` \| `HMAC-MD5` |
| `key` | object | 是 | `{ data, encoding }` |
| `message` | object | 是 | `{ data, encoding }` |
| `outputEncoding` | string | 否 | 默认 `hex` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mac` | string | 认证码 |
| `algorithm` | string | 实际算法 |

### 错误

`UNSUPPORTED_ALGORITHM`、`INVALID_FIELD`、`MISSING_FIELD`、`INVALID_ENCODING`。

### 警告

- `WEAK_HASH` —— 使用 `HMAC-MD5` / `HMAC-SHA1`。

---

## password-gen

随机密码生成。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `length` | int | 否 | 密码长度，默认 16（4–128） |
| `count` | int | 否 | 生成数量，默认 1（1–100） |
| `charsets` | array | 否 | 字符集集合：`lower` `upper` `digit` `symbol`；缺省为全部四类 |
| `excludeAmbiguous` | bool | 否 | 排除 `0O1lI` 等易混字符 |
| `requireEachClass` | bool | 否 | 每类字符至少出现一次 |
| `customCharset` | string | 否 | 自定义字符集（覆盖 `charsets`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `passwords` | array | 生成的密码列表 |
| `entropyBits` | number | 单个密码的估算熵（bit） |

### 错误

`INVALID_FIELD`、`MISSING_FIELD`。

> `customCharset` 与 `requireEachClass` **互斥**：同时提供返回 `INVALID_FIELD`。
> 随机源是 Tongsuo 的 CSPRNG，字符选择使用**拒绝采样**避免模偏差。
> **禁止**用 `math/rand`。

---

## hash-calc

摘要计算。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithms` | array | 是 | 如 `["SM3","SHA256"]` |
| `input` | object | 是 | `{ data, encoding }` |
| `outputEncoding` | string | 否 | 默认 `hex` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `digests` | array | `[{ algorithm, value }]` |
| `inputBytes` | int | 输入字节数 |

### 错误

`UNSUPPORTED_ALGORITHM`、`INVALID_FIELD`、`MISSING_FIELD`、`INVALID_ENCODING`。

### 警告

- `WEAK_HASH` —— 请求中包含 `MD5` / `SHA1`。

---

## password-policy

按策略检查口令合规性（Password Policy Checker）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `password` | string | 是 | 待检查口令 |
| `preset` | string | 否 | `basic` `medium` `strict`，默认 `medium` |
| `minLength` | int | 否 | 最小长度；缺省取预设值 |
| `maxLength` | int | 否 | 最大长度；`0` 表示不限 |
| `requireUpper` | bool | 否 | 是否要求大写字母 |
| `requireLower` | bool | 否 | 是否要求小写字母 |
| `requireDigit` | bool | 否 | 是否要求数字 |
| `requireSymbol` | bool | 否 | 是否要求符号 |
| `maxRepeatRun` | int | 否 | 连续重复字符上限；`0` 表示不限 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `preset` | string | 生效的预设 |
| `compliant` | bool | 是否满足整体策略 |
| `score` | int | 强度评分（0–100） |
| `strength` | string | 强度等级 |
| `checks` | array | 检查项 `[{ check, passed, message }]` |
| `suggestions` | array | 改进建议 |

### 错误

`INVALID_FIELD`。

> **纯规则检查**：不涉及密码学运算，也不存储口令；常见口令表内置在服务端。

---

## totp

生成与校验 TOTP 动态口令（TOTP Generator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 否 | `generate` 生成 / `verify` 校验，默认 `generate` |
| `secret` | string | 是 | Base32 密钥 |
| `algorithm` | string | 否 | `SHA1` `SHA256` `SHA512`，默认 `SHA1` |
| `digits` | int | 否 | 验证码位数，取值 6–10，默认 6 |
| `period` | int | 否 | 时间步长（秒），取值 1–300，默认 30 |
| `timestamp` | int | 否 | Unix 时间戳（秒）；`0` 表示当前时间 |
| `code` | string | 否 | 待校验的验证码；`verify` 模式必填 |
| `window` | int | 否 | 允许的时间步容差，默认 1 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `operation` | string | 操作 |
| `code` | string | 验证码 |
| `valid` | bool | 校验结论；生成模式下为 `null` |
| `algorithm` | string | 摘要算法 |
| `digits` | int | 验证码位数 |
| `period` | int | 时间步长 |
| `counter` | int | 时间步计数 |
| `validFrom` | string | 生效起点（RFC 3339） |
| `validUntil` | string | 失效点（RFC 3339） |
| `remainingSeconds` | int | 剩余有效秒数 |

### 错误

`MISSING_FIELD`、`INVALID_FIELD`。

> HMAC 计算经服务端转调 Tongsuo；标准向量见 RFC 6238。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封、通用类型、限流
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/03-crypto.md](../../standards/03-crypto.md) —— 密码学规范
- [../../upstream/tongsuo-go-issues.md](../../upstream/tongsuo-go-issues.md) —— 上游能力缺口登记
