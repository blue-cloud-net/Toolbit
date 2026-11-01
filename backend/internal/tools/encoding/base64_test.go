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

func TestBase64Convert(t *testing.T) {
	cases := []struct {
		name         string
		data         string
		variant      string
		target       string
		wantOutput   string
		wantEncoding string
	}{
		{"标准转 URL 安全", "+/8=", "", Base64VariantURL, "-_8=", "base64url"},
		{"URL 安全转标准", "-_8=", "", Base64VariantStandard, "+/8=", "base64"},
		{"自动识别输入变体", "-_8=", Base64VariantAuto, Base64VariantStandard, "+/8=", "base64"},
		{"显式指定输入变体", "-_8=", Base64VariantURL, Base64VariantStandard, "+/8=", "base64"},
		{"忽略换行", "+/\n8=", "", Base64VariantURL, "-_8=", "base64url"},
		{"忽略填充", "+/8", "", Base64VariantURL, "-_8=", "base64url"},
		{"无特殊字符时两变体等价", "SGVsbG8=", "", Base64VariantURL, "SGVsbG8=", "base64url"},
		{"缺省目标变体为 url", "+/8=", "", "", "-_8=", "base64url"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Base64(context.Background(), &Base64Request{
				Operation:     Base64Convert,
				Input:         ByteInput{Data: tc.data, Encoding: EncodingUTF8},
				Variant:       tc.variant,
				TargetVariant: tc.target,
			})
			if err != nil {
				t.Fatalf("转换失败: %v", err)
			}
			got := res.Data.(Base64Result)
			if got.Output != tc.wantOutput {
				t.Errorf("输出 = %q，期望 %q", got.Output, tc.wantOutput)
			}
			if got.Encoding != tc.wantEncoding {
				t.Errorf("编码 = %q，期望 %q", got.Encoding, tc.wantEncoding)
			}
		})
	}
}

func TestBase64ConvertPadding(t *testing.T) {
	no := false
	res, err := Base64(context.Background(), &Base64Request{
		Operation:     Base64Convert,
		Input:         ByteInput{Data: "+/8=", Encoding: EncodingUTF8},
		TargetVariant: Base64VariantURL,
		Padding:       &no,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if got := res.Data.(Base64Result).Output; got != "-_8" {
		t.Errorf("输出 = %q，期望 %q", got, "-_8")
	}
}

func TestBase64ConvertRoundTrip(t *testing.T) {
	original := "+/8="
	first, err := Base64(context.Background(), &Base64Request{
		Operation:     Base64Convert,
		Input:         ByteInput{Data: original, Encoding: EncodingUTF8},
		TargetVariant: Base64VariantURL,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	second, err := Base64(context.Background(), &Base64Request{
		Operation:     Base64Convert,
		Input:         ByteInput{Data: first.Data.(Base64Result).Output, Encoding: EncodingUTF8},
		TargetVariant: Base64VariantStandard,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if got := second.Data.(Base64Result).Output; got != original {
		t.Errorf("往返结果 = %q，期望 %q", got, original)
	}
}

func TestBase64ConvertMixedAlphabets(t *testing.T) {
	_, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Convert,
		Input:     ByteInput{Data: "+-_8=", Encoding: EncodingUTF8},
	})
	if err == nil {
		t.Fatal("期望混用字母表时失败")
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidBase64 {
		t.Errorf("错误码 = %s，期望 INVALID_BASE64", toolErr.Code)
	}
}

func TestBase64ConvertInvalidTargetVariant(t *testing.T) {
	_, err := Base64(context.Background(), &Base64Request{
		Operation:     Base64Convert,
		Input:         ByteInput{Data: "SGVsbG8=", Encoding: EncodingUTF8},
		TargetVariant: "base32",
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

func TestBase64ConvertRejectsHexInputEncoding(t *testing.T) {
	_, err := Base64(context.Background(), &Base64Request{
		Operation: Base64Convert,
		Input:     ByteInput{Data: "2f384d3d", Encoding: EncodingHex},
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
