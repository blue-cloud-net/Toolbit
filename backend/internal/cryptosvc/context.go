package cryptosvc

import "context"

// sessionKey 是 Session 在 context 中的键类型。
//
// 使用**未导出的空结构体**作为键，避免与其它包的 context 值冲突。
type sessionKey struct{}

// WithSession 把 Session 放入上下文。
func WithSession(ctx context.Context, s *Session) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil {
		return ctx
	}
	return context.WithValue(ctx, sessionKey{}, s)
}

// FromContext 从上下文取出 Session。
//
// 上下文里没有 Session 时返回 nil；调用方**必须**判空，
// 不得假设 Session 一定存在（例如直接调用 handler 的单元测试）。
func FromContext(ctx context.Context) *Session {
	if ctx == nil {
		return nil
	}
	s, _ := ctx.Value(sessionKey{}).(*Session)
	return s
}
