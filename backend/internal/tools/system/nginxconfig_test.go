package system

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// nginxResult 生成配置并取出结果。
func nginxResult(t *testing.T, req *NginxConfigRequest) NginxConfigResult {
	t.Helper()
	result, err := NginxConfig(t.Context(), req)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	data, ok := result.Data.(NginxConfigResult)
	if !ok {
		t.Fatalf("结果类型断言失败：%T", result.Data)
	}
	return data
}

// nginxFileByPath 按路径取文件。
func nginxFileByPath(files []NginxFile, path string) (NginxFile, bool) {
	for _, file := range files {
		if file.Path == path {
			return file, true
		}
	}
	return NginxFile{}, false
}

// TestNginxConfigSimpleBundle 断言简单模式生成多文件配置。
func TestNginxConfigSimpleBundle(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Preset: "static",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			ListenPort: 80,
			Root:       "/var/www/html",
		}},
	})

	if data.Output != NginxOutputBundle {
		t.Errorf("输出形态 = %q", data.Output)
	}
	if len(data.Files) < 3 {
		t.Fatalf("文件数 = %d，期望至少 3：%v", len(data.Files), data.Files)
	}

	main, ok := nginxFileByPath(data.Files, "nginx.conf")
	if !ok {
		t.Fatalf("缺少 nginx.conf：%v", data.Files)
	}
	if !strings.Contains(main.Content, "include conf.d/example.com.conf;") {
		t.Errorf("主配置应引入站点文件：\n%s", main.Content)
	}
	if !strings.Contains(main.Content, "include mime.types;") {
		t.Errorf("主配置应引入 mime.types：\n%s", main.Content)
	}

	site, ok := nginxFileByPath(data.Files, "conf.d/example.com.conf")
	if !ok {
		t.Fatalf("缺少站点文件：%v", data.Files)
	}
	if !strings.Contains(site.Content, "server_name example.com;") {
		t.Errorf("站点文件应含域名：\n%s", site.Content)
	}
	if !strings.Contains(site.Content, "try_files $uri $uri/ =404;") {
		t.Errorf("站点文件应含 try_files：\n%s", site.Content)
	}

	mime, ok := nginxFileByPath(data.Files, "mime.types")
	if !ok {
		t.Fatalf("缺少 mime.types：%v", data.Files)
	}
	// mime.types 自身就是 `types { ... }` 块：它被 include 进 http 上下文。
	if !strings.Contains(mime.Content, "types {") {
		t.Errorf("mime.types 应包含 types 包装：\n%s", mime.Content)
	}
	if !strings.Contains(mime.Content, "text/html html htm;") {
		t.Errorf("mime.types 应含常见类型且以分号结尾：\n%s", mime.Content)
	}
}

// TestNginxConfigIncludeNotDuplicated 断言同一片段文件只被引入一次。
func TestNginxConfigIncludeNotDuplicated(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Preset: "compress",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			Root:       "/var/www/html",
		}},
	})

	main, ok := nginxFileByPath(data.Files, "nginx.conf")
	if !ok {
		t.Fatalf("缺少 nginx.conf")
	}
	if count := strings.Count(main.Content, "include snippets/gzip.conf;"); count != 1 {
		t.Errorf("gzip 片段引入次数 = %d，期望 1：\n%s", count, main.Content)
	}
	if count := strings.Count(main.Content, "include snippets/cache.conf;"); count != 1 {
		t.Errorf("cache 片段引入次数 = %d，期望 1：\n%s", count, main.Content)
	}
}

// TestNginxConfigTLSUsesDistinctPorts 断言 HTTPS 站点的端口分配。
func TestNginxConfigTLSUsesDistinctPorts(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "tls",
		Sites: []NginxSite{{
			ID:             "s1",
			ServerName:     "secure.example.com",
			Root:           "/var/www/secure",
			EnableSSL:      true,
			SSLCertificate: "/etc/ssl/certs/secure.pem",
			SSLPrivateKey:  "/etc/ssl/private/secure.key",
		}},
	})

	content := data.Files[0].Content
	if !strings.Contains(content, "listen 443 ssl;") {
		t.Errorf("HTTPS 站点应监听 443：\n%s", content)
	}
	if !strings.Contains(content, "listen 80;") {
		t.Errorf("应保留 80 端口跳转：\n%s", content)
	}
	// 主 server 不能同时占用 80 与 443，否则跳转块永远匹配不到。
	if strings.Contains(content, "listen 80;\n        server_name secure.example.com;\n        listen 443 ssl;") {
		t.Errorf("同一 server 块不应同时监听 80 与 443：\n%s", content)
	}
	// 跳转块必须排在主 server 之后。
	if strings.Index(content, "return 301 https://") < strings.Index(content, "ssl_certificate") {
		t.Errorf("跳转块应排在主 server 之后：\n%s", content)
	}
	if data.Summary.SiteCount != 1 {
		t.Errorf("站点数 = %d，期望 1（跳转块不计入）", data.Summary.SiteCount)
	}
}

// TestNginxConfigUpstreamNameIsValid 断言上游组名不含点号。
func TestNginxConfigUpstreamNameIsValid(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "loadbalance",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "api.example.com",
			Upstreams:  []string{"127.0.0.1:8081"},
		}},
	})

	content := data.Files[0].Content
	if !strings.Contains(content, "upstream api-example-com-backend {") {
		t.Errorf("上游组名应去掉点号：\n%s", content)
	}
	if !strings.Contains(content, "proxy_pass http://api-example-com-backend;") {
		t.Errorf("proxy_pass 应指向上游组名：\n%s", content)
	}
}

// TestNginxConfigCoreDirectivesKnown 断言核心模块指令已被目录收录。
func TestNginxConfigCoreDirectivesKnown(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "static",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			Root:       "/var/www/html",
		}},
	})

	for _, diag := range data.Diagnostics {
		if diag.Code == "UNKNOWN_DIRECTIVE" {
			t.Errorf("核心指令不应报未知：%s", diag.Message)
		}
	}
}

// TestNginxConfigMimeEntriesTerminated 断言 MIME 映射每项以分号结尾。
//
// 缺分号会让 Nginx 报 "unknown directive" —— 只有真实 nginx -t 才能发现。
func TestNginxConfigMimeEntriesTerminated(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "static",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			Root:       "/var/www/html",
		}},
	})

	content := data.Files[0].Content
	typesStart := strings.Index(content, "types {")
	if typesStart < 0 {
		t.Fatalf("缺少 types 块：\n%s", content)
	}
	typesEnd := strings.Index(content[typesStart:], "}")
	if typesEnd < 0 {
		t.Fatalf("types 块未闭合：\n%s", content)
	}
	block := content[typesStart : typesStart+typesEnd]

	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "types {" {
			continue
		}
		if !strings.HasSuffix(trimmed, ";") {
			t.Errorf("MIME 映射项应以分号结尾：%q", trimmed)
		}
	}
}

// TestNginxConfigNoDuplicateLocation 断言同一 server 内不重复定义 location /。
func TestNginxConfigNoDuplicateLocation(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "static",
		Sites: []NginxSite{{
			ID:          "s1",
			ServerName:  "example.com",
			Root:        "/var/www/html",
			EnableProxy: true,
			ProxyPass:   "http://127.0.0.1:3000",
		}},
	})

	content := data.Files[0].Content
	if count := strings.Count(content, "location / {"); count > 1 {
		t.Errorf("location / 出现 %d 次，Nginx 会报 duplicate location：\n%s", count, content)
	}
}

// TestNginxConfigSelfContained 断言单文件模式不依赖外部文件。
func TestNginxConfigSelfContained(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "php",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "php.example.com",
			Root:       "/var/www/php",
			PHPSocket:  "unix:/run/php/php-fpm.sock",
		}},
	})

	content := data.Files[0].Content
	if strings.Contains(content, "include fastcgi_params;") {
		t.Errorf("单文件模式不应依赖 fastcgi_params：\n%s", content)
	}
	if !strings.Contains(content, "fastcgi_param SCRIPT_FILENAME") {
		t.Errorf("单文件模式应内联 FastCGI 参数：\n%s", content)
	}
}

// TestNginxConfigErrorPageNoDuplicateRoot 断言错误页不重复定义 root。
func TestNginxConfigErrorPageNoDuplicateRoot(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "errorpage",
		Sites: []NginxSite{{
			ID:           "s1",
			ServerName:   "err.example.com",
			Root:         "/var/www/err",
			ErrorPageDir: "/var/www/errors",
		}},
	})

	content := data.Files[0].Content
	if count := strings.Count(content, "root "); count > 1 {
		t.Errorf("root 指令出现 %d 次，Nginx 会报 duplicate directive：\n%s", count, content)
	}
	if !strings.Contains(content, "error_page 404 /errors/404.html;") {
		t.Errorf("应输出错误页映射：\n%s", content)
	}
}

// TestNginxConfigSingleIsSelfContained 断言单文件模式不依赖外部文件。
func TestNginxConfigSingleIsSelfContained(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Output: NginxOutputSingle,
		Preset: "static",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			Root:       "/var/www/html",
		}},
	})

	if len(data.Files) != 1 {
		t.Fatalf("文件数 = %d，期望 1：%v", len(data.Files), data.Files)
	}
	content := data.Files[0].Content
	if data.Files[0].Path != "nginx.conf" {
		t.Errorf("文件路径 = %q", data.Files[0].Path)
	}
	if strings.Contains(content, "include conf.d/") {
		t.Errorf("单文件模式不应引入站点文件：\n%s", content)
	}
	if strings.Contains(content, "include mime.types;") {
		t.Errorf("单文件模式不应引入 mime.types：\n%s", content)
	}
	if !strings.Contains(content, "types {") {
		t.Errorf("单文件模式应内联 types 块：\n%s", content)
	}
	if !strings.Contains(content, "server_name example.com;") {
		t.Errorf("单文件模式应内联站点配置：\n%s", content)
	}
}

// TestNginxConfigSameTreeSameSemantics 断言两种输出基于同一配置树。
func TestNginxConfigSameTreeSameSemantics(t *testing.T) {
	req := &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Preset: "compress",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			Root:       "/var/www/html",
		}},
	}

	bundle := nginxResult(t, req)
	single := nginxResult(t, &NginxConfigRequest{
		Mode:   req.Mode,
		Output: NginxOutputSingle,
		Preset: req.Preset,
		Sites:  req.Sites,
	})

	// 压缩参数在两种输出里都应出现，只是位置不同。
	bundleGzip, ok := nginxFileByPath(bundle.Files, "snippets/gzip.conf")
	if !ok {
		t.Fatalf("多文件模式缺少 gzip 片段：%v", bundle.Files)
	}
	if !strings.Contains(bundleGzip.Content, "gzip_comp_level 5;") {
		t.Errorf("gzip 片段内容异常：\n%s", bundleGzip.Content)
	}
	singleMain, ok := nginxFileByPath(single.Files, "nginx.conf")
	if !ok {
		t.Fatalf("单文件模式缺少主配置")
	}
	if !strings.Contains(singleMain.Content, "gzip_comp_level 5;") {
		t.Errorf("单文件模式应内联 gzip 参数：\n%s", singleMain.Content)
	}
	if strings.Contains(singleMain.Content, "include snippets/gzip.conf;") {
		t.Errorf("单文件模式不应引入外部片段：\n%s", singleMain.Content)
	}
}

// TestNginxConfigAdvancedTree 断言高级模式按配置树渲染。
func TestNginxConfigAdvancedTree(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeAdvanced,
		Output: NginxOutputSingle,
		Nodes: []NginxNode{
			{Type: NginxNodeDirective, Context: NginxContextMain, Name: "worker_processes", Args: "4"},
			{
				Type:    NginxNodeBlock,
				Context: NginxContextMain,
				Name:    "http",
				Children: []NginxNode{
					{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "sendfile", Args: "on"},
					{
						Type:    NginxNodeBlock,
						Context: NginxContextHTTP,
						Name:    "server",
						Children: []NginxNode{
							{Type: NginxNodeDirective, Context: NginxContextServer, Name: "server_name", Args: "a.example.com"},
							{Type: NginxNodeDirective, Context: NginxContextServer, Name: "return", Args: "204"},
						},
					},
				},
			},
		},
	})

	content := data.Files[0].Content
	if !strings.Contains(content, "worker_processes 4;") {
		t.Errorf("应含全局指令：\n%s", content)
	}
	if !strings.Contains(content, "server_name a.example.com;") {
		t.Errorf("应含站点指令：\n%s", content)
	}
	if !strings.Contains(content, "http {") {
		t.Errorf("应含 http 块：\n%s", content)
	}
}

// TestNginxConfigAdvancedSplitsFiles 断言高级模式按站点拆文件。
func TestNginxConfigAdvancedSplitsFiles(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode: NginxModeAdvanced,
		Nodes: []NginxNode{
			{
				Type:    NginxNodeBlock,
				Context: NginxContextMain,
				Name:    "http",
				Children: []NginxNode{
					{
						Type:    NginxNodeBlock,
						Context: NginxContextHTTP,
						Name:    "server",
						Children: []NginxNode{
							{Type: NginxNodeDirective, Context: NginxContextServer, Name: "server_name", Args: "a.example.com"},
						},
					},
				},
			},
		},
	})

	site, ok := nginxFileByPath(data.Files, "conf.d/a.example.com.conf")
	if !ok {
		t.Fatalf("缺少站点文件：%v", data.Files)
	}
	if !strings.HasPrefix(site.Content, "server {") {
		t.Errorf("站点文件应以 server 开头：\n%s", site.Content)
	}
	main, ok := nginxFileByPath(data.Files, "nginx.conf")
	if !ok {
		t.Fatalf("缺少主配置")
	}
	if !strings.Contains(main.Content, "include conf.d/a.example.com.conf;") {
		t.Errorf("主配置应引入站点文件：\n%s", main.Content)
	}
}

// TestNginxConfigArchive 断言 ZIP 产物可解压且与文件列表一致。
func TestNginxConfigArchive(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:          NginxModeSimple,
		ResponseShape: NginxShapeBundle,
		Preset:        "static",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			Root:       "/var/www/html",
		}},
	})

	if data.Archive == nil {
		t.Fatalf("缺少打包产物")
	}
	if data.Archive.Entries != len(data.Files) {
		t.Errorf("条目数 = %d，文件数 = %d", data.Archive.Entries, len(data.Files))
	}

	raw, err := base64.StdEncoding.DecodeString(data.Archive.Base64)
	if err != nil {
		t.Fatalf("Base64 解码失败：%v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("ZIP 解析失败：%v", err)
	}

	entries := map[string]string{}
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("打开条目失败：%v", err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("读取条目失败：%v", err)
		}
		entries[file.Name] = string(content)
	}

	for _, file := range data.Files {
		got, ok := entries[file.Path]
		if !ok {
			t.Errorf("ZIP 缺少条目 %q", file.Path)
			continue
		}
		if got != file.Content {
			t.Errorf("条目 %q 内容与文件列表不一致", file.Path)
		}
	}
}

// TestNginxConfigDeterministic 断言同样输入产出同样结果。
func TestNginxConfigDeterministic(t *testing.T) {
	req := &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Preset: "security",
		Sites: []NginxSite{
			{ID: "a", ServerName: "a.example.com", Root: "/var/www/a"},
			{ID: "b", ServerName: "b.example.com", Root: "/var/www/b"},
		},
	}

	first := nginxResult(t, req)
	second := nginxResult(t, req)
	if len(first.Files) != len(second.Files) {
		t.Fatalf("文件数不一致：%d vs %d", len(first.Files), len(second.Files))
	}
	for i := range first.Files {
		if first.Files[i].Path != second.Files[i].Path {
			t.Fatalf("文件路径不一致：%q vs %q", first.Files[i].Path, second.Files[i].Path)
		}
		if first.Files[i].Content != second.Files[i].Content {
			t.Errorf("文件 %q 内容不一致", first.Files[i].Path)
		}
	}
}

// TestNginxConfigDiagnostics 断言诊断能发现上下文与版本问题。
func TestNginxConfigDiagnostics(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeAdvanced,
		Output: NginxOutputSingle,
		Nodes: []NginxNode{
			{Type: NginxNodeDirective, Context: NginxContextMain, Name: "proxy_pass", Args: "http://127.0.0.1:8080"},
			{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "not_a_real_directive", Args: "1"},
		},
	})

	codes := map[string]bool{}
	for _, diag := range data.Diagnostics {
		codes[diag.Code] = true
	}
	if !codes["CONTEXT_MISMATCH"] {
		t.Errorf("应报告上下文不匹配：%v", data.Diagnostics)
	}
	if !codes["UNKNOWN_DIRECTIVE"] {
		t.Errorf("应报告未知指令：%v", data.Diagnostics)
	}
}

// TestNginxConfigBlockMismatch 断言块被写成指令时报错。
//
// `location /x;` 会让 Nginx 直接启动失败，属于必须提示的错误。
func TestNginxConfigBlockMismatch(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeAdvanced,
		Output: NginxOutputSingle,
		Nodes: []NginxNode{
			{Type: NginxNodeDirective, Context: NginxContextHTTP, Name: "location", Args: "/x"},
		},
	})

	found := false
	for _, diag := range data.Diagnostics {
		if diag.Code == "BLOCK_REQUIRED" && diag.Level == NginxDiagError {
			found = true
		}
	}
	if !found {
		t.Errorf("应报告配置块被写成指令：%v", data.Diagnostics)
	}
}

// TestNginxConfigSimpleDiagnostics 断言简单模式的语义提示。
func TestNginxConfigSimpleDiagnostics(t *testing.T) {
	data := nginxResult(t, &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Preset: "static",
		Sites: []NginxSite{
			{ID: "a", ServerName: "a.example.com", Root: "/var/www/a"},
			{ID: "b", ServerName: "a.example.com", Root: "/var/www/b"},
		},
	})

	found := false
	for _, diag := range data.Diagnostics {
		if diag.Code == "DUPLICATE_SERVER_NAME" {
			found = true
		}
	}
	if !found {
		t.Errorf("应报告重复域名：%v", data.Diagnostics)
	}
}

// TestNginxConfigInvalidMode 断言未知模式返回 INVALID_FIELD。
func TestNginxConfigInvalidMode(t *testing.T) {
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{Mode: "stream"})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestNginxConfigInvalidVersion 断言未知版本返回 INVALID_FIELD。
func TestNginxConfigInvalidVersion(t *testing.T) {
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{
		Mode:          NginxModeSimple,
		TargetVersion: "0.0.1",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestNginxConfigUnsafeFileName 断言非法文件名返回 INVALID_FIELD。
func TestNginxConfigUnsafeFileName(t *testing.T) {
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Preset: "static",
		Sites: []NginxSite{{
			ID:         "s1",
			ServerName: "example.com",
			FileName:   "../evil",
			Root:       "/var/www/html",
		}},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestNginxConfigMissingServerName 断言缺少域名返回 MISSING_FIELD。
func TestNginxConfigMissingServerName(t *testing.T) {
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{
		Mode:   NginxModeSimple,
		Preset: "static",
		Sites:  []NginxSite{{ID: "s1", Root: "/var/www/html"}},
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestNginxConfigTooManyNodes 断言节点数超限返回 INVALID_FIELD。
func TestNginxConfigTooManyNodes(t *testing.T) {
	nodes := make([]NginxNode, 0, maxNginxNodes+1)
	for i := 0; i <= maxNginxNodes; i++ {
		nodes = append(nodes, NginxNode{
			Type:    NginxNodeDirective,
			Context: NginxContextMain,
			Name:    "sendfile",
			Args:    "on",
		})
	}
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{
		Mode:  NginxModeAdvanced,
		Nodes: nodes,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestNginxConfigInvalidNodeType 断言非法节点类型返回 INVALID_FIELD。
func TestNginxConfigInvalidNodeType(t *testing.T) {
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{
		Mode: NginxModeAdvanced,
		Nodes: []NginxNode{
			{Type: "weird", Context: NginxContextMain, Name: "sendfile"},
		},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
