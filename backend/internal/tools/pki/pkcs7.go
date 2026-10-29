package pki

import (
	"context"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// --- p7b-pack ---

// P7BPackRequest 是 PKCS#7 打包工具的请求。
type P7BPackRequest struct {
	// Certs 是待打包的证书列表（首张视为主证书）。
	Certs []PemBundle `json:"certs"`
}

// P7BOutput 是 PKCS#7 产物的输出形态。
type P7BOutput struct {
	// Pem 是容器的 PEM 文本。
	Pem string `json:"pem"`
	// DerBase64 是容器的 DER（base64）。
	DerBase64 string `json:"derBase64"`
	// Length 是 DER 的字节长度。
	Length int `json:"length"`
	// CertCount 是打包的证书数量。
	CertCount int `json:"certCount"`
}

// Validate 做字段级校验。
func (r *P7BPackRequest) Validate() error {
	if len(r.Certs) == 0 {
		return tools.MissingField("certs")
	}
	if len(r.Certs) > pkisvc.MaxPKCS7Certs {
		return tools.InvalidField("certs", []int{pkisvc.MaxPKCS7Certs}, len(r.Certs))
	}
	for i := range r.Certs {
		if err := r.Certs[i].Validate("certs"); err != nil {
			return err
		}
	}
	return nil
}

// P7BPack 打包 PKCS#7 证书袋。
//
// ⚠️ 上游只支持**证书袋**：签名 / 加密容器在 v0.6.0 **不提供**
// （见 docs/upstream/tongsuo-go-issues.md I-11）。
func P7BPack(_ context.Context, req *P7BPackRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	certs, err := loadCertificates(req.Certs)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(certs)

	der, err := pkisvc.BuildPKCS7(certs)
	if err != nil {
		return tools.Result{}, err
	}

	result := &P7BOutput{
		Pem:       string(pkisvc.EncodePKCS7PEM(der)),
		DerBase64: encodeDER(der),
		Length:    len(der),
		CertCount: len(certs),
	}
	return tools.Result{Data: result}, nil
}

// --- p7b-extract ---

// P7BExtractRequest 是 PKCS#7 提取工具的请求。
type P7BExtractRequest struct {
	// Container 是 PKCS#7 容器（PEM / DER / base64 / hex 自动识别）。
	Container PemBundle `json:"container"`
}

// P7BExtractResult 是 PKCS#7 提取工具的响应。
type P7BExtractResult struct {
	// Certificates 是提取出的证书元信息。
	Certificates []CertInfoOutput `json:"certificates"`
	// Count 是证书数量。
	Count int `json:"count"`
	// Pems 是提取出的证书 PEM。
	Pems []string `json:"pems"`
}

// Validate 做字段级校验。
func (r *P7BExtractRequest) Validate() error {
	return r.Container.Validate("container")
}

// P7BExtract 从 PKCS#7 证书袋中提取证书。
func P7BExtract(_ context.Context, req *P7BExtractRequest) (tools.Result, error) {
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

	certs, err := pkisvc.ExtractPKCS7(der)
	if err != nil {
		return tools.Result{}, err
	}
	defer closeAll(certs)

	result := &P7BExtractResult{
		Count:        len(certs),
		Certificates: make([]CertInfoOutput, 0, len(certs)),
		Pems:         make([]string, 0, len(certs)),
	}
	for _, cert := range certs {
		info, err := pkisvc.Describe(cert, time.Time{}, nil)
		if err != nil {
			return tools.Result{}, err
		}
		pemBytes, err := cert.MarshalPEM()
		if err != nil {
			return tools.Result{}, err
		}
		result.Certificates = append(result.Certificates, certInfoOutput(info))
		result.Pems = append(result.Pems, string(pemBytes))
	}

	return tools.Result{Data: result}, nil
}
