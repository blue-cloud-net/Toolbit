package text

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回文本类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "regex-tester",
			NewRequest: func() tools.Request {
				return &RegexTesterRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				request := req.(*RegexTesterRequest)
				result, err := RegexTester(request)
				if err != nil {
					return tools.Result{}, err
				}
				warnings := []toolerr.Warning{}
				if hasDialectDifference(request.Pattern) {
					warnings = append(warnings, toolerr.NewWarning(
						toolerr.WarnRegexDialectDifference, map[string]any{
							"dialect": "RE2",
						}))
				}
				return tools.Result{Data: result, Warnings: warnings}, nil
			},
		},
		{
			Slug: "text-convert",
			NewRequest: func() tools.Request {
				return &TextConvertRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := TextConvert(req.(*TextConvertRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
	}
}
