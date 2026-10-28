package pkisvc

import (
	"encoding/pem"

	"github.com/blue-cloud-net/tongsuo-go/asym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// ConvertOptions 是密钥格式转换的参数。
type ConvertOptions struct {
	// Source 是源字节：PEM 文本或 DER 二进制。
	Source []byte
	// SourceIsPEM 报告 Source 是否是 PEM 文本。
	SourceIsPEM bool
	// Passphrase 是私钥口令；源为加密私钥时**必填**。
	Passphrase string
	// From 是源格式；FormatAuto 表示自动识别。
	From string
	// To 是目标格式；取值域见 KeyFormats。
	To string
}

// ConvertResult 是格式转换的结果。
type ConvertResult struct {
	// PEM 是转换后的 PEM 文本；目标为 der 时为空。
	PEM []byte
	// DER 是转换后的 DER 字节；目标非 der 时为空。
	DER []byte
	// Algorithm 是算法标识。
	Algorithm string
	// KeyType 是密钥类型（private / public）。
	KeyType string
	// Detected 是识别到的源格式。
	Detected string
	// Encrypted 报告源是否为加密私钥。
	Encrypted bool
	// Bits 是 RSA 位数；非 RSA 为 0。
	Bits int
	// Curve 是曲线名；非椭圆曲线为空。
	Curve string
}

// Convert 把密钥从一种格式转换为另一种格式。
//
// 源格式由 From 约束（FormatAuto 时自动识别）；
// 目标格式由 To 指定，取值域见 KeyFormats。
//
// 转换规则：
//
//   - 私钥 → spki：**报错**（spki 是公钥格式）；
//   - 公钥 → pkcs8 / pkcs1：**报错**（公钥没有这两种形态）；
//   - 私钥 → pkcs1：仅 RSA，且**源不得是加密私钥**；
//   - 源是加密私钥时，转换结果**保持加密**（AES-256-CBC + 原口令）。
func Convert(opts ConvertOptions) (*ConvertResult, error) {
	if len(opts.Source) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "key.data"})
	}
	if !containsString(KeyFormats(), opts.To) {
		return nil, unsupportedFormat(opts.To, KeyFormats())
	}

	source, err := detectSource(opts)
	if err != nil {
		return nil, err
	}
	if err := checkSourceFormat(source, opts.From); err != nil {
		return nil, err
	}

	switch source.Kind {
	case sourcePrivate:
		return convertPrivate(source, opts)
	case sourcePublic:
		return convertPublic(source, opts)
	default:
		return nil, toolerr.Newf(toolerr.CodeInvalidKey, map[string]any{
			"reason": "unsupported key material",
		})
	}
}

// sourceKind 是源密钥的种类。
type sourceKind int

const (
	// sourcePrivate 是私钥。
	sourcePrivate sourceKind = iota
	// sourcePublic 是公钥。
	sourcePublic
)

// sourceKey 是识别出的源密钥。
type sourceKey struct {
	// Kind 是私钥还是公钥。
	Kind sourceKind
	// Container 是外层容器：pem 或 der。
	Container string
	// Encoding 是密钥编码：pkcs1 / pkcs8 / spki。
	Encoding string
	// PEM 是 PEM 文本；DER 输入会被包装成 PEM。
	PEM []byte
	// Encrypted 报告源是否为加密私钥。
	Encrypted bool
}

// checkSourceFormat 校验识别结果是否满足 From 的约束。
func checkSourceFormat(source *sourceKey, from string) error {
	switch from {
	case "", FormatAuto:
		return nil
	case FormatPEM:
		if source.Container != FormatPEM {
			return unsupportedFormat(from, []string{FormatDER})
		}
		return nil
	case FormatDER:
		if source.Container != FormatDER {
			return unsupportedFormat(from, []string{FormatPEM})
		}
		return nil
	default:
		if source.Encoding != from {
			return unsupportedFormat(from, []string{source.Encoding})
		}
		return nil
	}
}

// detectSource 识别源密钥的种类与格式。
func detectSource(opts ConvertOptions) (*sourceKey, error) {
	if opts.SourceIsPEM {
		return detectPEMSource(opts.Source)
	}
	return detectDERSource(opts.Source)
}

// detectPEMSource 识别 PEM 源。
func detectPEMSource(pemBytes []byte) (*sourceKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field": "key.data",
		})
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return &sourceKey{
			Kind:      sourcePrivate,
			Container: FormatPEM,
			Encoding:  FormatPKCS1,
			PEM:       pemBytes,
		}, nil
	case "PRIVATE KEY":
		return &sourceKey{
			Kind:      sourcePrivate,
			Container: FormatPEM,
			Encoding:  FormatPKCS8,
			PEM:       pemBytes,
		}, nil
	case "ENCRYPTED PRIVATE KEY":
		return &sourceKey{
			Kind:      sourcePrivate,
			Container: FormatPEM,
			Encoding:  FormatPKCS8,
			PEM:       pemBytes,
			Encrypted: true,
		}, nil
	case "PUBLIC KEY":
		return &sourceKey{
			Kind:      sourcePublic,
			Container: FormatPEM,
			Encoding:  FormatSPKI,
			PEM:       pemBytes,
		}, nil
	default:
		// "RSA PUBLIC KEY"（PKCS#1 公钥）没有底层加载入口：
		// 铜锁只提供 SPKI 公钥的读取。此处**显式拒绝**而不是猜测转换，
		// 避免产出无法被上层解释的密钥（见 docs/upstream/tongsuo-go-issues.md I-14）。
		return nil, unsupportedFormat(block.Type, []string{
			"RSA PRIVATE KEY", "PRIVATE KEY", "ENCRYPTED PRIVATE KEY", "PUBLIC KEY",
		})
	}
}

// derCandidate 是一个 DER 试加载候选。
type derCandidate struct {
	// Label 是 PEM 块标签。
	Label string
	// Encoding 是对应的密钥编码。
	Encoding string
	// Kind 是私钥还是公钥。
	Kind sourceKind
}

// derCandidates 是 DER 试加载的候选顺序。
//
// 顺序有意义：私钥优先（私钥的 DER 不会被误判为公钥），
// 同类内部按「现代编码在前」排列。
var derCandidates = []derCandidate{
	{Label: "PRIVATE KEY", Encoding: FormatPKCS8, Kind: sourcePrivate},
	{Label: "RSA PRIVATE KEY", Encoding: FormatPKCS1, Kind: sourcePrivate},
	{Label: "EC PRIVATE KEY", Encoding: FormatPKCS8, Kind: sourcePrivate},
	{Label: "PUBLIC KEY", Encoding: FormatSPKI, Kind: sourcePublic},
}

// detectDERSource 识别 DER 源。
//
// DER 本身不携带格式标签，只能靠**试加载**判定：按候选顺序逐个包装成 PEM
// 试加载，第一个成功的即为识别结果。
func detectDERSource(der []byte) (*sourceKey, error) {
	for _, candidate := range derCandidates {
		pemBytes := pem.EncodeToMemory(&pem.Block{Type: candidate.Label, Bytes: der})
		var err error
		if candidate.Kind == sourcePrivate {
			_, err = asym.LoadPrivateKeyPEM(pemBytes)
		} else {
			_, err = asym.LoadPublicKeyPEM(pemBytes)
		}
		if err != nil {
			continue
		}
		return &sourceKey{
			Kind:      candidate.Kind,
			Container: FormatDER,
			Encoding:  candidate.Encoding,
			PEM:       pemBytes,
		}, nil
	}
	return nil, toolerr.Newf(toolerr.CodeInvalidDER, map[string]any{
		"field": "key.data",
	})
}

// convertPrivate 转换私钥。
func convertPrivate(source *sourceKey, opts ConvertOptions) (*ConvertResult, error) {
	priv, err := LoadPrivateKey(source.PEM, opts.Passphrase)
	if err != nil {
		return nil, err
	}
	defer func() { _ = priv.Close() }()

	info, err := describeKey(priv)
	if err != nil {
		return nil, err
	}

	result := &ConvertResult{
		Algorithm: info.Algorithm,
		KeyType:   KeyTypePrivate,
		Detected:  source.Encoding,
		Encrypted: source.Encrypted,
		Bits:      info.Bits,
		Curve:     info.Curve,
	}

	switch opts.To {
	case FormatPKCS8, FormatPEM:
		out, err := marshalCanonicalPrivate(priv, source.Encrypted, opts.Passphrase)
		if err != nil {
			return nil, err
		}
		result.PEM = out
	case FormatPKCS1:
		if source.Encrypted {
			return nil, unsupportedFormat(FormatPKCS1, []string{FormatPKCS8, FormatPEM, FormatDER})
		}
		out, err := priv.MarshalPKCS1PEM()
		if err != nil {
			return nil, err
		}
		result.PEM = out
	case FormatSPKI:
		return nil, unsupportedFormat(FormatSPKI, []string{FormatPKCS8, FormatPKCS1, FormatPEM, FormatDER})
	case FormatDER:
		out, err := marshalCanonicalPrivate(priv, source.Encrypted, opts.Passphrase)
		if err != nil {
			return nil, err
		}
		der, err := pemToDER(out)
		if err != nil {
			return nil, err
		}
		result.DER = der
	default:
		return nil, unsupportedFormat(opts.To, KeyFormats())
	}

	return result, nil
}

// convertPublic 转换公钥。
func convertPublic(source *sourceKey, opts ConvertOptions) (*ConvertResult, error) {
	pub, err := LoadPublicKey(source.PEM)
	if err != nil {
		return nil, err
	}
	defer func() { _ = pub.Close() }()

	info, err := describePublicKey(pub)
	if err != nil {
		return nil, err
	}

	result := &ConvertResult{
		Algorithm: info.Algorithm,
		KeyType:   KeyTypePublic,
		Detected:  source.Encoding,
		Bits:      info.Bits,
		Curve:     info.Curve,
	}

	switch opts.To {
	case FormatSPKI, FormatPEM:
		out, err := pub.MarshalPEM()
		if err != nil {
			return nil, err
		}
		result.PEM = out
	case FormatDER:
		out, err := pub.MarshalPEM()
		if err != nil {
			return nil, err
		}
		der, err := pemToDER(out)
		if err != nil {
			return nil, err
		}
		result.DER = der
	case FormatPKCS8:
		return nil, unsupportedFormat(FormatPKCS8, []string{FormatSPKI, FormatPEM, FormatDER})
	case FormatPKCS1:
		return nil, unsupportedFormat(FormatPKCS1, []string{FormatSPKI, FormatPEM, FormatDER})
	default:
		return nil, unsupportedFormat(opts.To, KeyFormats())
	}

	return result, nil
}

// marshalCanonicalPrivate 导出私钥的规范形态（PKCS#8 PEM）。
//
// encrypted 为真时保持加密（AES-256-CBC + 原口令），
// 避免转换动作把加密私钥「顺手解密」暴露出去。
func marshalCanonicalPrivate(priv *PrivateKey, encrypted bool, passphrase string) ([]byte, error) {
	if !encrypted {
		return priv.MarshalPKCS8PEM()
	}
	return asym.MarshalEncryptedPrivateKeyPEMWithCipher(priv.key, "aes-256-cbc", passphrase)
}

// pemToDER 剥掉 PEM 外壳，取出 DER 字节。
func pemToDER(pemBytes []byte) ([]byte, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field": "key.data",
		})
	}
	return block.Bytes, nil
}

// containsString 判断字符串切片是否包含目标值。
func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
