# syntax=docker/dockerfile:1.7

# ============================================================================
#  Toolbit — 单镜像（前端静态资源 + Go 后端 + Tongsuo 运行库）
# ----------------------------------------------------------------------------
#  构建上下文是**仓库根目录**：
#
#      docker build -f deployments/Dockerfile -t bluecloud/toolbit:local .
#
#  四个阶段：
#      tongsuo   golang:1.25-bookworm   → 编译 Tongsuo 8.4.0 到 /opt/tongsuo
#      frontend  node:22-alpine         → 产出 frontend/dist
#      backend   golang:1.25-bookworm   → 产出 /out/toolbit（cgo 链接 Tongsuo）
#      runtime   debian:bookworm-slim   → 携带二进制 + 前端产物 + Tongsuo 运行库
#
#  关键约束（见 docs/standards/08-deploy.md）：
#    1. tongsuo-go 的 cgo 指令把路径写死为 /opt/tongsuo/include 与
#       /opt/tongsuo/lib64，并自带 -Wl,-rpath,/opt/tongsuo/lib64。
#       因此安装前缀必须是 /opt/tongsuo，库目录必须是 lib64
#       （arm64 默认会装到 lib，所以显式 --libdir=lib64）。
#    2. 运行镜像**不得**设 LD_LIBRARY_PATH / LD_PRELOAD / GOPATH / GOPROXY：
#       rpath 已自带，多余的 LD_* 会遮蔽系统 OpenSSL 并破坏动态链接。
#    3. 运行镜像基于 debian（glibc），**禁止** alpine（musl）或 scratch。
#    4. 基础镜像固定 digest，避免上游漂移导致构建不可复现。
# ============================================================================

# 基础镜像（digest 固定）。
ARG GO_IMAGE=golang:1.25-bookworm@sha256:3b4a11519ad929d1e1d261a12cff056f0c85b735253d7d861346b9c6f8b36437
ARG NODE_IMAGE=node:22-alpine@sha256:fe4f6d74e20370acb04ad889077d14d0f64a9c8ea6a42bc60cddf1e844699914
ARG RUNTIME_IMAGE=debian:bookworm-slim@sha256:99c0b25c46d0214d5a2091092e7908907dcaecd6a962cde0ac700acee222b6f3

# 依赖版本（构建期可覆盖）。
ARG TONGSUO_VERSION=8.4.0
ARG TONGSUO_URL=https://github.com/Tongsuo-Project/Tongsuo/archive/refs/tags/8.4.0.tar.gz
# Tongsuo 8.4.0 源码包 sha256：网络受限时可用 TONGSUO_URL 指向镜像源，
# 但**必须**同时提供对应的 TONGSUO_SHA256，不允许跳过校验。
ARG TONGSUO_SHA256=57c2741750a699bfbdaa1bbe44a5733e9c8fc65d086c210151cfbc2bbd6fc975
# Go 模块代理：部分网络下 proxy.golang.org 不可达，构建时可覆盖。
ARG GOPROXY=https://goproxy.io,direct


# ---------- 阶段 1：Tongsuo ----------
FROM ${GO_IMAGE} AS tongsuo

ARG TONGSUO_VERSION
ARG TONGSUO_URL
ARG TONGSUO_SHA256

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        perl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# 源码包先校验再解包：校验失败直接中断构建（不留「看起来成功」的模糊状态）。
RUN set -eux; \
    curl -fsSL --retry 3 -o tongsuo.tar.gz "${TONGSUO_URL}"; \
    echo "${TONGSUO_SHA256}  tongsuo.tar.gz" | sha256sum -c -; \
    mkdir -p /src/tongsuo; \
    tar -xzf tongsuo.tar.gz -C /src/tongsuo --strip-components=1; \
    rm -f tongsuo.tar.gz; \
    echo "${TONGSUO_VERSION}" > /src/tongsuo/.toolbit-version

WORKDIR /src/tongsuo

# --libdir=lib64：与 tongsuo-go 的 cgo 指令和 rpath 保持一致（arm64 同样落在 lib64）。
# shared：产出共享库（.so），运行镜像只靠 rpath 找库。
# enable-ntls：后续国密双证书（NTLS）相关工具依赖它，开启后不影响其他算法。
# 注意：`--openssl-shared` 不是 Tongsuo 8.4.0 的 config 选项，会被透传给 gcc 导致构建失败。
# make 与 make install_sw 分开：安装步骤不参与并行编译。
RUN set -eux; \
    ./config --prefix=/opt/tongsuo --libdir=lib64 shared enable-ntls; \
    make -j"$(nproc)"; \
    make install_sw

# 构建期自查：库、头文件与 OPENSSLDIR 必须就位。
RUN set -eux; \
    test -f /opt/tongsuo/include/openssl/opensslv.h; \
    ls -l /opt/tongsuo/lib64/libcrypto.so* /opt/tongsuo/lib64/libssl.so*; \
    mkdir -p /opt/tongsuo/ssl


# ---------- 阶段 2：前端 ----------
FROM ${NODE_IMAGE} AS frontend

WORKDIR /src/frontend

# pnpm 版本与 frontend/package.json 的 packageManager 字段保持一致。
# 单独成层：只要版本不变，这一层始终命中缓存。
RUN --mount=type=cache,target=/root/.npm \
    npm install --global pnpm@12.11.2

# 依赖单独成层：package.json / pnpm-lock.yaml 不变时安装层可命中缓存。
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/pnpm-store \
    pnpm install --frozen-lockfile --store-dir /pnpm-store

COPY frontend/ ./

# pnpm build = vite build（类型检查在 make verify 里单独跑，构建不做二次校验）。
RUN pnpm build


# ---------- 阶段 3：后端 ----------
FROM ${GO_IMAGE} AS backend

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
ARG GOPROXY

# cgo 路径与 tongsuo-go 的硬编码指令一致；rpath 让二进制自带库路径。
ENV CGO_ENABLED=1 \
    CGO_CFLAGS="-I/opt/tongsuo/include -Wno-deprecated-declarations" \
    CGO_LDFLAGS="-L/opt/tongsuo/lib64 -Wl,-rpath,/opt/tongsuo/lib64"

# 构建期需要 Tongsuo 的头文件与共享库（与运行镜像保持同一路径）。
COPY --from=tongsuo /opt/tongsuo /opt/tongsuo

WORKDIR /src/backend

# 依赖单独成层：go.mod / go.sum 不变时下载层可命中缓存。
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    GOPROXY="${GOPROXY}" go mod download

COPY backend/ ./

# 版本信息由 ldflags 注入 internal/buildinfo（见该包注释）。
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOPROXY="${GOPROXY}" go build -trimpath \
      -ldflags "-s -w \
        -X github.com/blue-cloud-net/Toolbit/backend/internal/buildinfo.Version=${VERSION} \
        -X github.com/blue-cloud-net/Toolbit/backend/internal/buildinfo.Commit=${COMMIT} \
        -X github.com/blue-cloud-net/Toolbit/backend/internal/buildinfo.BuildTime=${BUILD_TIME}" \
      -o /out/toolbit ./cmd/toolbit

# 构建期自查：动态链接必须解析到镜像内的 Tongsuo 路径（rpath 生效），
# 且不能有未解析的库。产物一旦漏链，运行期才炸，代价更高。
RUN set -eux; \
    ldd /out/toolbit > /tmp/ldd.txt; \
    grep -q '/opt/tongsuo/lib64/libcrypto' /tmp/ldd.txt; \
    ! grep -q 'not found' /tmp/ldd.txt


# ---------- 阶段 4：运行 ----------
FROM ${RUNTIME_IMAGE} AS runtime

LABEL org.opencontainers.image.title="Toolbit" \
      org.opencontainers.image.description="蓝云工具箱：IT 工具箱后端与前端静态资源" \
      org.opencontainers.image.licenses="Apache-2.0"

# ca-certificates：出网工具访问 https 需要根证书；
# tzdata：容器时区（TZ），统计按日聚合依赖本地时区。
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 10001 app \
    && useradd --uid 10001 --gid app --no-create-home --home-dir /app --shell /usr/sbin/nologin app

WORKDIR /app

COPY --from=backend /out/toolbit /app/toolbit
# Tongsuo 运行库：路径必须与二进制里写死的 rpath 一致（/opt/tongsuo/lib64）。
COPY --from=tongsuo /opt/tongsuo/lib64/ /opt/tongsuo/lib64/
# OPENSSLDIR 指向 /opt/tongsuo/ssl：缺省配置会让部分算法/Provider 初始化失败。
COPY --from=tongsuo /opt/tongsuo/ssl/ /opt/tongsuo/ssl/
# 前端构建产物：由 Go 进程直接托管（静态资源 + SPA 回退）。
COPY --from=frontend /src/frontend/dist/ /app/web/

# 数据目录预建并把属主交给 app：命名卷首次挂载会继承该目录属主（10001:10001），
# 否则 SQLite 会因目录属主是 root 而报 unable to open database file。
RUN mkdir -p /app/data && chown -R 10001:10001 /app/data

USER 10001:10001

ENV TZ=Asia/Shanghai \
    TOOLBIT_ADDR=:30234 \
    TOOLBIT_WEB_ROOT=/app/web \
    TOOLBIT_DB_PATH=/app/data/toolbit.db

EXPOSE 30234

# exec 形式：让进程直接收到 SIGTERM，走 main 里的优雅退出路径。
STOPSIGNAL SIGTERM

# 探针用内置子命令（不引入 curl / wget）：契约见 docs/api/admin.md §9.1。
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD ["/app/toolbit", "healthcheck"]

ENTRYPOINT ["/app/toolbit"]
