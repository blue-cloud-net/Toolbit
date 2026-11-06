package system

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/system/nginx"
)

// Nginx 配置模式。
const (
	// NginxModeSimple 是简单模式：只暴露常用场景字段。
	NginxModeSimple = "simple"
	// NginxModeAdvanced 是高级模式：直接编辑配置树。
	NginxModeAdvanced = "advanced"
)

// Nginx 输出形态。
const (
	// NginxOutputBundle 输出多文件配置包。
	NginxOutputBundle = "bundle"
	// NginxOutputSingle 输出单个完整 nginx.conf。
	NginxOutputSingle = "single"
)

// Nginx 响应形态。
const (
	// NginxShapeFiles 只返回文件内容。
	NginxShapeFiles = "files"
	// NginxShapeBundle 额外返回 ZIP 打包内容。
	NginxShapeBundle = "bundle"
)

// Nginx 节点类型。
const (
	// NginxNodeDirective 是单条指令。
	NginxNodeDirective = "directive"
	// NginxNodeBlock 是带子节点的块。
	NginxNodeBlock = "block"
	// NginxNodeRaw 是原样片段。
	NginxNodeRaw = "raw"
)

// Nginx 配置上下文。
const (
	NginxContextMain        = "main"
	NginxContextEvents      = "events"
	NginxContextHTTP        = "http"
	NginxContextServer      = "server"
	NginxContextLocation    = "location"
	NginxContextUpstream    = "upstream"
	NginxContextStream      = "stream"
	NginxContextMail        = "mail"
	NginxContextIf          = "if"
	NginxContextLimitExcept = "limit_except"
	NginxContextTypes       = "types"
)

// Nginx 诊断级别。
const (
	// NginxDiagInfo 是提示。
	NginxDiagInfo = "info"
	// NginxDiagWarning 是警告。
	NginxDiagWarning = "warning"
	// NginxDiagError 是错误。
	NginxDiagError = "error"
)

// 生成规模上限。
//
// 上限用于避免超大配置把响应撑爆：文件数、单文件大小与总大小都有界，
// 超出时**明确报错**，不静默截断。
const (
	// maxNginxNodes 是配置树节点总数上限。
	maxNginxNodes = 2000
	// maxNginxDepth 是配置树嵌套深度上限。
	maxNginxDepth = 16
	// maxNginxFiles 是输出文件数上限。
	maxNginxFiles = 32
	// maxNginxFileSize 是单个文件大小上限（字节）。
	maxNginxFileSize = 128 * 1024
	// maxNginxTotalSize 是全部文件大小之和上限（字节）。
	maxNginxTotalSize = 1024 * 1024
	// maxNginxSites 是站点数上限。
	maxNginxSites = 32
	// maxNginxUpstreams 是单站点上游服务器数上限。
	maxNginxUpstreams = 64
	// maxNginxNodesArgs 是单个节点参数长度上限。
	maxNginxNodesArgs = 4096
)

// NginxConfigRequest 是 nginx-config 工具的请求。
type NginxConfigRequest struct {
	// Mode 是配置模式：simple / advanced；默认 simple。
	Mode string `json:"mode"`
	// TargetVersion 是目标 Nginx 版本；留空表示不按版本过滤。
	TargetVersion string `json:"targetVersion"`
	// Output 是输出形态：bundle（多文件）/ single（单个完整配置）；默认 bundle。
	Output string `json:"output"`
	// ResponseShape 是响应形态：files / bundle；默认 files。
	//
	// bundle 形态额外返回 ZIP 打包内容，用于「打包下载」。
	ResponseShape string `json:"responseShape"`
	// Preset 是简单模式的场景预设。
	Preset string `json:"preset"`
	// WorkerProcesses 是 worker 进程数；留空表示 auto。
	WorkerProcesses string `json:"workerProcesses"`
	// WorkerConnections 是单个 worker 的最大连接数；默认 1024。
	WorkerConnections int `json:"workerConnections"`
	// ClientMaxBodySize 是请求体上限；默认 10m。
	ClientMaxBodySize string `json:"clientMaxBodySize"`
	// Sites 是简单模式的站点列表。
	Sites []NginxSite `json:"sites"`
	// Nodes 是高级模式的配置树；根节点属于 main 上下文。
	Nodes []NginxNode `json:"nodes"`
}

// NginxSite 是简单模式中的一个站点。
type NginxSite struct {
	// ID 是站点标识；用于前端定位，不影响生成结果。
	ID string `json:"id"`
	// FileName 是站点配置文件名；留空表示按域名自动生成。
	FileName string `json:"fileName"`
	// Preset 是站点级预设；留空时沿用请求级预设。
	Preset string `json:"preset"`
	// ServerName 是站点域名。
	ServerName string `json:"serverName"`
	// ListenPort 是监听端口；默认 80。
	ListenPort int `json:"listenPort"`
	// ListenIPv6 是否同时监听 IPv6。
	ListenIPv6 bool `json:"listenIPv6"`
	// Root 是站点根目录。
	Root string `json:"root"`
	// IndexFiles 是索引文件列表。
	IndexFiles []string `json:"indexFiles"`
	// EnableSSL 是否启用 HTTPS。
	EnableSSL bool `json:"enableSsl"`
	// SSLCertificate 是证书路径。
	SSLCertificate string `json:"sslCertificate"`
	// SSLPrivateKey 是私钥路径。
	SSLPrivateKey string `json:"sslPrivateKey"`
	// TLSProfile 是 TLS 配置档位：modern / intermediate。
	TLSProfile string `json:"tlsProfile"`
	// EnableProxy 是否启用反向代理。
	EnableProxy bool `json:"enableProxy"`
	// ProxyPass 是代理目标地址。
	ProxyPass string `json:"proxyPass"`
	// ProxyPath 是被代理的路径前缀。
	ProxyPath string `json:"proxyPath"`
	// EnableWebSocket 是否启用 WebSocket 升级。
	EnableWebSocket bool `json:"enableWebSocket"`
	// EnableLoadBalance 是否启用负载均衡。
	EnableLoadBalance bool `json:"enableLoadBalance"`
	// Upstreams 是后端服务器列表。
	Upstreams []string `json:"upstreams"`
	// EnableGzip 是否启用 Gzip 压缩。
	EnableGzip bool `json:"enableGzip"`
	// EnableCache 是否启用静态资源缓存。
	EnableCache bool `json:"enableCache"`
	// EnableSecurityHeaders 是否输出安全响应头。
	EnableSecurityHeaders bool `json:"enableSecurityHeaders"`
	// EnableErrorPages 是否输出错误页。
	EnableErrorPages bool `json:"enableErrorPages"`
	// ErrorPageDir 是错误页目录。
	ErrorPageDir string `json:"errorPageDir"`
	// EnablePHP 是否输出 FastCGI 配置。
	EnablePHP bool `json:"enablePhp"`
	// PHPSocket 是 FastCGI 地址，如 unix:/run/php/php8.2-fpm.sock。
	PHPSocket string `json:"phpSocket"`
}

// NginxNode 是配置树中的一个节点。
type NginxNode struct {
	// ID 是节点标识；用于诊断定位。
	ID string `json:"id"`
	// Type 是节点类型：directive / block / raw。
	Type string `json:"type"`
	// Context 是节点所属上下文。
	Context string `json:"context"`
	// Name 是指令名或块名。
	Name string `json:"name"`
	// Args 是原样参数；保留变量、引号与转义。
	Args string `json:"args"`
	// Body 是原样片段内容，仅 raw 节点使用。
	Body string `json:"body"`
	// File 是目标文件；留空表示跟随所在层级。
	File string `json:"file"`
	// Comment 是节点前的注释。
	Comment string `json:"comment"`
	// Children 是子节点，仅块节点使用。
	Children []NginxNode `json:"children"`
}

// NginxFile 是生成结果中的一个文件。
type NginxFile struct {
	// Path 是文件路径。
	Path string `json:"path"`
	// Content 是文件内容。
	Content string `json:"content"`
	// Size 是文件字节数。
	Size int `json:"size"`
	// Description 是文件用途说明。
	Description string `json:"description"`
}

// NginxArchive 是打包产物。
type NginxArchive struct {
	// Filename 是建议的下载文件名。
	Filename string `json:"filename"`
	// ContentType 是 MIME 类型。
	ContentType string `json:"contentType"`
	// Base64 是 ZIP 内容的 Base64 编码。
	Base64 string `json:"base64"`
	// Entries 是压缩包内的条目数。
	Entries int `json:"entries"`
	// UncompressedBytes 是解压后的总字节数。
	UncompressedBytes int `json:"uncompressedBytes"`
}

// NginxDiagnostic 是一条生成诊断。
type NginxDiagnostic struct {
	// Level 是级别：info / warning / error。
	Level string `json:"level"`
	// Code 是诊断码。
	Code string `json:"code"`
	// Context 是相关上下文。
	Context string `json:"context"`
	// NodeID 是相关节点标识。
	NodeID string `json:"nodeId"`
	// Message 是诊断说明；**不**包含用户输入内容。
	Message string `json:"message"`
}

// NginxSummary 是生成结果摘要。
type NginxSummary struct {
	// SiteCount 是站点数。
	SiteCount int `json:"siteCount"`
	// FileCount 是文件数。
	FileCount int `json:"fileCount"`
	// NodeCount 是配置树节点数。
	NodeCount int `json:"nodeCount"`
	// DirectiveCount 是生成的指令条数。
	DirectiveCount int `json:"directiveCount"`
	// WarningCount 是警告数。
	WarningCount int `json:"warningCount"`
	// ErrorCount 是错误数。
	ErrorCount int `json:"errorCount"`
}

// NginxConfigResult 是 nginx-config 工具的响应。
type NginxConfigResult struct {
	// Mode 是实际使用的配置模式。
	Mode string `json:"mode"`
	// TargetVersion 是实际使用的目标版本。
	TargetVersion string `json:"targetVersion"`
	// Output 是实际使用的输出形态。
	Output string `json:"output"`
	// Files 是生成的文件列表。
	Files []NginxFile `json:"files"`
	// Archive 是打包产物；仅 bundle 响应形态有值。
	Archive *NginxArchive `json:"archive,omitempty"`
	// Diagnostics 是生成诊断。
	Diagnostics []NginxDiagnostic `json:"diagnostics"`
	// Summary 是结果摘要。
	Summary NginxSummary `json:"summary"`
}

// NginxConfig 生成 Nginx 配置。
//
// 这是**纯模板拼装**：不涉及密码学运算，也不出网。
func NginxConfig(_ context.Context, req *NginxConfigRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = NginxModeSimple
	}
	output := req.Output
	if output == "" {
		output = NginxOutputBundle
	}
	shape := req.ResponseShape
	if shape == "" {
		shape = NginxShapeFiles
	}
	version := req.TargetVersion
	if version == "" {
		version = defaultNginxVersion()
	}

	diagnostics := []NginxDiagnostic{}

	// 两种模式最终都得到**同一棵配置树**，渲染逻辑因此只有一份。
	var nodes []NginxNode
	if mode == NginxModeAdvanced {
		nodes = req.Nodes
	} else {
		nodes = buildNginxSimpleTree(req, &diagnostics)
	}

	// 版本与上下文兼容性检查：只做事实性判断，不替代目标环境的 nginx -t。
	checkNginxNodes(nodes, version, &diagnostics)

	files := renderNginx(nodes, output, version, &diagnostics)
	if err := checkNginxLimits(files); err != nil {
		return tools.Result{}, err
	}

	result := NginxConfigResult{
		Mode:          mode,
		TargetVersion: version,
		Output:        output,
		Files:         files,
		Diagnostics:   diagnostics,
	}
	result.Summary = summarizeNginx(files, nodes, diagnostics)

	if shape == NginxShapeBundle {
		archive, err := packNginxArchive(files)
		if err != nil {
			return tools.Result{}, err
		}
		result.Archive = archive
	}

	return tools.Result{Data: result}, nil
}

// summarizeNginx 汇总生成结果。
func summarizeNginx(files []NginxFile, nodes []NginxNode, diagnostics []NginxDiagnostic) NginxSummary {
	summary := NginxSummary{
		FileCount: len(files),
		NodeCount: countNginxNodes(nodes),
		SiteCount: countNginxSites(nodes),
	}
	for _, file := range files {
		summary.DirectiveCount += strings.Count(file.Content, ";")
	}
	for _, diag := range diagnostics {
		switch diag.Level {
		case NginxDiagWarning:
			summary.WarningCount++
		case NginxDiagError:
			summary.ErrorCount++
		}
	}
	return summary
}

// countNginxSites 统计站点数量。
//
// 只数**对外提供服务的** server 块：带 `return` 的跳转块不计入，
// 否则一个 HTTPS 站点会被算成两个。
func countNginxSites(nodes []NginxNode) int {
	total := 0
	for i := range nodes {
		node := &nodes[i]
		if node.Type == NginxNodeBlock && node.Name == "server" && !isRedirectServer(node) {
			total++
		}
		total += countNginxSites(node.Children)
	}
	return total
}

// isRedirectServer 判断 server 块是否只是跳转。
func isRedirectServer(node *NginxNode) bool {
	for _, child := range node.Children {
		if child.Name == "return" {
			return true
		}
	}
	return false
}

// countNginxNodes 统计配置树节点数。
func countNginxNodes(nodes []NginxNode) int {
	total := 0
	for _, node := range nodes {
		total++
		total += countNginxNodes(node.Children)
	}
	return total
}

// checkNginxLimits 校验输出规模上限。
func checkNginxLimits(files []NginxFile) error {
	if len(files) > maxNginxFiles {
		return tools.InvalidField("sites", []int{0, maxNginxFiles}, len(files))
	}
	total := 0
	for _, file := range files {
		if len(file.Content) > maxNginxFileSize {
			return tools.InvalidField("output", []int{0, maxNginxFileSize}, len(file.Content))
		}
		total += len(file.Content)
	}
	if total > maxNginxTotalSize {
		return tools.InvalidField("output", []int{0, maxNginxTotalSize}, total)
	}
	return nil
}

// defaultNginxVersion 返回默认目标版本。
func defaultNginxVersion() string {
	versions := nginx.Versions()
	if len(versions) == 0 {
		return ""
	}
	return versions[0]
}

// Validate 做字段级校验。
func (r *NginxConfigRequest) Validate() error {
	mode := r.Mode
	if mode == "" {
		mode = NginxModeSimple
	}
	if !containsString([]string{NginxModeSimple, NginxModeAdvanced}, mode) {
		return tools.InvalidField("mode", []string{NginxModeSimple, NginxModeAdvanced}, mode)
	}

	output := r.Output
	if output == "" {
		output = NginxOutputBundle
	}
	if !containsString([]string{NginxOutputBundle, NginxOutputSingle}, output) {
		return tools.InvalidField("output", []string{NginxOutputBundle, NginxOutputSingle}, output)
	}

	shape := r.ResponseShape
	if shape == "" {
		shape = NginxShapeFiles
	}
	if !containsString([]string{NginxShapeFiles, NginxShapeBundle}, shape) {
		return tools.InvalidField("responseShape", []string{NginxShapeFiles, NginxShapeBundle}, shape)
	}

	if r.TargetVersion != "" {
		versions := nginx.Versions()
		if !containsString(versions, r.TargetVersion) {
			return tools.InvalidField("targetVersion", versions, r.TargetVersion)
		}
	}

	if mode == NginxModeSimple {
		if len(r.Sites) > maxNginxSites {
			return tools.InvalidField("sites", []int{0, maxNginxSites}, len(r.Sites))
		}
		for i := range r.Sites {
			if err := r.Sites[i].validate(); err != nil {
				return err
			}
		}
		return nil
	}

	return validateNginxNodes(r.Nodes)
}

// validate 校验单个站点。
func (s *NginxSite) validate() error {
	if strings.TrimSpace(s.ServerName) == "" {
		return tools.MissingField("serverName")
	}
	if s.ListenPort < 0 || s.ListenPort > 65535 {
		return tools.InvalidField("listenPort", []int{0, 65535}, s.ListenPort)
	}
	if len(s.Upstreams) > maxNginxUpstreams {
		return tools.InvalidField("upstreams", []int{0, maxNginxUpstreams}, len(s.Upstreams))
	}
	if s.FileName != "" && !isSafeNginxFileName(s.FileName) {
		return tools.InvalidField("fileName", []string{"[A-Za-z0-9._-]"}, "unsafe")
	}
	if s.Preset != "" && !containsString(nginxPresets(), s.Preset) {
		return tools.InvalidField("preset", nginxPresets(), s.Preset)
	}
	if s.TLSProfile != "" && !containsString(nginxTLSProfiles(), s.TLSProfile) {
		return tools.InvalidField("tlsProfile", nginxTLSProfiles(), s.TLSProfile)
	}
	return nil
}

// validateNginxNodes 校验高级模式的配置树。
func validateNginxNodes(nodes []NginxNode) error {
	total := 0
	if err := walkNginxNodes(nodes, 0, &total); err != nil {
		return err
	}
	if total > maxNginxNodes {
		return tools.InvalidField("nodes", []int{0, maxNginxNodes}, total)
	}
	return nil
}

// walkNginxNodes 递归校验节点。
func walkNginxNodes(nodes []NginxNode, depth int, total *int) error {
	if depth > maxNginxDepth {
		return tools.InvalidField("nodes", []int{0, maxNginxDepth}, depth)
	}
	for i := range nodes {
		node := &nodes[i]
		*total++
		switch node.Type {
		case NginxNodeDirective, NginxNodeBlock, NginxNodeRaw:
		default:
			return tools.InvalidField("type", []string{NginxNodeDirective, NginxNodeBlock, NginxNodeRaw}, node.Type)
		}
		if node.Type == NginxNodeRaw {
			if strings.TrimSpace(node.Body) == "" {
				return tools.MissingField("body")
			}
			continue
		}
		if !isValidNginxDirectiveName(node.Name) {
			return tools.InvalidField("name", []string{"[A-Za-z0-9_]"}, "invalid")
		}
		if len(node.Args) > maxNginxNodesArgs {
			return tools.InvalidField("args", []int{0, maxNginxNodesArgs}, len(node.Args))
		}
		if node.File != "" && !isSafeNginxFilePath(node.File) {
			return tools.InvalidField("file", []string{"relative path"}, "unsafe")
		}
		if err := walkNginxNodes(node.Children, depth+1, total); err != nil {
			return err
		}
	}
	return nil
}
