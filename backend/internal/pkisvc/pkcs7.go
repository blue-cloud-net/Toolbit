package pkisvc

import (
	"github.com/blue-cloud-net/tongsuo-go/pkcs/pkcs7"
	tx509 "github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// MaxPKCS7Certs 是 PKCS#7 证书袋的证书数上限。
const MaxPKCS7Certs = 64

// BuildPKCS7 把多张证书打包为 PKCS#7 证书袋（DER）。
//
// ⚠️ 上游 `pkcs/pkcs7` **只支持证书袋**：没有签名 / 加密容器
// （见 docs/upstream/tongsuo-go-issues.md I-11）。
func BuildPKCS7(certs []*Certificate) ([]byte, error) {
	if len(certs) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "certs"})
	}
	if len(certs) > MaxPKCS7Certs {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "certs",
			"limit":  MaxPKCS7Certs,
			"actual": len(certs),
		})
	}

	upstreamCerts := make([]*tx509.Certificate, 0, len(certs))
	for _, cert := range certs {
		if cert == nil || cert.cert == nil {
			return nil, toolerrInvalidCert("certificate unavailable")
		}
		upstreamCerts = append(upstreamCerts, cert.cert)
	}

	der, err := pkcs7.Build(upstreamCerts)
	if err != nil {
		return nil, toolerrInvalidPKCS7("cannot build container")
	}
	return der, nil
}

// ExtractPKCS7 从 PKCS#7 证书袋中提取全部证书。
//
// 返回的证书句柄由调用方负责 Close。
// 签名 / 加密容器会返回 UNSUPPORTED_FORMAT（上游只支持证书袋）。
func ExtractPKCS7(der []byte) ([]*Certificate, error) {
	if len(der) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "container.data"})
	}

	certs, err := pkcs7.Extract(der)
	if err != nil {
		return nil, toolerrInvalidPKCS7("cannot parse container")
	}
	if len(certs) == 0 {
		return nil, toolerrInvalidPKCS7("no certificate in container")
	}
	if len(certs) > MaxPKCS7Certs {
		closeCertificates(wrapCertificates(certs))
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "container.data",
			"limit":  MaxPKCS7Certs,
			"actual": len(certs),
		})
	}

	out := wrapCertificates(certs)
	return out, nil
}

// wrapCertificates 把上游证书句柄包装为本层句柄。
func wrapCertificates(certs []*tx509.Certificate) []*Certificate {
	out := make([]*Certificate, 0, len(certs))
	for _, cert := range certs {
		if cert == nil {
			continue
		}
		out = append(out, &Certificate{cert: cert})
	}
	return out
}

// EncodePKCS7PEM 把 PKCS#7 的 DER 编成 PEM 文本。
func EncodePKCS7PEM(der []byte) []byte {
	return pkcs7.MarshalPEM(der)
}
