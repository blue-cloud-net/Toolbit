// Package api 是 HTTP 分发器：查目录、协商 Content-Type、限流、超时、
// Body 限制、Origin 校验与注册路由。
//
// 本包**不实现**任何业务算法。
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/config"
	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/stats"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Server 是 HTTP 服务器。
type Server struct {
	cfg     *config.Config
	logger  *slog.Logger
	catalog *catalog.Catalog
	tools   *tools.Registry
	stats   *stats.Store
	writer  *stats.Writer
	web     http.Handler

	limiter *rateLimiter
}

// Options 是 Server 的构造参数。
type Options struct {
	// Config 是运行配置。
	Config *config.Config
	// Logger 是日志器。
	Logger *slog.Logger
	// Catalog 是工具目录。
	Catalog *catalog.Catalog
	// Tools 是工具实现注册表。
	Tools *tools.Registry
	// Stats 是统计库；可为 nil（统计关闭）。
	Stats *stats.Store
	// StatsWriter 是统计写入器；可为 nil（统计关闭）。
	StatsWriter *stats.Writer
	// Web 是静态资源处理器；可为 nil（前端未部署）。
	Web http.Handler
}

// New 构造服务器。
func New(opts Options) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		cfg:     opts.Config,
		logger:  logger,
		catalog: opts.Catalog,
		tools:   opts.Tools,
		stats:   opts.Stats,
		writer:  opts.StatsWriter,
		web:     opts.Web,
		limiter: newRateLimiter(),
	}
}

// Handler 返回配置好的路由。
//
// 中间件顺序（由外向内）：
// Recovery → RequestID → SecurityHeaders → AccessLog → BodyLimit →
// OriginCheck → RateLimit（仅 /api 分组）→ Stats
func (s *Server) Handler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.RedirectTrailingSlash = false

	r.NoRoute(s.handleNotFound)
	r.NoMethod(s.handleMethodNotAllowed)

	// 全局中间件。
	r.Use(httpx.Recovery(s.logger))
	r.Use(httpx.RequestID())
	r.Use(httpx.SecurityHeaders())
	r.Use(httpx.AccessLog(s.logger))
	r.Use(httpx.BodyLimit(s.cfg.MaxBodyBytes))

	api := r.Group("/api")
	api.Use(httpx.OriginCheck(s.cfg.AllowedOrigins))
	if s.cfg.RateLimitEnabled {
		api.Use(s.rateLimit())
	}

	api.GET("/health", s.handleHealth)
	api.GET("/v1/meta", s.handleMeta)
	api.GET("/v1/tools", s.handleCatalog)
	api.Any("/v1/tools/:slug", s.handleTool)

	// 探针端点不启用限流，也不参与统计。
	r.GET("/healthz", s.handleHealthz)
	r.GET("/readyz", s.handleReadyz)

	// 内部端点：仅在显式开启时注册。
	if s.cfg.EnableInternalEndpoints {
		r.GET("/internal/version", s.handleInternalVersion)
	}

	// 静态资源与 SPA 回退。
	if s.web != nil {
		r.NoRoute(func(c *gin.Context) { s.web.ServeHTTP(c.Writer, c.Request) })
	}

	return r
}

// handleNotFound 处理未匹配的路径。
func (s *Server) handleNotFound(c *gin.Context) {
	if s.web != nil {
		s.web.ServeHTTP(c.Writer, c.Request)
		return
	}
	httpx.Fail(c, toolerr.New(toolerr.CodeToolNotFound))
}

// handleMethodNotAllowed 处理方法不匹配。
func (s *Server) handleMethodNotAllowed(c *gin.Context) {
	httpx.Fail(c, toolerr.Newf(toolerr.CodeMethodNotAllowed, map[string]any{
		"method":  c.Request.Method,
		"allowed": http.MethodPost,
	}))
}

// StartLimiterCleanup 启动限流桶的后台回收。
func (s *Server) StartLimiterCleanup(ctx context.Context) {
	s.limiter.StartCleanupLoop(ctx)
}

// categoryOf 返回工具所属分类。
func (s *Server) categoryOf(slug string) string {
	if t, ok := s.catalog.Tool(slug); ok {
		return t.Category
	}
	return ""
}
