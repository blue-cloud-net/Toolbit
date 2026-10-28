package cryptosvc

import (
	tsrand "github.com/blue-cloud-net/tongsuo-go/rand"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// RandomBytes 返回 n 个密码学安全随机字节。
//
// 随机源是 Tongsuo 的 RAND_bytes（CSPRNG）。
// 严禁改用 math/rand（见 docs/standards/03-crypto.md §14）。
//
// n 为 0 时返回空切片；n 为负时返回 INVALID_FIELD。
func RandomBytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":   "length",
			"expected": "non-negative",
		})
	}
	b, err := tsrand.Bytes(n)
	if err != nil {
		return nil, MapError(err)
	}
	return b, nil
}
