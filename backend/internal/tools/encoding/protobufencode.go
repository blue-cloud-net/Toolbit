package encoding

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ProtobufEncodeRequest 是 protobuf-encoder 工具的请求。
type ProtobufEncodeRequest struct {
	// Mode 是编码模式：wire（按字段规格）或 schema（按 protojson）。
	Mode string `json:"mode"`
	// Schema 是 .proto 描述文本；mode 为 schema 时必填。
	Schema string `json:"schema"`
	// MessageType 是顶层消息名；schema 中顶层消息唯一时可省略。
	MessageType string `json:"messageType"`
	// Message 是 protojson 文本；mode 为 schema 时必填。
	Message string `json:"message"`
	// Fields 是字段规格；mode 为 wire 时必填。
	Fields []ProtoEncodeField `json:"fields"`
	// OutputEncoding 是输出编码；缺省 base64。
	OutputEncoding string `json:"outputEncoding"`
}

// ProtobufEncodeResult 是 protobuf-encoder 工具的响应。
type ProtobufEncodeResult struct {
	// Output 是编码结果。
	Output string `json:"output"`
	// Encoding 是实际使用的输出编码。
	Encoding string `json:"encoding"`
	// ByteLength 是结果字节数。
	ByteLength int `json:"byteLength"`
	// FieldCount 是写入的字段总数（含嵌套字段）。
	FieldCount int `json:"fieldCount"`
}

// ProtobufEncode 按字段规格或 protojson 编码出 protobuf 字节流。
func ProtobufEncode(_ context.Context, req *ProtobufEncodeRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	var (
		raw   []byte
		count int
		err   error
	)

	switch req.mode() {
	case ProtoModeWire:
		raw, count, err = protoEncodeWire(req.Fields)
		if err != nil {
			return tools.Result{}, err
		}

	case ProtoModeSchema:
		schema, compileErr := compileProtoSchema(req.Schema)
		if compileErr != nil {
			return tools.Result{}, compileErr
		}
		md, findErr := schema.findMessage(req.MessageType)
		if findErr != nil {
			return tools.Result{}, findErr
		}
		raw, count, err = protoSchemaEncode(req.Message, schema, md, protoWireOptions{
			maxDepth:     ProtoDefaultMaxDepth,
			maxFields:    ProtoDefaultMaxFields,
			inferStrings: true,
		})
		if err != nil {
			return tools.Result{}, err
		}
	}

	enc, err := codecsvc.Normalize(req.outputEncoding())
	if err != nil {
		return tools.Result{}, err
	}
	encoded, err := codecsvc.Encode(raw, enc)
	if err != nil {
		return tools.Result{}, err
	}

	return tools.Result{Data: ProtobufEncodeResult{
		Output:     encoded,
		Encoding:   enc,
		ByteLength: len(raw),
		FieldCount: count,
	}}, nil
}

// mode 返回编码模式；缺省为 wire。
func (r *ProtobufEncodeRequest) mode() string {
	if r.Mode == "" {
		return ProtoModeWire
	}
	return r.Mode
}

// outputEncoding 返回输出编码；缺省 base64。
func (r *ProtobufEncodeRequest) outputEncoding() string {
	if r.OutputEncoding == "" {
		return codecsvc.EncodingBase64
	}
	return r.OutputEncoding
}

// Validate 做字段级校验。
func (r *ProtobufEncodeRequest) Validate() error {
	switch r.Mode {
	case "", ProtoModeWire, ProtoModeSchema:
	default:
		return tools.UnsupportedMode(r.Mode, []string{ProtoModeWire, ProtoModeSchema})
	}

	if r.Mode == ProtoModeSchema {
		if r.Schema == "" {
			return tools.MissingField("schema")
		}
		if r.Message == "" {
			return tools.MissingField("message")
		}
	}

	if r.Mode == ProtoModeWire && len(r.Fields) == 0 {
		return tools.MissingField("fields")
	}

	for _, f := range r.Fields {
		if f.Type == "bytes" && f.Encoding != "" {
			if _, err := codecsvc.Normalize(f.Encoding); err != nil {
				return err
			}
		}
	}

	if _, err := codecsvc.Normalize(r.outputEncoding()); err != nil {
		return err
	}

	return nil
}
