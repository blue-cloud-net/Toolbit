package crypto

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 签名的操作类型。
const (
	// SignOpSign 是签名。
	SignOpSign = "sign"
	// SignOpVerify 是验签。
	SignOpVerify = "verify"
)

// SignatureRequest 是数字签名工具的请求。
type SignatureRequest struct {
	// Algorithm 是签名算法。
	Algorithm string `json:"algorithm"`
	// Operation 是 sign 或 verify。
	Operation string `json:"operation"`
	// Message 是待签名 / 待验签的消息。
	Message ByteInput `json:"message"`
	// Key 是私钥；sign 时必填。
	Key KeyInput `json:"key"`
	// PublicKey 是公钥；verify 时必填。
	PublicKey KeyInput `json:"publicKey"`
	// Signature 是签名值；verify 时必填。
	Signature ByteInput `json:"signature"`
	// Hash 是摘要名；仅 RSA 接受。
	Hash string `json:"hash"`
	// RSAPadding 是 RSA 填充；默认 pss。
	RSAPadding string `json:"rsaPadding"`
	// SM2UserID 是 SM2 用户标识符。
	SM2UserID string `json:"sm2UserId"`
}

// SignatureResult 是数字签名工具的响应。
type SignatureResult struct {
	// Valid 是验签结论；仅 verify 时返回。
	Valid *bool `json:"valid,omitempty"`
	// Signature 是签名值；仅 sign 时返回。
	Signature string `json:"signature,omitempty"`
	// Algorithm 是实际使用的算法。
	Algorithm string `json:"algorithm"`
	// Hash 是实际使用的摘要。
	Hash string `json:"hash"`
}

// Validate 做字段级校验。
func (r *SignatureRequest) Validate() error {
	switch r.Operation {
	case SignOpSign, SignOpVerify:
	default:
		return tools.InvalidField("operation", []string{SignOpSign, SignOpVerify}, r.Operation)
	}

	algorithms := cryptosvc.SignAlgorithms()
	if !containsString(algorithms, r.Algorithm) {
		return tools.UnsupportedAlgorithm(r.Algorithm, algorithms)
	}

	hashes := cryptosvc.SignHashes(r.Algorithm)
	if r.Hash != "" && !containsString(hashes, r.Hash) {
		return tools.UnsupportedAlgorithm(r.Algorithm+"/"+r.Hash, hashes)
	}

	paddings := cryptosvc.SignPaddings(r.Algorithm)
	if r.RSAPadding != "" && !containsString(paddings, r.RSAPadding) {
		return tools.UnsupportedAlgorithm(r.Algorithm+"/"+r.RSAPadding, paddings)
	}

	if err := r.Message.Validate("message"); err != nil {
		return err
	}
	if err := r.Key.Validate("key"); err != nil {
		return err
	}
	if err := r.PublicKey.Validate("publicKey"); err != nil {
		return err
	}
	if err := r.Signature.Validate("signature"); err != nil {
		return err
	}
	return nil
}

// Signature 执行签名或验签。
//
// 验签不通过返回 `valid: false`（HTTP 200），**不是**错误。
func Signature(_ context.Context, req *SignatureRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	message, err := req.Message.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	hash := req.Hash
	if hash == "" {
		if hashes := cryptosvc.SignHashes(req.Algorithm); len(hashes) > 0 {
			hash = hashes[0]
		}
	}

	if req.Operation == SignOpSign {
		return signMessage(req, message, hash)
	}
	return verifySignature(req, message, hash)
}

// signMessage 签名。
func signMessage(req *SignatureRequest, message []byte, hash string) (tools.Result, error) {
	pemBytes, err := req.Key.PrivateKeyPEM()
	if err != nil {
		return tools.Result{}, err
	}
	priv, err := cryptosvc.LoadPrivateKey(pemBytes, req.Key.Passphrase)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = priv.Close() }()

	sig, err := cryptosvc.Sign(req.Algorithm, hash, req.RSAPadding, priv, message, req.SM2UserID)
	if err != nil {
		return tools.Result{}, err
	}

	value, err := codecsvc.Encode(sig, codecsvc.EncodingBase64)
	if err != nil {
		return tools.Result{}, err
	}

	return tools.Result{Data: SignatureResult{
		Signature: value,
		Algorithm: req.Algorithm,
		Hash:      hash,
	}}, nil
}

// verifySignature 验签。
func verifySignature(req *SignatureRequest, message []byte, hash string) (tools.Result, error) {
	pemBytes, err := req.PublicKey.PublicKeyPEM()
	if err != nil {
		return tools.Result{}, err
	}
	pub, err := cryptosvc.LoadPublicKey(pemBytes)
	if err != nil {
		return tools.Result{}, err
	}

	sig, err := req.Signature.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	valid, err := cryptosvc.Verify(req.Algorithm, hash, req.RSAPadding, pub, message, sig, req.SM2UserID)
	if err != nil {
		return tools.Result{}, err
	}

	return tools.Result{Data: SignatureResult{
		Valid:     &valid,
		Algorithm: req.Algorithm,
		Hash:      hash,
	}}, nil
}
