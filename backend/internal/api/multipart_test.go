package api

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// sm4TestKey 是测试用的 SM4 密钥（16 字节）。
var sm4TestKey = []byte("0123456789ABCDEF")

// multipartBody 构造一个 multipart 请求体。
//
// 文件部分**必须**是最后一个 part：分发器按顺序读取表单，
// 读到文件后就无法回头取后面的字段。
type multipartBody struct {
	// Fields 是普通字段。
	Fields map[string]string
	// FileName 是文件名。
	FileName string
	// FileContent 是文件内容。
	FileContent []byte
}

// build 序列化请求体。
func (m multipartBody) build() (string, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for name, value := range m.Fields {
		_ = w.WriteField(name, value)
	}
	if m.FileName != "" {
		part, err := w.CreateFormFile("file", m.FileName)
		if err != nil {
			panic(err)
		}
		_, _ = part.Write(m.FileContent)
	}
	_ = w.Close()
	return buf.String(), w.FormDataContentType()
}

// postMultipart 发起一次 multipart POST 请求。
func postMultipart(t *testing.T, h http.Handler, path string, body multipartBody) *httptest.ResponseRecorder {
	t.Helper()
	content, contentType := body.build()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, strings.NewReader(content))
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// keyField 生成密钥表单字段。
func keyField(key []byte) string {
	raw, _ := json.Marshal(map[string]string{
		"data":     hexString(key),
		"encoding": "hex",
	})
	return string(raw)
}

// hexString 把字节转成十六进制。
func hexString(b []byte) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 0, len(b)*2)
	for _, c := range b {
		out = append(out, hexDigits[c>>4], hexDigits[c&0x0f])
	}
	return string(out)
}

// TestSM4FileCryptRoundTrip 加密后解密，内容一致。
func TestSM4FileCryptRoundTrip(t *testing.T) {
	h := newTestServer(t)
	plain := []byte("机密内容机密内容机密内容机密内容")

	enc := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"mode":      "GCM",
			"key":       keyField(sm4TestKey),
		},
		FileName:    "report.txt",
		FileContent: plain,
	})
	if enc.Code != http.StatusOK {
		t.Fatalf("加密状态码 = %d，期望 200；body=%s", enc.Code, enc.Body.String())
	}
	if ct := enc.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("Content-Type = %q，期望 application/octet-stream", ct)
	}
	if cd := enc.Header().Get("Content-Disposition"); !strings.Contains(cd, "report.txt.sm4") {
		t.Errorf("Content-Disposition = %q，期望包含 report.txt.sm4", cd)
	}
	if cc := enc.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q，期望 no-store", cc)
	}

	container := enc.Body.Bytes()
	if len(container) < cryptosvc.TBFCHeaderSize {
		t.Fatalf("容器太短：%d 字节", len(container))
	}
	if string(container[:4]) != cryptosvc.TBFC {
		t.Errorf("魔数 = %q，期望 %q", container[:4], cryptosvc.TBFC)
	}

	dec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "decrypt",
			"mode":      "GCM",
			"key":       keyField(sm4TestKey),
		},
		FileName:    "report.txt.sm4",
		FileContent: container,
	})
	if dec.Code != http.StatusOK {
		t.Fatalf("解密状态码 = %d，期望 200；body=%s", dec.Code, dec.Body.String())
	}
	if !bytes.Equal(dec.Body.Bytes(), plain) {
		t.Errorf("解密结果 = %q，期望 %q", dec.Body.Bytes(), plain)
	}
}

// 密钥不对时必须返回干净的 JSON 错误，**不是** 200 + 截断的下载。
func TestSM4FileCryptWrongKeyReturnsError(t *testing.T) {
	h := newTestServer(t)

	enc := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField(sm4TestKey),
		},
		FileName:    "a.txt",
		FileContent: []byte("机密内容"),
	})
	if enc.Code != http.StatusOK {
		t.Fatalf("加密状态码 = %d，期望 200", enc.Code)
	}

	dec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "decrypt",
			"key":       keyField([]byte("FEDCBA9876543210")),
		},
		FileName:    "a.txt.sm4",
		FileContent: enc.Body.Bytes(),
	})
	if dec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400；body=%s", dec.Code, dec.Body.String())
	}
	env := decode(t, dec)
	if env.Error == nil {
		t.Fatal("期望有错误，实际无错误")
	}
	if env.Error.Code != string(toolerr.CodeTagInvalid) &&
		env.Error.Code != string(toolerr.CodeDecryptFailed) {
		t.Errorf("错误码 = %s，期望 %s 或 %s",
			env.Error.Code, toolerr.CodeTagInvalid, toolerr.CodeDecryptFailed)
	}
}

// 输入不是 TBFC 容器时必须报错。
func TestSM4FileCryptRejectsNonContainer(t *testing.T) {
	h := newTestServer(t)
	rec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "decrypt",
			"key":       keyField(sm4TestKey),
		},
		FileName:    "a.txt",
		FileContent: []byte("这不是容器"),
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400；body=%s", rec.Code, rec.Body.String())
	}
}

// 缺少文件部分时报 MISSING_FIELD。
func TestSM4FileCryptRequiresFile(t *testing.T) {
	h := newTestServer(t)
	rec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField(sm4TestKey),
		},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400；body=%s", rec.Code, rec.Body.String())
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != string(toolerr.CodeMissingField) {
		t.Fatalf("错误码 = %v，期望 %s", env.Error, toolerr.CodeMissingField)
	}
}

// 密钥长度不对时报 KEY_SIZE_MISMATCH。
func TestSM4FileCryptRejectsBadKeyLength(t *testing.T) {
	h := newTestServer(t)
	rec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField([]byte("short")),
		},
		FileName:    "a.txt",
		FileContent: []byte("x"),
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400；body=%s", rec.Code, rec.Body.String())
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != string(toolerr.CodeKeySizeMismatch) {
		t.Fatalf("错误码 = %v，期望 %s", env.Error, toolerr.CodeKeySizeMismatch)
	}
}

// 用 JSON 请求 multipart 工具必须报 UNSUPPORTED_MEDIA_TYPE。
func TestSM4FileCryptRejectsJSONContentType(t *testing.T) {
	h := newTestServer(t)
	rec := postJSON(t, h, "/api/v1/tools/sm4-file-crypt", `{"operation":"encrypt"}`)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("状态码 = %d，期望 415；body=%s", rec.Code, rec.Body.String())
	}
}

// 分块大小越界时报 INVALID_FIELD。
func TestSM4FileCryptRejectsBadChunkSize(t *testing.T) {
	h := newTestServer(t)
	rec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField(sm4TestKey),
			"chunkSize": "10",
		},
		FileName:    "a.txt",
		FileContent: []byte("x"),
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400；body=%s", rec.Code, rec.Body.String())
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != string(toolerr.CodeInvalidField) {
		t.Fatalf("错误码 = %v，期望 %s", env.Error, toolerr.CodeInvalidField)
	}
}

// 大文件必须跨块往返一致，且内存占用不随文件大小增长。
func TestSM4FileCryptLargeFileRoundTrip(t *testing.T) {
	h := newTestServer(t)

	// 3 块 + 一点余量。
	plain := make([]byte, cryptosvc.TBFCMinChunkSize*3+123)
	for i := range plain {
		plain[i] = byte(i % 251)
	}

	enc := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField(sm4TestKey),
			"chunkSize": itoa(cryptosvc.TBFCMinChunkSize),
		},
		FileName:    "big.bin",
		FileContent: plain,
	})
	if enc.Code != http.StatusOK {
		t.Fatalf("加密状态码 = %d，期望 200", enc.Code)
	}

	dec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "decrypt",
			"key":       keyField(sm4TestKey),
			"chunkSize": itoa(cryptosvc.TBFCMinChunkSize),
		},
		FileName:    "big.bin.sm4",
		FileContent: enc.Body.Bytes(),
	})
	if dec.Code != http.StatusOK {
		t.Fatalf("解密状态码 = %d，期望 200；body=%s", dec.Code, dec.Body.String())
	}
	if !bytes.Equal(dec.Body.Bytes(), plain) {
		t.Error("大文件往返不一致")
	}
}

// itoa 把整数转成字符串。
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

// 容器头的分块大小必须与请求一致。
func TestSM4FileCryptStoresChunkSize(t *testing.T) {
	h := newTestServer(t)
	rec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField(sm4TestKey),
			"chunkSize": itoa(cryptosvc.TBFCMinChunkSize),
		},
		FileName:    "a.txt",
		FileContent: []byte("x"),
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	raw := rec.Body.Bytes()
	got := binary.BigEndian.Uint32(raw[6:10])
	if got != cryptosvc.TBFCMinChunkSize {
		t.Errorf("容器头分块大小 = %d，期望 %d", got, cryptosvc.TBFCMinChunkSize)
	}
}

// 未实现的工具仍然返回 501。
func TestMultipartUnimplementedToolReturns501(t *testing.T) {
	h := newTestServer(t)
	// slug 从 catalog 派生：已启用但尚未实现的 multipart 工具。
	slug := unimplementedSlug(t, catalog.TransportMultipart)
	rec := postMultipart(t, h, "/api/v1/tools/"+slug, multipartBody{
		FileName:    "a.txt",
		FileContent: []byte("x"),
	})
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("状态码 = %d，期望 501；body=%s", rec.Code, rec.Body.String())
	}
}

// 表单字段超长必须被拒绝。
func TestSM4FileCryptRejectsOversizedField(t *testing.T) {
	h := newTestServer(t)
	rec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField(sm4TestKey),
			"junk":      strings.Repeat("A", 2<<20),
		},
		FileName:    "a.txt",
		FileContent: []byte("x"),
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400；body=%s", rec.Code, rec.Body.String())
	}
}

// 响应体必须**不**是 JSON 信封。
func TestSM4FileCryptResponseIsNotEnvelope(t *testing.T) {
	h := newTestServer(t)
	rec := postMultipart(t, h, "/api/v1/tools/sm4-file-crypt", multipartBody{
		Fields: map[string]string{
			"operation": "encrypt",
			"key":       keyField(sm4TestKey),
		},
		FileName:    "a.txt",
		FileContent: []byte("x"),
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	var probe map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &probe); err == nil {
		t.Error("响应体被解析成了 JSON 信封，期望二进制流")
	}
	// 确认响应体确实不是空。
	if rec.Body.Len() == 0 {
		t.Error("响应体为空")
	}
	_ = io.Discard
}
