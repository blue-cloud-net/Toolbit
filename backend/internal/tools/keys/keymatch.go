package keys

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// KeyMatchRequest 是密钥配对检测工具的请求。
type KeyMatchRequest struct {
	// A 是第一段密钥材料（私钥 / 公钥 / 证书）。
	A KeyInput `json:"a"`
	// B 是第二段密钥材料。
	B KeyInput `json:"b"`
}

// KeyMatchResult 是密钥配对检测工具的响应。
type KeyMatchResult struct {
	// Match 报告两段材料是否配对。
	Match bool `json:"match"`
	// Reason 是结论枚举，取值见 pkisvc.MatchReasons。
	Reason string `json:"reason"`
	// AFingerprint 是 A 的公钥指纹。
	AFingerprint string `json:"aFingerprint"`
	// BFingerprint 是 B 的公钥指纹。
	BFingerprint string `json:"bFingerprint"`
}

// Validate 做字段级校验。
func (r *KeyMatchRequest) Validate() error {
	if err := r.A.Validate("a"); err != nil {
		return err
	}
	if err := r.B.Validate("b"); err != nil {
		return err
	}
	return nil
}

// KeyMatch 检测两段密钥材料是否配对。
//
// 「不配对」是**业务结果**（返回 200 + match:false），
// 只有输入无法解析时才报错。
func KeyMatch(_ context.Context, req *KeyMatchRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	a, _, err := decodeInput(req.A)
	if err != nil {
		return tools.Result{}, err
	}
	b, _, err := decodeInput(req.B)
	if err != nil {
		return tools.Result{}, err
	}

	matched, err := pkisvc.Match(a, b)
	if err != nil {
		return tools.Result{}, err
	}

	result := &KeyMatchResult{
		Match:        matched.Match,
		Reason:       matched.Reason,
		AFingerprint: matched.AFingerprint,
		BFingerprint: matched.BFingerprint,
	}
	return tools.Result{Data: result}, nil
}
