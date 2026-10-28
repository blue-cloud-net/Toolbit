# 密码学与资源生命周期规范

> 状态：Draft ｜ 适用版本：v0.3.1 ｜ 最后更新：2026-10-28

适用包：`backend/internal/cryptosvc`（以及所有间接使用 Tongsuo 的代码）。

**本文是本项目最容易被违反、后果最严重的一套规范。**

---

## 1. 边界

- **必须**：所有 Tongsuo 调用集中在 `backend/internal/cryptosvc`。
- **禁止**：其他包 `import github.com/blue-cloud-net/tongsuo-go`。
- 服务层**不得**依赖 `net/http`、Gin、JSON 编解码。
- 服务层**不得**感知具体工具（不得出现 `toolSlug` 之类的概念）。

### 1.1 密码学运算的归属

- **必须**：分组模式链接、填充、KDF、签名与密文格式转换，**一律**经 `cryptosvc` 转调
  `tongsuo-go`。
- **禁止**：在 Toolbit 内用 Go 标准库 `crypto/*` 实现或拼装上述运算。
  标准库 `crypto/*` **仅**可用于**编码**（hex / base64 / PEM）与**结构解析**
  （不涉及密钥运算的只读查看）。
- **禁止**：为 `tongsuo-go` 的能力缺口自研兜底实现。缺口**必须**先登记到
  [../upstream/tongsuo-go-issues.md](../upstream/tongsuo-go-issues.md) 并向上游提 issue，
  本版**跳过**该能力。

该边界由 `.golangci.yml` 的 `depguard` 规则 `no-stdlib-crypto` 机器校验：
`crypto/aes`、`cipher`、`des`、`rc4`、`hmac`、`rsa`、`ecdsa`、`ed25519`、
`elliptic`、`ecdh` 在 `cryptosvc` 之外一律报错。

> 摘要（`crypto/sha256` 等）与随机源（`crypto/rand`）**不在**禁止范围内：
> 前者用于内容指纹（目录 `ETag`），后者用于非密钥标识（`requestId`）。
> 二者一旦用于密码学用途，仍属违规。

---

## 2. Session 模型

每个请求在 `backend/internal/api` 层创建**一个** `cryptosvc.Session`，并在请求结束时 `Close()`。

Session **经 `ctx` 传递**，**不**改变 `tools.Handler` 的签名：

```go
sess := cryptosvc.NewSession(ctx)
defer sess.Close()

result, err := def.Handle(sess.Context(), req)
```

服务层内部用 `cryptosvc.FromContext(ctx)` 取回当前 Session。

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
| `cryptosvc.PrivateKey` | 拥有句柄，**必须** `Close`（幂等） |
| `cryptosvc.PublicKey` | 拥有句柄，**必须** `Close`（幂等） |

> `PrivateKey.PublicKeyPEM()` 返回的是**字节**而不是 `PublicKey` 对象：
> `priv.Public()` 与私钥共享底层句柄，单独关闭会造成重复释放。

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
- **输入**：自动识别以下三种并给出正确结果：
  1. ASN.1 DER 密文（以 `30` 开头）；
  2. 裸格式 `C1C3C2`；
  3. 裸格式 `C1C2C3`。
- `C1` 带 `02` / `03` 压缩前缀**当前不支持**（上游 `asym.Format` 的 `parseSM2Raw`
  只识别未压缩点），返回 `UNSUPPORTED_FORMAT`；缺口登记为上游 issue I-09。
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

**结构预校验**（`cryptosvc.checkSignatureStructure`）在调用验签**之前**执行，
按算法区分：

| 算法 | 结构判定 |
| --- | --- |
| Ed25519 / Ed448 | 长度**固定**（64 / 114 字节） |
| RSA | 长度**等于**模长字节数（RSA-PSS 是**裸值**，**不是** DER） |
| SM2 / ECDSA | ASN.1 DER `SEQUENCE { INTEGER, INTEGER }` |

> RSA-PSS 签名是**裸值**而非 DER，与 OpenSSL 的输出不兼容（缺口 I-07）。
> 本版按 OpenSSL 行为处理：结构合法但验签失败 ⇒ `valid: false`。

---

## 7. 对称加密

- 一次性 ECB / CBC 用 `EncryptSM4CBC` 等（自带 PKCS#7）；零填充用 `*Zero` 变体。
- `padding: none` **必须**自行校验输入长度为分组长度整数倍。
- GCM **必须**使用 `EncryptSM4GCM` 或 `NewSM4GCM`；解密**必须**提供 `tag`。
- **禁止**对 GCM 复用 `(key, nonce)` 组合。
- ECB 模式**必须**产生 `INSECURE_ECB` 警告。

### 7.1 支持矩阵由本层维护

`tongsuo-go` 的按名入口对 AES 会**静默忽略** `padding` 参数（恒用 PKCS#7）。
若不显式拦截，`padding: none` 的 AES 密文将无法按预期解密。
因此 `cryptosvc.SymPaddings(algorithm, mode)` **自行维护**支持矩阵：

| 算法 | 模式 | 允许的填充 |
| --- | --- | --- |
| SM4 | `ECB` `CBC` | `pkcs7` `zero` `none` |
| SM4 | `CTR` `OFB` `CFB` `GCM` | `pkcs7` |
| AES | `ECB` `CBC` `CTR` `GCM` | `pkcs7` |

不在矩阵内的组合**必须**返回 `UNSUPPORTED_PADDING`，**禁止**放行。

### 7.2 GCM nonce 的生成与回传

- GCM **加密**未提供 `iv` 时，服务端生成 12 字节随机 nonce，并在响应的 `iv` 字段回传。
- GCM **解密**必须复用同一个 nonce，否则 tag 校验失败（`TAG_INVALID`）。
- **禁止**把随机 nonce 静默丢弃——那会让密文不可解密。

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

`sm4-file-crypt` 使用自定义容器 `TBFC`，实现在
`backend/internal/cryptosvc/tbfc.go`。

### 11.1 格式

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

| 派生量 | 取值 |
| --- | --- |
| 分块 nonce | `baseNonce[0:8] \|\| uint32be(chunkIndex)` |
| AAD | `容器头 \|\| uint32be(chunkIndex)` |

### 11.2 规则

**必须**

- 分块大小默认 1 MB，范围 4 KiB – 16 MiB，**必须**支持流式；
- nonce **必须**按块索引派生，**禁止**每块复用同一 nonce；
  基础 nonce 随机生成，因此跨容器也不复用；
- AAD **必须**绑定容器头与块索引：
  绑定容器头 → 改分块大小或基础 nonce 会认证失败；
  绑定块索引 → 块重排或删除会认证失败；
- **结束块必须带认证标签**：`length = 0` + 一个对空明文计算的 `tag`。
  因此**截断**的容器会被检测出来，而不会被误判成「文件到此为止」；
- 解密时任一块 `tag` 校验失败 → `TAG_INVALID`；
- 输入不是容器（魔数 / 版本 / 分块大小不符）→ `INVALID_FIELD`；
- 内存占用约 2×chunkSize，**不**随文件大小增长。

**禁止**

- 在 handler 里手写容器字节布局：容器读写**必须**经 `cryptosvc`；
- 把结束块当成「可选」：没有它就无法检测截断。

### 11.3 首块预校验

解密**必须**在响应头发出**之前**同步读出容器头并校验第一块
（`cryptosvc.TBFCReader.Probe`）。

**理由**：密钥不对、文件被篡改、文件不是容器——这三类错误占误用的绝大多数。
若等到流式输出中途才发现，响应头已经发出，用户只会拿到一个
「200 + 被截断的下载」，既无法判断原因，也无法重试。

因此解密失败**必须**返回 JSON 错误信封，而不是二进制流。

### 11.4 multipart 的顺序约束

multipart 表单是**顺序**的：服务端读到文件部分后就无法回头取后面的字段。

因此**必须**规定：**文件部分是最后一个 part**。

- 客户端**必须**先发所有普通字段，最后发文件；
- 服务端**必须**在读到文件部分时停止读取表单字段；
- 该约束**必须**写进 API 契约与前端实现（前端用 `FormData.append` 控制顺序）。

**禁止**使用 `ParseMultipartForm`：它会把超限文件**落盘**到临时目录，
违反「禁止落盘」的约束。**必须**用 `Request.MultipartReader()` 流式读取。

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

### 13.1 v0.3.0 已落地的向量

| 场景 | 向量 |
| --- | --- |
| SM3(`"abc"`) | 标准摘要 |
| SM4-ECB | 密钥 `0123456789ABCDEFFEDCBA9876543210`、明文同串 → `681EDF34D206965E86B3E94F536E4246`（`padding: none`） |
| SM4-CBC / SM4-GCM | 往返一致 |
| SM2 签名 / 验签 | 默认与自定义 `userId` 各一组 |
| SM2 加解密 | 往返一致 |
| RSA-OAEP / RSA-PSS / RSA-PKCS#1 | 往返一致 |
| Ed25519 / Ed448 | 往返一致 |
| ECDH P-256 + HKDF-SHA256 | 双方结果一致 |

> **陷阱**：SM4-ECB 的国标向量对应 `padding: none`，**不是** PKCS#7。
> 用 `pkcs7` 加密同一明文会得到不同密文——两者必须分别断言。

---

## 14. 禁止清单

- ❌ 在 `cryptosvc` 之外 import `tongsuo-go`。
- ❌ 用标准库 `crypto/*` 拼装密码学运算（编码与结构解析除外）。
- ❌ 为 `tongsuo-go` 的能力缺口自研兜底实现。
- ❌ 依赖 finalizer 释放句柄。
- ❌ 关闭 `priv.Public()`。
- ❌ 用 `x509.Name.Add` 处理用户输入。
- ❌ 在未 `CRL.Verify` 的情况下使用 `RevocationCheck`。
- ❌ 复用 GCM 的 `(key, nonce)`。
- ❌ 跨请求缓存密钥或证书对象。
- ❌ 把密钥材料写入日志或错误信息。
- ❌ 在密钥清零后仍继续使用缓冲区。
