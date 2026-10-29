package data

import (
	"context"
	"strconv"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 校验类型。
const (
	// IDTypeIDCard 是 18 位居民身份证号。
	IDTypeIDCard = "idCard"
	// IDTypeUSCC 是统一社会信用代码。
	IDTypeUSCC = "uscc"
	// IDTypePhone 是中国大陆手机号。
	IDTypePhone = "phone"
	// IDTypeBankCard 是银行卡号。
	IDTypeBankCard = "bankCard"
	// IDTypePostcode 是邮政编码。
	IDTypePostcode = "postcode"
)

// IDType 返回受支持的校验类型（顺序稳定）。
func IDType() []string {
	return []string{
		IDTypeIDCard, IDTypeUSCC, IDTypePhone,
		IDTypeBankCard, IDTypePostcode,
	}
}

// IDValidateRequest 是 id-validator 工具的请求。
type IDValidateRequest struct {
	// Type 是校验类型。
	Type string `json:"type"`
	// Value 是待校验的值。
	Value string `json:"value"`
}

// IDValidateResult 是 id-validator 工具的响应。
type IDValidateResult struct {
	// Type 是校验类型。
	Type string `json:"type"`
	// Value 是规范化后的值。
	Value string `json:"value"`
	// Valid 是否校验通过。
	Valid bool `json:"valid"`
	// Normalized 是规范化结果；校验不通过时为空串。
	Normalized string `json:"normalized"`
	// CheckDigit 是校验位；该类型无校验位时为空串。
	CheckDigit string `json:"checkDigit"`
	// Region 是归属地区；该类型无地区信息时为空串。
	Region string `json:"region"`
	// Birthday 是出生日期；该类型无出生日期时为空串。
	Birthday string `json:"birthday"`
	// Gender 是性别；该类型无性别时为空串。
	Gender string `json:"gender"`
	// Issue 是校验失败的原因；通过时为空串。
	Issue string `json:"issue"`
}

// maxIDValueLength 是待校验值的长度上限。
const maxIDValueLength = 64

// IDValidate 校验身份证号、统一社会信用代码、手机号、银行卡号与邮政编码。
//
// 校验位算法是**公开的定长加权校验和**，不属于密码学运算，
// 因此直接实现，不经 cryptosvc。
func IDValidate(_ context.Context, req *IDValidateRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	value := strings.ToUpper(strings.TrimSpace(req.Value))
	result := IDValidateResult{Type: req.Type, Value: value}

	switch req.Type {
	case IDTypeIDCard:
		fillIDCardResult(&result, value)
	case IDTypeUSCC:
		fillUSCCResult(&result, value)
	case IDTypePhone:
		fillPhoneResult(&result, value)
	case IDTypeBankCard:
		fillBankCardResult(&result, value)
	case IDTypePostcode:
		fillPostcodeResult(&result, value)
	}

	return tools.Result{Data: result}, nil
}

// idCardWeights 是 18 位身份证的加权因子。
var idCardWeights = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}

// idCardCheckChars 是 18 位身份证的校验码字符表。
const idCardCheckChars = "10X98765432"

// 身份证行政区划代码的前两位对应的省级名称。
//
// 只覆盖省级；地市一级不展开，避免数据集过大且准确率下降。
var idCardProvinces = map[string]string{
	"11": "北京", "12": "天津", "13": "河北", "14": "山西", "15": "内蒙古",
	"21": "辽宁", "22": "吉林", "23": "黑龙江",
	"31": "上海", "32": "江苏", "33": "浙江", "34": "安徽", "35": "福建",
	"36": "江西", "37": "山东",
	"41": "河南", "42": "湖北", "43": "湖南", "44": "广东", "45": "广西",
	"46": "海南",
	"50": "重庆", "51": "四川", "52": "贵州", "53": "云南", "54": "西藏",
	"61": "陕西", "62": "甘肃", "63": "青海", "64": "宁夏", "65": "新疆",
	"71": "台湾", "81": "香港", "82": "澳门", "83": "台湾",
}

// fillIDCardResult 填充身份证校验结果。
func fillIDCardResult(result *IDValidateResult, value string) {
	if len(value) != 18 {
		result.Issue = "长度必须是 18 位"
		return
	}
	for i := 0; i < 17; i++ {
		if value[i] < '0' || value[i] > '9' {
			result.Issue = "前 17 位必须是数字"
			return
		}
	}
	if value[17] != 'X' && (value[17] < '0' || value[17] > '9') {
		result.Issue = "最后一位必须是数字或 X"
		return
	}

	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(value[i]-'0') * idCardWeights[i]
	}
	expected := idCardCheckChars[sum%11]
	if value[17] != expected {
		result.Issue = "校验位不匹配"
		result.CheckDigit = string(expected)
		return
	}

	birthday := value[6:14]
	if !isValidIDBirthday(birthday) {
		result.Issue = "出生日期不合法"
		return
	}

	gender := "女"
	if int(value[16]-'0')%2 == 1 {
		gender = "男"
	}

	result.Valid = true
	result.Normalized = value
	result.CheckDigit = string(expected)
	result.Birthday = formatIDBirthday(birthday)
	result.Gender = gender
	if province, ok := idCardProvinces[value[:2]]; ok {
		result.Region = province
	}
}

// isValidIDBirthday 判断出生日期是否合法。
func isValidIDBirthday(birthday string) bool {
	year, err1 := strconv.Atoi(birthday[:4])
	month, err2 := strconv.Atoi(birthday[4:6])
	day, err3 := strconv.Atoi(birthday[6:8])
	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}
	if year < 1900 || year > 2100 {
		return false
	}
	if month < 1 || month > 12 {
		return false
	}
	if day < 1 || day > 31 {
		return false
	}
	days := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	limit := days[month-1]
	if month == 2 && !isLeapYear(year) {
		limit = 28
	}
	return day <= limit
}

// isLeapYear 判断是否闰年。
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// formatIDBirthday 把 YYYYMMDD 格式化成 YYYY-MM-DD。
func formatIDBirthday(birthday string) string {
	return birthday[:4] + "-" + birthday[4:6] + "-" + birthday[6:8]
}

// usccCharSet 是统一社会信用代码的字符集。
const usccCharSet = "0123456789ABCDEFGHJKLMNPQRTUWXY"

// usccWeights 是统一社会信用代码的加权因子。
var usccWeights = []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}

// fillUSCCResult 填充统一社会信用代码校验结果。
func fillUSCCResult(result *IDValidateResult, value string) {
	if len(value) != 18 {
		result.Issue = "长度必须是 18 位"
		return
	}

	index := map[rune]int{}
	for i, c := range usccCharSet {
		index[c] = i
	}

	sum := 0
	for i := 0; i < 17; i++ {
		pos, ok := index[rune(value[i])]
		if !ok {
			result.Issue = "包含不在字符集内的字符"
			return
		}
		sum += pos * usccWeights[i]
	}

	expected := rune(usccCharSet[31-sum%31])
	if rune(value[17]) != expected {
		result.Issue = "校验位不匹配"
		result.CheckDigit = string(expected)
		return
	}

	result.Valid = true
	result.Normalized = value
	result.CheckDigit = string(expected)
}

// fillPhoneResult 填充手机号校验结果。
func fillPhoneResult(result *IDValidateResult, value string) {
	// 允许带 +86 / 86 前缀或空格分隔。
	digits := ""
	for _, c := range value {
		if c >= '0' && c <= '9' {
			digits += string(c)
		} else if c != '+' && c != ' ' && c != '-' {
			result.Issue = "包含非法字符"
			return
		}
	}
	if len(digits) == 13 && strings.HasPrefix(digits, "86") {
		digits = digits[2:]
	}
	if len(digits) != 11 {
		result.Issue = "长度必须是 11 位"
		return
	}
	if digits[0] != '1' || digits[1] < '3' || digits[1] > '9' {
		result.Issue = "号段不合法"
		return
	}
	result.Valid = true
	result.Normalized = digits
}

// bankCardWeights 是银行卡 Luhn 校验的加权因子。
var bankCardWeights = []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}

// fillBankCardResult 填充银行卡校验结果。
func fillBankCardResult(result *IDValidateResult, value string) {
	digits := ""
	for _, c := range value {
		if c < '0' || c > '9' {
			result.Issue = "只能包含数字"
			return
		}
		digits += string(c)
	}
	if len(digits) < 12 || len(digits) > 19 {
		result.Issue = "长度必须在 12 到 19 位之间"
		return
	}

	// Luhn 校验：从右往左，偶数位（从右数第 2 位起）乘 2 后折位求和。
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		digit := int(digits[i] - '0')
		if double {
			digit *= bankCardWeights[i%len(bankCardWeights)]
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}

	if sum%10 != 0 {
		result.Issue = "Luhn 校验不通过"
		return
	}

	// 补齐到 19 位计算缺失的校验位。
	result.Valid = true
	result.Normalized = digits
}

// fillPostcodeResult 填充邮政编码校验结果。
func fillPostcodeResult(result *IDValidateResult, value string) {
	if len(value) != 6 {
		result.Issue = "长度必须是 6 位"
		return
	}
	for i := 0; i < 6; i++ {
		if value[i] < '0' || value[i] > '9' {
			result.Issue = "只能包含数字"
			return
		}
	}
	result.Valid = true
	result.Normalized = value
}

// Validate 做字段级校验。
func (r *IDValidateRequest) Validate() error {
	if r.Type == "" {
		return tools.MissingField("type")
	}
	if !containsString(IDType(), r.Type) {
		return tools.InvalidField("type", IDType(), r.Type)
	}
	if strings.TrimSpace(r.Value) == "" {
		return tools.MissingField("value")
	}
	if len(r.Value) > maxIDValueLength {
		return tools.InvalidField("value", []int{0, maxIDValueLength}, len(r.Value))
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
