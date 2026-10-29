# 密钥 API

> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29

覆盖 **7** 个密钥类工具。端点均为 `POST /api/v1/tools/{slug}`。

| slug | 名称 | 档位 |
| --- | --- | --- |
| [keygen](#keygen) | 密钥生成 | heavy |
| [key-convert](#key-convert) | 密钥格式转换 | standard |
| [key-parse](#key-parse) | 密钥解析 | standard |
| [key-ops](#key-ops) | 密钥运算 | standard |
| [key-match](#key-match) | 密钥配对检测 | standard |
| [jwk-xml-convert](#jwk-xml-convert) | JWK/XML 转换 | standard |
| [ssh-keygen](#ssh-keygen) | SSH 密钥生成 | light |

---

## keygen

生成密钥对。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 是 | `RSA` \| `SM2` \| `EC` \| `Ed25519` \| `Ed448` \| `X25519` \| `X448` |
| `bits` | int | RSA 时是 | 1024 / 2048（默认）/ 3072 / 4096 |
| `curve` | string | EC / SM2 时是 | `P-256` \| `P-384` \| `P-521` \| `SM2` |
| `privateFormat` | string | 否 | `pkcs8`（默认）\| `pkcs1` |
| `publicFormat` | string | 否 | `spki`（默认）；`pkcs1` **暂不支持**（上游无导出入口，见 [I-14](../../upstream/tongsuo-go-issues.md#i-14-asym-缺-rsa-pkcs1-公钥的导出--加载入口)） |
| `encrypt` | object | 否 | `{ cipher, passphrase }`；`cipher` ∈ `AES-256-CBC` \| `DES-EDE3-CBC` |
| `count` | int | 否 | 生成数量；当前**只支持 1**（响应结构只能表达一把密钥），其余取值返回 `INVALID_FIELD` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `privateKey` | object | `{ pem }` |
| `publicKey` | object | `{ pem }` |
| `algorithm` | string | 实际算法 |
| `bits` | int | 密钥位数（RSA） |
| `curve` | string | 曲线（EC / SM2） |

### 错误

`UNSUPPORTED_ALGORITHM`、`INVALID_FIELD`、`KEY_SIZE_MISMATCH`、`CRYPTO_OPERATION_FAILED`。

### 警告

`SMALL_RSA_KEY` —— RSA 位数小于 2048。

> 私钥**只**在响应中返回，服务端**不**落盘、不记录日志。

---

## key-convert

密钥格式转换。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `key` | object | 是 | `KeyInput` |
| `from` | string | 否 | `auto`（默认）\| `pkcs1` \| `pkcs8` \| `spki` \| `der` \| `pem`；非 `auto` 时**约束**源格式，不匹配返回 `UNSUPPORTED_FORMAT` |
| `to` | string | 是 | `pkcs1` \| `pkcs8` \| `spki` \| `der` \| `pem` |
| `passphrase` | string | 否 | 加密私钥的现有口令 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `key` | object | `{ pem, derBase64 }`；`to` 为 `der` 时 `pem` 为空、`derBase64` 有值，否则反之 |
| `detected` | object | `{ algorithm, format, encrypted, bits, curve }`；`format` 为识别到的编码（`pkcs1` / `pkcs8` / `spki`） |

### 错误

`INVALID_KEY`、`INVALID_PEM`、`INVALID_DER`、`UNSUPPORTED_FORMAT`、`PASSPHRASE_REQUIRED`、`PASSPHRASE_INCORRECT`。

> `pkcs1` 仅适用于 RSA；`spki` 为公钥格式。转换私钥到 `spki` **必须**报错。
>
> 源为加密私钥时，转换结果**保持加密**（AES-256-CBC + 原口令），不会因为转换而解密。
>
> `passphrase` 可写在顶层或 `key.passphrase`，两者等价。

---

## key-parse

解析密钥内容，展示其元信息。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `key` | object | 是 | `KeyInput` |
| `passphrase` | string | 否 | 加密私钥口令 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `algorithm` | string | 算法 |
| `keyType` | string | `private` \| `public` |
| `format` | string | 识别到的编码（`pkcs1` / `pkcs8` / `spki`） |
| `bits` | int | 位数（RSA） |
| `curve` | string | 曲线 |
| `encrypted` | bool | 源是否为加密私钥 |
| `fingerprintSha256` | string | 公钥 SPKI DER 的 SHA-256 指纹（小写十六进制） |
| `params` | object | RSA 参数 `{ n, e, d, p, q, dp, dq, qi }`（私钥才有） |
| `normalizedPem` | string | 规范化后的 PEM |

### 错误

`INVALID_KEY`、`PASSPHRASE_REQUIRED`、`PASSPHRASE_INCORRECT`。

> `params` 中的私钥参数**仅**在输入是**私钥**时返回；响应**禁止**被缓存（`no-store`）。
>
> 加密私钥的规范化输出**仍保持加密**，不会因为解析而解密。
>
> `passphrase` 可写在顶层或 `key.passphrase`，两者等价。

---

## key-ops

密钥操作：加解密私钥、提取公钥、修改口令。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `action` | string | 是 | `encrypt` \| `decrypt` \| `extract-public` \| `change-passphrase` |
| `key` | object | 是 | `KeyInput` |
| `passphrase` | string | 视 action | 现有口令 |
| `newPassphrase` | string | 视 action | 新口令 |
| `pbeCipher` | string | 否 | `AES-256-CBC`（默认）\| `DES-EDE3-CBC` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `key` | object | `{ pem }` |
| `algorithm` | string | 算法 |
| `format` | string | 输出格式 |

### 错误

`UNSUPPORTED_MODE`、`PASSPHRASE_REQUIRED`、`PASSPHRASE_INCORRECT`、`INVALID_KEY`、`INVALID_FIELD`。

> `extract-public` **不**需要口令，且**不**会修改原私钥。

---

## key-match

检测两个密钥是否配对（私钥 ↔ 公钥，或私钥 ↔ 证书）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `a` | object | 是 | `KeyInput`（私钥 / 公钥 / 证书 PEM） |
| `b` | object | 是 | `KeyInput` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `match` | bool | 是否配对 |
| `reason` | string | 结论说明（枚举） |
| `aFingerprint` | string | A 的 SHA-256 指纹 |
| `bFingerprint` | string | B 的 SHA-256 指纹 |

`reason` 取值：`matched`、`public_key_mismatch`、`not_a_key_pair`、`invalid_key`、`identical`。

### 错误

`INVALID_KEY`、`INVALID_CERT`。

> `match: false` 是**业务结果**，返回 `200`。

---

## jwk-xml-convert

JWK / PEM / XML 之间的密钥格式转换。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `input` | object | 是 | `{ format, data }`，`format` ∈ `jwk` \| `pem` \| `xml` |
| `to` | string | 是 | `jwk` \| `pem` \| `xml` |
| `keyType` | string | 否 | `private` \| `public` |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `output` | object | `{ format, data }` |

### 错误

`INVALID_JWK`、`INVALID_KEY`、`INVALID_XML`、`UNSUPPORTED_FORMAT`。

> XML 采用 .NET `RSAKeyValue` 表达，**只支持 RSA**
> （见 [I-13](../../upstream/tongsuo-go-issues.md#i-13-xmlrsa-仅支持-rsa)）；
> JWK **只支持 RSA 与 EC**，不支持 OKP（Ed25519 / X25519，
> 见 [I-12](../../upstream/tongsuo-go-issues.md#i-12-jwk-不支持-okp-ed25519--x25519-与对称密钥)）。
>
> JWK 导出私钥时**必须**包含 `d` 成员，且需显式 `keyType: "private"`。

---

## ssh-keygen

生成 OpenSSH 格式密钥对与指纹（SSH Key Generator）。

### 请求

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `algorithm` | string | 否 | `Ed25519` / `RSA`，默认 `Ed25519` |
| `bits` | string | 否 | RSA 位数；仅 RSA 有效 |
| `comment` | string | 否 | 密钥注释，单行 |

### 响应

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `algorithm` | string | 算法 |
| `privateKey` | string | OpenSSH 格式私钥 |
| `publicKey` | string | OpenSSH 格式公钥 |
| `fingerprint` | string | SHA256 指纹 |
| `bits` | int | 密钥位数 |
| `comment` | string | 注释 |

### 错误

`INVALID_FIELD`。

> 产出 OpenSSH 格式（`openssh-key-v1`）；**不含**加密私钥，私钥明文只在响应中出现一次。

---

## 相关文档

- [../README.md](../README.md) —— 通用信封与通用类型
- [../errors.md](../errors.md) —— 错误码全集
- [../../standards/03-crypto.md](../../standards/03-crypto.md) —— 密码学规范
