package cryptosvc

import (
	"github.com/blue-cloud-net/tongsuo-go/digest"
)

// DigestAlgorithms 返回受支持的摘要算法名（顺序稳定）。
//
// 这是本服务的**白名单**：库内还有其它摘要实现，但本项目只暴露这几种。
func DigestAlgorithms() []string {
	return []string{"SM3", "MD5", "SHA1", "SHA224", "SHA256", "SHA384", "SHA512"}
}

// IsWeakDigest 报告该摘要算法是否属于**弱**算法（MD5 / SHA-1）。
//
// 弱算法仍可计算，但必须附带 WEAK_HASH 警告
// （见 docs/standards/03-crypto.md 与 docs/api/errors.md §6）。
func IsWeakDigest(algorithm string) bool {
	switch algorithm {
	case "MD5", "SHA1":
		return true
	default:
		return false
	}
}

// Digest 计算摘要。
//
// algorithm 大小写不敏感；未登记的算法返回 UNSUPPORTED_ALGORITHM。
func Digest(algorithm string, data []byte) ([]byte, error) {
	sum, err := digest.Sum(algorithm, data)
	if err != nil {
		return nil, MapError(err)
	}
	return sum, nil
}
