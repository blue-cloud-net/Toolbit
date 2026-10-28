package documents

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回文档类工具的实现列表。
//
// 登记是**显式**的：**禁止**用 init() 副作用注册。
//
// `pdf-sign-verify` **不在**本列表里：它依赖 multipart 上传与
// PDF 签名能力，本版未启用，端点返回 NOT_IMPLEMENTED。
func Definitions() []tools.Definition {
	return []tools.Definition{
		{
			Slug: "markdown-preview",
			NewRequest: func() tools.Request {
				return &MarkdownPreviewRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := MarkdownPreview(ctx, req.(*MarkdownPreviewRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
		{
			Slug: "eml-parser",
			NewRequest: func() tools.Request {
				return &EMLParserRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				result, err := EMLParser(ctx, req.(*EMLParserRequest))
				if err != nil {
					return tools.Result{}, err
				}
				return tools.Result{Data: result}, nil
			},
		},
	}
}
