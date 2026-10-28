package encoding

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func TestGZipDeflateRoundTrip(t *testing.T) {
	payload := "蓝云工具箱 Toolbit"

	for _, format := range []string{GZipFormatGzip, GZipFormatDeflate, GZipFormatRaw} {
		t.Run(format, func(t *testing.T) {
			compressed, err := GZipDeflate(context.Background(), &GZipDeflateRequest{
				Operation: GZipCompress,
				Format:    format,
				Input:     ByteInput{Data: payload, Encoding: EncodingUTF8},
			})
			if err != nil {
				t.Fatalf("压缩失败: %v", err)
			}
			comp := compressed.Data.(GZipDeflateResult)
			if comp.Encoding != "base64" {
				t.Errorf("输出编码 = %q，期望 base64", comp.Encoding)
			}

			decompressed, err := GZipDeflate(context.Background(), &GZipDeflateRequest{
				Operation:      GZipDecompress,
				Format:         format,
				Input:          ByteInput{Data: comp.Output, Encoding: "base64"},
				OutputEncoding: EncodingUTF8,
			})
			if err != nil {
				t.Fatalf("解压失败: %v", err)
			}
			out := decompressed.Data.(GZipDeflateResult)
			if out.Output != payload {
				t.Errorf("输出 = %q，期望 %q", out.Output, payload)
			}
		})
	}
}

func TestGZipDeflateRejectsCorruptStream(t *testing.T) {
	_, err := GZipDeflate(context.Background(), &GZipDeflateRequest{
		Operation: GZipDecompress,
		Format:    GZipFormatGzip,
		Input:     ByteInput{Data: base64.StdEncoding.EncodeToString([]byte("not gzip")), Encoding: "base64"},
	})
	assertToolErrCode(t, err, toolerr.CodeDecryptFailed)
}

func TestGZipDeflateRejectsInvalidLevel(t *testing.T) {
	level := 99
	_, err := GZipDeflate(context.Background(), &GZipDeflateRequest{
		Operation: GZipCompress,
		Input:     ByteInput{Data: "x", Encoding: EncodingUTF8},
		Level:     &level,
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}

func TestGZipDeflateRejectsUnsupportedFormat(t *testing.T) {
	_, err := GZipDeflate(context.Background(), &GZipDeflateRequest{
		Operation: GZipCompress,
		Format:    "lzma",
		Input:     ByteInput{Data: "x", Encoding: EncodingUTF8},
	})
	assertToolErrCode(t, err, toolerr.CodeUnsupportedFormat)
}
