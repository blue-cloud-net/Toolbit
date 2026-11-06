package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/httpx"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/system/nginx"
)

// nginxDirectivesSlug 是提供指令目录的工具。
const nginxDirectivesSlug = "nginx-config"

// NginxDirectiveInfo 是一条指令的展示元信息。
type NginxDirectiveInfo struct {
	// ID 是指令唯一标识：模块与指令名。
	ID string `json:"id"`
	// Name 是指令名。
	Name string `json:"name"`
	// Module 是所属模块。
	Module string `json:"module"`
	// Syntax 是语法提示。
	Syntax string `json:"syntax"`
	// Default 是默认值。
	Default string `json:"default"`
	// Contexts 是可用上下文。
	Contexts []string `json:"contexts"`
	// Block 是否是块指令。
	Block bool `json:"block"`
	// Since 是首次出现的版本。
	Since string `json:"since"`
	// Deprecated 是废弃的版本。
	Deprecated string `json:"deprecated"`
	// Removed 是移除的版本。
	Removed string `json:"removed"`
	// CommercialParams 表示该指令含有仅商业订阅可用的参数。
	CommercialParams bool `json:"commercialParams"`
	// Available 表示该指令在请求的目标版本中可用。
	Available bool `json:"available"`
	// DeprecatedNow 表示该指令在请求的目标版本中已废弃。
	DeprecatedNow bool `json:"deprecatedNow"`
	// DocURL 是官方文档链接。
	DocURL string `json:"docUrl"`
	// SummaryZh 是中文简述。
	SummaryZh string `json:"summaryZh"`
	// SummaryEn 是英文简述。
	SummaryEn string `json:"summaryEn"`
}

// NginxModuleInfo 是一个模块的展示元信息。
type NginxModuleInfo struct {
	// Name 是模块标识。
	Name string `json:"name"`
	// Title 是模块标题。
	Title string `json:"title"`
	// BuildOption 是启用该模块的编译开关。
	BuildOption string `json:"buildOption"`
	// DefaultBuild 是否默认编译。
	DefaultBuild bool `json:"defaultBuild"`
	// Commercial 是否属于商业订阅提供的模块。
	Commercial bool `json:"commercial"`
	// DocURL 是官方文档链接。
	DocURL string `json:"docUrl"`
}

// NginxDirectivesResponse 是指令目录端点的响应数据。
type NginxDirectivesResponse struct {
	// Version 是本次响应对应的目标版本。
	Version string `json:"version"`
	// Versions 是可选的目标版本列表。
	Versions []string `json:"versions"`
	// DocsVersion 是快照对应的官方文档版本。
	DocsVersion string `json:"docsVersion"`
	// Source 是官方文档地址。
	Source string `json:"source"`
	// Modules 是模块列表。
	Modules []NginxModuleInfo `json:"modules"`
	// Directives 是指令列表。
	Directives []NginxDirectiveInfo `json:"directives"`
}

// handleNginxDirectives 返回 Nginx 指令目录。
//
// 目录是**只读**的静态快照：不含用户数据，可缓存，
// 因此与工具目录端点一样使用 ETag 与 Cache-Control。
func (s *Server) handleNginxDirectives(c *gin.Context) {
	start := time.Now()

	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		httpx.Fail(c, methodNotAllowed(c.Request.Method))
		return
	}

	slug := c.Param("slug")
	if slug != nginxDirectivesSlug {
		s.fail(c, start, toolerr.Newf(toolerr.CodeToolNotFound, map[string]any{"slug": slug}))
		return
	}

	snapshot := nginx.Load()
	versions := nginx.Versions()
	version := c.Query("version")
	if version == "" && len(versions) > 0 {
		version = versions[0]
	}
	if version != "" && !nginxDirectivesContains(versions, version) {
		s.fail(c, start, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "version",
			"expected": versions,
			"actual":   version,
		}))
		return
	}

	resp := buildNginxDirectivesResponse(snapshot, version, versions)
	etag := nginxDirectivesETag(resp)

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

// buildNginxDirectivesResponse 把目录快照转换为响应 DTO。
func buildNginxDirectivesResponse(snapshot *nginx.Snapshot, version string, versions []string) NginxDirectivesResponse {
	resp := NginxDirectivesResponse{
		Version:     version,
		Versions:    versions,
		DocsVersion: snapshot.DocsVersion,
		Source:      snapshot.Source,
		Modules:     make([]NginxModuleInfo, 0, len(snapshot.Modules)),
		Directives:  make([]NginxDirectiveInfo, 0, len(snapshot.Directives)),
	}

	for _, module := range snapshot.Modules {
		resp.Modules = append(resp.Modules, NginxModuleInfo{
			Name:         module.Name,
			Title:        module.Title,
			BuildOption:  module.BuildOption,
			DefaultBuild: module.DefaultBuild,
			Commercial:   module.Commercial,
			DocURL:       module.DocURL,
		})
	}

	for _, directive := range snapshot.Directives {
		resp.Directives = append(resp.Directives, NginxDirectiveInfo{
			ID:               directive.ID,
			Name:             directive.Name,
			Module:           directive.Module,
			Syntax:           directive.Syntax,
			Default:          directive.Default,
			Contexts:         directive.Contexts,
			Block:            directive.Block,
			Since:            directive.Since,
			Deprecated:       directive.Deprecated,
			Removed:          directive.Removed,
			CommercialParams: directive.CommercialParams,
			Available:        nginx.AvailableIn(directive, version),
			DeprecatedNow:    nginx.DeprecatedIn(directive, version),
			DocURL:           directive.DocURL,
			SummaryZh:        directive.SummaryZh,
			SummaryEn:        directive.SummaryEn,
		})
	}

	return resp
}

// nginxDirectivesETag 计算指令目录内容哈希。
func nginxDirectivesETag(resp NginxDirectivesResponse) string {
	raw, err := json.Marshal(resp)
	if err != nil {
		return `""`
	}
	sum := sha256.Sum256(raw)
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

// nginxDirectivesContains 判断字符串切片是否包含某值。
func nginxDirectivesContains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
