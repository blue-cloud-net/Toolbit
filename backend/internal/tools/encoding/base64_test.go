package encoding

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func TestBase64Encode(t *testing.T) {
	cases := []struct {
		name  string
		input ByteInput
		want  string
	}{
		{"空字符串", ByteInput{Data: "", Encoding: EncodingUTF8}, ""},
		{"Hello", ByteInput{Data: "Hello", Encoding: EncodingUTF8}, "SGVsbG8="},
		{"中文", ByteInput{Data: "机密内容", Encoding: EncodingUTF8}, "5py65a+G5YaF5a65"},
		{"十六进制输入", ByteInput{Data: "48656c6c6f", Encoding: EncodingHex}, "SGVsbG8="},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Base64(context.Background(), &Base64Request{
				Operation: Base64Encode,
				Input:     tc.input,
			})
			if err != nil {
				t.Fatalf("编码失败: %v", err)
			}
			got := res.Data.(Base64Result).Output
			if got != tc.want {
				t.Errorf("输出 = %q，期望 %q", got, tc.want)
			}
		})
	}
}

func TestBase64Decode(t *testing.T) {
	cases := []struct {
		name string
		data string
		want string
	}{
		{"标准", "SGVsbG8=", "Hello"},
		{"忽略换行", "SGVs\nbG8=", "Hello"},
		{"忽略空白", "  SGVsbG8=  ", "Hello"},
		{"无填充输入", "SGVsbG8", "Hello"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Base64(context.Background(), &Base64Request{
				Operation: Base64Decode,
				Input:     ByteInput{Data: tc.data, Encoding: EncodingUTF8},
			})
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}
			got := res.Data.(Base64Result).Output
			if got != tc.want {
				t.Errorf("输出 = %q，期望 %q", got, tc.want)
			}
		})
	}
}

func TestBase64RoundTrip(t *testing.T) {
	original := "往返测试 ✓ 123"
	enc, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Encode,
		Input:     ByteInput{Data: original, Encoding: EncodingUTF8},
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	encoded := enc.Data.(Base64Result).Output

	dec, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Decode,
		Input:     ByteInput{Data: encoded, Encoding: EncodingUTF8},
	})
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	if got := dec.Data.(Base64Result).Output; got != original {
		t.Errorf("往返结果 = %q，期望 %q", got, original)
	}
}

func TestBase64URLVariant(t *testing.T) {
	// 0xFB 0xFF 在标准字符集中会产生 + 与 /，URL 变体应为 - 与 _。
	raw := ByteInput{Data: "fbff", Encoding: EncodingHex}

	std, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Encode, Input: raw, Variant: Base64VariantStandard,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	url, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Encode, Input: raw, Variant: Base64VariantURL,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}

	stdOut := std.Data.(Base64Result).Output
	urlOut := url.Data.(Base64Result).Output
	if stdOut != "+/8=" {
		t.Errorf("标准变体输出 = %q，期望 %q", stdOut, "+/8=")
	}
	if urlOut != "-_8=" {
		t.Errorf("URL 变体输出 = %q，期望 %q", urlOut, "-_8=")
	}
}

func TestBase64PaddingDisabled(t *testing.T) {
	no := false
	res, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Encode,
		Input:     ByteInput{Data: "Hello", Encoding: EncodingUTF8},
		Padding:   &no,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	if got := res.Data.(Base64Result).Output; got != "SGVsbG8" {
		t.Errorf("输出 = %q，期望 %q", got, "SGVsbG8")
	}
}

func TestBase64InvalidInput(t *testing.T) {
	_, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Decode,
		Input:     ByteInput{Data: "!!!not base64!!!", Encoding: EncodingUTF8},
	})
	if err == nil {
		t.Fatal("期望解码失败")
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidBase64 {
		t.Errorf("错误码 = %s，期望 INVALID_BASE64", toolErr.Code)
	}
}

func TestBase64InvalidOperation(t *testing.T) {
	_, err := Base64(context.Background(), &Base64Request{
		Operation: "transcode",
		Input:     ByteInput{Data: "x", Encoding: EncodingUTF8},
	})
	if err == nil {
		t.Fatal("期望校验失败")
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 INVALID_FIELD", toolErr.Code)
	}
}

func TestBase64InvalidHexInput(t *testing.T) {
	_, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Encode,
		Input:     ByteInput{Data: "zz", Encoding: EncodingHex},
	})
	if err == nil {
		t.Fatal("期望十六进制解析失败")
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidHex {
		t.Errorf("错误码 = %s，期望 INVALID_HEX", toolErr.Code)
	}
}
