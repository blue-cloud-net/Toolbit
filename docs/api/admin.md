# 管理与运维接口

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

本文描述**非工具类**端点：健康检查、服务元信息、统计、版本与内部运维端点。

---

## 1. 端点总览

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/api/health` | 无 | 进程健康（DB 状态、版本、uptime） |
| `GET` | `/api/v1/meta` | 无 | 服务元信息 |
| `GET` | `/api/v1/stats/popular` | 无 | 公开聚合（工具使用排行） |
| `GET` | `/api/v1/stats/overview` | `X-Stats-Token` | 详细统计 |
| `GET` | `/healthz` | 无 | 存活探针（无 I/O） |
| `GET` | `/readyz` | 无 | 就绪探针 |
| `GET` | `/internal/version` | 内部 | 详细版本信息 |
| `GET` | `/internal/metrics` | 内部 | Prometheus 指标 |

**内部端点**（`/internal/*`）**必须**在 `TOOLBIT_ENABLE_INTERNAL_ENDPOINTS=true` 时才注册，且**禁止**绑定到公网地址。

---

## 2. GET /api/health

进程级健康检查。响应**不使用**统一信封（便于探针直接判断）。

```http
GET /api/health
```

```json
{
  "status": "ok",
  "version": "0.1.0",
  "commit": "a1b2c3d",
  "uptimeSeconds": 3600,
  "database": {
    "enabled": true,
    "ok": true,
    "events": 1284
  },
  "tools": {
    "total": 70,
    "enabled": 70
  }
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `status` | string | `ok` \| `degraded` |
| `version` | string | 语义化版本 |
| `commit` | string | 短 commit sha |
| `uptimeSeconds` | number | 进程运行时长 |
| `database.ok` | boolean | 统计库可写 |
| `tools.total` / `tools.enabled` | number | catalog 工具总数 / 启用数 |

**注意**：`status: degraded` 时 HTTP 状态仍为 `200`（进程还活着），由监控系统判断是否告警。

---

## 3. GET /api/v1/meta

```http
GET /api/v1/meta
```

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "name": "Toolbit",
    "version": "0.1.0",
    "commit": "a1b2c3d",
    "builtAt": "2026-10-27T08:00:00Z",
    "goVersion": "go1.25.0",
    "tongsuoVersion": "8.4.0",
    "locales": ["zh-CN", "en-US"],
    "defaultLocale": "zh-CN",
    "features": {
      "network": true,
      "stats": true,
      "rateLimit": true
    },
    "apiVersion": "v1"
  },
  "error": null,
  "meta": { "durationMs": 0, "inputBytes": 0, "outputBytes": 384, "warnings": [] },
  "requestId": "…"
}
```

前端**必须**用 `locales` / `defaultLocale` 决定可用语言，**禁止**硬编码语言列表。

---

## 4. GET /api/v1/stats/popular

公开的匿名聚合，用于首页展示「热门工具」。

```http
GET /api/v1/stats/popular?days=7&limit=10
```

| 参数 | 类型 | 默认 | 说明 |
| --- | --- | --- | --- |
| `days` | int | 7 | 统计窗口（1–90） |
| `limit` | int | 10 | 返回条数（1–50） |

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "days": 7,
    "totalCalls": 15420,
    "items": [
      { "slug": "base64", "category": "encoding", "calls": 3820, "successRate": 0.998 },
      { "slug": "json-format", "category": "formatting", "calls": 2140, "successRate": 0.995 }
    ]
  },
  "error": null,
  "meta": { "durationMs": 3, "inputBytes": 0, "outputBytes": 512, "warnings": [] },
  "requestId": "…"
}
```

**必须**

- 只返回**聚合**结果，**不**返回任何可关联到个人的信息；
- 参数越界**必须**回退到默认值（**不**报错），避免前端抖动；
- 统计**未启用**时返回 `data.items = []` 与 `totalCalls = 0`，**不**返回错误。

**禁止**在此端点暴露 IP、UA、时间戳序列等高基数维度。

---

## 5. GET /api/v1/stats/overview

详细统计，**需要** `X-Stats-Token`。

```http
GET /api/v1/stats/overview?days=30
X-Stats-Token: <token>
```

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "days": 30,
    "totalCalls": 68320,
    "byCategory": [
      { "category": "encoding", "calls": 20130, "successRate": 0.997 }
    ],
    "byDay": [
      { "day": "2026-10-27", "calls": 2410, "errors": 12 }
    ],
    "byErrorCode": [
      { "code": "INVALID_FIELD", "count": 87 }
    ],
    "byTool": [
      { "slug": "base64", "calls": 3820, "errors": 5, "avgDurationMs": 2 }
    ],
    "slowest": [
      { "slug": "keygen", "avgDurationMs": 312 }
    ]
  },
  "error": null,
  "meta": { "durationMs": 12, "inputBytes": 0, "outputBytes": 4096, "warnings": [] },
  "requestId": "…"
}
```

**鉴权行为**

| 情况 | 响应 |
| --- | --- |
| Token 正确 | `200` + 数据 |
| Token 错误 | `401` + `STATS_UNAUTHORIZED` |
| Token 缺失 | `401` + `STATS_UNAUTHORIZED` |
| **未配置** Token | `404` + `TOOL_NOT_FOUND`（端点不存在） |

**必须**

- Token 通过环境变量配置，**禁止**硬编码；
- Token 比较**必须**使用**常数时间**比较（防时序侧信道）；
- 端点**必须**不写缓存（`Cache-Control: no-store`）。

---

## 6. 探针端点

### 6.1 GET /healthz

存活探针。**不执行任何 I/O**，进程能响应即 `200`。

```http
GET /healthz
```

```text
ok
```

### 6.2 GET /readyz

就绪探针。检查项：

| 检查 | 失败含义 |
| --- | --- |
| 静态目录可读 | 前端资源未正确挂载 |
| catalog 已加载 | 工具目录解析失败 |
| SQLite 可写（若启用统计） | 数据卷权限错误 |
| 迁移版本匹配 | 需要执行数据库迁移 |

```http
GET /readyz
```

```json
{
  "status": "ready",
  "checks": {
    "static": "ok",
    "catalog": "ok",
    "database": "ok"
  }
}
```

任一检查失败 → `503` + `status: "not_ready"`。

**必须**给检查设置 ≤ 2 秒的超时。

---

## 7. 内部端点

**仅在** `TOOLBIT_ENABLE_INTERNAL_ENDPOINTS=true` 时注册。

### 7.1 GET /internal/version

```json
{
  "version": "0.1.0",
  "commit": "a1b2c3d4e5f6",
  "buildTime": "2026-10-27T08:00:00Z",
  "goVersion": "go1.25.0",
  "tongsuoVersion": "8.4.0",
  "nodeVersion": "v22.11.0",
  "platform": "linux/amd64",
  "catalogTools": 70
}
```

### 7.2 GET /internal/metrics

Prometheus 文本格式：

```text
# HELP toolbit_requests_total Total tool requests.
# TYPE toolbit_requests_total counter
toolbit_requests_total{slug="base64",status="200"} 3820
toolbit_requests_total{slug="base64",status="400"} 5
# HELP toolbit_request_duration_ms Tool request duration in milliseconds.
# TYPE toolbit_request_duration_ms histogram
toolbit_request_duration_ms_bucket{slug="base64",le="5"} 3800
toolbit_request_duration_ms_bucket{slug="base64",le="10"} 3820
toolbit_request_duration_ms_sum{slug="base64"} 8012
toolbit_request_duration_ms_count{slug="base64"} 3820
```

可用指标清单见 [../standards/07-observability.md §4](../standards/07-observability.md)。

**禁止**在指标标签中使用 `slug` 以外的高基数维度（如 `requestId`）。

---

## 8. 响应头

除工具端点外，本组端点**不**强制 `no-store`：

| 端点 | `Cache-Control` |
| --- | --- |
| `/api/health` | `no-store` |
| `/api/v1/meta` | `public, max-age=60` |
| `/api/v1/tools` | `public, max-age=60`（带 `ETag`） |
| `/api/v1/stats/popular` | `public, max-age=300` |
| `/api/v1/stats/overview` | `no-store` |
| `/healthz` / `/readyz` | `no-store` |
| `/internal/*` | `no-store` |

---

## 9. 运维脚本

### 9.1 容器内健康检查

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/app/toolbit", "healthcheck"]
```

`toolbit healthcheck` 子命令**必须**：

- 访问 `http://127.0.0.1:30234/readyz`；
- 状态非 `200` 时**退出码 1**；
- 设置 ≤ 3 秒超时；
- **不**输出敏感信息。

### 9.2 冒烟检查

```bash
curl -fsS http://localhost:30234/readyz >/dev/null
curl -fsS http://localhost:30234/api/v1/tools | head -c 200
curl -fsS -X POST http://localhost:30234/api/v1/tools/base64 \
  -H 'Content-Type: application/json' \
  -d '{"operation":"encode","input":{"data":"ok","encoding":"utf8"},"outputEncoding":"base64"}'
```

---

## 10. 统计模型

事件结构（**不**含任何用户内容）：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `ts` | int | Unix 毫秒 |
| `slug` | string | 工具标识 |
| `category` | string | 分类 |
| `ok` | bool | 是否成功 |
| `status` | int | HTTP 状态 |
| `errorCode` | string | 错误码（成功时为空） |
| `durationMs` | int | 耗时 |
| `reqBytes` | int | 请求字节数 |
| `respBytes` | int | 响应字节数 |
| `visitorHash` | string | 访客哈希（截断 / 哈希后的 IP，**不可逆**） |

**必须**

- 写入**异步**（内存 channel + 批量事务）；
- 队列满**丢弃**事件并计数，**禁止**阻塞请求；
- `visitorHash` **必须**不可逆，且**不得**单独用于跨工具关联。

**禁止**记录请求体、参数内容、原始 IP、User-Agent 原文。

---

## 11. 相关文档

- [README.md](README.md) —— API 总览
- [errors.md](errors.md) —— 错误码全集
- [../deploy.md](../deploy.md) —— 部署与配置
- [../standards/07-observability.md](../standards/07-observability.md) —— 可观测性规范
- [../standards/08-deploy.md](../standards/08-deploy.md) —— 部署规范
