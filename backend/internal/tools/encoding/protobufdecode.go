package encoding

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ProtobufDecodeRequest 是 protobuf-decoder 工具的请求。
type ProtobufDecodeRequest struct {
	// Mode 是解析模式：wire（无 schema）或 schema（按 .proto 描述）。
	Mode string `json:"mode"`
	// Input 是待解析的字节数据。
	Input tools.Bytes `json:"input"`
	// Schema 是 .proto 描述文本；mode 为 schema 时必填。
	Schema string `json:"schema"`
	// MessageType 是顶层消息名；schema 中顶层消息唯一时可省略。
	MessageType string `json:"messageType"`
	// MaxDepth 是嵌套解析深度上限；缺省 32，上限 128。
	MaxDepth *int `json:"maxDepth"`
	// MaxFields 是字段总数上限；缺省 10000，上限 100000。
	MaxFields *int `json:"maxFields"`
	// InferStrings 是否优先把长度分隔字段识别为 UTF-8 文本；缺省 true。
	InferStrings *bool `json:"inferStrings"`
	// IncludeHexDump 是否返回十六进制转储；缺省 false。
	IncludeHexDump *bool `json:"includeHexDump"`
}

// ProtobufDecodeResult 是 protobuf-decoder 工具的响应。
type ProtobufDecodeResult struct {
	// Mode 是实际使用的解析模式。
	Mode string `json:"mode"`
	// MessageType 是解析所用的消息名；wire 模式为空。
	MessageType string `json:"messageType"`
	// ByteLength 是输入字节数。
	ByteLength int `json:"byteLength"`
	// FieldCount 是解析出的字段总数（含嵌套字段）。
	FieldCount int `json:"fieldCount"`
	// Depth 是实际达到的最大嵌套深度。
	Depth int `json:"depth"`
	// MaxDepth 是生效的深度上限。
	MaxDepth int `json:"maxDepth"`
	// Fields 是顶层字段列表。
	Fields []ProtoField `json:"fields"`
	// JSON 是解析结果的 JSON 文本。
	JSON string `json:"json"`
	// UnknownFieldCount 是 schema 模式下无法识别的字段数。
	UnknownFieldCount int `json:"unknownFieldCount"`
	// HexDump 是十六进制转储；未请求时为空。
	HexDump string `json:"hexDump"`
	// Truncated 表示结果因上限被截断。
	Truncated bool `json:"truncated"`
}

// ProtobufDecode 解析 protobuf 字节流。
func ProtobufDecode(_ context.Context, req *ProtobufDecodeRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := req.Input.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	opts := protoWireOptions{
		maxDepth:     req.maxDepth(),
		maxFields:    req.maxFields(),
		inferStrings: req.inferStrings(),
	}

	result := ProtobufDecodeResult{
		Mode:       req.mode(),
		ByteLength: len(raw),
		MaxDepth:   opts.maxDepth,
		Fields:     []ProtoField{},
	}

	switch req.mode() {
	case ProtoModeWire:
		parsed, err := parseProtoWire(raw, opts)
		if err != nil {
			return tools.Result{}, err
		}
		result.Fields = parsed.fields
		result.FieldCount = parsed.fieldCount
		result.Depth = parsed.depth
		result.Truncated = parsed.truncated
		result.JSON = protoWireJSON(parsed.fields)

	case ProtoModeSchema:
		schema, err := compileProtoSchema(req.Schema)
		if err != nil {
			return tools.Result{}, err
		}
		md, err := schema.findMessage(req.MessageType)
		if err != nil {
			return tools.Result{}, err
		}
		fields, jsonText, unknown, err := protoSchemaDecode(raw, schema, md, opts)
		if err != nil {
			return tools.Result{}, err
		}
		result.MessageType = string(md.FullName())
		result.Fields = fields
		result.FieldCount = protoCountFields(fields)
		result.Depth = protoTreeDepth(fields)
		result.JSON = jsonText
		result.UnknownFieldCount = unknown
	}

	if req.includeHexDump() {
		result.HexDump = pkisvc.HexDump(raw)
	}

	return tools.Result{Data: result}, nil
}

// mode 返回解析模式；缺省为 wire。
func (r *ProtobufDecodeRequest) mode() string {
	if r.Mode == "" {
		return ProtoModeWire
	}
	return r.Mode
}

// maxDepth 返回嵌套解析深度上限；缺省 32。
func (r *ProtobufDecodeRequest) maxDepth() int {
	if r.MaxDepth == nil {
		return ProtoDefaultMaxDepth
	}
	return *r.MaxDepth
}

// maxFields 返回字段总数上限；缺省 10000。
func (r *ProtobufDecodeRequest) maxFields() int {
	if r.MaxFields == nil {
		return ProtoDefaultMaxFields
	}
	return *r.MaxFields
}

// inferStrings 返回是否优先把长度分隔字段识别为文本；缺省 true。
func (r *ProtobufDecodeRequest) inferStrings() bool {
	if r.InferStrings == nil {
		return true
	}
	return *r.InferStrings
}

// includeHexDump 返回是否输出十六进制转储；缺省 false。
func (r *ProtobufDecodeRequest) includeHexDump() bool {
	if r.IncludeHexDump == nil {
		return false
	}
	return *r.IncludeHexDump
}

// Validate 做字段级校验。
func (r *ProtobufDecodeRequest) Validate() error {
	switch r.Mode {
	case "", ProtoModeWire, ProtoModeSchema:
	default:
		return tools.UnsupportedMode(r.Mode, []string{ProtoModeWire, ProtoModeSchema})
	}

	if r.Mode == ProtoModeSchema && r.Schema == "" {
		return tools.MissingField("schema")
	}

	if r.MaxDepth != nil && (*r.MaxDepth < 0 || *r.MaxDepth > ProtoLimitMaxDepth) {
		return tools.InvalidField("maxDepth", []int{0, ProtoLimitMaxDepth}, *r.MaxDepth)
	}

	if r.MaxFields != nil && (*r.MaxFields < 1 || *r.MaxFields > ProtoLimitMaxFields) {
		return tools.InvalidField("maxFields", []int{1, ProtoLimitMaxFields}, *r.MaxFields)
	}

	return r.Input.Validate("input")
}

// protoCountFields 统计字段树中的字段总数。
func protoCountFields(fields []ProtoField) int {
	count := 0
	for _, f := range fields {
		count++
		count += protoCountFields(f.Children)
	}
	return count
}

// protoTreeDepth 返回字段树的最大嵌套层数（扁平消息为 0）。
func protoTreeDepth(fields []ProtoField) int {
	depth := 0
	for _, f := range fields {
		if len(f.Children) == 0 {
			continue
		}
		if sub := protoTreeDepth(f.Children) + 1; sub > depth {
			depth = sub
		}
	}
	return depth
}
