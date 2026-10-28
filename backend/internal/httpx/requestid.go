package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// RequestIDHeader 是请求标识的请求头与响应头名。
const RequestIDHeader = "X-Request-Id"

// requestIDKey 是上下文中的键名。
const requestIDKey = "httpx.requestId"

// metaKey 是上下文中元信息的键名。
const metaKey = "httpx.meta"

// requestIDPattern 是允许的入站 requestId 格式：仅字母、数字、下划线与连字符。
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// RequestID 生成并透传请求标识。
//
// **必须**在**最外层**中间件使用：每个请求生成一个 requestId，
// 写入上下文、随日志与响应返回。
//
// 优先采用入站 X-Request-Id，但**必须**校验格式；
// 格式非法时重新生成，不透传。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}
		c.Set(requestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// RequestIDFrom 取出当前请求的 requestId。
func RequestIDFrom(c *gin.Context) string {
	if v, ok := c.Get(requestIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// newRequestID 生成 16 字节随机十六进制字符串。
func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// 随机源不可用时退化为空标识：宁可标识为空，也不要把错误抛到业务层。
		return ""
	}
	return hex.EncodeToString(buf)
}

// StatusRecorder 记录响应状态码与字节数。
type StatusRecorder struct {
	gin.ResponseWriter
	status int
	size   int
}

// WriteHeader 记录状态码。
func (r *StatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Write 记录字节数。
func (r *StatusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// WriteString 记录字节数。
func (r *StatusRecorder) WriteString(s string) (int, error) {
	n, err := r.ResponseWriter.WriteString(s)
	r.size += n
	return n, err
}

// Status 返回响应状态码，遮蔽内嵌 ResponseWriter 的同名方法。
func (r *StatusRecorder) Status() int {
	return r.status
}

// Size 返回已写出的字节数，遮蔽内嵌 ResponseWriter 的同名方法。
func (r *StatusRecorder) Size() int {
	return r.size
}

// StatusCode 返回记录到的状态码。
func (r *StatusRecorder) StatusCode() int {
	return r.status
}

// BytesWritten 返回已写出的字节数。
func (r *StatusRecorder) BytesWritten() int {
	return r.size
}

// Recorder 返回包裹了状态记录的 ResponseWriter。
func Recorder(c *gin.Context) *StatusRecorder {
	if rec, ok := c.Writer.(*StatusRecorder); ok {
		return rec
	}
	rec := &StatusRecorder{ResponseWriter: c.Writer, status: http.StatusOK}
	c.Writer = rec
	return rec
}
