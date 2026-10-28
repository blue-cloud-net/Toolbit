// Package system 实现系统类工具。
//
// 本包只做「解码输入 → 校验字段 → 计算 → 编码 DTO」，
// 不执行任何系统命令，也不访问文件系统。
package system

import (
	"strconv"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 目标 Shell。
const (
	// ShellBash 是 Bash / Zsh。
	ShellBash = "bash"
	// ShellSh 是 POSIX sh。
	ShellSh = "sh"
	// ShellPowershell 是 PowerShell。
	ShellPowershell = "powershell"
	// ShellCmd 是 Windows 命令提示符。
	ShellCmd = "cmd"
)

// Shells 返回受支持的目标 Shell（顺序稳定）。
func Shells() []string {
	return []string{ShellBash, ShellSh, ShellPowershell, ShellCmd}
}

// 认证方式。
const (
	// AuthNone 不带认证。
	AuthNone = "none"
	// AuthBasic 是 HTTP Basic。
	AuthBasic = "basic"
	// AuthBearer 是 Bearer Token。
	AuthBearer = "bearer"
)

// AuthTypes 返回受支持的认证方式（顺序稳定）。
func AuthTypes() []string {
	return []string{AuthNone, AuthBasic, AuthBearer}
}

// CurlGeneratorRequest 是 curl-generator 工具的请求。
type CurlGeneratorRequest struct {
	// Method 是 HTTP 方法；缺省 GET。
	Method string `json:"method"`
	// URL 是请求地址。
	URL string `json:"url"`
	// Headers 是请求头。
	Headers []CurlHeader `json:"headers"`
	// Body 是请求体。
	Body string `json:"body"`
	// BodyFormat 是请求体类型：json / form / raw。
	BodyFormat string `json:"bodyFormat"`
	// Auth 是认证方式。
	Auth string `json:"auth"`
	// Username 是 Basic 认证的用户名。
	Username string `json:"username"`
	// Password 是 Basic 认证的口令。
	Password string `json:"password"`
	// Token 是 Bearer Token。
	Token string `json:"token"`
	// Insecure 是否跳过 TLS 证书校验。
	Insecure bool `json:"insecure"`
	// FollowRedirects 是否跟随重定向。
	FollowRedirects bool `json:"followRedirects"`
	// Timeout 是超时秒数。
	Timeout int `json:"timeout"`
	// Shell 是目标 Shell；缺省 bash。
	Shell string `json:"shell"`
}

// CurlHeader 是一个请求头。
type CurlHeader struct {
	// Name 是头名称。
	Name string `json:"name"`
	// Value 是头值。
	Value string `json:"value"`
	// Sensitive 为真时值以占位符呈现，**不**内联明文。
	Sensitive bool `json:"sensitive"`
}

// 请求体类型。
const (
	// BodyFormatJSON 是 JSON 请求体。
	BodyFormatJSON = "json"
	// BodyFormatForm 是表单请求体。
	BodyFormatForm = "form"
	// BodyFormatRaw 是原始请求体。
	BodyFormatRaw = "raw"
)

// CurlGeneratorResult 是 curl-generator 工具的响应。
type CurlGeneratorResult struct {
	// Command 是生成的命令文本。
	Command string `json:"command"`
	// Shell 是目标 Shell。
	Shell string `json:"shell"`
	// Placeholders 是被占位符替代的凭据字段名。
	Placeholders []string `json:"placeholders,omitempty"`
}

// CurlGenerator 生成 cURL 命令。
//
// 口令与 Token **一律**以占位符呈现，**不**内联明文凭据：
// 命令通常会被贴进聊天窗口或存进脚本。
func CurlGenerator(req *CurlGeneratorRequest) (CurlGeneratorResult, error) {
	if err := req.Validate(); err != nil {
		return CurlGeneratorResult{}, err
	}

	shell := req.shell()
	parts := []string{}
	placeholders := []string{}

	parts = append(parts, "curl")

	if req.Method != "" && !strings.EqualFold(req.Method, "GET") {
		parts = append(parts, "-X "+req.Method)
	}

	for _, header := range req.Headers {
		if strings.TrimSpace(header.Name) == "" {
			continue
		}
		value := header.Value
		if header.Sensitive {
			placeholder := "<" + strings.ToUpper(strings.ReplaceAll(header.Name, "-", "_")) + ">"
			placeholders = append(placeholders, header.Name)
			value = placeholder
		}
		parts = append(parts, "-H "+quoteArg(shell, header.Name+": "+value))
	}

	switch req.Auth {
	case AuthBasic:
		parts = append(parts, "-u "+quoteArg(shell, req.Username+":"+placeholderOrValue(req.Password, "PASSWORD", &placeholders)))
	case AuthBearer:
		parts = append(parts, "-H "+quoteArg(shell, "Authorization: Bearer "+placeholderOrValue(req.Token, "TOKEN", &placeholders)))
	}

	if req.Insecure {
		parts = append(parts, "-k")
	}
	if req.FollowRedirects {
		parts = append(parts, "-L")
	}
	if req.Timeout > 0 {
		parts = append(parts, "--max-time "+strconv.Itoa(req.Timeout))
	}

	if req.Body != "" {
		switch req.BodyFormat {
		case BodyFormatForm:
			parts = append(parts, "--data-raw "+quoteArg(shell, req.Body))
		case BodyFormatRaw:
			parts = append(parts, "--data-binary "+quoteArg(shell, req.Body))
		default:
			parts = append(parts, "--data-raw "+quoteArg(shell, req.Body))
		}
	}

	parts = append(parts, quoteArg(shell, req.URL))

	command := strings.Join(parts, " ")

	return CurlGeneratorResult{
		Command:      command,
		Shell:        shell,
		Placeholders: placeholders,
	}, nil
}

// Validate 做字段级校验。
func (r *CurlGeneratorRequest) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return tools.MissingField("url")
	}
	switch r.Auth {
	case "", AuthNone, AuthBasic, AuthBearer:
	default:
		return tools.InvalidField("auth", AuthTypes(), r.Auth)
	}
	switch r.BodyFormat {
	case "", BodyFormatJSON, BodyFormatForm, BodyFormatRaw:
	default:
		return tools.InvalidField("bodyFormat",
			[]string{BodyFormatJSON, BodyFormatForm, BodyFormatRaw}, r.BodyFormat)
	}
	switch r.Shell {
	case "", ShellBash, ShellSh, ShellPowershell, ShellCmd:
	default:
		return tools.InvalidField("shell", Shells(), r.Shell)
	}
	if r.Timeout < 0 || r.Timeout > 3600 {
		return tools.InvalidField("timeout", []int{0, 3600}, r.Timeout)
	}
	if r.Auth == AuthBearer && r.Token == "" {
		return tools.MissingField("token")
	}
	if r.Auth == AuthBasic && r.Username == "" {
		return tools.MissingField("username")
	}
	return nil
}

// shell 返回目标 Shell；缺省为 bash。
func (r *CurlGeneratorRequest) shell() string {
	if r.Shell == "" {
		return ShellBash
	}
	return r.Shell
}

// placeholderOrValue 返回凭据的占位符，并把字段名登记到列表。
func placeholderOrValue(value string, name string, placeholders *[]string) string {
	if value == "" {
		return ""
	}
	*placeholders = append(*placeholders, name)
	return "<" + name + ">"
}

// quoteArg 按目标 Shell 的规则给参数加引号。
func quoteArg(shell string, arg string) string {
	switch shell {
	case ShellPowershell:
		return "'" + strings.ReplaceAll(arg, "'", "''") + "'"
	case ShellCmd:
		return "\"" + strings.ReplaceAll(arg, "\"", "\\\"") + "\""
	default:
		return "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}
}
