# 加密 API

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

覆盖 **8** 个加密类工具。所有端点为 `POST /api/v1/tools/{slug}`，请求与响应使用[统一信封](../README.md)。

| slug | 名称 | 传输 | 档位 |
| --- | --- | --- | --- |
| [signature](#signature) | 数字签名 | JSON | standard |
| [asymmetric-crypt](#asymmetric-crypt) | 非对称加解密 | JSON | standard |
| [symmetric-crypt](#symmetric-crypt) | 对称加解密 | JSON | standard |
| [ecdh](#ecdh) | ECDH 密钥协商 | JSON | standard |
| [sm4-file-crypt](#sm4-file-crypt) | SM4 文件加解密 | multipart | file |
| [hmac-gen](#hmac-gen) | HMAC 计算 | JSON | light |
| [password-gen](#password-gen) | 密码生成 | JSON | light |
| [hash-calc](#hash-calc) | 摘要计算 | JSON | light |

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
| `format` | string | 否 | `der`（默认）\| `raw` |
| `hash` | string | 否 | RSA / ECDSA 摘要：`SHA256`（默认）\| `SHA384` \| `SHA512` \| `SM3` |
| `rsaPadding` | string | 否 | `PSS`（默认）\| `PKCS1` |
| `saltLength` | int | 否 | PSS salt 长度（默认 = 摘要长度） |
| `sm2UserId` | string | 否 | SM2 userId（默认 `1234567812345678`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `valid` | bool | **仅 `verify` 时**返回 |
| `signature` | object | **仅 `sign` 时**返回，`{ data, encoding }` |
| `algorithm` | string | 实际使用的算法 |
| `hash` | string | 实际使用的摘要 |
| `format` | string | 签名编码 |

### 错误

`UNSUPPORTED_ALGORITHM`、`KEY_SIZE_MISMATCH`、`SIGNATURE_INVALID`、`INVALID_KEY`、`INVALID_FIELD`。

> **验签不通过**返回 `200` + `valid: false`，**不**是错误。

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
    "signature": { "data": "MEUCIQD…", "encoding": "base64" },
    "algorithm": "SM2",
    "hash": "SM3",
    "format": "der"
  },
  "error": null,
  "meta": { "durationMs": 8, "inputBytes": 512, "outputBytes": 384, "warnings": [] },
  "requestId": "…"
}
```

> SM2 的 `userId` **必须**签名与验签一致，否则验签失败。Ed25519 / Ed448 **不**支持 `hash` 参数。

---

## asymmetric-crypt

非对称加解密（RSA、SM2、ECIES）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `RSA` \| `SM2` \| `ECIES` |
| `operation` | string | 是 | `encrypt` \| `decrypt` |
| `key` | object | 是 | 加密用公钥，解密用私钥（`KeyInput`） |
| `input` | object | 是 | `{ data, encoding }` |
| `sm2Order` | string | 否 | `C1C3C2`（默认）\| `C1C2C3` |
| `sm2Prefix04` | bool | 否 | 是否保留 `04` 前缀（默认 `true`） |
| `sm2Format` | string | 否 | `raw`（默认）\| `der` |
| `rsaPadding` | string | 否 | `OAEP`（默认）\| `PKCS1` |
| `oaepHash` | string | 否 | OAEP 摘要（默认 `SHA256`） |
| `outputEncoding` | string | 否 | 输出编码（默认 `base64`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 密文 |
| `algorithm` | string | 实际算法 |
| `order` | string | SM2 密文顺序 |
| `format` | string | SM2 密文格式 |

### 错误

`UNSUPPORTED_ALGORITHM`、`KEY_SIZE_MISMATCH`、`DECRYPT_FAILED`、`PLAINTEXT_TOO_LONG`、`INVALID_KEY`。

> SM2 默认输出为 **`C1C3C2` + `04` 前缀 + 裸格式**，兼容 BouncyCastle 等主流实现。解密时服务端**自动识别** DER / `C1C3C2` / `C1C2C3` / 压缩前缀四种输入。

---

## symmetric-crypt

对称加解密（SM4、AES、DES、3DES）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `SM4` \| `AES` \| `DES` \| `3DES` |
| `operation` | string | 是 | `encrypt` \| `decrypt` |
| `mode` | string | 是 | `ECB` \| `CBC` \| `CTR` \| `OFB` \| `CFB` \| `GCM` |
| `padding` | string | 否 | `pkcs7`（默认）\| `zero` \| `none` |
| `key` | object | 是 | `{ data, encoding }` |
| `iv` | object | 视模式 | `{ data, encoding }`；GCM 为 12 字节，其余为 16 字节 |
| `aad` | object | 否 | GCM 附加认证数据 |
| `tag` | object | GCM 解密时是 | 认证标签（16 字节） |
| `input` | object | 是 | `{ data, encoding }` |
| `outputEncoding` | string | 否 | 默认 `base64` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | string | 密文 / 明文 |
| `tag` | string | GCM 认证标签（十六进制） |
| `algorithm` / `mode` / `padding` | string | 实际参数 |
| `keyBytes` / `ivBytes` | int | 实际长度 |

### 错误

`UNSUPPORTED_MODE`、`UNSUPPORTED_PADDING`、`KEY_SIZE_MISMATCH`、`IV_SIZE_MISMATCH`、`TAG_REQUIRED`、`TAG_INVALID`、`DECRYPT_FAILED`。

### 警告

- `INSECURE_ECB` —— 使用 ECB 模式；
- `NO_AUTHENTICATION` —— 非 GCM 模式（无认证标签）。

> `padding: none` 要求输入长度为分组长度（16 字节）的整数倍。GCM 模式**忽略** `padding`。

---

## ecdh

ECDH 密钥协商 + KDF 派生。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `curve` | string | 是 | `P-256` \| `P-384` \| `P-521` \| `X25519` \| `X448` |
| `privateKey` | object | 是 | 本方私钥（`KeyInput`） |
| `peerPublicKey` | object | 是 | 对端公钥（`KeyInput`） |
| `kdf` | object | 否 | `{ type, hash, salt, info, length }` |
| `outputEncoding` | string | 否 | 默认 `base64` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `sharedSecret` | string | 共享密钥 |
| `derivedKey` | string | KDF 派生结果（提供 `kdf` 时） |
| `curve` | string | 实际曲线 |
| `length` | int | 共享密钥字节数 |

### 错误

`UNSUPPORTED_ALGORITHM`、`KEY_SIZE_MISMATCH`、`INVALID_KEY`、`INVALID_FIELD`。

> X25519 / X448 **不**支持 `kdf.type` 之外的派生参数；ECDH 输出的共享密钥**必须**经过 KDF 后再使用。

---

## sm4-file-crypt

SM4 文件级加解密，使用自定义容器 `TBFC`（分块 GCM）。

### 请求（multipart）

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `operation` | string | 是 | `encrypt` \| `decrypt` |
| `mode` | string | 否 | 仅 `GCM` |
| `key` | string | 是 | JSON 串 `{"data":"…","encoding":"hex"}` |
| `iv` | string | 否 | JSON 串；缺省时生成随机 12 字节并随响应头返回 |
| `chunkSize` | int | 否 | 分块大小，默认 1048576 |
| `file` | file | 是 | 待处理文件（单文件） |

### 响应

**二进制流**（不套信封）：

| 场景 | `Content-Type` | `Content-Disposition` |
| --- | --- | --- |
| 加密 | `application/octet-stream` | `attachment; filename*=UTF-8''<原名>.sm4` |
| 解密 | `application/octet-stream` | `attachment; filename*=UTF-8''<原名>` |

响应头额外携带 `X-Warnings`。

### 错误

`TOO_LARGE`、`KEY_SIZE_MISMATCH`、`IV_SIZE_MISMATCH`、`DECRYPT_FAILED`、`INVALID_FIELD`。

> 容器格式：`TBFC`(4) + version(1) + flags(1) + chunkSize(4) + baseNonce(12)，随后为分块。分块 nonce 由 `baseNonce[0:8] || uint32be(块索引)` 派生，AAD 绑定块索引。

---

## hmac-gen

HMAC 计算。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `SM3` \| `SHA1` \| `SHA224` \| `SHA256` \| `SHA384` \| `SHA512` \| `MD5` |
| `key` | object | 是 | `{ data, encoding }` |
| `message` | object | 二选一 | `{ data, encoding }` |
| `file` | file | 二选一 | 以文件内容为消息 |
| `outputEncoding` | string | 否 | 默认 `hex` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mac` | object | `{ data, encoding }` |
| `algorithm` | string | 实际算法 |

### 错误

`UNSUPPORTED_ALGORITHM`、`INVALID_FIELD`、`MISSING_FIELD`、`INVALID_ENCODING`。

### 警告

`WEAK_HASH` —— 使用 `MD5` / `SHA1`。

---

## password-gen

随机密码生成。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `length` | int | 否 | 密码长度，默认 16（4–128） |
| `count` | int | 否 | 生成数量，默认 1（1–100） |
| `charsets` | array | 否 | 字符集集合：`lower` `upper` `digit` `symbol` |
| `excludeAmbiguous` | bool | 否 | 排除 `0O1lI` 等易混字符 |
| `requireEachClass` | bool | 否 | 每类字符至少出现一次 |
| `customCharset` | string | 否 | 自定义字符集（覆盖 `charsets`） |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `passwords` | array | 生成的密码列表 |
| `entropyBits` | number | 单个密码的估算熵（bit） |

### 错误

`INVALID_FIELD`、`TOO_MANY_ITEMS`。

> 使用密码学安全随机源（CSPRNG）。**禁止**用 `math/rand`。

---

## hash-calc

摘要计算。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithms` | array | 是 | 如 `["SM3","SHA256"]` |
| `input` | object | 二选一 | `{ data, encoding }` |
| `file` | file | 二选一 | 以文件内容为输入 |
| `outputEncoding` | string | 否 | 默认 `hex` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `digests` | array | `[{ algorithm, value }]` |
| `inputBytes` | int | 输入字节数 |

### 错误

`UNSUPPORTED_ALGORITHM`、`INVALID_FIELD`、`MISSING_FIELD`。

### 警告

`WEAK_HASH` —— 请求中包含 `MD5` / `SHA1`。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封、通用类型、限流
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/03-crypto.md](../../standards/03-crypto.md) —— 密码学规范
