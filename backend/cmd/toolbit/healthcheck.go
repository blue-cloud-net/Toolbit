// 容器内探活子命令。
//
// 契约见 docs/api/admin.md §9.1：
//   - 访问 `http://127.0.0.1:<TOOLBIT_ADDR 的端口>/readyz`；
//   - 状态非 200 时退出码为 1；
//   - 总超时不超过 3 秒；
//   - **不**输出任何敏感信息（不打印响应体、不打印配置值）。
//
// 之所以内置子命令而不是在镜像里装 curl：运行镜像只需 glibc + Tongsuo 运行库，
// 体积最小且不引入额外进程依赖（见 docs/standards/08-deploy.md §2）。
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/config"
)

const (
	// healthcheckTimeout 是探活请求的总超时（含连接、TLS 无关、读响应）。
	healthcheckTimeout = 3 * time.Second
	// healthcheckPath 是探活端点。
	//
	// 用 `/readyz` 而非 `/healthz`：前者额外检查静态资源与统计库可写，
	// 因此「容器 healthy」等价于「能真正对外服务」。
	healthcheckPath = "/readyz"
	// healthcheckLoopback 是监听地址无法直接连接时的回环替代。
	healthcheckLoopback = "127.0.0.1"
)

// 退出码。
const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

// runHealthcheck 执行一次探活，返回进程退出码。
//
// lookup 是环境变量读取入口，便于测试注入；生产环境传 os.LookupEnv。
func runHealthcheck(stdout, stderr io.Writer, lookup func(string) (string, bool)) int {
	url, err := healthcheckURL(listenAddr(lookup))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "健康检查失败：%v\n", err)
		return exitFail
	}

	// Proxy 显式置空：宿主或编排注入的 HTTP_PROXY 会把回环请求转发到外部代理，
	// 表现为「容器内 curl 正常、healthcheck 超时」。
	client := &http.Client{
		Timeout:   healthcheckTimeout,
		Transport: &http.Transport{Proxy: nil},
	}

	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "健康检查失败：请求构造失败（%s）\n", healthcheckPath)
		return exitFail
	}

	resp, err := client.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "健康检查失败：%s 不可达\n", healthcheckPath)
		return exitFail
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		_, _ = fmt.Fprintf(stderr, "健康检查失败：%s 返回 %d\n", healthcheckPath, resp.StatusCode)
		return exitFail
	}

	_, _ = fmt.Fprintln(stdout, "ok")
	return exitOK
}

// listenAddr 读取监听地址配置，未设置时回落默认值。
func listenAddr(lookup func(string) (string, bool)) string {
	if v, ok := lookup(config.EnvAddr); ok && v != "" {
		return v
	}
	return config.DefaultAddr
}

// healthcheckURL 把监听地址换算成探活 URL。
//
// 监听地址常见四种形态：`:30234`、`0.0.0.0:30234`、`[::]:30234`、`127.0.0.1:30234`。
// 前三种都不能直接连接（或不该连接），统一换成回环地址。
func healthcheckURL(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("监听地址缺少端口（%s）", config.EnvAddr)
	}
	if port == "" {
		return "", fmt.Errorf("监听地址缺少端口（%s）", config.EnvAddr)
	}

	switch host {
	case "", "0.0.0.0", "::":
		host = healthcheckLoopback
	}

	return "http://" + net.JoinHostPort(host, port) + healthcheckPath, nil
}
