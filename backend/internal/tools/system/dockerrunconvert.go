package system

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 转换方向。
const (
	// DockerConvertRunToCompose 表示把 docker run 命令转成 Compose 配置。
	DockerConvertRunToCompose = "runToCompose"
	// DockerConvertComposeToRun 表示把 Compose 配置转成 docker run 命令。
	DockerConvertComposeToRun = "composeToRun"
)

// dockerConvertInputLimit 是输入文本的字节上限。
const dockerConvertInputLimit = 64 * 1024

// dockerConvertMaxServices 是 composeToRun 方向可处理的服务数量上限。
const dockerConvertMaxServices = 32

// dockerServiceNamePattern 是合法的 Compose 服务名。
var dockerServiceNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// dockerPlainKeyPattern 是可省略引号的 YAML 映射键。
var dockerPlainKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// dockerShellSafePattern 是无需引号即可安全出现在 shell 中的字符集。
var dockerShellSafePattern = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// dockerYAMLLinePattern 从 YAML 解析错误中提取行号。
var dockerYAMLLinePattern = regexp.MustCompile(`line (\d+)`)

// dockerRunCommandPattern 匹配以 `docker run` / `podman run` 开头的命令。
//
// 允许前置 `sudo`；`run` 后面必须跟空白或结束，避免把 `docker runc` 误判成命令。
var dockerRunCommandPattern = regexp.MustCompile(`^(sudo\s+)?(docker|podman)\s+run(\s|$)`)

// DockerConvertDirections 返回支持的转换方向（顺序稳定）。
func DockerConvertDirections() []string {
	return []string{DockerConvertRunToCompose, DockerConvertComposeToRun}
}

// DockerRunConvertRequest 是 docker-run-convert 工具的请求。
type DockerRunConvertRequest struct {
	// Direction 是转换方向；缺省 runToCompose。
	Direction string `json:"direction"`
	// Input 是源文本：docker run 命令或 Compose 配置。
	Input string `json:"input"`
	// ServiceName 是 runToCompose 方向的服务名；留空则由镜像名推导。
	ServiceName string `json:"serviceName"`
	// Strict 是否严格模式：遇到未支持的项直接报错，而不是忽略并列出。
	Strict bool `json:"strict"`
}

// DockerConvertService 是转换出的单个服务。
type DockerConvertService struct {
	// Name 是服务名。
	Name string `json:"name"`
	// Image 是镜像。
	Image string `json:"image"`
	// Exec 是容器内执行的命令（runToCompose 方向）。
	Exec string `json:"exec"`
	// RunCommand 是生成的 docker run 命令（composeToRun 方向）。
	RunCommand string `json:"runCommand"`
}

// DockerRunConvertResult 是 docker-run-convert 工具的响应。
type DockerRunConvertResult struct {
	// Direction 是实际使用的转换方向。
	Direction string `json:"direction"`
	// Output 是转换结果文本。
	Output string `json:"output"`
	// Image 是解析到的镜像（runToCompose 方向）。
	Image string `json:"image"`
	// ServiceName 是服务名（runToCompose 方向）。
	ServiceName string `json:"serviceName"`
	// Services 是转换出的服务列表。
	Services []DockerConvertService `json:"services"`
	// Unsupported 是未转换的项：**只有**参数名或 Compose 键名，**不含**取值。
	Unsupported []string `json:"unsupported"`
}

// DockerRunConvert 在 docker run 命令与 Compose 配置之间互转。
//
// 这是**纯文本转换**：不连接任何守护进程，也不拉取镜像。
func DockerRunConvert(_ context.Context, req *DockerRunConvertRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	// 方向**优先**由输入形态决定：用户选的方向与输入不符时自动纠正，
	// 避免把 `docker run` 命令当 YAML 解析后报「expected a mapping」。
	direction := req.Direction
	if direction == "" {
		direction = DockerConvertRunToCompose
	}
	warnings := []toolerr.Warning{}
	if detected := detectConvertDirection(req.Input); detected != "" && detected != direction {
		direction = detected
		warnings = append(warnings, toolerr.Warning{
			Code:   toolerr.WarnDirectionAdjusted,
			Params: map[string]any{"direction": detected},
		})
	}

	var (
		result DockerRunConvertResult
		err    error
	)
	if direction == DockerConvertComposeToRun {
		result, err = dockerComposeToRun(req.Input)
	} else {
		result, err = dockerRunToCompose(req)
	}
	if err != nil {
		return tools.Result{}, err
	}

	if req.Strict && len(result.Unsupported) > 0 {
		return tools.Result{}, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":       "input",
			"reason":      "unsupported options",
			"unsupported": result.Unsupported,
		})
	}
	return tools.Result{Data: result, Warnings: warnings}, nil
}

// detectConvertDirection 按输入形态判断转换方向。
//
// 返回值：
//   - DockerConvertRunToCompose：输入是 `docker run` / `podman run` 命令；
//   - DockerConvertComposeToRun：输入是含 `services` 的 YAML 映射；
//   - 空串：形态不明确，交回调用方选定的方向。
//
// 「含 services 的 YAML 映射」才算 Compose：只判断「是不是映射」会把
// Kubernetes 清单等 YAML 也当成 Compose，掩盖「缺少 services」的真实原因。
func detectConvertDirection(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}

	var document yaml.Node
	if err := yaml.Unmarshal([]byte(trimmed), &document); err == nil {
		root := &document
		if root.Kind == yaml.DocumentNode {
			if len(root.Content) == 0 {
				return ""
			}
			root = root.Content[0]
		}
		if root.Kind == yaml.MappingNode {
			if yamlMappingValue(root, "services") != nil {
				return DockerConvertComposeToRun
			}
			return ""
		}
	}

	if dockerRunCommandPattern.MatchString(trimmed) {
		return DockerConvertRunToCompose
	}
	return ""
}

// Validate 做字段级校验。
func (r *DockerRunConvertRequest) Validate() error {
	switch r.Direction {
	case "", DockerConvertRunToCompose, DockerConvertComposeToRun:
	default:
		return tools.UnsupportedMode(r.Direction, DockerConvertDirections())
	}
	if strings.TrimSpace(r.Input) == "" {
		return tools.MissingField("input")
	}
	if len(r.Input) > dockerConvertInputLimit {
		return toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"field":    "input",
			"maxBytes": dockerConvertInputLimit,
		})
	}
	if r.ServiceName != "" && !dockerServiceNamePattern.MatchString(r.ServiceName) {
		return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "serviceName",
			"expected": "^[A-Za-z0-9][A-Za-z0-9_.-]*$",
			"actual":   len(r.ServiceName),
		})
	}
	return nil
}

// ── docker run → Compose ────────────────────────────────────────────────

// dockerRunSpec 是 docker run 命令的解析结果。
type dockerRunSpec struct {
	image         string
	command       []string
	containerName string
	ports         []string
	environment   []string
	volumes       []string
	networks      []string
	networkMode   string
	restart       string
	entrypoint    string
	workingDir    string
	user          string
	hostname      string
	capAdd        []string
	capDrop       []string
	labels        []string
	tmpfs         []string
	devices       []string
	memLimit      string
	cpus          string
	privileged    bool
	tty           bool
	stdinOpen     bool
	init          bool
	readOnly      bool
	unsupported   []string
}

// dockerRunValueFlag 是带值长参数的处理函数。
type dockerRunValueFlag func(spec *dockerRunSpec, value string)

// dockerRunBoolFlag 是开关长参数的处理函数。
type dockerRunBoolFlag func(spec *dockerRunSpec)

// dockerRunValueFlags 是**已支持**的带值长参数。
var dockerRunValueFlags = map[string]dockerRunValueFlag{
	"name":       func(s *dockerRunSpec, v string) { s.containerName = v },
	"publish":    func(s *dockerRunSpec, v string) { s.ports = append(s.ports, v) },
	"env":        func(s *dockerRunSpec, v string) { s.environment = append(s.environment, v) },
	"volume":     func(s *dockerRunSpec, v string) { s.volumes = append(s.volumes, v) },
	"network":    func(s *dockerRunSpec, v string) { s.networks = append(s.networks, v) },
	"net":        func(s *dockerRunSpec, v string) { s.networks = append(s.networks, v) },
	"restart":    func(s *dockerRunSpec, v string) { s.restart = v },
	"entrypoint": func(s *dockerRunSpec, v string) { s.entrypoint = v },
	"workdir":    func(s *dockerRunSpec, v string) { s.workingDir = v },
	"user":       func(s *dockerRunSpec, v string) { s.user = v },
	"hostname":   func(s *dockerRunSpec, v string) { s.hostname = v },
	"cap-add":    func(s *dockerRunSpec, v string) { s.capAdd = append(s.capAdd, v) },
	"cap-drop":   func(s *dockerRunSpec, v string) { s.capDrop = append(s.capDrop, v) },
	"label":      func(s *dockerRunSpec, v string) { s.labels = append(s.labels, v) },
	"tmpfs":      func(s *dockerRunSpec, v string) { s.tmpfs = append(s.tmpfs, v) },
	"device":     func(s *dockerRunSpec, v string) { s.devices = append(s.devices, v) },
	"memory":     func(s *dockerRunSpec, v string) { s.memLimit = v },
	"cpus":       func(s *dockerRunSpec, v string) { s.cpus = v },
}

// dockerRunBoolFlags 是**已支持**的开关长参数。
var dockerRunBoolFlags = map[string]dockerRunBoolFlag{
	"privileged":  func(s *dockerRunSpec) { s.privileged = true },
	"tty":         func(s *dockerRunSpec) { s.tty = true },
	"interactive": func(s *dockerRunSpec) { s.stdinOpen = true },
	"init":        func(s *dockerRunSpec) { s.init = true },
	"read-only":   func(s *dockerRunSpec) { s.readOnly = true },
}

// dockerRunSkippedValueFlags 是**未支持但已知带值**的长参数。
//
// 命中它们时必须跳过取值，否则会把取值误当成镜像名。
var dockerRunSkippedValueFlags = map[string]bool{
	"add-host": true, "annotation": true, "attach": true, "blkio-weight": true,
	"blkio-weight-device": true, "cgroup-parent": true, "cgroupns": true, "cidfile": true,
	"cpu-period": true, "cpu-quota": true, "cpu-shares": true, "cpuset-cpus": true,
	"cpuset-mems": true, "device-cgroup-rule": true, "device-read-bps": true,
	"device-read-iops": true, "device-write-bps": true, "device-write-iops": true,
	"dns": true, "dns-option": true, "dns-search": true, "env-file": true, "expose": true,
	"gpus": true, "group-add": true, "health-cmd": true, "health-interval": true,
	"health-retries": true, "health-start-interval": true, "health-start-period": true,
	"health-timeout": true, "ip": true, "ip6": true, "ipc": true, "isolation": true,
	"kernel-memory": true, "label-file": true, "link": true, "link-local-ip": true,
	"log-driver": true, "log-opt": true, "memory-reservation": true, "memory-swap": true,
	"mount": true, "network-alias": true, "net-alias": true, "oom-score-adj": true,
	"pid": true, "pids-limit": true, "platform": true, "pull": true, "runtime": true,
	"security-opt": true, "shm-size": true, "stop-signal": true, "stop-timeout": true,
	"storage-opt": true, "sysctl": true, "ulimit": true, "userns": true, "uts": true,
	"volume-driver": true, "volumes-from": true,
}

// dockerShortValueFlags 是**已支持**的带值短参数。
var dockerShortValueFlags = map[byte]bool{
	'p': true, 'e': true, 'v': true, 'w': true, 'u': true, 'l': true, 'm': true, 'h': true,
}

// dockerRunToCompose 把 docker run 命令转成 Compose 配置。
func dockerRunToCompose(req *DockerRunConvertRequest) (DockerRunConvertResult, error) {
	tokens, err := splitShellWords(req.Input)
	if err != nil {
		return DockerRunConvertResult{}, err
	}
	if len(tokens) == 0 {
		return DockerRunConvertResult{}, tools.MissingField("input")
	}

	spec, err := parseDockerRun(tokens)
	if err != nil {
		return DockerRunConvertResult{}, err
	}

	serviceName := req.ServiceName
	if serviceName == "" {
		serviceName = deriveDockerServiceName(spec.image)
	}

	return DockerRunConvertResult{
		Direction:   DockerConvertRunToCompose,
		Output:      renderDockerCompose(spec, serviceName),
		Image:       spec.image,
		ServiceName: serviceName,
		Services: []DockerConvertService{{
			Name:  serviceName,
			Image: spec.image,
			Exec:  strings.Join(spec.command, " "),
		}},
		Unsupported: spec.unsupported,
	}, nil
}

// parseDockerRun 解析 docker run 命令的 token 序列。
func parseDockerRun(tokens []string) (*dockerRunSpec, error) {
	spec := &dockerRunSpec{unsupported: []string{}}
	tokens = trimDockerCommandPrefix(tokens)

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch {
		case token == "--":
			spec.command = append(spec.command, tokens[i+1:]...)
			i = len(tokens)
		case strings.HasPrefix(token, "--"):
			name, value, hasValue := strings.Cut(token[2:], "=")
			if handler, ok := dockerRunValueFlags[name]; ok {
				if !hasValue {
					if i+1 >= len(tokens) {
						return nil, tools.MissingField("--" + name)
					}
					i++
					value = tokens[i]
				}
				handler(spec, value)
				continue
			}
			if handler, ok := dockerRunBoolFlags[name]; ok {
				handler(spec)
				continue
			}
			spec.unsupported = append(spec.unsupported, "--"+name)
			if dockerRunSkippedValueFlags[name] && !hasValue && i+1 < len(tokens) {
				i++
			}
		case strings.HasPrefix(token, "-") && len(token) > 1:
			next, err := parseDockerShortFlags(spec, tokens, i)
			if err != nil {
				return nil, err
			}
			i = next
		case spec.image == "":
			spec.image = token
		default:
			spec.command = append(spec.command, token)
		}
	}

	if spec.image == "" {
		return nil, tools.MissingField("image")
	}
	return spec, nil
}

// parseDockerShortFlags 解析短参数簇，返回消耗到的 token 下标。
func parseDockerShortFlags(spec *dockerRunSpec, tokens []string, index int) (int, error) {
	cluster := tokens[index][1:]
	for i := 0; i < len(cluster); i++ {
		char := cluster[i]
		if !dockerShortValueFlags[char] {
			switch char {
			case 't':
				spec.tty = true
			case 'i':
				spec.stdinOpen = true
			default:
				spec.unsupported = append(spec.unsupported, "-"+string(char))
			}
			continue
		}
		value := cluster[i+1:]
		if value == "" {
			if index+1 >= len(tokens) {
				return 0, tools.MissingField("-" + string(char))
			}
			index++
			value = tokens[index]
		}
		applyDockerShortValue(spec, char, value)
		return index, nil
	}
	return index, nil
}

// applyDockerShortValue 施加一个带值短参数。
func applyDockerShortValue(spec *dockerRunSpec, char byte, value string) {
	switch char {
	case 'p':
		spec.ports = append(spec.ports, value)
	case 'e':
		spec.environment = append(spec.environment, value)
	case 'v':
		spec.volumes = append(spec.volumes, value)
	case 'w':
		spec.workingDir = value
	case 'u':
		spec.user = value
	case 'l':
		spec.labels = append(spec.labels, value)
	case 'm':
		spec.memLimit = value
	case 'h':
		spec.hostname = value
	}
}

// trimDockerCommandPrefix 去掉命令前缀（docker run / podman run）。
func trimDockerCommandPrefix(tokens []string) []string {
	if len(tokens) > 0 && (tokens[0] == "docker" || tokens[0] == "podman") {
		tokens = tokens[1:]
	}
	if len(tokens) > 0 && tokens[0] == "run" {
		tokens = tokens[1:]
	}
	return tokens
}

// deriveDockerServiceName 从镜像名推导服务名。
func deriveDockerServiceName(image string) string {
	name := image
	if index := strings.LastIndex(name, "/"); index >= 0 {
		name = name[index+1:]
	}
	if index := strings.Index(name, "@"); index >= 0 {
		name = name[:index]
	}
	if index := strings.Index(name, ":"); index >= 0 {
		name = name[:index]
	}

	var builder strings.Builder
	for _, char := range name {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z',
			char >= '0' && char <= '9', char == '.', char == '_', char == '-':
			builder.WriteRune(char)
		default:
			builder.WriteRune('-')
		}
	}
	result := strings.Trim(builder.String(), "-._")
	if result == "" {
		return "app"
	}
	return result
}

// renderDockerCompose 把解析结果渲染成 Compose 配置文本。
func renderDockerCompose(spec *dockerRunSpec, serviceName string) string {
	lines := []string{"services:", "  " + serviceName + ":"}
	add := func(key string, value string) {
		lines = append(lines, "    "+key+": "+value)
	}
	addList := func(key string, values []string) {
		if len(values) == 0 {
			return
		}
		lines = append(lines, "    "+key+":")
		for _, value := range values {
			lines = append(lines, "      - "+quoteComposeScalar(value))
		}
	}

	add("image", quoteComposeScalar(spec.image))
	if spec.containerName != "" {
		add("container_name", quoteComposeScalar(spec.containerName))
	}
	if spec.entrypoint != "" {
		add("entrypoint", quoteComposeScalar(spec.entrypoint))
	}
	addList("command", spec.command)
	addList("ports", spec.ports)
	addList("volumes", spec.volumes)
	lines = appendComposeKeyValue(lines, "environment", spec.environment)
	lines = appendComposeKeyValue(lines, "labels", spec.labels)
	addList("cap_add", spec.capAdd)
	addList("cap_drop", spec.capDrop)
	addList("tmpfs", spec.tmpfs)
	addList("devices", spec.devices)
	if spec.workingDir != "" {
		add("working_dir", quoteComposeScalar(spec.workingDir))
	}
	if spec.user != "" {
		add("user", quoteComposeScalar(spec.user))
	}
	if spec.hostname != "" {
		add("hostname", quoteComposeScalar(spec.hostname))
	}
	if spec.restart != "" {
		add("restart", quoteComposeScalar(spec.restart))
	}
	if spec.memLimit != "" {
		add("mem_limit", quoteComposeScalar(spec.memLimit))
	}
	if spec.cpus != "" {
		add("cpus", quoteComposeScalar(spec.cpus))
	}
	if spec.privileged {
		add("privileged", "true")
	}
	if spec.tty {
		add("tty", "true")
	}
	if spec.stdinOpen {
		add("stdin_open", "true")
	}
	if spec.init {
		add("init", "true")
	}
	if spec.readOnly {
		add("read_only", "true")
	}

	named := []string{}
	for _, network := range spec.networks {
		switch network {
		case "host", "none", "bridge":
			spec.networkMode = network
		case "default":
		default:
			named = append(named, network)
		}
	}
	if spec.networkMode != "" {
		add("network_mode", quoteComposeScalar(spec.networkMode))
	}
	addList("networks", named)

	if len(named) > 0 {
		lines = append(lines, "", "networks:")
		for _, network := range named {
			lines = append(lines, "  "+network+":", "    driver: bridge")
		}
	}
	return strings.Join(lines, "\n")
}

// appendComposeKeyValue 把 K=V 列表渲染成映射或序列。
//
// 全部条目都带 `=` 时用映射形式（键排序），否则用序列形式。
func appendComposeKeyValue(lines []string, key string, entries []string) []string {
	if len(entries) == 0 {
		return lines
	}
	for _, entry := range entries {
		if !strings.Contains(entry, "=") {
			lines = append(lines, "    "+key+":")
			for _, item := range entries {
				lines = append(lines, "      - "+quoteComposeScalar(item))
			}
			return lines
		}
	}

	values := make(map[string]string, len(entries))
	for _, entry := range entries {
		name, value, _ := strings.Cut(entry, "=")
		values[name] = value
	}
	lines = append(lines, "    "+key+":")
	for _, name := range sortedComposeKeys(values) {
		lines = append(lines, "      "+quoteComposeKey(name)+": "+quoteComposeScalar(values[name]))
	}
	return lines
}

// quoteComposeScalar 把值渲染成安全的 YAML 标量。
//
// 含冒号的值必须加引号：YAML 1.1 的六十进制数会把 `8080:80` 解析成数字。
// 含空格的值也加引号，与 docker-compose-gen 的输出风格保持一致。
func quoteComposeScalar(value string) string {
	if value == "" {
		return `""`
	}
	if !strings.ContainsAny(value, ": ") && isPlainYAMLScalar(value) {
		return value
	}
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`).Replace(value) + `"`
}

// quoteComposeKey 把映射键渲染成安全的 YAML 键。
func quoteComposeKey(key string) string {
	if dockerPlainKeyPattern.MatchString(key) {
		return key
	}
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(key) + `"`
}

// isPlainYAMLScalar 判断值能否不加引号地写成 YAML 纯量。
func isPlainYAMLScalar(value string) bool {
	if strings.TrimSpace(value) != value {
		return false
	}
	switch value[0] {
	case '-', '?', ':', ',', '[', ']', '{', '}', '#', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`':
		return false
	}
	if strings.Contains(value, ": ") || strings.Contains(value, " #") {
		return false
	}
	if strings.HasSuffix(value, ":") {
		return false
	}
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return false
		}
	}
	return true
}

// ── Compose → docker run ────────────────────────────────────────────────

// dockerComposeToRun 把 Compose 配置转成 docker run 命令。
func dockerComposeToRun(input string) (DockerRunConvertResult, error) {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(input), &document); err != nil {
		return DockerRunConvertResult{}, dockerYAMLParseError(err)
	}

	root := &document
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return DockerRunConvertResult{}, toolerr.Newf(toolerr.CodeInvalidYAML, map[string]any{
				"field":  "input",
				"reason": "empty document",
			})
		}
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return DockerRunConvertResult{}, toolerr.Newf(toolerr.CodeInvalidYAML, map[string]any{
			"field":  "input",
			"reason": "expected a mapping",
		})
	}

	services := yamlMappingValue(root, "services")
	if services == nil {
		return DockerRunConvertResult{}, tools.MissingField("services")
	}
	if services.Kind != yaml.MappingNode {
		return DockerRunConvertResult{}, toolerr.Newf(toolerr.CodeInvalidYAML, map[string]any{
			"field":  "services",
			"reason": "expected a mapping",
		})
	}
	if len(services.Content) == 0 {
		return DockerRunConvertResult{}, tools.MissingField("services")
	}
	if len(services.Content)/2 > dockerConvertMaxServices {
		return DockerRunConvertResult{}, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":    "services",
			"maxItems": dockerConvertMaxServices,
		})
	}

	result := DockerRunConvertResult{
		Direction:   DockerConvertComposeToRun,
		Services:    []DockerConvertService{},
		Unsupported: []string{},
	}
	blocks := []string{}
	for i := 0; i+1 < len(services.Content); i += 2 {
		name := services.Content[i].Value
		body := services.Content[i+1]
		if body.Kind != yaml.MappingNode {
			return DockerRunConvertResult{}, toolerr.Newf(toolerr.CodeInvalidYAML, map[string]any{
				"field":  "services",
				"reason": "expected a mapping",
			})
		}
		service, unsupported, err := composeServiceToRun(name, body)
		if err != nil {
			return DockerRunConvertResult{}, err
		}
		result.Services = append(result.Services, service)
		result.Unsupported = append(result.Unsupported, unsupported...)
		blocks = append(blocks, "# "+name+"\n"+service.RunCommand)
	}
	result.Output = strings.Join(blocks, "\n\n")
	return result, nil
}

// composeServiceToRun 把单个 Compose 服务转成 docker run 命令。
func composeServiceToRun(name string, body *yaml.Node) (DockerConvertService, []string, error) {
	service := DockerConvertService{Name: name}
	flags := []string{}
	unsupported := []string{}
	command := []string{}

	for i := 0; i+1 < len(body.Content); i += 2 {
		key := body.Content[i].Value
		value := body.Content[i+1]
		switch key {
		case "image":
			service.Image = value.Value
		case "container_name":
			flags = append(flags, "--name", shellQuote(value.Value))
		case "command":
			command = append(command, composeCommandArgs(value)...)
		case "entrypoint":
			flags = append(flags, "--entrypoint", shellQuote(composeScalarText(value)))
		case "ports":
			for _, port := range composePortArgs(value) {
				flags = append(flags, "-p", shellQuote(port))
			}
		case "environment":
			for _, item := range composeEnvArgs(value) {
				flags = append(flags, "-e", shellQuote(item))
			}
		case "volumes":
			for _, item := range composeStringList(value) {
				flags = append(flags, "-v", shellQuote(item))
			}
		case "labels":
			for _, item := range composeEnvArgs(value) {
				flags = append(flags, "-l", shellQuote(item))
			}
		case "restart":
			flags = append(flags, "--restart", shellQuote(value.Value))
		case "working_dir":
			flags = append(flags, "-w", shellQuote(value.Value))
		case "user":
			flags = append(flags, "-u", shellQuote(value.Value))
		case "hostname":
			flags = append(flags, "-h", shellQuote(value.Value))
		case "privileged":
			if composeIsTrue(value) {
				flags = append(flags, "--privileged")
			}
		case "cap_add":
			for _, item := range composeStringList(value) {
				flags = append(flags, "--cap-add", shellQuote(item))
			}
		case "cap_drop":
			for _, item := range composeStringList(value) {
				flags = append(flags, "--cap-drop", shellQuote(item))
			}
		case "tty":
			if composeIsTrue(value) {
				flags = append(flags, "-t")
			}
		case "stdin_open":
			if composeIsTrue(value) {
				flags = append(flags, "-i")
			}
		case "init":
			if composeIsTrue(value) {
				flags = append(flags, "--init")
			}
		case "read_only":
			if composeIsTrue(value) {
				flags = append(flags, "--read-only")
			}
		case "tmpfs":
			for _, item := range composeStringList(value) {
				flags = append(flags, "--tmpfs", shellQuote(item))
			}
		case "devices":
			for _, item := range composeStringList(value) {
				flags = append(flags, "--device", shellQuote(item))
			}
		case "mem_limit":
			flags = append(flags, "-m", shellQuote(value.Value))
		case "cpus":
			flags = append(flags, "--cpus", shellQuote(value.Value))
		case "network_mode":
			flags = append(flags, "--network", shellQuote(value.Value))
		case "networks":
			networks := composeStringList(value)
			if len(networks) > 0 {
				flags = append(flags, "--network", shellQuote(networks[0]))
			}
			if len(networks) > 1 {
				unsupported = append(unsupported, name+".networks")
			}
		default:
			unsupported = append(unsupported, name+"."+key)
		}
	}

	if service.Image == "" {
		return DockerConvertService{}, nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{
			"field": "services." + name + ".image",
		})
	}

	parts := []string{"docker", "run"}
	parts = append(parts, flags...)
	parts = append(parts, service.Image)
	for _, arg := range command {
		parts = append(parts, shellQuote(arg))
	}
	service.RunCommand = strings.Join(parts, " ")
	service.Exec = strings.Join(command, " ")
	return service, unsupported, nil
}

// yamlMappingValue 返回映射节点中指定键的值节点。
func yamlMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// composeStringList 把节点读成字符串列表。
func composeStringList(node *yaml.Node) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.SequenceNode:
		items := make([]string, 0, len(node.Content))
		for _, item := range node.Content {
			items = append(items, item.Value)
		}
		return items
	case yaml.ScalarNode:
		if node.Tag == "!!null" {
			return nil
		}
		return []string{node.Value}
	default:
		return nil
	}
}

// composeScalarText 返回标量节点的文本。
func composeScalarText(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return node.Value
}

// composeIsTrue 判断节点是否为真值。
func composeIsTrue(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode {
		return false
	}
	return strings.EqualFold(node.Value, "true")
}

// composeCommandArgs 把 command 节点读成参数列表。
func composeCommandArgs(node *yaml.Node) []string {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.SequenceNode {
		return composeStringList(node)
	}
	if node.Tag == "!!null" {
		return nil
	}
	words, err := splitShellWords(node.Value)
	if err != nil || len(words) == 0 {
		return []string{node.Value}
	}
	return words
}

// composeEnvArgs 把 environment / labels 节点读成 K=V 列表。
func composeEnvArgs(node *yaml.Node) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.MappingNode:
		items := make([]string, 0, len(node.Content)/2)
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := node.Content[i+1]
			if value.Kind == yaml.ScalarNode && value.Tag == "!!null" {
				items = append(items, key)
				continue
			}
			items = append(items, key+"="+value.Value)
		}
		return items
	case yaml.SequenceNode:
		return composeStringList(node)
	case yaml.ScalarNode:
		return []string{node.Value}
	default:
		return nil
	}
}

// composePortArgs 把 ports 节点读成端口映射列表。
func composePortArgs(node *yaml.Node) []string {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.SequenceNode:
		items := make([]string, 0, len(node.Content))
		for _, item := range node.Content {
			if text := composeLongPort(item); text != "" {
				items = append(items, text)
				continue
			}
			if item.Kind == yaml.MappingNode {
				// 既不是短格式也不是可识别的长格式：跳过，不产出空参数。
				continue
			}
			items = append(items, item.Value)
		}
		return items
	case yaml.ScalarNode:
		return []string{node.Value}
	default:
		return nil
	}
}

// composeLongPort 把长格式端口映射转成短格式。
func composeLongPort(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.MappingNode {
		return ""
	}
	published := yamlMappingValue(node, "published")
	target := yamlMappingValue(node, "target")
	if published == nil || target == nil {
		return ""
	}
	text := published.Value + ":" + target.Value
	if host := yamlMappingValue(node, "host_ip"); host != nil && host.Value != "" {
		text = host.Value + ":" + text
	}
	if protocol := yamlMappingValue(node, "protocol"); protocol != nil && protocol.Value != "" {
		text += "/" + protocol.Value
	}
	return text
}

// dockerYAMLParseError 把 YAML 解析错误映射成工具错误。
//
// 只回显行号等元信息，**不回显**用户输入。
func dockerYAMLParseError(err error) error {
	params := map[string]any{"field": "input"}
	if match := dockerYAMLLinePattern.FindStringSubmatch(err.Error()); match != nil {
		if line, convErr := strconv.Atoi(match[1]); convErr == nil {
			params["line"] = line
		}
	}
	return toolerr.Newf(toolerr.CodeInvalidYAML, params)
}

// ── 词法工具 ────────────────────────────────────────────────────────────

// splitShellWords 按 shell 词法切分命令行。
//
// 支持单引号、双引号与反斜杠转义；不处理变量展开与命令替换。
func splitShellWords(input string) ([]string, error) {
	var (
		words   []string
		current strings.Builder
		hasWord bool
		quote   byte
	)
	for i := 0; i < len(input); i++ {
		char := input[i]
		switch {
		case quote == '\'':
			if char == '\'' {
				quote = 0
				continue
			}
			current.WriteByte(char)
		case quote == '"':
			if char == '\\' && i+1 < len(input) {
				i++
				current.WriteByte(input[i])
				continue
			}
			if char == '"' {
				quote = 0
				continue
			}
			current.WriteByte(char)
		case char == '\'' || char == '"':
			quote = char
			hasWord = true
		case char == '\\' && i+1 < len(input):
			i++
			current.WriteByte(input[i])
			hasWord = true
		case char == ' ' || char == '\t' || char == '\n' || char == '\r':
			if hasWord {
				words = append(words, current.String())
				current.Reset()
				hasWord = false
			}
		default:
			current.WriteByte(char)
			hasWord = true
		}
	}
	if quote != 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "input",
			"reason": "unbalanced quote",
		})
	}
	if hasWord {
		words = append(words, current.String())
	}
	return words, nil
}

// shellQuote 用单引号包裹参数，避免被 shell 解释。
func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if dockerShellSafePattern.MatchString(value) {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
