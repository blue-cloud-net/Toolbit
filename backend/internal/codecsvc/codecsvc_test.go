package codecsvc

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":           EncodingUTF8,
		"utf8":       EncodingUTF8,
		"UTF8":       EncodingUTF8,
		"hex":        EncodingHex,
		"base64":     EncodingBase64,
		"base64url":  EncodingBase64URL,
		" base64url": EncodingBase64URL,
	}
	for in, want := range cases {
		got, err := Normalize(in)
		if err != nil {
			t.Fatalf("Normalize(%q) 失败：%v", in, err)
		}
		if got != want {
			t.Errorf("Normalize(%q) = %q，期望 %q", in, got, want)
		}
	}
	if _, err := Normalize("rot13"); err == nil {
		t.Error("未知编码应报错")
	}
}

func TestDecodeEncodeRoundTrip(t *testing.T) {
	raw := []byte{0x00, 0x01, 0xfe, 0xff, 'A'}
	for _, enc := range Encodings() {
		if enc == EncodingUTF8 {
			// UTF-8 不是字节级可逆编码，跳过。
			continue
		}
		text, err := Encode(raw, enc)
		if err != nil {
			t.Fatalf("Encode(%s) 失败：%v", enc, err)
		}
		got, err := Decode(text, enc)
		if err != nil {
			t.Fatalf("Decode(%s) 失败：%v", enc, err)
		}
		if string(got) != string(raw) {
			t.Errorf("%s 往返不一致：%x", enc, got)
		}
	}
}

func TestDecodeBase64IgnoresWhitespace(t *testing.T) {
	raw := []byte("hello world")
	text := base64.StdEncoding.EncodeToString(raw)
	spaced := text[:4] + "\n" + text[4:8] + "  " + text[8:]
	got, err := Decode(spaced, EncodingBase64)
	if err != nil {
		t.Fatalf("解码失败：%v", err)
	}
	if string(got) != string(raw) {
		t.Errorf("忽略空白后解码结果 = %q", string(got))
	}
}

func TestDecodeBase64AcceptsMissingPadding(t *testing.T) {
	got, err := Decode("aGVsbG8", EncodingBase64)
	if err != nil {
		t.Fatalf("解码失败：%v", err)
	}
	if string(got) != "hello" {
		t.Errorf("解码结果 = %q，期望 hello", string(got))
	}
}

func TestDecodeInvalidHex(t *testing.T) {
	_, err := Decode("zz", EncodingHex)
	if err == nil {
		t.Fatal("非法十六进制应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("期望 *toolerr.Error，实际 %T", err)
	}
	if toolErr.Code != toolerr.CodeInvalidHex {
		t.Errorf("错误码 = %v，期望 %v", toolErr.Code, toolerr.CodeInvalidHex)
	}
}

func TestPEMRoundTrip(t *testing.T) {
	der := []byte{0x30, 0x03, 0x02, 0x01, 0x05}
	pemBytes, err := ToPEM(der, BlockPublicKey)
	if err != nil {
		t.Fatalf("ToPEM 失败：%v", err)
	}
	if !LooksLikePEM(string(pemBytes)) {
		t.Error("生成的 PEM 未被识别")
	}
	got, blockType, err := FromPEM(pemBytes)
	if err != nil {
		t.Fatalf("FromPEM 失败：%v", err)
	}
	if blockType != BlockPublicKey {
		t.Errorf("块类型 = %q，期望 %q", blockType, BlockPublicKey)
	}
	if hex.EncodeToString(got) != hex.EncodeToString(der) {
		t.Errorf("内容不一致：%x", got)
	}
}

func TestFromPEMInvalid(t *testing.T) {
	_, _, err := FromPEM([]byte("not a pem"))
	if err == nil {
		t.Fatal("非法 PEM 应报错")
	}
}

func TestLooksLikeDER(t *testing.T) {
	if !LooksLikeDER([]byte{0x30, 0x00}) {
		t.Error("0x30 开头应判定为 DER")
	}
	if LooksLikeDER([]byte{0x04, 0x00}) {
		t.Error("0x04 开头不应判定为 DER")
	}
}
