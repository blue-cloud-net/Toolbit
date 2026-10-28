// Package tools 定义工具的抽象与注册表。
//
// 每个工具 handler 的职责**只有三件事**：
//
//  1. 解码输入（JSON 结构体 / multipart 表单）并做**字段级**校验；
//  2. 调用 cryptosvc / codecsvc / netguard 的服务；
//  3. 编码响应（DTO）。
//
// handler **不做**算法实现，**不**直接接触 Tongsuo，**不**自己拼错误码。
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// Request 是工具请求的最小接口。
//
// 请求结构体**必须**是具名类型，字段带 json tag，并实现 Validate。
type Request interface {
	// Validate 做字段级校验，返回 *toolerr.Error。
	Validate() error
}

// Result 是工具的结果。
//
// 结果**必须**是具名结构体，字段带 json tag 与注释；
// **禁止**用 map[string]any 作为响应 data。
type Result struct {
	// Data 是响应数据，序列化后即 data 字段。
	Data any
	// Warnings 是警告列表，可为空。
	Warnings []toolerr.Warning
}

// Handler 是工具处理函数。
type Handler func(ctx context.Context, req Request) (Result, error)

// MultipartRequest 是 multipart 工具的请求。
//
// 与 Request 的区别：输入来自 multipart/form-data，
// 且可能携带**大文件**——因此表单是**流式**的。
type MultipartRequest interface {
	// BindMultipart 绑定 multipart 表单。
	//
	// 表单按 part 顺序到达。实现**必须**先把所有小字段读完，
	// 再处理文件部分：一旦开始读文件，就无法回头取后面的字段。
	//
	// 因此**文件部分必须是最后一个 part**。
	// 实现**不得**把文件内容读入内存，也**不得**落盘。
	//
	// 返回 *toolerr.Error 表示表单非法。
	BindMultipart(r *multipart.Reader) error

	// Validate 做字段级校验，返回 *toolerr.Error。
	Validate() error
}

// MultipartResult 是 multipart 工具的响应。
//
// 与 Result 的区别：响应体是**二进制流**，**不套**统一信封。
type MultipartResult struct {
	// ContentType 是响应的 Content-Type；为空时用 application/octet-stream。
	ContentType string
	// Filename 是建议的下载文件名；为空时不带 Content-Disposition。
	Filename string
	// Body 是响应体。
	//
	// 分发器负责关闭它。
	Body io.ReadCloser
	// Warnings 是警告列表，可为空。
	Warnings []toolerr.Warning
}

// MultipartHandler 处理 multipart 请求。
type MultipartHandler func(ctx context.Context, req MultipartRequest) (MultipartResult, error)

// Definition 描述一个工具的实现。
type Definition struct {
	// Slug 是工具标识，必须与 catalog 中的 slug 一致。
	Slug string
	// NewRequest 返回一个空的请求结构体，供分发器解码。
	NewRequest func() Request
	// Handle 执行工具。
	Handle Handler

	// NewMultipart 返回一个空的 multipart 请求结构体。
	//
	// 仅 `transport: multipart` 的工具需要；为 nil 表示该工具不支持 multipart。
	NewMultipart func() MultipartRequest
	// HandleMultipart 执行 multipart 工具。
	//
	// 仅 `transport: multipart` 的工具需要；为 nil 表示该工具不支持 multipart。
	HandleMultipart MultipartHandler
}

// SupportsMultipart 报告该工具是否支持 multipart。
func (d Definition) SupportsMultipart() bool {
	return d.NewMultipart != nil && d.HandleMultipart != nil
}

// DecodeJSON 严格解码 JSON 请求体到 req。
//
// 严格模式：**拒绝**未知字段。
// 语法错误 → INVALID_JSON；字段类型错误 → INVALID_FIELD；未知字段 → INVALID_FIELD。
func DecodeJSON(r *http.Request, req Request) error {
	if r.Body == nil {
		return toolerr.New(toolerr.CodeInvalidRequest)
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(req); err != nil {
		return decodeError(err)
	}
	// 拒绝尾随内容：请求体只允许一个 JSON 值。
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return toolerr.Newf(toolerr.CodeInvalidJSON, map[string]any{
			"reason": "trailing content",
		})
	}
	return nil
}

// decodeError 把 JSON 解码错误映射为工具错误码。
func decodeError(err error) *toolerr.Error {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return toolerr.Newf(toolerr.CodeInvalidJSON, map[string]any{
			"offset": syntaxErr.Offset,
		})
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    typeErr.Field,
			"expected": typeErr.Type.String(),
		})
	}
	if errors.Is(err, io.EOF) {
		return toolerr.New(toolerr.CodeInvalidRequest)
	}
	// 截断的 JSON 表现为 ErrUnexpectedEOF，同样归为语法错误。
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return toolerr.Newf(toolerr.CodeInvalidJSON, map[string]any{
			"reason": "unexpected end of input",
		})
	}
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"maxBytes": maxErr.Limit,
		})
	}
	// 未知字段等情况：统一归为字段非法。
	return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
		"reason": "unknown field or malformed body",
	})
}

// Validate 调用请求的字段级校验，并把普通错误归一化为工具错误。
func Validate(req Request) error {
	if err := req.Validate(); err != nil {
		return Normalize(err)
	}
	return nil
}

// Normalize 把任意错误归一化为 *toolerr.Error。
func Normalize(err error) *toolerr.Error {
	if err == nil {
		return nil
	}
	var toolErr *toolerr.Error
	if errors.As(err, &toolErr) {
		return toolErr
	}
	return toolerr.Wrap(toolerr.CodeInternalError, err)
}

// MissingField 构造缺少必填字段的错误。
func MissingField(field string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": field})
}

// InvalidField 构造字段取值非法的错误。
func InvalidField(field string, expected, actual any) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
		"field":    field,
		"expected": expected,
		"actual":   actual,
	})
}

// UnsupportedAlgorithm 构造算法不支持的错误。
func UnsupportedAlgorithm(algorithm string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedAlgorithm, map[string]any{
		"algorithm": algorithm,
		"supported": supported,
	})
}

// UnsupportedFormat 构造格式不支持的错误。
func UnsupportedFormat(format string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedFormat, map[string]any{
		"format":    format,
		"supported": supported,
	})
}

// UnsupportedMode 构造模式不支持的错误。
func UnsupportedMode(mode string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedMode, map[string]any{
		"mode":      mode,
		"supported": supported,
	})
}

// UnsupportedPadding 构造填充方式不支持的错误。
func UnsupportedPadding(padding string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedPadding, map[string]any{
		"padding":   padding,
		"supported": supported,
	})
}

// String 返回用于日志的描述，不含用户输入。
func (d Definition) String() string {
	return fmt.Sprintf("tool(%s)", d.Slug)
}
