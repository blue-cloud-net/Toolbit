package formatting

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回格式化类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "json-format",
			NewRequest: func() tools.Request {
				return &JSONFormatRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*JSONFormatRequest), JSONFormat)
			},
		},
		{
			Slug: "yaml-format",
			NewRequest: func() tools.Request {
				return &YAMLFormatRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*YAMLFormatRequest), YAMLFormat)
			},
		},
		{
			Slug: "toml-format",
			NewRequest: func() tools.Request {
				return &TOMLFormatRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*TOMLFormatRequest), TOMLFormat)
			},
		},
		{
			Slug: "xml-format",
			NewRequest: func() tools.Request {
				return &XMLFormatRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*XMLFormatRequest), XMLFormat)
			},
		},
		{
			Slug: "html-format",
			NewRequest: func() tools.Request {
				return &HTMLFormatRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*HTMLFormatRequest), HTMLFormat)
			},
		},
		{
			Slug: "js-format",
			NewRequest: func() tools.Request {
				return &JSFormatRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*JSFormatRequest), JSFormat)
			},
		},
		{
			Slug: "sql-format",
			NewRequest: func() tools.Request {
				return &SQLFormatRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*SQLFormatRequest), SQLFormat)
			},
		},
		{
			Slug: "format-convert",
			NewRequest: func() tools.Request {
				return &FormatConvertRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return runFormat(ctx, req.(*FormatConvertRequest), FormatConvert)
			},
		},
	}
}

// formatter 是格式化工具的统一签名。
//
// 各格式化器的请求与结果类型各不相同（JSON / YAML / …），
// 因此这里用泛型把「校验 → 执行 → 包信封」这一步统一掉。
type formatter[Req any, Res any] func(*Req) (Res, error)

// runFormat 执行格式化工具。
//
// 语法错误**不是**工具错误：它们被放进响应的 `error` 字段里
// （`valid: false`），状态码仍是 200 —— 因为「这段内容语法不对」
// 本身就是工具给出的**结论**，而不是请求失败。
//
// 只有字段级校验失败（枚举越界、缺字段）才返回错误码。
func runFormat[Req any, Res any](_ context.Context, req *Req, fn formatter[Req, Res]) (tools.Result, error) {
	result, err := fn(req)
	if err != nil {
		return tools.Result{}, err
	}
	return tools.Result{Data: result}, nil
}
