# 后端规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

适用目录：`backend/`。

---

## 1. 技术栈

- Go 1.25，`CGO_ENABLED=1`（`tongsuo-go` 依赖 cgo）
- Gin（HTTP 路由与中间件）
- `github.com/blue-cloud-net/tongsuo-go` v0.5.0（**仅** `backend/internal/cryptosvc` 可导入）
- `modernc.org/sqlite`（纯 Go 驱动）

---

## 2. 包布局与依赖方向

```text
backend/cmd/toolbit  →  backend/internal/api  →  backend/internal/tools
                     →  backend/internal/cryptosvc  →  tongsuo-go
```

后端是**独立 Go 模块**，模块根为 `backend/`，import 路径形如
`github.com/blue-cloud-net/Toolbit/backend/internal/api`。

**必须**

- 依赖**单向向下**，禁止反向 import 与循环依赖；
- `backend/internal/cryptosvc` 是**唯一**允许 `import tongsuo-go` 的包；
- `backend/internal/httpx` **不得**依赖任何领域包；
- `backend/internal/catalog` **不得**认识工具实现；
- 工具之间**禁止**互相 import；需要共用能力时下沉到 `cryptosvc` / `codecsvc` / `netguard`。

---

## 3. 分层职责

| 层 | 允许做 | 禁止做 |
| --- | --- | --- |
| `backend/internal/api` | 查目录、协商 Content-Type、限流、超时、Body 限制、Origin 校验、注册路由 | 实现业务算法 |
| `backend/internal/tools/*` | 解码输入、字段校验、调用服务、编码 DTO | 直接调 `tongsuo-go`、自造错误码、写 SQL |
| `backend/internal/cryptosvc` | 调 Tongsuo、管理句柄生命周期、算法分发、限额校验 | 感知 HTTP / JSON / Gin |
| `backend/internal/codecsvc` | 编码归一化、PEM/DER 识别 | 感知具体工具 |
| `backend/internal/netguard` | SSRF 防护、出网策略 | 感知具体工具 |
| `backend/internal/stats` | 事件写入与聚合 | 感知具体工具的业务语义 |

---

## 4. handler 形态

```go
// backend/internal/tools/crypto/symmetric.go
type symmetricRequest struct {
    Algorithm string `json:"algorithm"`
    Operation string `json:"operation"`
    Mode      string `json:"mode"`
    // ...
}

func (r symmetricRequest) Validate() error { /* 字段级校验，返回 toolerr */ }

func SymmetricCrypt(ctx context.Context, in symmetricRequest) (tools.Result, error) {
    // 1. 校验 → 2. 调 cryptosvc → 3. 编码 DTO
}
```

**必须**

- 请求结构体使用**具名类型**，字段带 `json` tag；
- 每类请求提供 `Validate() error`；
- 错误**必须**是 `*toolerr.Error`（带 `Code` 与 `Params`）；
- handler **不得**拼接 `map[string]any` 作为响应 `data`（前端类型失去保证）；
- 响应 DTO **必须**是具名结构体，字段带 `json` tag 与注释。

**禁止**

- handler 内出现 `panic`（编程错误除外，且必须被 `Recovery` 捕获）；
- handler 内直接 `fmt.Sprintf` 拼错误信息给前端；
- handler 内出现 `os.ReadFile` / `sql.Query` 等基础设施调用。

---

## 5. 请求解码

- JSON 解码**必须**开启严格模式（拒绝未知字段）；
- JSON 语法错误 → `INVALID_JSON`；字段类型错误 → `INVALID_FIELD`（带字段名）；
- 缺失必填字段 → `MISSING_FIELD`（带字段名）；
- multipart **必须**使用字段白名单，单文件，**禁止**落盘；
- 请求体大小**必须**按 catalog 的 `maxBodyBytes` 用 `http.MaxBytesReader` 限制；
- 超过限制 → `413` + `TOO_LARGE`（带 `maxBytes`）。

---

## 6. 上下文与超时

- **必须**使用 `c.Request.Context()`，**禁止**使用 `context.Background()` 覆盖。
- 每个工具有独立超时（catalog `timeoutMs`），由中间件统一施加。
- 流式处理**必须**周期性检查 `ctx.Err()`（建议每 1 MB）。
- **必须**尊重 `ctx` 取消；不得因为客户端断开而继续占用 CPU。

---

## 7. 错误处理

见 [04-errors.md](04-errors.md)。要点：

- 业务结果（如「验签不通过」「证书已过期」）**用 200 + 字段表达**，**不**用 4xx；
- 只有「请求无法被处理」才用 4xx；
- 错误 `message` **不得**包含用户输入。

---

## 8. 日志

使用标准库 `log/slog`，输出 JSON（生产）或文本（开发）。

**允许记录的字段**

| 字段 | 说明 |
| --- | --- |
| `ts` | 时间 |
| `level` | 级别 |
| `msg` | 固定消息 |
| `requestId` | 请求标识（贯穿日志与响应） |
| `slug` | 工具标识 |
| `status` | HTTP 状态 |
| `durationMs` | 耗时 |
| `reqBytes` / `respBytes` | 字节数 |
| `errorCode` | 机器可读错误码 |

**禁止记录**

- ❌ 请求体、响应体
- ❌ 私钥、口令、密钥材料
- ❌ 证书 / CSR / CRL 正文
- ❌ 原始 IP、User-Agent 原文
- ❌ 底层 C 库错误文本中的输入片段

底层错误**只**进日志（带 `requestId`），对外只给错误码。

---

## 9. 配置

- 全部配置通过**环境变量**注入，前缀 `TOOLBIT_`；
- 读取集中在 `backend/internal/config`，启动时一次性解析并校验；
- 解析失败**必须** `fail-fast`（打印错误并退出），**禁止**用「静默默认值」掩盖配置错误；
- 配置项**不得**在运行时通过接口修改（v1 无热更新）；
- 每个配置项**必须**有：名称、类型、默认值、取值范围说明、是否必填。

---

## 10. 并发

- 共享状态**必须**加锁或改用无共享设计；`go test -race` **必须**通过。
- 重操作（密钥生成、证书签发、PKCS#12 打包）**必须**经全局信号量限并发（≤ 2）。
- 信号量获取**必须**响应 `ctx`。
- cgo 调用**不可中断**：超时后底层仍会跑完，**必须**记录 orphan 日志并确保信号量被释放。
- **禁止**在 handler 中启动无生命周期管理的 goroutine。

---

## 11. SQLite 统计

- 驱动：`modernc.org/sqlite`（纯 Go）。
- 开启 WAL；设置 `busy_timeout`。
- 写入：内存 channel + 批量事务（2 秒 / 100 条触发）；**队列满时丢弃事件**，**禁止**阻塞请求。
- 读取：聚合查询用 `day` 索引。
- 保留期默认 90 天，每日清理；清理在后台任务中执行。
- 迁移**必须**版本化（`schema_version` 表），**禁止**在启动时无条件重建表。

---

## 12. 静态资源

见 [architecture.md §5](../architecture.md)。要点：

- 静态目录由 `TOOLBIT_WEB_ROOT` 指定；
- 路径先经 `fs.ValidPath` 校验；
- `/api/*` 未命中返回 JSON 404，**禁止**回落到 `index.html`。

---

## 13. 测试

见 [05-testing.md](05-testing.md)。要点：

- handler **必须**有正例 + 反例；
- 错误码**必须**断言 `error.code`（而非仅断言 HTTP 状态）；
- 涉及密码学的**必须**有标准向量测试。

---

## 14. 禁止清单

- ❌ 在 `backend/internal/api` 之外注册路由。
- ❌ handler 直接 import `tongsuo-go`。
- ❌ 用 `map[string]any` 作为响应 `data`。
- ❌ 关闭 `priv.Public()` 返回的公钥对象。
- ❌ 用 `x509.Name.Add` 处理用户输入。
- ❌ 依赖 finalizer 释放原生句柄。
- ❌ 在日志中输出敏感数据。
- ❌ 静默吞掉错误（`_ = err` 只允许在明确的「忽略」注释旁出现）。
