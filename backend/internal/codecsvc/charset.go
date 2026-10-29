package codecsvc

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 字符集名。
const (
	// CharsetAuto 表示自动探测源字符集。
	CharsetAuto = "auto"
	// CharsetUTF8 是 UTF-8。
	CharsetUTF8 = "utf-8"
	// CharsetUTF16 是 UTF-16：BOM 决定字节序，无 BOM 时按小端。
	CharsetUTF16 = "utf-16"
	// CharsetUTF16LE 是 UTF-16 小端。
	CharsetUTF16LE = "utf-16le"
	// CharsetUTF16BE 是 UTF-16 大端。
	CharsetUTF16BE = "utf-16be"
	// CharsetGBK 是 GBK（覆盖 GB2312 常用子集）。
	CharsetGBK = "gbk"
	// CharsetBig5 是 Big5。
	CharsetBig5 = "big5"
	// CharsetShiftJIS 是 Shift_JIS。
	CharsetShiftJIS = "shift_jis"
	// CharsetLatin1 是 ISO-8859-1（Latin-1）。
	CharsetLatin1 = "latin-1"
)

// 字节序标记。
var (
	// bomUTF8 是 UTF-8 的 BOM。
	bomUTF8 = []byte{0xEF, 0xBB, 0xBF}
	// bomUTF16LE 是 UTF-16 小端的 BOM。
	bomUTF16LE = []byte{0xFF, 0xFE}
	// bomUTF16BE 是 UTF-16 大端的 BOM。
	bomUTF16BE = []byte{0xFE, 0xFF}
)

// Charsets 返回受支持的字符集名（顺序稳定）。
func Charsets() []string {
	return []string{
		CharsetAuto,
		CharsetUTF8,
		CharsetUTF16,
		CharsetUTF16LE,
		CharsetUTF16BE,
		CharsetGBK,
		CharsetBig5,
		CharsetShiftJIS,
		CharsetLatin1,
	}
}

// charsetAliases 是字符集别名表。
//
// 别名**不**出现在 `Charsets()` 里：对外暴露的取值集合是封闭的，
// 别名只是输入侧的宽容处理。
var charsetAliases = map[string]string{
	"utf8":        CharsetUTF8,
	"utf-8":       CharsetUTF8,
	"utf16":       CharsetUTF16,
	"utf-16":      CharsetUTF16,
	"utf16le":     CharsetUTF16LE,
	"utf-16le":    CharsetUTF16LE,
	"utf16be":     CharsetUTF16BE,
	"utf-16be":    CharsetUTF16BE,
	"gbk":         CharsetGBK,
	"gb2312":      CharsetGBK,
	"gb18030":     CharsetGBK,
	"big5":        CharsetBig5,
	"big-5":       CharsetBig5,
	"shift_jis":   CharsetShiftJIS,
	"shift-jis":   CharsetShiftJIS,
	"sjis":        CharsetShiftJIS,
	"latin1":      CharsetLatin1,
	"latin-1":     CharsetLatin1,
	"iso-8859-1":  CharsetLatin1,
	"iso8859-1":   CharsetLatin1,
	"utf8bom":     CharsetUTF8,
	"utf-8-bom":   CharsetUTF8,
	"utf16bom":    CharsetUTF16,
	"utf-16-bom":  CharsetUTF16,
	"unicode":     CharsetUTF16,
	"utf-16-le":   CharsetUTF16LE,
	"utf-16-be":   CharsetUTF16BE,
	"windows-31j": CharsetShiftJIS,
}

// charsetSpec 描述一个字符集。
type charsetSpec struct {
	// dec 把源字符集的字节解码为 UTF-8。
	dec encoding.Encoding
	// enc 把 UTF-8 编码为目标字符集。
	enc encoding.Encoding
	// bom 是该字符集的字节序标记；nil 表示该字符集没有 BOM。
	bom []byte
	// utf16 为真表示按 UTF-16 族处理：BOM 决定字节序。
	utf16 bool
}

// charsetSpecs 是字符集表。
var charsetSpecs = map[string]charsetSpec{
	CharsetUTF8: {
		dec: unicode.UTF8,
		enc: unicode.UTF8,
		bom: bomUTF8,
	},
	CharsetUTF16: {
		dec:   unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
		enc:   unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
		bom:   bomUTF16LE,
		utf16: true,
	},
	CharsetUTF16LE: {
		dec: unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
		enc: unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
		bom: bomUTF16LE,
	},
	CharsetUTF16BE: {
		dec: unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),
		enc: unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),
		bom: bomUTF16BE,
	},
	CharsetGBK: {
		dec: simplifiedchinese.GBK,
		enc: simplifiedchinese.GBK,
	},
	CharsetBig5: {
		dec: traditionalchinese.Big5,
		enc: traditionalchinese.Big5,
	},
	CharsetShiftJIS: {
		dec: japanese.ShiftJIS,
		enc: japanese.ShiftJIS,
	},
	CharsetLatin1: {
		dec: charmap.ISO8859_1,
		enc: charmap.ISO8859_1,
	},
}

// NormalizeCharset 归一化字符集名。
//
// 大小写与分隔符不敏感；`utf8` 之类的常见别名会被归一化；
// 未登记的字符集返回 INVALID_ENCODING。
func NormalizeCharset(name string) (string, error) {
	key := normalizeCharsetKey(name)
	if key == CharsetAuto {
		return CharsetAuto, nil
	}
	if canonical, ok := charsetAliases[key]; ok {
		return canonical, nil
	}
	return "", toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
		"encoding":  name,
		"supported": Charsets(),
	})
}

// normalizeCharsetKey 把字符集名归一化成比较用的键。
func normalizeCharsetKey(name string) string {
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'A' && c <= 'Z':
			out = append(out, c+('a'-'A'))
		case c == '_' || c == ' ':
			out = append(out, '-')
		default:
			out = append(out, c)
		}
	}
	return string(out)
}

// DetectCharset 探测字节序列的源字符集。
//
// 探测顺序：BOM → UTF-8 合法性。
//
// 两者都不成立时**不**猜测：GBK、Big5 等双字节编码几乎覆盖任意字节序列，
// 猜测结果不可靠，因此要求调用方显式指定源字符集。
func DetectCharset(raw []byte) (string, error) {
	switch {
	case bytes.HasPrefix(raw, bomUTF8):
		return CharsetUTF8, nil
	case bytes.HasPrefix(raw, bomUTF16LE), bytes.HasPrefix(raw, bomUTF16BE):
		return CharsetUTF16, nil
	case utf8.Valid(raw):
		return CharsetUTF8, nil
	default:
		return "", toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
			"reason": "cannot detect source charset; specify it explicitly",
		})
	}
}

// CharsetResult 是字符集转换的结果。
type CharsetResult struct {
	// Output 是按目标字符集编码的字节。
	Output []byte
	// Detected 是自动探测出的源字符集；未探测时为空串。
	Detected string
	// Lossy 为真表示转换过程中发生了替换（有信息丢失）。
	Lossy bool
}

// ConvertCharset 把字节从源字符集转换到目标字符集。
//
// from 为 `auto` 时自动探测源字符集。
//
// 非法字节与不可映射字符**不**导致失败：一律被替换，
// 并把 Lossy 置为真——转换结果因此可能与原文不完全等价。
func ConvertCharset(raw []byte, from, to string, bom bool) (*CharsetResult, error) {
	fromName, err := NormalizeCharset(from)
	if err != nil {
		return nil, err
	}
	toName, err := NormalizeCharset(to)
	if err != nil {
		return nil, err
	}
	if toName == CharsetAuto {
		return nil, toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
			"encoding":  to,
			"supported": Charsets(),
		})
	}

	detected := ""
	if fromName == CharsetAuto {
		detected, err = DetectCharset(raw)
		if err != nil {
			return nil, err
		}
		fromName = detected
	}

	fromSpec := charsetSpecs[fromName]
	toSpec := charsetSpecs[toName]

	text, replaced, err := decodeCharset(fromSpec, raw)
	if err != nil {
		return nil, err
	}
	out, lossyOut, err := encodeCharset(toSpec, text, bom)
	if err != nil {
		return nil, err
	}

	return &CharsetResult{
		Output:   out,
		Detected: detected,
		Lossy:    replaced || lossyOut,
	}, nil
}

// decodeCharset 把源字符集的字节解码为 UTF-8 文本。
//
// 第二个返回值标记是否发生了替换。
func decodeCharset(spec charsetSpec, raw []byte) (string, bool, error) {
	src := raw
	dec := spec.dec

	if spec.utf16 {
		// UTF-16 族：BOM 决定字节序，无 BOM 时按小端。
		switch {
		case bytes.HasPrefix(src, bomUTF16LE):
			src = src[len(bomUTF16LE):]
			dec = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		case bytes.HasPrefix(src, bomUTF16BE):
			src = src[len(bomUTF16BE):]
			dec = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
		}
	} else if spec.bom != nil && bytes.HasPrefix(src, spec.bom) {
		src = src[len(spec.bom):]
	}

	return decodeReplacing(dec, src)
}

// decodeReplacing 把字节解码为 UTF-8 字符串。
//
// 非法字节序列**不**导致失败：解码器把它们表示为 U+FFFD。
// 因此是否发生替换只能从**输出**上判断：结果中含 U+FFFD 即视为有损。
//
// 注意：原文本身合法地包含 U+FFFD 字符时也会命中该判断——
// 这在字节层面本就无法区分，属于编码格式的固有性质。
func decodeReplacing(enc encoding.Encoding, src []byte) (string, bool, error) {
	dec := enc.NewDecoder()
	var out bytes.Buffer

	rest := src
	for len(rest) > 0 {
		// 输出缓冲区按「一个源字节最多产出 3 字节 UTF-8」估算；
		// 真的不够时由内层循环扩容，避免把缓冲区不足误当成非法字节。
		size := len(rest)*3 + 64
		var nDst, nSrc int
		var err error
		for {
			dst := make([]byte, size)
			nDst, nSrc, err = dec.Transform(dst, rest, true)
			if err != transform.ErrShortDst {
				out.Write(dst[:nDst])
				break
			}
			size *= 2
		}

		rest = rest[nSrc:]
		if err == nil {
			break
		}

		// 出错的那个字节无法解码：替换后跳过它继续。
		out.WriteRune(utf8.RuneError)
		rest = rest[1:]
		dec.Reset()
	}

	text := out.String()
	return text, strings.ContainsRune(text, utf8.RuneError), nil
}

// encodeCharset 把 UTF-8 文本编码为目标字符集。
//
// 不可映射的字符被替换为该字符集的替换字节（通常是 `?`），
// 第二个返回值标记是否发生了替换。
func encodeCharset(spec charsetSpec, text string, bom bool) ([]byte, bool, error) {
	strict, _, err := transform.Bytes(spec.enc.NewEncoder(), []byte(text))
	if err == nil {
		return prependBOM(strict, spec.bom, bom), false, nil
	}

	lossy, _, err := transform.Bytes(
		encoding.ReplaceUnsupported(spec.enc.NewEncoder()),
		[]byte(text),
	)
	if err != nil {
		return nil, false, toolerr.Newf(toolerr.CodeInvalidEncoding, map[string]any{
			"reason": "failed to encode text",
		})
	}
	return prependBOM(lossy, spec.bom, bom), true, nil
}

// prependBOM 按需在数据前加上字节序标记。
func prependBOM(data []byte, bom []byte, want bool) []byte {
	if !want || bom == nil {
		return data
	}
	out := make([]byte, 0, len(bom)+len(data))
	out = append(out, bom...)
	out = append(out, data...)
	return out
}
