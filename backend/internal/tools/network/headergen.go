package network

import (
	"context"
	"sort"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 受支持的平台。
const (
	// HeaderPlatformChrome 是 Chrome 桌面版。
	HeaderPlatformChrome = "chrome"
	// HeaderPlatformFirefox 是 Firefox 桌面版。
	HeaderPlatformFirefox = "firefox"
	// HeaderPlatformSafari 是 Safari 桌面版。
	HeaderPlatformSafari = "safari"
	// HeaderPlatformEdge 是 Edge 桌面版。
	HeaderPlatformEdge = "edge"
	// HeaderPlatformAndroid 是 Android Chrome。
	HeaderPlatformAndroid = "android"
	// HeaderPlatformIOS 是 iOS Safari。
	HeaderPlatformIOS = "ios"
	// HeaderPlatformCurl 是 curl。
	HeaderPlatformCurl = "curl"
)

// HeaderPlatforms 返回受支持的平台（顺序稳定）。
func HeaderPlatforms() []string {
	return []string{
		HeaderPlatformChrome, HeaderPlatformFirefox, HeaderPlatformSafari,
		HeaderPlatformEdge, HeaderPlatformAndroid, HeaderPlatformIOS,
		HeaderPlatformCurl,
	}
}

// HeaderGeneratorRequest 是 ua-header-gen 工具的请求。
type HeaderGeneratorRequest struct {
	// Platform 是平台；留空用 chrome。
	Platform string `json:"platform"`
	// Version 是浏览器版本号；留空用内置默认值。
	Version string `json:"version"`
	// OS 是操作系统名；留空用平台默认值。
	OS string `json:"os"`
	// AcceptLanguage 是 Accept-Language 头；留空用默认值。
	AcceptLanguage string `json:"acceptLanguage"`
	// Referer 是 Referer 头；留空则不输出该头。
	Referer string `json:"referer"`
	// CustomHeaders 是额外的自定义请求头。
	CustomHeaders map[string]string `json:"customHeaders"`
}

// HeaderEntry 是单个请求头。
type HeaderEntry struct {
	// Name 是请求头名称。
	Name string `json:"name"`
	// Value 是请求头取值。
	Value string `json:"value"`
}

// HeaderGeneratorResult 是 ua-header-gen 工具的响应。
type HeaderGeneratorResult struct {
	// Headers 是生成的请求头。
	Headers []HeaderEntry `json:"headers"`
	// UserAgent 是生成的 UA 串。
	UserAgent string `json:"userAgent"`
}

// maxCustomHeaders 是自定义请求头的数量上限。
const maxCustomHeaders = 20

// 各平台的默认版本与操作系统。
var headerDefaults = map[string]struct {
	version string
	os      string
}{
	HeaderPlatformChrome:  {version: "131.0.0.0", os: "Windows NT 10.0; Win64; x64"},
	HeaderPlatformFirefox: {version: "133.0", os: "Windows NT 10.0; Win64; x64"},
	HeaderPlatformSafari:  {version: "18.1", os: "Mac OS X 10_15_7"},
	HeaderPlatformEdge:    {version: "131.0.0.0", os: "Windows NT 10.0; Win64; x64"},
	HeaderPlatformAndroid: {version: "131.0.0.0", os: "Android 15; Pixel 9"},
	HeaderPlatformIOS:     {version: "18.1", os: "iPhone; iOS 18_1"},
	HeaderPlatformCurl:    {version: "8.9.1", os: ""},
}

// HeaderGenerator 按平台生成一组客户端请求头。
//
// 首版只支持**内置**的常见平台与用户自定义填写，
// **不**引入任何在线 UA 查询（外部依赖会带来出网与隐私问题）。
func HeaderGenerator(_ context.Context, req *HeaderGeneratorRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	platform := req.Platform
	if platform == "" {
		platform = HeaderPlatformChrome
	}
	defaults := headerDefaults[platform]

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

	headers := []HeaderEntry{
		{Name: "User-Agent", Value: userAgent},
		{Name: "Accept", Value: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		{Name: "Accept-Language", Value: language},
		{Name: "Accept-Encoding", Value: "gzip, deflate, br"},
		{Name: "Upgrade-Insecure-Requests", Value: "1"},
	}
	if req.Referer != "" {
		headers = append(headers, HeaderEntry{Name: "Referer", Value: req.Referer})
	}
	for _, name := range sortedHeaderNames(req.CustomHeaders) {
		headers = append(headers, HeaderEntry{Name: name, Value: req.CustomHeaders[name]})
	}

	return tools.Result{Data: HeaderGeneratorResult{
		Headers:   headers,
		UserAgent: userAgent,
	}}, nil
}

// sortedHeaderNames 返回自定义请求头的名称（按字典序，顺序稳定）。
//
// map 的遍历顺序随机，排序后同一组自定义请求头每次输出都一致。
func sortedHeaderNames(headers map[string]string) []string {
	if len(headers) == 0 {
		return nil
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// buildUserAgent 拼装 UA 串。
func buildUserAgent(platform, version, osName string) string {
	switch platform {
	case HeaderPlatformCurl:
		return "curl/" + version
	case HeaderPlatformFirefox:
		return "Mozilla/5.0 (" + osName + "; rv:" + version + ") Gecko/20100101 Firefox/" + version
	case HeaderPlatformSafari:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/605.1.15 (KHTML, like Gecko) Version/" + version + " Safari/605.1.15"
	case HeaderPlatformEdge:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36 Edg/" + version
	case HeaderPlatformAndroid:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Mobile Safari/537.36"
	case HeaderPlatformIOS:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/605.1.15 (KHTML, like Gecko) Version/" + version + " Mobile/15E148 Safari/604.1"
	default:
		return "Mozilla/5.0 (" + osName + ") AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + " Safari/537.36"
	}
}

// Validate 做字段级校验。
func (r *HeaderGeneratorRequest) Validate() error {
	if r.Platform != "" && !containsString(HeaderPlatforms(), r.Platform) {
		return tools.InvalidField("platform", HeaderPlatforms(), r.Platform)
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
