# API 总览

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

本文描述 Toolbit 的 HTTP 接口约定。所有工具共用**一个分发端点**，请求与响应结构由 catalog 驱动。

工具的逐项契约见 [tools/](tools/) 目录：

| 文件 | 分类 | 工具数 |
| --- | --- | --- |
| [crypto.md](tools/crypto.md) | 加密 | 8 |
| [keys.md](tools/keys.md) | 密钥 | 6 |
| [certificates.md](tools/certificates.md) | 证书 | 23 |
| [encoding.md](tools/encoding.md) | 编码 | 6 |
| [formatting.md](tools/formatting.md) | 格式化 | 8 |
| [text.md](tools/text.md) | 文本 | 2 |
| [data.md](tools/data.md) | 数据 | 6 |
| [system.md](tools/system.md) | 系统 | 3 |
| [network.md](tools/network.md) | 网络 | 5 |
| [documents.md](tools/documents.md) | 文档 | 3 |

错误码见 [errors.md](errors.md)，管理与运维接口见 [admin.md](admin.md)。

---

## 1. 端点

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/health` | 进程健康（DB 状态、版本、uptime） |
| `GET` | `/api/v1/meta` | 服务元信息（版本、commit、构建时间、locales、features） |
| `GET` | `/api/v1/tools` | 工具目录（带 `ETag`） |
| `POST` | `/api/v1/tools/{slug}` | 执行工具（JSON 或 multipart） |
| `GET` | `/api/v1/tools/{slug}/stream` | WebSocket 流式（仅 `ssl-check`） |
| `GET` | `/api/v1/stats/popular` | 公开聚合（工具使用排行） |
| `GET` | `/api/v1/stats/overview` | 详细统计（需 `X-Stats-Token`） |

> `/api/v1` 为版本化前缀。v1 内**只允许**向后兼容地新增字段；删除或改变字段语义**必须**引入 v2。

---

## 2. 统一信封

### 2.1 成功

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "output": "..."
  },
  "error": null,
  "meta": {
    "durationMs": 12,
    "inputBytes": 256,
    "outputBytes": 512,
    "warnings": []
  },
  "requestId": "0193f4c2-7a1b-7c3d-9e55-8f2a1b3c4d5e"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | number | 固定为 `0` |
| `message` | string | 固定为 `"ok"` |
| `data` | object \| null | 业务结果 |
| `error` | null | 成功时恒为 `null` |
| `meta` | object | 执行元信息 |
| `requestId` | string | 请求标识，用于日志关联 |

### 2.2 失败

```json
{
  "code": 400,
  "message": "invalid field value",
  "data": null,
  "error": {
    "code": "INVALID_FIELD",
    "params": { "field": "key", "expected": 16, "actual": 24 }
  },
  "meta": {
    "durationMs": 1,
    "inputBytes": 64,
    "outputBytes": 0,
    "warnings": []
  },
  "requestId": "0193f4c2-7a1b-7c3d-9e55-8f2a1b3c4d5e"
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | number | **等于 HTTP 状态码** |
| `message` | string | 英文兜底文案，**不含**用户输入 |
| `data` | null | 失败时恒为 `null` |
| `error.code` | string | 机器可读错误码，见 [errors.md](errors.md) |
| `error.params` | object | 可插值参数，可缺省 |

> `code` 同时出现在顶层与 HTTP 状态，是为了让**不支持读取状态码**的客户端也能区分结果。

### 2.3 二进制响应

文件类工具（如 `sm4-file-crypt`、`gzip-deflate`）**不**使用信封，直接返回二进制流：

| 响应头 | 说明 |
| --- | --- |
| `Content-Type` | 如 `application/octet-stream` |
| `Content-Disposition` | `attachment; filename*=UTF-8''<percent-encoded>` |
| `X-Request-Id` | 请求标识 |
| `X-Duration-Ms` | 耗时（毫秒） |
| `X-Input-Bytes` / `X-Output-Bytes` | 字节数 |
| `X-Warnings` | 逗号分隔的 warning code |
| `Cache-Control` | 固定 `no-store` |

---

## 3. meta

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `durationMs` | number | 服务端处理耗时（毫秒，向上取整） |
| `inputBytes` | number | 请求体字节数 |
| `outputBytes` | number | 响应体字节数 |
| `warnings` | array | 警告列表，见 §6 |

**必须**：`meta` 字段**只允许新增**，不得改变语义。

---

## 4. 通用类型

### 4.1 Encoding

```json
"utf8" | "hex" | "base64" | "base64url"
```

### 4.2 ByteInput / Output

```json
{ "data": "48656c6c6f", "encoding": "hex" }
```

`data` **始终**是字符串，字节序列由 `encoding` 解释。**禁止**在 JSON 中直接放二进制。

### 4.3 KeyInput

```json
{
  "data": "-----BEGIN PRIVATE KEY-----...",
  "format": "auto",
  "passphrase": "optional"
}
```

| 字段 | 取值 |
| --- | --- |
| `format` | `auto` \| `pem` \| `der` \| `hex` \| `base64` |
| `passphrase` | 可选；加密私钥必填 |

`format: auto` 的识别顺序：**PEM 头 → DER → base64 → hex**，全部失败返回 `INVALID_PEM`。

### 4.4 PemBundle

```json
{ "data": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----" }
```

自动识别单个 / 多个 PEM 拼接、裸 DER、base64、hex。

### 4.5 FileInput

multipart 的 `file` 字段，**单文件**，**流式**读取，**禁止**落盘。

### 4.6 TimeInput

RFC3339 字符串，或 Unix 秒 / 毫秒整数（自动识别位数）。

### 4.7 Output

```json
{ "data": "…", "encoding": "utf8" }
```

---

## 5. 限额与限流

限流档位由 catalog 的 `rateClass` 决定。

| 档位 | 速率 / 突发 | 请求体上限 | 超时 | 典型工具 |
| --- | --- | --- | --- | --- |
| `light` | 30 rps / burst 60 | 1 MB | 5 s | base64、json-format、timestamp、cidr-calc、chmod-calc |
| `standard` | 10 rps / burst 20 | 4 MB | 15 s | symmetric-crypt、cert-parse、pfx-pack、eml-parser |
| `heavy` | 2 rps / burst 4 | 20 MB | 30 s | keygen、cert-gen、crl-gen、pfx-pack |
| `file` | 1 并发 / 1 rps | 100 MB | 180 s | sm4-file-crypt、pdf-sign-verify |
| `network` | 1 rps / burst 2 | 256 KB | 20 s | ssl-check、dns-lookup、ct-search、domain-verify |

**输出上限**：JSON 响应 8 MB；文件响应 400 MB（输入上限 100 MB）。

**限流响应**：

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 2
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1760000000
```

- 限流键：**IP**；
- **不**信任 `X-Forwarded-For`，除非显式设置 `TOOLBIT_TRUST_PROXY=true`；
- `Retry-After` 单位为**秒**。

---

## 6. warnings

`meta.warnings` 中的每一项：

```json
{ "code": "INSECURE_ECB", "params": { "mode": "ECB" } }
```

警告**不改变** HTTP 状态，也不影响 `data` 的可用性。完整列表见 [errors.md §4](errors.md)。

---

## 7. 通用响应头

| 头 | 值 |
| --- | --- |
| `Content-Type` | `application/json; charset=utf-8`（或二进制类型） |
| `Cache-Control` | `no-store` |
| `X-Request-Id` | 请求标识 |
| `X-Content-Type-Options` | `nosniff` |
| `Referrer-Policy` | `no-referrer` |

---

## 8. 请求处理顺序

服务端按以下顺序校验，**命中第一个失败即返回**：

1. 路由与工具存在性 → `TOOL_NOT_FOUND` / `TOOL_DISABLED`
2. HTTP 方法 → `METHOD_NOT_ALLOWED`
3. `Content-Type` → `UNSUPPORTED_MEDIA_TYPE`
4. `Origin` / `Sec-Fetch-Site` → `CROSS_ORIGIN_BLOCKED`
5. 限流 → `RATE_LIMITED`
6. 请求体大小 → `TOO_LARGE`
7. 语法解码 → `INVALID_JSON` / `INVALID_FIELD`
8. 字段语义 → `INVALID_FIELD` / `MISSING_FIELD`
9. 跨字段一致性 → `KEY_MISMATCH` 等
10. 算法相关 → `UNSUPPORTED_ALGORITHM` / `KEY_SIZE_MISMATCH`

---

## 9. 跨站请求防护

- **不**开启 CORS（无 `Access-Control-Allow-Origin`）。
- 对 `POST` 等不安全方法校验 `Origin`（与 `Host` 比对）与 `Sec-Fetch-Site`。
- `Sec-Fetch-Site: cross-site` **一律拒绝**（403 + `CROSS_ORIGIN_BLOCKED`）。
- 部署在反向代理后时，反代**必须**覆写 `Origin` / `Sec-Fetch-Site`，**禁止**透传客户端伪造值。

---

## 10. 工具目录端点

```http
GET /api/v1/tools
```

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "version": 1,
    "categories": [
      { "slug": "crypto", "name": "加密", "nameEn": "Cryptography", "order": 1, "icon": "lock" }
    ],
    "tools": [
      {
        "slug": "symmetric-crypt",
        "category": "crypto",
        "order": 3,
        "name": "对称加解密",
        "nameEn": "Symmetric Encryption",
        "transport": "json",
        "rateClass": "standard",
        "maxBodyBytes": 4194304,
        "timeoutMs": 15000,
        "sensitive": true,
        "requiresNetwork": false,
        "autoRun": false,
        "wide": false,
        "i18nKey": "tools.symmetric-crypt"
      }
    ]
  },
  "error": null,
  "meta": { "durationMs": 1, "inputBytes": 0, "outputBytes": 2048, "warnings": [] },
  "requestId": "…"
}
```

- 带 `ETag`（内容哈希），客户端**应当**用 `If-None-Match` 协商；
- 响应 `Cache-Control: public, max-age=60`（目录不含用户数据，可缓存）。

---

## 11. WebSocket 协议

仅 `ssl-check` 使用流式端点。

```http
GET /api/v1/tools/ssl-check/stream
Upgrade: websocket
```

### 11.1 客户端首帧

连接建立后**必须**在 5 秒内发送一帧 JSON，否则服务端以 `1008` 关闭：

```json
{
  "host": "example.com",
  "port": 443,
  "sni": "example.com",
  "depth": 2,
  "versions": ["TLS1.3", "TLS1.2"],
  "timeoutMs": 30000
}
```

### 11.2 服务端帧

| `type` | 载荷 |
| --- | --- |
| `progress` | `{ "phase": "handshake", "done": 1, "total": 2 }` |
| `result` | `{ "type": "result", "data": { … } }` |
| `error` | `{ "type": "error", "error": { "code": "TLS_ERROR", "message": "…", "params": {} } }` |

### 11.3 关闭码

| 码 | 含义 |
| --- | --- |
| `1000` | 正常结束 |
| `1008` | 输入非法或被策略拒绝 |
| `1011` | 服务端内部错误 |

### 11.4 生命周期

- 心跳 ping 间隔 30 秒；
- 空闲 60 秒关闭；
- 单次任务总预算 30 秒；
- 同一 IP **同时最多 1 条**连接。

---

## 12. 调用示例

### 12.1 Base64 编码

```bash
curl -sS -X POST http://localhost:30234/api/v1/tools/base64 \
  -H 'Content-Type: application/json' \
  -d '{"operation":"encode","input":{"data":"Hello","encoding":"utf8"},"outputEncoding":"base64"}'
```

```json
{
  "code": 0,
  "message": "ok",
  "data": { "output": "SGVsbG8=" },
  "error": null,
  "meta": { "durationMs": 1, "inputBytes": 78, "outputBytes": 152, "warnings": [] },
  "requestId": "…"
}
```

### 12.2 SM4-GCM 加密

```bash
curl -sS -X POST http://localhost:30234/api/v1/tools/symmetric-crypt \
  -H 'Content-Type: application/json' \
  -d '{
    "algorithm": "SM4",
    "operation": "encrypt",
    "mode": "GCM",
    "key": { "data": "0123456789ABCDEFFEDCBA9876543210", "encoding": "hex" },
    "iv":  { "data": "000102030405060708090a0b", "encoding": "hex" },
    "input": { "data": "机密内容", "encoding": "utf8" },
    "outputEncoding": "base64"
  }'
```

### 12.3 证书解析

```bash
curl -sS -X POST http://localhost:30234/api/v1/tools/cert-parse \
  -H 'Content-Type: application/json' \
  -d '{"input":{"data":"-----BEGIN CERTIFICATE-----\n…\n-----END CERTIFICATE-----"}}'
```

### 12.4 文件加密（multipart）

```bash
curl -sS -X POST http://localhost:30234/api/v1/tools/sm4-file-crypt \
  -F 'operation=encrypt' \
  -F 'mode=GCM' \
  -F 'key={"data":"0123456789ABCDEFFEDCBA9876543210","encoding":"hex"}' \
  -F 'file=@report.pdf' \
  -o report.pdf.sm4
```

### 12.5 触发错误

```bash
curl -sS -X POST http://localhost:30234/api/v1/tools/symmetric-crypt \
  -H 'Content-Type: application/json' \
  -d '{"algorithm":"SM4","operation":"encrypt","mode":"GCM","key":{"data":"00","encoding":"hex"},"input":{"data":"x","encoding":"utf8"}}'
```

```json
{
  "code": 400,
  "message": "key size mismatch",
  "data": null,
  "error": {
    "code": "KEY_SIZE_MISMATCH",
    "params": { "field": "key", "expected": 16, "actual": 1 }
  },
  "meta": { "durationMs": 0, "inputBytes": 141, "outputBytes": 0, "warnings": [] },
  "requestId": "…"
}
```

---

## 13. 版本与兼容

- URL 路径前缀 `/api/v1`。
- v1 内**只允许**新增**可选**字段；新增字段对旧客户端**必须**可忽略。
- 改变字段语义、删除字段、改变错误码含义 → **必须**引入 `/api/v2`。
- `GET /api/v1/meta` 返回的 `version` 为**服务版本**（语义化），与 API 版本**独立**。

---

## 14. 相关文档

- [errors.md](errors.md) —— 错误码与警告全集
- [admin.md](admin.md) —— 管理与运维接口
- [tools/](tools/) —— 逐工具契约
- [../standards/04-errors.md](../standards/04-errors.md) —— 错误处理规范
- [../architecture.md](../architecture.md) —— 架构与关键决策
