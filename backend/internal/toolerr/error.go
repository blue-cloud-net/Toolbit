package toolerr

import (
	"errors"
	"fmt"
	"net/http"
)

// Error 是 Toolbit 的统一错误类型。
//
// 它同时承载三层语义：HTTP 状态（传输层）、Code（语义层）与 Params（可插值参数）。
//
// **必须**通过 New 构造，不要手工拼接错误码字符串。
type Error struct {
	// Code 是机器可读的错误码。
	Code Code
	// Params 是可插值参数，只放结构化元信息（字段名、期望长度、索引、上限）。
	// **禁止**放入用户输入原文。
	Params map[string]any
	// cause 是底层错误，仅进日志，**不**回显给客户端。
	cause error
}

// New 用给定错误码构造一个 Error。
func New(code Code) *Error {
	return &Error{Code: code}
}

// Newf 用给定错误码构造一个 Error，并附加可插值参数。
func Newf(code Code, params map[string]any) *Error {
	return &Error{Code: code, Params: params}
}

// Wrap 用给定错误码包装一个底层错误。
//
// 底层错误文本**不**进入对外 message，只保留在 cause 中供日志排查。
func Wrap(code Code, cause error) *Error {
	return &Error{Code: code, cause: cause}
}

// Error 实现 error 接口，返回对外的英文兜底文案。
//
// 该文案来自注册表，**不含**用户输入。
func (e *Error) Error() string {
	spec, ok := Registry[e.Code]
	if !ok {
		return "unknown error"
	}
	return spec.Message
}

// HTTPStatus 返回该错误码对应的 HTTP 状态码。
//
// 未知错误码一律返回 500，避免把未登记的错误码泄露到传输层。
func (e *Error) HTTPStatus() int {
	spec, ok := Registry[e.Code]
	if !ok {
		return http.StatusInternalServerError
	}
	return spec.HTTP
}

// I18nKey 返回该错误码的前端本地化词条 key。
func (e *Error) I18nKey() string {
	spec, ok := Registry[e.Code]
	if !ok {
		return "errors.INTERNAL_ERROR"
	}
	return spec.I18nKey
}

// Unwrap 返回底层错误，供 errors.Is / errors.As 使用。
func (e *Error) Unwrap() error {
	return e.cause
}

// WithParam 追加一个可插值参数并返回自身，便于链式调用。
func (e *Error) WithParam(key string, value any) *Error {
	if e.Params == nil {
		e.Params = make(map[string]any, 1)
	}
	e.Params[key] = value
	return e
}

// From 把任意 error 归一化为 *Error。
//
// 已经是 *Error 的原样返回；否则按内部错误处理。
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return Wrap(CodeInternalError, err)
}

// Message 返回该错误码的对外兜底文案。
func Message(code Code) string {
	spec, ok := Registry[code]
	if !ok {
		return "unknown error"
	}
	return spec.Message
}

// HTTPStatusOf 返回错误码对应的 HTTP 状态码，未登记时返回 500。
func HTTPStatusOf(code Code) int {
	spec, ok := Registry[code]
	if !ok {
		return http.StatusInternalServerError
	}
	return spec.HTTP
}

// I18nKeyOf 返回错误码的本地化 key，未登记时回退到内部错误。
func I18nKeyOf(code Code) string {
	spec, ok := Registry[code]
	if !ok {
		return "errors.INTERNAL_ERROR"
	}
	return spec.I18nKey
}

// String 返回用于日志的简短描述，含错误码而不含用户输入。
func (e *Error) String() string {
	return fmt.Sprintf("%s(%d)", e.Code, e.HTTPStatus())
}
