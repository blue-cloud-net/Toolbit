package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/config"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolset"
)

// newTestServer 构造一个带两个工具实现的测试服务器。
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	return newTestServerWithLimit(t, config.DefaultMaxBodyBytes)
}

// newTestServerWithLimit 构造一个带指定全局请求体上限的测试服务器。
func newTestServerWithLimit(t *testing.T, maxBodyBytes int64) http.Handler {
	t.Helper()

	cfg := &config.Config{
		Addr:             config.DefaultAddr,
		WebRoot:          config.DefaultWebRoot,
		DBPath:           filepath.Join(t.TempDir(), "stats.db"),
		LogLevel:         "error",
		LogFormat:        "text",
		RateLimitEnabled: false,
		StatsEnabled:     false,
		NetworkEnabled:   true,
		RequestTimeout:   config.DefaultRequestTimeout,
		MaxBodyBytes:     maxBodyBytes,
		ShutdownTimeout:  config.DefaultShutdownTimeout,
	}

	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("加载目录失败: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(Options{
		Config:  cfg,
		Logger:  logger,
		Catalog: cat,
		Tools:   tools.NewRegistry(toolset.Definitions()),
	})
	return srv.Handler()
}

// postJSON 发起一次 JSON POST 请求。
func postJSON(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// get 发起一次 GET 请求。
func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// envelope 解析响应信封。
type envelope struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	Error     *errorPayload   `json:"error"`
	Meta      metaPayload     `json:"meta"`
	RequestID string          `json:"requestId"`
}

// metaPayload 是信封中的元信息。
type metaPayload struct {
	DurationMs int64            `json:"durationMs"`
	Warnings   []warningPayload `json:"warnings"`
}

// warningPayload 是单条警告。
type warningPayload struct {
	Code   string         `json:"code"`
	Params map[string]any `json:"params"`
}

// errorPayload 是信封中的错误部分。
type errorPayload struct {
	Code   string         `json:"code"`
	Params map[string]any `json:"params"`
}

// decode 解析响应信封。
func decode(t *testing.T, rec *httptest.ResponseRecorder) envelope {
	t.Helper()
	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("响应不是合法 JSON: %v；body=%s", err, rec.Body.String())
	}
	return env
}

func TestCatalogEndpointReturns70Tools(t *testing.T) {
	h := newTestServer(t)
	rec := get(t, h, "/api/v1/tools")

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	env := decode(t, rec)
	if env.Code != 0 {
		t.Errorf("code = %d，期望 0", env.Code)
	}

	var data struct {
		Version    int `json:"version"`
		Categories []struct {
			Slug string `json:"slug"`
		} `json:"categories"`
		Tools []struct {
			Slug      string `json:"slug"`
			Category  string `json:"category"`
			Transport string `json:"transport"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("解析 data 失败: %v", err)
	}
	if len(data.Tools) != 70 {
		t.Errorf("工具数量 = %d，期望 70", len(data.Tools))
	}
	if len(data.Categories) != 10 {
		t.Errorf("分类数量 = %d，期望 10", len(data.Categories))
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("缺少 ETag")
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=60" {
		t.Errorf("Cache-Control = %q，期望 public, max-age=60", got)
	}
}

func TestCatalogETagNegotiation(t *testing.T) {
	h := newTestServer(t)
	first := get(t, h, "/api/v1/tools")
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("缺少 ETag")
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tools", nil)
	req.Header.Set("If-None-Match", etag)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotModified {
		t.Errorf("状态码 = %d，期望 304", rec.Code)
	}
}

func TestBase64ToolEndToEnd(t *testing.T) {
	h := newTestServer(t)
	rec := postJSON(t, h, "/api/v1/tools/base64",
		`{"operation":"encode","input":{"data":"Hello","encoding":"utf8"}}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200；body=%s", rec.Code, rec.Body.String())
	}
	env := decode(t, rec)
	if env.Error != nil {
		t.Fatalf("错误 = %+v", env.Error)
	}
	if !strings.Contains(string(env.Data), "SGVsbG8=") {
		t.Errorf("data = %s，期望包含 SGVsbG8=", env.Data)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q，期望 no-store", got)
	}
}

func TestNumberBaseToolEndToEnd(t *testing.T) {
	h := newTestServer(t)
	rec := postJSON(t, h, "/api/v1/tools/number-base",
		`{"value":"255","from":10,"to":16}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200；body=%s", rec.Code, rec.Body.String())
	}
	env := decode(t, rec)
	if !strings.Contains(string(env.Data), `"ff"`) {
		t.Errorf("data = %s，期望包含 ff", env.Data)
	}
}

func TestUnknownToolReturns404(t *testing.T) {
	h := newTestServer(t)
	rec := postJSON(t, h, "/api/v1/tools/no-such-tool", `{}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("状态码 = %d，期望 404", rec.Code)
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "TOOL_NOT_FOUND" {
		t.Errorf("error = %+v，期望 TOOL_NOT_FOUND", env.Error)
	}
}

func TestUnimplementedToolReturns501(t *testing.T) {
	h := newTestServer(t)
	// keygen 在 catalog 中已启用，但尚未实现（v0.4.0 交付）。
	rec := postJSON(t, h, "/api/v1/tools/keygen", `{}`)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("状态码 = %d，期望 501；body=%s", rec.Code, rec.Body.String())
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "NOT_IMPLEMENTED" {
		t.Errorf("error = %+v，期望 NOT_IMPLEMENTED", env.Error)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestServer(t)
	rec := get(t, h, "/api/v1/tools/base64")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("状态码 = %d，期望 405", rec.Code)
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "METHOD_NOT_ALLOWED" {
		t.Errorf("error = %+v，期望 METHOD_NOT_ALLOWED", env.Error)
	}
}

func TestUnsupportedMediaType(t *testing.T) {
	h := newTestServer(t)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/tools/base64", strings.NewReader("plain"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("状态码 = %d，期望 415", rec.Code)
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "UNSUPPORTED_MEDIA_TYPE" {
		t.Errorf("error = %+v，期望 UNSUPPORTED_MEDIA_TYPE", env.Error)
	}
}

func TestMissingContentType(t *testing.T) {
	h := newTestServer(t)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/tools/base64", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("状态码 = %d，期望 415", rec.Code)
	}
}

func TestInvalidJSONReturns400(t *testing.T) {
	h := newTestServer(t)
	rec := postJSON(t, h, "/api/v1/tools/base64", `{"operation":`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400", rec.Code)
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "INVALID_JSON" {
		t.Errorf("error = %+v，期望 INVALID_JSON", env.Error)
	}
}

func TestUnknownFieldReturns400(t *testing.T) {
	h := newTestServer(t)
	rec := postJSON(t, h, "/api/v1/tools/base64",
		`{"operation":"encode","input":{"data":"x","encoding":"utf8"},"unknown":1}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400", rec.Code)
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "INVALID_FIELD" {
		t.Errorf("error = %+v，期望 INVALID_FIELD", env.Error)
	}
}

func TestFieldValidationReturns400(t *testing.T) {
	h := newTestServer(t)
	rec := postJSON(t, h, "/api/v1/tools/base64",
		`{"operation":"transcode","input":{"data":"x","encoding":"utf8"}}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400", rec.Code)
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "INVALID_FIELD" {
		t.Errorf("error = %+v，期望 INVALID_FIELD", env.Error)
	}
}

func TestBodyTooLargeReturns413(t *testing.T) {
	// 构造一个全局上限极小的服务器，验证超限路径。
	h := newTestServerWithLimit(t, 128)
	rec := postJSON(t, h, "/api/v1/tools/base64",
		`{"operation":"encode","input":{"data":"`+strings.Repeat("a", 1024)+`","encoding":"utf8"}}`)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("状态码 = %d，期望 413；body=%s", rec.Code, rec.Body.String())
	}
	env := decode(t, rec)
	if env.Error == nil || env.Error.Code != "TOO_LARGE" {
		t.Errorf("error = %+v，期望 TOO_LARGE", env.Error)
	}
}

func TestHealthEndpoints(t *testing.T) {
	h := newTestServer(t)

	rec := get(t, h, "/healthz")
	if rec.Code != http.StatusOK {
		t.Errorf("/healthz 状态码 = %d，期望 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("/healthz 响应 = %q，期望 ok", rec.Body.String())
	}

	rec = get(t, h, "/readyz")
	if rec.Code != http.StatusOK {
		t.Errorf("/readyz 状态码 = %d，期望 200", rec.Code)
	}

	rec = get(t, h, "/api/health")
	if rec.Code != http.StatusOK {
		t.Errorf("/api/health 状态码 = %d，期望 200", rec.Code)
	}
	var health HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("解析健康信息失败: %v", err)
	}
	if health.Tools.Total != 70 {
		t.Errorf("工具总数 = %d，期望 70", health.Tools.Total)
	}
	if health.Tools.Enabled != 70 {
		t.Errorf("启用工具数 = %d，期望 70", health.Tools.Enabled)
	}
}

func TestMetaEndpoint(t *testing.T) {
	h := newTestServer(t)
	rec := get(t, h, "/api/v1/meta")

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	env := decode(t, rec)
	var meta MetaResponse
	if err := json.Unmarshal(env.Data, &meta); err != nil {
		t.Fatalf("解析 meta 失败: %v", err)
	}
	if len(meta.Locales) != 2 {
		t.Errorf("locales = %v，期望 2 项", meta.Locales)
	}
	if meta.DefaultLocale != "zh-CN" {
		t.Errorf("defaultLocale = %q，期望 zh-CN", meta.DefaultLocale)
	}
	if meta.APIVersion != "v1" {
		t.Errorf("apiVersion = %q，期望 v1", meta.APIVersion)
	}
}

func TestInternalEndpointsDisabledByDefault(t *testing.T) {
	h := newTestServer(t)
	rec := get(t, h, "/internal/version")
	if rec.Code != http.StatusNotFound {
		t.Errorf("状态码 = %d，期望 404（内部端点默认关闭）", rec.Code)
	}
}

func TestContractDataFilesExist(t *testing.T) {
	// 契约测试数据必须落档。
	slugs := []string{
		"base64", "number-base",
		"hash-calc", "hmac-gen", "password-gen",
		"symmetric-crypt", "signature", "asymmetric-crypt", "ecdh",
	}
	for _, slug := range slugs {
		path := filepath.Join("..", "..", "testdata", "contract", slug+".json")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("契约数据缺失：%s", path)
		}
	}
}
