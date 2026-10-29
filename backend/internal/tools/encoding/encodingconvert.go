package encoding

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// EncodingConvertRequest 是 encoding-convert 工具的请求。
type EncodingConvertRequest struct {
	// Input 是待转换内容；encoding 指明 data 的传输编码。
	Input tools.Bytes `json:"input"`
	// From 是源字符集；auto 表示自动探测。
	From string `json:"from"`
	// To 是目标字符集。
	To string `json:"to"`
	// BOM 是否写入字节序标记。
	BOM *bool `json:"bom"`
	// OutputEncoding 是输出编码；默认 base64。
	OutputEncoding string `json:"outputEncoding"`
}

// EncodingConvertResult 是 encoding-convert 工具的响应。
type EncodingConvertResult struct {
	// Output 是转换结果。
	Output string `json:"output"`
	// Encoding 是输出编码。
	Encoding string `json:"encoding"`
	// Detected 是 from 为 auto 时的探测结果。
	Detected string `json:"detected,omitempty"`
}

// EncodingConvert 在字符集之间转换。
//
// 非法字节与不可映射字符**不**导致失败：一律被替换，
// 并在 meta.warnings 中给出 LOSSY_FORMAT 提示。
func EncodingConvert(_ context.Context, req *EncodingConvertRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := req.Input.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	result, err := codecsvc.ConvertCharset(raw, req.From, req.To, req.bom())
	if err != nil {
		return tools.Result{}, err
	}

	enc, err := codecsvc.Normalize(req.outputEncoding())
	if err != nil {
		return tools.Result{}, err
	}
	out, err := codecsvc.Encode(result.Output, enc)
	if err != nil {
		return tools.Result{}, err
	}

	warnings := []toolerr.Warning{}
	if result.Lossy {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnLossyFormat, map[string]any{
			"reason": "some characters could not be represented",
		}))
	}

	return tools.Result{
		Data: EncodingConvertResult{
			Output:   out,
			Encoding: enc,
			Detected: result.Detected,
		},
		Warnings: warnings,
	}, nil
}

// bom 返回是否写入字节序标记；缺省为否。
func (r *EncodingConvertRequest) bom() bool {
	if r.BOM == nil {
		return false
	}
	return *r.BOM
}

// outputEncoding 返回输出编码；缺省为 base64。
func (r *EncodingConvertRequest) outputEncoding() string {
	if r.OutputEncoding == "" {
		return codecsvc.EncodingBase64
	}
	return r.OutputEncoding
}

// Validate 做字段级校验。
func (r *EncodingConvertRequest) Validate() error {
	if r.From == "" {
		return tools.MissingField("from")
	}
	if _, err := codecsvc.NormalizeCharset(r.From); err != nil {
		return err
	}

	if r.To == "" {
		return tools.MissingField("to")
	}
	to, err := codecsvc.NormalizeCharset(r.To)
	if err != nil {
		return err
	}
	if to == codecsvc.CharsetAuto {
		return tools.InvalidField("to", codecsvc.Charsets(), r.To)
	}

	if _, err := codecsvc.Normalize(r.outputEncoding()); err != nil {
		return err
	}

	return r.Input.Validate("input")
}
