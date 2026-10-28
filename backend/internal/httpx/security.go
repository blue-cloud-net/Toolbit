package httpx

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 安全响应头。
const (
	// HeaderContentTypeOptions 禁止浏览器嗅探响应类型。
	HeaderContentTypeOptions = "X-Content-Type-Options"
	// HeaderReferrerPolicy 禁止在 Referer 中泄露路径。
	HeaderReferrerPolicy = "Referrer-Policy"
	// HeaderVersion 是服务版本响应头。
	HeaderVersion = "X-Toolbit-Version"
	// HeaderCommit 是 commit 响应头。
	HeaderCommit = "X-Toolbit-Commit"
	// HeaderBuildTime 是构建时间响应头。
	HeaderBuildTime = "X-Toolbit-Build-Time"
	// HeaderDurationMs 是耗时响应头。
	HeaderDurationMs = "X-Duration-Ms"
	// HeaderInputBytes 是请求体字节数响应头。
	HeaderInputBytes = "X-Input-Bytes"
	// HeaderOutputBytes 是响应体字节数响应头。
	HeaderOutputBytes = "X-Output-Bytes"
)

// SecurityHeaders 设置安全响应头。
//
// **不**设置 Cache-Control：缓存策略由各个响应自己决定
// （工具响应为 no-store，目录端点为 public）。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set(HeaderContentTypeOptions, "nosniff")
		h.Set(HeaderReferrerPolicy, "no-referrer")
		c.Next()
	}
}

// AccessLog 记录访问日志。
//
// 只记录白名单字段：requestId、slug、status、durationMs、字节数、errorCode。
// **禁止**记录请求体、响应体、UA 原文、Referer 原文与原始 IP。
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		rec := Recorder(c)

		c.Next()

		duration := time.Since(start)
		attrs := []any{
			"requestId", RequestIDFrom(c),
			"status", rec.StatusCode(),
			"durationMs", duration.Milliseconds(),
			"reqBytes", c.Request.ContentLength,
			"respBytes", rec.BytesWritten(),
		}
		if slug := c.GetString(slugKey); slug != "" {
			attrs = append(attrs, "slug", slug)
		}
		if code := c.GetString(errorCodeKey); code != "" {
			attrs = append(attrs, "errorCode", code)
		}
		logger.Info("request", attrs...)
	}
}

// slugKey 是上下文中的 slug 键名。
const slugKey = "httpx.slug"

// errorCodeKey 是上下文中的错误码键名。
const errorCodeKey = "httpx.errorCode"

// SetSlug 把工具标识写入上下文，供访问日志使用。
func SetSlug(c *gin.Context, slug string) {
	c.Set(slugKey, slug)
}

// SetErrorCode 把错误码写入上下文，供访问日志使用。
func SetErrorCode(c *gin.Context, code string) {
	c.Set(errorCodeKey, code)
}

// Recovery 捕获 panic，返回 500 而不是断开连接。
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic 已恢复",
					"requestId", RequestIDFrom(c),
					"error", r,
				)
				if !c.Writer.Written() {
					Failure(c, toolerr.New(toolerr.CodeInternalError), NewMeta())
					return
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
