package keys

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 密钥材料格式。
const (
	// MaterialFormatJWK 是 JSON Web Key。
	MaterialFormatJWK = pkisvc.FormatJWK
	// MaterialFormatPEM 是 PEM 文本。
	MaterialFormatPEM = pkisvc.FormatPEM
	// MaterialFormatXML 是 .NET RSAKeyValue XML。
	MaterialFormatXML = pkisvc.FormatXML
)

// 密钥类型。
const (
	// KeyTypePrivate 是私钥。
	KeyTypePrivate = pkisvc.KeyTypePrivate
	// KeyTypePublic 是公钥。
	KeyTypePublic = pkisvc.KeyTypePublic
)

// MaterialInput 是待转换的密钥材料。
type MaterialInput struct {
	// Format 是源格式：jwk / pem / xml。
	Format string `json:"format"`
	// Data 是材料文本。
	Data string `json:"data"`
}

// JWKXMLConvertRequest 是 JWK / PEM / XML 转换工具的请求。
type JWKXMLConvertRequest struct {
	// Input 是待转换的密钥材料。
	Input MaterialInput `json:"input"`
	// To 是目标格式：jwk / pem / xml。
	To string `json:"to"`
	// KeyType 是 private 或 public；影响 JWK 与 XML 的导出形态。
	KeyType string `json:"keyType"`
}

// MaterialOutput 是转换后的密钥材料。
type MaterialOutput struct {
	// Format 是目标格式。
	Format string `json:"format"`
	// Data 是材料文本。
	Data string `json:"data"`
}

// JWKXMLConvertResult 是 JWK / PEM / XML 转换工具的响应。
type JWKXMLConvertResult struct {
	// Output 是转换结果。
	Output MaterialOutput `json:"output"`
}

// Validate 做字段级校验。
func (r *JWKXMLConvertRequest) Validate() error {
	switch r.Input.Format {
	case MaterialFormatJWK, MaterialFormatPEM, MaterialFormatXML:
	case "":
		return tools.MissingField("input.format")
	default:
		return tools.UnsupportedFormat(r.Input.Format, pkisvc.ConvertibleFormats())
	}
	if r.Input.Data == "" {
		return tools.MissingField("input.data")
	}
	if r.To == "" {
		return tools.MissingField("to")
	}
	if !containsString(pkisvc.ConvertibleFormats(), r.To) {
		return tools.UnsupportedFormat(r.To, pkisvc.ConvertibleFormats())
	}
	if r.KeyType == "" {
		r.KeyType = KeyTypePrivate
	}
	switch r.KeyType {
	case KeyTypePrivate, KeyTypePublic:
	default:
		return tools.InvalidField("keyType", []string{KeyTypePrivate, KeyTypePublic}, r.KeyType)
	}
	return nil
}

// JWKXMLConvert 在 JWK / PEM / XML 之间转换密钥。
//
// 上游能力边界（转为 UNSUPPORTED_FORMAT，不自研兜底）：
//
//   - JWK 只支持 RSA 与 EC，OKP（Ed25519 / X25519）不支持
//     （见 docs/upstream/tongsuo-go-issues.md I-12）；
//   - XML 只支持 RSA（见 docs/upstream/tongsuo-go-issues.md I-13）。
func JWKXMLConvert(_ context.Context, req *JWKXMLConvertRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	private := req.KeyType == KeyTypePrivate

	// 源 → PEM（中间形态）。
	pemBytes, err := toPEM(req.Input, private)
	if err != nil {
		return tools.Result{}, err
	}

	// PEM → 目标。
	var (
		data   string
		format = req.To
	)
	switch req.To {
	case MaterialFormatJWK:
		out, err := pkisvc.JWKFromPEM(pemBytes, private)
		if err != nil {
			return tools.Result{}, err
		}
		data = string(out)
	case MaterialFormatXML:
		out, err := pkisvc.XMLFromPEM(pemBytes, private)
		if err != nil {
			return tools.Result{}, err
		}
		data = string(out)
	case MaterialFormatPEM:
		data = string(pemBytes)
	default:
		return tools.Result{}, tools.UnsupportedFormat(req.To, pkisvc.ConvertibleFormats())
	}

	result := &JWKXMLConvertResult{
		Output: MaterialOutput{
			Format: format,
			Data:   data,
		},
	}
	return tools.Result{Data: result}, nil
}

// toPEM 把任意源格式归一到 PEM。
func toPEM(in MaterialInput, private bool) ([]byte, error) {
	switch in.Format {
	case MaterialFormatJWK:
		return pkisvc.PEMFromJWK([]byte(in.Data), private)
	case MaterialFormatXML:
		return pkisvc.PEMFromXML([]byte(in.Data), private)
	case MaterialFormatPEM:
		return []byte(in.Data), nil
	default:
		return nil, tools.UnsupportedFormat(in.Format, pkisvc.ConvertibleFormats())
	}
}
