package system

import (
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools/system/nginx"
)

// nginxIndentUnit 是缩进单位。
const nginxIndentUnit = "    "

// nginxFileModel 是一个文件的渲染模型。
type nginxFileModel struct {
	// path 是文件路径。
	path string
	// items 是文件内的条目。
	items []nginxItem
}

// nginxItem 是文件内的一个条目。
type nginxItem struct {
	// node 是内联节点；include 为空时有效。
	node *NginxNode
	// include 是被引入到其他文件的路径；非空时本条目只渲染一行 include。
	include string
	// children 是块节点的子条目。
	children []nginxItem
}

// nginxRenderer 把配置树渲染成文件集合。
//
// 多文件与单文件**共用同一棵配置树**：区别只在于节点是否允许被拆到其他文件。
type nginxRenderer struct {
	// output 是输出形态。
	output string
	// version 是目标版本。
	version string
	// diag 收集诊断。
	diag *[]NginxDiagnostic
	// models 是各文件的渲染模型。
	models map[string]*nginxFileModel
	// order 是文件顺序。
	order []string
	// usedNames 是已占用的派生文件名。
	usedNames map[string]bool
}

// renderNginx 把配置树渲染成文件列表。
func renderNginx(nodes []NginxNode, output, version string, diag *[]NginxDiagnostic) []NginxFile {
	r := &nginxRenderer{
		output:    output,
		version:   version,
		diag:      diag,
		models:    map[string]*nginxFileModel{},
		usedNames: map[string]bool{},
	}

	root := &nginxFileModel{path: nginxMainFile}
	r.models[nginxMainFile] = root
	r.order = append(r.order, nginxMainFile)
	root.items = r.buildItems(nodes, nginxMainFile, NginxContextMain)

	files := make([]NginxFile, 0, len(r.order))
	for _, path := range r.order {
		model := r.models[path]
		content := r.renderItems(model.items, 0, path)
		files = append(files, NginxFile{
			Path:        path,
			Content:     content,
			Size:        len(content),
			Description: describeNginxFile(path),
		})
	}
	return files
}

// nginxMainFile 是主配置文件名。
const nginxMainFile = "nginx.conf"

// buildItems 把节点列表转成文件条目，并解析被拆到其他文件的节点。
func (r *nginxRenderer) buildItems(nodes []NginxNode, currentFile, ctx string) []nginxItem {
	items := make([]nginxItem, 0, len(nodes))
	// included 记录本层已引入过的文件：同一个文件只 include 一次，
	// 其余同文件节点仍然并入该文件。
	included := map[string]bool{}
	for i := range nodes {
		node := &nodes[i]
		target := r.targetFile(node, currentFile, ctx)
		if target != currentFile {
			// 节点属于另一个文件：此处改为 include 该文件。
			r.addToFile(target, node, ctx)
			if included[target] {
				continue
			}
			included[target] = true
			items = append(items, nginxItem{include: target})
			continue
		}
		item := nginxItem{node: node}
		if node.Type == NginxNodeBlock {
			item.children = r.buildItems(node.Children, currentFile, blockContextOf(node, ctx))
		}
		items = append(items, item)
	}
	return items
}

// addToFile 把节点并入目标文件。
func (r *nginxRenderer) addToFile(path string, node *NginxNode, ctx string) {
	model, ok := r.models[path]
	if !ok {
		model = &nginxFileModel{path: path}
		r.models[path] = model
		r.order = append(r.order, path)
	}
	item := nginxItem{node: node}
	if node.Type == NginxNodeBlock {
		item.children = r.buildItems(node.Children, path, blockContextOf(node, ctx))
	}
	model.items = append(model.items, item)
}

// targetFile 计算节点的目标文件。
func (r *nginxRenderer) targetFile(node *NginxNode, currentFile, ctx string) string {
	if r.output == NginxOutputSingle {
		// 单文件模式不拆文件：全部内容原地内联。
		return currentFile
	}
	if node.File != "" {
		return node.File
	}
	if node.Type != NginxNodeBlock {
		return currentFile
	}
	switch node.Name {
	case "server":
		switch ctx {
		case NginxContextHTTP:
			return "conf.d/" + r.uniqueFileName(deriveServerFileName(node)) + ".conf"
		case NginxContextStream:
			return "stream.d/" + r.uniqueFileName(deriveServerFileName(node)) + ".conf"
		case NginxContextMail:
			return "mail.d/" + r.uniqueFileName(deriveServerFileName(node)) + ".conf"
		}
	case "types":
		if ctx == NginxContextHTTP {
			return "mime.types"
		}
	}
	return currentFile
}

// uniqueFileName 为派生文件名去重。
func (r *nginxRenderer) uniqueFileName(name string) string {
	base := name
	candidate := base
	for i := 2; r.usedNames[candidate]; i++ {
		candidate = base + "-" + itoa(i)
	}
	r.usedNames[candidate] = true
	return candidate
}

// deriveServerFileName 从 server 块推导配置文件名。
//
// 优先取 server_name 的第一个名字；取不到时用节点标识；
// 都不可用时回退到 site。
func deriveServerFileName(node *NginxNode) string {
	for _, child := range node.Children {
		if child.Name != "server_name" {
			continue
		}
		fields := strings.Fields(child.Args)
		if len(fields) == 0 {
			continue
		}
		if name := sanitizeNginxFileStem(fields[0]); name != "" {
			return name
		}
	}
	if name := sanitizeNginxFileStem(node.ID); name != "" {
		return name
	}
	return "site"
}

// sanitizeNginxName 把任意字符串整理成合法的指令名。
//
// 上游组名会出现在 `proxy_pass http://<name>` 里，
// 因此**不允许**点号等 URL 语义字符。
func sanitizeNginxName(raw string) string {
	lower := strings.ToLower(raw)
	var b strings.Builder
	lastDash := false
	for _, r := range lower {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 48 {
		out = strings.Trim(out[:48], "-")
	}
	return out
}

// sanitizeNginxFileStem 把任意字符串整理成安全的文件名主干。
//
// 保留字母、数字与点号（域名形态的文件名很常见），
// 其余字符统一替换为连字符。
func sanitizeNginxFileStem(raw string) string {
	lower := strings.ToLower(raw)
	var b strings.Builder
	lastDash := false
	for _, r := range lower {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case r == '.':
			if b.Len() > 0 && !lastDash {
				b.WriteRune(r)
				lastDash = false
			}
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-.")
	if len(out) > 48 {
		out = strings.Trim(out[:48], "-.")
	}
	return out
}

// blockContextOf 返回块节点的子节点上下文。
func blockContextOf(node *NginxNode, parent string) string {
	switch node.Name {
	case "events":
		return NginxContextEvents
	case "http":
		return NginxContextHTTP
	case "stream":
		return NginxContextStream
	case "mail":
		return NginxContextMail
	case "server":
		return NginxContextServer
	case "location":
		return NginxContextLocation
	case "upstream":
		return NginxContextUpstream
	case "if":
		return NginxContextIf
	case "limit_except":
		return NginxContextLimitExcept
	case "types":
		return NginxContextTypes
	default:
		return parent
	}
}

// renderItems 渲染文件条目。
func (r *nginxRenderer) renderItems(items []nginxItem, indent int, path string) string {
	var sb strings.Builder
	for _, item := range items {
		if item.include != "" {
			sb.WriteString(r.indent(indent) + "include " + item.include + ";\n")
			continue
		}
		sb.WriteString(r.renderItem(item, indent, path))
	}
	return sb.String()
}

// renderItem 渲染单个条目。
func (r *nginxRenderer) renderItem(item nginxItem, indent int, path string) string {
	node := item.node
	var sb strings.Builder
	if node.Comment != "" {
		sb.WriteString(r.indent(indent) + "# " + node.Comment + "\n")
	}
	switch node.Type {
	case NginxNodeRaw:
		body := strings.TrimRight(node.Body, "\n")
		for _, line := range strings.Split(body, "\n") {
			sb.WriteString(r.indent(indent) + line + "\n")
		}
	case NginxNodeBlock:
		header := node.Name
		if node.Args != "" {
			header += " " + node.Args
		}
		sb.WriteString(r.indent(indent) + header + " {\n")
		sb.WriteString(r.renderItems(item.children, indent+1, path))
		sb.WriteString(r.indent(indent) + "}\n")
	default:
		line := node.Name
		if node.Args != "" {
			line += " " + node.Args
		}
		sb.WriteString(r.indent(indent) + line + ";\n")
	}
	return sb.String()
}

// indent 生成缩进前缀。
func (r *nginxRenderer) indent(level int) string {
	if level <= 0 {
		return ""
	}
	return strings.Repeat(nginxIndentUnit, level)
}

// describeNginxFile 返回文件的用途标识。
//
// 返回的是**稳定标识**而非文案：文案由前端按当前语言渲染。
func describeNginxFile(path string) string {
	switch {
	case path == nginxMainFile:
		return "main"
	case path == "mime.types":
		return "mimeTypes"
	case strings.HasPrefix(path, "conf.d/"):
		return "site"
	case strings.HasPrefix(path, "stream.d/"):
		return "streamSite"
	case strings.HasPrefix(path, "mail.d/"):
		return "mailSite"
	case strings.HasPrefix(path, "snippets/"):
		return "snippet"
	default:
		return "custom"
	}
}

// isValidNginxDirectiveName 判断指令名是否合法。
func isValidNginxDirectiveName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
	}
	return true
}

// isSafeNginxFileName 判断文件名是否安全。
//
// 只允许字母、数字、点、下划线与连字符；不含路径分隔符。
func isSafeNginxFileName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	if name == "." || name == ".." {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

// isSafeNginxFilePath 判断相对文件路径是否安全。
//
// 禁止绝对路径、上跳路径与路径分隔符之外的字符。
func isSafeNginxFilePath(path string) bool {
	if path == "" || len(path) > 128 {
		return false
	}
	if strings.HasPrefix(path, "/") || strings.Contains(path, "\\") {
		return false
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
		if !isSafeNginxFileName(segment) {
			return false
		}
	}
	return true
}

// checkNginxNodes 校验配置树与目标版本的兼容性。
//
// 这里只做**事实性**检查：指令是否存在、是否在目标版本可用、
// 是否适用于所在上下文。配置能否真正跑起来仍以目标环境的 nginx -t 为准。
func checkNginxNodes(nodes []NginxNode, version string, diag *[]NginxDiagnostic) {
	snapshot := nginx.Load()
	walkNginxForCheck(nodes, NginxContextMain, version, snapshot, diag)
}

// walkNginxForCheck 递归检查节点。
func walkNginxForCheck(nodes []NginxNode, ctx, version string, snapshot *nginx.Snapshot, diag *[]NginxDiagnostic) {
	for i := range nodes {
		node := &nodes[i]
		if node.Type == NginxNodeRaw {
			continue
		}
		matches := nginx.FindDirective(snapshot, node.Name)
		if len(matches) == 0 {
			appendNginxDiag(diag, NginxDiagInfo, "UNKNOWN_DIRECTIVE", node.Context, node.ID,
				"指令 "+node.Name+" 不在官方目录中，可能是第三方模块或自定义指令")
			continue
		}
		available := false
		for _, match := range matches {
			if nginx.AvailableIn(match, version) {
				available = true
			}
			if match.Deprecated != "" && nginx.DeprecatedIn(match, version) {
				appendNginxDiag(diag, NginxDiagWarning, "DEPRECATED_DIRECTIVE", node.Context, node.ID,
					"该指令在目标版本中已废弃")
			}
		}
		if !available {
			appendNginxDiag(diag, NginxDiagWarning, "UNSUPPORTED_VERSION", node.Context, node.ID,
				"该指令在目标版本中不可用")
		}
		if !nginx.AllowsContext(snapshot, node.Name, node.Context) {
			appendNginxDiag(diag, NginxDiagWarning, "CONTEXT_MISMATCH", node.Context, node.ID,
				"指令 "+node.Name+" 不适用于所在上下文")
		}
		// 块与指令写反是常见错误：`location /x;` 会让 Nginx 直接启动失败。
		if node.Type == NginxNodeDirective && isNginxBlockDirective(snapshot, node.Name) {
			appendNginxDiag(diag, NginxDiagError, "BLOCK_REQUIRED", node.Context, node.ID,
				"指令 "+node.Name+" 是配置块，应使用配置块节点")
		}
		if node.Type == NginxNodeBlock && !isNginxBlockDirective(snapshot, node.Name) && len(matches) > 0 {
			appendNginxDiag(diag, NginxDiagWarning, "BLOCK_NOT_EXPECTED", node.Context, node.ID,
				"指令 "+node.Name+" 不是配置块，写成块可能无法启动")
		}
		if node.Type == NginxNodeBlock {
			walkNginxForCheck(node.Children, blockContextOf(node, ctx), version, snapshot, diag)
		}
	}
}

// isNginxBlockDirective 判断指令在官方目录里是否是配置块。
//
// 同名指令可能跨模块出现，只要有一条是块就按块处理。
func isNginxBlockDirective(snapshot *nginx.Snapshot, name string) bool {
	for _, directive := range nginx.FindDirective(snapshot, name) {
		if directive.Block {
			return true
		}
	}
	return false
}

// appendNginxDiag 追加一条诊断。
func appendNginxDiag(diag *[]NginxDiagnostic, level, code, ctx, nodeID, message string) {
	*diag = append(*diag, NginxDiagnostic{
		Level:   level,
		Code:    code,
		Context: ctx,
		NodeID:  nodeID,
		Message: message,
	})
}
