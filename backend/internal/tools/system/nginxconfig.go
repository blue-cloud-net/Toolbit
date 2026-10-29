package system

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Nginx 配置模式。
const (
	// NginxModeGlobal 是全局配置。
	NginxModeGlobal = "global"
	// NginxModeSite 是站点配置。
	NginxModeSite = "site"
)

// NginxMode 返回受支持的配置模式（顺序稳定）。
func NginxMode() []string {
	return []string{NginxModeGlobal, NginxModeSite}
}

// NginxConfigRequest 是 nginx-config 工具的请求。
type NginxConfigRequest struct {
	// Mode 是配置模式：global / site；默认 site。
	Mode string `json:"mode"`
	// ServerName 是站点域名；site 模式必填。
	ServerName string `json:"serverName"`
	// ListenPort 是监听端口；默认 80。
	ListenPort int `json:"listenPort"`
	// Root 是站点根目录；site 模式必填。
	Root string `json:"root"`
	// IndexFiles 是索引文件列表。
	IndexFiles []string `json:"indexFiles"`
	// EnableSSL 是否启用 HTTPS。
	EnableSSL bool `json:"enableSsl"`
	// SSLCertificate 是证书路径；启用 SSL 时必填。
	SSLCertificate string `json:"sslCertificate"`
	// SSLPrivateKey 是私钥路径；启用 SSL 时必填。
	SSLPrivateKey string `json:"sslPrivateKey"`
	// EnableProxy 是否启用反向代理。
	EnableProxy bool `json:"enableProxy"`
	// ProxyPass 是代理目标地址。
	ProxyPass string `json:"proxyPass"`
	// ProxyPath 是被代理的路径前缀。
	ProxyPath string `json:"proxyPath"`
	// EnableGzip 是否启用 Gzip 压缩。
	EnableGzip bool `json:"enableGzip"`
	// EnableCORS 是否启用跨域。
	EnableCORS bool `json:"enableCORS"`
	// AllowOrigin 是允许的来源；留空表示允许全部。
	AllowOrigin string `json:"allowOrigin"`
	// EnableWebSocket 是否启用 WebSocket 升级。
	EnableWebSocket bool `json:"enableWebSocket"`
	// EnableLoadBalance 是否启用负载均衡。
	EnableLoadBalance bool `json:"enableLoadBalance"`
	// Upstreams 是后端服务器列表。
	Upstreams []string `json:"upstreams"`
	// ClientMaxBodySize 是请求体上限；默认 10m。
	ClientMaxBodySize string `json:"clientMaxBodySize"`
	// WorkerProcesses 是 worker 进程数；全局模式使用。
	WorkerProcesses int `json:"workerProcesses"`
}

// NginxConfigResult 是 nginx-config 工具的响应。
type NginxConfigResult struct {
	// Mode 是实际使用的配置模式。
	Mode string `json:"mode"`
	// Config 是生成的配置文本。
	Config string `json:"config"`
	// Sections 是生成的配置段标题。
	Sections []string `json:"sections"`
	// Warnings 是配置提示。
	Warnings []string `json:"warnings"`
}

// NginxConfig 生成 Nginx 配置片段。
//
// 这是**纯模板拼装**：不涉及密码学运算，也不出网。
func NginxConfig(_ context.Context, req *NginxConfigRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = NginxModeSite
	}

	sections := []string{}
	warnings := []string{}
	lines := []string{}

	if mode == NginxModeGlobal {
		lines = append(lines, buildNginxGlobal(req)...)
		sections = append(sections, "全局配置")
	} else {
		lines = append(lines, buildNginxSite(req, &warnings)...)
		sections = append(sections, "站点配置")
	}

	if req.EnableGzip {
		lines = append(lines, buildNginxGzip()...)
		sections = append(sections, "Gzip 压缩")
	}
	if req.EnableCORS {
		lines = append(lines, buildNginxCORS(req)...)
		sections = append(sections, "跨域配置")
	}

	return tools.Result{Data: NginxConfigResult{
		Mode:     mode,
		Config:   strings.Join(lines, "\n"),
		Sections: sections,
		Warnings: warnings,
	}}, nil
}

// buildNginxGlobal 生成全局配置段。
func buildNginxGlobal(req *NginxConfigRequest) []string {
	workers := req.WorkerProcesses
	if workers <= 0 {
		workers = 1
	}
	return []string{
		"worker_processes  " + itoa(workers) + ";",
		"",
		"events {",
		"    worker_connections  1024;",
		"}",
		"",
		"http {",
		"    include       mime.types;",
		"    default_type  application/octet-stream;",
		"    sendfile      on;",
		"    keepalive_timeout  65;",
		"",
		"    server {",
		"        listen       80;",
		"        server_name  localhost;",
		"        location / {",
		"            root   html;",
		"            index  index.html;",
		"        }",
		"    }",
		"}",
	}
}

// buildNginxSite 生成站点配置段。
func buildNginxSite(req *NginxConfigRequest, warnings *[]string) []string {
	port := req.ListenPort
	if port <= 0 {
		port = 80
	}

	indexFiles := req.IndexFiles
	if len(indexFiles) == 0 {
		indexFiles = []string{"index.html"}
	}

	bodySize := req.ClientMaxBodySize
	if bodySize == "" {
		bodySize = "10m"
	}

	lines := []string{
		"server {",
		"    listen       " + itoa(port) + ";",
		"    server_name  " + req.ServerName + ";",
		"    client_max_body_size  " + bodySize + ";",
		"",
		"    location / {",
		"        root   " + req.Root + ";",
		"        index  " + strings.Join(indexFiles, " ") + ";",
		"    }",
	}

	if req.EnableSSL {
		lines = append(lines,
			"",
			"    ssl_certificate      "+req.SSLCertificate+";",
			"    ssl_certificate_key  "+req.SSLPrivateKey+";",
			"    ssl_protocols        TLSv1.2 TLSv1.3;",
			"    ssl_ciphers          HIGH:!aNULL:!MD5;",
		)
		if port == 80 {
			*warnings = append(*warnings, "已启用 SSL 但监听端口仍是 80，建议另加 443 监听段或改用 443")
		}
	}

	if req.EnableProxy {
		proxyPath := req.ProxyPath
		if proxyPath == "" {
			proxyPath = "/"
		}
		proxyPass := req.ProxyPass
		if proxyPass == "" {
			proxyPass = "http://127.0.0.1:8080"
		}
		lines = append(lines,
			"",
			"    location "+proxyPath+" {",
			"        proxy_pass         "+proxyPass+";",
			"        proxy_set_header   Host              $host;",
			"        proxy_set_header   X-Real-IP         $remote_addr;",
			"        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;",
			"        proxy_set_header   X-Forwarded-Proto $scheme;",
		)
		if req.EnableWebSocket {
			lines = append(lines,
				"        proxy_http_version  1.1;",
				"        proxy_set_header    Upgrade    $http_upgrade;",
				"        proxy_set_header    Connection \"upgrade\";",
			)
		}
		lines = append(lines, "    }")
	}

	if req.EnableLoadBalance && len(req.Upstreams) > 0 {
		lines = append(lines, "", "    upstream "+req.ServerName+"_backend {")
		for _, upstream := range req.Upstreams {
			lines = append(lines, "        server  "+upstream+";")
		}
		lines = append(lines, "    }")
		*warnings = append(*warnings, "已生成 upstream 段；代理配置需手动把 proxy_pass 指向该 upstream 名称")
	}

	lines = append(lines, "}")
	return lines
}

// buildNginxGzip 生成 Gzip 配置段。
func buildNginxGzip() []string {
	return []string{
		"",
		"gzip  on;",
		"gzip_comp_level  5;",
		"gzip_min_length  1k;",
		"gzip_types  text/plain text/css application/json application/javascript text/xml application/xml;",
	}
}

// buildNginxCORS 生成跨域配置段。
func buildNginxCORS(req *NginxConfigRequest) []string {
	origin := req.AllowOrigin
	if origin == "" {
		origin = "*"
	}
	return []string{
		"",
		"# CORS：允许跨域请求",
		"add_header  Access-Control-Allow-Origin      " + origin + " always;",
		"add_header  Access-Control-Allow-Methods     \"GET, POST, PUT, DELETE, OPTIONS\" always;",
		"add_header  Access-Control-Allow-Headers     \"Content-Type, Authorization\" always;",
	}
}

// Validate 做字段级校验。
func (r *NginxConfigRequest) Validate() error {
	mode := r.Mode
	if mode == "" {
		mode = NginxModeSite
	}
	if !containsString(NginxMode(), mode) {
		return tools.InvalidField("mode", NginxMode(), mode)
	}

	if mode == NginxModeSite {
		if strings.TrimSpace(r.ServerName) == "" {
			return tools.MissingField("serverName")
		}
		if strings.TrimSpace(r.Root) == "" {
			return tools.MissingField("root")
		}
	}

	if r.ListenPort < 0 || r.ListenPort > 65535 {
		return tools.InvalidField("listenPort", []int{0, 65535}, r.ListenPort)
	}
	if r.EnableSSL {
		if strings.TrimSpace(r.SSLCertificate) == "" {
			return tools.MissingField("sslCertificate")
		}
		if strings.TrimSpace(r.SSLPrivateKey) == "" {
			return tools.MissingField("sslPrivateKey")
		}
	}
	if r.EnableProxy && strings.TrimSpace(r.ProxyPass) == "" && !r.EnableLoadBalance {
		// 默认代理目标已在生成阶段兜底，此处不额外报错。
		_ = r.ProxyPass
	}
	if r.WorkerProcesses < 0 || r.WorkerProcesses > 1024 {
		return tools.InvalidField("workerProcesses", []int{0, 1024}, r.WorkerProcesses)
	}
	if len(r.Upstreams) > 64 {
		return tools.InvalidField("upstreams", []int{0, 64}, len(r.Upstreams))
	}
	return nil
}
