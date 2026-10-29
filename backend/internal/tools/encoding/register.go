package encoding

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回编码类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "base64",
			NewRequest: func() tools.Request {
				return &Base64Request{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return Base64(ctx, req.(*Base64Request))
			},
		},
		{
			Slug: "number-base",
			NewRequest: func() tools.Request {
				return &NumberBaseRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return NumberBase(ctx, req.(*NumberBaseRequest))
			},
		},
		{
			Slug: "url-codec",
			NewRequest: func() tools.Request {
				return &URLCodecRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return URLCodec(ctx, req.(*URLCodecRequest))
			},
		},
		{
			Slug: "jwt-codec",
			NewRequest: func() tools.Request {
				return &JWTCodecRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return JWTCodec(ctx, req.(*JWTCodecRequest))
			},
		},
		{
			Slug: "encoding-convert",
			NewRequest: func() tools.Request {
				return &EncodingConvertRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return EncodingConvert(ctx, req.(*EncodingConvertRequest))
			},
		},
		{
			Slug: "gzip-deflate",
			NewRequest: func() tools.Request {
				return &GZipDeflateRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return GZipDeflate(ctx, req.(*GZipDeflateRequest))
			},
		},
	}
}
