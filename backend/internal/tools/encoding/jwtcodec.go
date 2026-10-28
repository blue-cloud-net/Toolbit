package encoding

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// JWT 操作。
const (
	// JWTDecode 解析令牌的头部、载荷与签名。
	JWTDecode = "decode"
	// JWTEncode 生成令牌。
	JWTEncode = "encode"
	// JWTVerify 校验 HS 系列令牌的签名。
	JWTVerify = "verify"
)

// JWT 支持的签名算法。
const (
	// JWTHS256 是 HMAC-SHA256。
	JWTHS256 = "HS256"
	// JWTHS384 是 HMAC-SHA384。
	JWTHS384 = "HS384"
	// JWTHS512 是 HMAC-SHA512。
	JWTHS512 = "HS512"
)

// JWTCodecRequest 是 jwt-codec 工具的请求。
type JWTCodecRequest struct {
	// Operation 是 decode、encode 或 verify。
	Operation string `json:"operation"`
	// Token 是令牌文本；decode 与 verify 时必填。
	Token string `json:"token"`
	// Header 是头部；encode 时必填。
	Header map[string]any `json:"header"`
	// Payload 是载荷；encode 时必填。
	Payload map[string]any `json:"payload"`
	// Secret 是 HMAC 密钥；encode 与 verify 时必填。
	Secret string `json:"secret"`
	// Algorithm 是签名算法；默认 HS256。
	Algorithm string `json:"algorithm"`
}

// JWTStatus 是令牌的时间状态。
type JWTStatus struct {
	// Expired 为真表示令牌已过期。
	Expired bool `json:"expired"`
	// NotYetValid 为真表示令牌尚未生效。
	NotYetValid bool `json:"notYetValid"`
}

// JWTCodecResult 是 jwt-codec 工具的响应。
type JWTCodecResult struct {
	// Header 是头部。
	Header map[string]any `json:"header,omitempty"`
	// Payload 是载荷。
	Payload map[string]any `json:"payload,omitempty"`
	// Signature 是签名（base64url，无填充）。
	Signature string `json:"signature,omitempty"`
	// Token 是生成的完整令牌。
	Token string `json:"token,omitempty"`
	// Valid 是校验结论；仅 verify 时出现。
	Valid *bool `json:"valid,omitempty"`
	// Status 是时间状态；仅载荷含 exp / nbf 时出现。
	Status *JWTStatus `json:"status,omitempty"`
}

// JWTCodec 执行 JWT 的解析、生成与校验。
//
// **只**支持 HS 系列签名：RS / ES 系列的签名真实性校验不在本工具范围内，
// 遇到非 HS 算法一律返回 UNSUPPORTED_ALGORITHM，而不是「假装校验通过」。
func JWTCodec(_ context.Context, req *JWTCodecRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	switch req.Operation {
	case JWTDecode:
		return jwtDecode(req.Token)
	case JWTEncode:
		return jwtEncode(req)
	case JWTVerify:
		return jwtVerify(req)
	default:
		return tools.Result{}, tools.InvalidField("operation",
			[]string{JWTDecode, JWTEncode, JWTVerify}, req.Operation)
	}
}

// Validate 做字段级校验。
func (r *JWTCodecRequest) Validate() error {
	switch r.Operation {
	case JWTDecode, JWTEncode, JWTVerify:
	default:
		return tools.InvalidField("operation",
			[]string{JWTDecode, JWTEncode, JWTVerify}, r.Operation)
	}

	switch r.Algorithm {
	case "", JWTHS256, JWTHS384, JWTHS512:
	default:
		return tools.UnsupportedAlgorithm(r.Algorithm,
			[]string{JWTHS256, JWTHS384, JWTHS512})
	}

	switch r.Operation {
	case JWTDecode:
		if strings.TrimSpace(r.Token) == "" {
			return tools.MissingField("token")
		}
	case JWTEncode:
		if r.Header == nil {
			return tools.MissingField("header")
		}
		if r.Payload == nil {
			return tools.MissingField("payload")
		}
		if r.Secret == "" {
			return tools.MissingField("secret")
		}
	case JWTVerify:
		if strings.TrimSpace(r.Token) == "" {
			return tools.MissingField("token")
		}
		if r.Secret == "" {
			return tools.MissingField("secret")
		}
	}

	return nil
}

// jwtDecode 解析令牌。
func jwtDecode(token string) (tools.Result, error) {
	header, payload, signature, err := splitJWT(token)
	if err != nil {
		return tools.Result{}, err
	}

	return tools.Result{Data: JWTCodecResult{
		Header:    header,
		Payload:   payload,
		Signature: base64.RawURLEncoding.EncodeToString(signature),
		Status:    jwtStatus(payload),
	}}, nil
}

// jwtEncode 生成令牌。
func jwtEncode(req *JWTCodecRequest) (tools.Result, error) {
	algorithm := req.algorithm()

	header := make(map[string]any, len(req.Header)+2)
	for k, v := range req.Header {
		header[k] = v
	}
	if declared, ok := header["alg"].(string); ok && !strings.EqualFold(declared, algorithm) {
		return tools.Result{}, tools.InvalidField("header.alg",
			[]string{algorithm}, declared)
	}
	header["alg"] = algorithm
	if _, ok := header["typ"]; !ok {
		header["typ"] = "JWT"
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return tools.Result{}, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "header",
			"reason": "header is not serializable",
		})
	}
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		return tools.Result{}, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "payload",
			"reason": "payload is not serializable",
		})
	}

	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON)

	sig, err := signJWT(algorithm, req.Secret, signingInput)
	if err != nil {
		return tools.Result{}, err
	}

	return tools.Result{Data: JWTCodecResult{
		Header:    header,
		Payload:   req.Payload,
		Signature: base64.RawURLEncoding.EncodeToString(sig),
		Token:     signingInput + "." + base64.RawURLEncoding.EncodeToString(sig),
	}}, nil
}

// jwtVerify 校验 HS 系列令牌的签名。
func jwtVerify(req *JWTCodecRequest) (tools.Result, error) {
	header, payload, signature, err := splitJWT(req.Token)
	if err != nil {
		return tools.Result{}, err
	}

	declared, _ := header["alg"].(string)
	if declared == "" {
		return tools.Result{}, toolerr.Newf(toolerr.CodeInvalidJWT, map[string]any{
			"reason": "missing alg header",
		})
	}
	if !isHSAlgorithm(declared) {
		return tools.Result{}, tools.UnsupportedAlgorithm(declared,
			[]string{JWTHS256, JWTHS384, JWTHS512})
	}
	// 显式指定的算法必须与头部一致：否则等于允许算法混淆。
	if req.Algorithm != "" && !strings.EqualFold(req.Algorithm, declared) {
		return tools.Result{}, tools.InvalidField("algorithm",
			[]string{declared}, req.Algorithm)
	}

	parts := strings.Split(req.Token, ".")
	if len(parts) != 3 {
		return tools.Result{}, toolerr.Newf(toolerr.CodeInvalidJWT, map[string]any{
			"reason": "malformed token",
		})
	}

	expected, err := signJWT(declared, req.Secret, parts[0]+"."+parts[1])
	if err != nil {
		return tools.Result{}, err
	}

	valid := cryptosvc.MACEqual(expected, signature)

	return tools.Result{Data: JWTCodecResult{
		Header:    header,
		Payload:   payload,
		Signature: base64.RawURLEncoding.EncodeToString(signature),
		Valid:     &valid,
		Status:    jwtStatus(payload),
	}}, nil
}

// algorithm 返回归一化后的签名算法；缺省为 HS256。
func (r *JWTCodecRequest) algorithm() string {
	if r.Algorithm == "" {
		return JWTHS256
	}
	return strings.ToUpper(r.Algorithm)
}

// isHSAlgorithm 报告算法是否属于受支持的 HMAC-SHA 族。
func isHSAlgorithm(alg string) bool {
	switch strings.ToUpper(alg) {
	case JWTHS256, JWTHS384, JWTHS512:
		return true
	default:
		return false
	}
}

// signJWT 计算 HMAC 签名。
func signJWT(algorithm, secret, signingInput string) ([]byte, error) {
	macName, ok := jwtMACName(algorithm)
	if !ok {
		return nil, tools.UnsupportedAlgorithm(algorithm,
			[]string{JWTHS256, JWTHS384, JWTHS512})
	}
	sig, err := cryptosvc.MAC(macName, []byte(secret), []byte(signingInput))
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// jwtMACName 把 JWT 算法名映射到 cryptosvc 的 MAC 算法名。
func jwtMACName(algorithm string) (string, bool) {
	switch strings.ToUpper(algorithm) {
	case JWTHS256:
		return "HMAC-SHA256", true
	case JWTHS384:
		return "HMAC-SHA384", true
	case JWTHS512:
		return "HMAC-SHA512", true
	default:
		return "", false
	}
}

// splitJWT 拆分令牌并解码头部、载荷与签名。
func splitJWT(token string) (map[string]any, map[string]any, []byte, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return nil, nil, nil, toolerr.Newf(toolerr.CodeInvalidJWT, map[string]any{
			"reason": "malformed token",
		})
	}

	header, err := decodeJWTPart(parts[0])
	if err != nil {
		return nil, nil, nil, err
	}
	payload, err := decodeJWTPart(parts[1])
	if err != nil {
		return nil, nil, nil, err
	}
	signature, err := decodeBase64URL(parts[2])
	if err != nil {
		return nil, nil, nil, err
	}

	return header, payload, signature, nil
}

// decodeJWTPart 解码并解析令牌的 JSON 段。
func decodeJWTPart(part string) (map[string]any, error) {
	raw, err := decodeBase64URL(part)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidJWT, map[string]any{
			"reason": "segment is not a JSON object",
		})
	}
	return out, nil
}

// decodeBase64URL 解码 base64url 文本。
//
// 容忍标准 base64 的 `+` / `/` 字符与 `=` 填充：现实中不少实现并不严格。
func decodeBase64URL(s string) ([]byte, error) {
	normalized := strings.TrimRight(strings.ReplaceAll(
		strings.ReplaceAll(s, "+", "-"), "/", "_"), "=")

	raw, err := base64.RawURLEncoding.DecodeString(normalized)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidJWT, map[string]any{
			"reason": "segment is not valid base64url",
		})
	}
	return raw, nil
}

// jwtStatus 从载荷的 exp / nbf 推出时间状态。
//
// 过期与未生效都是**结论**而不是错误：令牌本身仍然可解析。
func jwtStatus(payload map[string]any) *JWTStatus {
	exp, hasExp := numericClaim(payload, "exp")
	nbf, hasNbf := numericClaim(payload, "nbf")
	if !hasExp && !hasNbf {
		return nil
	}

	now := time.Now().Unix()
	status := &JWTStatus{}
	if hasExp && exp < now {
		status.Expired = true
	}
	if hasNbf && nbf > now {
		status.NotYetValid = true
	}
	return status
}

// numericClaim 读取载荷中的数值型声明。
func numericClaim(payload map[string]any, name string) (int64, bool) {
	raw, ok := payload[name]
	if !ok {
		return 0, false
	}

	switch v := raw.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}
