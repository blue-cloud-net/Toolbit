package system

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// DockerComposeRequest 是 docker-compose-gen 工具的请求。
type DockerComposeRequest struct {
	// ProjectName 是项目名；用作服务名前缀与容器名。
	ProjectName string `json:"projectName"`
	// Services 是服务定义列表。
	Services []ComposeService `json:"services"`
	// NetworkName 是自定义网络名；留空用项目名。
	NetworkName string `json:"networkName"`
	// IncludeVolumes 是否生成顶层 volumes 段。
	IncludeVolumes bool `json:"includeVolumes"`
	// IncludeNetwork 是否生成顶层 networks 段。
	IncludeNetwork bool `json:"includeNetwork"`
}

// ComposeService 是一个服务定义。
type ComposeService struct {
	// Name 是服务名。
	Name string `json:"name"`
	// Image 是镜像名。
	Image string `json:"image"`
	// Ports 是端口映射，格式 `宿主机:容器`。
	Ports []string `json:"ports"`
	// Environment 是环境变量键值。
	Environment map[string]string `json:"environment"`
	// Volumes 是卷挂载，格式 `宿主机:容器`。
	Volumes []string `json:"volumes"`
	// DependsOn 是依赖的服务名列表。
	DependsOn []string `json:"dependsOn"`
	// Restart 是重启策略；默认 always。
	Restart string `json:"restart"`
	// Command 是启动命令。
	Command string `json:"command"`
}

// DockerComposeResult 是 docker-compose-gen 工具的响应。
type DockerComposeResult struct {
	// Compose 是生成的 compose 文件内容。
	Compose string `json:"compose"`
	// ServiceCount 是服务数量。
	ServiceCount int `json:"serviceCount"`
	// Warnings 是配置提示。
	Warnings []string `json:"warnings"`
}

// 默认重启策略。
const composeDefaultRestart = "always"

// DockerComposeGen 生成 Docker Compose 配置。
//
// 这是**纯模板拼装**：不连接任何守护进程，也不拉取镜像。
func DockerComposeGen(_ context.Context, req *DockerComposeRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	projectName := req.ProjectName
	networkName := req.NetworkName
	if networkName == "" {
		networkName = projectName + "_default"
	}

	warnings := []string{}
	lines := []string{
		"# 由 Toolbit 生成的 Docker Compose 配置",
		"services:",
	}

	for _, service := range req.Services {
		lines = append(lines, "  "+service.Name+":")
		lines = append(lines, "    image: "+service.Image)
		if service.Restart != "" {
			lines = append(lines, "    restart: "+service.Restart)
		} else {
			lines = append(lines, "    restart: "+composeDefaultRestart)
		}
		if len(service.Ports) > 0 {
			lines = append(lines, "    ports:")
			for _, port := range service.Ports {
				lines = append(lines, "      - "+quoteComposeValue(port))
			}
		}
		if len(service.Environment) > 0 {
			lines = append(lines, "    environment:")
			for _, key := range sortedComposeKeys(service.Environment) {
				lines = append(lines, "      "+key+": "+quoteComposeValue(service.Environment[key]))
			}
		}
		if len(service.Volumes) > 0 {
			lines = append(lines, "    volumes:")
			for _, volume := range service.Volumes {
				lines = append(lines, "      - "+quoteComposeValue(volume))
			}
		}
		if len(service.DependsOn) > 0 {
			lines = append(lines, "    depends_on:")
			for _, dependency := range service.DependsOn {
				lines = append(lines, "      - "+dependency)
			}
		}
		if service.Command != "" {
			lines = append(lines, "    command: "+quoteComposeValue(service.Command))
		}
		if !req.IncludeNetwork {
			lines = append(lines, "    networks: ["+networkName+"]")
		}
	}

	if req.IncludeVolumes {
		lines = append(lines, "", "volumes:")
		for _, service := range req.Services {
			for _, volume := range service.Volumes {
				parts := strings.Split(volume, ":")
				if len(parts) >= 2 && !strings.HasPrefix(parts[1], "/") {
					warnings = append(warnings, "卷 "+volume+" 的容器侧路径不是绝对路径")
				}
			}
		}
		if len(lines) == 0 || lines[len(lines)-1] != "volumes:" {
			// 没有可列出的卷时给出空段，避免生成非法 YAML。
			lines = append(lines, "  # 未声明命名卷")
		}
	}

	if req.IncludeNetwork {
		lines = append(lines, "", "networks:", "  "+networkName+":", "    driver: bridge")
	}

	return tools.Result{Data: DockerComposeResult{
		Compose:      strings.Join(lines, "\n"),
		ServiceCount: len(req.Services),
		Warnings:     warnings,
	}}, nil
}

// quoteComposeValue 给值加引号。
//
// YAML 中含 `:`、空格或 `#` 的值必须加引号，否则解析出错。
func quoteComposeValue(value string) string {
	if value == "" {
		return "\"\""
	}
	needsQuote := false
	for _, c := range value {
		if c == ':' || c == ' ' || c == '#' || c == '"' || c == '\'' {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return value
	}
	return "\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""
}

// sortedComposeKeys 返回排序后的键列表。
func sortedComposeKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// Validate 做字段级校验。
func (r *DockerComposeRequest) Validate() error {
	if strings.TrimSpace(r.ProjectName) == "" {
		return tools.MissingField("projectName")
	}
	if len(r.ProjectName) > 64 {
		return tools.InvalidField("projectName", []int{0, 64}, len(r.ProjectName))
	}
	if len(r.Services) == 0 {
		return tools.MissingField("services")
	}
	if len(r.Services) > 32 {
		return tools.InvalidField("services", []int{0, 32}, len(r.Services))
	}
	for _, service := range r.Services {
		if strings.TrimSpace(service.Name) == "" {
			return tools.MissingField("services.name")
		}
		if strings.TrimSpace(service.Image) == "" {
			return tools.MissingField("services.image")
		}
	}
	return nil
}
