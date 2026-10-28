package api

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 限流响应头。
const (
	// HeaderRateLimitLimit 是窗口内的限额。
	HeaderRateLimitLimit = "X-RateLimit-Limit"
	// HeaderRateLimitRemaining 是剩余可用次数。
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	// HeaderRateLimitReset 是窗口重置的 Unix 时间戳。
	HeaderRateLimitReset = "X-RateLimit-Reset"
)

// bucketIdleTTL 是空闲桶的回收时长；超过该时长的桶会被清理。
const bucketIdleTTL = 10 * time.Minute

// rateSpec 描述一个限流档位。
type rateSpec struct {
	// limit 是每秒允许的请求数。
	limit rate.Limit
	// burst 是突发容量。
	burst int
}

// rateSpecs 是 catalog 的 rateClass 到限流参数的映射。
//
// 与 docs/api/README.md §5 的表格一致。
var rateSpecs = map[catalog.RateClass]rateSpec{
	catalog.RateLight:    {limit: 30, burst: 60},
	catalog.RateStandard: {limit: 10, burst: 20},
	catalog.RateHeavy:    {limit: 2, burst: 4},
	catalog.RateFile:     {limit: 1, burst: 1},
	catalog.RateNetwork:  {limit: 1, burst: 2},
}

// limiterEntry 是一个限流桶及其最后使用时间。
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiter 按（IP，限流档位）维护令牌桶。
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*limiterEntry
	now     func() time.Time
}

// newRateLimiter 构造限流器。
func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*limiterEntry),
		now:     time.Now,
	}
}

// allow 判断请求是否放行，并返回限额信息。
func (rl *rateLimiter) allow(key string, spec rateSpec) (ok bool, remaining int, reset time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, found := rl.buckets[key]
	if !found {
		entry = &limiterEntry{
			limiter: rate.NewLimiter(spec.limit, spec.burst),
		}
		rl.buckets[key] = entry
	}
	entry.lastSeen = rl.now()

	reservation := entry.limiter.Reserve()
	if !reservation.OK() {
		return false, 0, rl.now()
	}
	delay := reservation.Delay()
	// 令牌桶拒绝时 reservation.Delay 必然大于 0，需要撤销预留。
	if delay > 0 {
		reservation.Cancel()
		return false, 0, rl.now().Add(delay)
	}

	reset = rl.now().Add(time.Duration(float64(time.Second) / float64(spec.limit)))
	remaining = int(entry.limiter.TokensAt(rl.now()))
	return true, remaining, reset
}

// cleanup 回收长时间未使用的桶。
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := rl.now().Add(-bucketIdleTTL)
	for key, entry := range rl.buckets {
		if entry.lastSeen.Before(cutoff) {
			delete(rl.buckets, key)
		}
	}
}

// StartCleanupLoop 启动后台回收协程。
//
// 空闲桶会随 IP 数量增长，**必须**定期回收，否则内存无界增长。
func (rl *rateLimiter) StartCleanupLoop(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(bucketIdleTTL)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.cleanup()
			}
		}
	}()
}

// rateLimit 返回限流中间件。
//
// 限流键为**直连 IP**；未开启 TRUST_PROXY 时**不**信任 X-Forwarded-For。
// 限流**只作用于 /api 分组**：首屏会并发拉取几十个带内容 hash 的 chunk，
// 若全局限流会直接 429。
func (s *Server) rateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" {
			c.Next()
			return
		}
		tool, ok := s.catalog.Tool(slug)
		if !ok {
			// 工具不存在：交给分发器返回 404，不占用限流配额。
			c.Next()
			return
		}
		spec, ok := rateSpecs[tool.RateClass]
		if !ok {
			c.Next()
			return
		}

		key := s.clientIP(c) + "|" + string(tool.RateClass)
		allowed, remaining, reset := s.limiter.allow(key, spec)
		h := c.Writer.Header()
		h.Set(HeaderRateLimitLimit, strconv.Itoa(spec.burst))
		h.Set(HeaderRateLimitRemaining, strconv.Itoa(remaining))
		h.Set(HeaderRateLimitReset, strconv.FormatInt(reset.Unix(), 10))

		if !allowed {
			retryAfter := int(time.Until(reset).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			h.Set("Retry-After", strconv.Itoa(retryAfter))
			err := toolerr.Newf(toolerr.CodeRateLimited, map[string]any{
				"limit":    spec.burst,
				"windowMs": int64(time.Second / time.Duration(spec.limit)),
			})
			httpx.SetErrorCode(c, string(toolerr.CodeRateLimited))
			httpx.Failure(c, err, httpx.MetaOf(c))
			c.Abort()
			return
		}

		c.Next()
	}
}

// clientIP 返回限流所用的客户端 IP。
//
// 默认使用直连 IP；只有在显式信任代理时才读 X-Forwarded-For 的第一段。
func (s *Server) clientIP(c *gin.Context) string {
	if s.cfg.TrustProxy {
		if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
			if idx := indexByte(xff, ','); idx >= 0 {
				return trimASCII(xff[:idx])
			}
			return trimASCII(xff)
		}
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return host
}

// indexByte 返回字节在字符串中的首次下标。
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// methodNotAllowed 构造 METHOD_NOT_ALLOWED 错误。
func methodNotAllowed(method string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeMethodNotAllowed, map[string]any{
		"method":  method,
		"allowed": http.MethodPost,
	})
}
