package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// BodyLimit 施加全局请求体上限。
//
// maxBytes 为全局兜底上限（TOOLBIT_MAX_BODY_BYTES）；
// 每个工具还有各自的 maxBodyBytes，由分发器在路由层施加。
//
// 超限 → 413 + TOO_LARGE。
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 {
			c.Next()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// LimitBody 施加工具级的请求体上限。
//
// 必须在解码请求体**之前**调用；超限时解码会返回 *http.MaxBytesError，
// 由分发器统一映射为 TOO_LARGE。
func LimitBody(c *gin.Context, maxBytes int64) {
	if maxBytes <= 0 {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
}

// TooLargeError 构造 TOO_LARGE 错误，附带上限与实际大小。
func TooLargeError(maxBytes, actual int64) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
		"maxBytes": maxBytes,
		"actual":   actual,
	})
}
