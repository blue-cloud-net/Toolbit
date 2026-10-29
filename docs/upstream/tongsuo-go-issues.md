# 上游依赖缺口台账（tongsuo-go）

> 状态：Draft ｜ 适用版本：v0.5.0 ｜ 最后更新：2026-10-29

本文登记 Toolbit 在使用 [`tongsuo-go`](https://github.com/blue-cloud-net/tongsuo-go) 时遇到的**能力缺口**。

每条缺口都对应一份可直接提交给上游的 issue 草稿。Toolbit 的应对策略是：

> **上游未实现的能力一律跳过**——不进 API 契约、不留枚举分支、不自研兜底实现。
> 待上游补齐后，按 MINOR 版本把该能力加回契约。

## 0. 核实环境

| 项 | 值 |
| --- | --- |
| tongsuo-go | v0.5.0 |
| Tongsuo | 8.5.0-pre1（OpenSSL 兼容层 3.5.4） |
| 平台 | Linux / x86_64 / Go 1.27 |
| 核实方式 | 直接阅读模块源码 + 运行最小复现程序 |

## 1. 缺口总览

| # | 缺口 | 影响的能力 | Toolbit 承接版本 |
| --- | --- | --- | --- |
| I-01 | `sym` 缺 AES 的 OFB / CFB | `symmetric-crypt` | 0.3.x |
| I-02 | `sym` 缺 DES / 3DES | `symmetric-crypt` | 0.4.0 |
| I-03 | `sym` 未公布 `Names()` × `Options.Padding` 支持矩阵 | `symmetric-crypt` | 0.3.x |
| I-04 | `sym.Encrypt` 对 AES **静默忽略** `Options.Padding` | `symmetric-crypt` | 0.3.x |
| I-05 | `asym` 缺 ECIES | `asymmetric-crypt` | 待定 |
| I-06 | RSA 的签名 / OAEP 不支持 `sm3` 摘要 | `signature` | 0.3.x |
| I-07 | `asym` 缺签名 DER ↔ 裸值互转 | `signature` | 0.3.x |
| I-08 | RSA-PSS 签名是**裸值**，与 OpenSSL 不互通 | `signature` | 0.3.x |
| I-09 | `ecdh` 的密钥对象**没有 `Close()`** | `ecdh` | 0.3.x |
| I-10 | SM2 **压缩点**裸密文无法解密 | `asymmetric-crypt` | 0.4.0 |
| I-11 | `pkcs/pkcs7` 仅支持**证书袋**，无签名 / 加密 | `p7b-pack` `p7b-extract` | 0.6.0 |
| I-12 | `jwk` 不支持 OKP（Ed25519 / X25519）与对称密钥 | `jwk-xml-convert` | 0.4.0 |
| I-13 | `xml/rsa` **仅支持 RSA** | `jwk-xml-convert` | 0.4.0 |
| I-14 | `asym` 缺 RSA **PKCS#1 公钥**的导出 / 加载入口 | `keygen` `key-convert` | 0.4.0 |
| I-15 | `x509.CertificateRequest` **没有公开的 `Close()`** | `csr-parse` | 0.5.0 |
| I-16 | `x509/asn1.Parse` **没有深度 / 节点数上限参数** | `asn1-view` | 0.5.0 |
| I-17 | `x509.Store` **没有加载系统信任库**的入口 | `cert-chain-verify` | 0.5.0 |
| I-18 | `x509.RevokedEntry.Serial` 是 `int64`，超 8 字节返回 `-1` | `crl-parse` | 0.5.0 |

---

## I-01 `sym` 缺 AES 的 OFB / CFB 模式

**现象**　`sym` 对 SM4 提供了 `EncryptSM4OFB` / `EncryptSM4CFB` 及其解密入口，
但 AES 只有 `EncryptAESECB` / `CBC` / `CTR` / `GCM`。按名分发路径 `sym.Names()` 中
也只存在 `AES-{128,192,256}-{CBC,CTR,ECB,GCM}`。

**期望**　补齐 `EncryptAESOFB` / `DecryptAESOFB` / `EncryptAESCFB` / `DecryptAESCFB`
（IV 长度为 `BlockSize`），并纳入 `sym.Names()`，使 AES 与 SM4 的模式集合对齐。

**验收**

- `sym.Names()` 含 `AES-128-OFB`、`AES-192-CFB` 等条目；
- 往返测试通过；
- `-tags tongsuocli` 下与 `openssl enc -aes-128-ofb` / `-aes-256-cfb` 交叉验证一致。

**下游影响**　Toolbit 的 `symmetric-crypt` 契约里 AES 只能提供 ECB / CBC / CTR / GCM。

---

## I-02 `sym` 缺 DES / 3DES

**现象**　`sym` 仅覆盖 SM4 / AES / ChaCha20，没有任何 DES 或 3DES 入口，
`sym.Names()` 中亦无对应条目。

**期望**　补 `DES-ECB` / `DES-CBC` / `DES-EDE3-ECB` / `DES-EDE3-CBC` 的一次性入口
与按名分发，`Options.Padding` 支持 `pkcs7` / `zero`。ECB 的安全提示沿用现有注释风格。

**验收**

- `sym.Names()` 含上述四个名字；
- 与 `openssl enc -des-cbc` / `openssl enc -des-ede3-cbc` 结果一致；
- 文档注明仅用于兼容既有数据。

**下游影响**　Toolbit 的 `symmetric-crypt` 当前只提供 SM4 与 AES。

---

## I-03 `sym` 未公布 `Names()` × `Options.Padding` 支持矩阵

**现象**　`sym.Encrypt(name, key, iv, data, opts)` 接受 `Options.Padding`
（`pkcs7` / `zero` / `none`），但包文档未说明**哪些名字 × 哪些填充**是合法组合。
调用方只能在运行期靠错误反推，无法在请求校验阶段判定。

**期望**　二选一：

- (a) 在 `sym` 包文档中给出完整矩阵表；或
- (b) 提供静态查询入口，如 `sym.Supports(name string, padding string) (bool, error)`。

**验收**　任意非法组合可在**不提交数据**的前提下被判定。

**下游影响**　Toolbit 目前在 `cryptosvc` 自行维护该矩阵。

---

## I-04 `sym.Encrypt` 对 AES **静默忽略** `Options.Padding`

**现象**　`sym.Encrypt` 的按名路径中，AES 分支直接调用
`EncryptAESCBC` / `EncryptAESCTR` / `EncryptAESECB`，**完全忽略** `opts.Padding`；
而 SM4 的 ECB / CBC 走 `encryptSM4WithPad(..., pad)`，真正应用了填充。
因此 `padding: "none"` 或 `"zero"` 传给 AES 时会被静默当成 pkcs7 处理。

**影响**　调用方以为拿到的是无填充密文，实际却是 pkcs7 密文；若另一端按
`none` 解密，必然失败，且失败原因被掩盖。

**期望**　AES 分支要么应用 `opts.Padding`，要么对不受支持的填充值**显式报错**
（返回包装 `sym.ErrUnsupported` 的错误），不得静默忽略。

**验收**　`sym.Encrypt("AES-128-CBC", key, iv, data, &sym.Options{Padding: "none"})`
要么产出与 `none` 一致的密文，要么返回明确错误。

**下游影响**　Toolbit 已在 `cryptosvc` 层显式拦截：只有 SM4 的 ECB / CBC 接受
`zero` / `none`，其余组合返回 `UNSUPPORTED_PADDING`。

---

## I-05 `asym` 缺 ECIES

**现象**　`asym` 覆盖 SM2 / RSA / EC / Ed25519 / Ed448 / X25519 / X448，
`ecdh` 只做共享密钥派生，库内没有 ECIES（ECDH + KDF + 对称 AEAD 的组合封装）。

**期望**　二选一——

- (a) 提供 `EncryptECIES` / `DecryptECIES`（建议默认 P-256/384/521 搭配
  HKDF-SHA256 与 AES-256-GCM，并给出自描述的容器头）；
- (b) 在 README 与 roadmap 中写明「不计划提供 ECIES，请自行组合
  `ecdh` + `kdf` + `sym`」。

**说明**　若选 (b)，下游会按「组合使用」实现；但这与下游「密码学运算不自行拼装」
的约束冲突，需要上游明确表态。

**下游影响**　Toolbit 的 `asymmetric-crypt` 当前只提供 RSA 与 SM2。

---

## I-06 RSA 的签名 / OAEP 不支持 `sm3` 摘要

**现象**　`SignPSS` / `SignPKCS1v15` / `VerifyPSS` / `VerifyPKCS1v15` 与
`EncryptOAEP` / `DecryptOAEP` 的 `digest` 参数只接受
`sha1` / `sha224` / `sha256` / `sha384` / `sha512`。库内已有 SM3 摘要与
SM2withSM3，唯独 RSA 缺 `sm3`。

**期望**　上述入口接受 `"sm3"`（RSA-PKCS1v15-SM3 / RSA-PSS-SM3 / RSA-OAEP-SM3），
并补标准向量测试。

**验收**　RSA-SM3 签名与 `openssl dgst -sm3 -sign` 的结果可互验；
`Verify*` 对摘要不匹配的情况返回明确错误。

**下游影响**　Toolbit 的 `signature` 中 RSA 的 `hash` 仅有 SHA256 / SHA384 / SHA512。

---

## I-07 `asym` 缺签名 DER ↔ 裸值互转

**现象**　`asym.Format(ct, from, to)` 只处理 **SM2 密文**（`der` / `c1c3c2` / `c1c2c3`）。
签名没有与裸 `r‖s` 互转的入口。

**期望**　新增 `asym.FormatSignature(sig []byte, from, to string) ([]byte, error)`，
`from` / `to` ∈ `der` / `raw`；或按算法提供 `SignRaw` / `VerifyRaw`。

**验收**　`der → raw → der` 往返字节一致；SM2 与 ECDSA 各有用例；
与 `openssl` 的 `-sigopt` 输出可互验。

**下游影响**　Toolbit 的 `signature` 暂不支持裸 `r‖s` 形式的签名。

---

## I-08 RSA-PSS 签名是**裸值**，与 OpenSSL 不互通

**现象**　`asym.SignPSS` / `asym.SignPKCS1v15` 返回的是**裸的 RSA 签名值**
（长度等于模数字节数），**不是** OpenSSL 惯用的 ASN.1 DER
`SEQUENCE { INTEGER r, INTEGER s }`。实测 2048 位 RSA 的 PSS 签名首字节为
`99 c9 c2 62 …`，可确认是裸值。

与之相对，SM2 与 ECDSA 的签名**是** DER（首字节 `0x30`）。

**影响**　OpenSSL / Java / Python 的 RSA-PSS 验签接口期望 DER 输入，
因此用本库产出的 RSA 签名无法被这些实现直接验签，影响跨语言互通。

**期望**　`SignPSS` / `VerifyPSS`（以及 PKCS#1 v1.5 的对应入口）按 OpenSSL 惯例
产出 / 接受 ASN.1 DER；如需保留裸值语义，请另设显式入口（如 `SignPSSRaw`），
并在包文档中说明两种格式。

**验收**　本库产出的 RSA-PSS 签名可被 `openssl dgst -verify` 直接验证；
`VerifyPSS` 同时接受 DER 与裸值（或明确只接受其一并写入文档）。

**下游影响**　Toolbit 已在 `cryptosvc` 中按实际格式做结构校验：
RSA 签名按「长度 == 模数字节数」判定，SM2 / ECDSA 按 DER 判定。

---

## I-09 `ecdh` 的密钥对象**没有 `Close()`**

**现象**　`ecdh.PrivateKey` 与 `ecdh.PublicKey` 持有原生 `EVP_PKEY` 句柄
（内部经 `EVP_PKEY_dup` 复制），但包内**没有**导出 `Close` 方法，
只有 `MarshalPEM` / `Curve` / `ECDH` 等方法。

**影响**　调用方只能依赖 Go 的 GC 与 finalizer 释放句柄。对短生命周期对象
尚可接受，但无法满足「显式释放、禁止依赖 finalizer」这类严格的句柄管理要求，
也无法在请求边界精确控制释放时机。

**期望**　为 `ecdh.PrivateKey` 与 `ecdh.PublicKey` 提供 `Close() error`，
幂等且在释放后返回错误（与 `asym` 的 `Close` 语义一致）。

**验收**　`Close` 幂等；释放后再调用 `ECDH` / `MarshalPEM` 返回错误而非崩溃。

**下游影响**　Toolbit 的 `ecdh` 工具每次请求会留下 2 个待 GC 回收的句柄副本，
量级可控但**不满足**项目「显式释放」的规范要求。

---

## I-10 SM2 **压缩点**裸密文无法解密

**现象**　`asym.Format` 的 `parseSM2Raw` 只在首字节为 `0x04`（未压缩点）时
填充 `x` / `y`；首字节为 `0x02` / `0x03`（压缩点）时虽然 `sm2C1Len` 能正确
返回 33，但坐标**不会被解析**。随后 `buildSM2Raw` 用空的 `x` / `y` 重建 C1，
坐标就此丢失，`DecryptWithOrder` 随即失败。

**期望**　`Format` 支持压缩点输入：解析时按曲线参数解出 `x` / `y`，
输出时按 `to` 指定的表示（未压缩 / 压缩）重建；或在文档中明确
「压缩点不受支持」并让 `Format` 对压缩点输入返回明确错误。

**验收**　`c1c3c2`（压缩）→ `der` → `c1c3c2`（未压缩）往返后，
用未压缩形式可成功解密；或在文档与错误中明确拒绝压缩点。

**下游影响**　Toolbit 的 `asymmetric-crypt` 对压缩点密文返回
`UNSUPPORTED_FORMAT`（不误报为「私钥不匹配」）。

---

## I-11 `pkcs/pkcs7` 仅支持**证书袋**，无签名 / 加密

**现象**　`pkcs/pkcs7` 包的公开入口只有三个：`Build(certs)`（仅证书）、
`MarshalPEM(der)` 与 `Extract(data)`（仅取证书）。包的 API 清单里
`Sign` / `Verify` / `Encrypt` / `Decrypt` 均标注为「规划中（0.7.0）」。
`internal/core.PKCS7` 也只实现了 `NewPKCS7SignedData` 的 certificates-only 形态。

**期望**　补齐 PKCS#7 的签名与加密流程：`Sign` / `Verify` / `Encrypt` / `Decrypt`。

**验收**

- 能产出带 signerInfo 的 SignedData 并验签；
- 能产出 EnvelopedData 并解密；
- 与 `openssl smime -sign` / `-encrypt` 互通。

**下游影响**　Toolbit 的 `p7b-pack` / `p7b-extract` **只做证书袋**（certs-only）。
请求中出现「签名 / 加密容器」语义时返回 `UNSUPPORTED_FORMAT`。

---

## I-12 `jwk` 不支持 OKP（Ed25519 / X25519）与对称密钥

**现象**　`jwk.Parse` 只接受 `kty` 为 `RSA` 与 `EC`，其它一律报
`unsupported kty`；`jwk.MarshalKey` 对 OKP 类型同样不支持。
包文档也写明「目前只支持 RSA 与 EC 参数导出」。

**期望**　补齐 OKP（Ed25519 / Ed448 / X25519 / X448，RFC 8037 / 7748）
与对称密钥（`kty: "oct"`）的导入导出。

**验收**

- `kty: "OKP"` 的 JWK 与 PEM 双向往返一致；
- `kty: "oct"` 的对称密钥可导出与导入；
- 与 `openssl` / RFC 7517 附录的样例一致。

**下游影响**　Toolbit 的 `jwk-xml-convert` 的 JWK 侧**只支持 RSA 与 EC**；
Ed25519 / X448 等输入返回 `UNSUPPORTED_FORMAT`。

---

## I-13 `xml/rsa` **仅支持 RSA**

**现象**　`xml/rsa` 包的四个入口（`MarshalPrivate` / `MarshalPublic` /
`UnmarshalPrivate` / `UnmarshalPublic`）都以 `asym.KeyParams` 的
`Type == "RSA"` 为前提；非 RSA 密钥直接返回
`not an RSA private key` / `not an RSA public key`。

**期望**　要么补齐 EC / SM2 的 XML 表达（.NET 的 `RSAKeyValue` 本身是
RSA 专用，建议另起 `ECKeyValue` 规范），要么在包文档中显式写明
「本包只处理 RSA」。

**验收**　包文档或 README 明确写出适用算法范围。

**下游影响**　Toolbit 的 `jwk-xml-convert` 的 XML 目标**仅支持 RSA**；
EC / SM2 输入返回 `UNSUPPORTED_FORMAT`。

---

## I-14 `asym` 缺 RSA **PKCS#1 公钥**的导出 / 加载入口

**现象**　`asym` 提供了 `MarshalRSAPrivateKeyPKCS1PEM`（PKCS#1 **私钥**），
但**没有**对应的 PKCS#1 **公钥**入口（`RSA PUBLIC KEY` 形态）。
`asym.LoadPublicKeyPEM` 也只接受 SPKI（`PUBLIC KEY`），
无法加载 `-----BEGIN RSA PUBLIC KEY-----` 块。

**期望**　补 `MarshalRSAPublicKeyPKCS1PEM` 与对应的加载入口；
或在包文档中写明「PKCS#1 公钥不支持」，让调用方不必试错。

**验收**　同一把 RSA 密钥的 PKCS#1 公钥 PEM 可被加载并与私钥 `Match` 成功。

**下游影响**　Toolbit 的 `keygen` 的 `publicFormat: pkcs1` 与
`key-convert` 的 `to: pkcs1`（公钥输入）返回 `UNSUPPORTED_FORMAT`；
PEM 块标签为 `RSA PUBLIC KEY` 的输入同样被显式拒绝。

---

## I-15 `x509.CertificateRequest` **没有公开的 `Close()`**

**现象**　`x509.Certificate` 与 `x509.CRL` 都提供了幂等的 `Close()`，
但 `x509.CertificateRequest` **没有**。CSR 的底层句柄只能靠 `finalizer` 回收，
这与 Toolbit「原生句柄必须显式释放」的硬性要求冲突
（见 [AGENTS.md §4.7](../AGENTS.md)）。

**期望**　给 `CertificateRequest` 补一个幂等的 `Close()`，
与 `Certificate.Close()` / `CRL.Close()` 行为一致。

**验收**

- `req.Close()` 可重复调用而不 panic；
- `req.Close()` 后再调 `MarshalDER()` / `Verify()` 会报错，而不是拿到脏数据；
- 与 `Certificate` 的句柄释放语义对齐。

**下游影响**　Toolbit 的 `csr-parse` 只能依赖上游 `finalizer` 释放 CSR 句柄，
**无法**显式释放；这是 v0.5.0 唯一一处句柄释放例外，已在
[密码学规范 §8.6](../standards/03-crypto.md) 登记。

---

## I-16 `x509/asn1.Parse` **没有深度 / 节点数上限参数**

**现象**　`asn1.Parse(der []byte) (*Node, error)` **只**接受字节，
既没有深度上限也没有节点数上限；恶意构造的深层嵌套 DER
（例如几万个嵌套 SEQUENCE）会让它递归展开并占用大量内存。

**期望**　补 `ParseWithLimits(der []byte, maxDepth, maxNodes int) (*Node, error)`，
或为 `Parse` 增加带上限的变体；在超出上限时返回可识别的错误。

**验收**

- 超过深度上限时返回错误而不是展开到底；
- 超过节点数上限时返回错误；
- 默认 `Parse` 行为保持不变。

**下游影响**　Toolbit 的 `asn1-view` 在**调用上游之前**自行测量字节长度与节点数
（限额见 [密码学规范 §9](../standards/03-crypto.md)），超出即返回 `TOO_LARGE`；
这是**调用方侧的保护**，不是上游修复。

---

## I-17 `x509.Store` **没有加载系统信任库**的入口

**现象**　`x509.NewStore()` 只能通过 `AddCert` / `AddCRL` 逐个添加对象，
没有对应 OpenSSL `X509_STORE_set_default_paths()` 的入口，
因此**无法**把操作系统信任库（`/etc/ssl/certs` 等）加载进 store。

**期望**　补 `NewSystemStore()`（或 `SetDefaultPaths()`），
使调用方能像 `openssl verify` 一样默认使用系统信任库。

**验收**

- `NewSystemStore()` 创建的 store 能验证一条系统信任链；
- 不带参数的 `openssl verify` 与上游入口的结论一致。

**下游影响**　Toolbit 的 `cert-chain-verify` **没有系统信任根可用**：
`trustedRoots` 为空时把链内的自签名根当作**隐含信任锚**，
此时 `valid: true` 只说明链内自洽，响应里的 `implicitTrust` 为 `true`。
这一语义已在 API 契约与产品文档中显式说明。

---

## I-18 `x509.RevokedEntry.Serial` 是 `int64`，超 8 字节返回 `-1`

**现象**　`CRL.RevokedEntries()` 返回的 `RevokedEntry.Serial` 是 `int64`，
底层走 `ASN1_INTEGER_get`；当被吊销证书的序列号长度超过 8 字节时，
该函数返回 `-1`，于是**所有**超长序列号都被折叠成同一个值 `-1`，
无法区分。

**期望**　把序列号改为**字符串**（或 `*big.Int`）承载，
至少要能无损表达 20 字节的序列号。

**验收**

- 一个序列号为 20 字节的吊销条目能读出完整值；
- 与 `openssl crl -text` 的输出逐项一致。

**下游影响**　Toolbit 的 `crl-parse` 在序列号为负（溢出标记）时输出**空字符串**，
**不**输出错误的 `-1`；该行为已写入 API 契约的注意事项。

---

## 2. 维护约定

- 上游发布新版本后，**逐条**复核本台账：已修复的条目直接删除，并在
  对应的 Toolbit 版本把该能力加回 API 契约；
- 新发现的缺口追加到本文件，并在同一次提交里向 upstream 提交 issue；
- 本文件只登记**能力缺口**，不登记实现偏好或风格建议。
