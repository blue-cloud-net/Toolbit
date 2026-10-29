package text

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 正则生成器的结构化选项。
//
// 每项对应生成器参数区的一个字段；未勾选的项不出现在结果里。
const (
	// RegexPartPrefix 是固定前缀。
	RegexPartPrefix = "prefix"
	// RegexPartAnyChar 是任意字符。
	RegexPartAnyChar = "anyChar"
	// RegexPartDigits 是数字。
	RegexPartDigits = "digits"
	// RegexPartLetters 是字母。
	RegexPartLetters = "letters"
	// RegexPartWord 是单词字符。
	RegexPartWord = "word"
	// RegexPartSpace 是空白字符。
	RegexPartSpace = "space"
	// RegexPartUpper 是大写字母。
	RegexPartUpper = "upper"
	// RegexPartLower 是小写字母。
	RegexPartLower = "lower"
	// RegexPartChinese 是中文字符。
	RegexPartChinese = "chinese"
	// RegexPartHyphen 是连字符。
	RegexPartHyphen = "hyphen"
	// RegexPartDot 是点号。
	RegexPartDot = "dot"
	// RegexPartAt 是邮箱符号。
	RegexPartAt = "at"
	// RegexPartSlash 是斜杠。
	RegexPartSlash = "slash"
	// RegexPartSuffix 是固定后缀。
	RegexPartSuffix = "suffix"
)

// RegexPart 返回受支持的生成部件（顺序稳定）。
func RegexPart() []string {
	return []string{
		RegexPartPrefix, RegexPartAnyChar, RegexPartDigits, RegexPartLetters,
		RegexPartWord, RegexPartSpace, RegexPartUpper, RegexPartLower,
		RegexPartChinese, RegexPartHyphen, RegexPartDot, RegexPartAt,
		RegexPartSlash, RegexPartSuffix,
	}
}

// RegexQuantifier 是量词。
type RegexQuantifier struct {
	// Min 是最小重复次数。
	Min int `json:"min"`
	// Max 是最大重复次数；-1 表示无上限。
	Max int `json:"max"`
}

// RegexBuildRequest 是 regex-build 工具的请求。
type RegexBuildRequest struct {
	// Parts 是启用的部件列表。
	Parts []string `json:"parts"`
	// Prefix 是固定前缀文本。
	Prefix string `json:"prefix"`
	// Suffix 是固定后缀文本。
	Suffix string `json:"suffix"`
	// Quantifier 是统一的量词设置；不填则不套量词。
	Quantifier *RegexQuantifier `json:"quantifier"`
	// Anchored 是否在首尾加锚点 ^ $。
	Anchored bool `json:"anchored"`
}

// RegexBuildResult 是 regex-build 工具的响应。
type RegexBuildResult struct {
	// Pattern 是生成的表达式。
	Pattern string `json:"pattern"`
	// Valid 是否是合法表达式；生成结果按 RE2 方言校验。
	Valid bool `json:"valid"`
	// Explanation 是逐段释义。
	Explanation []RegexPartExplain `json:"explanation"`
	// Error 是校验错误；合法时为空。
	Error *RegexError `json:"error"`
	// CaptureGroups 是捕获组数量。
	CaptureGroups int `json:"captureGroups"`
}

// RegexPartExplain 是一段释义。
type RegexPartExplain struct {
	// Part 是部件标识。
	Part string `json:"part"`
	// Fragment 是对应的表达式片段。
	Fragment string `json:"fragment"`
	// Description 是片段含义。
	Description string `json:"description"`
}

// 部件到片段与释义的映射。
var regexPartTemplates = map[string]struct {
	fragment string
	desc     string
}{
	RegexPartAnyChar: {".", "任意字符（除换行外）"},
	RegexPartDigits:  {`\d`, "数字"},
	RegexPartLetters: {`[A-Za-z]`, "英文字母"},
	RegexPartWord:    {`\w`, "单词字符（字母、数字、下划线）"},
	RegexPartSpace:   {`\s`, "空白字符"},
	RegexPartUpper:   {`[A-Z]`, "大写字母"},
	RegexPartLower:   {`[a-z]`, "小写字母"},
	RegexPartChinese: {`[\p{Han}]`, "中文字符"},
	RegexPartHyphen:  {`\-`, "连字符"},
	RegexPartDot:     {`\.`, "点号"},
	RegexPartAt:      {`@`, "邮箱符号"},
	RegexPartSlash:   {`\/`, "斜杠"},
	RegexPartPrefix:  {"", "固定前缀"},
	RegexPartSuffix:  {"", "固定后缀"},
}

// RegexBuild 由结构化选项拼装正则表达式。
//
// 生成是纯字符串拼装，不涉及密码学运算；产出结果仍按 RE2 方言校验，
// 非法组合即时报错。
func RegexBuild(_ context.Context, req *RegexBuildRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	explanation := []RegexPartExplain{}
	fragments := []string{}

	if req.Prefix != "" {
		fragments = append(fragments, regexp.QuoteMeta(req.Prefix))
		explanation = append(explanation, RegexPartExplain{
			Part:        RegexPartPrefix,
			Fragment:    regexp.QuoteMeta(req.Prefix),
			Description: "固定前缀",
		})
	}

	quantifier := ""
	if req.Quantifier != nil {
		quantifier = buildQuantifier(req.Quantifier)
	}

	for _, part := range req.Parts {
		template, ok := regexPartTemplates[part]
		if !ok {
			continue
		}
		if part == RegexPartPrefix || part == RegexPartSuffix {
			continue
		}
		fragment := template.fragment
		if quantifier != "" {
			fragment += quantifier
		}
		fragments = append(fragments, fragment)
		explanation = append(explanation, RegexPartExplain{
			Part:        part,
			Fragment:    fragment,
			Description: template.desc,
		})
	}

	if req.Suffix != "" {
		fragments = append(fragments, regexp.QuoteMeta(req.Suffix))
		explanation = append(explanation, RegexPartExplain{
			Part:        RegexPartSuffix,
			Fragment:    regexp.QuoteMeta(req.Suffix),
			Description: "固定后缀",
		})
	}

	pattern := strings.Join(fragments, "")
	if req.Anchored && pattern != "" {
		pattern = "^" + pattern + "$"
	}

	result := RegexBuildResult{
		Pattern:     pattern,
		Valid:       true,
		Explanation: explanation,
		Error:       nil,
	}
	if pattern == "" {
		result.Valid = false
		result.Error = &RegexError{
			Code:     "empty",
			Message:  "未选择任何部件，生成的表达式为空",
			Offset:   -1,
			Position: "",
		}
		return tools.Result{Data: result}, nil
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		result.Valid = false
		result.Error = buildRegexError(pattern, err)
		return tools.Result{Data: result}, nil
	}
	result.CaptureGroups = compiled.NumSubexp()
	return tools.Result{Data: result}, nil
}

// buildQuantifier 把量词设置渲染成表达式后缀。
func buildQuantifier(q *RegexQuantifier) string {
	if q.Max < 0 {
		return "{" + strconv.Itoa(q.Min) + ",}"
	}
	if q.Min == q.Max {
		return "{" + strconv.Itoa(q.Min) + "}"
	}
	return "{" + strconv.Itoa(q.Min) + "," + strconv.Itoa(q.Max) + "}"
}

// Validate 做字段级校验。
func (r *RegexBuildRequest) Validate() error {
	supported := RegexPart()
	for _, part := range r.Parts {
		if !containsString(supported, part) {
			return tools.InvalidField("parts", supported, part)
		}
	}
	if len(r.Prefix) > 256 {
		return tools.InvalidField("prefix", []int{0, 256}, len(r.Prefix))
	}
	if len(r.Suffix) > 256 {
		return tools.InvalidField("suffix", []int{0, 256}, len(r.Suffix))
	}
	if r.Quantifier != nil {
		if r.Quantifier.Min < 0 || r.Quantifier.Min > 100 {
			return tools.InvalidField("quantifier.min", []int{0, 100}, r.Quantifier.Min)
		}
		if r.Quantifier.Max < -1 || r.Quantifier.Max > 100 {
			return tools.InvalidField("quantifier.max", []int{-1, 100}, r.Quantifier.Max)
		}
		if r.Quantifier.Max >= 0 && r.Quantifier.Min > r.Quantifier.Max {
			return tools.InvalidField("quantifier.min", []int{0, r.Quantifier.Max}, r.Quantifier.Min)
		}
	}
	return nil
}
