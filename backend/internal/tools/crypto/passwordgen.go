package crypto

import (
	"context"
	"math"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 密码字符集。
const (
	// CharsetLower 是小写字母。
	CharsetLower = "lower"
	// CharsetUpper 是大写字母。
	CharsetUpper = "upper"
	// CharsetDigit 是数字。
	CharsetDigit = "digit"
	// CharsetSymbol 是符号。
	CharsetSymbol = "symbol"
)

const (
	// charsetLowerChars 是小写字母表。
	charsetLowerChars = "abcdefghijklmnopqrstuvwxyz"
	// charsetUpperChars 是大写字母表。
	charsetUpperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// charsetDigitChars 是数字表。
	charsetDigitChars = "0123456789"
	// charsetSymbolChars 是符号表。
	charsetSymbolChars = "!@#$%^&*()-_=+[]{}|;:,.<>?/~"
	// ambiguousChars 是易混字符。
	ambiguousChars = "0Oo1lI"
)

// 密码参数边界。
const (
	// MinPasswordLength 是密码最短长度。
	MinPasswordLength = 4
	// MaxPasswordLength 是密码最长长度。
	MaxPasswordLength = 128
	// MaxPasswordCount 是单次生成的最大数量。
	MaxPasswordCount = 100
)

// 口令策略模板。
const (
	// PasswordTemplateSimple 是简单口令模板。
	PasswordTemplateSimple = "simple"
	// PasswordTemplateStrong 是强口令模板。
	PasswordTemplateStrong = "strong"
	// PasswordTemplatePin 是数字口令模板。
	PasswordTemplatePin = "pin"
	// PasswordTemplatePassphrase 是可发音口令模板。
	PasswordTemplatePassphrase = "passphrase"
)

// PasswordTemplate 返回受支持的策略模板（顺序稳定）。
func PasswordTemplate() []string {
	return []string{
		PasswordTemplateSimple, PasswordTemplateStrong,
		PasswordTemplatePin, PasswordTemplatePassphrase,
	}
}

// PasswordGenRequest 是密码生成工具的请求。
type PasswordGenRequest struct {
	// Length 是密码长度；默认 16。
	Length int `json:"length"`
	// Count 是生成数量；默认 1。
	Count int `json:"count"`
	// Charsets 是字符集集合。
	Charsets []string `json:"charsets"`
	// ExcludeAmbiguous 是否排除易混字符。
	ExcludeAmbiguous bool `json:"excludeAmbiguous"`
	// RequireEachClass 是否要求每类字符至少出现一次。
	RequireEachClass bool `json:"requireEachClass"`
	// CustomCharset 是自定义字符集；非空时覆盖 Charsets。
	CustomCharset string `json:"customCharset"`
	// Template 是策略模板；留空表示不套模板。
	Template string `json:"template"`
	// Separator 是批量生成时的分隔符；留空用换行。
	Separator string `json:"separator"`
	// ExcludeChars 是要额外排除的字符。
	ExcludeChars string `json:"excludeChars"`
	// WordCount 是可发音口令的单词数；默认 3。
	WordCount int `json:"wordCount"`
}

// PasswordGenResult 是密码生成工具的响应。
type PasswordGenResult struct {
	// Passwords 是生成的密码列表。
	Passwords []string `json:"passwords"`
	// EntropyBits 是单个密码的估算熵（bit）。
	EntropyBits float64 `json:"entropyBits"`
	// Template 是实际使用的策略模板。
	Template string `json:"template"`
	// PoolSize 是字符池大小；可发音口令为空。
	PoolSize int `json:"poolSize"`
}

// Validate 做字段级校验。
func (r *PasswordGenRequest) Validate() error {
	if r.Length == 0 {
		r.Length = 16
	}
	if r.Count == 0 {
		r.Count = 1
	}
	if r.Length < MinPasswordLength || r.Length > MaxPasswordLength {
		return tools.InvalidField("length",
			[]int{MinPasswordLength, MaxPasswordLength}, r.Length)
	}
	if r.Count < 1 || r.Count > MaxPasswordCount {
		return tools.InvalidField("count", []int{1, MaxPasswordCount}, r.Count)
	}
	if r.Template != "" && !containsString(PasswordTemplate(), r.Template) {
		return tools.InvalidField("template", PasswordTemplate(), r.Template)
	}
	if r.WordCount < 0 || r.WordCount > 12 {
		return tools.InvalidField("wordCount", []int{0, 12}, r.WordCount)
	}
	if len(r.ExcludeChars) > 64 {
		return tools.InvalidField("excludeChars", []int{0, 64}, len(r.ExcludeChars))
	}
	if len(r.Separator) > 16 {
		return tools.InvalidField("separator", []int{0, 16}, len(r.Separator))
	}

	// 套用策略模板：模板给出字符集与约束的默认值。
	switch r.Template {
	case PasswordTemplateSimple:
		r.applyTemplate([]string{CharsetLower, CharsetUpper, CharsetDigit}, false)
	case PasswordTemplateStrong:
		r.applyTemplate([]string{CharsetLower, CharsetUpper, CharsetDigit, CharsetSymbol}, true)
	case PasswordTemplatePin:
		r.applyTemplate([]string{CharsetDigit}, false)
	case PasswordTemplatePassphrase:
		// 可发音口令走独立的单词表，不使用字符池。
		if r.WordCount == 0 {
			r.WordCount = 3
		}
		if r.Length < 8 || r.Length > 64 {
			return tools.InvalidField("length", []int{8, 64}, r.Length)
		}
	}

	if len(r.Charsets) == 0 && r.CustomCharset == "" {
		r.Charsets = []string{CharsetLower, CharsetUpper, CharsetDigit, CharsetSymbol}
	}
	for _, cs := range r.Charsets {
		switch cs {
		case CharsetLower, CharsetUpper, CharsetDigit, CharsetSymbol:
		default:
			return tools.InvalidField("charsets",
				[]string{CharsetLower, CharsetUpper, CharsetDigit, CharsetSymbol}, cs)
		}
	}
	if r.CustomCharset != "" && r.RequireEachClass {
		return tools.InvalidField("requireEachClass", false, true)
	}
	return nil
}

// applyTemplate 套用策略模板的默认参数。
//
// 模板只填充**未显式指定**的字段，用户显式给出的值优先。
func (r *PasswordGenRequest) applyTemplate(charsets []string, requireEachClass bool) {
	if len(r.Charsets) == 0 && r.CustomCharset == "" {
		r.Charsets = charsets
	}
	if !r.RequireEachClass {
		r.RequireEachClass = requireEachClass
	}
}

// PasswordGen 生成随机密码。
//
// 随机源是 Tongsuo 的 CSPRNG；字符选择使用**拒绝采样**避免模偏差。
func PasswordGen(_ context.Context, req *PasswordGenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	pool, classes, err := buildPool(req)
	if err != nil {
		return tools.Result{}, err
	}

	passwords := make([]string, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		pwd, err := generatePassword(req, pool, classes)
		if err != nil {
			return tools.Result{}, err
		}
		passwords = append(passwords, pwd)
	}

	entropy := 0.0
	if len(pool) > 1 {
		entropy = float64(req.Length) * math.Log2(float64(len(pool)))
	}

	return tools.Result{Data: PasswordGenResult{
		Passwords:   passwords,
		EntropyBits: entropy,
		Template:    req.Template,
		PoolSize:    len(pool),
	}}, nil
}

// buildPool 构造字符池与各类字符集。
func buildPool(req *PasswordGenRequest) (string, []string, error) {
	if req.CustomCharset != "" {
		pool := req.CustomCharset
		if req.ExcludeAmbiguous {
			pool = removeChars(pool, ambiguousChars)
		}
		pool = removeChars(pool, req.ExcludeChars)
		if len(pool) == 0 {
			return "", nil, tools.InvalidField("customCharset", "non-empty", "")
		}
		return pool, nil, nil
	}

	classes := make([]string, 0, len(req.Charsets))
	pool := ""
	for _, cs := range req.Charsets {
		var chars string
		switch cs {
		case CharsetLower:
			chars = charsetLowerChars
		case CharsetUpper:
			chars = charsetUpperChars
		case CharsetDigit:
			chars = charsetDigitChars
		case CharsetSymbol:
			chars = charsetSymbolChars
		}
		if req.ExcludeAmbiguous {
			chars = removeChars(chars, ambiguousChars)
		}
		chars = removeChars(chars, req.ExcludeChars)
		if chars == "" {
			continue
		}
		classes = append(classes, chars)
		pool += chars
	}
	if len(classes) == 0 {
		return "", nil, tools.InvalidField("charsets", "non-empty after filtering", req.Charsets)
	}
	return pool, classes, nil
}

// generatePassword 生成一个密码。
func generatePassword(req *PasswordGenRequest, pool string, classes []string) (string, error) {
	// 可发音口令走独立路径：它由单词表拼成，不使用字符池。
	if req.Template == PasswordTemplatePassphrase {
		return generatePassphrase(req)
	}

	out := make([]byte, 0, req.Length)

	// 每类至少一个。
	if req.RequireEachClass && len(classes) > 0 {
		if len(classes) > req.Length {
			return "", tools.InvalidField("length",
				"at least the number of character classes", req.Length)
		}
		for _, class := range classes {
			c, err := pick(class)
			if err != nil {
				return "", err
			}
			out = append(out, c)
		}
	}

	for len(out) < req.Length {
		c, err := pick(pool)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}

	// 打乱顺序，使「每类一个」的固定位置不泄露。
	if err := shuffle(out); err != nil {
		return "", err
	}
	return string(out), nil
}

// pick 用拒绝采样从字符集中均匀选一个字符。
func pick(charset string) (byte, error) {
	n := len(charset)
	if n == 0 {
		return 0, tools.InvalidField("charset", "non-empty", "")
	}
	// 拒绝采样上界：最大的 n 的倍数，避免取模偏差。
	limit := (256 / n) * n
	buf := make([]byte, 1)
	for {
		raw, err := cryptosvc.RandomBytes(len(buf))
		if err != nil {
			return 0, err
		}
		v := int(raw[0])
		if v < limit {
			return charset[v%n], nil
		}
	}
}

// shuffle 用 Fisher-Yates 洗牌。
func shuffle(b []byte) error {
	for i := len(b) - 1; i > 0; i-- {
		j, err := randomIndex(i + 1)
		if err != nil {
			return err
		}
		b[i], b[j] = b[j], b[i]
	}
	return nil
}

// randomIndex 返回 [0, n) 内的均匀随机下标。
func randomIndex(n int) (int, error) {
	if n <= 1 {
		return 0, nil
	}
	limit := (256 / n) * n
	for {
		raw, err := cryptosvc.RandomBytes(1)
		if err != nil {
			return 0, err
		}
		v := int(raw[0])
		if v < limit {
			return v % n, nil
		}
	}
}

// removeChars 从字符集中移除指定字符。
func removeChars(charset, remove string) string {
	out := make([]byte, 0, len(charset))
	for i := 0; i < len(charset); i++ {
		if !containsRune(remove, charset[i]) {
			out = append(out, charset[i])
		}
	}
	return string(out)
}

// containsRune 报告字符串是否含有该字节。
func containsRune(s string, b byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return true
		}
	}
	return false
}
