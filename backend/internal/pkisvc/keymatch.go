package pkisvc

import (
	"bytes"
	"encoding/pem"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 配对检测结论。
const (
	// MatchMatched 表示两段材料是同一把密钥的私钥与公钥。
	MatchMatched = "matched"
	// MatchIdentical 表示两段材料是**完全相同**的私钥。
	MatchIdentical = "identical"
	// MatchPublicMismatch 表示算法相同但公钥不同。
	MatchPublicMismatch = "public_key_mismatch"
	// MatchNotPair 表示算法不同，不可能是密钥对。
	MatchNotPair = "not_a_key_pair"
)

// MatchReasons 返回全部配对检测结论（顺序稳定）。
func MatchReasons() []string {
	return []string{MatchMatched, MatchIdentical, MatchPublicMismatch, MatchNotPair}
}

// MatchResult 是配对检测的结果。
type MatchResult struct {
	// Match 报告两段材料是否配对。
	Match bool
	// Reason 是结论枚举，取值见 MatchReasons。
	Reason string
	// AFingerprint 是 A 的公钥指纹（小写十六进制）。
	AFingerprint string
	// BFingerprint 是 B 的公钥指纹（小写十六进制）。
	BFingerprint string
}

// materialKind 是密钥材料的种类。
type materialKind int

const (
	// materialPrivate 是私钥。
	materialPrivate materialKind = iota
	// materialPublic 是公钥。
	materialPublic
	// materialCertificate 是证书。
	materialCertificate
)

// material 是一段已解析的密钥材料。
type material struct {
	// Kind 是材料种类。
	Kind materialKind
	// Algorithm 是算法标识。
	Algorithm string
	// SPKI 是公钥的 SPKI PEM；配对判定只看它。
	SPKI []byte
	// PrivatePEM 是私钥的 PKCS#8 PEM；仅私钥有值。
	PrivatePEM []byte
}

// Match 检测两段密钥材料是否配对。
//
// 每一段都可以是私钥、公钥或证书；判定依据是**公钥的 SPKI**，
// 因此「私钥 ↔ 证书」这种组合也能正确判定。
//
// 输入无法解析时返回错误（INVALID_KEY / INVALID_CERT）；
// 能解析但不配对时返回 Match=false 与对应结论，**不**是错误。
func Match(a, b []byte) (*MatchResult, error) {
	matA, err := parseMaterial(a)
	if err != nil {
		return nil, err
	}
	matB, err := parseMaterial(b)
	if err != nil {
		return nil, err
	}

	fpA, err := publicFingerprintFromPEM(matA.SPKI)
	if err != nil {
		return nil, err
	}
	fpB, err := publicFingerprintFromPEM(matB.SPKI)
	if err != nil {
		return nil, err
	}

	result := &MatchResult{
		AFingerprint: fpA,
		BFingerprint: fpB,
	}

	switch {
	case bytes.Equal(matA.SPKI, matB.SPKI):
		result.Match = true
		if matA.Kind == materialPrivate && matB.Kind == materialPrivate &&
			bytes.Equal(matA.PrivatePEM, matB.PrivatePEM) {
			result.Reason = MatchIdentical
		} else {
			result.Reason = MatchMatched
		}
	case matA.Algorithm != matB.Algorithm:
		result.Reason = MatchNotPair
	default:
		result.Reason = MatchPublicMismatch
	}

	return result, nil
}

// parseMaterial 解析一段密钥材料。
//
// 判定顺序：私钥 → 公钥 → 证书。
func parseMaterial(pemBytes []byte) (*material, error) {
	if len(pemBytes) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "key.data"})
	}

	if priv, err := LoadPrivateKey(pemBytes, ""); err == nil {
		spki, err := priv.MarshalPublicKeyPEM()
		if err != nil {
			_ = priv.Close()
			return nil, err
		}
		algorithm := priv.Algorithm()
		pkcs8, err := priv.MarshalPKCS8PEM()
		_ = priv.Close()
		if err != nil {
			return nil, err
		}
		return &material{
			Kind:       materialPrivate,
			Algorithm:  algorithm,
			SPKI:       spki,
			PrivatePEM: pkcs8,
		}, nil
	}

	if pub, err := LoadPublicKey(pemBytes); err == nil {
		spki, err := pub.MarshalPEM()
		algorithm := pub.Algorithm()
		_ = pub.Close()
		if err != nil {
			return nil, err
		}
		return &material{
			Kind:      materialPublic,
			Algorithm: algorithm,
			SPKI:      spki,
		}, nil
	}

	cert, err := x509.LoadCertificatePEM(pemBytes)
	if err != nil {
		// 输入若**看起来**是证书（PEM 块标签为 CERTIFICATE），
		// 报 INVALID_CERT；否则就是「不是任何一种密钥材料」，报 INVALID_KEY。
		if isCertificatePEM(pemBytes) {
			return nil, toolerr.Newf(toolerr.CodeInvalidCert, map[string]any{
				"field": "key.data",
			})
		}
		return nil, toolerrInvalidKey("not a key, certificate or public key")
	}
	defer func() { _ = cert.Close() }()

	pub, err := cert.PublicKey()
	if err != nil {
		return nil, MapError(err)
	}
	defer func() { _ = asym.Close(pub) }()

	spki, err := pub.MarshalPublicKeyPEM()
	if err != nil {
		return nil, MapError(err)
	}

	return &material{
		Kind:      materialCertificate,
		Algorithm: algorithmFromUpstream(pub.Algorithm().String()),
		SPKI:      spki,
	}, nil
}

// isCertificatePEM 判断输入的 PEM 块标签是否是 CERTIFICATE。
func isCertificatePEM(pemBytes []byte) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}
	return block.Type == "CERTIFICATE"
}
