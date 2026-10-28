package text

import (
	"strings"
	"unicode"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 命名风格。
const (
	// StyleUpper 大写：HTTPSERVER。
	StyleUpper = "upper"
	// StyleLower 小写：httpserver。
	StyleLower = "lower"
	// StyleCamel 驼峰：httpserver → httpServer。
	StyleCamel = "camel"
	// StylePascal 帕斯卡：HttpServer。
	StylePascal = "pascal"
	// StyleSnake 下划线：http_server。
	StyleSnake = "snake"
	// StyleKebab 短横线：http-server。
	StyleKebab = "kebab"
	// StyleConstant 常量：HTTP_SERVER。
	StyleConstant = "constant"
	// StyleTitle 标题：Http Server。
	StyleTitle = "title"
	// StyleDot 点分：http.server。
	StyleDot = "dot"
	// StylePath 路径：http/server。
	StylePath = "path"
)

// TextStyles 返回受支持的命名风格（顺序稳定）。
func TextStyles() []string {
	return []string{
		StyleUpper, StyleLower, StyleCamel, StylePascal,
		StyleSnake, StyleKebab, StyleConstant, StyleTitle,
		StyleDot, StylePath,
	}
}

// TextConvertRequest 是 text-convert 工具的请求。
type TextConvertRequest struct {
	// Input 是待转换的文本。
	Input TextInput `json:"input"`
	// Style 是目标命名风格。
	Style string `json:"style"`
	// Separator 是自定义分隔符（可选）；给出时覆盖风格的默认分隔符。
	Separator string `json:"separator"`
}

// TextConvertResult 是 text-convert 工具的响应。
type TextConvertResult struct {
	// Output 是转换结果。
	Output string `json:"output"`
	// Style 是实际使用的风格。
	Style string `json:"style"`
	// Words 是切分出的词元。
	Words []string `json:"words"`
}

// TextConvert 转换文本的命名风格。
//
// 切分规则：非字母数字字符视为分隔符。
// 因此中文段落会被逐字切开——这是规则本身的限制，
// 不做「按语言分词」的处理。
func TextConvert(req *TextConvertRequest) (TextConvertResult, error) {
	if err := req.Validate(); err != nil {
		return TextConvertResult{}, err
	}

	source, err := req.Input.Text()
	if err != nil {
		return TextConvertResult{}, err
	}

	words := splitWords(source)
	out := joinWords(words, req.Style, req.Separator)

	return TextConvertResult{
		Output: out,
		Style:  req.Style,
		Words:  words,
	}, nil
}

// Validate 做字段级校验。
func (r *TextConvertRequest) Validate() error {
	switch r.Style {
	case "", StyleUpper, StyleLower, StyleCamel, StylePascal,
		StyleSnake, StyleKebab, StyleConstant, StyleTitle, StyleDot, StylePath:
	default:
		return tools.InvalidField("style", TextStyles(), r.Style)
	}
	return r.Input.Validate("input")
}

// splitWords 把文本切分成词元。
//
// 非字母数字字符一律视为分隔符；连续分隔符不产生空词元。
func splitWords(source string) []string {
	words := []string{}
	var current strings.Builder

	for _, r := range source {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
			continue
		}
		if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// joinWords 按目标风格拼接词元。
func joinWords(words []string, style string, separator string) string {
	if len(words) == 0 {
		return ""
	}

	switch style {
	case StyleUpper:
		return strings.ToUpper(strings.Join(words, ""))
	case StyleLower:
		return strings.ToLower(strings.Join(words, ""))
	case StyleCamel:
		return strings.ToLower(words[0]) + pascalWords(words[1:])
	case StylePascal:
		return pascalWords(words)
	case StyleSnake:
		return strings.ToLower(strings.Join(words, "_"))
	case StyleKebab:
		return strings.ToLower(strings.Join(words, "-"))
	case StyleConstant:
		return strings.ToUpper(strings.Join(words, "_"))
	case StyleTitle:
		return strings.ToLower(strings.Join(words, " "))
	case StyleDot:
		return strings.ToLower(strings.Join(words, "."))
	case StylePath:
		return strings.ToLower(strings.Join(words, "/"))
	case "":
		return strings.Join(words, separator)
	default:
		return strings.Join(words, separator)
	}
}

// pascalWords 把词元拼成帕斯卡风格。
func pascalWords(words []string) string {
	var b strings.Builder
	for _, w := range words {
		b.WriteString(capitalize(w))
	}
	return b.String()
}

// capitalize 把词元的首字母转成大写。
func capitalize(word string) string {
	if word == "" {
		return word
	}
	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
