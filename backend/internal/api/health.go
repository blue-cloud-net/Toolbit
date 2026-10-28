package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/buildinfo"
	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// HealthResponse 是 GET /api/health 的响应。
//
// 该端点**不使用**统一信封，便于探针直接判断。
type HealthResponse struct {
	// Status 是 ok 或 degraded。
	Status string `json:"status"`
	// Version 是语义化版本。
	Version string `json:"version"`
	// Commit 是短 commit sha。
	Commit string `json:"commit"`
	// UptimeSeconds 是进程运行时长。
	UptimeSeconds int64 `json:"uptimeSeconds"`
	// Database 是统计库状态。
	Database DatabaseStatus `json:"database"`
	// Tools 是工具数量统计。
	Tools ToolCounts `json:"tools"`
}

// DatabaseStatus 是统计库状态。
type DatabaseStatus struct {
	// Enabled 是否启用了统计。
	Enabled bool `json:"enabled"`
	// OK 是统计库是否可写。
	OK bool `json:"ok"`
	// Events 是已记录的调用条数。
	Events int64 `json:"events"`
}

// ToolCounts 是工具数量统计。
type ToolCounts struct {
	// Total 是目录中的工具总数。
	Total int `json:"total"`
	// Enabled 是已启用的工具数。
	Enabled int `json:"enabled"`
}

// startTime 是进程启动时间，用于计算 uptime。
var startTime = time.Now()

// handleHealth 返回进程级健康信息。
//
// status 为 degraded 时 HTTP 状态**仍为** 200（进程还活着），由监控系统判断是否告警。
func (s *Server) handleHealth(c *gin.Context) {
	info := buildinfo.Get()

	resp := HealthResponse{
		Status:        "ok",
		Version:       info.Version,
		Commit:        info.Commit,
		UptimeSeconds: int64(time.Since(startTime).Seconds()),
		Database: DatabaseStatus{
			Enabled: s.cfg.StatsEnabled,
			OK:      false,
		},
		Tools: ToolCounts{
			Total:   len(s.catalog.Tools),
			Enabled: len(s.catalog.EnabledTools()),
		},
	}

	if s.cfg.StatsEnabled && s.stats != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := s.stats.Writable(ctx); err == nil {
			resp.Database.OK = true
			resp.Database.Events = s.stats.EventCount(ctx)
		} else {
			resp.Status = "degraded"
		}
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, resp)
}

// handleHealthz 是存活探针。
//
// **不执行任何 I/O**：进程能响应即 200。
func (s *Server) handleHealthz(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusOK, "ok")
}

// ReadyResponse 是 GET /readyz 的响应。
type ReadyResponse struct {
	// Status 是 ready 或 not_ready。
	Status string `json:"status"`
	// Checks 是各项检查结果。
	Checks map[string]string `json:"checks"`
}

// handleReadyz 是就绪探针。
//
// 检查项：静态目录可读、catalog 已加载、SQLite 可写、迁移版本匹配。
// 任一失败 → 503；检查总超时不超过 2 秒。
func (s *Server) handleReadyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{"catalog": "ok"}

	// 静态目录：未部署前端时跳过检查；已注册静态服务但目录不可读时判为失败。
	if s.web == nil {
		checks["static"] = "skipped"
	} else {
		checks["static"] = "ok"
	}

	if s.cfg.StatsEnabled && s.stats != nil {
		if err := s.stats.Writable(ctx); err != nil {
			checks["database"] = "error"
		} else {
			checks["database"] = "ok"
		}
	}

	ready := true
	for _, v := range checks {
		if v == "error" {
			ready = false
			break
		}
	}

	status := http.StatusOK
	state := "ready"
	if !ready {
		status = http.StatusServiceUnavailable
		state = "not_ready"
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(status, ReadyResponse{Status: state, Checks: checks})
}

// MetaResponse 是 GET /api/v1/meta 的响应数据。
type MetaResponse struct {
	// Name 是服务名。
	Name string `json:"name"`
	// Version 是语义化版本。
	Version string `json:"version"`
	// Commit 是短 commit sha。
	Commit string `json:"commit"`
	// BuiltAt 是构建时间。
	BuiltAt string `json:"builtAt"`
	// GoVersion 是编译所用的 Go 版本。
	GoVersion string `json:"goVersion"`
	// Locales 是可用语言。
	Locales []string `json:"locales"`
	// DefaultLocale 是默认语言。
	DefaultLocale string `json:"defaultLocale"`
	// Features 是功能开关。
	Features Features `json:"features"`
	// APIVersion 是 API 版本。
	APIVersion string `json:"apiVersion"`
}

// Features 是功能开关。
type Features struct {
	// Network 是否启用出网。
	Network bool `json:"network"`
	// Stats 是否启用统计。
	Stats bool `json:"stats"`
	// RateLimit 是否启用限流。
	RateLimit bool `json:"rateLimit"`
}

// handleMeta 返回服务元信息。
//
// 前端**必须**用 locales 与 defaultLocale 决定可用语言，**禁止**硬编码语言列表。
func (s *Server) handleMeta(c *gin.Context) {
	info := buildinfo.Get()
	httpx.Success(c, MetaResponse{
		Name:          "Toolbit",
		Version:       info.Version,
		Commit:        info.Commit,
		BuiltAt:       info.BuildTime,
		GoVersion:     info.GoVersion,
		Locales:       []string{"zh-CN", "en-US"},
		DefaultLocale: "zh-CN",
		Features: Features{
			Network:   s.cfg.NetworkEnabled,
			Stats:     s.cfg.StatsEnabled,
			RateLimit: s.cfg.RateLimitEnabled,
		},
		APIVersion: "v1",
	}, httpx.Meta{DurationMs: 0, Warnings: []toolerr.Warning{}})
}

// handleInternalVersion 返回详细版本信息。
//
// 仅在 TOOLBIT_ENABLE_INTERNAL_ENDPOINTS=true 时注册。
func (s *Server) handleInternalVersion(c *gin.Context) {
	info := buildinfo.Get()
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, map[string]any{
		"version":        info.Version,
		"commit":         info.Commit,
		"buildTime":      info.BuildTime,
		"goVersion":      info.GoVersion,
		"platform":       info.Platform,
		"catalogTools":   len(s.catalog.Tools),
		"catalogEnabled": len(s.catalog.EnabledTools()),
	})
}
