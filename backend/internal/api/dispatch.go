package api

import (
	"context"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/stats"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// handleTool 是工具执行端点。
//
// 校验顺序严格按 docs/api/README.md §8，命中第一个失败即返回：
//
//  1. 路由与工具存在性 → TOOL_NOT_FOUND / TOOL_DISABLED
//  2. 本版是否已实现 → NOT_IMPLEMENTED
//  3. HTTP 方法 → METHOD_NOT_ALLOWED
//  4. Content-Type → UNSUPPORTED_MEDIA_TYPE
//  5. 跨站请求 → CROSS_ORIGIN_BLOCKED（中间件）
//  6. 限流 → RATE_LIMITED（中间件）
//  7. 请求体大小 → TOO_LARGE
//  8. 语法解码 → INVALID_JSON / INVALID_FIELD
//  9. 字段语义 → INVALID_FIELD / MISSING_FIELD
//  10. 跨字段一致性 → KEY_MISMATCH 等
//  11. 算法相关 → UNSUPPORTED_ALGORITHM / KEY_SIZE_MISMATCH
func (s *Server) handleTool(c *gin.Context) {
	start := time.Now()
	slug := c.Param("slug")
	httpx.SetSlug(c, slug)

	// 1. 工具存在性与启用状态。
	tool, ok := s.catalog.Tool(slug)
	if !ok {
		s.fail(c, start, toolerr.Newf(toolerr.CodeToolNotFound, map[string]any{"slug": slug}))
		return
	}
	if !tool.Enabled {
		s.fail(c, start, toolerr.Newf(toolerr.CodeToolDisabled, map[string]any{"slug": slug}))
		return
	}

	// 2. 本版是否已实现该工具。
	def, ok := s.tools.Lookup(slug)
	if !ok {
		s.fail(c, start, toolerr.Newf(toolerr.CodeNotImplemented, map[string]any{"slug": slug}))
		return
	}

	// 3. HTTP 方法。
	if c.Request.Method != http.MethodPost {
		s.fail(c, start, toolerr.Newf(toolerr.CodeMethodNotAllowed, map[string]any{
			"method":  c.Request.Method,
			"allowed": http.MethodPost,
		}))
		return
	}

	// 4. Content-Type。
	if !hasJSONContentType(c) {
		s.fail(c, start, toolerr.Newf(toolerr.CodeUnsupportedMediaType, map[string]any{
			"contentType": c.GetHeader("Content-Type"),
			"supported":   []string{"application/json"},
		}))
		return
	}

	// 7. 请求体上限。
	httpx.LimitBody(c, tool.MaxBodyBytes)

	// 8. 语法解码。
	req := def.NewRequest()
	if err := tools.DecodeJSON(c.Request, req); err != nil {
		s.fail(c, start, tools.Normalize(err))
		return
	}

	// 9 + 10. 字段语义与跨字段一致性。
	if err := tools.Validate(req); err != nil {
		s.fail(c, start, tools.Normalize(err))
		return
	}

	// 11. 执行工具，施加工具级超时。
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(tool.TimeoutMs)*time.Millisecond)
	defer cancel()

	result, err := def.Handle(ctx, req)
	if err != nil {
		s.fail(c, start, tools.Normalize(err))
		return
	}

	meta := httpx.MetaOf(c)
	meta.DurationMs = time.Since(start).Milliseconds()
	meta.InputBytes = c.Request.ContentLength
	meta.OutputBytes = int64(c.Writer.Size())
	if len(result.Warnings) > 0 {
		meta.Warnings = result.Warnings
	}
	httpx.SetMeta(c, meta)

	s.record(c, slug, http.StatusOK, "", time.Since(start))
	httpx.Success(c, result.Data, meta)
}

// fail 写出失败信封并记录统计事件。
func (s *Server) fail(c *gin.Context, start time.Time, err *toolerr.Error) {
	httpx.SetErrorCode(c, string(err.Code))
	meta := httpx.MetaOf(c)
	meta.DurationMs = time.Since(start).Milliseconds()
	meta.InputBytes = c.Request.ContentLength
	httpx.SetMeta(c, meta)

	s.record(c, c.Param("slug"), err.HTTPStatus(), string(err.Code), time.Since(start))
	httpx.Failure(c, err, meta)
}

// record 记录调用事件。
func (s *Server) record(c *gin.Context, slug string, status int, errCode string, duration time.Duration) {
	if s.writer == nil || !s.cfg.StatsEnabled {
		return
	}
	s.writer.Write(statsEvent(slug, s.categoryOf(slug), status, errCode, duration, c))
}

// hasJSONContentType 判断 Content-Type 是否为 JSON。
//
// 允许 application/json 及其带 charset 参数的变体；
// 空的 Content-Type 视为不支持（客户端必须显式声明）。
func hasJSONContentType(c *gin.Context) bool {
	ct := c.GetHeader("Content-Type")
	if ct == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return strings.EqualFold(mediaType, "application/json")
}

// statsEvent 构造一条统计事件。
func statsEvent(slug, category string, status int, errCode string, duration time.Duration, c *gin.Context) stats.Event {
	return stats.Event{
		Slug:       slug,
		Category:   category,
		Status:     status,
		OK:         status < http.StatusBadRequest,
		ErrorCode:  errCode,
		DurationMs: duration.Milliseconds(),
		ReqBytes:   c.Request.ContentLength,
		RespBytes:  int64(c.Writer.Size()),
	}
}
