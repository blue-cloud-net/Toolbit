package encoding

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func TestEncodingConvertGBKToUTF8(t *testing.T) {
	// “中文”的 GBK 编码。
	gbk := []byte{0xD6, 0xD0, 0xCE, 0xC4}

	res, err := EncodingConvert(context.Background(), &EncodingConvertRequest{
		Input:          ByteInput{Data: base64.StdEncoding.EncodeToString(gbk), Encoding: "base64"},
		From:           "gbk",
		To:             "utf-8",
		OutputEncoding: "utf8",
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	got := res.Data.(EncodingConvertResult)
	if got.Output != "中文" {
		t.Errorf("输出 = %q，期望 中文", got.Output)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("不应有警告，实际 %v", res.Warnings)
	}
}

func TestEncodingConvertDetectsUTF8(t *testing.T) {
	res, err := EncodingConvert(context.Background(), &EncodingConvertRequest{
		Input:          ByteInput{Data: "hello", Encoding: EncodingUTF8},
		From:           "auto",
		To:             "utf-8",
		OutputEncoding: "utf8",
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	got := res.Data.(EncodingConvertResult)
	if got.Detected != codecsvc.CharsetUTF8 {
		t.Errorf("探测结果 = %q，期望 utf-8", got.Detected)
	}
	if got.Output != "hello" {
		t.Errorf("输出 = %q，期望 hello", got.Output)
	}
}

func TestEncodingConvertReplacesInvalidBytes(t *testing.T) {
	// 0x81 在 GBK 中不是合法序列。
	res, err := EncodingConvert(context.Background(), &EncodingConvertRequest{
		Input:          ByteInput{Data: base64.StdEncoding.EncodeToString([]byte{0x81}), Encoding: "base64"},
		From:           "gbk",
		To:             "utf-8",
		OutputEncoding: "utf8",
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if len(res.Warnings) == 0 {
		t.Error("非法字节应产生警告")
	}
}

func TestEncodingConvertRejectsUnknownCharset(t *testing.T) {
	_, err := EncodingConvert(context.Background(), &EncodingConvertRequest{
		Input: ByteInput{Data: "x", Encoding: EncodingUTF8},
		From:  "nope",
		To:    "utf-8",
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidEncoding)
}

func TestEncodingConvertRejectsAutoTarget(t *testing.T) {
	_, err := EncodingConvert(context.Background(), &EncodingConvertRequest{
		Input: ByteInput{Data: "x", Encoding: EncodingUTF8},
		From:  "utf-8",
		To:    "auto",
	})
	// `auto` 是合法的字符集取值，但在「目标字符集」这个字段上不合法。
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}
