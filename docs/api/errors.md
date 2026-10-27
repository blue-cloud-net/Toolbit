# 错误码与警告

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

错误码的**唯一事实来源**是 `internal/toolerr/registry.go`。本文是它的镜像文档，由 `scripts/check-docs.mjs` 校验覆盖率。

---

## 1. 约定

- 错误码全大写、下划线分隔，全局唯一。
- HTTP 状态表达「请求能否被处理」；`error.code` 表达**具体原因**。
- `error.params` 承载可插值参数（字段名、期望长度、索引、上限）。
- `error.message` 是**英文兜底文案**，**禁止**包含用户输入。
- 前端**必须**按 `error.code` 本地化，**禁止**依赖 `message` 做逻辑判断。

---

## 2. 请求与格式（Request & Format）

| 错误码 | HTTP | 含义 | 常见 `params` |
| --- | --- | --- | --- |
| `INVALID_REQUEST` | 400 | 请求体结构非法 | — |
| `MISSING_FIELD` | 400 | 缺少必填字段 | `field` |
| `INVALID_FIELD` | 400 | 字段取值非法 | `field`, `expected`, `actual` |
| `INVALID_ENCODING` | 400 | 编码名不支持 | `encoding` |
| `INVALID_HEX` | 400 | 十六进制串非法 | `index` |
| `INVALID_BASE64` | 400 | Base64 串非法 | `index` |
| `INVALID_PEM` | 400 | PEM 解析失败 | `label` |
| `INVALID_DER` | 400 | DER 解析失败 | — |
| `INVALID_KEY` | 400 | 密钥非法或无法加载 | `format` |
| `INVALID_CERT` | 400 | 证书非法 | — |
| `INVALID_CSR` | 400 | CSR 非法 | — |
| `INVALID_CRL` | 400 | CRL 非法 | — |
| `INVALID_PKCS7` | 400 | PKCS#7 非法 | — |
| `INVALID_PKCS12` | 400 | PKCS#12 非法 | — |
| `INVALID_JWK` | 400 | JWK 非法 | — |
| `INVALID_JSON` | 400 | JSON 语法错误 | `line`, `column` |
| `INVALID_YAML` | 400 | YAML 语法错误 | `line`, `column` |
| `INVALID_TOML` | 400 | TOML 语法错误 | `line`, `column` |
| `INVALID_XML` | 400 | XML 语法错误 | `line`, `column` |
| `INVALID_JWT` | 400 | JWT 结构非法 | `part` |
| `INVALID_CIDR` | 400 | CIDR 非法 | `value` |
| `INVALID_REGEX` | 400 | 正则编译失败 | `pattern` |
| `INVALID_CRON` | 400 | Crontab 表达式非法 | `field` |
| `INVALID_CHMOD` | 400 | 权限值非法 | `value` |
| `INVALID_TIME` | 400 | 时间无法解析 | `value` |
| `INVALID_DOMAIN` | 400 | 域名格式非法 | `value` |

---

## 3. 算法与参数（Algorithm & Parameter）

| 错误码 | HTTP | 含义 | 常见 `params` |
| --- | --- | --- | --- |
| `UNSUPPORTED_ALGORITHM` | 400 | 算法不支持 | `algorithm`, `supported` |
| `UNSUPPORTED_MODE` | 400 | 模式不支持 | `mode`, `supported` |
| `UNSUPPORTED_PADDING` | 400 | 填充方式不支持 | `padding`, `supported` |
| `UNSUPPORTED_FORMAT` | 400 | 格式不支持 | `format`, `supported` |
| `UNSUPPORTED_VERSION` | 400 | 版本不支持 | `version`, `supported` |
| `KEY_SIZE_MISMATCH` | 400 | 密钥长度不符 | `field`, `expected`, `actual` |
| `IV_SIZE_MISMATCH` | 400 | IV 长度不符 | `field`, `expected`, `actual` |
| `PASSPHRASE_REQUIRED` | 400 | 需要口令但未提供 | `field` |
| `PASSPHRASE_INCORRECT` | 400 | 口令错误 | — |
| `TAG_REQUIRED` | 400 | AEAD 缺少 tag | `field` |
| `TAG_INVALID` | 400 | tag 校验失败 | — |
| `DECRYPT_FAILED` | 400 | 解密失败 | — |
| `SIGNATURE_INVALID` | 400 | 签名**结构**非法（非验签不通过） | — |
| `PLAINTEXT_TOO_LONG` | 400 | 明文超出算法上限 | `maxBytes`, `actual` |
| `KEY_MISMATCH` | 400 | 私钥与证书/公钥不配对 | — |
| `SAN_INVALID` | 400 | SAN 条目非法 | `value` |
| `SERIAL_INVALID` | 400 | 序列号非法 | `value` |
| `NOT_AFTER_BEFORE_NOW` | 400 | 有效期非法（终止早于起始） | `notBefore`, `notAfter` |

> **注意**：`SIGNATURE_INVALID` 只表示签名**无法解析**。「验签不通过」是**业务结果**，返回 `200` + `data.valid = false`。

---

## 4. 限制与传输（Limit & Transport）

| 错误码 | HTTP | 含义 | 常见 `params` |
| --- | --- | --- | --- |
| `TOO_LARGE` | 413 | 请求体超出上限 | `maxBytes`, `actual` |
| `TOO_MANY_ITEMS` | 400 | 条目数超出上限 | `max`, `actual` |
| `TIMEOUT` | 504 | 处理超时 | `timeoutMs` |
| `RATE_LIMITED` | 429 | 触发限流 | `limit`, `windowMs` |
| `TOOL_NOT_FOUND` | 404 | 工具不存在 | `slug` |
| `TOOL_DISABLED` | 404 | 工具已禁用 | `slug` |
| `METHOD_NOT_ALLOWED` | 405 | 方法不匹配 | `method`, `allowed` |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | Content-Type 不支持 | `contentType`, `supported` |
| `CROSS_ORIGIN_BLOCKED` | 403 | 跨站请求被拒 | `origin` |
| `NETWORK_DISABLED` | 403 | 出网功能已关闭 | — |
| `SSRF_BLOCKED` | 400 | 目标地址被安全策略拦截 | `host`, `reason` |
| `DNS_ERROR` | 502 | DNS 解析失败 | `host`, `type` |
| `TLS_ERROR` | 502 | TLS 握手失败 | `host`, `reason` |
| `HOST_UNREACHABLE` | 502 | 目标不可达 | `host` |
| `UPSTREAM_TIMEOUT` | 504 | 上游超时 | `host`, `timeoutMs` |
| `UPSTREAM_ERROR` | 502 | 上游返回异常 | `host`, `status` |
| `REDIRECT_LIMIT` | 502 | 重定向次数超限 | `maxRedirects` |

---

## 5. 服务端（Server）

| 错误码 | HTTP | 含义 |
| --- | --- | --- |
| `INTERNAL_ERROR` | 500 | 未分类的服务端错误 |
| `CRYPTO_OPERATION_FAILED` | 500 | 密码学操作失败（已归一化的兜底） |
| `CRYPTO_INIT_FAILED` | 500 | 密码学库初始化失败 |
| `STORAGE_ERROR` | 500 | 统计存储写入失败 |
| `NOT_IMPLEMENTED` | 501 | 端点未实现 |
| `STATS_UNAUTHORIZED` | 401 | 统计端点缺少或错误的 Token |

---

## 6. 警告（Warnings）

警告**不改变** HTTP 状态，出现在 `meta.warnings` 中。

| 警告码 | 含义 |
| --- | --- |
| `INSECURE_ECB` | 使用了 ECB 模式（无扩散） |
| `NO_AUTHENTICATION` | 对称加密未提供认证标签 |
| `SMALL_RSA_KEY` | RSA 密钥长度小于 2048 |
| `WEAK_HASH` | 使用了 MD5 / SHA-1 |
| `EXPIRED_CERT` | 证书已过期 |
| `NOT_YET_VALID` | 证书尚未生效 |
| `SELF_SIGNED_CERT` | 自签名证书 |
| `UNTRUSTED_CHAIN` | 证书链无法追溯到受信任根 |
| `PADDING_ORACLE_RISK` | 使用了 PKCS#1 v1.5 填充 |
| `LARGE_INPUT` | 输入超过常见阈值，响应可能很大 |
| `LOSSY_FORMAT` | 格式转换有损（信息丢失） |
| `REGEX_DIALECT_DIFFERENCE` | 正则方言与目标语言存在差异 |
| `TIMEZONE_ASSUMED` | 时间转换假定了某个时区 |
| `COMMENTS_LOST` | 注释被丢弃 |

---

## 7. 业务结果 vs 错误

工具**执行成功并得出结论** → `200`；工具**无法执行** → `4xx` / `5xx`。

| 场景 | 响应 |
| --- | --- |
| 验签不通过 | `200` + `data.valid = false` |
| 证书已过期 | `200` + `data.validity.status = "expired"` |
| 证书链不可信 | `200` + `data.valid = false` + `data.errors[]` |
| 域名不匹配 | `200` + `data.hostnameMatch = false` |
| 证书被吊销 | `200` + `data.status = "revoked"` |
| OCSP 状态未知 | `200` + `data.certStatus = "unknown"` |
| 私钥与证书不配对 | `400` + `KEY_MISMATCH` |
| 输入不是合法证书 | `400` + `INVALID_CERT` |
| 密钥长度不符 | `400` + `KEY_SIZE_MISMATCH` |
| 参数超出允许范围 | `400` + `INVALID_FIELD` |

---

## 8. 错误响应示例

```json
{
  "code": 400,
  "message": "key size mismatch",
  "data": null,
  "error": {
    "code": "KEY_SIZE_MISMATCH",
    "params": { "field": "key", "expected": 16, "actual": 24 }
  },
  "meta": { "durationMs": 0, "inputBytes": 128, "outputBytes": 0, "warnings": [] },
  "requestId": "0193f4c2-7a1b-7c3d-9e55-8f2a1b3c4d5e"
}
```

---

## 9. 国际化

- 每个错误码**必须**有 i18n key：`errors.<CODE>`。
- `zh-CN` 与 `en-US` **必须**都提供。
- 词条可使用 `params` 中的占位符：`{field}`、`{expected}`、`{actual}`。
- **禁止**在文案中拼接用户输入。

示例（`zh-CN`）：

```json
{
  "errors": {
    "KEY_SIZE_MISMATCH": "密钥长度不符合要求（期望 {expected} 字节，实际 {actual} 字节）",
    "MISSING_FIELD": "缺少必填字段：{field}"
  }
}
```

---

## 10. 相关文档

- [README.md](README.md) —— API 总览与信封
- [admin.md](admin.md) —— 管理与运维接口
- [../standards/04-errors.md](../standards/04-errors.md) —— 错误处理规范
