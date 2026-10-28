package crypto

import "encoding/hex"

// hexEncode 把字节编码为小写十六进制。
//
// 只用于把 GCM 认证标签输出为文本；大块数据一律走 codecsvc.Encode。
func hexEncode(b []byte) string {
	return hex.EncodeToString(b)
}
