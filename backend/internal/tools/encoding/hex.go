package encoding

import (
	"encoding/hex"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// hexDecode 解析十六进制串。
//
// 非法字符 → INVALID_HEX（带字符下标）；奇数长度 → INVALID_HEX。
func hexDecode(s string) ([]byte, error) {
	cleaned := strings.TrimSpace(s)
	if cleaned == "" {
		return []byte{}, nil
	}
	if len(cleaned)%2 != 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidHex, map[string]any{
			"index":  len(cleaned) - 1,
			"reason": "odd length",
		})
	}
	raw, err := hex.DecodeString(cleaned)
	if err != nil {
		idx := hexErrorIndex(cleaned)
		return nil, toolerr.Newf(toolerr.CodeInvalidHex, map[string]any{
			"index":  idx,
			"reason": "invalid hex character",
		})
	}
	return raw, nil
}

// hexErrorIndex 定位首个非法十六进制字符的下标。
func hexErrorIndex(s string) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case '0' <= c && c <= '9':
		case 'a' <= c && c <= 'f':
		case 'A' <= c && c <= 'F':
		default:
			return i
		}
	}
	return len(s)
}

// hexEncode 把字节序列编码为小写十六进制。
func hexEncode(raw []byte) string {
	return hex.EncodeToString(raw)
}
