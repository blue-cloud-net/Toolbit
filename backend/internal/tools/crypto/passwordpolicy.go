package crypto

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 口令策略检查项。
const (
	// PolicyCheckLength 是长度要求。
	PolicyCheckLength = "length"
	// PolicyCheckUpper 是大写字母要求。
	PolicyCheckUpper = "upper"
	// PolicyCheckLower 是小写字母要求。
	PolicyCheckLower = "lower"
	// PolicyCheckDigit 是数字要求。
	PolicyCheckDigit = "digit"
	// PolicyCheckSymbol 是符号要求。
	PolicyCheckSymbol = "symbol"
	// PolicyCheckRepeat 是重复字符要求。
	PolicyCheckRepeat = "repeat"
	// PolicyCheckSequential 是顺序字符要求。
	PolicyCheckSequential = "sequential"
	// PolicyCheckCommon 是常见弱口令要求。
	PolicyCheckCommon = "common"
)

// PasswordPolicyCheck 返回受支持的检查项（顺序稳定）。
func PasswordPolicyCheck() []string {
	return []string{
		PolicyCheckLength, PolicyCheckUpper, PolicyCheckLower,
		PolicyCheckDigit, PolicyCheckSymbol, PolicyCheckRepeat,
		PolicyCheckSequential, PolicyCheckCommon,
	}
}

// 策略预设。
const (
	// PolicyPresetBasic 是基础策略。
	PolicyPresetBasic = "basic"
	// PolicyPresetMedium 是中等策略。
	PolicyPresetMedium = "medium"
	// PolicyPresetStrict 是严格策略。
	PolicyPresetStrict = "strict"
)

// PasswordPolicyPreset 返回受支持的策略预设（顺序稳定）。
func PasswordPolicyPreset() []string {
	return []string{PolicyPresetBasic, PolicyPresetMedium, PolicyPresetStrict}
}

// PasswordPolicyRequest 是 password-policy 工具的请求。
type PasswordPolicyRequest struct {
	// Password 是待检查的口令。
	Password string `json:"password"`
	// Preset 是策略预设；默认 medium。
	Preset string `json:"preset"`
	// MinLength 是最小长度；留空用预设值。
	MinLength int `json:"minLength"`
	// MaxLength 是最大长度；0 表示不限。
	MaxLength int `json:"maxLength"`
	// RequireUpper 是否要求大写字母。
	RequireUpper bool `json:"requireUpper"`
	// RequireLower 是否要求小写字母。
	RequireLower bool `json:"requireLower"`
	// RequireDigit 是否要求数字。
	RequireDigit bool `json:"requireDigit"`
	// RequireSymbol 是否要求符号。
	RequireSymbol bool `json:"requireSymbol"`
	// MaxRepeatRun 是允许的最大连续重复次数；0 表示不限。
	MaxRepeatRun int `json:"maxRepeatRun"`
}

// PolicyCheckResult 是一项检查的结果。
type PolicyCheckResult struct {
	// Check 是检查项标识。
	Check string `json:"check"`
	// Passed 是否通过。
	Passed bool `json:"passed"`
	// Message 是检查说明。
	Message string `json:"message"`
}

// PasswordPolicyResult 是 password-policy 工具的响应。
type PasswordPolicyResult struct {
	// Preset 是实际使用的策略预设。
	Preset string `json:"preset"`
	// Compliant 是否满足全部启用的检查项。
	Compliant bool `json:"compliant"`
	// Score 是通过率，取值 0~100。
	Score int `json:"score"`
	// Strength 是强度等级：weak / fair / good / strong。
	Strength string `json:"strength"`
	// Checks 是各项检查结果。
	Checks []PolicyCheckResult `json:"checks"`
	// Suggestions 是改进建议。
	Suggestions []string `json:"suggestions"`
}

// 常见弱口令。
var commonPasswords = []string{
	"123456", "password", "12345678", "qwerty", "123456789",
	"111111", "1234567", "abc123", "password1", "admin",
	"letmein", "welcome", "monkey", "dragon", "iloveyou",
}

// 预设的默认要求。
var policyPresets = map[string]struct {
	minLength int
	upper     bool
	lower     bool
	digit     bool
	symbol    bool
}{
	PolicyPresetBasic:  {minLength: 8},
	PolicyPresetMedium: {minLength: 12, upper: true, lower: true, digit: true},
	PolicyPresetStrict: {minLength: 16, upper: true, lower: true, digit: true, symbol: true},
}

// PasswordPolicy 检查口令是否符合给定的策略要求。
//
// 这是**纯规则检查**：不涉及密码学运算，也不存储口令。
func PasswordPolicy(_ context.Context, req *PasswordPolicyRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	preset := req.Preset
	if preset == "" {
		preset = PolicyPresetMedium
	}
	defaults := policyPresets[preset]

	minLength := req.MinLength
	if minLength == 0 {
		minLength = defaults.minLength
	}
	requireUpper := req.RequireUpper || defaults.upper
	requireLower := req.RequireLower || defaults.lower
	requireDigit := req.RequireDigit || defaults.digit
	requireSymbol := req.RequireSymbol || defaults.symbol

	password := req.Password
	runes := []rune(password)

	checks := []PolicyCheckResult{}

	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckLength,
		Passed:  len(runes) >= minLength && (req.MaxLength == 0 || len(runes) <= req.MaxLength),
		Message: lengthMessage(len(runes), minLength, req.MaxLength),
	})

	upperCount := countByClass(password, isUpper)
	lowerCount := countByClass(password, isLower)
	digitCount := countByClass(password, isDigit)
	symbolCount := countByClass(password, isSymbol)

	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckUpper,
		Passed:  !requireUpper || upperCount > 0,
		Message: classMessage("大写字母", upperCount, requireUpper),
	})
	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckLower,
		Passed:  !requireLower || lowerCount > 0,
		Message: classMessage("小写字母", lowerCount, requireLower),
	})
	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckDigit,
		Passed:  !requireDigit || digitCount > 0,
		Message: classMessage("数字", digitCount, requireDigit),
	})
	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckSymbol,
		Passed:  !requireSymbol || symbolCount > 0,
		Message: classMessage("符号", symbolCount, requireSymbol),
	})

	repeatRun := longestRepeatRun(password)
	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckRepeat,
		Passed:  req.MaxRepeatRun <= 0 || repeatRun <= req.MaxRepeatRun,
		Message: repeatMessage(repeatRun, req.MaxRepeatRun),
	})

	sequential := longestSequentialRun(password)
	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckSequential,
		Passed:  sequential <= 3,
		Message: sequentialMessage(sequential),
	})

	isCommon := isCommonPassword(password)
	checks = append(checks, PolicyCheckResult{
		Check:   PolicyCheckCommon,
		Passed:  !isCommon,
		Message: commonMessage(isCommon),
	})

	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		}
	}
	score := passed * 100 / len(checks)

	compliant := true
	for _, check := range checks {
		if !check.Passed {
			compliant = false
			break
		}
	}

	suggestions := []string{}
	for _, check := range checks {
		if !check.Passed {
			suggestions = append(suggestions, check.Message)
		}
	}

	return tools.Result{Data: PasswordPolicyResult{
		Preset:      preset,
		Compliant:   compliant,
		Score:       score,
		Strength:    strengthLevel(score, len(runes)),
		Checks:      checks,
		Suggestions: suggestions,
	}}, nil
}

// lengthMessage 生成长度检查说明。
func lengthMessage(length, minLength, maxLength int) string {
	switch {
	case length < minLength:
		return "长度不足：当前 " + itoa(length) + " 位，至少需要 " + itoa(minLength) + " 位"
	case maxLength > 0 && length > maxLength:
		return "长度超出上限：当前 " + itoa(length) + " 位，上限为 " + itoa(maxLength) + " 位"
	default:
		return "长度符合要求：当前 " + itoa(length) + " 位"
	}
}

// classMessage 生成字符类检查说明。
func classMessage(name string, count int, required bool) string {
	if !required {
		return "未要求包含" + name
	}
	if count == 0 {
		return "缺少" + name
	}
	return "包含 " + itoa(count) + " 个" + name
}

// repeatMessage 生成重复字符检查说明。
func repeatMessage(run, limit int) string {
	if limit <= 0 {
		return "未限制连续重复字符（当前最长 " + itoa(run) + " 个）"
	}
	if run > limit {
		return "存在 " + itoa(run) + " 个连续重复字符，超过上限 " + itoa(limit)
	}
	return "连续重复字符未超过上限 " + itoa(limit)
}

// sequentialMessage 生成顺序字符检查说明。
func sequentialMessage(run int) string {
	if run > 3 {
		return "存在 " + itoa(run) + " 个连续递增或递减的字符"
	}
	return "没有明显的顺序字符"
}

// commonMessage 生成常见弱口令检查说明。
func commonMessage(isCommon bool) string {
	if isCommon {
		return "属于常见弱口令，容易被字典攻击"
	}
	return "不在常见弱口令列表中"
}

// strengthLevel 根据通过率与长度给出强度等级。
func strengthLevel(score, length int) string {
	switch {
	case score >= 90 && length >= 16:
		return "strong"
	case score >= 75:
		return "good"
	case score >= 50:
		return "fair"
	default:
		return "weak"
	}
}

// countByClass 统计满足条件的字符数。
func countByClass(value string, predicate func(rune) bool) int {
	count := 0
	for _, r := range value {
		if predicate(r) {
			count++
		}
	}
	return count
}

// isUpper 判断是否是大写字母。
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// isLower 判断是否是小写字母。
func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// isDigit 判断是否是数字。
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// isSymbol 判断是否是符号。
func isSymbol(r rune) bool {
	return !isUpper(r) && !isLower(r) && !isDigit(r) && r != ' ' && r >= 33
}

// longestRepeatRun 返回最长连续重复字符数。
func longestRepeatRun(value string) int {
	runes := []rune(value)
	longest := 0
	current := 0
	for i := 0; i < len(runes); i++ {
		if i > 0 && runes[i] == runes[i-1] {
			current++
		} else {
			current = 1
		}
		if current > longest {
			longest = current
		}
	}
	return longest
}

// longestSequentialRun 返回最长连续递增或递减的字符数。
func longestSequentialRun(value string) int {
	runes := []rune(value)
	longest := 0
	increasing := 0
	decreasing := 0
	for i := 0; i < len(runes); i++ {
		if i == 0 {
			increasing = 1
			decreasing = 1
			continue
		}
		diff := int(runes[i]) - int(runes[i-1])
		switch diff {
		case 1:
			increasing++
			decreasing = 1
		case -1:
			decreasing++
			increasing = 1
		default:
			increasing = 1
			decreasing = 1
		}
		if increasing > longest {
			longest = increasing
		}
		if decreasing > longest {
			longest = decreasing
		}
	}
	return longest
}

// isCommonPassword 判断是否属于常见弱口令。
func isCommonPassword(value string) bool {
	lowered := strings.ToLower(value)
	for _, password := range commonPasswords {
		if lowered == password {
			return true
		}
	}
	return false
}

// itoa 把整数转成字符串。
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// Validate 做字段级校验。
func (r *PasswordPolicyRequest) Validate() error {
	if r.Password == "" {
		return tools.MissingField("password")
	}
	if len(r.Password) > 1024 {
		return tools.InvalidField("password", []int{0, 1024}, len(r.Password))
	}
	preset := r.Preset
	if preset == "" {
		preset = PolicyPresetMedium
	}
	if !containsString(PasswordPolicyPreset(), preset) {
		return tools.InvalidField("preset", PasswordPolicyPreset(), preset)
	}
	if r.MinLength < 0 || r.MinLength > 256 {
		return tools.InvalidField("minLength", []int{0, 256}, r.MinLength)
	}
	if r.MaxLength < 0 || r.MaxLength > 1024 {
		return tools.InvalidField("maxLength", []int{0, 1024}, r.MaxLength)
	}
	if r.MaxLength > 0 && r.MinLength > r.MaxLength {
		return tools.InvalidField("minLength", []int{0, r.MaxLength}, r.MinLength)
	}
	if r.MaxRepeatRun < 0 || r.MaxRepeatRun > 32 {
		return tools.InvalidField("maxRepeatRun", []int{0, 32}, r.MaxRepeatRun)
	}
	return nil
}
