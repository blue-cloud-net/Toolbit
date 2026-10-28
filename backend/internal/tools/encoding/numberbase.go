package encoding

import (
	"context"
	"math/big"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 进制取值范围。
const (
	// MinBase 是支持的最小进制（二进制）。
	MinBase = 2
	// MaxBase 是支持的最大进制（三十六进制）。
	MaxBase = 36
)

// NumberBaseRequest 是 number-base 工具的请求。
type NumberBaseRequest struct {
	// Value 是待转换的值（字符串，支持大整数）。
	Value string `json:"value"`
	// From 是源进制（2–36）。
	From int `json:"from"`
	// To 是目标进制（2–36）。
	To int `json:"to"`
	// Prefix 是否加进制前缀（0x / 0o / 0b）。
	Prefix bool `json:"prefix"`
}

// NumberBaseResult 是 number-base 工具的响应。
type NumberBaseResult struct {
	// Result 是转换结果。
	Result string `json:"result"`
	// From 是源进制。
	From int `json:"from"`
	// To 是目标进制。
	To int `json:"to"`
}

// Validate 做字段级校验。
func (r *NumberBaseRequest) Validate() error {
	if strings.TrimSpace(r.Value) == "" {
		return tools.MissingField("value")
	}
	if r.From < MinBase || r.From > MaxBase {
		return tools.InvalidField("from", "2..36", r.From)
	}
	if r.To < MinBase || r.To > MaxBase {
		return tools.InvalidField("to", "2..36", r.To)
	}
	return nil
}

// NumberBase 执行进制转换。
//
// 使用任意精度整数，**不**受 64 位限制。
func NumberBase(_ context.Context, req *NumberBaseRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	text := strings.TrimSpace(req.Value)
	negative := strings.HasPrefix(text, "-")
	if negative {
		text = text[1:]
	}

	value, ok := new(big.Int).SetString(text, req.From)
	if !ok {
		return tools.Result{}, tools.InvalidField("value", "可被 "+itoa(req.From)+" 进制解析的整数", req.Value)
	}

	out := value.Text(req.To)
	if negative {
		out = "-" + out
	}
	if req.Prefix {
		out = withPrefix(out, req.To)
	}

	return tools.Result{Data: NumberBaseResult{
		Result: out,
		From:   req.From,
		To:     req.To,
	}}, nil
}

// withPrefix 按进制加前缀。
//
// 只有 2、8、16 进制有惯用前缀，其余进制不加。
func withPrefix(s string, base int) string {
	switch base {
	case 2:
		return "0b" + s
	case 8:
		return "0o" + s
	case 16:
		return "0x" + s
	default:
		return s
	}
}

// itoa 把进制转成字符串。
func itoa(base int) string {
	if base < 10 {
		return string(rune('0' + base))
	}
	return string(rune('a' + base - 10))
}
