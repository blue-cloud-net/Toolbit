package pkisvc

import (
	"encoding/json"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/jwk"
	rsaxml "github.com/blue-cloud-net/tongsuo-go/xml/rsa"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// JWK 与 XML 的格式标识。
const (
	// FormatJWK 是 JSON Web Key。
	FormatJWK = "jwk"
	// FormatXML 是 .NET RSAKeyValue XML。
	FormatXML = "xml"
)

// ConvertibleFormats 返回 JWK / XML / PEM 三种可互转格式（顺序稳定）。
func ConvertibleFormats() []string {
	return []string{FormatJWK, FormatXML, FormatPEM}
}

// JWKAlgorithms 返回 JWK 支持的算法（顺序稳定）。
//
// 上游 `jwk` 包**只支持 RSA 与 EC**：OKP（Ed25519 / X25519）与
// 对称密钥（oct）尚未实现，见 docs/upstream/tongsuo-go-issues.md I-12。
func JWKAlgorithms() []string {
	return []string{AlgRSA, AlgEC}
}

// JWKFromPEM 把 PEM 密钥转成 JWK JSON。
//
// private 为真时导出私钥 JWK（必须含 `d` 成员），
// 否则只导出公钥分量。
func JWKFromPEM(pemBytes []byte, private bool) ([]byte, error) {
	if private {
		priv, err := LoadPrivateKey(pemBytes, "")
		if err != nil {
			return nil, err
		}
		defer func() { _ = priv.Close() }()

		key, err := jwk.MarshalKey(priv.key)
		if err != nil {
			return nil, toolerr.Wrap(toolerr.CodeInvalidJWK, err)
		}
		return marshalJWK(key)
	}

	pub, err := LoadPublicKey(pemBytes)
	if err != nil {
		return nil, err
	}
	defer func() { _ = pub.Close() }()

	key, err := jwk.MarshalKey(pub.key)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidJWK, err)
	}
	return marshalJWK(key)
}

// PEMFromJWK 把 JWK JSON 转成 PEM。
//
// private 为真时输出 PKCS#8 私钥 PEM，否则输出 SPKI 公钥 PEM。
func PEMFromJWK(jsonBytes []byte, private bool) ([]byte, error) {
	key, err := jwk.Parse(jsonBytes)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidJWK, err)
	}

	if private {
		if !key.IsPrivate() {
			return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"field":    "keyType",
				"expected": []string{KeyTypePrivate},
			})
		}
		out, err := key.ToPEM()
		if err != nil {
			return nil, toolerr.Wrap(toolerr.CodeInvalidJWK, err)
		}
		return out, nil
	}

	out, err := key.ToPublicPEM()
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidJWK, err)
	}
	return out, nil
}

// XMLFromPEM 把 PEM 密钥转成 .NET RSAKeyValue XML。
//
// **只支持 RSA**：上游 `xml/rsa` 包只处理 RSA 密钥，
// 见 docs/upstream/tongsuo-go-issues.md I-13。
func XMLFromPEM(pemBytes []byte, private bool) ([]byte, error) {
	if private {
		priv, err := LoadPrivateKey(pemBytes, "")
		if err != nil {
			return nil, err
		}
		defer func() { _ = priv.Close() }()

		if priv.Algorithm() != AlgRSA {
			return nil, unsupportedFormat(priv.Algorithm(), JWKAlgorithms())
		}
		out, err := rsaxml.MarshalPrivate(priv.key)
		if err != nil {
			return nil, toolerr.Wrap(toolerr.CodeInvalidXML, err)
		}
		return out, nil
	}

	pub, err := LoadPublicKey(pemBytes)
	if err != nil {
		return nil, err
	}
	defer func() { _ = pub.Close() }()

	if pub.Algorithm() != AlgRSA {
		return nil, unsupportedFormat(pub.Algorithm(), JWKAlgorithms())
	}
	out, err := rsaxml.MarshalPublic(pub.key)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidXML, err)
	}
	return out, nil
}

// PEMFromXML 把 .NET RSAKeyValue XML 转成 PEM。
//
// private 为真时输出 PKCS#8 私钥 PEM，否则输出 SPKI 公钥 PEM。
func PEMFromXML(xmlBytes []byte, private bool) ([]byte, error) {
	if private {
		priv, err := rsaxml.UnmarshalPrivate(xmlBytes)
		if err != nil {
			return nil, toolerr.Wrap(toolerr.CodeInvalidXML, err)
		}
		defer func() { _ = asym.Close(priv) }()

		out, err := priv.MarshalPrivateKeyPEM()
		if err != nil {
			return nil, MapError(err)
		}
		return out, nil
	}

	pub, err := rsaxml.UnmarshalPublic(xmlBytes)
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidXML, err)
	}
	defer func() { _ = asym.Close(pub) }()

	out, err := pub.MarshalPublicKeyPEM()
	if err != nil {
		return nil, MapError(err)
	}
	return out, nil
}

// marshalJWK 把 JWK 结构序列化为 JSON。
func marshalJWK(key *jwk.Key) ([]byte, error) {
	out, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInvalidJWK, err)
	}
	return out, nil
}
