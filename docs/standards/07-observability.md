# 可观测性规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

---

## 1. 目标

| 目标 | 手段 |
| --- | --- |
| 定位一次失败请求 | `requestId` 贯穿日志 / 响应 / 统计 |
| 看清服务健康度 | 结构化日志 + 健康检查端点 |
| 看清使用分布 | 匿名统计（SQLite） |
| 及时发现异常 | 关键指标埋点与阈值告警 |

**非目标**：不做全链路追踪（APM）、不做用户级监控、不做长周期指标存储。

---

## 2. requestId

**必须**

- 每个请求在**最外层中间件**生成 `requestId`（UUIDv4 或 16 字节随机 hex）；
- 优先采用入站 `X-Request-Id`（**必须**校验格式：仅 `[A-Za-z0-9_-]{1,64}`，否则重新生成）；
- 写入 `context`，随 `request.Context()` 传递；
- 写进**所有**该请求产生的日志；
- 通过响应头 `X-Request-Id` 返回；
- 错误响应的 `meta.requestId` 也返回同一值。

**禁止**在业务层自己生成新的 `requestId`。

---

## 3. 日志

### 3.1 格式

- 生产：**JSON**，每行一条。
- 开发：文本，带颜色（`slog` 的 `TextHandler`）。
- 字段名**必须**用小驼峰或点号分层（`http.status`、`tool.slug`）。

### 3.2 级别

| 级别 | 用途 | 示例 |
| --- | --- | --- |
| `DEBUG` | 开发调试 | 请求体字段名（**不含值**） |
| `INFO` | 常规事件 | 请求完成、工具调用、启动、配置摘要 |
| `WARN` | 可恢复异常 | 限流、超时、孤儿任务、上游降级 |
| `ERROR` | 失败 | panic、5xx、数据库写入失败 |

**必须**：生产环境默认级别 `INFO`；`DEBUG` **禁止**在生产开启（可用 `TOOLBIT_LOG_LEVEL` 显式开启，但不记录敏感值）。

### 3.3 必记字段

| 字段 | 说明 |
| --- | --- |
| `requestId` | 贯穿 |
| `slug` | 工具标识 |
| `status` | HTTP 状态 |
| `durationMs` | 耗时 |
| `errorCode` | 错误码（失败时） |
| `reqBytes` / `respBytes` | 字节数 |

### 3.4 禁止字段

见 [06-security.md §12](06-security.md)。

### 3.5 访问日志

**必须**记录一次请求一条：

```json
{"ts":"2026-10-27T10:00:00Z","level":"INFO","msg":"request","requestId":"...","slug":"sm4-cbc","status":200,"durationMs":12,"reqBytes":256,"respBytes":512}
```

**禁止**在访问日志里额外打印 UA / Referer 原文。

---

## 4. 指标

**必须**暴露（内部端点，不对公网开放）：

| 指标 | 类型 | 说明 |
| --- | --- | --- |
| `toolbit_requests_total` | counter | 按 `slug` / `status` |
| `toolbit_request_duration_ms` | histogram | 按 `slug` |
| `toolbit_active_requests` | gauge | 进行中的请求数 |
| `toolbit_errors_total` | counter | 按 `error.code` |
| `toolbit_rate_limited_total` | counter | 429 次数 |
| `toolbit_crypto_ops_total` | counter | 密码学操作数（按算法） |
| `toolbit_stats_queue_depth` | gauge | 统计队列积压 |
| `toolbit_stats_dropped_total` | counter | 统计丢弃数 |

暴露方式：v1 **必须**提供 `GET /internal/metrics`（Prometheus 文本格式），仅在 `TOOLBIT_ENABLE_INTERNAL_ENDPOINTS=true` 时开启。

**禁止**把内部端点暴露到公网监听端口。

---

## 5. 健康检查

| 端点 | 语义 | 判定 |
| --- | --- | --- |
| `GET /healthz` | 存活 | 进程能响应即 200 |
| `GET /readyz` | 就绪 | 静态目录可读、SQLite 可写、catalog 加载成功 → 200，否则 503 |

**必须**：`/healthz` **不**做任何 I/O；`/readyz` 的检查**必须**有超时（≤ 2 秒）。

**禁止**把 `/healthz` 用作负载均衡的就绪探针。

---

## 6. 统计

见 [architecture.md §8](../architecture.md)。要点：

- 只记录：`slug`、`day`、`status_class`、`error_code`、`duration_bucket`、`region`（可选）；
- **禁止**记录请求体、参数内容、IP 原值；
- 写入**必须**异步、不阻塞请求；
- 队列满**必须**丢弃并计数（`toolbit_stats_dropped_total`）。

---

## 7. 告警建议

| 指标 | 阈值 | 级别 |
| --- | --- | --- |
| 5xx 比例 | > 1%（5 分钟） | 重要 |
| `toolbit_error_total` 突增 | 相对基线 × 3 | 重要 |
| `toolbit_stats_dropped_total` 增长 | 持续 5 分钟 | 一般 |
| `toolbit_active_requests` | 持续 > 200 | 一般 |
| 磁盘占用（数据卷） | > 80% | 一般 |
| SQLite 写入失败 | 任意 | 严重 |

告警**必须**带上 `requestId` 样例。

---

## 8. 审计

**必须**记录以下**元事件**（不含内容）：

| 事件 | 字段 |
| --- | --- |
| 服务启动 | 版本、配置摘要（**脱敏**）、catalog 工具数 |
| 配置错误 | 缺哪个变量（**不**含值） |
| 工具禁用 | slug、原因 |
| 上游 DNS 失败 | 目标主机、错误类型 |

**禁止**记录任何用户提交的材料。

---

## 9. 版本与兼容

**必须**在响应头或 `/internal/version` 暴露：

| 头 / 端点 | 内容 |
| --- | --- |
| `X-Toolbit-Version` | 语义化版本 |
| `X-Toolbit-Commit` | 短 commit sha |
| `X-Toolbit-Build-Time` | 构建时间（RFC3339 UTC） |
| `GET /internal/version` | 同上 + Go / Node 版本 |

**禁止**在响应中暴露内部路径或主机名。

---

## 10. 禁止清单

- ❌ 在日志中记录请求体、响应体、密钥。
- ❌ 在多处生成不同的 `requestId`。
- ❌ 用 `/healthz` 做就绪探针。
- ❌ 统计写入阻塞请求路径。
- ❌ 把内部指标端点绑定到公网。
- ❌ 在生产开启 `DEBUG` 并打印输入值。
