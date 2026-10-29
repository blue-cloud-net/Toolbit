package data

import (
	"encoding/base64"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestFileAnalyzePNG 断言识别 PNG。
func TestFileAnalyzePNG(t *testing.T) {
	raw := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x01}
	result, err := FileAnalyze(t.Context(), &FileAnalyzeRequest{
		Data:            base64.StdEncoding.EncodeToString(raw),
		FileName:        "a.png",
		DigestAlgorithm: "SHA256",
	})
	data := result.Data.(FileAnalyzeResult)
	if err != nil {
		t.Fatalf("识别失败：%v", err)
	}
	if data.Extension != "png" {
		t.Errorf("扩展名 = %q，期望 png", data.Extension)
	}
	if data.Digest == "" {
		t.Error("应给出摘要")
	}
}

// TestFileAnalyzeMissingData 断言缺少数据返回 MISSING_FIELD。
func TestFileAnalyzeMissingData(t *testing.T) {
	_, err := FileAnalyze(t.Context(), &FileAnalyzeRequest{FileName: "a.txt"})
	assertCode(t, err, toolerr.CodeMissingField)
}
