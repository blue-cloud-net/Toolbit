package documents

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// sampleEML 是一封带纯文本、HTML 与附件的邮件。
const sampleEML = "From: sender@example.com\r\n" +
	"To: receiver@example.com\r\n" +
	"Subject: 测试邮件\r\n" +
	"Date: Mon, 14 Nov 2023 22:13:20 +0000\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Type: multipart/mixed; boundary=\"BOUND\"\r\n" +
	"\r\n" +
	"--BOUND\r\n" +
	"Content-Type: text/plain; charset=utf-8\r\n" +
	"\r\n" +
	"纯文本正文\r\n" +
	"--BOUND\r\n" +
	"Content-Type: text/html; charset=utf-8\r\n" +
	"\r\n" +
	"<p>HTML 正文</p>\r\n" +
	"--BOUND\r\n" +
	"Content-Type: application/octet-stream; name=\"data.bin\"\r\n" +
	"Content-Disposition: attachment; filename=\"data.bin\"\r\n" +
	"\r\n" +
	"ATTACHMENT-BYTES\r\n" +
	"--BOUND--\r\n"

// TestEMLParserBasic 断言头部与正文解析。
func TestEMLParserBasic(t *testing.T) {
	result, err := EMLParser(context.Background(), &EMLParserRequest{
		Input: EMLInput{Text: sampleEML},
	})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}

	if result.HeaderMap["subject"] != "测试邮件" {
		t.Errorf("主题 = %q", result.HeaderMap["subject"])
	}
	if result.HeaderMap["from"] != "sender@example.com" {
		t.Errorf("发件人 = %q", result.HeaderMap["from"])
	}
	if len(result.Headers) == 0 {
		t.Error("缺少原始头字段")
	}

	if len(result.Bodies) != 2 {
		t.Fatalf("正文数 = %d，期望 2", len(result.Bodies))
	}
	if !strings.Contains(result.Bodies[0].Text, "纯文本正文") {
		t.Errorf("纯文本正文 = %q", result.Bodies[0].Text)
	}
	if result.Bodies[1].HTML == "" {
		t.Error("缺少 HTML 正文")
	}
}

// TestEMLParserAttachment 断言附件解析。
func TestEMLParserAttachment(t *testing.T) {
	result, err := EMLParser(context.Background(), &EMLParserRequest{
		Input: EMLInput{Text: sampleEML},
	})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if len(result.Attachments) != 1 {
		t.Fatalf("附件数 = %d，期望 1", len(result.Attachments))
	}
	attachment := result.Attachments[0]
	if attachment.Filename != "data.bin" {
		t.Errorf("文件名 = %q", attachment.Filename)
	}
	if attachment.MIME != "application/octet-stream" {
		t.Errorf("媒体类型 = %q", attachment.MIME)
	}
	raw, err := base64.StdEncoding.DecodeString(attachment.RawBase64)
	if err != nil {
		t.Fatalf("附件 base64 解析失败：%v", err)
	}
	if string(raw) != "ATTACHMENT-BYTES" {
		t.Errorf("附件内容 = %q", string(raw))
	}
}

// TestEMLParserAttachmentTruncated 断言超限附件被截断。
func TestEMLParserAttachmentTruncated(t *testing.T) {
	result, err := EMLParser(context.Background(), &EMLParserRequest{
		Input:              EMLInput{Text: sampleEML},
		MaxAttachmentBytes: 4,
	})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if len(result.Attachments) != 1 {
		t.Fatalf("附件数 = %d，期望 1", len(result.Attachments))
	}
	attachment := result.Attachments[0]
	if !attachment.Truncated {
		t.Error("期望 truncated 为真")
	}
	if attachment.RawBase64 != "" {
		t.Error("超限附件不应返回内容")
	}
}

// TestEMLParserInvalid 断言非法邮件报 INVALID_FIELD。
func TestEMLParserInvalid(t *testing.T) {
	_, err := EMLParser(context.Background(), &EMLParserRequest{
		Input: EMLInput{Text: "这不是一封邮件"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestEMLParserMissingInput 断言缺少输入时报 MISSING_FIELD。
func TestEMLParserMissingInput(t *testing.T) {
	_, err := EMLParser(context.Background(), &EMLParserRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestEMLParserInvalidLimit 断言超限的附件上限报 INVALID_FIELD。
func TestEMLParserInvalidLimit(t *testing.T) {
	_, err := EMLParser(context.Background(), &EMLParserRequest{
		Input:              EMLInput{Text: sampleEML},
		MaxAttachmentBytes: 1 << 30,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestSanitizeHeaderValue 断言头字段换行被折叠。
func TestSanitizeHeaderValue(t *testing.T) {
	got := sanitizeHeaderValue("a\r\nb")
	if strings.ContainsAny(got, "\r\n") {
		t.Errorf("头字段仍含换行：%q", got)
	}
}
