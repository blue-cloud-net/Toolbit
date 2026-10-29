package system

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回系统类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "curl-generator",
			NewRequest: func() tools.Request {
				return &CurlGeneratorRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := CurlGenerator(req.(*CurlGeneratorRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "crontab-generator",
			NewRequest: func() tools.Request {
				return &CrontabGeneratorRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := CrontabGenerator(req.(*CrontabGeneratorRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "chmod-calc",
			NewRequest: func() tools.Request {
				return &ChmodCalcRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := ChmodCalc(req.(*ChmodCalcRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
	}
}
