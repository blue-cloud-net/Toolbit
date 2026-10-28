package api

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// handleMultipartTool 执行 multipart 工具。
//
// 与 JSON 通道的差别：
//   - 请求体是 multipart/form-data；
//   - 响应体是**二进制流**，**不套**统一信封；
//   - 警告通过 `X-Warnings` 响应头传递（流式响应没有信封可挂）。
//
// 校验顺序与 JSON 通道一致（见 handleTool 的注释），
// 只是第 4 步按 transport 分流。
func (s *Server) handleMultipartTool(
	c *gin.Context,
	start time.Time,
	tool catalog.Tool,
	def tools.Definition,
) {
	// 4. Content-Type 必须是 multipart/form-data。
	if !hasMultipartContentType(c) {
		s.fail(c, start, toolerr.Newf(toolerr.CodeUnsupportedMediaType, map[string]any{
			"contentType": c.GetHeader("Content-Type"),
			"supported":   []string{"multipart/form-data"},
		}))
		return
	}

	// 工具实现与 catalog 的 transport 必须一致。
	if !def.SupportsMultipart() {
		s.fail(c, start, toolerr.Newf(toolerr.CodeNotImplemented, map[string]any{
			"slug": tool.Slug,
		}))
		return
	}

	// 7. 请求体上限。
	httpx.LimitBody(c, tool.MaxBodyBytes)

	// 8. 表单解码（**流式**，不落盘）。
	reader, err := tools.NewMultipartReader(c.Request)
	if err != nil {
		s.fail(c, start, tools.Normalize(err))
		return
	}

	req := def.NewMultipart()
	if err := req.BindMultipart(reader); err != nil {
		s.fail(c, start, tools.Normalize(err))
		return
	}

	// 9 + 10. 字段语义与跨字段一致性。
	if err := req.Validate(); err != nil {
		s.fail(c, start, tools.Normalize(err))
		return
	}

	// 11. 执行工具，施加工具级超时。
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(tool.TimeoutMs)*time.Millisecond)
	defer cancel()

	sess := cryptosvc.NewSession(ctx)
	defer sess.Close()

	result, err := def.HandleMultipart(sess.Context(), req)

	s.logOrphans(tool.Slug, sess)
	if closeErr := sess.Close(); closeErr != nil {
		s.logger.Warn("密码学句柄释放失败", "slug", tool.Slug, "error", closeErr.Error())
	}

	if err != nil {
		s.failWith(c, start, tools.Normalize(err), sess.Warnings())
		return
	}
	if result.Body == nil {
		s.failWith(c, start, toolerr.New(toolerr.CodeInternalError), sess.Warnings())
		return
	}
	defer func() { _ = result.Body.Close() }()

	warnings := mergeWarnings(result.Warnings, sess.Warnings())

	// 写二进制响应。
	contentType := result.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-store")
	if result.Filename != "" {
		if disposition := tools.ContentDisposition(result.Filename); disposition != "" {
			c.Header("Content-Disposition", disposition)
		}
	}
	if len(warnings) > 0 {
		if raw, err := json.Marshal(warnings); err == nil {
			c.Header(httpx.HeaderWarnings, string(raw))
		}
	}

	c.Status(http.StatusOK)
	written, copyErr := io.Copy(c.Writer, result.Body)

	meta := httpx.MetaOf(c)
	meta.DurationMs = time.Since(start).Milliseconds()
	meta.InputBytes = c.Request.ContentLength
	meta.OutputBytes = written
	meta.Warnings = warnings
	httpx.SetMeta(c, meta)

	if copyErr != nil {
		// 响应头已发出，**无法**再改状态码。
		// 只能记录日志与统计：客户端会看到一份被截断的下载。
		s.logger.Warn("流式响应中断",
			"slug", tool.Slug,
			"error", copyErr.Error(),
			"writtenBytes", written,
		)
		httpx.SetErrorCode(c, string(toolerr.CodeDecryptFailed))
		s.record(c, tool.Slug, http.StatusOK, string(toolerr.CodeDecryptFailed), time.Since(start))
		return
	}

	s.record(c, tool.Slug, http.StatusOK, "", time.Since(start))
}

// hasMultipartContentType 判断 Content-Type 是否为 multipart/form-data。
func hasMultipartContentType(c *gin.Context) bool {
	ct := c.GetHeader("Content-Type")
	if ct == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return strings.EqualFold(mediaType, "multipart/form-data")
}
