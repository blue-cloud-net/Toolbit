package encoding

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// Protobuf 解析与编码模式。
const (
	// ProtoModeWire 是无 schema 的线格式模式：只按 protobuf 二进制线格式解释字段。
	ProtoModeWire = "wire"
	// ProtoModeSchema 是 .proto schema 模式：按描述文件解释字段名与声明类型。
	ProtoModeSchema = "schema"
)

// Protobuf 线类型标识（响应里的 `wireType` 取值）。
const (
	// ProtoWireVarint 是变长整数。
	ProtoWireVarint = "varint"
	// ProtoWireFixed32 是 32 位定长。
	ProtoWireFixed32 = "fixed32"
	// ProtoWireFixed64 是 64 位定长。
	ProtoWireFixed64 = "fixed64"
	// ProtoWireBytes 是长度分隔。
	ProtoWireBytes = "bytes"
	// ProtoWireGroup 是分组（protobuf 2 的分组线格式，已废弃）。
	ProtoWireGroup = "group"
)

// protobuf 字段号边界。
const (
	// ProtoMinFieldNumber 是最小合法字段号。
	ProtoMinFieldNumber = 1
	// ProtoMaxFieldNumber 是最大合法字段号（2^29-1）。
	ProtoMaxFieldNumber = 536870911
)

// protobuf 解析上限。
const (
	// ProtoDefaultMaxDepth 是默认嵌套解析深度。
	ProtoDefaultMaxDepth = 32
	// ProtoLimitMaxDepth 是嵌套解析深度上限。
	ProtoLimitMaxDepth = 128
	// ProtoDefaultMaxFields 是默认字段总数上限。
	ProtoDefaultMaxFields = 10000
	// ProtoLimitMaxFields 是字段总数上限。
	ProtoLimitMaxFields = 100000
	// protoHexPreviewBytes 是二进制值在字段表里展示的十六进制字节上限。
	protoHexPreviewBytes = 64
)

// ProtoField 是解析出的一个 protobuf 字段。
type ProtoField struct {
	// Path 是字段路径：根为 `root`，子字段逐级追加 `/<字段号>`，
	// 重复字段的元素追加 `[<下标>]`。
	Path string `json:"path"`
	// Offset 是该字段在线格式字节流中的起始偏移。
	Offset int `json:"offset"`
	// FieldNumber 是字段号。
	FieldNumber int `json:"fieldNumber"`
	// Name 是字段名；仅 schema 模式有值。
	Name string `json:"name,omitempty"`
	// WireType 是线类型。
	WireType string `json:"wireType"`
	// DeclaredType 是声明类型；仅 schema 模式有值。
	DeclaredType string `json:"declaredType,omitempty"`
	// Value 是字段值的文本表示。
	Value string `json:"value"`
	// Length 是长度分隔字段的字节数。
	Length int `json:"length,omitempty"`
	// Text 表示该值被识别为 UTF-8 文本。
	Text bool `json:"text"`
	// Message 表示该值被识别为嵌套消息。
	Message bool `json:"message"`
	// Hex 是二进制值的十六进制预览（最多 64 字节）。
	Hex string `json:"hex,omitempty"`
	// Float 是 32 / 64 位浮点字段的浮点解释。
	Float *float64 `json:"float,omitempty"`
	// Children 是嵌套消息的子字段。
	Children []ProtoField `json:"children,omitempty"`

	// raw 是字段的原始字节；只用于生成 JSON 视图，**不参与序列化**。
	raw []byte
}

// invalidProtobuf 构造 protobuf 解析错误。
//
// 错误文案**不**回显用户输入；参数只放结构性元信息
// （偏移、字段名、可用取值列表）。
func invalidProtobuf(reason string, params map[string]any) *toolerr.Error {
	if params == nil {
		params = map[string]any{}
	}
	params["reason"] = reason
	return toolerr.Newf(toolerr.CodeInvalidProtobuf, params)
}
