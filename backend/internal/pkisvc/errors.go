package pkisvc

import (
	"errors"

	"github.com/blue-cloud-net/tongsuo-go/asym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// MapError 把底层错误映射为 toolerr 错误码。
//
// 映射结果**只含**结构化元信息，**不含**用户输入，
// 也不回显底层库的英文错误文本（见 docs/standards/04-errors.md）。
//
// 本函数只覆盖 asym 的哨兵错误；JWK / XML / X.509 层的错误由各自入口
// 显式映射为 INVALID_JWK / INVALID_XML / INVALID_CERT 等更精确的码。
func MapError(err error) *toolerr.Error {
	if err == nil {
		return nil
	}

	// 已经是 toolerr 的原样返回。
	var toolErr *toolerr.Error
	if errors.As(err, &toolErr) {
		return toolErr
	}

	switch {
	case errors.Is(err, asym.ErrInvalidKey):
		return toolerr.Wrap(toolerr.CodeInvalidKey, err)
	case errors.Is(err, asym.ErrUnknownAlgorithm), errors.Is(err, asym.ErrUnsupported):
		return toolerr.Wrap(toolerr.CodeUnsupportedAlgorithm, err)
	default:
		return toolerr.Wrap(toolerr.CodeCryptoOperationFailed, err)
	}
}

// unsupportedAlgorithm 构造「算法不支持」错误，附带受支持的取值列表。
func unsupportedAlgorithm(algorithm string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedAlgorithm, map[string]any{
		"algorithm": algorithm,
		"supported": supported,
	})
}

// unsupportedFormat 构造「格式不支持」错误，附带受支持的取值列表。
func unsupportedFormat(format string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedFormat, map[string]any{
		"format":    format,
		"supported": supported,
	})
}

// toolerrInvalidKey 构造「密钥无效」错误。
//
// reason 是**枚举化**的失败原因，不含任何用户输入。
func toolerrInvalidKey(reason string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeInvalidKey, map[string]any{
		"reason": reason,
	})
}

// toolerrInvalidCert 构造「证书无效」错误。
//
// reason 是**枚举化**的失败原因，不含任何用户输入。
func toolerrInvalidCert(reason string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeInvalidCert, map[string]any{
		"reason": reason,
	})
}

// toolerrInvalidCSR 构造「证书签名请求无效」错误。
func toolerrInvalidCSR(reason string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeInvalidCSR, map[string]any{
		"reason": reason,
	})
}

// toolerrInvalidCRL 构造「吊销列表无效」错误。
func toolerrInvalidCRL(reason string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeInvalidCRL, map[string]any{
		"reason": reason,
	})
}

// toolerrInvalidPKCS12 构造「PKCS#12 容器无效」错误。
func toolerrInvalidPKCS12(reason string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeInvalidPKCS12, map[string]any{
		"reason": reason,
	})
}

// toolerrInvalidPKCS7 构造「PKCS#7 容器无效」错误。
func toolerrInvalidPKCS7(reason string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeInvalidPKCS7, map[string]any{
		"reason": reason,
	})
}
