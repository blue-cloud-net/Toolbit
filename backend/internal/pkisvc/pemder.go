package pkisvc

import (
	"encoding/base64"
	"encoding/pem"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// PEM 块类型（OpenSSL 的标签）。
const (
	// PEMBlockCertificate 是 X.509 证书。
	PEMBlockCertificate = "CERTIFICATE"
	// PEMBlockCertificateRequest 是 PKCS#10 证书签名请求。
	PEMBlockCertificateRequest = "CERTIFICATE REQUEST"
	// PEMBlockCRL 是证书吊销列表。
	PEMBlockCRL = "X509 CRL"
	// PEMBlockPrivateKey 是 PKCS#8 私钥。
	PEMBlockPrivateKey = "PRIVATE KEY"
	// PEMBlockEncryptedPrivateKey 是加密的 PKCS#8 私钥。
	PEMBlockEncryptedPrivateKey = "ENCRYPTED PRIVATE KEY"
	// PEMBlockRSAKey 是 PKCS#1 私钥。
	PEMBlockRSAKey = "RSA PRIVATE KEY"
	// PEMBlockPublicKey 是 SPKI 公钥。
	PEMBlockPublicKey = "PUBLIC KEY"
)

// PEM 类型过滤值。
const (
	// PEMTypeCert 只保留证书。
	PEMTypeCert = "cert"
	// PEMTypeKey 只保留密钥。
	PEMTypeKey = "key"
	// PEMTypeCSR 只保留证书签名请求。
	PEMTypeCSR = "csr"
	// PEMTypeCRL 只保留吊销列表。
	PEMTypeCRL = "crl"
	// PEMTypeAny 保留全部。
	PEMTypeAny = "any"
)

// 转换目标格式。
const (
	// ConvertToPEM 输出 PEM 文本。
	ConvertToPEM = "pem"
	// ConvertToDER 输出 DER 二进制。
	ConvertToDER = "der"
)

// MaxPEMBlocks 是单个输入允许的 PEM 块数上限。
const MaxPEMBlocks = 64

// PEMTypes 返回受支持的 PEM 类型过滤值（顺序稳定）。
func PEMTypes() []string {
	return []string{PEMTypeCert, PEMTypeKey, PEMTypeCSR, PEMTypeCRL, PEMTypeAny}
}

// ConvertTargets 返回受支持的转换目标（顺序稳定）。
func ConvertTargets() []string {
	return []string{ConvertToPEM, ConvertToDER}
}

// PEMBlock 是一个 PEM 块。
type PEMBlock struct {
	// Type 是 PEM 标签（如 CERTIFICATE）。
	Type string
	// Label 是 PEM 标签的**规范化**形式（去掉空格）。
	//
	// 例如 "CERTIFICATE REQUEST" 的 Label 是 "CERTIFICATEREQUEST"，
	// 便于做不区分空格的匹配。
	Label string
	// Bytes 是块体的原始字节（DER）。
	Bytes []byte
}

// PEMConvertOptions 是 PEM ↔ DER 转换的参数。
type PEMConvertOptions struct {
	// Source 是源字节：PEM 文本或 DER 二进制。
	Source []byte
	// SourceIsPEM 报告 Source 是否是 PEM 文本。
	SourceIsPEM bool
	// Type 是类型过滤值，取值域见 PEMTypes；PEMTypeAny 表示不过滤。
	Type string
	// To 是目标格式，取值域见 ConvertTargets。
	To string
	// Label 是输出 PEM 的标签；为空时沿用源标签。
	Label string
}

// PEMConvertResult 是 PEM ↔ DER 转换的结果。
type PEMConvertResult struct {
	// Data 是输出文本：目标为 der 时是 base64，否则是 PEM 文本。
	Data string
	// Format 是输出格式（pem / der）。
	Format string
	// Encoding 是 Data 的传输编码（base64 / utf8）。
	Encoding string
	// Labels 是**输入**中识别到的 PEM 标签（按出现顺序）。
	Labels []string
	// BlockCount 是输出的 PEM 块数。
	BlockCount int
}

// SplitPEM 把 PEM 文本拆成若干块。
//
// 输入不是 PEM（或一个块都没有）时返回 INVALID_PEM；
// 块数超过 MaxPEMBlocks 时返回 TOO_MANY_ITEMS。
func SplitPEM(data []byte) ([]PEMBlock, error) {
	if len(data) == 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field": "input.data",
		})
	}

	var blocks []PEMBlock
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		blocks = append(blocks, PEMBlock{
			Type:  block.Type,
			Label: normalizePEMLabel(block.Type),
			Bytes: block.Bytes,
		})
		if len(blocks) > MaxPEMBlocks {
			return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
				"field":  "input.data",
				"limit":  MaxPEMBlocks,
				"actual": len(blocks),
			})
		}
	}

	if len(blocks) == 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field": "input.data",
		})
	}
	return blocks, nil
}

// NormalizeToDER 把输入统一成 DER 字节。
//
// isPEM 为真时剥掉 PEM 外壳；否则原样返回（调用方负责已解码为二进制）。
func NormalizeToDER(source []byte, isPEM bool) ([]byte, error) {
	if len(source) == 0 {
		return nil, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field": "input.data",
		})
	}
	if !isPEM {
		return source, nil
	}
	block, _ := pem.Decode(source)
	if block == nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidPEM, map[string]any{
			"field": "input.data",
		})
	}
	return block.Bytes, nil
}

// ConvertPEMDER 在 PEM 与 DER 之间转换。
//
// 规则：
//
//   - 源是 PEM 时先按 Type 过滤块；过滤后为空返回 UNSUPPORTED_FORMAT；
//   - 源是 DER 时视为单个块，Type 只用于推断默认标签；
//   - 目标为 der 时**只允许一个块**（DER 装不下多块），多个块返回 UNSUPPORTED_FORMAT；
//   - 目标为 pem 时按源顺序拼接，Label 非空时覆盖每一块的标签。
func ConvertPEMDER(opts PEMConvertOptions) (*PEMConvertResult, error) {
	if len(opts.Source) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "input.data"})
	}
	if !containsString(ConvertTargets(), opts.To) {
		return nil, unsupportedFormat(opts.To, ConvertTargets())
	}
	if opts.Type == "" {
		opts.Type = PEMTypeAny
	}
	if !containsString(PEMTypes(), opts.Type) {
		return nil, unsupportedFormat(opts.Type, PEMTypes())
	}

	blocks, labels, err := splitSourceBlocks(opts)
	if err != nil {
		return nil, err
	}

	filtered := filterBlocks(blocks, opts.Type)
	if len(filtered) == 0 {
		return nil, unsupportedFormat(opts.Type, PEMTypes())
	}

	if opts.To == ConvertToDER && len(filtered) > 1 {
		return nil, unsupportedFormat(ConvertToDER, []string{ConvertToPEM})
	}

	result := &PEMConvertResult{
		Format:     opts.To,
		Labels:     labels,
		BlockCount: len(filtered),
	}

	if opts.To == ConvertToDER {
		result.Data = base64.StdEncoding.EncodeToString(filtered[0].Bytes)
		result.Encoding = "base64"
		return result, nil
	}

	var out []byte
	for _, block := range filtered {
		label := block.Type
		if opts.Label != "" {
			label = opts.Label
		}
		out = append(out, pem.EncodeToMemory(&pem.Block{Type: label, Bytes: block.Bytes})...)
	}
	result.Data = string(out)
	result.Encoding = "utf8"
	return result, nil
}

// splitSourceBlocks 把源拆成块，并收集识别到的标签。
func splitSourceBlocks(opts PEMConvertOptions) ([]PEMBlock, []string, error) {
	if !opts.SourceIsPEM {
		return []PEMBlock{{
			Type:  defaultLabelForType(opts.Type),
			Label: normalizePEMLabel(defaultLabelForType(opts.Type)),
			Bytes: opts.Source,
		}}, nil, nil
	}

	blocks, err := SplitPEM(opts.Source)
	if err != nil {
		return nil, nil, err
	}
	labels := make([]string, 0, len(blocks))
	for _, block := range blocks {
		labels = append(labels, block.Type)
	}
	return blocks, labels, nil
}

// filterBlocks 按类型过滤 PEM 块。
func filterBlocks(blocks []PEMBlock, filter string) []PEMBlock {
	if filter == PEMTypeAny {
		return blocks
	}
	out := make([]PEMBlock, 0, len(blocks))
	for _, block := range blocks {
		if pemBlockMatches(block.Label, filter) {
			out = append(out, block)
		}
	}
	return out
}

// pemBlockMatches 判断块标签是否属于指定类型。
func pemBlockMatches(label, filter string) bool {
	switch filter {
	case PEMTypeCert:
		return label == normalizePEMLabel(PEMBlockCertificate)
	case PEMTypeKey:
		switch label {
		case normalizePEMLabel(PEMBlockPrivateKey),
			normalizePEMLabel(PEMBlockEncryptedPrivateKey),
			normalizePEMLabel(PEMBlockRSAKey),
			normalizePEMLabel(PEMBlockPublicKey):
			return true
		}
		return false
	case PEMTypeCSR:
		return label == normalizePEMLabel(PEMBlockCertificateRequest)
	case PEMTypeCRL:
		return label == normalizePEMLabel(PEMBlockCRL)
	default:
		return false
	}
}

// defaultLabelForType 按类型推断默认 PEM 标签。
func defaultLabelForType(filter string) string {
	switch filter {
	case PEMTypeCert:
		return PEMBlockCertificate
	case PEMTypeCSR:
		return PEMBlockCertificateRequest
	case PEMTypeCRL:
		return PEMBlockCRL
	case PEMTypeKey:
		return PEMBlockPublicKey
	default:
		return PEMBlockCertificate
	}
}

// normalizePEMLabel 归一化 PEM 标签：去掉空白并转大写。
func normalizePEMLabel(label string) string {
	return strings.ToUpper(strings.ReplaceAll(label, " ", ""))
}

// LooksLikePEM 判断字节是否是 PEM 文本。
func LooksLikePEM(data []byte) bool {
	return strings.Contains(string(data), "-----BEGIN ")
}

// pemEncode 把 DER 字节包装成指定标签的 PEM。
func pemEncode(label string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: label, Bytes: der})
}
