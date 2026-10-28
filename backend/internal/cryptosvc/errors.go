package cryptosvc

import (
	"errors"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/digest"
	"github.com/blue-cloud-net/tongsuo-go/ecdh"
	"github.com/blue-cloud-net/tongsuo-go/sym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// MapError 把底层错误映射为 toolerr 错误码。
//
// 映射结果**只含**结构化元信息，**不含**用户输入，
// 也不回显底层库的英文错误文本（见 docs/standards/04-errors.md）。
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
	case errors.Is(err, asym.ErrSignature):
		// 验签不通过。调用方（签名服务）会先把它当作**业务结果**处理，
		// 走到这里说明签名结构本身非法。
		return toolerr.Wrap(toolerr.CodeSignatureInvalid, err)
	case errors.Is(err, asym.ErrInvalidKey):
		return toolerr.Wrap(toolerr.CodeInvalidKey, err)
	case errors.Is(err, asym.ErrUnknownAlgorithm), errors.Is(err, asym.ErrUnsupported):
		return toolerr.Wrap(toolerr.CodeUnsupportedAlgorithm, err)
	case errors.Is(err, digest.ErrUnknownAlgorithm):
		return toolerr.Wrap(toolerr.CodeUnsupportedAlgorithm, err)
	case errors.Is(err, sym.ErrInvalidKeyLength):
		return toolerr.Wrap(toolerr.CodeKeySizeMismatch, err)
	case errors.Is(err, sym.ErrUnknownAlgorithm):
		return toolerr.Wrap(toolerr.CodeUnsupportedAlgorithm, err)
	case errors.Is(err, sym.ErrUnsupported):
		return toolerr.Wrap(toolerr.CodeUnsupportedMode, err)
	case errors.Is(err, ecdh.ErrUnknownCurve):
		return toolerr.Wrap(toolerr.CodeUnsupportedAlgorithm, err)
	case errors.Is(err, ecdh.ErrUnsupportedKey):
		return toolerr.Wrap(toolerr.CodeInvalidKey, err)
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

// unsupportedMode 构造「模式不支持」错误。
func unsupportedMode(mode string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedMode, map[string]any{
		"mode":      mode,
		"supported": supported,
	})
}

// unsupportedPadding 构造「填充不支持」错误。
func unsupportedPadding(padding string, supported []string) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeUnsupportedPadding, map[string]any{
		"padding":   padding,
		"supported": supported,
	})
}
