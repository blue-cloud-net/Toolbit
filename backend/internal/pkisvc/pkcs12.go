package pkisvc

import (
	"strings"

	"github.com/blue-cloud-net/tongsuo-go/pkcs/pkcs12"
	tx509 "github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// PKCS12PackOptions 是打包 PKCS#12 的参数。
type PKCS12PackOptions struct {
	// Password 是容器口令；可为空（不加密）。
	Password string
	// FriendlyName 是友好名称；可为空。
	FriendlyName string
	// CACerts 是随包附带的 CA 链（不含主证书）。
	CACerts []*Certificate
}

// PackPKCS12 把证书、私钥与 CA 链打包为 PKCS#12（DER）。
//
// 私钥与证书必须配对：不配对的容器对使用者毫无意义。
func PackPKCS12(cert *Certificate, key *PrivateKey, opts PKCS12PackOptions) ([]byte, error) {
	if cert == nil || cert.cert == nil {
		return nil, toolerrInvalidCert("certificate unavailable")
	}
	if key == nil || key.key == nil {
		return nil, toolerrInvalidKey("private key unavailable")
	}
	if !signerMatchesCertificate(key, cert) {
		return nil, toolerr.Newf(toolerr.CodeKeyMismatch, map[string]any{
			"field": "key",
		})
	}
	if len(opts.CACerts) > MaxChainLength {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "caCerts",
			"limit":  MaxChainLength,
			"actual": len(opts.CACerts),
		})
	}

	ca := make([]*tx509.Certificate, 0, len(opts.CACerts))
	for _, item := range opts.CACerts {
		if item == nil || item.cert == nil {
			continue
		}
		ca = append(ca, item.cert)
	}

	der, err := pkcs12.Pack(cert.cert, key.key, ca, opts.Password, opts.FriendlyName)
	if err != nil {
		return nil, mapPKCS12Error(err)
	}
	return der, nil
}

// PKCS12Bundle 是解析出的 PKCS#12 内容。
//
// 句柄（PrivateKey / Certificate / CACerts）由调用方负责 Close。
type PKCS12Bundle struct {
	// PrivateKey 是容器内的私钥；容器不含私钥时为 nil。
	PrivateKey *PrivateKey
	// Certificate 是主证书；容器不含证书时为 nil。
	Certificate *Certificate
	// CACerts 是随包的 CA 链。
	CACerts []*Certificate
}

// Close 释放容器内的全部句柄。
//
// **幂等**：重复调用只返回首次的结果。
func (b *PKCS12Bundle) Close() error {
	if b == nil {
		return nil
	}
	var firstErr error
	for _, cert := range b.CACerts {
		if err := cert.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	b.CACerts = nil
	if b.Certificate != nil {
		if err := b.Certificate.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		b.Certificate = nil
	}
	if b.PrivateKey != nil {
		if err := b.PrivateKey.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		b.PrivateKey = nil
	}
	return firstErr
}

// ParsePKCS12 解析 PKCS#12 容器。
//
// 口令错误返回 PASSPHRASE_INCORRECT；结构损坏返回 INVALID_PKCS12。
func ParsePKCS12(der []byte, password string) (*PKCS12Bundle, error) {
	if len(der) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "container.data"})
	}

	bundle, err := pkcs12.Parse(der, password)
	if err != nil {
		return nil, mapPKCS12Error(err)
	}

	out := &PKCS12Bundle{}
	if bundle.PrivateKey != nil {
		out.PrivateKey = &PrivateKey{key: bundle.PrivateKey}
	}
	if bundle.Certificate != nil {
		out.Certificate = &Certificate{cert: bundle.Certificate}
	}
	for _, item := range bundle.CACerts {
		if item == nil {
			continue
		}
		out.CACerts = append(out.CACerts, &Certificate{cert: item})
	}
	return out, nil
}

// ChangePKCS12Password 修改 PKCS#12 容器口令（输入输出均为 DER）。
func ChangePKCS12Password(der []byte, oldPassword, newPassword string) ([]byte, error) {
	if len(der) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "container.data"})
	}
	if oldPassword == "" {
		return nil, toolerr.Newf(toolerr.CodePassphraseRequired, map[string]any{
			"field": "oldPassphrase",
		})
	}
	if newPassword == "" {
		return nil, toolerr.Newf(toolerr.CodePassphraseRequired, map[string]any{
			"field": "newPassphrase",
		})
	}

	out, err := pkcs12.ChangePassword(der, oldPassword, newPassword)
	if err != nil {
		return nil, mapPKCS12Error(err)
	}
	return out, nil
}

// mapPKCS12Error 把 PKCS#12 层的错误映射为 toolerr 错误码。
//
// 上游对「口令错误」与「容器损坏」返回的都是同一类 OpError，
// 只能按底层 OpenSSL 的错误描述区分。判据是**枚举化的片段**，
// 上游英文原文**不会**出现在响应里（见 AGENTS.md §4.5）。
func mapPKCS12Error(err error) *toolerr.Error {
	if err == nil {
		return nil
	}
	if isPassphraseError(err) {
		return toolerr.Newf(toolerr.CodePassphraseIncorrect, map[string]any{
			"field": "passphrase",
		})
	}
	return toolerrInvalidPKCS12("cannot parse container")
}

// passphraseErrorOps 列出「口令错误」在底层操作名里的特征片段。
//
// 判据用**操作名**而不是 OpenSSL 的错误描述：错误描述来自线程错误栈
// （`ERR_error_string_n`），而该栈在成功调用后**不会被清空**，
// 同进程内前一次失败遗留的条目会覆盖本次的真实原因
// （见 docs/upstream/tongsuo-go-issues.md I-22）。
// 操作名由上游 `OpError.Op` 固定给出，不受错误栈影响。
var passphraseErrorOps = []string{
	"pkcs12_parse",
	"pkcs12_newpass",
}

// isPassphraseError 判断错误是否由「口令不正确」引起。
//
// 上游在结构损坏时于 `LoadPKCS12DER` 阶段就失败，
// 因此能走到 `PKCS12_parse` / `PKCS12_newpass` 说明容器结构本身可用，
// 此时失败的最可能原因就是口令不匹配。
func isPassphraseError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	for _, op := range passphraseErrorOps {
		if strings.Contains(text, op) {
			return true
		}
	}
	return false
}
