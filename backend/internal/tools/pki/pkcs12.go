package pki

import (
	"context"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// --- pfx-pack ---

// PFXPackRequest 是 PKCS#12 打包工具的请求。
type PFXPackRequest struct {
	// Cert 是主证书。
	Cert PemBundle `json:"cert"`
	// Key 是与证书配对的私钥。
	Key KeyInput `json:"key"`
	// CaCerts 是随包的 CA 链。
	CaCerts []PemBundle `json:"caCerts"`
	// Passphrase 是容器口令；为空表示不加密。
	Passphrase string `json:"passphrase"`
	// FriendlyName 是友好名称。
	FriendlyName string `json:"friendlyName"`
}

// PFXOutput 是 PKCS#12 二进制产物的输出形态。
type PFXOutput struct {
	// Base64 是容器的 DER（base64）。
	Base64 string `json:"base64"`
	// Length 是字节长度。
	Length int `json:"length"`
	// Encoding 固定为 base64：JSON 装不下二进制。
	Encoding string `json:"encoding"`
}

// Validate 做字段级校验。
func (r *PFXPackRequest) Validate() error {
	if err := r.Cert.Validate("cert"); err != nil {
		return err
	}
	if err := r.Key.Validate("key"); err != nil {
		return err
	}
	for i := range r.CaCerts {
		if err := r.CaCerts[i].Validate("caCerts"); err != nil {
			return err
		}
	}
	return nil
}

// PFXPack 打包 PKCS#12。
func PFXPack(_ context.Context, req *PFXPackRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	cert, err := loadPemCertificate(req.Cert)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = cert.Close() }()

	key, err := loadPemPrivateKey(req.Key)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = key.Close() }()

	ca, err := loadCertificates(req.CaCerts)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(ca)

	der, err := pkisvc.PackPKCS12(cert, key, pkisvc.PKCS12PackOptions{
		Password:     req.Passphrase,
		FriendlyName: req.FriendlyName,
		CACerts:      ca,
	})
	if err != nil {
		return tools.Result{}, err
	}

	result := &PFXOutput{
		Base64:   encodeDER(der),
		Length:   len(der),
		Encoding: "base64",
	}
	return tools.Result{Data: result, Warnings: pfxWarnings(req.Passphrase)}, nil
}

// --- pfx-parse ---

// PFXParseRequest 是 PKCS#12 解析工具的请求。
type PFXParseRequest struct {
	// Container 是 PKCS#12 容器（DER / base64 / hex 自动识别）。
	Container PemBundle `json:"container"`
	// Passphrase 是容器口令。
	Passphrase string `json:"passphrase"`
}

// PFXParseResult 是 PKCS#12 解析工具的响应。
type PFXParseResult struct {
	// HasPrivateKey 报告容器是否含私钥。
	HasPrivateKey bool `json:"hasPrivateKey"`
	// Certificate 是主证书的元信息；容器不含证书时为 null。
	Certificate *CertInfoOutput `json:"certificate"`
	// PrivateKeyPem 是容器内私钥的 PKCS#8 PEM。
	PrivateKeyPem string `json:"privateKeyPem"`
	// CaCertificates 是随包 CA 链的元信息。
	CaCertificates []CertInfoOutput `json:"caCertificates"`
	// CaCount 是 CA 证书数量。
	CaCount int `json:"caCount"`
}

// Validate 做字段级校验。
func (r *PFXParseRequest) Validate() error {
	return r.Container.Validate("container")
}

// PFXParse 解析 PKCS#12 容器。
func PFXParse(_ context.Context, req *PFXParseRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Container.Resolve()
	if err != nil {
		return tools.Result{}, err
	}
	der, err := pkisvc.NormalizeToDER(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}

	bundle, err := pkisvc.ParsePKCS12(der, req.Passphrase)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = bundle.Close() }()

	result := &PFXParseResult{
		HasPrivateKey: bundle.PrivateKey != nil,
		CaCount:       len(bundle.CACerts),
	}
	if bundle.PrivateKey != nil {
		pemBytes, err := bundle.PrivateKey.MarshalPKCS8PEM()
		if err != nil {
			return tools.Result{}, err
		}
		result.PrivateKeyPem = string(pemBytes)
	}
	if bundle.Certificate != nil {
		info, err := pkisvc.Describe(bundle.Certificate, time.Time{}, nil)
		if err != nil {
			return tools.Result{}, err
		}
		out := certInfoOutput(info)
		result.Certificate = &out
	}
	for _, cert := range bundle.CACerts {
		info, err := pkisvc.Describe(cert, time.Time{}, nil)
		if err != nil {
			return tools.Result{}, err
		}
		result.CaCertificates = append(result.CaCertificates, certInfoOutput(info))
	}

	return tools.Result{Data: result}, nil
}

// --- pfx-change-password ---

// PFXChangePasswordRequest 是 PKCS#12 改密工具的请求。
type PFXChangePasswordRequest struct {
	// Container 是 PKCS#12 容器。
	Container PemBundle `json:"container"`
	// OldPassphrase 是当前口令。
	OldPassphrase string `json:"oldPassphrase"`
	// NewPassphrase 是新口令。
	NewPassphrase string `json:"newPassphrase"`
}

// Validate 做字段级校验。
func (r *PFXChangePasswordRequest) Validate() error {
	if err := r.Container.Validate("container"); err != nil {
		return err
	}
	if r.OldPassphrase == "" {
		return tools.MissingField("oldPassphrase")
	}
	if r.NewPassphrase == "" {
		return tools.MissingField("newPassphrase")
	}
	return nil
}

// PFXChangePassword 修改 PKCS#12 口令。
func PFXChangePassword(_ context.Context, req *PFXChangePasswordRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Container.Resolve()
	if err != nil {
		return tools.Result{}, err
	}
	der, err := pkisvc.NormalizeToDER(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}

	updated, err := pkisvc.ChangePKCS12Password(der, req.OldPassphrase, req.NewPassphrase)
	if err != nil {
		return tools.Result{}, err
	}

	result := &PFXOutput{
		Base64:   encodeDER(updated),
		Length:   len(updated),
		Encoding: "base64",
	}
	return tools.Result{Data: result}, nil
}

// pfxWarnings 汇总 PKCS#12 相关的警告。
//
// 无口令的容器不再受口令保护，值得提醒一句。
func pfxWarnings(passphrase string) []toolerr.Warning {
	if passphrase != "" {
		return nil
	}
	return []toolerr.Warning{{
		Code:   toolerr.WarnLossyFormat,
		Params: map[string]any{"reason": "container has no passphrase"},
	}}
}
