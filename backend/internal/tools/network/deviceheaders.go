package network

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 请求头生成模式。
const (
	// DeviceModeGenerate 是生成请求头。
	DeviceModeGenerate = "generate"
	// DeviceModeParseUA 是解析 User-Agent。
	DeviceModeParseUA = "parseUa"
)

// DeviceMode 返回受支持的模式（顺序稳定）。
func DeviceMode() []string {
	return []string{DeviceModeGenerate, DeviceModeParseUA}
}

// 受支持的平台。
const (
	// DevicePlatformChrome 是 Chrome 桌面版。
	DevicePlatformChrome = "chrome"
	// DevicePlatformFirefox 是 Firefox 桌面版。
	DevicePlatformFirefox = "firefox"
	// DevicePlatformSafari 是 Safari 桌面版。
	DevicePlatformSafari = "safari"
	// DevicePlatformEdge 是 Edge 桌面版。
	DevicePlatformEdge = "edge"
	// DevicePlatformAndroid 是 Android Chrome。
	DevicePlatformAndroid = "android"
	// DevicePlatformIOS 是 iOS Safari。
	DevicePlatformIOS = "ios"
	// DevicePlatformCurl 是 curl。
	DevicePlatformCurl = "curl"
)

// DevicePlatform 返回受支持的平台（顺序稳定）。
func DevicePlatform() []string {
	return []string{
		DevicePlatformChrome, DevicePlatformFirefox, DevicePlatformSafari,
		DevicePlatformEdge, DevicePlatformAndroid, DevicePlatformIOS,
		DevicePlatformCurl,
	}
}

// DeviceHeadersRequest 是 device-headers 工具的请求。
type DeviceHeadersRequest struct {
	// Mode 是模式：generate / parseUa；默认 generate。
	Mode string `json:"mode"`
	// Platform 是平台；仅 generate 模式使用。
	Platform string `json:"platform"`
	// Version 是浏览器版本号；留空用内置默认值。
	Version string `json:"version"`
	// OS 是操作系统名；留空用平台默认值。
	OS string `json:"os"`
	// UserAgent 是待解析的 UA 串；仅 parseUa 模式使用。
	UserAgent string `json:"userAgent"`
	// AcceptLanguage 是 Accept-Language 头。
	AcceptLanguage string `json:"acceptLanguage"`
	// Referer 是 Referer 头。
	Referer string `json:"referer"`
	// CustomHeaders 是额外的自定义请求头。
	CustomHeaders map[string]string `json:"customHeaders"`
}

// DeviceHeader 是单个请求头。
type DeviceHeader struct {
	// Name 是请求头名称。
	Name string `json:"name"`
	// Value 是请求头取值。
	Value string `json:"value"`
}

// UAInfo 是解析出的 UA 信息。
type UAInfo struct {
	// Browser 是浏览器名称。
	Browser string `json:"browser"`
	// BrowserVersion 是浏览器版本。
	BrowserVersion string `json:"browserVersion"`
	// OS 是操作系统名称。
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

// DeviceHeadersResult 是 device-headers 工具的响应。
type DeviceHeadersResult struct {
	// Mode 是实际使用的模式。
	Mode string `json:"mode"`
	// Headers 是生成的请求头。
	Headers []DeviceHeader `json:"headers"`
	// UserAgent 是生成或解析的 UA 串。
	UserAgent string `json:"userAgent"`
	// Parsed 是解析出的 UA 信息；generate 模式为 nil。
	Parsed *UAInfo `json:"parsed"`
}

// maxUserAgentLength 是 UA 串的长度上限。
const maxUserAgentLength = 1024

// maxCustomHeaders 是自定义请求头的数量上限。
const maxCustomHeaders = 20

// 各平台的默认版本与操作系统。
var deviceDefaults = map[string]struct {
	version string
	os      string
}{
	DevicePlatformChrome:  {version: "131.0.0.0", os: "Windows NT 10.0; Win64; x64"},
	DevicePlatformFirefox: {version: "133.0", os: "Windows NT 10.0; Win64; x64"},
	DevicePlatformSafari:  {version: "18.1", os: "Mac OS X 10_15_7"},
	DevicePlatformEdge:    {version: "131.0.0.0", os: "Windows NT 10.0; Win64; x64"},
	DevicePlatformAndroid: {version: "131.0.0.0", os: "Android 15; Pixel 9"},
	DevicePlatformIOS:     {version: "18.1", os: "iPhone; iOS 18_1"},
	DevicePlatformCurl:    {version: "8.9.1", os: ""},
}

// DeviceHeaders 生成客户端请求头或解析 User-Agent。
//
// 首版只支持**内置**的常见平台与用户自定义填写，
// **不**引入任何在线 UA 查询（外部依赖会带来出网与隐私问题）。
func DeviceHeaders(_ context.Context, req *DeviceHeadersRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = DeviceModeGenerate
	}

	if mode == DeviceModeParseUA {
		return parseUserAgent(req)
	}
	return generateDeviceHeaders(req)
}

// generateDeviceHeaders 生成一组请求头。
func generateDeviceHeaders(req *DeviceHeadersRequest) (tools.Result, error) {
	platform := req.Platform
	if platform == "" {
		platform = DevicePlatformChrome
	}
	defaults, ok := deviceDefaults[platform]
	if !ok {
		return tools.Result{}, tools.InvalidField("platform", DevicePlatform(), platform)
	}

	version := req.Version
	if version == "" {
		version = defaults.version
	}
	osName := req.OS
	if osName == "" {
		osName = defaults.os
	}

	userAgent := buildUserAgent(platform, version, osName)

	language := req.AcceptLanguage
	if language == "" {
		language = "zh-CN,zh;q=0.9,en;q=0.8"
	}

	headers := []DeviceHeader{
		{Name: "User-Agent", Value: userAgent},
		{Name: "Accept", Value: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		{Name: "Accept-Language", Value: language},
		{Name: "Accept-Encoding", Value: "gzip, deflate, br"},
		{Name: "Upgrade-Insecure-Requests", Value: "1"},
	}
	if req.Referer != "" {
		headers = append(headers, DeviceHeader{Name: "Referer", Value: req.Referer})
	}
	for name, value := range req.CustomHeaders {
		headers = append(headers, DeviceHeader{Name: name, Value: value})
	}

	return tools.Result{Data: DeviceHeadersResult{
		Mode:      DeviceModeGenerate,
		Headers:   headers,
		UserAgent: userAgent,
		Parsed:    nil,
	}}, nil
}

// buildUserAgent 拼装 UA 串。
func buildUserAgent(platform, version, osName string) string {
	switch platform {
	case DevicePlatformCurl:
		return "curl/" + version
	case DevicePlatformFirefox:
		return "Mozilla/5.0 (" + osName + "; rv:" + version + ") Gecko/20100101 Firefox/" + version
	case DevicePlatformSafari:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/605.1.15 (KHTML, like Gecko) Version/" + version + " Safari/605.1.15"
	case DevicePlatformEdge:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36 Edg/" + version
	case DevicePlatformAndroid:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Mobile Safari/537.36"
	case DevicePlatformIOS:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/605.1.15 (KHTML, like Gecko) Version/" + version + " Mobile/15E148 Safari/604.1"
	default:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36"
	}
}

// parseUserAgent 解析 UA 串。
func parseUserAgent(req *DeviceHeadersRequest) (tools.Result, error) {
	ua := strings.TrimSpace(req.UserAgent)
	info := parseUAString(ua)

	return tools.Result{Data: DeviceHeadersResult{
		Mode:      DeviceModeParseUA,
		Headers:   []DeviceHeader{},
		UserAgent: ua,
		Parsed:    &info,
	}}, nil
}

// 常见爬虫标识。
var botKeywords = []string{"bot", "spider", "crawler", "slurp", "curl", "wget", "python-requests", "headlesschrome"}

// parseUAString 从 UA 串中提取浏览器、操作系统与设备类型。
func parseUAString(ua string) UAInfo {
	info := UAInfo{
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
	case strings.Contains(ua, "Mac OS X"):
		info.OS = "macOS"
		info.OSVersion = extractVersionAfter(ua, "Mac OS X ")
	case strings.Contains(ua, "Android"):
		info.OS = "Android"
		info.OSVersion = extractVersionAfter(ua, "Android ")
	case strings.Contains(ua, "iPhone"):
		info.OS = "iOS"
		info.OSVersion = extractVersionAfter(ua, "iPhone OS ")
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
	marker := "Windows NT "
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

// Validate 做字段级校验。
func (r *DeviceHeadersRequest) Validate() error {
	mode := r.Mode
	if mode == "" {
		mode = DeviceModeGenerate
	}
	if !containsString(DeviceMode(), mode) {
		return tools.InvalidField("mode", DeviceMode(), mode)
	}

	if mode == DeviceModeParseUA {
		if strings.TrimSpace(r.UserAgent) == "" {
			return tools.MissingField("userAgent")
		}
		if len(r.UserAgent) > maxUserAgentLength {
			return tools.InvalidField("userAgent", []int{0, maxUserAgentLength}, len(r.UserAgent))
		}
		return nil
	}

	if r.Platform != "" && !containsString(DevicePlatform(), r.Platform) {
		return tools.InvalidField("platform", DevicePlatform(), r.Platform)
	}
	if len(r.Version) > 32 {
		return tools.InvalidField("version", []int{0, 32}, len(r.Version))
	}
	if len(r.OS) > 128 {
		return tools.InvalidField("os", []int{0, 128}, len(r.OS))
	}
	if len(r.AcceptLanguage) > 128 {
		return tools.InvalidField("acceptLanguage", []int{0, 128}, len(r.AcceptLanguage))
	}
	if len(r.Referer) > 2048 {
		return tools.InvalidField("referer", []int{0, 2048}, len(r.Referer))
	}
	if len(r.CustomHeaders) > maxCustomHeaders {
		return tools.InvalidField("customHeaders", []int{0, maxCustomHeaders}, len(r.CustomHeaders))
	}
	for name := range r.CustomHeaders {
		if !isValidHeaderName(name) {
			return tools.InvalidField("customHeaders", "valid header name", "invalid format")
		}
	}
	return nil
}

// isValidHeaderName 判断请求头名是否合法。
//
// RFC 9110 的 token 字符集：不允许空格与分隔符。
func isValidHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", c):
		default:
			return false
		}
	}
	return true
}
