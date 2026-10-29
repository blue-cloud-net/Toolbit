package crypto

import (
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
)

// passphraseWords 是可发音口令的单词表。
//
// 全部是**常见、易读、无歧义**的英文单词；不含真实人名、地名或
// 任何与凭据相关的词条。
var passphraseWords = []string{
	"amber", "anchor", "apple", "autumn", "basket", "beacon", "bishop",
	"boulder", "branch", "bridge", "brook", "candle", "canvas", "canyon",
	"cedar", "circle", "cobalt", "comet", "copper", "coral", "cosmos",
	"crater", "crystal", "daisy", "delta", "dolphin", "dragon", "ember",
	"falcon", "feather", "fern", "flint", "forest", "fountain", "galaxy",
	"garden", "glacier", "granite", "harbor", "harvest", "hazel", "heron",
	"hollow", "island", "ivory", "jasper", "jungle", "kernel", "lantern",
	"laurel", "lemon", "lilac", "lotus", "maple", "marble", "meadow",
	"meteor", "mint", "mosaic", "nebula", "nectar", "oasis", "olive",
	"onyx", "opal", "orbit", "otter", "pebble", "pepper", "pigeon",
	"pillar", "pineapple", "plum", "prairie", "quartz", "quiver", "raccoon",
	"radish", "rainbow", "raven", "reef", "ribbon", "river", "rocket",
	"rooster", "saffron", "sage", "sailor", "sandbar", "sapphire", "shadow",
	"sparrow", "spruce", "squash", "starling", "stone", "summit", "sunset",
	"thistle", "thunder", "timber", "topaz", "tulip", "tundra", "violet",
	"walnut", "willow", "winter", "wren", "yarrow", "zephyr",
}

// passphraseSeparators 是单词之间的分隔符。
var passphraseSeparators = []string{"-", ".", "_"}

// generatePassphrase 生成可发音口令。
//
// 由若干常见单词加数字与符号拼成；比纯随机字符更好读易记，
// 熵主要来自**单词数量**而不是字符集大小。
func generatePassphrase(req *PasswordGenRequest) (string, error) {
	words := req.WordCount
	if words == 0 {
		words = 3
	}

	parts := make([]string, 0, words+2)
	for i := 0; i < words; i++ {
		word, err := pickWord()
		if err != nil {
			return "", err
		}
		parts = append(parts, word)
	}

	separator := passphraseSeparators[0]
	if req.Separator != "" {
		separator = req.Separator
	}
	out := strings.Join(parts, separator)

	// 补足到目标长度：追加数字，必要时再追加符号。
	if len(out) < req.Length {
		digit, err := pick(charsetDigitChars)
		if err != nil {
			return "", err
		}
		out += string(digit)
	}
	for len(out) < req.Length {
		symbol, err := pick(charsetSymbolChars)
		if err != nil {
			return "", err
		}
		out += string(symbol)
	}

	// 目标长度比拼出的结果短时，从末尾截断。
	if len(out) > req.Length {
		out = out[:req.Length]
	}
	return out, nil
}

// pickWord 从单词表里随机选一个词。
func pickWord() (string, error) {
	raw, err := cryptosvc.RandomBytes(2)
	if err != nil {
		return "", err
	}
	index := int(raw[0])<<8 | int(raw[1])
	return passphraseWords[index%len(passphraseWords)], nil
}
