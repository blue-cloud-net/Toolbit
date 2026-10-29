package pki

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ASN1ViewRequest 是 ASN.1 视图工具的请求。
type ASN1ViewRequest struct {
	// Input 是待查看的 DER 材料（PEM 会先剥掉外壳）。
	Input PemBundle `json:"input"`
	// MaxDepth 是最大展开深度；为空时用默认 32。
	MaxDepth int `json:"maxDepth"`
}

// ASN1ViewResult 是 ASN.1 视图工具的响应。
type ASN1ViewResult struct {
	// Length 是输入的字节长度。
	Length int `json:"length"`
	// Tree 是根节点。
	Tree ASN1NodeOutput `json:"tree"`
	// NodeCount 是解析出的节点总数。
	NodeCount int `json:"nodeCount"`
	// Depth 是解析出的最大深度。
	Depth int `json:"depth"`
	// MaxDepth 是实际展开的深度上限。
	MaxDepth int `json:"maxDepth"`
	// HexDump 是十六进制转储。
	HexDump string `json:"hexDump"`
}

// ASN1NodeOutput 是一个 DER TLV 节点的输出形态。
type ASN1NodeOutput struct {
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
	Children []ASN1NodeOutput `json:"children,omitempty"`
}

// Validate 做字段级校验。
func (r *ASN1ViewRequest) Validate() error {
	if err := r.Input.Validate("input"); err != nil {
		return err
	}
	if r.MaxDepth < 0 {
		return tools.InvalidField("maxDepth", []int{0, pkisvc.ASN1MaxDepth}, r.MaxDepth)
	}
	if r.MaxDepth > pkisvc.ASN1MaxDepth {
		return tools.InvalidField("maxDepth", []int{0, pkisvc.ASN1MaxDepth}, r.MaxDepth)
	}
	return nil
}

// ASN1View 把 DER 展开成可读的 TLV 树。
func ASN1View(_ context.Context, req *ASN1ViewRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source, isPEM, err := req.Input.Resolve()
	if err != nil {
		return tools.Result{}, err
	}

	der, err := pkisvc.NormalizeToDER(source, isPEM)
	if err != nil {
		return tools.Result{}, err
	}

	out, err := pkisvc.ParseASN1(der, req.MaxDepth)
	if err != nil {
		return tools.Result{}, err
	}

	result := &ASN1ViewResult{
		Length:    out.Length,
		Tree:      asn1NodeOutput(out.Tree),
		NodeCount: out.NodeCount,
		Depth:     out.Depth,
		MaxDepth:  out.MaxDepth,
		HexDump:   out.HexDump,
	}

	return tools.Result{Data: result}, nil
}

// asn1NodeOutput 把节点树转成输出形态。
func asn1NodeOutput(node pkisvc.ASN1Node) ASN1NodeOutput {
	out := ASN1NodeOutput{
		Offset:      node.Offset,
		Tag:         node.Tag,
		Class:       node.Class,
		Number:      node.Number,
		Constructed: node.Constructed,
		Length:      node.Length,
		Value:       node.Value,
		Truncated:   node.Truncated,
	}
	if len(node.Children) == 0 {
		return out
	}
	out.Children = make([]ASN1NodeOutput, 0, len(node.Children))
	for _, child := range node.Children {
		out.Children = append(out.Children, asn1NodeOutput(child))
	}
	return out
}
