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
		{
			Slug: "nginx-config",
			NewRequest: func() tools.Request {
				return &NginxConfigRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return NginxConfig(ctx, req.(*NginxConfigRequest))
			},
		},
		{
			Slug: "firewall-rules",
			NewRequest: func() tools.Request {
				return &FirewallRulesRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return FirewallRules(ctx, req.(*FirewallRulesRequest))
			},
		},
		{
			Slug: "docker-compose-gen",
			NewRequest: func() tools.Request {
				return &DockerComposeRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return DockerComposeGen(ctx, req.(*DockerComposeRequest))
			},
		},
		{
			Slug: "lookup-table",
			NewRequest: func() tools.Request {
				return &LookupTableRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return LookupTable(ctx, req.(*LookupTableRequest))
			},
		},
	}
}
