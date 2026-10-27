# 密码学与资源生命周期规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

适用包：`internal/cryptosvc`（以及所有间接使用 Tongsuo 的代码）。

**本文是本项目最容易被违反、后果最严重的一套规范。**

---

## 1. 边界

- **必须**：所有 Tongsuo 调用集中在 `internal/cryptosvc`。
- **禁止**：其他包 `import github.com/blue-cloud-net/tongsuo-go`。
- 服务层**不得**依赖 `net/http`、Gin、JSON 编解码。
- 服务层**不得**感知具体工具（不得出现 `toolSlug` 之类的概念）。

---

## 2. Session 模型

每个请求在 `internal/api` 层创建**一个** `cryptosvc.Session`，并在请求结束时 `Close()`。

```go
type Session struct {
    ctx      context.Context
    warnings []toolerr.Warning
    closers  []func() error
    sems     *semaphores
}
```

**必须**

- `Close()` **幂等**，按**注册逆序**释放；
- 所有原生句柄（`asym.PrivateKey`、`x509.Certificate`、`x509.CRL`、`x509.Store`、`x509.Response`）**必须**注册到 Session；
- 释放**必须**显式调用，**禁止**依赖 `runtime.SetFinalizer`；
- Session 结束前**必须**清零密钥与明文缓冲区。

**禁止**

- 把句柄存到包级变量或跨请求缓存；
- 把句柄返回给 handler 之外的地方长期持有；
- 在 Session 之外手动 `Close` 同一个句柄。

---

## 3. 句柄所有权要点

| 对象 | 规则 |
| --- | --- |
| `priv.Public()` | **与私钥共享底层句柄**。**禁止**单独 `Close`；只关闭私钥即可 |
| `x509.Certificate` | 拥有句柄，**必须** `Close` |
| `x509.CRL` | 拥有句柄，**必须** `Close` |
| `x509.Store` | 拥有句柄，**必须** `Close`；释放后链验证会失败 |
| `x509.Response`（OCSP） | 拥有句柄，**必须** `Close` |
| `x509.CertificateRequest` | **无公开 `Close`**，只能靠 finalizer。**禁止**跨请求持有 |
| `sym` 的 `cipher.AEAD` | **持有 key 切片引用**。使用后**必须**清零 key 切片 |
| 私钥 PEM 源缓冲 | 加载后**必须**立即清零 |

**禁止**在 `Session.Close()` 之后继续使用任何句柄（会得到明确错误，而非崩溃）。

---

## 4. 算法默认值

| 项目 | 默认值 | 备注 |
| --- | --- | --- |
| SM2 密文顺序 | `C1C3C2` | GM/T 0009-2012 之后的标准顺序 |
| SM2 密文格式 | 裸格式 + `04` 前缀 | 兼容 BouncyCastle 等主流实现 |
| SM2 签名格式 | ASN.1 DER | 与 `openssl` 一致 |
| SM2 `userId` | `1234567812345678` | GM/T 0003-2012 规定值；两端必须一致 |
| SM4 填充 | `pkcs7` | `zero` / `none` 需显式选择 |
| SM4 / AES GCM nonce | 12 字节 | |
| SM4 / AES GCM tag | 16 字节 | |
| SM4 / AES CBC / CTR / OFB / CFB IV | 16 字节 | |
| RSA 密钥长度 | 2048 | 允许 1024–4096；<2048 产生 `SMALL_RSA_KEY` |
| RSA OAEP 摘要 | SHA-256 | 加解密**必须**使用同一摘要 |
| ECDSA 摘要 | SHA-256 | 当前库固定；需其它摘要时自行预哈希 |
| 证书有效期上限 | 30 年 | 超过返回 `NOT_AFTER_BEFORE_NOW` 或参数错误 |
| 证书序列号 | CSPRNG 随机正整数 | 用户显式指定时优先 |

---

## 5. SM2 密文格式处理

SM2 密文的 `C1C3C2` / `C1C2C3` 顺序与 `04` 前缀在各实现间不统一，是联调失败的首要原因。

**必须**

- **输出**：`EncryptWithOrder(pub, data, "c1c3c2")`，裸格式，`C1` 未压缩（`04` 前缀）。
- **输入**：自动识别以下四种并给出正确结果：
  1. ASN.1 DER 密文（以 `30` 开头）；
  2. 裸格式 `C1C3C2`；
  3. 裸格式 `C1C2C3`；
  4. `C1` 带 `02` / `03` 压缩前缀。
- 转换统一走 `asym.Format(ct, from, to)`，`from` / `to` 取 `der` / `c1c3c2` / `c1c2c3`。
- 解密失败**必须**区分「格式不认识」与「私钥不匹配」两种错误码。

**禁止**在 handler 里手写密文顺序的字节拼接。

---

## 6. 签名与验签

- SM2 签名走 `Sign` / `SignWithID`（DER）；自定义 `userId` 时**必须**两端一致。
- **禁止**对 Ed25519 / Ed448 的消息**预哈希**（纯签名语义，预哈希会破坏互通）。
- RSA-PSS 的 `saltLen` 与摘要**必须**在签名与验签时一致。
- 验签失败**必须**返回 `valid: false`（业务结果），**禁止**当作错误抛出。
- 只有签名**结构非法**（无法解析为 DER / 长度不符）才返回 `SIGNATURE_INVALID`。

---

## 7. 对称加密

- 一次性 ECB / CBC 用 `EncryptSM4CBC` 等（自带 PKCS#7）；零填充用 `*Zero` 变体。
- `padding: none` **必须**自行校验输入长度为分组长度整数倍。
- GCM **必须**使用 `EncryptSM4GCM` 或 `NewSM4GCM`；解密**必须**提供 `tag`。
- **禁止**对 GCM 复用 `(key, nonce)` 组合。
- ECB 模式**必须**产生 `INSECURE_ECB` 警告。

---

## 8. PKI

### 8.1 DN 构造

- **必须**：用户输入的 DN 字段一律用 `x509.Name.TryAdd`。
- **禁止**：`x509.Name.Add` 会因非法字段名或超长值 **panic**。
- 允许的字段白名单：`CN` `C` `O` `OU` `L` `ST` `serialNumber` `emailAddress` `title` `SN` `givenName` `initials` `generationQualifier` `pseudonym` `businessCategory` `DC`。

### 8.2 证书签发

- 自签用 `CreateSelfSigned`（自动补 SKID / AKID）。
- CA 签发用 `CreateCertificate`（**必须**手工补 SKID / AKID）。
- CSR 签发用 `IssueCertificateFromRequest`；`CopyRequestExtensions` **默认 false**；开启时**只**拷贝白名单（SAN / KU / EKU / BC）——这与 `openssl ca` 的 `copy_extensions none` 一致，是防止请求方自选 SAN 的安全底线。
- 签名前**必须**用 `asym.Match` 校验签发者私钥与证书公钥是否配对。

### 8.3 链验证

- 每请求**必须**新建 `x509.Store`（`SetTime` 会影响该 store 后续所有验证）。
- 验证时刻回拨**必须**谨慎使用：仅在用户显式指定历史时刻时使用。
- 域名匹配用 `Certificate.VerifyHostname`；它**只**做名字匹配，**不**验证链与有效期。
- 链验证失败的 `VerifyError.Code` / `Depth` / `Message` **必须**映射为结构化错误项。

### 8.4 CRL 与 OCSP

- `RevocationCheck` **不**校验 CRL 的签名与有效期。**必须**先 `CRL.Verify(issuer)`，再调用吊销检查。
- CRL 原因码取 RFC 5280 §5.3.1；码值 `7` 为保留值，**禁止**使用。
- CRL 签发要求 CA 证书带 SKID（CRL 的 AKID 取自该扩展）。
- OCSP 解析**必须**同时提供 `cert` 与 `issuer`；响应级失败（`Status != 0`）时只展示响应级字段。

---

## 9. 限额

| 项目 | 上限 | 超限错误 |
| --- | --- | --- |
| RSA 密钥长度 | 1024–4096 | `INVALID_FIELD` |
| PBKDF2 迭代 | 1,000,000 | `TOO_MANY_ITEMS` |
| scrypt `N` | 2²⁰ | `TOO_MANY_ITEMS` |
| Argon2 内存 | 256 MB | `TOO_LARGE` |
| KDF 输出长度 | 1 MB | `TOO_LARGE` |
| 随机字节数 | 1 MB | `TOO_LARGE` |
| PEM / 证书输入 | 1 MB | `TOO_LARGE` |
| 证书链张数 | 64 | `TOO_MANY_ITEMS` |
| CRL 吊销条目 | 10,000 | `TOO_MANY_ITEMS` |
| ASN.1 解析深度 | 64 | `TOO_LARGE` |
| ASN.1 节点数 | 10,000 | `TOO_LARGE` |

---

## 10. 密钥清零

- 私钥、口令、明文缓冲在**使用完毕后立即清零**。
- Go 编译器**可能**消除「看似无副作用」的清零循环。清零代码**必须**配合 `runtime.KeepAlive` 保证不被优化掉。
- **禁止**把密钥材料放进 `fmt.Stringer` / 结构体 `String()` 方法。
- **禁止**把密钥放进 panic 信息或错误文本。

---

## 11. 文件加密容器

`sm4-file-crypt` 使用自定义容器 `TBFC`：

| 部分 | 内容 |
| --- | --- |
| Header | `TBFC`(4) + version(1) + flags(1) + chunkSize(4) + baseNonce(12) |
| Chunk | `uint32 length` + ciphertext + `tag(16)`；末块 `length = 0` |
| Nonce | `baseNonce[0:8]` + `uint32be(chunkIndex)` |
| AAD | header + `uint32be(chunkIndex)` |

**必须**

- 分块大小默认 1 MB，**必须**支持流式；
- nonce **必须**按块索引派生，**禁止**每块复用同一 nonce；
- AAD **必须**绑定块索引，防止块重排；
- 解密时任一块 `tag` 校验失败 → `DECRYPT_FAILED`；
- 缓冲池复用，内存占用约 2×chunkSize。

---

## 12. 并发

- 密钥生成、证书签发、PKCS#12 打包**必须**经信号量限并发（≤ 2）。
- 信号量获取**必须**响应 `ctx`。
- cgo 调用**不可中断**：超时后请求返回 `504`，底层继续跑完。**必须**记录 orphan 日志，并确保信号量在任务结束时释放。
- **禁止**在 Session 之外缓存任何 Tongsuo 对象。

---

## 13. 测试要求

**必须**包含：

- SM3(`"abc"`) 标准向量；
- SM4-ECB：密钥 `0123456789ABCDEFFEDCBA9876543210`、明文同串 → 密文 `681EDF34D206965E86B3E94F536E4246`；
- SM4-CBC / SM4-GCM 往返；
- SM2：签名验签（默认与自定义 `userId`）、DER/raw 互转、`C1C2C3`/`C1C3C2` 互转、加解密往返；
- RSA：OAEP / PSS 往返、最大明文边界、错误口令；
- 全算法密钥生成 + PEM 往返 + 加密私钥；
- PKI：CSR → 证书 → 链验证 → hostname；自签 CA → 叶证书；CRL 生成 → 解析 → 验签 → 吊销检查；PKCS#12 打包 / 解析 / 改密；PKCS#7；
- 生命周期：`Close` 幂等、`Public` 别名不重复关闭、`-race` 并发；
- Fuzz：PEM / DER 输入、ASN.1 视图、SM2 格式转换。

**禁止**在测试中使用真实生产密钥；示例数据**必须**来自测试向量。

---

## 14. 禁止清单

- ❌ 在 `cryptosvc` 之外 import `tongsuo-go`。
- ❌ 依赖 finalizer 释放句柄。
- ❌ 关闭 `priv.Public()`。
- ❌ 用 `x509.Name.Add` 处理用户输入。
- ❌ 在未 `CRL.Verify` 的情况下使用 `RevocationCheck`。
- ❌ 复用 GCM 的 `(key, nonce)`。
- ❌ 跨请求缓存密钥或证书对象。
- ❌ 把密钥材料写入日志或错误信息。
- ❌ 在密钥清零后仍继续使用缓冲区。
