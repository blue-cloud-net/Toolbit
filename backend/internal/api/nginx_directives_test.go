package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// nginxDirectivesEnvelope 是端点响应的信封。
type nginxDirectivesEnvelope struct {
	Data nginxDirectivesData `json:"data"`
}

// nginxDirectivesData 是端点响应数据。
type nginxDirectivesData struct {
	Version     string               `json:"version"`
	Versions    []string             `json:"versions"`
	DocsVersion string               `json:"docsVersion"`
	Modules     []NginxModuleInfo    `json:"modules"`
	Directives  []NginxDirectiveInfo `json:"directives"`
}

// TestNginxDirectivesEndpoint 断言指令目录端点返回可用数据。
func TestNginxDirectivesEndpoint(t *testing.T) {
	h := newTestServer(t)
	rec := get(t, h, "/api/v1/tools/nginx-config/directives")

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200；body=%s", rec.Code, rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=60" {
		t.Errorf("Cache-Control = %q", cc)
	}
	if rec.Header().Get("ETag") == "" {
		t.Errorf("缺少 ETag")
	}

	var env nginxDirectivesEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("解析响应失败：%v", err)
	}
	if len(env.Data.Directives) < 100 {
		t.Errorf("指令数 = %d，期望至少 100", len(env.Data.Directives))
	}
	if len(env.Data.Modules) < 10 {
		t.Errorf("模块数 = %d，期望至少 10", len(env.Data.Modules))
	}
	if len(env.Data.Versions) == 0 {
		t.Errorf("缺少可选版本列表")
	}
	if env.Data.DocsVersion == "" {
		t.Errorf("缺少文档版本")
	}

	found := false
	for _, directive := range env.Data.Directives {
		if directive.Name == "proxy_pass" && directive.Module == "ngx_http_proxy_module" {
			found = true
			if directive.Syntax == "" {
				t.Errorf("proxy_pass 缺少语法")
			}
			if len(directive.Contexts) == 0 {
				t.Errorf("proxy_pass 缺少上下文")
			}
			if directive.DocURL == "" {
				t.Errorf("proxy_pass 缺少文档链接")
			}
		}
	}
	if !found {
		t.Errorf("目录中缺少 proxy_pass")
	}
}

// TestNginxDirectivesETag 断言条件请求返回 304。
func TestNginxDirectivesETag(t *testing.T) {
	h := newTestServer(t)
	first := get(t, h, "/api/v1/tools/nginx-config/directives")
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("缺少 ETag")
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet,
		"/api/v1/tools/nginx-config/directives", nil)
	req.Header.Set("If-None-Match", etag)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotModified {
		t.Errorf("状态码 = %d，期望 304", rec.Code)
	}
}

// TestNginxDirectivesUnknownSlug 断言其它工具没有指令目录。
func TestNginxDirectivesUnknownSlug(t *testing.T) {
	h := newTestServer(t)
	rec := get(t, h, "/api/v1/tools/base64/directives")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("状态码 = %d，期望 404；body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "TOOL_NOT_FOUND") {
		t.Errorf("错误码不正确：%s", rec.Body.String())
	}
}

// TestNginxDirectivesInvalidVersion 断言未知版本被拒绝。
func TestNginxDirectivesInvalidVersion(t *testing.T) {
	h := newTestServer(t)
	rec := get(t, h, "/api/v1/tools/nginx-config/directives?version=0.0.1")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400；body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_FIELD") {
		t.Errorf("错误码不正确：%s", rec.Body.String())
	}
}
