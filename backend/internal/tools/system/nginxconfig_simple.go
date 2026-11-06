package system

import (
	"strings"
)

// nginxPresetPlan 是一个预设展开后的具体计划。
type nginxPresetPlan struct {
	// base 是站点骨架：static / spa / proxy / loadbalance / php。
	base string
	// ssl 强制启用 HTTPS。
	ssl bool
	// proxy 强制启用反向代理。
	proxy bool
	// loadBalance 强制启用负载均衡。
	loadBalance bool
	// gzip 强制启用压缩。
	gzip bool
	// cache 强制启用静态缓存。
	cache bool
	// securityHeaders 强制输出安全响应头。
	securityHeaders bool
	// errorPages 强制输出错误页。
	errorPages bool
	// php 强制输出 FastCGI 配置。
	php bool
}

// nginxPresets 返回受支持的简单模式预设。
func nginxPresets() []string {
	return []string{
		"static", "spa", "tls", "proxy", "loadbalance",
		"compress", "security", "errorpage", "php",
	}
}

// nginxTLSProfiles 返回受支持的 TLS 档位。
func nginxTLSProfiles() []string {
	return []string{"modern", "intermediate"}
}

// resolveNginxPreset 把预设名展开成具体计划。
func resolveNginxPreset(preset string) nginxPresetPlan {
	switch preset {
	case "spa":
		return nginxPresetPlan{base: "spa"}
	case "tls":
		return nginxPresetPlan{base: "static", ssl: true}
	case "proxy":
		return nginxPresetPlan{base: "proxy", proxy: true}
	case "loadbalance":
		return nginxPresetPlan{base: "loadbalance", proxy: true, loadBalance: true}
	case "compress":
		return nginxPresetPlan{base: "static", gzip: true, cache: true}
	case "security":
		return nginxPresetPlan{base: "static", securityHeaders: true}
	case "errorpage":
		return nginxPresetPlan{base: "static", errorPages: true}
	case "php":
		return nginxPresetPlan{base: "php", php: true}
	default:
		return nginxPresetPlan{base: "static"}
	}
}

// buildNginxSimpleTree 把简单模式请求展开成配置树。
//
// 展开后的树与高级模式的树**结构完全一致**，
// 因此渲染、诊断与打包逻辑都只有一份。
func buildNginxSimpleTree(req *NginxConfigRequest, diag *[]NginxDiagnostic) []NginxNode {
	requestPreset := req.Preset
	if requestPreset == "" {
		requestPreset = "static"
	}

	nodes := []NginxNode{
		{Type: NginxNodeDirective, Context: NginxContextMain, Name: "worker_processes", Args: nginxWorkerProcesses(req)},
		{
			Type:    NginxNodeBlock,
			Context: NginxContextMain,
			Name:    "events",
			Children: []NginxNode{
				{Type: NginxNodeDirective, Context: NginxContextEvents, Name: "worker_connections", Args: itoa(nginxWorkerConnections(req))},
				{Type: NginxNodeDirective, Context: NginxContextEvents, Name: "multi_accept", Args: "on"},
			},
		},
	}

	httpChildren := []NginxNode{
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "sendfile", Args: "on"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "tcp_nopush", Args: "on"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "keepalive_timeout", Args: "65"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "server_tokens", Args: "off"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "client_max_body_size", Args: nginxClientMaxBodySize(req)},
		buildNginxTypesBlock(),
	}

	// 站点级内容按预设归并：同一类片段只生成一份，避免重复 include。
	gzipNeeded := false
	cacheNeeded := false

	// 先归并全局开关，再生成站点：片段与上游要排在站点之前。
	for i := range req.Sites {
		site := req.Sites[i]
		preset := site.Preset
		if preset == "" {
			preset = requestPreset
		}
		plan := resolveNginxPreset(preset)
		if plan.gzip || site.EnableGzip {
			gzipNeeded = true
		}
		if plan.cache || site.EnableCache {
			cacheNeeded = true
		}
	}

	// 全局片段放在站点之前，便于阅读。
	//
	// 片段单独成文件：多文件模式下是 snippets/ 下的 include，
	// 单文件模式下被原地内联，两种输出内容一致。
	if gzipNeeded {
		httpChildren = append(httpChildren, withNginxFile(buildNginxGzipNodes(), "snippets/gzip.conf")...)
	}
	if cacheNeeded {
		httpChildren = append(httpChildren, withNginxFile(buildNginxCacheNodes(), "snippets/cache.conf")...)
	}

	seenNames := map[string]bool{}
	usedFiles := map[string]bool{}
	upstreams := []NginxNode{}

	for i := range req.Sites {
		site := req.Sites[i]
		preset := site.Preset
		if preset == "" {
			preset = requestPreset
		}
		plan := resolveNginxPreset(preset)

		if seenNames[site.ServerName] {
			appendNginxDiag(diag, NginxDiagWarning, "DUPLICATE_SERVER_NAME", NginxContextServer, site.ID,
				"存在重复的 server_name，Nginx 会按顺序匹配第一个")
		}
		seenNames[site.ServerName] = true

		fileName := nginxSiteFileName(site, usedFiles)
		servers, upstream := buildNginxSiteNodes(site, plan, fileName, diag)
		httpChildren = append(httpChildren, servers...)
		if upstream != nil {
			upstreams = append(upstreams, *upstream)
		}
	}

	if len(upstreams) > 0 {
		httpChildren = append(httpChildren, upstreams...)
	}

	nodes = append(nodes, NginxNode{
		Type:     NginxNodeBlock,
		Context:  NginxContextMain,
		Name:     "http",
		Children: httpChildren,
	})

	return nodes
}

// nginxWorkerProcesses 返回 worker 进程数取值。
func nginxWorkerProcesses(req *NginxConfigRequest) string {
	if strings.TrimSpace(req.WorkerProcesses) == "" {
		return "auto"
	}
	return req.WorkerProcesses
}

// nginxWorkerConnections 返回单个 worker 的最大连接数。
func nginxWorkerConnections(req *NginxConfigRequest) int {
	if req.WorkerConnections <= 0 {
		return 1024
	}
	return req.WorkerConnections
}

// withNginxFile 给一组节点统一设置目标文件。
func withNginxFile(nodes []NginxNode, file string) []NginxNode {
	for i := range nodes {
		nodes[i].File = file
	}
	return nodes
}

// nginxClientMaxBodySize 返回请求体上限取值。
func nginxClientMaxBodySize(req *NginxConfigRequest) string {
	if strings.TrimSpace(req.ClientMaxBodySize) == "" {
		return "10m"
	}
	return req.ClientMaxBodySize
}

// nginxSiteFileName 生成站点配置文件名并去重。
func nginxSiteFileName(site NginxSite, used map[string]bool) string {
	base := sanitizeNginxFileStem(site.FileName)
	if base == "" {
		base = sanitizeNginxFileStem(site.ServerName)
	}
	if base == "" {
		base = "site"
	}
	candidate := base
	for i := 2; used[candidate]; i++ {
		candidate = base + "-" + itoa(i)
	}
	used[candidate] = true
	return "conf.d/" + candidate + ".conf"
}

// buildNginxSiteNodes 生成一个站点的 server 块。
//
// 返回 server 块列表与可选的 upstream 块。
func buildNginxSiteNodes(site NginxSite, plan nginxPresetPlan, fileName string, diag *[]NginxDiagnostic) ([]NginxNode, *NginxNode) {
	port := site.ListenPort
	if port <= 0 {
		port = 80
	}
	useSSL := plan.ssl || site.EnableSSL
	useProxy := plan.proxy || site.EnableProxy
	useLoadBalance := plan.loadBalance || site.EnableLoadBalance
	usePHP := plan.php || site.EnablePHP
	useSecurity := plan.securityHeaders || site.EnableSecurityHeaders
	useErrorPages := plan.errorPages || site.EnableErrorPages

	names := strings.Fields(site.ServerName)
	if len(names) == 0 {
		names = []string{"_"}
	}

	serverChildren := []NginxNode{}
	// 启用 HTTPS 时，80 端口由跳转站点接管，主 server 只监听 443。
	primaryPort := port
	listenArgs := itoa(port)
	if useSSL {
		primaryPort = 443
		listenArgs = "443 ssl"
	}
	serverChildren = append(serverChildren,
		NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "listen", Args: listenArgs})
	if site.ListenIPv6 {
		ipv6Args := "[::]:" + itoa(primaryPort)
		if useSSL {
			ipv6Args = "[::]:443 ssl"
		}
		serverChildren = append(serverChildren,
			NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "listen", Args: ipv6Args})
	}
	serverChildren = append(serverChildren,
		NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "server_name", Args: strings.Join(names, " ")})

	if useSSL {
		if strings.TrimSpace(site.SSLCertificate) == "" || strings.TrimSpace(site.SSLPrivateKey) == "" {
			appendNginxDiag(diag, NginxDiagError, "SSL_MISSING_CERT", NginxContextServer, site.ID,
				"启用 HTTPS 时必须提供证书与私钥路径")
		}
		if port != 443 {
			appendNginxDiag(diag, NginxDiagInfo, "SSL_REDIRECT_PORT", NginxContextServer, site.ID,
				"HTTPS 站点在 80 端口上做跳转，HTTPS 本身监听 443")
		}
		serverChildren = append(serverChildren,
			NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_certificate", Args: site.SSLCertificate},
			NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_certificate_key", Args: site.SSLPrivateKey},
		)
		serverChildren = append(serverChildren, buildNginxTLSNodes(site.TLSProfile)...)
	}

	// 静态类站点才有根目录与索引。
	switch plan.base {
	case "static", "spa", "php":
		if strings.TrimSpace(site.Root) != "" {
			serverChildren = append(serverChildren,
				NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "root", Args: site.Root})
		}
		indexFiles := site.IndexFiles
		if len(indexFiles) == 0 {
			indexFiles = []string{"index.html"}
		}
		serverChildren = append(serverChildren,
			NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "index", Args: strings.Join(indexFiles, " ")})
	}

	if useSecurity {
		serverChildren = append(serverChildren, buildNginxSecurityHeaderNodes(useSSL)...)
	}
	if useErrorPages {
		dir := strings.TrimSpace(site.ErrorPageDir)
		if dir == "" {
			dir = "/var/www/errors"
		}
		// 错误页目录用 alias 挂到 /errors/：与站点根目录并存时
		// 不能复用 root（同一层级只允许一个 root）。
		serverChildren = append(serverChildren,
			NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "error_page", Args: "404 /errors/404.html"},
			NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "error_page", Args: "500 502 503 504 /errors/50x.html"},
			NginxNode{
				Type:    NginxNodeBlock,
				Context: NginxContextServer,
				Name:    "location",
				Args:    "/errors/",
				Children: []NginxNode{
					{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "alias", Args: dir + "/"},
					{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "internal"},
				},
			},
		)
	}

	// location 部分。
	//
	// 一个上下文里**只能有一个** `location /`：启用代理时不再生成
	// 静态兜底 location，否则 Nginx 会报 duplicate location。
	locations := []NginxNode{}
	switch plan.base {
	case "static":
		if !useProxy {
			locations = append(locations, NginxNode{
				Type:    NginxNodeBlock,
				Context: NginxContextServer,
				Name:    "location",
				Args:    "/",
				Children: []NginxNode{
					{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "try_files", Args: "$uri $uri/ =404"},
				},
			})
		}
	case "spa":
		if !useProxy {
			locations = append(locations, NginxNode{
				Type:    NginxNodeBlock,
				Context: NginxContextServer,
				Name:    "location",
				Args:    "/",
				Children: []NginxNode{
					{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "try_files", Args: "$uri $uri/ /index.html"},
				},
			})
		}
	case "php":
		if !useProxy {
			locations = append(locations, NginxNode{
				Type:    NginxNodeBlock,
				Context: NginxContextServer,
				Name:    "location",
				Args:    "/",
				Children: []NginxNode{
					{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "try_files", Args: "$uri $uri/ =404"},
				},
			})
		}
	case "proxy", "loadbalance":
		// 代理类站点的 location 在下面统一生成。
	}

	// FastCGI 段：预设或开关任一开启即生成。
	//
	// PHP 的 location 用正则匹配 `\.php$`，与代理的 `location /` 不冲突；
	// 但静态兜底 location 与代理会冲突，因此上面已按 useProxy 跳过。
	if usePHP {
		socket := strings.TrimSpace(site.PHPSocket)
		if socket == "" {
			socket = "unix:/run/php/php8.2-fpm.sock"
			appendNginxDiag(diag, NginxDiagWarning, "PHP_SOCKET_DEFAULTED", NginxContextServer, site.ID,
				"未填写 FastCGI 地址，已使用常见默认值，请按实际路径调整")
		}
		locations = append(locations, NginxNode{
			Type:    NginxNodeBlock,
			Context: NginxContextServer,
			Name:    "location",
			Args:    `~ \.php$`,
			Children: []NginxNode{
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "try_files", Args: "$fastcgi_script_name =404"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_pass", Args: socket},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_index", Args: "index.php"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_split_path_info", Args: "^(.+\\.php)(/.+)$"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "SCRIPT_FILENAME $document_root$fastcgi_script_name"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "QUERY_STRING $query_string"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "REQUEST_METHOD $request_method"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "CONTENT_TYPE $content_type"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "CONTENT_LENGTH $content_length"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "SCRIPT_NAME $fastcgi_script_name"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "REQUEST_URI $request_uri"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "DOCUMENT_URI $document_uri"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "DOCUMENT_ROOT $document_root"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "SERVER_PROTOCOL $server_protocol"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "GATEWAY_INTERFACE CGI/1.1"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "SERVER_SOFTWARE nginx/$nginx_version"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "REMOTE_ADDR $remote_addr"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "REMOTE_PORT $remote_port"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "SERVER_ADDR $server_addr"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "SERVER_PORT $server_port"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "SERVER_NAME $server_name"},
				{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "fastcgi_param", Args: "REDIRECT_STATUS 200"},
			},
		})
	}

	var upstream *NginxNode
	if useProxy {
		proxyPath := strings.TrimSpace(site.ProxyPath)
		if proxyPath == "" {
			proxyPath = "/"
		}
		target := strings.TrimSpace(site.ProxyPass)
		if useLoadBalance {
			upstreamName := sanitizeNginxName(site.ServerName) + "-backend"
			if upstreamName == "-backend" {
				upstreamName = "backend"
			}
			upstreamChildren := []NginxNode{}
			for _, addr := range site.Upstreams {
				addr = strings.TrimSpace(addr)
				if addr == "" {
					continue
				}
				upstreamChildren = append(upstreamChildren,
					NginxNode{Type: NginxNodeDirective, Context: NginxContextUpstream, Name: "server", Args: addr})
			}
			if len(upstreamChildren) == 0 {
				appendNginxDiag(diag, NginxDiagWarning, "UPSTREAM_EMPTY", NginxContextServer, site.ID,
					"启用了负载均衡但没有填写后端地址")
			}
			upstream = &NginxNode{
				Type:     NginxNodeBlock,
				Context:  NginxContextHTTP,
				Name:     "upstream",
				Args:     upstreamName,
				Children: upstreamChildren,
			}
			target = "http://" + upstreamName
		}
		if target == "" {
			target = "http://127.0.0.1:8080"
			appendNginxDiag(diag, NginxDiagWarning, "PROXY_TARGET_DEFAULTED", NginxContextServer, site.ID,
				"未填写代理目标，已使用常见默认值，请按实际地址调整")
		}
		proxyChildren := []NginxNode{
			{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_pass", Args: target},
			{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_set_header", Args: "Host $host"},
			{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_set_header", Args: "X-Real-IP $remote_addr"},
			{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_set_header", Args: "X-Forwarded-For $proxy_add_x_forwarded_for"},
			{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_set_header", Args: "X-Forwarded-Proto $scheme"},
		}
		if site.EnableWebSocket {
			proxyChildren = append(proxyChildren,
				NginxNode{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_http_version", Args: "1.1"},
				NginxNode{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_set_header", Args: "Upgrade $http_upgrade"},
				NginxNode{Type: NginxNodeDirective, Context: NginxContextLocation, Name: "proxy_set_header", Args: `Connection "upgrade"`},
			)
		}
		locations = append(locations, NginxNode{
			Type:     NginxNodeBlock,
			Context:  NginxContextServer,
			Name:     "location",
			Args:     proxyPath,
			Children: proxyChildren,
		})
	}

	serverChildren = append(serverChildren, locations...)

	server := NginxNode{
		Type:     NginxNodeBlock,
		Context:  NginxContextHTTP,
		Name:     "server",
		File:     fileName,
		Children: serverChildren,
	}

	servers := []NginxNode{server}

	// 启用 HTTPS 时补一个 HTTP → HTTPS 跳转站点。
	//
	// 跳转站点**放在后面**：Nginx 按出现顺序匹配同名虚拟主机，
	// 放在前面会让 443 的 HTTPS 请求也落到跳转块上。
	if useSSL {
		redirect := NginxNode{
			Type:    NginxNodeBlock,
			Context: NginxContextHTTP,
			Name:    "server",
			File:    fileName,
			Children: []NginxNode{
				{Type: NginxNodeDirective, Context: NginxContextServer, Name: "listen", Args: itoa(port)},
				{Type: NginxNodeDirective, Context: NginxContextServer, Name: "listen", Args: "[::]:" + itoa(port)},
				{Type: NginxNodeDirective, Context: NginxContextServer, Name: "server_name", Args: strings.Join(names, " ")},
				{Type: NginxNodeDirective, Context: NginxContextServer, Name: "return", Args: "301 https://$host$request_uri"},
			},
		}
		servers = append(servers, redirect)
	}

	return servers, upstream
}

// buildNginxTLSNodes 生成 TLS 参数。
func buildNginxTLSNodes(profile string) []NginxNode {
	ciphers := "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:" +
		"ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305"
	prefer := "off"
	if profile == "intermediate" {
		ciphers += ":DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384"
		prefer = "on"
	}
	return []NginxNode{
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_protocols", Args: "TLSv1.2 TLSv1.3"},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_ciphers", Args: ciphers},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_prefer_server_ciphers", Args: prefer},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_session_cache", Args: "shared:SSL:10m"},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_session_timeout", Args: "1d"},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "ssl_session_tickets", Args: "off"},
	}
}

// buildNginxSecurityHeaderNodes 生成安全响应头。
func buildNginxSecurityHeaderNodes(useSSL bool) []NginxNode {
	nodes := []NginxNode{
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "add_header", Args: `X-Content-Type-Options "nosniff" always`},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "add_header", Args: `X-Frame-Options "SAMEORIGIN" always`},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "add_header", Args: `Referrer-Policy "strict-origin-when-cross-origin" always`},
		{Type: NginxNodeDirective, Context: NginxContextServer, Name: "add_header", Args: `Permissions-Policy "geolocation=(), microphone=()" always`},
	}
	if useSSL {
		nodes = append(nodes,
			NginxNode{Type: NginxNodeDirective, Context: NginxContextServer, Name: "add_header", Args: `Strict-Transport-Security "max-age=31536000; includeSubDomains" always`})
	}
	return nodes
}

// buildNginxGzipNodes 生成压缩参数。
func buildNginxGzipNodes() []NginxNode {
	return []NginxNode{
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "gzip", Args: "on"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "gzip_vary", Args: "on"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "gzip_proxied", Args: "any"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "gzip_comp_level", Args: "5"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "gzip_min_length", Args: "1k"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "gzip_types", Args: "text/plain text/css text/xml application/json application/javascript application/rss+xml image/svg+xml"},
	}
}

// buildNginxCacheNodes 生成静态缓存参数。
func buildNginxCacheNodes() []NginxNode {
	return []NginxNode{
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "expires", Args: "7d"},
		{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "add_header", Args: `Cache-Control "public, immutable" always`},
	}
}

// buildNginxTypesBlock 生成 MIME 类型映射。
//
// 这是一份**精简**的自有映射：只覆盖常见扩展名，
// 单文件模式下内联使用，多文件模式下写入 mime.types。
func buildNginxTypesBlock() NginxNode {
	entries := []string{
		"text/html html htm",
		"text/css css",
		"text/plain txt",
		"text/xml xml",
		"application/javascript js mjs",
		"application/json json",
		"application/rss+xml rss",
		"application/atom+xml atom",
		"application/pdf pdf",
		"application/zip zip",
		"application/gzip gz",
		"application/octet-stream bin exe dll",
		"image/gif gif",
		"image/jpeg jpg jpeg",
		"image/png png",
		"image/svg+xml svg svgz",
		"image/webp webp",
		"image/avif avif",
		"image/x-icon ico",
		"font/woff2 woff2",
		"font/ttf ttf",
		"font/otf otf",
		"font/woff woff",
		"application/vnd.ms-fontobject eot",
		"audio/mpeg mp3",
		"audio/ogg ogg",
		"video/mp4 mp4",
		"video/webm webm",
		"application/wasm wasm",
	}
	children := make([]NginxNode, 0, len(entries))
	for _, entry := range entries {
		// MIME 类型映射的每一项都必须以分号结尾。
		children = append(children, NginxNode{
			Type:    NginxNodeRaw,
			Context: NginxContextTypes,
			Body:    entry + ";",
		})
	}
	return NginxNode{
		Type:     NginxNodeBlock,
		Context:  NginxContextHTTP,
		Name:     "types",
		Children: children,
	}
}
