package text

import (
	"context"
	"regexp"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 脱敏类型。
const (
	// MaskTypePhone 是手机号。
	MaskTypePhone = "phone"
	// MaskTypeIDCard 是身份证号。
	MaskTypeIDCard = "idCard"
	// MaskTypeBankCard 是银行卡号。
	MaskTypeBankCard = "bankCard"
	// MaskTypeEmail 是邮箱。
	MaskTypeEmail = "email"
	// MaskTypeName 是姓名。
	MaskTypeName = "name"
	// MaskTypeIP 是 IP 地址。
	MaskTypeIP = "ip"
)

// MaskType 返回受支持的脱敏类型（顺序稳定）。
func MaskType() []string {
	return []string{
		MaskTypePhone, MaskTypeIDCard, MaskTypeBankCard,
		MaskTypeEmail, MaskTypeName, MaskTypeIP,
	}
}

// 脱敏位置策略。
const (
	// MaskKeepHead 保留开头若干字符。
	MaskKeepHead = "keepHead"
	// MaskKeepTail 保留结尾若干字符。
	MaskKeepTail = "keepTail"
	// MaskKeepBoth 保留首尾各若干字符。
	MaskKeepBoth = "keepBoth"
	// MaskHash 全量替换为哈希摘要。
	MaskHash = "hash"
)

// MaskPosition 返回受支持的保留策略（顺序稳定）。
func MaskPosition() []string {
	return []string{MaskKeepHead, MaskKeepTail, MaskKeepBoth, MaskHash}
}

// DataMaskRequest 是 data-mask 工具的请求。
type DataMaskRequest struct {
	// Text 是待脱敏文本。
	Text string `json:"text"`
	// Types 是要脱敏的类型列表；为空表示全部类型。
	Types []string `json:"types"`
	// Position 是保留策略；默认 keepHead。
	Position string `json:"position"`
	// KeepCount 是保留的字符数；默认 3。
	KeepCount int `json:"keepCount"`
}

// MaskMatch 是一处脱敏记录。
type MaskMatch struct {
	// Type 是脱敏类型。
	Type string `json:"type"`
	// Count 是命中次数。
	Count int `json:"count"`
	// Sample 是脱敏后的样例；不回显原文。
	Sample string `json:"sample"`
}

// DataMaskResult 是 data-mask 工具的响应。
type DataMaskResult struct {
	// Masked 是脱敏后的文本。
	Masked string `json:"masked"`
	// Matches 是各类脱敏的统计。
	Matches []MaskMatch `json:"matches"`
	// Total 是脱敏总数。
	Total int `json:"total"`
}

// maskRule 是一条脱敏规则。
type maskRule struct {
	// kind 是脱敏类型。
	kind string
	// pattern 是识别用的正则。
	pattern *regexp.Regexp
}

// 各类敏感信息的识别正则。
//
// 顺序有意义：先匹配更具体的长模式，再匹配宽松模式。
var maskRules = []maskRule{
	{
		kind:    MaskTypeIDCard,
		pattern: regexp.MustCompile(`[1-9]\d{5}(?:19|20)\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01])\d{3}[\dXx]`),
	},
	{
		kind:    MaskTypeBankCard,
		pattern: regexp.MustCompile(`\b\d{16,19}\b`),
	},
	{
		kind:    MaskTypePhone,
		pattern: regexp.MustCompile(`1[3-9]\d{9}`),
	},
	{
		kind:    MaskTypeEmail,
		pattern: regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`),
	},
	{
		kind:    MaskTypeIP,
		pattern: regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
	},
}

// 姓名的匹配规则：中文姓名或英文姓名。
var maskNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`[\p{Han}]{2,4}`),
	regexp.MustCompile(`\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`),
}

// maskAsterisk 是遮蔽字符。
const maskAsterisk = "*"

// maxMaskTextBytes 是待脱敏文本的字节上限。
const maxMaskTextBytes = 256 * 1024

// DataMask 对文本中的敏感信息做脱敏。
//
// 脱敏在**服务端**完成：文本不会离开后端，返回的只有脱敏结果与统计。
func DataMask(_ context.Context, req *DataMaskRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	position := req.Position
	if position == "" {
		position = MaskKeepHead
	}
	keep := req.KeepCount
	if keep == 0 {
		keep = 3
	}

	enabled := map[string]bool{}
	if len(req.Types) == 0 {
		for _, kind := range MaskType() {
			enabled[kind] = true
		}
	} else {
		for _, kind := range req.Types {
			enabled[kind] = true
		}
	}

	masked := req.Text
	stats := map[string]*MaskMatch{}
	order := []string{}

	applyRule := func(kind string, pattern *regexp.Regexp) {
		if !enabled[kind] {
			return
		}
		count := 0
		masked = pattern.ReplaceAllStringFunc(masked, func(match string) string {
			count++
			return maskValue(match, position, keep)
		})
		if count > 0 {
			stats[kind] = &MaskMatch{Type: kind, Count: count}
			order = append(order, kind)
		}
	}

	for _, rule := range maskRules {
		applyRule(rule.kind, rule.pattern)
	}

	if enabled[MaskTypeName] {
		for _, pattern := range maskNamePatterns {
			applyRule(MaskTypeName, pattern)
		}
	}

	matches := []MaskMatch{}
	total := 0
	for _, kind := range order {
		entry := stats[kind]
		entry.Sample = maskValue(maskSampleFor(kind, masked), position, keep)
		matches = append(matches, *entry)
		total += entry.Count
	}

	return tools.Result{Data: DataMaskResult{
		Masked:  masked,
		Matches: matches,
		Total:   total,
	}}, nil
}

// maskValue 按策略遮蔽一个值。
func maskValue(value string, position string, keep int) string {
	runes := []rune(value)
	length := len(runes)
	if length == 0 {
		return value
	}

	switch position {
	case MaskHash:
		return strings.Repeat(maskAsterisk, length)
	case MaskKeepTail:
		if length <= keep {
			return strings.Repeat(maskAsterisk, length)
		}
		return strings.Repeat(maskAsterisk, length-keep) + string(runes[length-keep:])
	case MaskKeepBoth:
		if length <= keep*2 {
			return strings.Repeat(maskAsterisk, length)
		}
		head := string(runes[:keep])
		tail := string(runes[length-keep:])
		return head + strings.Repeat(maskAsterisk, length-keep*2) + tail
	default:
		if length <= keep {
			return strings.Repeat(maskAsterisk, length)
		}
		return string(runes[:keep]) + strings.Repeat(maskAsterisk, length-keep)
	}
}

// maskSampleFor 造一个脱敏样例。
//
// 样例用**占位字符**而不是真实命中值，避免回显用户数据。
func maskSampleFor(kind string, _ string) string {
	switch kind {
	case MaskTypePhone:
		return "138****8000"
	case MaskTypeIDCard:
		return "110101********1234"
	case MaskTypeBankCard:
		return "6222****1234"
	case MaskTypeEmail:
		return "user@example.com"
	case MaskTypeIP:
		return "192.168.*.*"
	default:
		return "张*"
	}
}

// Validate 做字段级校验。
func (r *DataMaskRequest) Validate() error {
	if strings.TrimSpace(r.Text) == "" {
		return tools.MissingField("text")
	}
	if len(r.Text) > maxMaskTextBytes {
		return tools.InvalidField("text", []int{0, maxMaskTextBytes}, len(r.Text))
	}
	supported := MaskType()
	for _, kind := range r.Types {
		if !containsString(supported, kind) {
			return tools.InvalidField("types", supported, kind)
		}
	}
	if r.Position != "" && !containsString(MaskPosition(), r.Position) {
		return tools.InvalidField("position", MaskPosition(), r.Position)
	}
	if r.KeepCount < 0 || r.KeepCount > 32 {
		return tools.InvalidField("keepCount", []int{0, 32}, r.KeepCount)
	}
	return nil
}
