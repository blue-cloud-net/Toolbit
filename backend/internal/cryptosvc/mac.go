package cryptosvc

import (
	"github.com/blue-cloud-net/tongsuo-go/mac"
)

// MACAlgorithms 返回受支持的 HMAC 算法名（顺序稳定）。
//
// 库内还提供 CMAC / GMAC / KMAC 等其它 MAC，本项目本版只暴露 HMAC 族。
func MACAlgorithms() []string {
	return []string{
		"HMAC-SM3",
		"HMAC-MD5",
		"HMAC-SHA1",
		"HMAC-SHA224",
		"HMAC-SHA256",
		"HMAC-SHA384",
		"HMAC-SHA512",
	}
}

// MACAlgorithmForDigest 返回摘要算法对应的 HMAC 算法名。
//
// 例如 "SHA256" → "HMAC-SHA256"。摘要不被支持时返回空串。
func MACAlgorithmForDigest(digestName string) string {
	for _, name := range MACAlgorithms() {
		if name == "HMAC-"+digestName {
			return name
		}
	}
	return ""
}

// IsWeakMAC 报告该 HMAC 算法是否基于**弱**摘要（MD5 / SHA-1）。
func IsWeakMAC(algorithm string) bool {
	switch algorithm {
	case "HMAC-MD5", "HMAC-SHA1":
		return true
	default:
		return false
	}
}

// MAC 计算消息认证码。
//
// algorithm 大小写不敏感；未登记的算法返回 UNSUPPORTED_ALGORITHM。
func MAC(algorithm string, key, data []byte) ([]byte, error) {
	sum, err := mac.Sum(algorithm, key, data, nil)
	if err != nil {
		return nil, MapError(err)
	}
	return sum, nil
}
