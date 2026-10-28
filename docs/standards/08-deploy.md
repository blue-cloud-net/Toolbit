# 部署规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

---

## 1. 交付形态

| 形态 | 用途 | 说明 |
| --- | --- | --- |
| Docker 镜像 | 主力 | 单容器包含前端静态资源 + Go 后端 + Tongsuo 运行库 |
| 源码构建 | 二次开发 | `make build` 产出 `bin/toolbit` + `frontend/dist` |

**必须**提供 Dockerfile 与 `docker-compose.yml`。

**禁止**提供「需要用户在宿主机单独安装 Tongsuo」的部署方式作为唯一路径。

---

## 2. 镜像

### 2.1 多阶段

```text
builder-base  →  go build（CGO_ENABLED=1, tongsuo headers/libs）
                →  pnpm build → frontend/dist
runtime-base  →  拷贝二进制 + dist + tongsuo 运行库 + CA 证书
```

**必须**

- 运行时镜像**不含** Go / Node / 编译工具链；
- Tongsuo 共享库用 `-Wl,-rpath,/opt/tongsuo/lib64` 链接（**禁止**依赖 `LD_LIBRARY_PATH`）；
- 运行时镜像内 Tongsuo 库路径与 rpath **必须**一致；
- 基础镜像**必须**固定 digest；
- 非 root 用户：`app`，UID/GID **必须**为 `10001:10001`；
- `WORKDIR /app`；
- 数据目录 `/app/data`（统计库），**必须**属主 `10001:10001`；
- 暴露 `8080`；
- `ENTRYPOINT ["/app/toolbit"]`（exec 形式，**必须**让进程收到 SIGTERM）。

### 2.2 基础镜像选择

- 运行时基于 **debian slim**（需 `ca-certificates` 与 glibc）；
- **禁止**基于 `alpine`（musl 与 tongsuo-go 的 cgo 兼容性风险高）；
- **禁止**基于 `scratch`（Tongsuo 需要 glibc 与 CA）。

---

## 3. 配置

全部通过**环境变量**注入，前缀 `TOOLBIT_`。

| 变量 | 类型 | 默认 | 说明 |
| --- | --- | --- | --- |
| `TOOLBIT_ADDR` | string | `:8080` | 监听地址 |
| `TOOLBIT_WEB_ROOT` | string | `./frontend/dist` | 静态资源目录 |
| `TOOLBIT_DB_PATH` | string | `./data/stats.db` | SQLite 路径 |
| `TOOLBIT_LOG_LEVEL` | string | `info` | `debug`/`info`/`warn`/`error` |
| `TOOLBIT_LOG_FORMAT` | string | `json` | `json`/`text` |
| `TOOLBIT_TRUST_PROXY` | bool | `false` | 是否信任 `X-Forwarded-For` |
| `TOOLBIT_ALLOWED_ORIGINS` | string | 空 | 逗号分隔的 Origin 白名单 |
| `TOOLBIT_ENABLE_INTERNAL_ENDPOINTS` | bool | `false` | 是否开放 `/internal/*` |
| `TOOLBIT_RATE_LIMIT_ENABLED` | bool | `true` | 限流开关 |
| `TOOLBIT_STATS_ENABLED` | bool | `true` | 匿名统计开关 |
| `TOOLBIT_NETWORK_ENABLED` | bool | `true` | 出网工具总开关 |
| `TOOLBIT_NETWORK_ALLOWLIST` | string | 空 | 额外允许出网的主机白名单 |
| `TOOLBIT_REQUEST_TIMEOUT` | duration | `30s` | 全局兜底超时 |
| `TOOLBIT_MAX_BODY_BYTES` | int | `104857600` | 全局兜底上限（100 MB） |
| `TOOLBIT_SHUTDOWN_TIMEOUT` | duration | `15s` | 优雅退出超时 |
| `TOOLBIT_STATS_TOKEN` | string | 空 | 统计详情端点的 Token（未配置则端点 404） |

**必须**

- 启动时校验全部配置，**禁止**静默使用非法值的默认值；
- 未知的环境变量**必须**在启动日志中列出（帮助排错），但**不**报错；
- 配置**不**支持运行时热更新。

---

## 4. 优雅退出

**必须**

- 收到 `SIGINT` / `SIGTERM` 后：停止接受新请求 → 等待进行中请求（≤ `TOOLBIT_SHUTDOWN_TIMEOUT`）→ 关闭统计写入 → 关闭数据库 → 退出；
- 超时未完成**必须**强制退出（exit code 1）；
- **禁止**在收到信号后立即 `os.Exit` 而不等待。

**必须**在 Docker 中用 `STOPSIGNAL SIGTERM`（默认即是）。

---

## 5. Docker Compose

```yaml
services:
  toolbit:
    build: .
    image: bluecloud/toolbit:0.1.0
    container_name: toolbit
    ports:
      - "8080:8080"
    environment:
      TOOLBIT_ADDR: ":8080"
      TOOLBIT_DB_PATH: "/app/data/stats.db"
    volumes:
      - toolbit-data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/app/toolbit", "healthcheck"]
      interval: 30s
      timeout: 5s
      retries: 3
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL

volumes:
  toolbit-data:
```

**必须**

- **必须**使用**命名卷**（不用 bind mount，避免属主与权限问题，见 §9）；
- `read_only: true` + `tmpfs` 只在确有必要时启用；
- `cap_drop: ALL`、`no-new-privileges:true`；
- **禁止**在 compose 中写明文密钥（本项目无密钥需求，但口令类配置**必须**走 `.env` 且 `.env` 已在 `.gitignore`）。

---

## 6. 反向代理

本服务**不内置 TLS**（见 [architecture.md D8](../architecture.md)）。对外暴露**必须**由反向代理终止 TLS。

**必须**：

- 设置 `X-Forwarded-Proto` / `X-Forwarded-For`；
- 若启用 `TOOLBIT_TRUST_PROXY=true`，代理**必须**覆写而非追加这些头；
- 反向代理**必须**限制上传体积（与 `TOOLBIT_MAX_BODY_BYTES` 协调）；
- WebSocket 路径**必须**允许升级（`websocket-test`）。

示例（Nginx）：

```nginx
server {
    listen 443 ssl http2;
    server_name toolbit.example.com;
    ssl_certificate     /etc/letsencrypt/live/toolbit.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/toolbit.example.com/privkey.pem;
    client_max_body_size 100m;
    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Request-Id      $request_id;
    }
}
```

---

## 7. 资源限制

**必须**在部署侧设置：

| 资源 | 建议值 |
| --- | --- |
| CPU | ≤ 2 核（重操作并发已限为 2） |
| 内存 | ≤ 1 GB（文件加密峰值 ≈ 2×chunkSize） |
| 磁盘 | 数据卷 ≥ 1 GB（统计库 + 90 天保留） |
| `ulimit -n` | ≥ 4096 |
| `ulimit -u` | ≥ 4096 |

**必须**设置容器内存上限，避免单次大文件请求 OOM。

---

## 8. CI/CD

### 8.1 流水线阶段

| 阶段 | 内容 |
| --- | --- |
| lint | `gofmt` / `go vet` / `golangci-lint` / `eslint` / `prettier` / `markdownlint` |
| test | `go test ./... -race` / `pnpm test` / `pnpm build` |
| scan | `govulncheck` / `pnpm audit` / Trivy 镜像扫描 |
| build | 构建前端 → 构建 Go 二进制（`CGO_ENABLED=1`）→ 构建镜像 |
| docs | `scripts/check-docs.mjs` + `lychee` |
| publish | 推送到镜像仓库（打 tag = 语义化版本） |

### 8.2 镜像标签

| 标签 | 指向 |
| --- | --- |
| `0.1.0` | 语义化版本（不可变） |
| `0.1` | 次版本（可移动） |
| `latest` | 最新正式版（可移动） |
| `sha-<short>` | 特定 commit（不可变） |

**必须**至少推 `0.1.0` 与 `sha-<short>`。

**禁止**给可变标签做「就地覆盖」而不重新构建（缓存失效）。

---

## 9. 运维陷阱（已实测）

| 陷阱 | 现象 | 处理 |
| --- | --- | --- |
| bind mount 目录属主 | 容器内 UID 10001 **写不进去** → `unable to open database file` | **必须**用命名卷 |
| `docker cp` 恢复的数据 | 属主是执行 cp 的宿主机用户 | 恢复后 `chown 10001:10001` |
| `.env` 被提交 | 一次 `git add .` 泄露 | `.gitignore` **必须**含 `.env`（已含） |
| 宿主机 socket 挂载 | 镜像内 UID 不在 socket 所属组 → 权限拒绝 | `--group-add <socket gid>` |

---

## 10. 升级流程

1. 读 [CHANGELOG.md](../../CHANGELOG.md) 的破坏性变更；
2. `docker compose pull`；
3. `docker compose up -d`；
4. 检查 `/readyz` 与 `/internal/version`；
5. 确认统计库迁移日志正常。

**必须**：升级**不得**删除数据卷。数据库迁移**必须**向前兼容（只增列 / 加表）。

**禁止**跨大版本直接升级（需按 changelog 逐版本执行）。

---

## 11. 备份

需要备份的内容：

| 内容 | 位置 | 必要性 |
| --- | --- | --- |
| 统计库 | `/app/data/stats.db` | 可选（匿名数据） |
| 配置 | 部署环境变量 / `.env` | 必需（外部管理） |
| 镜像 | 镜像仓库 | 必需 |

**禁止**备份用户提交的密钥材料（本项目不持久化）。

备份 SQLite **必须**用 `.backup` 命令或 `VACUUM INTO`，**禁止**在写入期间直接拷贝文件。

---

## 12. 禁止清单

- ❌ 基于 `alpine` / `scratch` 构建镜像。
- ❌ 依赖 `LD_LIBRARY_PATH` 找 Tongsuo 库。
- ❌ 容器内以 root 运行服务。
- ❌ bind mount 数据目录（属主问题）。
- ❌ 内置 TLS 到应用层。
- ❌ 公开 `/internal/*` 端点。
- ❌ 升级时删除数据卷。
- ❌ 在 compose / 仓库中写明文口令。
