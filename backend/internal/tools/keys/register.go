package keys

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回密钥类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "keygen",
			NewRequest: func() tools.Request {
				return &KeyGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return KeyGen(ctx, req.(*KeyGenRequest))
			},
		},
		{
			Slug: "key-convert",
			NewRequest: func() tools.Request {
				return &KeyConvertRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return KeyConvert(ctx, req.(*KeyConvertRequest))
			},
		},
		{
			Slug: "key-parse",
			NewRequest: func() tools.Request {
				return &KeyParseRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return KeyParse(ctx, req.(*KeyParseRequest))
			},
		},
		{
			Slug: "key-ops",
			NewRequest: func() tools.Request {
				return &KeyOpsRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return KeyOps(ctx, req.(*KeyOpsRequest))
			},
		},
		{
			Slug: "key-match",
			NewRequest: func() tools.Request {
				return &KeyMatchRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return KeyMatch(ctx, req.(*KeyMatchRequest))
			},
		},
		{
			Slug: "jwk-xml-convert",
			NewRequest: func() tools.Request {
				return &JWKXMLConvertRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return JWKXMLConvert(ctx, req.(*JWKXMLConvertRequest))
			},
		},
	}
}
