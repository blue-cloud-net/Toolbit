package encoding

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func TestJWTCodecDecode(t *testing.T) {
	// 头部 {"alg":"HS256","typ":"JWT"}，载荷 {"sub":"1234567890"}。
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIn0." +
		"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	res, err := JWTCodec(context.Background(), &JWTCodecRequest{
		Operation: JWTDecode,
		Token:     token,
	})
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	got := res.Data.(JWTCodecResult)
	if got.Header["alg"] != "HS256" {
		t.Errorf("alg = %v，期望 HS256", got.Header["alg"])
	}
	if got.Payload["sub"] != "1234567890" {
		t.Errorf("sub = %v，期望 1234567890", got.Payload["sub"])
	}
	if got.Signature == "" {
		t.Error("签名不应为空")
	}
}

func TestJWTCodecEncodeAndVerify(t *testing.T) {
	req := &JWTCodecRequest{
		Operation: JWTEncode,
		Header:    map[string]any{"alg": JWTHS256},
		Payload:   map[string]any{"sub": "42"},
		Secret:    "s3cret",
	}
	res, err := JWTCodec(context.Background(), req)
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}
	token := res.Data.(JWTCodecResult).Token
	if token == "" {
		t.Fatal("令牌不应为空")
	}

	verifyRes, err := JWTCodec(context.Background(), &JWTCodecRequest{
		Operation: JWTVerify,
		Token:     token,
		Secret:    "s3cret",
	})
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	verified := verifyRes.Data.(JWTCodecResult)
	if verified.Valid == nil || !*verified.Valid {
		t.Error("签名应校验通过")
	}

	// 错误密钥必须校验失败（而不是报错）。
	wrongRes, err := JWTCodec(context.Background(), &JWTCodecRequest{
		Operation: JWTVerify,
		Token:     token,
		Secret:    "wrong",
	})
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	wrong := wrongRes.Data.(JWTCodecResult)
	if wrong.Valid == nil || *wrong.Valid {
		t.Error("错误密钥不应校验通过")
	}
}

func TestJWTCodecVerifyRejectsNonHSAlgorithm(t *testing.T) {
	// 头部 {"alg":"RS256"}。
	token := "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxIn0.c2ln"

	_, err := JWTCodec(context.Background(), &JWTCodecRequest{
		Operation: JWTVerify,
		Token:     token,
		Secret:    "s3cret",
	})
	assertToolErrCode(t, err, toolerr.CodeUnsupportedAlgorithm)
}

func TestJWTCodecRejectsMalformedToken(t *testing.T) {
	_, err := JWTCodec(context.Background(), &JWTCodecRequest{
		Operation: JWTDecode,
		Token:     "not-a-token",
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidJWT)
}

func TestJWTCodecRejectsMissingSecret(t *testing.T) {
	_, err := JWTCodec(context.Background(), &JWTCodecRequest{
		Operation: JWTVerify,
		Token:     "a.b.c",
	})
	assertToolErrCode(t, err, toolerr.CodeMissingField)
}

func TestJWTCodecStatusReportsExpiry(t *testing.T) {
	// 载荷 {"exp":1} —— 早已过期。
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"exp":1}`))
	token := header + "." + payload + ".c2ln"

	res, err := JWTCodec(context.Background(), &JWTCodecRequest{
		Operation: JWTDecode,
		Token:     token,
	})
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	status := res.Data.(JWTCodecResult).Status
	if status == nil || !status.Expired {
		t.Error("过期令牌应标记 expired")
	}
}
