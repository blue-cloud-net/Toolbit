// Command toolbit 是蓝云工具箱的后端进程入口。
//
// 职责：解析配置 → 装配依赖 → 启动 HTTP 服务 → 等待信号 → 优雅退出。
//
// **不**包含任何业务逻辑；工具实现位于 internal/tools。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// 内嵌 IANA 时区数据库：基础镜像通常不含 zoneinfo，
	// 而 timestamp / date-calc 需要按 IANA 名解析时区。
	_ "time/tzdata"

	"github.com/blue-cloud-net/Toolbit/backend/internal/api"
	"github.com/blue-cloud-net/Toolbit/backend/internal/buildinfo"
	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/config"
	"github.com/blue-cloud-net/Toolbit/backend/internal/stats"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolset"
	"github.com/blue-cloud-net/Toolbit/backend/internal/webui"
)

func main() {
	if err := run(); err != nil {
		// 启动失败：只打印错误，不打印任何配置值。
		fmt.Fprintln(os.Stderr, "启动失败：", err)
		os.Exit(1)
	}
}

// run 负责启动与优雅退出，错误全部返回给 main 处理。
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)

	cat, err := catalog.Load()
	if err != nil {
		return fmt.Errorf("加载工具目录失败: %w", err)
	}

	registry := tools.NewRegistry(toolset.Definitions())
	missing, orphan := tools.CrossCheck(cat, registry)
	logger.Info("工具注册表",
		"implemented", registry.Len(),
		"missing", len(missing),
		"orphan", len(orphan),
		"catalog", cat.String(),
	)
	if len(orphan) > 0 {
		// 孤儿实现意味着代码与目录不一致，属于部署错误。
		return fmt.Errorf("存在未登记的工具实现：%v", orphan)
	}

	var (
		store  *stats.Store
		writer *stats.Writer
	)
	if cfg.StatsEnabled {
		store, err = stats.Open(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("初始化统计库失败: %w", err)
		}
		writer = stats.NewWriter(store, logger)
		defer func() { _ = store.Close() }()
	}

	var web http.Handler
	if cfg.WebRoot != "" {
		if h, err := webui.New(cfg.WebRoot); err == nil {
			web = h
			logger.Info("静态资源已挂载", "root", cfg.WebRoot)
		} else {
			// 静态资源不可用**不**阻止启动：/readyz 会把它报为未就绪。
			logger.Warn("静态资源不可用", "root", cfg.WebRoot, "error", err)
		}
	}

	srv := api.New(api.Options{
		Config:      cfg,
		Logger:      logger,
		Catalog:     cat,
		Tools:       registry,
		Stats:       store,
		StatsWriter: writer,
		Web:         web,
	})

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 启动后台清理任务。
	bgCtx, cancelBg := context.WithCancel(context.Background())
	defer cancelBg()
	if store != nil {
		store.StartRetentionLoop(bgCtx, logger)
	}
	srv.StartLimiterCleanup(bgCtx)

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("服务已启动", "addr", cfg.Addr, "version", buildinfo.Version)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("HTTP 服务异常退出: %w", err)
		}
		return nil
	case sig := <-sigCh:
		logger.Info("收到退出信号，开始优雅退出", "signal", sig.String())
	}

	// 停止接收新连接，并在超时内等完在途请求。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("优雅退出超时，强制关闭", "error", err)
		_ = httpServer.Close()
	}

	// 落盘剩余统计事件。
	if writer != nil {
		if err := writer.Close(shutdownCtx); err != nil {
			logger.Error("统计写入器关闭失败", "error", err)
		}
	}
	cancelBg()

	logger.Info("已退出")
	return nil
}

// newLogger 按配置构造日志器。
func newLogger(cfg *config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	if cfg.LogFormat == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
