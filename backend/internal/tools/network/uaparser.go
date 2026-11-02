package network

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// UAParserRequest 是 ua-parser 工具的请求。
type UAParserRequest struct {
	// UserAgent 是待解析的 User-Agent 串。
	UserAgent string `json:"userAgent"`
}

// UAParserResult 是 ua-parser 工具的响应。
type UAParserResult struct {
	// UserAgent 是规范化后的 UA 串。
	UserAgent string `json:"userAgent"`
	// Browser 是浏览器名称；无法识别时为 `unknown`。
	Browser string `json:"browser"`
	// BrowserVersion 是浏览器版本。
	BrowserVersion string `json:"browserVersion"`
	// OS 是操作系统名称；无法识别时为 `unknown`。
	OS string `json:"os"`
	// OSVersion 是操作系统版本。
	OSVersion string `json:"osVersion"`
	// DeviceType 是设备类型：desktop / mobile / tablet / bot / unknown。
	DeviceType string `json:"deviceType"`
	// IsBot 是否是爬虫。
	IsBot bool `json:"isBot"`
	// Engine 是渲染引擎。
	Engine string `json:"engine"`
}

// maxUserAgentLength 是 UA 串的长度上限。
const maxUserAgentLength = 1024

// 常见爬虫标识。
var botKeywords = []string{
	"bot", "spider", "crawler", "slurp", "curl", "wget",
	"python-requests", "headlesschrome",
}

// UAParser 解析 User-Agent 串，拆出浏览器、系统与设备类型。
//
// 只做**结构化识别**：不查在线数据库、不做指纹追踪，
// 识别规则是内置的关键词与前缀匹配。
func UAParser(_ context.Context, req *UAParserRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	ua := strings.TrimSpace(req.UserAgent)

	return tools.Result{Data: parseUAString(ua)}, nil
}

// parseUAString 从 UA 串中提取浏览器、操作系统与设备类型。
func parseUAString(ua string) UAParserResult {
	info := UAParserResult{
		UserAgent:  ua,
		Browser:    "unknown",
		OS:         "unknown",
		DeviceType: "unknown",
		IsBot:      false,
		Engine:     "",
	}

	lowered := strings.ToLower(ua)
	for _, keyword := range botKeywords {
		if strings.Contains(lowered, keyword) {
			info.IsBot = true
			info.DeviceType = "bot"
			break
		}
	}

	switch {
	case strings.Contains(ua, "Edg/"):
		info.Browser = "Edge"
		info.BrowserVersion = extractVersionAfter(ua, "Edg/")
	case strings.Contains(ua, "Firefox/"):
		info.Browser = "Firefox"
		info.BrowserVersion = extractVersionAfter(ua, "Firefox/")
	case strings.Contains(ua, "Chrome/"):
		info.Browser = "Chrome"
		info.BrowserVersion = extractVersionAfter(ua, "Chrome/")
	case strings.Contains(ua, "Version/"):
		info.Browser = "Safari"
		info.BrowserVersion = extractVersionAfter(ua, "Version/")
	}

	switch {
	case strings.Contains(ua, "Windows NT"):
		info.OS = "Windows"
		info.OSVersion = extractWindowsVersion(ua)
	case strings.Contains(ua, "iPhone"), strings.Contains(ua, "iPad"):
		// iOS 的 UA 串里含 `like Mac OS X`，必须排在 macOS 之前判定。
		info.OS = "iOS"
		info.OSVersion = extractIOSVersion(ua)
	case strings.Contains(ua, "Android"):
		info.OS = "Android"
		info.OSVersion = extractVersionAfter(ua, "Android ")
	case strings.Contains(ua, "Mac OS X"):
		info.OS = "macOS"
		info.OSVersion = extractVersionAfter(ua, "Mac OS X ")
	case strings.Contains(ua, "Linux"):
		info.OS = "Linux"
	}

	switch {
	case strings.Contains(ua, "iPad"), strings.Contains(ua, "Tablet"):
		info.DeviceType = "tablet"
	case strings.Contains(ua, "Mobile"), strings.Contains(ua, "iPhone"), strings.Contains(ua, "Android"):
		if info.DeviceType != "bot" {
			info.DeviceType = "mobile"
		}
	case strings.Contains(ua, "Windows"), strings.Contains(ua, "Mac OS X"), strings.Contains(ua, "Linux"):
		if info.DeviceType != "bot" {
			info.DeviceType = "desktop"
		}
	}

	switch {
	case strings.Contains(ua, "Gecko/") && strings.Contains(ua, "Firefox"):
		info.Engine = "Gecko"
	case strings.Contains(ua, "AppleWebKit"):
		info.Engine = "Blink/WebKit"
	}

	return info
}

// extractVersionAfter 抽取前缀后面的版本号。
func extractVersionAfter(text, prefix string) string {
	idx := strings.Index(text, prefix)
	if idx < 0 {
		return ""
	}
	rest := text[idx+len(prefix):]
	end := 0
	for end < len(rest) {
		c := rest[end]
		if (c < '0' || c > '9') && c != '.' {
			break
		}
		end++
	}
	return strings.TrimSuffix(rest[:end], ".")
}

// extractWindowsVersion 抽取 Windows 版本号。
func extractWindowsVersion(ua string) string {
	const marker = "Windows NT "
	idx := strings.Index(ua, marker)
	if idx < 0 {
		return ""
	}
	rest := ua[idx+len(marker):]
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	digits := rest[:end]
	switch digits {
	case "10.0":
		return "10"
	case "6.3":
		return "8.1"
	case "6.2":
		return "8"
	case "6.1":
		return "7"
	default:
		return digits
	}
}

// extractIOSVersion 抽取 iOS 版本号。
//
// iPhone 的 UA 串用 `iPhone OS 18_1`，iPad 用 `CPU OS 18_1`，两种都要认；
// 版本号里的分隔符是下划线（`18_1`），与其它平台的点号不同。
func extractIOSVersion(ua string) string {
	for _, marker := range []string{"iPhone OS ", "CPU OS "} {
		idx := strings.Index(ua, marker)
		if idx < 0 {
			continue
		}
		rest := ua[idx+len(marker):]
		end := 0
		for end < len(rest) {
			c := rest[end]
			if (c < '0' || c > '9') && c != '.' && c != '_' {
				break
			}
			end++
		}
		if version := strings.TrimSuffix(rest[:end], "."); version != "" {
			return version
		}
	}
	return ""
}

// Validate 做字段级校验。
func (r *UAParserRequest) Validate() error {
	if strings.TrimSpace(r.UserAgent) == "" {
		return tools.MissingField("userAgent")
	}
	if len(r.UserAgent) > maxUserAgentLength {
		return tools.InvalidField("userAgent", []int{0, maxUserAgentLength}, len(r.UserAgent))
	}
	return nil
}
