package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/catalog"
	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// CatalogResponse 是 GET /api/v1/tools 的响应数据。
type CatalogResponse struct {
	// Version 是目录结构版本。
	Version int `json:"version"`
	// Categories 是全部分类。
	Categories []CategoryResponse `json:"categories"`
	// Tools 是全部工具。
	Tools []ToolResponse `json:"tools"`
}

// CategoryResponse 是分类的展示元信息。
type CategoryResponse struct {
	// Slug 是分类标识。
	Slug string `json:"slug"`
	// Name 是中文名称。
	Name string `json:"name"`
	// NameEn 是英文名称。
	NameEn string `json:"nameEn"`
	// Order 是分类内排序。
	Order int `json:"order"`
	// Icon 是图标名。
	Icon string `json:"icon"`
}

// ToolResponse 是工具的展示元信息。
type ToolResponse struct {
	// Slug 是工具标识。
	Slug string `json:"slug"`
	// Category 是所属分类。
	Category string `json:"category"`
	// Order 是分类内排序。
	Order int `json:"order"`
	// Name 是中文名称。
	Name string `json:"name"`
	// NameEn 是英文名称。
	NameEn string `json:"nameEn"`
	// Enabled 是否启用。
	Enabled bool `json:"enabled"`
	// Transport 是传输方式。
	Transport catalog.Transport `json:"transport"`
	// RateClass 是限流档位。
	RateClass catalog.RateClass `json:"rateClass"`
	// MaxBodyBytes 是请求体上限。
	MaxBodyBytes int64 `json:"maxBodyBytes"`
	// TimeoutMs 是服务端处理超时。
	TimeoutMs int `json:"timeoutMs"`
	// Sensitive 是否涉及敏感材料。
	Sensitive bool `json:"sensitive"`
	// RequiresNetwork 是否需要出网。
	RequiresNetwork bool `json:"requiresNetwork"`
	// AutoRun 是否在参数变更后自动触发。
	AutoRun bool `json:"autoRun"`
	// Wide 是否在宽屏下双栏布局。
	Wide bool `json:"wide"`
	// I18nKey 是前端词条命名空间。
	I18nKey string `json:"i18nKey"`
}

// buildCatalogResponse 把目录转换为响应 DTO。
func buildCatalogResponse(c *catalog.Catalog) CatalogResponse {
	resp := CatalogResponse{
		Version:    c.Version,
		Categories: make([]CategoryResponse, 0, len(c.Categories)),
		Tools:      make([]ToolResponse, 0, len(c.Tools)),
	}
	for _, cat := range c.Categories {
		resp.Categories = append(resp.Categories, CategoryResponse{
			Slug:   cat.Slug,
			Name:   cat.Name,
			NameEn: cat.NameEn,
			Order:  cat.Order,
			Icon:   cat.Icon,
		})
	}
	for _, t := range c.Tools {
		resp.Tools = append(resp.Tools, ToolResponse{
			Slug:            t.Slug,
			Category:        t.Category,
			Order:           t.Order,
			Name:            t.Name,
			NameEn:          t.NameEn,
			Enabled:         t.Enabled,
			Transport:       t.Transport,
			RateClass:       t.RateClass,
			MaxBodyBytes:    t.MaxBodyBytes,
			TimeoutMs:       t.TimeoutMs,
			Sensitive:       t.Sensitive,
			RequiresNetwork: t.RequiresNetwork,
			AutoRun:         t.AutoRun,
			Wide:            t.Wide,
			I18nKey:         t.I18nKey,
		})
	}
	return resp
}

// handleCatalog 返回工具目录。
//
// 带 ETag（内容哈希），客户端**应当**用 If-None-Match 协商。
// 目录不含用户数据，可缓存：Cache-Control: public, max-age=60。
func (s *Server) handleCatalog(c *gin.Context) {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		httpx.Fail(c, methodNotAllowed(c.Request.Method))
		return
	}

	resp := buildCatalogResponse(s.catalog)
	etag := catalogETag(resp)

	if match := c.GetHeader("If-None-Match"); match != "" && etagMatches(match, etag) {
		c.Header("ETag", etag)
		c.Header("Cache-Control", "public, max-age=60")
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=60")
	httpx.Success(c, resp, httpx.Meta{
		DurationMs: 0,
		Warnings:   []toolerr.Warning{},
	})
}

// catalogETag 计算目录内容哈希。
func catalogETag(resp CatalogResponse) string {
	sum := sha256.Sum256(catalogBytes(resp))
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

// etagMatches 判断 If-None-Match 是否命中。
func etagMatches(header, etag string) bool {
	for _, candidate := range splitETags(header) {
		if candidate == "*" || candidate == etag {
			return true
		}
	}
	return false
}

// splitETags 拆分 If-None-Match 头。
func splitETags(header string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(header); i++ {
		if i == len(header) || header[i] == ',' {
			tag := trimASCII(header[start:i])
			if tag != "" {
				out = append(out, tag)
			}
			start = i + 1
		}
	}
	return out
}

// trimASCII 去除首尾空白。
func trimASCII(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// catalogBytes 返回目录的规范化字节表示，用于计算 ETag。
func catalogBytes(resp CatalogResponse) []byte {
	// 目录内容稳定，用 JSON 序列化即可得到确定的哈希输入。
	b, err := json.Marshal(resp)
	if err != nil {
		return nil
	}
	return b
}
