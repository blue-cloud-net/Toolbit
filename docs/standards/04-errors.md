# 错误处理规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

---

## 1. 两层结构

错误由两部分表达：

| 层 | 字段 | 作用 |
| --- | --- | --- |
| 传输层 | HTTP 状态码 | 表达「请求能否被处理」的大类 |
| 语义层 | `error.code` | 表达**具体原因**，前端据此本地化 |

`error.params` 承载可插值参数（如期望长度、字段名、索引）。

**禁止**让前端依赖 `error.message` 做逻辑判断——它只是人类可读的兜底文案。

---

## 2. 业务结果 vs 错误

这是本项目最重要的错误语义区分。

| 情形 | 处理 |
| --- | --- |
| 验签不通过 | **200** + `valid: false` |
| 证书已过期 / 尚未生效 | **200** + `validity.status: "expired"` |
| 证书链不可信 | **200** + `valid: false` + `errors[]` |
| 域名与证书不匹配 | **200** + `hostnameMatch: false` |
| 证书被吊销 | **200** + `status: "revoked"` |
| OCSP 状态未知 | **200** + `certStatus: "unknown"` |
| 私钥与证书不匹配 | **400** + `KEY_MISMATCH` |
| 输入不是合法证书 | **400** + `INVALID_CERT` |
| 密钥长度不对 | **400** + `KEY_SIZE_MISMATCH` |
| 参数超出允许范围 | **400** + `INVALID_FIELD` |

**原则**：工具**成功执行**并得出了一个结论 → 200；工具**无法执行** → 4xx/5xx。

---

## 3. 错误码注册表

错误码**唯一来源**是 `backend/internal/toolerr/registry.go`。

```go
var Registry = map[Code]Spec{
    CodeInvalidHex: {
        HTTP:    http.StatusBadRequest,
        Message: "invalid hex input",
        I18nKey: "errors.INVALID_HEX",
    },
    // ...
}
```

**必须**

- 每个错误码**唯一**，全大写下划线；
- 每个错误码有：HTTP 状态、兜底英文 `message`、i18n key；
- 每个错误码在 `zh-CN` 与 `en-US` **都**有对应文案；
- 新增错误码**必须**同步 [docs/api/errors.md](../api/errors.md)（CI 校验覆盖率）。

**禁止**

- 在 handler 里临时构造错误码字符串；
- 复用语义不符的错误码（如把「参数非法」报成 `INVALID_KEY`）；
- 同一语义在不同工具使用不同错误码。

---

## 4. 命名规则

| 前缀 | 含义 | 例 |
| --- | --- | --- |
| `INVALID_*` | 输入格式 / 取值非法 | `INVALID_HEX`、`INVALID_CERT` |
| `UNSUPPORTED_*` | 功能或算法不受支持 | `UNSUPPORTED_MODE` |
| `MISSING_*` | 必填项缺失 | `MISSING_FIELD` |
| `KEY_SIZE_*` / `IV_SIZE_*` | 长度不匹配 | `KEY_SIZE_MISMATCH` |
| `PASSPHRASE_*` | 口令相关 | `PASSPHRASE_INCORRECT` |
| `TAG_*` | AEAD 标签相关 | `TAG_INVALID` |
| `DECRYPT_*` | 解密失败 | `DECRYPT_FAILED` |
| `SIGNATURE_*` | 签名结构非法 | `SIGNATURE_INVALID` |
| `TOO_LARGE` / `TOO_MANY_*` | 超出限额 | `TOO_LARGE` |
| `TIMEOUT` | 超时 | `TIMEOUT` |
| `RATE_LIMITED` | 触发限流 | `RATE_LIMITED` |
| `TOOL_NOT_FOUND` / `TOOL_DISABLED` | 工具不可用 | |
| `METHOD_NOT_ALLOWED` | 方法不匹配 | |
| `UNSUPPORTED_MEDIA_TYPE` | Content-Type 不支持 | |
| `CROSS_ORIGIN_BLOCKED` | 跨站请求被拒 | |
| `NETWORK_DISABLED` / `SSRF_BLOCKED` | 出网被禁 / 被拦 | |
| `DNS_ERROR` / `TLS_ERROR` / `HOST_UNREACHABLE` | 上游连接失败 | |
| `UPSTREAM_*` | 上游服务异常 | `UPSTREAM_TIMEOUT` |
| `INTERNAL_ERROR` | 未分类的服务端错误 | |

完整表见 [docs/api/errors.md](../api/errors.md)。

---

## 5. params

`params` 用于让前端插入**具体值**（字段名、期望长度、索引、上限）。

**必须**

- 只放**结构化、非敏感**的值；
- key 用小驼峰；
- **禁止**放入用户输入的原文（只放长度、类型、字段名这类元信息）。

**允许**

```json
{ "field": "key", "expected": 16, "actual": 24, "maxBytes": 4194304, "index": 12 }
```

**禁止**

```json
{ "value": "用户提交的密钥原文" }
```

---

## 6. 底层错误映射

Tongsuo / Go 标准库的错误**不得**直接回显。

| 来源 | 映射 |
| --- | --- |
| `asym.ErrSignature` | 验签 `valid: false`（**不是错误**） |
| `asym.ErrInvalidKey` | `INVALID_KEY` |
| `sym.ErrInvalidKeyLength` | `KEY_SIZE_MISMATCH`（带 expected / actual） |
| `sym.ErrUnknownAlgorithm` | `UNSUPPORTED_ALGORITHM` |
| `asym.ErrUnsupported` | `UNSUPPORTED_ALGORITHM` |
| `*core.OpError` | 按 library / reason 细化；兜底 `CRYPTO_OPERATION_FAILED` |
| `context.DeadlineExceeded` | `TIMEOUT`（504） |
| `context.Canceled` | `TIMEOUT`（504） |
| `http.MaxBytesError` | `TOO_LARGE`（413） |
| JSON 语法错误 | `INVALID_JSON`（400） |
| JSON 类型错误 | `INVALID_FIELD`（400，带字段名） |

**必须**：底层错误的原始文本**只**进日志（带 `requestId`），对外只给错误码与参数化文案。

---

## 7. HTTP 状态映射

| 状态 | 何时使用 |
| --- | --- |
| 200 | 成功；也用于「执行成功并得出否定结论」 |
| 400 | 输入或参数问题 |
| 403 | 跨站被拒、工具被禁用（网络类） |
| 404 | 工具不存在、端点不存在 |
| 405 | 方法不匹配 |
| 413 | 请求体超限 |
| 415 | Content-Type 不支持 |
| 429 | 触发限流（**必须**带 `Retry-After`） |
| 500 | 未分类服务端错误 |
| 501 | 端点未实现 |
| 502 | 上游失败（DNS / TLS / 重定向超限） |
| 504 | 超时 |

---

## 8. warnings

`meta.warnings` 表达**不阻断执行、但需要注意**的情况。

**必须**

- 每条 warning 有 `code` 与可选 `params`；
- warning **不改变** HTTP 状态；
- 前端**必须**以醒目但不打断的方式展示。

常用 warning：

| code | 含义 |
| --- | --- |
| `INSECURE_ECB` | 使用了 ECB 模式 |
| `NO_AUTHENTICATION` | 对称加密未提供认证（无 GCM tag） |
| `SMALL_RSA_KEY` | RSA 密钥长度小于 2048 |
| `WEAK_HASH` | 使用了 MD5 / SHA-1 |
| `EXPIRED_CERT` / `NOT_YET_VALID` | 证书有效期问题 |
| `SELF_SIGNED_CERT` / `UNTRUSTED_CHAIN` | 证书信任问题 |
| `PADDING_ORACLE_RISK` | 使用 PKCS#1 v1.5 |
| `LOSSY_FORMAT` | 格式转换有损（如 TOML 注释丢失） |
| `REGEX_DIALECT_DIFFERENCE` | 正则方言差异（PCRE vs JS） |
| `TIMEZONE_ASSUMED` | 时间戳转换假定时区 |
| `COMMENTS_LOST` | 注释被丢弃 |

---

## 9. panic 与恢复

- handler **禁止** `panic`（编程错误除外）。
- 全局 `Recovery` 中间件**必须**捕获 panic，返回 `500` + `INTERNAL_ERROR`，并记录带堆栈的日志。
- panic 信息**不得**回显给前端。

---

## 10. 校验顺序

同一请求存在多处校验时，**必须**按以下顺序短路：

1. 路由与工具存在性 → `TOOL_NOT_FOUND` / `TOOL_DISABLED`
2. 方法 → `METHOD_NOT_ALLOWED`
3. Content-Type → `UNSUPPORTED_MEDIA_TYPE`
4. Origin / Sec-Fetch-Site → `CROSS_ORIGIN_BLOCKED`
5. 限流 → `RATE_LIMITED`
6. 请求体大小 → `TOO_LARGE`
7. JSON / multipart 语法 → `INVALID_JSON` / `INVALID_FIELD`
8. 字段级语义 → `INVALID_FIELD` / `MISSING_FIELD`
9. 跨字段一致性 → `KEY_MISMATCH` 等
10. 算法与算法相关校验 → `UNSUPPORTED_*` / `KEY_SIZE_MISMATCH`

**必须**返回**第一个**失败的原因，不做「收集全部错误」。

---

## 11. 测试要求

- 每个错误码**必须**至少有一个测试断言其 `error.code`。
- 每个 handler **必须**有反例测试，断言具体的 `error.code`（**不**只断言 HTTP 状态）。
- 错误码注册表**必须**有测试：唯一性、HTTP 状态合法、i18n key 存在、文档已收录。

---

## 12. 禁止清单

- ❌ 前端依赖 `error.message` 做逻辑判断。
- ❌ 在 handler 里临时拼错误码字符串。
- ❌ 回显用户输入到错误信息。
- ❌ 把「验签失败」当成 4xx。
- ❌ 同一语义在不同工具用不同错误码。
- ❌ 新增错误码但不同步 i18n 与文档。
