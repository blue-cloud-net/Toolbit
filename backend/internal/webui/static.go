// Package webui 提供前端静态资源服务与 SPA 回退。
//
// 行为约定（见 docs/architecture.md §5）：
//
//   - /assets/*  带内容 hash，Cache-Control: public, max-age=31536000, immutable
//   - index.html 及其余文件  Cache-Control: no-cache
//   - 未命中且无扩展名  回退到 index.html（history 路由）
//   - 未命中且有扩展名  404（缺失 chunk 不返回 HTML）
//   - 含 .. 或非法路径  404
//   - /api/* 未命中  JSON 404，绝不回落 index.html
//   - 静态目录缺失  503，并提示检查 TOOLBIT_WEB_ROOT
package webui

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"
)

// 缓存策略常量。
const (
	// ImmutableCacheControl 用于带内容 hash 的构建产物。
	ImmutableCacheControl = "public, max-age=31536000, immutable"
	// NoCacheControl 用于需要回源校验的文件。
	NoCacheControl = "no-cache"
	// AssetsPrefix 是带内容 hash 的资源前缀。
	AssetsPrefix = "/assets/"
	// APIPrefix 是 API 路径前缀。
	APIPrefix = "/api/"
	// IndexFile 是 SPA 入口文件名。
	IndexFile = "index.html"
)

// Handler 提供静态资源服务。
type Handler struct {
	// root 是静态资源根目录。
	root string
	// indexData 是 index.html 的内容，在启动时读入内存。
	//
	// 预读有两个原因：SPA 回退需要频繁返回它，且目录缺失时要在首个请求就给出
	// 明确的 503 而不是逐次读盘。
	indexData []byte
}

// New 创建一个静态资源处理器。
//
// 目录不存在或 index.html 不可读时返回错误，由调用方在启动阶段决定行为。
func New(root string) (*Handler, error) {
	if root == "" {
		return nil, os.ErrNotExist
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}
	indexData, err := os.ReadFile(path.Join(root, IndexFile))
	if err != nil {
		return nil, err
	}
	return &Handler{root: root, indexData: indexData}, nil
}

// ServeHTTP 实现 http.Handler。
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	// API 路径**绝不**回落到 index.html。
	if strings.HasPrefix(r.URL.Path, APIPrefix) {
		writeJSONError(w, http.StatusNotFound, "TOOL_NOT_FOUND", "not found")
		return
	}

	cleaned, ok := cleanPath(r.URL.Path)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "TOOL_NOT_FOUND", "not found")
		return
	}

	// index.html 直接用预读内容返回：既避免 ServeFile 对 index.html 的重定向，
	// 也让 SPA 回退不必反复读盘。
	if cleaned == IndexFile {
		h.serveIndex(w)
		return
	}

	fullPath := path.Join(h.root, cleaned)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		// 未命中：带扩展名按 404 处理（缺失 chunk 不返回 HTML），
		// 无扩展名则回退到 index.html（history 路由）。
		if path.Ext(cleaned) != "" {
			writeJSONError(w, http.StatusNotFound, "TOOL_NOT_FOUND", "not found")
			return
		}
		h.serveIndex(w)
		return
	}

	setCacheControl(w, cleaned)
	file, err := os.Open(fullPath)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "TOOL_NOT_FOUND", "not found")
		return
	}
	defer func() { _ = file.Close() }()
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

// serveIndex 返回 SPA 入口。
func (h *Handler) serveIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", NoCacheControl)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.indexData)
}

// setCacheControl 按路径设置缓存头。
func setCacheControl(w http.ResponseWriter, cleaned string) {
	if strings.HasPrefix("/"+cleaned, AssetsPrefix) {
		w.Header().Set("Cache-Control", ImmutableCacheControl)
		return
	}
	w.Header().Set("Cache-Control", NoCacheControl)
}

// cleanPath 清洗请求路径。
//
// 返回 false 表示路径非法（含 ".." 、"." 或空分段），调用方应返回 404。
//
// 注意：**必须**在清洗**之前**检查分段。path.Clean 会把根部的 ".." 直接消解掉，
// 清洗后再查就永远查不到 "/../etc/passwd" 这类输入。
func cleanPath(p string) (string, bool) {
	if p == "" || p == "/" {
		return IndexFile, true
	}
	if !strings.HasPrefix(p, "/") {
		return "", false
	}
	trimmed := strings.TrimPrefix(p, "/")
	for _, seg := range strings.Split(trimmed, "/") {
		switch seg {
		case "", ".", "..":
			return "", false
		}
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == "/" || strings.HasPrefix(cleaned, "../") {
		return "", false
	}
	return cleaned, true
}

// writeJSONError 写出 JSON 风格的错误响应。
func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    status,
		"message": message,
		"data":    nil,
		"error":   map[string]any{"code": code},
		"meta":    map[string]any{"warnings": []any{}},
	})
}
