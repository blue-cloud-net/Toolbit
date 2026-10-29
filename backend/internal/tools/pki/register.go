package pki

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回证书类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "cert-parse",
			NewRequest: func() tools.Request {
				return &CertParseRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CertParse(ctx, req.(*CertParseRequest))
			},
		},
		{
			Slug: "csr-parse",
			NewRequest: func() tools.Request {
				return &CSRParseRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CSRParse(ctx, req.(*CSRParseRequest))
			},
		},
		{
			Slug: "pem-der",
			NewRequest: func() tools.Request {
				return &PEMDerRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return PEMDer(ctx, req.(*PEMDerRequest))
			},
		},
		{
			Slug: "asn1-view",
			NewRequest: func() tools.Request {
				return &ASN1ViewRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return ASN1View(ctx, req.(*ASN1ViewRequest))
			},
		},
		{
			Slug: "cert-chain-parse",
			NewRequest: func() tools.Request {
				return &ChainParseRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return ChainParse(ctx, req.(*ChainParseRequest))
			},
		},
		{
			Slug: "cert-chain-verify",
			NewRequest: func() tools.Request {
				return &ChainVerifyRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return ChainVerify(ctx, req.(*ChainVerifyRequest))
			},
		},
		{
			Slug: "cert-chain-repair",
			NewRequest: func() tools.Request {
				return &ChainRepairRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return ChainRepair(ctx, req.(*ChainRepairRequest))
			},
		},
		{
			Slug: "crl-parse",
			NewRequest: func() tools.Request {
				return &CRLParseRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CRLParse(ctx, req.(*CRLParseRequest))
			},
		},
	}
}
