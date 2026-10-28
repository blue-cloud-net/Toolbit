package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestHandler 构造一个带最小前端产物的处理器。
func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, IndexFile), "<!doctype html><title>Toolbit</title>")
	mustWrite(t, filepath.Join(root, "assets", "app.abc123.js"), "console.log(1)")
	mustWrite(t, filepath.Join(root, "favicon.ico"), "icon")

	h, err := New(root)
	if err != nil {
		t.Fatalf("创建静态资源处理器失败: %v", err)
	}
	return h
}

// mustWrite 写入测试夹具文件。
func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}
}

// do 发起一次请求并返回响应。
func do(h http.Handler, method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, target, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAssetsAreImmutable(t *testing.T) {
	h := newTestHandler(t)
	rec := do(h, http.MethodGet, "/assets/app.abc123.js")
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != ImmutableCacheControl {
		t.Errorf("Cache-Control = %q，期望 %q", got, ImmutableCacheControl)
	}
}

func TestIndexIsNoCache(t *testing.T) {
	h := newTestHandler(t)
	rec := do(h, http.MethodGet, "/"+IndexFile)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != NoCacheControl {
		t.Errorf("Cache-Control = %q，期望 %q", got, NoCacheControl)
	}
}

func TestSPAFallback(t *testing.T) {
	h := newTestHandler(t)
	// 无扩展名的未命中路径应回退到 index.html。
	for _, target := range []string{"/tools/base64", "/category/encoding", "/some/deep/route"} {
		rec := do(h, http.MethodGet, target)
		if rec.Code != http.StatusOK {
			t.Errorf("%s 状态码 = %d，期望 200", target, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Toolbit") {
			t.Errorf("%s 未回退到 index.html", target)
		}
		if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
			t.Errorf("%s Content-Type = %q，期望 text/html", target, got)
		}
	}
}

func TestMissingChunkReturns404(t *testing.T) {
	h := newTestHandler(t)
	// 带扩展名的未命中路径必须是 404，不能回退到 HTML。
	rec := do(h, http.MethodGet, "/assets/missing.deadbeef.js")
	if rec.Code != http.StatusNotFound {
		t.Errorf("状态码 = %d，期望 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<!doctype html>") {
		t.Error("缺失 chunk 不应返回 HTML")
	}
}

func TestPathTraversalBlocked(t *testing.T) {
	h := newTestHandler(t)
	for _, target := range []string{"/../etc/passwd", "/assets/../../etc/passwd", "/..%2f..%2fetc/passwd"} {
		rec := do(h, http.MethodGet, target)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s 状态码 = %d，期望 404", target, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "root:") {
			t.Errorf("%s 泄露了系统文件", target)
		}
	}
}

func TestAPIPathNeverFallsBack(t *testing.T) {
	h := newTestHandler(t)
	rec := do(h, http.MethodGet, "/api/v1/not-exist")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("状态码 = %d，期望 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<!doctype html>") {
		t.Error("/api/* 未命中不得回落到 index.html")
	}
	// 必须是 JSON 风格错误。
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if payload["error"] == nil {
		t.Error("错误响应缺少 error 字段")
	}
}

func TestAPIPathNoStore(t *testing.T) {
	h := newTestHandler(t)
	rec := do(h, http.MethodGet, "/api/v1/not-exist")
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q，期望 no-store", got)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestHandler(t)
	rec := do(h, http.MethodPost, "/assets/app.abc123.js")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("状态码 = %d，期望 405", rec.Code)
	}
}

func TestNewFailsWhenRootMissing(t *testing.T) {
	if _, err := New(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("目录缺失时应当返回错误")
	}
	if _, err := New(""); err == nil {
		t.Error("根目录为空时应当返回错误")
	}
}

func TestNewFailsWhenIndexMissing(t *testing.T) {
	root := t.TempDir()
	if _, err := New(root); err == nil {
		t.Error("index.html 缺失时应当返回错误")
	}
}

func TestCleanPath(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"", IndexFile, true},
		{"/", IndexFile, true},
		{"/index.html", "index.html", true},
		{"/assets/app.js", "assets/app.js", true},
		{"/a/b/../c", "", false},
		{"/../etc", "", false},
		{"no-leading-slash", "", false},
	}
	for _, tc := range cases {
		got, ok := cleanPath(tc.in)
		if ok != tc.wantOK {
			t.Errorf("cleanPath(%q) ok = %t，期望 %t", tc.in, ok, tc.wantOK)
			continue
		}
		if ok && got != tc.want {
			t.Errorf("cleanPath(%q) = %q，期望 %q", tc.in, got, tc.want)
		}
	}
}
