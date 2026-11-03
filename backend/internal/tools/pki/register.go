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
		{
			Slug: "csr-gen",
			NewRequest: func() tools.Request {
				return &CSRGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CSRGen(ctx, req.(*CSRGenRequest))
			},
		},
		{
			Slug: "cert-gen",
			NewRequest: func() tools.Request {
				return &CertGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CertGen(ctx, req.(*CertGenRequest))
			},
		},
		{
			Slug: "crl-gen",
			NewRequest: func() tools.Request {
				return &CRLGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CRLGen(ctx, req.(*CRLGenRequest))
			},
		},
		{
			Slug: "revocation-check",
			NewRequest: func() tools.Request {
				return &RevocationCheckRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return RevocationCheck(ctx, req.(*RevocationCheckRequest))
			},
		},
		{
			Slug: "ocsp-request",
			NewRequest: func() tools.Request {
				return &OCSPRequestInput{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return OCSPRequest(ctx, req.(*OCSPRequestInput))
			},
		},
		{
			Slug: "ocsp-response",
			NewRequest: func() tools.Request {
				return &OCSPResponseInput{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return OCSPResponse(ctx, req.(*OCSPResponseInput))
			},
		},
		{
			Slug: "pfx-pack",
			NewRequest: func() tools.Request {
				return &PFXPackRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return PFXPack(ctx, req.(*PFXPackRequest))
			},
		},
		{
			Slug: "pfx-parse",
			NewRequest: func() tools.Request {
				return &PFXParseRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return PFXParse(ctx, req.(*PFXParseRequest))
			},
		},
		{
			Slug: "pfx-change-password",
			NewRequest: func() tools.Request {
				return &PFXChangePasswordRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return PFXChangePassword(ctx, req.(*PFXChangePasswordRequest))
			},
		},
		{
			Slug: "p7b-pack",
			NewRequest: func() tools.Request {
				return &P7BPackRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return P7BPack(ctx, req.(*P7BPackRequest))
			},
		},
		{
			Slug: "p7b-extract",
			NewRequest: func() tools.Request {
				return &P7BExtractRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return P7BExtract(ctx, req.(*P7BExtractRequest))
			},
		},
		{
			Slug: "caa-gen",
			NewRequest: func() tools.Request {
				return &CAAGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CAAGen(ctx, req.(*CAAGenRequest))
			},
		},
	}
}
