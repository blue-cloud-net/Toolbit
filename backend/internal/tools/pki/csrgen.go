package pki

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CSR 生成类型。
const (
	// CSRTypeServer 是服务端证书请求。
	CSRTypeServer = "server"
	// CSRTypeClient 是客户端证书请求。
	CSRTypeClient = "client"
	// CSRTypeEmail 是邮件证书请求。
	CSRTypeEmail = "email"
	// CSRTypeCodeSigning 是代码签名请求。
	CSRTypeCodeSigning = "code-signing"
)

// CSRTypes 返回受支持的 CSR 生成类型（顺序稳定）。
func CSRTypes() []string {
	return []string{CSRTypeServer, CSRTypeClient, CSRTypeEmail, CSRTypeCodeSigning}
}

// CSRGenRequest 是 CSR 生成工具的请求。
type CSRGenRequest struct {
	// CertType 是生成类型，决定默认的密钥用途与扩展密钥用途。
	CertType string `json:"certType"`
	// Subject 是主体 DN。
	Subject SubjectInput `json:"subject"`
	// Key 是密钥生成参数；与 ExistingPrivateKey 二选一。
	Key KeyGenInput `json:"key"`
	// ExistingPrivateKey 是已有私钥；提供时不生成新密钥，也不返回私钥。
	ExistingPrivateKey *KeyInput `json:"existingPrivateKey"`
	// Sans 是主体备用名条目。
	Sans []string `json:"sans"`
	// ChallengePassword 是挑战口令；响应只报告存在性。
	ChallengePassword string `json:"challengePassword"`
	// KeyUsages 是密钥用途位名；为空时按 CertType 取默认值。
	KeyUsages []string `json:"keyUsages"`
	// ExtKeyUsages 是扩展密钥用途；为空时按 CertType 取默认值。
	ExtKeyUsages []string `json:"extKeyUsages"`
	// BasicConstraints 是基本约束。
	BasicConstraints *BasicConstraintsInput `json:"basicConstraints"`
}

// CSRGenResult 是 CSR 生成工具的响应。
type CSRGenResult struct {
	// CsrPem 是 CSR 的 PEM。
	CsrPem string `json:"csrPem"`
	// PrivateKeyPem 是新生成时的私钥 PEM；使用已有私钥时为空。
	PrivateKeyPem string `json:"privateKeyPem"`
	// KeyType / KeyBits / KeyCurve 是密钥参数。
	KeyType  string `json:"keyType"`
	KeyBits  int    `json:"keyBits"`
	KeyCurve string `json:"keyCurve"`
	// Info 是刚生成的 CSR 的元信息。
	Info CSRInfoOutput `json:"info"`
}

// CSRInfoOutput 是 CSR 元信息的输出形态。
type CSRInfoOutput struct {
	// SubjectText 是主体的单行文本。
	SubjectText string `json:"subjectText"`
	// Sans 是主体备用名条目。
	Sans []string `json:"sans"`
	// PublicKey 是公钥元信息。
	PublicKey PublicKeyOutput `json:"publicKey"`
	// SignatureAlgorithm 是签名算法短名。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// ChallengePasswordPresent 报告是否携带挑战口令。
	ChallengePasswordPresent bool `json:"challengePasswordPresent"`
	// SignatureValid 报告自签名是否校验通过。
	SignatureValid bool `json:"signatureValid"`
	// FingerprintSha256 是 CSR 的 SHA-256 指纹。
	FingerprintSha256 string `json:"fingerprintSha256"`
}

// Validate 做字段级校验。
func (r *CSRGenRequest) Validate() error {
	if r.CertType == "" {
		r.CertType = CSRTypeServer
	}
	if !containsString(CSRTypes(), r.CertType) {
		return tools.InvalidField("certType", CSRTypes(), r.CertType)
	}
	if r.Subject.isEmpty() {
		return tools.MissingField("subject")
	}
	if err := r.Key.Validate(); err != nil {
		return err
	}
	if r.ExistingPrivateKey != nil {
		if err := r.ExistingPrivateKey.Validate("existingPrivateKey"); err != nil {
			return err
		}
	}
	for _, san := range r.Sans {
		if san == "" {
			return tools.InvalidField("sans", []string{"DNS:...", "IP:..."}, san)
		}
	}
	return nil
}

// CSRGen 生成 PKCS#10 证书签名请求。
func CSRGen(_ context.Context, req *CSRGenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	material, err := resolveKeyMaterial(req.Key, req.ExistingPrivateKey)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = material.Key.Close() }()

	keyUsages, extKeyUsages := csrUsages(req)
	out, err := pkisvc.CreateCSR(pkisvc.CSRGenOptions{
		Subject:           req.Subject.entries(),
		SAN:               req.Sans,
		KeyUsage:          keyUsages,
		ExtendedKeyUsage:  extKeyUsages,
		BasicConstraints:  req.BasicConstraints.option(),
		ChallengePassword: req.ChallengePassword,
	}, material.Key)
	if err != nil {
		return tools.Result{}, err
	}

	result := &CSRGenResult{
		CsrPem:        string(out.PEM),
		PrivateKeyPem: material.PrivatePem,
		KeyType:       material.Info.Algorithm,
		KeyBits:       material.Info.Bits,
		KeyCurve:      material.Info.Curve,
		Info:          csrInfoOutput(out.Info),
	}
	return tools.Result{Data: result}, nil
}

// csrUsages 返回本次请求生效的密钥用途。
//
// 显式传入的用途优先；未传时按 CertType 取一组安全检查的默认值。
func csrUsages(req *CSRGenRequest) (keyUsage, extKeyUsage []string) {
	keyUsage = req.KeyUsages
	extKeyUsage = req.ExtKeyUsages
	if len(keyUsage) == 0 {
		switch req.CertType {
		case CSRTypeServer:
			keyUsage = []string{"critical", "digitalSignature", "keyEncipherment"}
		default:
			keyUsage = []string{"critical", "digitalSignature"}
		}
	}
	if len(extKeyUsage) == 0 {
		switch req.CertType {
		case CSRTypeServer:
			extKeyUsage = []string{"serverAuth"}
		case CSRTypeClient:
			extKeyUsage = []string{"clientAuth"}
		case CSRTypeEmail:
			extKeyUsage = []string{"emailProtection"}
		case CSRTypeCodeSigning:
			extKeyUsage = []string{"codeSigning"}
		}
	}
	return keyUsage, extKeyUsage
}

// csrInfoOutput 把 CSR 元信息转成输出形态。
func csrInfoOutput(info *pkisvc.CSRInfo) CSRInfoOutput {
	if info == nil {
		return CSRInfoOutput{}
	}
	return CSRInfoOutput{
		SubjectText:              info.SubjectText,
		Sans:                     info.SAN,
		PublicKey:                publicKeyOutput(info.PublicKey),
		SignatureAlgorithm:       info.SignatureAlgorithm,
		ChallengePasswordPresent: info.ChallengePasswordPresent,
		SignatureValid:           info.SignatureValid,
		FingerprintSha256:        info.FingerprintSHA256,
	}
}

// KeyMaterial 是密钥材料的解析结果。
type KeyMaterial struct {
	// Key 是私钥句柄。
	Key *pkisvc.PrivateKey
	// PrivatePem 是新生成私钥时的 PEM 文本；使用已有私钥时为空。
	PrivatePem string
	// Info 是密钥元信息。
	Info pkisvc.PublicKeyInfo
}

// resolveKeyMaterial 按「已有私钥优先，否则新建」解析密钥材料。
//
// 返回值中的 Key 由调用方负责 Close。
func resolveKeyMaterial(gen KeyGenInput, existing *KeyInput) (*KeyMaterial, error) {
	if existing != nil && existing.Data != "" {
		source, isPEM, err := existing.Resolve()
		if err != nil {
			return nil, err
		}
		if !isPEM {
			return nil, tools.UnsupportedFormat("der", []string{"pem"})
		}
		key, err := pkisvc.LoadPrivateKey(source, existing.Passphrase)
		if err != nil {
			return nil, err
		}
		info, err := pkisvc.DescribePrivateKey(key)
		if err != nil {
			_ = key.Close()
			return nil, err
		}
		return &KeyMaterial{Key: key, Info: publicKeyInfoOf(info)}, nil
	}

	pair, err := pkisvc.Generate(gen.algorithm(), gen.Bits, gen.Curve)
	if err != nil {
		return nil, err
	}
	// 公钥句柄独立持有，这里用不到，立即释放；私钥转移给调用方。
	_ = pair.Public.Close()

	pemBytes, err := pair.Private.MarshalPKCS8PEM()
	if err != nil {
		_ = pair.Private.Close()
		return nil, err
	}
	info, err := pkisvc.DescribePrivateKey(pair.Private)
	if err != nil {
		_ = pair.Private.Close()
		return nil, err
	}
	return &KeyMaterial{Key: pair.Private, PrivatePem: string(pemBytes), Info: publicKeyInfoOf(info)}, nil
}

// publicKeyInfoOf 从密钥元信息里取公钥参数。
func publicKeyInfoOf(info *pkisvc.KeyInfo) pkisvc.PublicKeyInfo {
	if info == nil {
		return pkisvc.PublicKeyInfo{}
	}
	return pkisvc.PublicKeyInfo{
		Algorithm: info.Algorithm,
		Bits:      info.Bits,
		Curve:     info.Curve,
	}
}
