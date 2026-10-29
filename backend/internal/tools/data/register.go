package data

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回数据类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "timestamp",
			NewRequest: func() tools.Request {
				return &TimestampRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := Timestamp(ctx, req.(*TimestampRequest))
				if err != nil {
					return tools.Result{}, err
				}
				warnings := []toolerr.Warning{}
				request := req.(*TimestampRequest)
				if request.Timezone == "" {
					warnings = append(warnings, toolerr.NewWarning(
						toolerr.WarnTimezoneAssumed, map[string]any{
							"timezone": "UTC",
						}))
				}
				return tools.Result{Data: result, Warnings: warnings}, nil
			},
		},
		{
			Slug: "uuid-gen",
			NewRequest: func() tools.Request {
				return &UUIDGenRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := UUIDGen(ctx, req.(*UUIDGenRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "date-calc",
			NewRequest: func() tools.Request {
				return &DateCalcRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := DateCalc(ctx, req.(*DateCalcRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "country-code",
			NewRequest: func() tools.Request {
				return &CountryCodeRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := CountryCode(ctx, req.(*CountryCodeRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "cn-regions",
			NewRequest: func() tools.Request {
				return &CNRegionsRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := CNRegions(ctx, req.(*CNRegionsRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "qr-code",
			NewRequest: func() tools.Request {
				return &QRCodeRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := QRCode(ctx, req.(*QRCodeRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "id-validator",
			NewRequest: func() tools.Request {
				return &IDValidateRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return IDValidate(ctx, req.(*IDValidateRequest))
			},
		},
		{
			Slug: "csv-convert",
			NewRequest: func() tools.Request {
				return &CSVConvertRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CSVConvert(ctx, req.(*CSVConvertRequest))
			},
		},
		{
			Slug: "json-to-types",
			NewRequest: func() tools.Request {
				return &JSONToTypesRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return JSONToTypes(ctx, req.(*JSONToTypesRequest))
			},
		},
		{
			Slug: "mock-data",
			NewRequest: func() tools.Request {
				return &MockDataRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return MockData(ctx, req.(*MockDataRequest))
			},
		},
		{
			Slug: "archive-inspect",
			NewRequest: func() tools.Request {
				return &ArchiveInspectRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return ArchiveInspect(ctx, req.(*ArchiveInspectRequest))
			},
		},
		{
			Slug: "file-analyze",
			NewRequest: func() tools.Request {
				return &FileAnalyzeRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return FileAnalyze(ctx, req.(*FileAnalyzeRequest))
			},
		},
		{
			Slug: "wifi-qr",
			NewRequest: func() tools.Request {
				return &WiFiQRRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return WiFiQR(ctx, req.(*WiFiQRRequest))
			},
		},
		{
			Slug: "id-generator",
			NewRequest: func() tools.Request {
				return &IDGeneratorRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return IDGenerator(ctx, req.(*IDGeneratorRequest))
			},
		},
	}
}
