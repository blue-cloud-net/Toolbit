package formatting

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 颜色格式。
const (
	// ColorFormatHex 是十六进制格式。
	ColorFormatHex = "hex"
	// ColorFormatRGB 是 RGB 格式。
	ColorFormatRGB = "rgb"
	// ColorFormatHSL 是 HSL 格式。
	ColorFormatHSL = "hsl"
	// ColorFormatOKLCH 是 OKLCH 格式。
	ColorFormatOKLCH = "oklch"
)

// ColorFormat 返回受支持的颜色格式（顺序稳定）。
func ColorFormat() []string {
	return []string{ColorFormatHex, ColorFormatRGB, ColorFormatHSL, ColorFormatOKLCH}
}

// RGBColor 是 sRGB 颜色分量。
type RGBColor struct {
	// R 是红分量，取值 0~255。
	R int `json:"r"`
	// G 是绿分量，取值 0~255。
	G int `json:"g"`
	// B 是蓝分量，取值 0~255。
	B int `json:"b"`
	// A 是透明度，取值 0~1。
	A float64 `json:"a"`
}

// ColorConvertRequest 是 color-convert 工具的请求。
type ColorConvertRequest struct {
	// Input 是颜色值。
	Input string `json:"input"`
	// InputFormat 是输入格式；空表示自动识别。
	InputFormat string `json:"inputFormat"`
	// OutputFormats 是要输出的格式列表；空表示全部。
	OutputFormats []string `json:"outputFormats"`
}

// ColorConvertResult 是 color-convert 工具的响应。
type ColorConvertResult struct {
	// Input 是规范化后的输入值。
	Input string `json:"input"`
	// InputFormat 是识别出的输入格式。
	InputFormat string `json:"inputFormat"`
	// Hex 是十六进制表示。
	Hex string `json:"hex"`
	// RGB 是 RGB 表示。
	RGB string `json:"rgb"`
	// HSL 是 HSL 表示。
	HSL string `json:"hsl"`
	// OKLCH 是 OKLCH 表示。
	OKLCH string `json:"oklch"`
	// ContrastOnWhite 是与白色的对比度。
	ContrastOnWhite float64 `json:"contrastOnWhite"`
	// ContrastOnBlack 是与黑色的对比度。
	ContrastOnBlack float64 `json:"contrastOnBlack"`
	// IsDark 是否是深色。
	IsDark bool `json:"isDark"`
}

// ColorConvert 转换颜色格式并计算对比度。
//
// 这是**纯数学计算**：不涉及密码学，也不联网。
func ColorConvert(_ context.Context, req *ColorConvertRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	input := strings.TrimSpace(req.Input)
	format := req.InputFormat
	if format == "" {
		format = detectColorFormat(input)
	}

	color, ok := parseColor(input, format)
	if !ok {
		return tools.Result{}, tools.InvalidField("input", ColorFormat(), "unrecognized format")
	}

	result := ColorConvertResult{
		Input:       input,
		InputFormat: format,
		Hex:         formatHex(color),
		RGB:         formatRGB(color),
		HSL:         formatHSL(color),
		OKLCH:       formatOKLCH(color),
	}
	result.ContrastOnWhite = contrastRatio(color, RGBColor{R: 255, G: 255, B: 255, A: 1})
	result.ContrastOnBlack = contrastRatio(color, RGBColor{R: 0, G: 0, B: 0, A: 1})
	result.IsDark = relativeLuminance(color) < 0.5

	return tools.Result{Data: result}, nil
}

// detectColorFormat 识别颜色格式。
func detectColorFormat(input string) string {
	lowered := strings.ToLower(strings.TrimSpace(input))
	switch {
	case strings.HasPrefix(lowered, "#"):
		return ColorFormatHex
	case strings.HasPrefix(lowered, "rgba("), strings.HasPrefix(lowered, "rgb("):
		return ColorFormatRGB
	case strings.HasPrefix(lowered, "hsla("), strings.HasPrefix(lowered, "hsl("):
		return ColorFormatHSL
	case strings.HasPrefix(lowered, "oklch("):
		return ColorFormatOKLCH
	default:
		return ColorFormatHex
	}
}

// parseColor 解析颜色值。
func parseColor(input, format string) (RGBColor, bool) {
	switch format {
	case ColorFormatHex:
		return parseHexColor(input)
	case ColorFormatRGB:
		return parseRGBColor(input)
	case ColorFormatHSL:
		return parseHSLColor(input)
	case ColorFormatOKLCH:
		return parseOKLCHColor(input)
	default:
		return RGBColor{}, false
	}
}

// parseHexColor 解析十六进制颜色。
func parseHexColor(input string) (RGBColor, bool) {
	value := strings.TrimSpace(input)
	value = strings.TrimPrefix(value, "#")
	if len(value) == 3 {
		expanded := ""
		for _, c := range value {
			expanded += string(c) + string(c)
		}
		value = expanded
	}
	if len(value) != 6 && len(value) != 8 {
		return RGBColor{}, false
	}
	parsed, err := strconv.ParseUint(value[0:2], 16, 8)
	if err != nil {
		return RGBColor{}, false
	}
	r := int(parsed)
	parsed, err = strconv.ParseUint(value[2:4], 16, 8)
	if err != nil {
		return RGBColor{}, false
	}
	g := int(parsed)
	parsed, err = strconv.ParseUint(value[4:6], 16, 8)
	if err != nil {
		return RGBColor{}, false
	}
	b := int(parsed)
	alpha := 1.0
	if len(value) == 8 {
		parsed, err = strconv.ParseUint(value[6:8], 16, 8)
		if err != nil {
			return RGBColor{}, false
		}
		alpha = float64(parsed) / 255
	}
	return RGBColor{R: r, G: g, B: b, A: alpha}, true
}

// parseRGBColor 解析 RGB 颜色。
func parseRGBColor(input string) (RGBColor, bool) {
	values, ok := parseColorFunction(input, "rgb")
	if !ok || len(values) < 3 {
		return RGBColor{}, false
	}
	r, err1 := strconv.Atoi(values[0])
	g, err2 := strconv.Atoi(values[1])
	b, err3 := strconv.Atoi(values[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return RGBColor{}, false
	}
	alpha := 1.0
	if len(values) >= 4 {
		parsed, err := strconv.ParseFloat(values[3], 64)
		if err != nil {
			return RGBColor{}, false
		}
		alpha = parsed
	}
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return RGBColor{}, false
	}
	return RGBColor{R: r, G: g, B: b, A: alpha}, true
}

// parseHSLColor 解析 HSL 颜色。
func parseHSLColor(input string) (RGBColor, bool) {
	values, ok := parseColorFunction(input, "hsl")
	if !ok || len(values) < 3 {
		return RGBColor{}, false
	}
	h, err1 := strconv.ParseFloat(values[0], 64)
	s, err2 := strconv.ParseFloat(values[1], 64)
	l, err3 := strconv.ParseFloat(values[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return RGBColor{}, false
	}
	alpha := 1.0
	if len(values) >= 4 {
		parsed, err := strconv.ParseFloat(values[3], 64)
		if err != nil {
			return RGBColor{}, false
		}
		alpha = parsed
	}
	return hslToRGB(h, s, l, alpha), true
}

// parseOKLCHColor 解析 OKLCH 颜色。
func parseOKLCHColor(input string) (RGBColor, bool) {
	values, ok := parseColorFunction(input, "oklch")
	if !ok || len(values) < 3 {
		return RGBColor{}, false
	}
	l, err1 := strconv.ParseFloat(values[0], 64)
	c, err2 := strconv.ParseFloat(values[1], 64)
	h, err3 := strconv.ParseFloat(values[2], 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return RGBColor{}, false
	}
	alpha := 1.0
	if len(values) >= 4 {
		parsed, err := strconv.ParseFloat(values[3], 64)
		if err != nil {
			return RGBColor{}, false
		}
		alpha = parsed
	}
	return oklchToRGB(l, c, h, alpha), true
}

// parseColorFunction 解析 `name(a, b, c)` 形式的颜色函数。
func parseColorFunction(input, name string) ([]string, bool) {
	lowered := strings.ToLower(strings.TrimSpace(input))
	prefix := name + "("
	if !strings.HasPrefix(lowered, prefix) {
		return nil, false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(lowered, prefix), ")")
	parts := strings.Split(body, ",")
	out := []string{}
	for _, part := range parts {
		out = append(out, strings.TrimSpace(part))
	}
	return out, true
}

// hslToRGB 把 HSL 转成 RGB。
func hslToRGB(h, s, l, alpha float64) RGBColor {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	s = clampFloat(s, 0, 1)
	l = clampFloat(l, 0, 1)

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return RGBColor{
		R: int(math.Round((r + m) * 255)),
		G: int(math.Round((g + m) * 255)),
		B: int(math.Round((b + m) * 255)),
		A: alpha,
	}
}

// oklchToRGB 把 OKLCH 转成 sRGB。
//
// 采用 OKLab 的标准矩阵变换；超出 sRGB 色域时按通道截断处理。
func oklchToRGB(l, c, h, alpha float64) RGBColor {
	hRad := h * math.Pi / 180
	a := c * math.Cos(hRad)
	b := c * math.Sin(hRad)

	l_ := l + 0.3963377774*a + 0.2158037573*b
	m_ := l - 0.1055613458*a - 0.0638541728*b
	s_ := l - 0.0894841775*a - 1.2914855480*b

	l3 := l_ * l_ * l_
	m3 := m_ * m_ * m_
	s3 := s_ * s_ * s_

	r := +4.0767416621*l3 - 3.3077115913*m3 + 0.2309699292*s3
	g := -1.2684380046*l3 + 2.6097574011*m3 - 0.3413193965*s3
	bl := -0.0041960863*l3 - 0.7034186147*m3 + 1.7076147010*s3

	return RGBColor{
		R: clampInt(int(math.Round(r*255)), 0, 255),
		G: clampInt(int(math.Round(g*255)), 0, 255),
		B: clampInt(int(math.Round(bl*255)), 0, 255),
		A: alpha,
	}
}

// formatHex 渲染十六进制表示。
func formatHex(color RGBColor) string {
	return fmt.Sprintf("#%02X%02X%02X", color.R, color.G, color.B)
}

// formatRGB 渲染 RGB 表示。
func formatRGB(color RGBColor) string {
	if color.A < 1 {
		return fmt.Sprintf("rgba(%d, %d, %d, %g)", color.R, color.G, color.B, color.A)
	}
	return fmt.Sprintf("rgb(%d, %d, %d)", color.R, color.G, color.B)
}

// formatHSL 渲染 HSL 表示。
func formatHSL(color RGBColor) string {
	h, s, l := rgbToHSL(color)
	if color.A < 1 {
		return fmt.Sprintf("hsla(%g, %g%%, %g%%, %g)", h, s*100, l*100, color.A)
	}
	return fmt.Sprintf("hsl(%g, %g%%, %g%%)", h, s*100, l*100)
}

// formatOKLCH 渲染 OKLCH 表示。
func formatOKLCH(color RGBColor) string {
	l, c, h := rgbToOKLCH(color)
	if color.A < 1 {
		return fmt.Sprintf("oklch(%g %g %g / %g)", l, c, h, color.A)
	}
	return fmt.Sprintf("oklch(%g %g %g)", l, c, h)
}

// rgbToHSL 把 RGB 转成 HSL。
func rgbToHSL(color RGBColor) (float64, float64, float64) {
	r := float64(color.R) / 255
	g := float64(color.G) / 255
	b := float64(color.B) / 255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	l := (max + min) / 2
	if delta == 0 {
		return 0, 0, l
	}

	var s float64
	if l > 0.5 {
		s = delta / (2 - max - min)
	} else {
		s = delta / (max + min)
	}

	var h float64
	switch max {
	case r:
		h = (g - b) / delta
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/delta + 2
	default:
		h = (r-g)/delta + 4
	}
	h *= 60

	return h, s, l
}

// rgbToOKLCH 把 RGB 转成 OKLCH。
func rgbToOKLCH(color RGBColor) (float64, float64, float64) {
	r := float64(color.R) / 255
	g := float64(color.G) / 255
	b := float64(color.B) / 255

	l := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*b)
	m := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*b)
	s := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*b)

	l_ := 0.2104542553*l + 0.7936177850*m - 0.0040720468*s
	m_ := 1.9779984951*l - 2.4285922050*m + 0.4505937099*s
	s_ := 0.0259040371*l + 0.7827717662*m - 0.8086757660*s

	labL := 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_
	labA := 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_
	labB := 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_

	c := math.Sqrt(labA*labA + labB*labB)
	h := math.Atan2(labB, labA) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return labL, c, h
}

// relativeLuminance 计算相对亮度。
//
// 依据 WCAG 2.1 的 sRGB 相对亮度定义。
func relativeLuminance(color RGBColor) float64 {
	return 0.2126*linearize(color.R) + 0.7152*linearize(color.G) + 0.0722*linearize(color.B)
}

// linearize 把 sRGB 分量线性化。
func linearize(value int) float64 {
	channel := float64(value) / 255
	if channel <= 0.03928 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}

// contrastRatio 计算两个颜色的对比度。
func contrastRatio(a, b RGBColor) float64 {
	l1 := relativeLuminance(a)
	l2 := relativeLuminance(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return math.Round(((l1+0.05)/(l2+0.05))*100) / 100
}

// clampFloat 把数值限制在区间内。
func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// clampInt 把整数限制在区间内。
func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Validate 做字段级校验。
func (r *ColorConvertRequest) Validate() error {
	if strings.TrimSpace(r.Input) == "" {
		return tools.MissingField("input")
	}
	if len(r.Input) > 128 {
		return tools.InvalidField("input", []int{0, 128}, len(r.Input))
	}
	if r.InputFormat != "" && !containsString(ColorFormat(), r.InputFormat) {
		return tools.InvalidField("inputFormat", ColorFormat(), r.InputFormat)
	}
	for _, format := range r.OutputFormats {
		if !containsString(ColorFormat(), format) {
			return tools.InvalidField("outputFormats", ColorFormat(), format)
		}
	}
	return nil
}

// containsString 判断切片是否包含某值。
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
