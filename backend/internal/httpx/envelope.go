// Package httpx 提供统一响应信封与 HTTP 中间件。
//
// 本包**不依赖**任何领域包（catalog / tools / cryptosvc …），
// 可被任意层使用。
//
// 响应信封的字段定义见 docs/api/README.md §2，**必须**逐字对齐。
package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// Envelope 是所有 JSON 响应的统一信封。
type Envelope struct {
	// Code 等于 HTTP 状态码；成功时固定为 0。
	Code int `json:"code"`
	// Message 是英文兜底文案；成功时固定为 "ok"。
	Message string `json:"message"`
	// Data 是业务结果；失败时为 null。
	Data any `json:"data"`
	// Error 是错误详情；成功时为 null。
	Error *ErrorPayload `json:"error"`
	// Meta 是执行元信息。
	Meta Meta `json:"meta"`
	// RequestID 是请求标识，用于日志关联。
	RequestID string `json:"requestId"`
}

// ErrorPayload 是信封中的错误部分。
type ErrorPayload struct {
	// Code 是机器可读错误码。
	Code string `json:"code"`
	// Params 是可插值参数，可缺省。
	Params map[string]any `json:"params,omitempty"`
}

// Meta 是执行元信息。
//
// 字段**只允许新增**，不得改变语义。
type Meta struct {
	// DurationMs 是服务端处理耗时（毫秒）。
	DurationMs int64 `json:"durationMs"`
	// InputBytes 是请求体字节数。
	InputBytes int64 `json:"inputBytes"`
	// OutputBytes 是响应体字节数。
	OutputBytes int64 `json:"outputBytes"`
	// Warnings 是警告列表；无警告时为空数组，**不**为 null。
	Warnings []toolerr.Warning `json:"warnings"`
}

// NewMeta 构造一个带空警告列表的 Meta。
func NewMeta() Meta {
	return Meta{Warnings: []toolerr.Warning{}}
}

// Success 写出成功信封。
//
// 响应固定带 Cache-Control: no-store（调用方已显式设置时不覆盖）。
func Success(c *gin.Context, data any, meta Meta) {
	if meta.Warnings == nil {
		meta.Warnings = []toolerr.Warning{}
	}
	if meta.OutputBytes == 0 {
		meta.OutputBytes = payloadSize(data)
	}
	env := Envelope{
		Code:      0,
		Message:   "ok",
		Data:      data,
		Error:     nil,
		Meta:      meta,
		RequestID: RequestIDFrom(c),
	}
	writeJSON(c, http.StatusOK, env)
}

// payloadSize 返回数据载荷的字节数。
//
// 信封里包含 meta 本身，响应体总长无法在写出前得知；
// 因此 outputBytes 记的是**业务数据**的字节数。
func payloadSize(data any) int64 {
	if data == nil {
		return 0
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	return int64(len(raw))
}

// Failure 写出失败信封。
//
// HTTP 状态取自错误码注册表；错误文案来自注册表，**不含**用户输入。
func Failure(c *gin.Context, err *toolerr.Error, meta Meta) {
	if err == nil {
		err = toolerr.New(toolerr.CodeInternalError)
	}
	if meta.Warnings == nil {
		meta.Warnings = []toolerr.Warning{}
	}
	status := err.HTTPStatus()
	env := Envelope{
		Code:      status,
		Message:   err.Error(),
		Data:      nil,
		Error:     &ErrorPayload{Code: string(err.Code), Params: err.Params},
		Meta:      meta,
		RequestID: RequestIDFrom(c),
	}
	writeJSON(c, status, env)
}

// Fail 是 Failure 的便捷别名。
func Fail(c *gin.Context, err *toolerr.Error) {
	Failure(c, err, MetaOf(c))
}

// MetaOf 从上下文取出已记录的元信息。
func MetaOf(c *gin.Context) Meta {
	if v, ok := c.Get(metaKey); ok {
		if m, ok := v.(Meta); ok {
			return m
		}
	}
	return NewMeta()
}

// SetMeta 把元信息写回上下文。
func SetMeta(c *gin.Context, meta Meta) {
	c.Set(metaKey, meta)
}

// writeJSON 写出 JSON 响应。
//
// 默认带 Cache-Control: no-store；若调用方已显式设置（如目录端点的
// public, max-age=60），则**不**覆盖。
func writeJSON(c *gin.Context, status int, payload any) {
	if c.Writer.Header().Get("Cache-Control") == "" {
		c.Writer.Header().Set("Cache-Control", "no-store")
	}
	c.JSON(status, payload)
}
