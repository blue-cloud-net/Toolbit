package crypto

import (
	"context"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// TOTP 操作。
const (
	// TOTPOperationGenerate 是生成验证码。
	TOTPOperationGenerate = "generate"
	// TOTPOperationVerify 是校验验证码。
	TOTPOperationVerify = "verify"
)

// TOTPOperation 返回受支持的操作（顺序稳定）。
func TOTPOperation() []string {
	return []string{TOTPOperationGenerate, TOTPOperationVerify}
}

// TOTP 支持的摘要算法。
const (
	// TOTPAlgorithmSHA1 是 HMAC-SHA1。
	TOTPAlgorithmSHA1 = "SHA1"
	// TOTPAlgorithmSHA256 是 HMAC-SHA256。
	TOTPAlgorithmSHA256 = "SHA256"
	// TOTPAlgorithmSHA512 是 HMAC-SHA512。
	TOTPAlgorithmSHA512 = "SHA512"
)

// TOTPAlgorithm 返回受支持的摘要算法（顺序稳定）。
func TOTPAlgorithm() []string {
	return []string{TOTPAlgorithmSHA1, TOTPAlgorithmSHA256, TOTPAlgorithmSHA512}
}

// TOTPRequest 是 totp 工具的请求。
type TOTPRequest struct {
	// Operation 是操作：generate / verify；默认 generate。
	Operation string `json:"operation"`
	// Secret 是 Base32 编码的密钥。
	Secret string `json:"secret"`
	// Algorithm 是摘要算法；默认 SHA1。
	Algorithm string `json:"algorithm"`
	// Digits 是验证码位数；默认 6。
	Digits int `json:"digits"`
	// Period 是时间步长秒数；默认 30。
	Period int `json:"period"`
	// Timestamp 是时间戳秒数；留空用当前时间。
	Timestamp int64 `json:"timestamp"`
	// Code 是待校验的验证码；仅 verify 操作使用。
	Code string `json:"code"`
	// Window 是校验时允许的前后时间步数；默认 1。
	Window int `json:"window"`
}

// TOTPResult 是 totp 工具的响应。
type TOTPResult struct {
	// Operation 是实际执行的操作。
	Operation string `json:"operation"`
	// Code 是生成的验证码；校验通过时也回填实际值。
	Code string `json:"code"`
	// Valid 是校验结论；generate 操作时为 nil。
	Valid *bool `json:"valid"`
	// Algorithm 是使用的摘要算法。
	Algorithm string `json:"algorithm"`
	// Digits 是验证码位数。
	Digits int `json:"digits"`
	// Period 是时间步长秒数。
	Period int `json:"period"`
	// Counter 是当前时间步计数器。
	Counter int64 `json:"counter"`
	// ValidFrom 是当前时间步的起始时间。
	ValidFrom string `json:"validFrom"`
	// ValidUntil 是当前时间步的结束时间。
	ValidUntil string `json:"validUntil"`
	// RemainingSeconds 是当前时间步剩余秒数。
	RemainingSeconds int `json:"remainingSeconds"`
}

// 默认参数。
const (
	defaultTOTPDigits = 6
	defaultTOTPPeriod = 30
	defaultTOTPWindow = 1
)

// maxTOTPSecretLength 是密钥长度上限。
const maxTOTPSecretLength = 512

// TOTP 生成与校验动态口令。
//
// HMAC 计算**必须**经 cryptosvc 转调 Tongsuo（AGENTS §15）：
// 本文件只用标准库做 **Base32 解码**与 **RFC 6238 定义的动态截断**，
// 不自行实现 HMAC。
func TOTP(_ context.Context, req *TOTPRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	operation := req.Operation
	if operation == "" {
		operation = TOTPOperationGenerate
	}
	algorithm := req.Algorithm
	if algorithm == "" {
		algorithm = TOTPAlgorithmSHA1
	}
	digits := req.Digits
	if digits == 0 {
		digits = defaultTOTPDigits
	}
	period := req.Period
	if period == 0 {
		period = defaultTOTPPeriod
	}
	window := req.Window
	if window == 0 {
		window = defaultTOTPWindow
	}

	secret, err := decodeTOTPSecret(req.Secret)
	if err != nil {
		return tools.Result{}, err
	}

	timestamp := req.Timestamp
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	counter := timestamp / int64(period)
	result := TOTPResult{
		Operation:        operation,
		Algorithm:        algorithm,
		Digits:           digits,
		Period:           period,
		Counter:          counter,
		ValidFrom:        time.Unix(counter*int64(period), 0).UTC().Format(time.RFC3339),
		ValidUntil:       time.Unix((counter+1)*int64(period), 0).UTC().Format(time.RFC3339),
		RemainingSeconds: int(int64(period) - timestamp%int64(period)),
	}

	switch operation {
	case TOTPOperationVerify:
		valid := verifyTOTPCode(secret, algorithm, digits, period, counter, window, req.Code)
		result.Valid = &valid
		if valid {
			result.Code = req.Code
		}
	default:
		code, err := generateTOTPCode(secret, algorithm, digits, counter)
		if err != nil {
			return tools.Result{}, err
		}
		result.Code = code
		result.Valid = nil
	}

	return tools.Result{Data: result}, nil
}

// generateTOTPCode 生成当前时间步的验证码。
func generateTOTPCode(secret []byte, algorithm string, digits int, counter int64) (string, error) {
	macAlgorithm := cryptosvc.MACAlgorithmForDigest(algorithm)
	if macAlgorithm == "" {
		return "", tools.UnsupportedAlgorithm(algorithm, TOTPAlgorithm())
	}

	macValue, err := cryptosvc.MAC(macAlgorithm, secret, counterBytes(counter))
	if err != nil {
		return "", err
	}

	return dynamicTruncate(macValue, digits), nil
}

// verifyTOTPCode 校验验证码。
//
// 在允许的时间步窗口内逐个尝试，任一匹配即通过。
func verifyTOTPCode(secret []byte, algorithm string, digits, period int, counter int64, window int, code string) bool {
	if strings.TrimSpace(code) == "" {
		return false
	}

	for offset := -window; offset <= window; offset++ {
		candidate := counter + int64(offset)
		expected, err := generateTOTPCode(secret, algorithm, digits, candidate)
		if err != nil {
			return false
		}
		if cryptosvc.MACEqual([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

// counterBytes 把计数器转成 8 字节大端。
func counterBytes(counter int64) []byte {
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, uint64(counter))
	return buffer
}

// dynamicTruncate 执行 RFC 4226 定义的动态截断。
//
// 这不是密码学原语，而是标准规定的**取值步骤**：
// 取 MAC 末字节的低 4 位作为偏移，再取 4 字节并对 10^digits 取模。
func dynamicTruncate(macValue []byte, digits int) string {
	offset := int(macValue[len(macValue)-1] & 0x0f)
	value := binary.BigEndian.Uint32(macValue[offset:offset+4]) & 0x7fffffff
	modulus := uint32(1)
	for i := 0; i < digits; i++ {
		modulus *= 10
	}
	return fmt.Sprintf("%0*d", digits, value%modulus)
}

// decodeTOTPSecret 解码 Base32 密钥。
func decodeTOTPSecret(secret string) ([]byte, error) {
	normalized := strings.ToUpper(strings.TrimSpace(secret))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	if normalized == "" {
		return nil, tools.MissingField("secret")
	}
	// 补齐到 8 的倍数。
	if pad := len(normalized) % 8; pad != 0 {
		normalized += strings.Repeat("=", 8-pad)
	}
	decoded, err := base32.StdEncoding.DecodeString(normalized)
	if err != nil {
		return nil, tools.InvalidField("secret", "Base32 encoded secret", "decode error")
	}
	if len(decoded) == 0 {
		return nil, tools.InvalidField("secret", "Base32 encoded secret", "decode error")
	}
	return decoded, nil
}

// Validate 做字段级校验。
func (r *TOTPRequest) Validate() error {
	operation := r.Operation
	if operation == "" {
		operation = TOTPOperationGenerate
	}
	if !containsString(TOTPOperation(), operation) {
		return tools.InvalidField("operation", TOTPOperation(), operation)
	}
	if strings.TrimSpace(r.Secret) == "" {
		return tools.MissingField("secret")
	}
	if len(r.Secret) > maxTOTPSecretLength {
		return tools.InvalidField("secret", []int{0, maxTOTPSecretLength}, len(r.Secret))
	}
	algorithm := r.Algorithm
	if algorithm == "" {
		algorithm = TOTPAlgorithmSHA1
	}
	if !containsString(TOTPAlgorithm(), algorithm) {
		return tools.InvalidField("algorithm", TOTPAlgorithm(), algorithm)
	}
	if r.Digits != 0 && (r.Digits < 6 || r.Digits > 10) {
		return tools.InvalidField("digits", []int{6, 10}, r.Digits)
	}
	if r.Period != 0 && (r.Period < 1 || r.Period > 300) {
		return tools.InvalidField("period", []int{1, 300}, r.Period)
	}
	if r.Window < 0 || r.Window > 10 {
		return tools.InvalidField("window", []int{0, 10}, r.Window)
	}
	if operation == TOTPOperationVerify && strings.TrimSpace(r.Code) == "" {
		return tools.MissingField("code")
	}
	return nil
}
