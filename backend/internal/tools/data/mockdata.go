package data

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 假数据字段类型。
const (
	// MockTypeString 是随机字符串。
	MockTypeString = "string"
	// MockTypeInt 是随机整数。
	MockTypeInt = "int"
	// MockTypeFloat 是随机小数。
	MockTypeFloat = "float"
	// MockTypeBool 是随机布尔值。
	MockTypeBool = "bool"
	// MockTypeEmail 是随机邮箱。
	MockTypeEmail = "email"
	// MockTypePhone 是随机手机号。
	MockTypePhone = "phone"
	// MockTypeName 是随机姓名。
	MockTypeName = "name"
	// MockTypeURL 是随机 URL。
	MockTypeURL = "url"
	// MockTypeUUID 是随机 UUID v4。
	MockTypeUUID = "uuid"
	// MockTypeDate 是随机日期。
	MockTypeDate = "date"
	// MockTypeDateTime 是随机日期时间。
	MockTypeDateTime = "datetime"
	// MockTypeIP 是随机 IP。
	MockTypeIP = "ip"
	// MockTypeColor 是随机颜色。
	MockTypeColor = "color"
)

// MockType 返回受支持的字段类型（顺序稳定）。
func MockType() []string {
	return []string{
		MockTypeString, MockTypeInt, MockTypeFloat, MockTypeBool,
		MockTypeEmail, MockTypePhone, MockTypeName, MockTypeURL,
		MockTypeUUID, MockTypeDate, MockTypeDateTime, MockTypeIP, MockTypeColor,
	}
}

// 支持的语言。
const (
	// MockLocaleZHCN 是简体中文。
	MockLocaleZHCN = "zh-CN"
	// MockLocaleENUS 是英文。
	MockLocaleENUS = "en-US"
)

// MockLocale 返回受支持的语言（顺序稳定）。
func MockLocale() []string {
	return []string{MockLocaleZHCN, MockLocaleENUS}
}

// MockDataRequest 是 mock-data 工具的请求。
type MockDataRequest struct {
	// Types 是要生成的字段类型列表；为空表示全部类型。
	Types []string `json:"types"`
	// Count 是每种类型生成的条数；默认 3。
	Count int `json:"count"`
	// Locale 是姓名等文本的语言；默认 zh-CN。
	Locale string `json:"locale"`
	// Prefix 是给生成值加的前缀。
	Prefix string `json:"prefix"`
}

// MockField 是一个生成的字段值。
type MockField struct {
	// Type 是字段类型。
	Type string `json:"type"`
	// Value 是生成的值；一律字符串化。
	Value string `json:"value"`
}

// MockDataResult 是 mock-data 工具的响应。
type MockDataResult struct {
	// Types 是实际生成的类型列表。
	Types []string `json:"types"`
	// Fields 是生成的字段值列表。
	Fields []MockField `json:"fields"`
	// Count 是每种类型生成的条数。
	Count int `json:"count"`
	// Locale 是使用的语言。
	Locale string `json:"locale"`
	// Seed 是本次生成使用的随机种子，便于复现。
	Seed string `json:"seed"`
}

// maxMockCount 是每种类型生成条数的上限。
const maxMockCount = 100

// 中文姓氏与名字用字。
// mockPhonePrefixes 是中国大陆手机号的第二位前缀。
var mockPhonePrefixes = []string{"3", "4", "5", "6", "7", "8", "9"}

var mockSurnamesZH = []string{"张", "王", "李", "赵", "陈", "刘", "杨", "黄", "周", "吴"}
var mockGivenZH = []string{"三", "明", "伟", "芳", "娜", "敏", "静", "磊", "洋", "强"}

// 英文名用字。
var mockFirstNamesEN = []string{"Alice", "Bob", "Carol", "David", "Emma", "Frank", "Grace", "Henry"}
var mockLastNamesEN = []string{"Smith", "Johnson", "Brown", "Miller", "Wilson", "Moore", "Taylor"}

// MockData 按字段类型生成假数据。
//
// 随机源用 `math/rand/v2` 的全局函数：Go 1.25 起全局函数自动随机播种，
// 无需手写种子，也不引入额外依赖。
func MockData(_ context.Context, req *MockDataRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	types := req.Types
	if len(types) == 0 {
		types = MockType()
	}
	count := req.Count
	if count == 0 {
		count = 3
	}
	locale := req.Locale
	if locale == "" {
		locale = MockLocaleZHCN
	}

	seed := rand.Uint64()
	source := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))

	fields := []MockField{}
	for _, kind := range types {
		for i := 0; i < count; i++ {
			fields = append(fields, MockField{
				Type:  kind,
				Value: req.Prefix + generateMockValue(kind, locale, source),
			})
		}
	}

	return tools.Result{Data: MockDataResult{
		Types:  types,
		Fields: fields,
		Count:  count,
		Locale: locale,
		Seed:   fmt.Sprintf("%016x", seed),
	}}, nil
}

// generateMockValue 按类型生成一个值。
func generateMockValue(kind, locale string, source *rand.Rand) string {
	switch kind {
	case MockTypeString:
		return randomAlpha(source, 6+source.IntN(7))
	case MockTypeInt:
		return fmt.Sprintf("%d", 1+source.IntN(10000))
	case MockTypeFloat:
		return fmt.Sprintf("%.2f", float64(source.IntN(1000000))/100)
	case MockTypeBool:
		if source.IntN(2) == 0 {
			return "true"
		}
		return "false"
	case MockTypeEmail:
		return randomAlpha(source, 8) + "@example.com"
	case MockTypePhone:
		return "1" + mockPhonePrefixes[source.IntN(len(mockPhonePrefixes))] +
			fmt.Sprintf("%09d", source.IntN(1000000000))
	case MockTypeName:
		if locale == MockLocaleENUS {
			return mockFirstNamesEN[source.IntN(len(mockFirstNamesEN))] + " " +
				mockLastNamesEN[source.IntN(len(mockLastNamesEN))]
		}
		return mockSurnamesZH[source.IntN(len(mockSurnamesZH))] +
			mockGivenZH[source.IntN(len(mockGivenZH))]
	case MockTypeURL:
		return "https://example.com/" + randomAlpha(source, 8)
	case MockTypeUUID:
		return randomUUID(source)
	case MockTypeDate:
		return randomDate(source)
	case MockTypeDateTime:
		return randomDate(source) + "T" +
			fmt.Sprintf("%02d:%02d:%02d", source.IntN(24), source.IntN(60), source.IntN(60))
	case MockTypeIP:
		return fmt.Sprintf("192.168.%d.%d", source.IntN(256), source.IntN(256))
	case MockTypeColor:
		return fmt.Sprintf("#%02X%02X%02X", source.IntN(256), source.IntN(256), source.IntN(256))
	default:
		return ""
	}
}

// randomAlpha 生成随机小写字母串。
func randomAlpha(source *rand.Rand, length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	out := make([]byte, length)
	for i := range out {
		out[i] = letters[source.IntN(len(letters))]
	}
	return string(out)
}

// randomUUID 生成 UUID v4 形态的标识符。
func randomUUID(source *rand.Rand) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, 0, 36)
	for i := 0; i < 36; i++ {
		switch i {
		case 8, 13, 18, 23:
			out = append(out, '-')
		case 14:
			// 版本位固定为 4。
			out = append(out, '4')
		case 19:
			// 变体位取 8/9/a/b。
			out = append(out, "89ab"[source.IntN(4)])
		default:
			out = append(out, hexDigits[source.IntN(16)])
		}
	}
	return string(out)
}

// randomDate 生成 2020~2030 之间的随机日期。
func randomDate(source *rand.Rand) string {
	year := 2020 + source.IntN(11)
	month := 1 + source.IntN(12)
	day := 1 + source.IntN(28)
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// Validate 做字段级校验。
func (r *MockDataRequest) Validate() error {
	for _, kind := range r.Types {
		if !containsString(MockType(), kind) {
			return tools.InvalidField("types", MockType(), kind)
		}
	}
	if r.Count < 0 || r.Count > maxMockCount {
		return tools.InvalidField("count", []int{0, maxMockCount}, r.Count)
	}
	locale := r.Locale
	if locale == "" {
		locale = MockLocaleZHCN
	}
	if !containsString(MockLocale(), locale) {
		return tools.InvalidField("locale", MockLocale(), locale)
	}
	if len(r.Prefix) > 64 {
		return tools.InvalidField("prefix", []int{0, 64}, len(r.Prefix))
	}
	return nil
}
