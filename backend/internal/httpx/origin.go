package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// OriginCheck 校验跨站请求。
//
// 规则：
//   - Sec-Fetch-Site: cross-site 一律拒绝；
//   - 携带 Origin 时，Origin 的 host 必须与请求的 Host 一致；
//   - 无 Origin 且无 Sec-Fetch-Site 的请求放行（非浏览器客户端）。
//
// 被拒 → 403 + CROSS_ORIGIN_BLOCKED。
//
// 部署在反向代理后时，代理**必须**覆写 Origin / Sec-Fetch-Site，
// **禁止**透传客户端伪造值。
func OriginCheck(allowedOrigins []string) gin.HandlerFunc {
	allowlist := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowlist[strings.ToLower(strings.TrimSpace(o))] = true
	}

	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		if site := c.GetHeader("Sec-Fetch-Site"); site != "" {
			if strings.EqualFold(site, "cross-site") {
				abortCrossOrigin(c, c.GetHeader("Origin"))
				return
			}
		}

		origin := c.GetHeader("Origin")
		if origin == "" || origin == "null" {
			c.Next()
			return
		}

		if len(allowlist) > 0 && allowlist[strings.ToLower(origin)] {
			c.Next()
			return
		}

		if !sameHost(origin, c.Request.Host) {
			abortCrossOrigin(c, origin)
			return
		}

		c.Next()
	}
}

// sameHost 判断 Origin 的 host 是否与请求的 Host 相同。
func sameHost(origin, host string) bool {
	origin = strings.TrimPrefix(origin, "http://")
	origin = strings.TrimPrefix(origin, "https://")
	// 去掉可能的路径部分（Origin 规范上不应带路径，但防御性处理）。
	if idx := strings.IndexByte(origin, '/'); idx >= 0 {
		origin = origin[:idx]
	}
	return strings.EqualFold(origin, host)
}

// abortCrossOrigin 拒绝跨站请求。
func abortCrossOrigin(c *gin.Context, origin string) {
	err := toolerr.Newf(toolerr.CodeCrossOriginBlocked, map[string]any{"origin": origin})
	SetErrorCode(c, string(toolerr.CodeCrossOriginBlocked))
	Failure(c, err, NewMeta())
	c.Abort()
}
