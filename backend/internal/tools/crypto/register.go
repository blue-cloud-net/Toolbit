package crypto

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回加密类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "hash-calc",
			NewRequest: func() tools.Request {
				return &HashCalcRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return HashCalc(ctx, req.(*HashCalcRequest))
			},
		},
		{
			Slug: "hmac-gen",
			NewRequest: func() tools.Request {
				return &HmacGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return HmacGen(ctx, req.(*HmacGenRequest))
			},
		},
		{
			Slug: "password-gen",
			NewRequest: func() tools.Request {
				return &PasswordGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return PasswordGen(ctx, req.(*PasswordGenRequest))
			},
		},
		{
			Slug: "symmetric-crypt",
			NewRequest: func() tools.Request {
				return &SymmetricCryptRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return SymmetricCrypt(ctx, req.(*SymmetricCryptRequest))
			},
		},
		{
			Slug: "signature",
			NewRequest: func() tools.Request {
				return &SignatureRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return Signature(ctx, req.(*SignatureRequest))
			},
		},
		{
			Slug: "asymmetric-crypt",
			NewRequest: func() tools.Request {
				return &AsymmetricCryptRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return AsymmetricCrypt(ctx, req.(*AsymmetricCryptRequest))
			},
		},
		{
			Slug: "ecdh",
			NewRequest: func() tools.Request {
				return &ECDHRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return ECDH(ctx, req.(*ECDHRequest))
			},
		},
	}
}
