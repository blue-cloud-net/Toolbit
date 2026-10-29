package text

import (
	"context"
	"strings"
	"unicode"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// TextStatsRequest 是 text-stats 工具的请求。
type TextStatsRequest struct {
	// Text 是待统计的文本。
	Text string `json:"text"`
}

// TextStatsResult 是 text-stats 工具的响应。
type TextStatsResult struct {
	// Characters 是字符总数（含空白）。
	Characters int `json:"characters"`
	// CharactersNoSpace 是不含空白的字符数。
	CharactersNoSpace int `json:"charactersNoSpace"`
	// Words 是词数；中文按字计、英文按空白分隔。
	Words int `json:"words"`
	// ChineseChars 是中文字符数。
	ChineseChars int `json:"chineseChars"`
	// EnglishWords 是英文单词数。
	EnglishWords int `json:"englishWords"`
	// Digits 是数字字符数。
	Digits int `json:"digits"`
	// Punctuation 是标点符号数。
	Punctuation int `json:"punctuation"`
	// Lines 是行数。
	Lines int `json:"lines"`
	// Paragraphs 是段落数（空行分隔）。
	Paragraphs int `json:"paragraphs"`
	// Bytes 是 UTF-8 字节数。
	Bytes int `json:"bytes"`
}

// maxTextStatsBytes 是待统计文本的字节上限。
const maxTextStatsBytes = 1024 * 1024

// TextStats 统计文本的字数、行数与字符构成。
func TextStats(_ context.Context, req *TextStatsRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	text := req.Text
	result := TextStatsResult{
		Characters:        len([]rune(text)),
		CharactersNoSpace: 0,
		Words:             0,
		ChineseChars:      0,
		EnglishWords:      0,
		Digits:            0,
		Punctuation:       0,
		Lines:             countLines(text),
		Paragraphs:        countParagraphs(text),
		Bytes:             len(text),
	}

	for _, r := range text {
		switch {
		case unicode.IsSpace(r):
			continue
		case unicode.IsDigit(r):
			result.Digits++
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			result.Punctuation++
		case unicode.Is(unicode.Han, r):
			result.ChineseChars++
		default:
			// 非空白、非数字、非标点的可打印字符计入词数总量。
		}
		result.CharactersNoSpace++
	}

	result.EnglishWords = countEnglishWords(text)
	// 中文按字计、英文按词计，两部分相加即总词数。
	result.Words = result.ChineseChars + result.EnglishWords

	return tools.Result{Data: result}, nil
}

// countLines 统计行数。
//
// 空文本算 0 行；带结尾换行的文本按实际行数计，不额外多算一行。
func countLines(text string) int {
	if text == "" {
		return 0
	}
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return len(lines)
}

// countParagraphs 统计段落数（空行分隔）。
func countParagraphs(text string) int {
	if strings.TrimSpace(text) == "" {
		return 0
	}
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	blocks := []string{}
	current := []string{}
	for _, line := range strings.Split(normalized, "\n") {
		if strings.TrimSpace(line) == "" {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
				current = []string{}
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return len(blocks)
}

// countEnglishWords 统计英文单词数。
//
// 按空白与标点切分，连续的字母数字算一个词。
func countEnglishWords(text string) int {
	count := 0
	inWord := false
	for _, r := range text {
		if unicode.IsLetter(r) && !unicode.Is(unicode.Han, r) {
			inWord = true
			continue
		}
		if unicode.IsDigit(r) {
			// 数字接在词尾时仍算同一个词（如 `abc123`）。
			inWord = true
			continue
		}
		if inWord {
			count++
			inWord = false
		}
	}
	if inWord {
		count++
	}
	return count
}

// Validate 做字段级校验。
func (r *TextStatsRequest) Validate() error {
	if r.Text == "" {
		return tools.MissingField("text")
	}
	if len(r.Text) > maxTextStatsBytes {
		return tools.InvalidField("text", []int{0, maxTextStatsBytes}, len(r.Text))
	}
	return nil
}
