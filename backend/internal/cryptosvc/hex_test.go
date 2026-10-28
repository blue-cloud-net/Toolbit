package cryptosvc

import "encoding/hex"

// hexEncode 把字节编码为小写十六进制。
//
// 只用于测试断言，避免与 codecsvc 形成循环依赖。
func hexEncode(b []byte) string {
	return hex.EncodeToString(b)
}
