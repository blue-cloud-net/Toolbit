# 测试规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-11-01

---

## 1. 测试分层

| 层 | 范围 | 工具 | 何时跑 |
| --- | --- | --- | --- |
| 单元测试 | 单个函数 / 服务方法 | Go `testing` / Vitest | 每次提交 |
| 契约测试 | 单个工具端点的请求 → 响应 | Go `httptest` | 每次提交 |
| 集成测试 | 跨包协作（dispatcher + handler + service） | Go `httptest` | 每次提交 |
| 端到端 | 浏览器真实操作 | Playwright | PR |
| 模糊测试 | 不可信输入解析 | Go `testing.F` | 定期（CI 每日） |
| 基准测试 | 关键路径性能 | Go `testing.B` | 定期 |

**必须**：`go test ./... -race` 与 `pnpm test` 在 CI 中通过。

---

## 2. 后端测试

### 2.1 通用要求

- 测试文件与被测文件同目录，命名 `<file>_test.go`。
- 测试函数名 `Test<被测标识>`；表驱动用 `Test<标识>_Table`。
- 断言**必须**针对具体值或具体 `error.code`，**禁止**只断言「没报错」。
- 错误断言**必须**校验 `error.code`（而非仅 HTTP 状态）。
- 测试**禁止**依赖真实网络、真实时间（用注入时钟）、真实文件系统（用 `t.TempDir()`）。
- 测试数据**必须**来自标准向量或固定夹具，**禁止**使用真实密钥。

### 2.2 handler 测试

每个工具**必须**至少包含：

| 用例 | 断言 |
| --- | --- |
| 正例 | `code == 0`，`data` 的关键字段等于预期值 |
| 反例 | `code == 4xx`，`error.code` 等于预期码 |
| 边界 | 必测最小 / 最大合法输入 |

**示例**

```go
func TestSymmetricCrypt(t *testing.T) {
    t.Run("sm4-cbc-roundtrip", func(t *testing.T) { /* ... */ })
    t.Run("invalid-key-length", func(t *testing.T) {
        _, err := SymmetricCrypt(ctx, symmetricRequest{ /* key 24 字节 */ })
        require.Equal(t, toolerr.CodeKeySizeMismatch, err.(*toolerr.Error).Code)
    })
}
```

### 2.3 标准向量

**必须**覆盖：

| 算法 | 向量 |
| --- | --- |
| SM3 | `SM3("abc") = 66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0` |
| SM4-ECB | 密钥 `0123456789ABCDEFFEDCBA9876543210`、明文同串 → 密文 `681EDF34D206965E86B3E94F536E4246` |
| SM4-CBC | 密钥与 IV 同上 → 密文 `2677F46B09C122CC975533105BD4A22AF6125F7275CE552C3A2BBCF533DE8A3B` |
| SM2 | 签名 / 验签（默认与自定义 `userId`）、加解密、格式互转 |
| RSA | OAEP / PSS 往返、最大明文边界 |
| HMAC-SM3 | 固定密钥与消息的标签值 |

向量来源写在测试文件注释里（国标编号或 RFC）。

### 2.4 契约测试

每个工具在 `backend/testdata/contract/<slug>.json` 存放一组 golden 请求 / 响应：

```json
{
  "request": { "...": "..." },
  "expect": { "code": 0, "dataContains": { "...": "..." } }
}
```

契约测试通过 `httptest` 起真实路由，验证：

- HTTP 状态正确；
- 信封结构完整（`code` / `message` / `data` / `error` / `meta` / `requestId`）；
- `data` 包含预期字段；
- 响应带 `Cache-Control: no-store`。

**必须**：98 个工具**全部**有契约测试。

### 2.5 模糊测试

**必须**对以下入口建立 fuzz：

| 目标 | 入口 |
| --- | --- |
| PEM / DER 识别 | `codecsvc` 的自动识别函数 |
| ASN.1 解析 | `asn1` 视图 |
| SM2 密文格式转换 | `asym.Format` 封装 |
| Base64 / Hex 解码 | `codecsvc` |
| 证书链排序 | 链解析 |

Fuzz 目标**不得** panic、**不得**超内存、**不得**死循环。

### 2.6 基准测试

**必须**为以下路径提供基准：

- SM3 摘要（1 MB）
- SM4-GCM 加解密吞吐
- SM2 签名 / 验签
- RSA-2048 密钥生成
- 文件加解密（10 MB）
- 证书解析

---

## 3. 资源与并发测试

**必须**包含：

| 测试 | 目的 |
| --- | --- |
| `Close` 幂等 | 重复释放不崩溃 |
| `Public` 别名不重复关闭 | 关闭私钥后公钥不单独释放 |
| 句柄泄漏 | 1 万次请求后 RSS 稳定 |
| 并发混合调用 | 100 goroutine 同时调用不同服务，无竞态 |
| 信号量释放 | 超时后信号量不泄漏 |
| 限流 | 超过 burst 后返回 429，且 `Retry-After` 存在 |
| Body 限制 | 超限返回 413 + `TOO_LARGE` |
| 超时 | 慢操作返回 504 + `TIMEOUT` |

---

## 4. 前端测试

### 4.1 单元测试（Vitest + `@vue/test-utils`）

**必须**覆盖：

- 组合式函数（`useTheme` / `useAccent` / `useLocale` / `useFavorites` / `useRecent` / `useToolRun` / `useDownload`）；
- API client（信封拆包、错误码映射、超时、取消）；
- 共享组件（`EncodingSelect` / `SensitiveInput` / `ResultPanel` / `ErrorAlert` 等）；
- 每个工具组件的**基本渲染**与**错误态**。

**禁止**

- 在测试中访问真实网络（**必须** `vi.mock` 拦截 `api/client`）；
- 使用 `setTimeout` 固定等待（**应当**用 `vi.useFakeTimers()`）。

### 4.2 端到端（Playwright）

**必须**覆盖：

1. 全部 98 个工具路由可访问，页面无控制台错误；
2. 首页搜索 → 工具页 → 运行 → 复制 / 下载 全链路；
3. 语言切换后刷新仍保持，`<html lang>` 正确；
4. 主题与主色切换后刷新仍保持；
5. 侧栏折叠状态持久化；移动端抽屉可开合；
6. 键盘可完成：打开搜索 → 选择 → 运行 → 复制 → 关闭对话框；
7. 截图基线：浅色 / 深色 × 首页 / 工具页 / 统计页。

### 4.3 可访问性

- 用 `axe-core` 扫描关键页面，**不得**有 serious / critical 问题；
- 对比度：预设主色在浅色与深色下均需 ≥ 4.5:1。

---

## 5. 文档校验

CI 中**必须**执行：

| 检查 | 工具 |
| --- | --- |
| Markdown 风格 | `markdownlint-cli2` |
| 链接可达 | `lychee` |
| catalog ↔ 文档覆盖率 | `scripts/check-docs.mjs` |
| 错误码 ↔ 文档覆盖率 | `scripts/check-docs.mjs` |
| 环境变量 ↔ 文档覆盖率 | `scripts/check-docs.mjs` |
| 禁词 | `scripts/check-docs.mjs` |

---

## 6. 覆盖率目标

| 范围 | 目标 |
| --- | --- |
| `backend/internal/cryptosvc` | ≥ 80% |
| `backend/internal/toolerr` | 100% |
| `backend/internal/httpx` | ≥ 80% |
| `backend/internal/tools/*` | 每个工具的正例 + 反例路径全覆盖（行覆盖 ≥ 70%） |
| 前端 `composables/` `api/` | ≥ 80% |
| 前端 `components/` | ≥ 60% |

覆盖率**不是**唯一质量指标；关键路径的**断言正确性**优先于行覆盖率。

---

## 7. 测试数据

- **必须**放在 `testdata/` 目录，不写在测试文件里。
- 私钥 / 证书等夹具**必须**是专用于测试的固定值，**禁止**使用生产材料。
- 夹具**必须**注明来源（自签生成 / 国标向量 / 第三方公开向量）。
- 大文件（> 1 MB）**必须**在测试中生成，不入库。

---

## 8. 禁止清单

- ❌ 只断言「没报错」。
- ❌ 只断言 HTTP 状态而不校验 `error.code`。
- ❌ 测试访问真实网络。
- ❌ 在测试中使用真实密钥或生产证书。
- ❌ 用固定 `sleep` 等待异步结果。
- ❌ 提交大体积二进制夹具到仓库。
- ❌ 跳过 `-race`。
