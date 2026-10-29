package data

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/yeqown/go-qrcode/v2"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// WiFi 加密方式。
const (
	// WiFiSecurityWPA 是 WPA / WPA2 / WPA3。
	WiFiSecurityWPA = "WPA"
	// WiFiSecurityWEP 是 WEP。
	WiFiSecurityWEP = "WEP"
	// WiFiSecurityNone 是无密码开放网络。
	WiFiSecurityNone = "nopass"
)

// WiFiSecurity 返回受支持的加密方式（顺序稳定）。
func WiFiSecurity() []string {
	return []string{WiFiSecurityWPA, WiFiSecurityWEP, WiFiSecurityNone}
}

// WiFiQRRequest 是 wifi-qr 工具的请求。
type WiFiQRRequest struct {
	// SSID 是网络名称。
	SSID string `json:"ssid"`
	// Password 是网络密码；开放网络留空。
	Password string `json:"password"`
	// Security 是加密方式；默认 WPA。
	Security string `json:"security"`
	// Hidden 是否是隐藏网络。
	Hidden bool `json:"hidden"`
	// Size 是二维码边长像素；默认 256。
	Size int `json:"size"`
}

// WiFiQRResult 是 wifi-qr 工具的响应。
type WiFiQRResult struct {
	// Payload 是二维码承载的文本内容。
	Payload string `json:"payload"`
	// PNG 是 PNG 图片的 Base64 编码。
	PNG string `json:"png"`
	// SVG 是 SVG 图片内容。
	SVG string `json:"svg"`
	// Size 是图片边长像素。
	Size int `json:"size"`
}

// 默认尺寸与上限。
const (
	defaultWiFiQRSize = 256
	maxWiFiQRSize     = 1024
)

// WiFiQR 生成访客 WiFi 的连接二维码。
//
// 二维码内容遵循 `WIFI:` 约定格式，各平台扫码后可直接连接。
func WiFiQR(_ context.Context, req *WiFiQRRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	security := req.Security
	if security == "" {
		security = WiFiSecurityWPA
	}
	size := req.Size
	if size == 0 {
		size = defaultWiFiQRSize
	}

	payload := buildWiFiPayload(req.SSID, req.Password, security, req.Hidden)

	encoder, err := qrcode.NewWith(payload, qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium))
	if err != nil {
		return tools.Result{}, tools.InvalidField("ssid", "encodable SSID", "encode error")
	}

	pngBytes, err := renderWiFiPNG(encoder, size)
	if err != nil {
		return tools.Result{}, err
	}

	svgBytes, err := renderWiFiSVG(encoder, size)
	if err != nil {
		return tools.Result{}, err
	}

	return tools.Result{Data: WiFiQRResult{
		Payload: payload,
		PNG:     base64Std(pngBytes),
		SVG:     string(svgBytes),
		Size:    size,
	}}, nil
}

// buildWiFiPayload 拼装 `WIFI:` 格式的二维码内容。
//
// 特殊字符按规范转义：分号、逗号、冒号、反斜杠与空格。
func buildWiFiPayload(ssid, password, security string, hidden bool) string {
	payload := "WIFI:" +
		"T:" + security + ";" +
		"S:" + escapeWiFiValue(ssid) + ";" +
		"P:" + escapeWiFiValue(password) + ";"
	if hidden {
		payload += "H:true;"
	}
	return payload
}

// escapeWiFiValue 转义 WiFi 格式中的特殊字符。
func escapeWiFiValue(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		";", "\\;",
		",", "\\,",
		":", "\\:",
		"\"", "\\\"",
		" ", "\\ ",
	)
	return replacer.Replace(value)
}

// renderWiFiPNG 渲染 PNG 图片。
func renderWiFiPNG(encoder *qrcode.QRCode, size int) ([]byte, error) {
	matrix, err := qrMatrix(encoder)
	if err != nil {
		return nil, err
	}

	image := renderQRImage(matrix, size)
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image); err != nil {
		return nil, tools.InvalidField("ssid", "encodable SSID", "render error")
	}
	return buffer.Bytes(), nil
}

// renderWiFiSVG 渲染 SVG 图片。
func renderWiFiSVG(encoder *qrcode.QRCode, size int) ([]byte, error) {
	matrix, err := qrMatrix(encoder)
	if err != nil {
		return nil, err
	}

	modules := len(matrix)
	if modules == 0 {
		return nil, tools.InvalidField("ssid", "encodable SSID", "encode error")
	}

	moduleSize := size / modules
	if moduleSize < 1 {
		moduleSize = 1
	}
	total := moduleSize * modules

	var buffer strings.Builder
	fmt.Fprintf(&buffer,
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">`,
		total, total, total, total)
	buffer.WriteString(`<rect width="100%" height="100%" fill="#ffffff"/>`)
	buffer.WriteString(`<path fill="#000000" d="`)
	for row := 0; row < modules; row++ {
		for col := 0; col < modules; col++ {
			if matrix[row][col] {
				fmt.Fprintf(&buffer, "M%d %dh1v1h-1z", col*moduleSize, row*moduleSize)
			}
		}
	}
	buffer.WriteString(`"/></svg>`)
	return []byte(buffer.String()), nil
}

// renderQRImage 把模块矩阵渲染成图片。
func renderQRImage(matrix [][]bool, size int) image.Image {
	modules := len(matrix)
	if modules == 0 {
		return image.NewRGBA(image.Rect(0, 0, size, size))
	}

	moduleSize := size / modules
	if moduleSize < 1 {
		moduleSize = 1
	}
	total := moduleSize * modules

	result := image.NewRGBA(image.Rect(0, 0, total, total))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	for y := 0; y < total; y++ {
		for x := 0; x < total; x++ {
			row := y / moduleSize
			col := x / moduleSize
			if row < modules && col < modules && matrix[row][col] {
				result.Set(x, y, black)
				continue
			}
			result.Set(x, y, white)
		}
	}
	return result
}

// base64Std 把字节编码成标准 Base64。
func base64Std(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Validate 做字段级校验。
func (r *WiFiQRRequest) Validate() error {
	if strings.TrimSpace(r.SSID) == "" {
		return tools.MissingField("ssid")
	}
	if len(r.SSID) > 32 {
		return tools.InvalidField("ssid", []int{0, 32}, len(r.SSID))
	}
	if len(r.Password) > 128 {
		return tools.InvalidField("password", []int{0, 128}, len(r.Password))
	}
	security := r.Security
	if security == "" {
		security = WiFiSecurityWPA
	}
	if !containsString(WiFiSecurity(), security) {
		return tools.InvalidField("security", WiFiSecurity(), security)
	}
	if r.Size < 0 || r.Size > maxWiFiQRSize {
		return tools.InvalidField("size", []int{0, maxWiFiQRSize}, r.Size)
	}
	return nil
}
