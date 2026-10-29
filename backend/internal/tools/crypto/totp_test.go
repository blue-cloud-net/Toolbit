package crypto

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// RFC 6238 的 SHA1 测试向量密钥（Base32）。
const rfc6238Secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

// TestTOTPRFC6238Vectors 断言 RFC 6238 的 SHA1 标准向量。
//
// 标准向量：时间戳 59 → 94287082，1111111109 → 07081804。
func TestTOTPRFC6238Vectors(t *testing.T) {
	cases := []struct {
		timestamp int64
		want      string
	}{
		{59, "94287082"},
		{1111111109, "07081804"},
		{1111111111, "14050471"},
		{1234567890, "89005924"},
		{2000000000, "69279037"},
		{20000000000, "65353130"},
	}

	for _, c := range cases {
		result, err := TOTP(t.Context(), &TOTPRequest{
			Operation: TOTPOperationGenerate,
			Secret:    rfc6238Secret,
			Algorithm: TOTPAlgorithmSHA1,
			Digits:    8,
			Period:    30,
			Timestamp: c.timestamp,
		})
		if err != nil {
			t.Fatalf("生成失败（时间戳 %d）：%v", c.timestamp, err)
		}
		data := result.Data.(TOTPResult)
		if data.Code != c.want {
			t.Errorf("时间戳 %d 的验证码 = %q，期望 %q", c.timestamp, data.Code, c.want)
		}
	}
}

// TestTOTPVerify 断言校验通过与不通过。
func TestTOTPVerify(t *testing.T) {
	result, err := TOTP(t.Context(), &TOTPRequest{
		Operation: TOTPOperationVerify,
		Secret:    rfc6238Secret,
		Algorithm: TOTPAlgorithmSHA1,
		Digits:    8,
		Timestamp: 59,
		Code:      "94287082",
	})
	data := result.Data.(TOTPResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.Valid == nil || !*data.Valid {
		t.Error("正确验证码应校验通过")
	}

	wrong, err := TOTP(t.Context(), &TOTPRequest{
		Operation: TOTPOperationVerify,
		Secret:    rfc6238Secret,
		Algorithm: TOTPAlgorithmSHA1,
		Digits:    8,
		Timestamp: 59,
		Code:      "00000000",
	})
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	wrongData := wrong.Data.(TOTPResult)
	if wrongData.Valid == nil || *wrongData.Valid {
		t.Error("错误验证码不应校验通过")
	}
}

// TestTOTPMissingSecret 断言缺少密钥返回 MISSING_FIELD。
func TestTOTPMissingSecret(t *testing.T) {
	_, err := TOTP(t.Context(), &TOTPRequest{Operation: TOTPOperationGenerate})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestTOTPInvalidSecret 断言非法 Base32 密钥被拒。
func TestTOTPInvalidSecret(t *testing.T) {
	_, err := TOTP(t.Context(), &TOTPRequest{
		Operation: TOTPOperationGenerate,
		Secret:    "0189",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
