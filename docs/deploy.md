# 部署指南

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

本文面向**运维 / 使用者**，说明如何把 Toolbit 跑起来。

架构与规范见 [architecture.md](architecture.md) 与 [standards/08-deploy.md](standards/08-deploy.md)。

---

## 1. 方式一：Docker（推荐）

镜像是**单容器**：Go 后端 + 前端静态资源 + Tongsuo 运行库，不需要额外的 nginx、Node 或数据库容器。
编排与 Dockerfile 都在 `deployments/`，构建上下文是**仓库根目录**。

### 1.1 快速开始

```bash
git clone git@github.com:blue-cloud-net/Toolbit.git
cd Toolbit
make docker-up          # 首次会编译 Tongsuo，约 5-10 分钟
```

缺 `deployments/.env` 时会自动用 `.env.example` 生成一份（**全部变量都有默认值**，可直接使用）。
访问 <http://localhost:30234>。

### 1.2 直接使用 compose

```bash
cp deployments/.env.example deployments/.env     # 可选，需要改默认值时再建
docker compose -f deployments/docker-compose.yml up -d --build
```

### 1.3 使用预构建镜像

```bash
docker run -d --name toolbit \
  -p 30234:30234 \
  -v toolbit-data:/app/data \
  --restart unless-stopped \
  --cap-drop ALL \
  --security-opt no-new-privileges:true \
  --memory 1g \
  bluecloud/toolbit:0.5.0
```

> 镜像内已设 `TOOLBIT_WEB_ROOT=/app/web` 与 `TOOLBIT_DB_PATH=/app/data/toolbit.db`，
> 因此 `docker run` **不需要**再传这两个变量。

### 1.4 验证

```bash
make docker-smoke        # 容器状态 + healthz / readyz / 目录 / 前端首页
docker exec toolbit /app/toolbit healthcheck; echo $?   # 0 = 就绪
```

### 1.5 编排变量

除 `TOOLBIT_*` 之外，编排层还有 6 个变量（见 `deployments/.env.example`）：

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `TAG` | `local` | 镜像标签，同时注入 `buildinfo.Version` |
| `COMMIT` | `unknown` | 短 commit，注入 `buildinfo.Commit` |
| `BUILD_TIME` | `unknown` | 构建时间（RFC3339 UTC），注入 `buildinfo.BuildTime` |
| `HTTP_PORT` | `30234` | 宿主机对外端口 |
| `TZ` | `Asia/Shanghai` | 容器时区（统计按日聚合依赖它） |
| `MEMORY_LIMIT` | `1g` | 容器内存上限 |

---

## 2. 方式二：源码构建

### 2.1 前置条件

| 依赖 | 版本 | 说明 |
| --- | --- | --- |
| Go | 1.25+ | 需要 `CGO_ENABLED=1` |
| Node.js | 22 LTS | 构建前端 |
| pnpm | 12+ | 包管理（`frontend/package.json` 的 `packageManager` 为准） |
| Tongsuo | 8.4.0 | 提供头文件与运行库 |

> **用 Docker 时不需要在本机安装 Tongsuo**：镜像内由 `tongsuo` 阶段自行编译安装到 `/opt/tongsuo`。

### 2.2 安装 Tongsuo

```bash
# 编译 Tongsuo（shared 产出 .so；lib64 与 tongsuo-go 的 cgo 路径一致）
cd tongsuo
./config --prefix=/opt/tongsuo --libdir=lib64 shared enable-ntls
make -j"$(nproc)"
sudo make install_sw
```

> `--openssl-shared` **不是** Tongsuo 8.4.0 的 config 选项，传了会被透传给 gcc 并报
> `unrecognized command-line option`。共享库用 `shared` 开启。

### 2.3 构建

后端与前端是**两个独立构建单元**，分别在各自目录内执行。

```bash
# 前端（frontend/）
cd frontend
pnpm install
pnpm build          # 产出 frontend/dist

# 后端（backend/，独立 Go 模块）
cd ../backend
export CGO_ENABLED=1
export CGO_CFLAGS="-I/opt/tongsuo/include"
export CGO_LDFLAGS="-L/opt/tongsuo/lib64 -Wl,-rpath,/opt/tongsuo/lib64"
go build -o ../bin/toolbit ./cmd/toolbit
```

### 2.4 运行

```bash
cd ..
TOOLBIT_ADDR=:30234 \
TOOLBIT_WEB_ROOT=./frontend/dist \
TOOLBIT_DB_PATH=./data/stats.db \
./bin/toolbit
```

> `-Wl,-rpath` 让二进制自带库路径，**无需**设置 `LD_LIBRARY_PATH`。

---

## 3. 配置项

全部通过环境变量，前缀 `TOOLBIT_`。

| 变量 | 类型 | 默认 | 说明 |
| --- | --- | --- | --- |
| `TOOLBIT_ADDR` | string | `:30234` | 监听地址 |
| `TOOLBIT_WEB_ROOT` | string | `./frontend/dist` | 前端静态资源目录 |
| `TOOLBIT_DB_PATH` | string | `./data/stats.db` | SQLite 统计库路径 |
| `TOOLBIT_LOG_LEVEL` | string | `info` | `debug` / `info` / `warn` / `error` |
| `TOOLBIT_LOG_FORMAT` | string | `json` | `json` / `text` |
| `TOOLBIT_TRUST_PROXY` | bool | `false` | 是否信任 `X-Forwarded-For` |
| `TOOLBIT_ALLOWED_ORIGINS` | string | 空 | 逗号分隔的 Origin 白名单（默认不启用 CORS） |
| `TOOLBIT_ENABLE_INTERNAL_ENDPOINTS` | bool | `false` | 是否开放 `/internal/*` |
| `TOOLBIT_RATE_LIMIT_ENABLED` | bool | `true` | 限流开关 |
| `TOOLBIT_STATS_ENABLED` | bool | `true` | 匿名统计开关 |
| `TOOLBIT_NETWORK_ENABLED` | bool | `true` | 出网工具总开关 |
| `TOOLBIT_NETWORK_ALLOWLIST` | string | 空 | 额外允许出网的主机（逗号分隔） |
| `TOOLBIT_REQUEST_TIMEOUT` | duration | `30s` | 全局兜底超时 |
| `TOOLBIT_MAX_BODY_BYTES` | int | `104857600` | 全局兜底请求体上限（100 MB） |
| `TOOLBIT_SHUTDOWN_TIMEOUT` | duration | `15s` | 优雅退出等待时长 |
| `TOOLBIT_STATS_TOKEN` | string | 空 | 统计详情端点的 Token（未配置则端点 404） |

**必须**

- 配置在启动时**一次性校验**；非法值导致启动失败，**不**静默回退；
- 未知的环境变量会在启动日志中列出（便于排错），但**不**视为错误；
- 配置**不**支持运行时热更新。

---

## 4. 反向代理与 TLS

> **本服务不内置 TLS。** 容器内监听**明文 HTTP** :30234。
> 公网暴露**必须**由反向代理终止 TLS，否则密钥与明文会经明文网络传输。

### 4.1 Nginx

```nginx
server {
    listen 443 ssl http2;
    server_name toolbit.example.com;

    ssl_certificate     /etc/letsencrypt/live/toolbit.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/toolbit.example.com/privkey.pem;

    client_max_body_size 100m;

    location / {
        proxy_pass         http://127.0.0.1:30234;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Request-Id      $request_id;
    }
}
```

### 4.2 代理侧注意

| 项 | 要求 |
| --- | --- |
| `X-Forwarded-For` | **必须覆写**，**禁止**直接透传客户端值 |
| `Origin` / `Sec-Fetch-Site` | 代理**禁止**注入这两个头（会绕过跨站校验） |
| `X-Request-Id` | 可透传，服务端会校验格式后复用 |
| 上传体积 | 与 `TOOLBIT_MAX_BODY_BYTES` 协调 |
| WebSocket | 需支持 Upgrade（`ssl-check` 使用） |

### 4.3 启用代理信任

只有当反代**确实覆写**上述头时，才设置：

```bash
TOOLBIT_TRUST_PROXY=true
```

**未正确配置反代就开启此项，会导致限流可被轻易绕过。**

---

## 5. 资源与限制

| 资源 | 建议 |
| --- | --- |
| CPU | ≤ 2 核（重操作并发已限为 2） |
| 内存 | ≤ 1 GB（文件加密峰值约 2×分块大小） |
| 磁盘 | 数据卷 ≥ 1 GB |
| `ulimit -n` | ≥ 4096 |

**必须**在容器上设置内存上限，避免单次大文件请求触发 OOM。
`deployments/docker-compose.yml` 已设 `MEMORY_LIMIT`（默认 `1g`），直接 `docker run` 时需自行传 `--memory`。

---

## 6. 数据与持久化

| 内容 | 位置 | 是否必须持久化 |
| --- | --- | --- |
| 匿名统计库 | `/app/data/toolbit.db` | 可选（统计数据） |
| 前端静态资源 | `/app/web`（镜像内） | 否 |
| 配置 | 外部环境变量 / `deployments/.env` | 是 |

**必须**使用**命名卷**：

```yaml
volumes:
  toolbit-data:
```

> ⚠️ **不要**用 bind mount 指向宿主机目录 —— 目录属主通常不是 UID 10001，容器内将无法写库，报 `unable to open database file`。

---

## 7. 运维端点

| 端点 | 用途 |
| --- | --- |
| `GET /healthz` | 存活探针（无 I/O） |
| `GET /readyz` | 就绪探针（静态资源 / catalog / 数据库） |
| `GET /api/health` | 详细健康信息 |
| `GET /api/v1/meta` | 版本与特性 |
| `GET /internal/metrics` | Prometheus 指标（需 `TOOLBIT_ENABLE_INTERNAL_ENDPOINTS=true`） |

**内部端点禁止绑定到公网地址。**

---

## 8. 升级

```bash
docker compose -f deployments/docker-compose.yml pull
docker compose -f deployments/docker-compose.yml up -d
curl -fsS http://localhost:30234/readyz
```

从源码升级（自己构建镜像）：

```bash
make docker-up            # 重新构建并滚动替换容器
make docker-smoke         # 确认 /readyz 与目录端点正常
```

**必须**：升级**不得**删除数据卷；数据库迁移**必须**向前兼容（只增列 / 加表）。

**禁止**跨大版本直接升级 —— 按 [CHANGELOG.md](../CHANGELOG.md) 逐版本执行。

---

## 9. 备份与恢复

需要备份的内容极少：统计数据可丢，配置由外部管理。

**必须**用 SQLite 的 `.backup` 或 `VACUUM INTO` 做备份，**禁止**在写入期间直接拷贝文件。

```bash
docker exec toolbit sqlite3 /app/data/toolbit.db ".backup /app/data/backup.db"
docker cp toolbit:/app/data/backup.db ./stats-backup.db
```

**恢复注意**：`docker cp` 放入的属主是执行 cp 的宿主机用户，需修正：

```bash
docker exec -u 0 toolbit chown 10001:10001 /app/data/toolbit.db
```

---

## 10. 故障排查

| 现象 | 原因 | 处理 |
| --- | --- | --- |
| `unable to open database file` | 数据卷属主 / 权限 | 改用命名卷；或 `docker exec -u 0 toolbit chown -R 10001:10001 /app/data` |
| 容器启动即退出、日志含 `error while loading shared libraries` | Tongsuo 运行库未随镜像发布 / rpath 失效 | 用本仓库 `deployments/Dockerfile` 构建；`docker run --rm --entrypoint ldd <image>` 应指向 `/opt/tongsuo/lib64` |
| `/readyz` 返回 503 | 静态资源不可读或统计库不可写 | 检查 `TOOLBIT_WEB_ROOT`（镜像内为 `/app/web`）与数据卷属主 |
| 容器一直是 `starting` / `unhealthy` | 探针不通过 | `docker exec toolbit /app/toolbit healthcheck; echo $?`，再用 `make docker-logs` 看启动日志 |
| 429 频繁 | 限流触发 | 检查是否有压测或反代未覆写 `X-Forwarded-For` |
| 出网工具报 `SSRF_BLOCKED` | 目标是内网地址 | 预期行为；如需放行请显式配置白名单 |
| 出网工具报 `NETWORK_DISABLED` | 总开关关闭 | 设置 `TOOLBIT_NETWORK_ENABLED=true` |
| 出网工具报证书错误 | 镜像内缺 CA 根证书（自建镜像时才可能） | 安装 `ca-certificates`，官方镜像已含 |
| WebSocket 连不上 | 代理未允许 Upgrade | 检查反代配置 |

---

## 11. 安全建议

1. **必须**在反代后启用 TLS；
2. **必须**保持 `TOOLBIT_TRUST_PROXY=false`（除非反代已正确覆写头）；
3. **必须**保持 `TOOLBIT_ENABLE_INTERNAL_ENDPOINTS=false`（除非内网可达）；
4. **应当**定期更新基础镜像（CI 已用 Trivy 扫描）；
5. **不应当**把服务直接暴露在公网 30234 端口；
6. **应当**在反代层再加一层速率限制。

---

## 12. 相关文档

- [architecture.md](architecture.md) —— 架构与关键决策
- [standards/08-deploy.md](standards/08-deploy.md) —— 部署规范
- [api/admin.md](api/admin.md) —— 管理与运维接口
- [privacy.md](privacy.md) —— 隐私说明
- [../CHANGELOG.md](../CHANGELOG.md) —— 变更记录
