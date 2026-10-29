package pkisvc

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/blue-cloud-net/tongsuo-go/asn1"
	"github.com/blue-cloud-net/tongsuo-go/x509"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// ASN.1 视图的默认与硬上限。
const (
	// ASN1DefaultDepth 是默认展开深度。
	ASN1DefaultDepth = 32
	// ASN1MaxDepth 是展开深度的硬上限。
	//
	// 超过该深度的输入整体拒绝（TOO_LARGE）。
	ASN1MaxDepth = 64
	// ASN1MaxNodes 是节点数的硬上限。
	//
	// 超过该数量的输入整体拒绝（TOO_MANY_ITEMS）。
	ASN1MaxNodes = 10000
	// ASN1MaxBytes 是 ASN.1 输入的字节上限。
	ASN1MaxBytes = 1 << 20
)

// ASN1Node 是一个 DER TLV 节点。
type ASN1Node struct {
	// Offset 是该节点的 tag 字节在输入中的绝对偏移。
	Offset int `json:"offset"`
	// Tag 是首个 tag 字节。
	Tag byte `json:"tag"`
	// Class 是 ASN.1 类（0 通用 / 1 应用 / 2 上下文 / 3 私有）。
	Class int `json:"class"`
	// Number 是 tag 编号。
	Number int `json:"number"`
	// Constructed 报告是否为构造类型。
	Constructed bool `json:"constructed"`
	// Length 是 value 的字节长度。
	Length int `json:"length"`
	// Value 是 value 的十六进制；构造类型为空。
	Value string `json:"value"`
	// Truncated 报告该节点因超出展开深度而未展开子节点。
	Truncated bool `json:"truncated"`
	// Children 是子节点。
	Children []ASN1Node `json:"children,omitempty"`
}

// ASN1Result 是 ASN.1 视图的结果。
type ASN1Result struct {
	// Length 是输入的字节长度。
	Length int `json:"length"`
	// Tree 是根节点。
	Tree ASN1Node `json:"tree"`
	// NodeCount 是解析出的节点总数（含未展开的）。
	NodeCount int `json:"nodeCount"`
	// Depth 是解析出的最大深度。
	Depth int `json:"depth"`
	// MaxDepth 是实际展开的深度上限。
	MaxDepth int `json:"maxDepth"`
	// HexDump 是十六进制转储。
	HexDump string `json:"hexDump"`
}

// ParseASN1 把 DER 解析成可读的 TLV 树。
//
// maxDepth 是**展开**深度（0 表示默认 ASN1DefaultDepth，上限 ASN1MaxDepth）；
// 超出展开深度的子节点不再展开，但**仍计入**节点总数。
//
// 输入字节数、节点数或实际深度超过硬上限时整体拒绝：
// 分别返回 TOO_LARGE 与 TOO_MANY_ITEMS。
//
// ⚠️ 上游 `asn1.Parse` **没有**深度 / 节点限制参数，只能先全量解析再统计
// （已登记于 docs/upstream/tongsuo-go-issues.md I-16）；
// 因此输入体积上限必须由 catalog 的 maxBodyBytes 先行约束。
func ParseASN1(der []byte, maxDepth int) (*ASN1Result, error) {
	if len(der) == 0 {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "input.data"})
	}
	if len(der) > ASN1MaxBytes {
		return nil, toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"field":  "input.data",
			"limit":  ASN1MaxBytes,
			"actual": len(der),
		})
	}

	if maxDepth <= 0 {
		maxDepth = ASN1DefaultDepth
	}
	if maxDepth > ASN1MaxDepth {
		maxDepth = ASN1MaxDepth
	}

	root, err := asn1.Parse(der)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidDER, map[string]any{
			"field": "input.data",
		})
	}

	count := 0
	depth := measureASN1(root, 1, &count)
	if count > ASN1MaxNodes {
		return nil, toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
			"field":  "input.data",
			"limit":  ASN1MaxNodes,
			"actual": count,
		})
	}
	if depth > ASN1MaxDepth {
		return nil, toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"field":  "input.data",
			"limit":  ASN1MaxDepth,
			"actual": depth,
		})
	}

	tree := buildASN1Node(root, 1, maxDepth)

	return &ASN1Result{
		Length:    len(der),
		Tree:      tree,
		NodeCount: count,
		Depth:     depth,
		MaxDepth:  maxDepth,
		HexDump:   HexDump(der),
	}, nil
}

// measureASN1 统计节点数与最大深度。
func measureASN1(node *asn1.Node, level int, count *int) int {
	if node == nil {
		return 0
	}
	*count++
	max := level
	for _, child := range node.Children {
		if d := measureASN1(child, level+1, count); d > max {
			max = d
		}
	}
	return max
}

// buildASN1Node 把底层节点转成对外视图，并按 maxDepth 截断展开。
func buildASN1Node(node *asn1.Node, level, maxDepth int) ASN1Node {
	if node == nil {
		return ASN1Node{}
	}

	out := ASN1Node{
		Offset:      node.Offset,
		Tag:         node.Tag,
		Class:       node.Class,
		Number:      node.Number,
		Constructed: node.Constructed,
		Length:      node.Length,
	}

	if len(node.Children) == 0 {
		out.Value = hex.EncodeToString(node.Value)
		return out
	}

	if level >= maxDepth {
		out.Truncated = true
		return out
	}

	out.Children = make([]ASN1Node, 0, len(node.Children))
	for _, child := range node.Children {
		out.Children = append(out.Children, buildASN1Node(child, level+1, maxDepth))
	}
	return out
}

// sanFromExtensions 从扩展列表里提取主体备用名条目。
//
// 上游只给 Certificate 提供了 SAN() 方法，CSR 只能自己从扩展的
// DER 值里解析（扩展值是 GeneralNames 的 DER 编码）。
//
// 找不到 subjectAltName 扩展时返回空切片。
func sanFromExtensions(extensions []x509.Extension) []string {
	for _, ext := range extensions {
		if ext.Field != subjectAltNameName && ext.OID != subjectAltNameOID {
			continue
		}
		return sanFromExtensionDER(ext.Data)
	}
	return []string{}
}

// subjectAltNameName 是 subjectAltName 扩展的短名。
const subjectAltNameName = "subjectAltName"

// subjectAltNameOID 是 subjectAltName 扩展的点分 OID。
const subjectAltNameOID = "2.5.29.17"

// sanFromExtensionDER 从 subjectAltName 扩展的 DER 值中提取条目。
//
// 条目格式与上游 Certificate.SAN() 一致：DNS: / IP: / email: / URI: / RegisteredID:。
func sanFromExtensionDER(der []byte) []string {
	if len(der) == 0 {
		return []string{}
	}
	root, err := asn1.Parse(der)
	if err != nil {
		return []string{}
	}

	out := make([]string, 0, len(root.Children))
	for _, child := range root.Children {
		if child.Class != asn1.ClassContextSpecific {
			continue
		}
		switch child.Number {
		case 1:
			out = append(out, "email:"+string(child.Value))
		case 2:
			out = append(out, "DNS:"+string(child.Value))
		case 6:
			out = append(out, "URI:"+string(child.Value))
		case 7:
			out = append(out, "IP:"+formatIP(child.Value))
		case 8:
			out = append(out, "RegisteredID:"+formatOID(child.Value))
		default:
			continue
		}
	}
	return out
}

// formatIP 把 iPAddress 的原始字节渲染成点分或冒号分隔的地址。
func formatIP(raw []byte) string {
	switch len(raw) {
	case 4:
		return fmt.Sprintf("%d.%d.%d.%d", raw[0], raw[1], raw[2], raw[3])
	case 16:
		parts := make([]string, 0, 8)
		for i := 0; i < 16; i += 2 {
			parts = append(parts, fmt.Sprintf("%x", int(raw[i])<<8|int(raw[i+1])))
		}
		return strings.Join(parts, ":")
	default:
		return hex.EncodeToString(raw)
	}
}

// formatOID 把 OID 的 DER 内容渲染成点分文本。
func formatOID(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	parts := []string{
		fmt.Sprintf("%d", raw[0]/40),
		fmt.Sprintf("%d", raw[0]%40),
	}
	value := uint64(0)
	for _, b := range raw[1:] {
		value = value<<7 | uint64(b&0x7f)
		if b&0x80 == 0 {
			parts = append(parts, fmt.Sprintf("%d", value))
			value = 0
		}
	}
	return strings.Join(parts, ".")
}

// HexDump 生成十六进制转储。
func HexDump(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var sb strings.Builder
	for offset := 0; offset < len(data); offset += 16 {
		end := offset + 16
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]

		fmt.Fprintf(&sb, "%08x: ", offset)
		for i := 0; i < 16; i++ {
			if i%2 == 0 {
				sb.WriteByte(' ')
			}
			if i < len(chunk) {
				fmt.Fprintf(&sb, "%02x", chunk[i])
			} else {
				sb.WriteString("  ")
			}
		}
		sb.WriteString("  ")
		for _, b := range chunk {
			if b >= 0x20 && b < 0x7f {
				sb.WriteByte(b)
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
