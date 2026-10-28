package keys

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 密钥操作动作。
const (
	// ActionEncrypt 是用口令加密私钥。
	ActionEncrypt = "encrypt"
	// ActionDecrypt 是解密私钥。
	ActionDecrypt = "decrypt"
	// ActionExtractPublic 是提取配对公钥。
	ActionExtractPublic = "extract-public"
	// ActionChangePassphrase 是修改私钥口令。
	ActionChangePassphrase = "change-passphrase"
)

// KeyOpsActions 返回受支持的操作动作（顺序稳定）。
func KeyOpsActions() []string {
	return []string{
		ActionEncrypt,
		ActionDecrypt,
		ActionExtractPublic,
		ActionChangePassphrase,
	}
}

// KeyOpsRequest 是密钥运算工具的请求。
type KeyOpsRequest struct {
	// Action 是操作动作。
	Action string `json:"action"`
	// Key 是待操作的私钥。
	Key KeyInput `json:"key"`
	// Passphrase 是现有口令；encrypt 时是新口令，decrypt / change-passphrase 时是旧口令。
	Passphrase string `json:"passphrase"`
	// NewPassphrase 是新口令；仅 change-passphrase 时必填。
	NewPassphrase string `json:"newPassphrase"`
	// PbeCipher 是加密算法；默认 AES-256-CBC。
	PbeCipher string `json:"pbeCipher"`
}

// KeyOpsResult 是密钥运算工具的响应。
type KeyOpsResult struct {
	// Key 是操作后的密钥。
	Key PEMOutput `json:"key"`
	// Algorithm 是算法标识。
	Algorithm string `json:"algorithm"`
	// Format 是输出格式（pkcs8 / spki）。
	Format string `json:"format"`
}

// Validate 做字段级校验。
func (r *KeyOpsRequest) Validate() error {
	if r.Action == "" {
		return tools.MissingField("action")
	}
	if !containsString(KeyOpsActions(), r.Action) {
		return tools.UnsupportedMode(r.Action, KeyOpsActions())
	}
	if err := r.Key.Validate("key"); err != nil {
		return err
	}

	switch r.Action {
	case ActionEncrypt:
		if r.Passphrase == "" {
			return tools.MissingField("passphrase")
		}
	case ActionDecrypt:
		if r.Passphrase == "" {
			return tools.MissingField("passphrase")
		}
	case ActionChangePassphrase:
		if r.Passphrase == "" {
			return tools.MissingField("passphrase")
		}
		if r.NewPassphrase == "" {
			return tools.MissingField("newPassphrase")
		}
	case ActionExtractPublic:
		// 提取公钥**不需要**口令：口令只保护私钥材料。
	}

	if r.PbeCipher == "" {
		r.PbeCipher = pkisvc.PBEAES256CBC
	}
	if !containsString(pkisvc.PBECiphers(), r.PbeCipher) {
		return tools.InvalidField("pbeCipher", pkisvc.PBECiphers(), r.PbeCipher)
	}

	return nil
}

// KeyOps 执行密钥操作。
func KeyOps(_ context.Context, req *KeyOpsRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, _, err := decodeInput(req.Key)
	if err != nil {
		return tools.Result{}, err
	}

	switch req.Action {
	case ActionEncrypt:
		out, err := pkisvc.EncryptPrivateKey(source, req.Passphrase, req.PbeCipher)
		if err != nil {
			return tools.Result{}, err
		}
		return opsResult(out, pkisvc.FormatPKCS8, req.Passphrase)
	case ActionDecrypt:
		out, err := pkisvc.DecryptPrivateKey(source, req.Passphrase)
		if err != nil {
			return tools.Result{}, err
		}
		return opsResult(out, pkisvc.FormatPKCS8, "")
	case ActionChangePassphrase:
		out, err := pkisvc.ChangePassphrase(source, req.Passphrase, req.NewPassphrase)
		if err != nil {
			return tools.Result{}, err
		}
		return opsResult(out, pkisvc.FormatPKCS8, req.NewPassphrase)
	case ActionExtractPublic:
		out, err := pkisvc.ExtractPublicKey(source)
		if err != nil {
			return tools.Result{}, err
		}
		return opsResult(out, pkisvc.FormatSPKI, "")
	default:
		return tools.Result{}, tools.UnsupportedMode(req.Action, KeyOpsActions())
	}
}

// opsResult 构造密钥运算的响应。
//
// passphrase 是**输出**私钥的口令：输出为加密私钥时必须带上，
// 否则无法读回输出密钥的元信息。
func opsResult(pemBytes []byte, format string, passphrase string) (tools.Result, error) {
	info, _, err := pkisvc.Inspect(pemBytes, true, passphrase)
	if err != nil {
		return tools.Result{}, err
	}
	result := &KeyOpsResult{
		Key:       PEMOutput{PEM: encodePEM(pemBytes)},
		Algorithm: info.Algorithm,
		Format:    format,
	}
	return tools.Result{Data: result}, nil
}
