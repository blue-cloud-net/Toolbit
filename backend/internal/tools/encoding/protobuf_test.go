package encoding

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// protoHex 把十六进制串解析成字节。
func protoHex(t *testing.T, s string) []byte {
	t.Helper()
	raw, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		t.Fatalf("十六进制串非法: %v", err)
	}
	return raw
}

// protoDecodeHex 走 wire 模式解码一段十六进制字节流。
func protoDecodeHex(t *testing.T, s string) ProtobufDecodeResult {
	t.Helper()
	res, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:  ProtoModeWire,
		Input: ByteInput{Data: s, Encoding: EncodingHex},
	})
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	return res.Data.(ProtobufDecodeResult)
}

func TestProtobufDecodeWireVarint(t *testing.T) {
	// 字段 1，varint，值 150。
	res := protoDecodeHex(t, "08 96 01")

	if res.FieldCount != 1 {
		t.Fatalf("字段数 = %d，期望 1", res.FieldCount)
	}
	field := res.Fields[0]
	if field.FieldNumber != 1 {
		t.Errorf("字段号 = %d，期望 1", field.FieldNumber)
	}
	if field.WireType != ProtoWireVarint {
		t.Errorf("线类型 = %q，期望 %q", field.WireType, ProtoWireVarint)
	}
	if field.Value != "150" {
		t.Errorf("值 = %q，期望 150", field.Value)
	}
	if field.Path != "root/1" {
		t.Errorf("路径 = %q，期望 root/1", field.Path)
	}
}

func TestProtobufDecodeWireString(t *testing.T) {
	// 字段 1，长度分隔，文本 "testing"。
	res := protoDecodeHex(t, "0a 07 74 65 73 74 69 6e 67")

	if len(res.Fields) != 1 {
		t.Fatalf("字段数 = %d，期望 1", len(res.Fields))
	}
	field := res.Fields[0]
	if !field.Text {
		t.Error("应识别为文本字段")
	}
	if field.Value != "testing" {
		t.Errorf("值 = %q，期望 testing", field.Value)
	}
	if field.Length != 7 {
		t.Errorf("长度 = %d，期望 7", field.Length)
	}
}

func TestProtobufDecodeWireNestedMessage(t *testing.T) {
	// 字段 1 是嵌套消息 {1: 1}。
	res := protoDecodeHex(t, "0a 02 08 01")

	if len(res.Fields) != 1 {
		t.Fatalf("字段数 = %d，期望 1", len(res.Fields))
	}
	field := res.Fields[0]
	if !field.Message {
		t.Fatal("应识别为嵌套消息")
	}
	if len(field.Children) != 1 {
		t.Fatalf("子字段数 = %d，期望 1", len(field.Children))
	}
	if field.Children[0].Value != "1" {
		t.Errorf("子字段值 = %q，期望 1", field.Children[0].Value)
	}
	if res.Depth != 1 {
		t.Errorf("深度 = %d，期望 1", res.Depth)
	}
}

func TestProtobufDecodeWireSIntNegative(t *testing.T) {
	// sint32 的 -1 编码为 zigzag 后的 1。
	res := protoDecodeHex(t, "08 01")
	if res.Fields[0].Value != "1" {
		t.Errorf("值 = %q，期望 1", res.Fields[0].Value)
	}
}

func TestProtobufDecodeWireFixed(t *testing.T) {
	// 字段 2，fixed32，值 1。
	res := protoDecodeHex(t, "15 01 00 00 00")
	field := res.Fields[0]
	if field.WireType != ProtoWireFixed32 {
		t.Errorf("线类型 = %q，期望 %q", field.WireType, ProtoWireFixed32)
	}
	if field.Value != "1" {
		t.Errorf("值 = %q，期望 1", field.Value)
	}
	if field.Float == nil {
		t.Fatal("应给出浮点解释")
	}
}

func TestProtobufDecodeWireRejectsMalformed(t *testing.T) {
	// 标签声明 varint，但缺少值。
	_, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:  ProtoModeWire,
		Input: ByteInput{Data: "08", Encoding: EncodingHex},
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidProtobuf)
}

func TestProtobufDecodeWireRejectsInvalidTag(t *testing.T) {
	// 字段号 0 是非法的。
	_, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:  ProtoModeWire,
		Input: ByteInput{Data: "00", Encoding: EncodingHex},
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidProtobuf)
}

func TestProtobufDecodeWireTruncatesOnFieldLimit(t *testing.T) {
	limit := 2
	res, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:      ProtoModeWire,
		Input:     ByteInput{Data: "08 01 08 02 08 03 08 04", Encoding: EncodingHex},
		MaxFields: &limit,
	})
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	out := res.Data.(ProtobufDecodeResult)
	if !out.Truncated {
		t.Error("应标记为已截断")
	}
	if out.FieldCount != limit {
		t.Errorf("字段数 = %d，期望 %d", out.FieldCount, limit)
	}
}

func TestProtobufDecodeWireJSON(t *testing.T) {
	res := protoDecodeHex(t, "0a 05 41 6c 69 63 65 10 01")
	if !strings.Contains(res.JSON, `"1": "Alice"`) {
		t.Errorf("JSON 应含字段 1 的文本值，实际 = %s", res.JSON)
	}
	if !strings.Contains(res.JSON, `"2": 1`) {
		t.Errorf("JSON 应含字段 2 的整数值，实际 = %s", res.JSON)
	}
}

func TestProtobufDecodeRejectsUnknownMode(t *testing.T) {
	_, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:  "yaml",
		Input: ByteInput{Data: "08 01", Encoding: EncodingHex},
	})
	assertToolErrCode(t, err, toolerr.CodeUnsupportedMode)
}

func TestProtobufDecodeRequiresSchemaInSchemaMode(t *testing.T) {
	_, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:  ProtoModeSchema,
		Input: ByteInput{Data: "08 01", Encoding: EncodingHex},
	})
	assertToolErrCode(t, err, toolerr.CodeMissingField)
}

func TestProtobufDecodeSchemaMode(t *testing.T) {
	const schema = `
syntax = "proto3";
message Person {
  string name = 1;
  int32 id = 2;
}
`
	// 字段 1 = "Alice"，字段 2 = 7。
	raw := protoHex(t, "0a 05 41 6c 69 63 65 10 07")

	res, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:        ProtoModeSchema,
		Input:       ByteInput{Data: hex.EncodeToString(raw), Encoding: EncodingHex},
		Schema:      schema,
		MessageType: "Person",
	})
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	out := res.Data.(ProtobufDecodeResult)
	if out.MessageType != "Person" {
		t.Errorf("消息名 = %q，期望 Person", out.MessageType)
	}
	if len(out.Fields) != 2 {
		t.Fatalf("字段数 = %d，期望 2", len(out.Fields))
	}
	if out.Fields[0].Name != "name" {
		t.Errorf("字段名 = %q，期望 name", out.Fields[0].Name)
	}
	if out.Fields[0].DeclaredType != "string" {
		t.Errorf("声明类型 = %q，期望 string", out.Fields[0].DeclaredType)
	}
	if out.Fields[1].Value != "7" {
		t.Errorf("值 = %q，期望 7", out.Fields[1].Value)
	}
	if !strings.Contains(out.JSON, `"name": "Alice"`) {
		t.Errorf("JSON 应含 name 字段，实际 = %s", out.JSON)
	}
}

func TestProtobufDecodeSchemaModeImportsWellKnown(t *testing.T) {
	const schema = `
syntax = "proto3";
import "google/protobuf/timestamp.proto";
message Event {
  google.protobuf.Timestamp at = 1;
}
`
	res, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:        ProtoModeSchema,
		Input:       ByteInput{Data: "0a 00", Encoding: EncodingHex},
		Schema:      schema,
		MessageType: "Event",
	})
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	out := res.Data.(ProtobufDecodeResult)
	if out.MessageType != "Event" {
		t.Errorf("消息名 = %q，期望 Event", out.MessageType)
	}
}

func TestProtobufDecodeSchemaModeRejectsUnknownMessage(t *testing.T) {
	const schema = `
syntax = "proto3";
message Person { string name = 1; }
`
	_, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:        ProtoModeSchema,
		Input:       ByteInput{Data: "0a 00", Encoding: EncodingHex},
		Schema:      schema,
		MessageType: "Missing",
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}

func TestProtobufDecodeSchemaModeRejectsBrokenSchema(t *testing.T) {
	_, err := ProtobufDecode(context.Background(), &ProtobufDecodeRequest{
		Mode:        ProtoModeSchema,
		Input:       ByteInput{Data: "0a 00", Encoding: EncodingHex},
		Schema:      "syntax = \"proto3\"; message {",
		MessageType: "X",
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidProtobuf)
}

func TestProtobufEncodeWireVarint(t *testing.T) {
	res, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode: ProtoModeWire,
		Fields: []ProtoEncodeField{
			{Field: 1, Type: "int32", Value: "150"},
		},
		OutputEncoding: EncodingHex,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	out := res.Data.(ProtobufEncodeResult)
	if out.Output != "089601" {
		t.Errorf("输出 = %q，期望 089601", out.Output)
	}
	if out.ByteLength != 3 {
		t.Errorf("字节数 = %d，期望 3", out.ByteLength)
	}
}

func TestProtobufEncodeWireRoundTrip(t *testing.T) {
	res, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode: ProtoModeWire,
		Fields: []ProtoEncodeField{
			{Field: 1, Type: "string", Value: "Alice"},
			{Field: 2, Type: "int32", Value: "7"},
		},
		OutputEncoding: EncodingHex,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	out := res.Data.(ProtobufEncodeResult)

	decoded := protoDecodeHex(t, out.Output)
	if len(decoded.Fields) != 2 {
		t.Fatalf("字段数 = %d，期望 2", len(decoded.Fields))
	}
	if decoded.Fields[0].Value != "Alice" {
		t.Errorf("字段 1 = %q，期望 Alice", decoded.Fields[0].Value)
	}
	if decoded.Fields[1].Value != "7" {
		t.Errorf("字段 2 = %q，期望 7", decoded.Fields[1].Value)
	}
}

func TestProtobufEncodeWireNestedMessage(t *testing.T) {
	res, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode: ProtoModeWire,
		Fields: []ProtoEncodeField{
			{
				Field: 1,
				Type:  "message",
				Fields: []ProtoEncodeField{
					{Field: 1, Type: "int32", Value: "1"},
				},
			},
		},
		OutputEncoding: EncodingHex,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	out := res.Data.(ProtobufEncodeResult)
	if out.Output != "0a020801" {
		t.Errorf("输出 = %q，期望 0a020801", out.Output)
	}
}

func TestProtobufEncodeWireBytesEncoding(t *testing.T) {
	res, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode: ProtoModeWire,
		Fields: []ProtoEncodeField{
			{Field: 1, Type: "bytes", Value: "48656c6c6f", Encoding: EncodingHex},
		},
		OutputEncoding: EncodingHex,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	out := res.Data.(ProtobufEncodeResult)
	if out.Output != "0a0548656c6c6f" {
		t.Errorf("输出 = %q，期望 0a0548656c6c6f", out.Output)
	}
}

func TestProtobufEncodeRejectsInvalidFieldNumber(t *testing.T) {
	_, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode:   ProtoModeWire,
		Fields: []ProtoEncodeField{{Field: 0, Type: "int32", Value: "1"}},
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}

func TestProtobufEncodeRejectsInvalidType(t *testing.T) {
	_, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode:   ProtoModeWire,
		Fields: []ProtoEncodeField{{Field: 1, Type: "int7", Value: "1"}},
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}

func TestProtobufEncodeRejectsMissingFields(t *testing.T) {
	_, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode: ProtoModeWire,
	})
	assertToolErrCode(t, err, toolerr.CodeMissingField)
}

func TestProtobufEncodeSchemaMode(t *testing.T) {
	const schema = `
syntax = "proto3";
message Person {
  string name = 1;
  int32 id = 2;
}
`
	res, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode:           ProtoModeSchema,
		Schema:         schema,
		MessageType:    "Person",
		Message:        `{"name": "Alice", "id": 7}`,
		OutputEncoding: EncodingHex,
	})
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}
	out := res.Data.(ProtobufEncodeResult)
	if out.Output != "0a05416c6963651007" {
		t.Errorf("输出 = %q，期望 0a05416c6963651007", out.Output)
	}
	if out.FieldCount != 2 {
		t.Errorf("字段数 = %d，期望 2", out.FieldCount)
	}
}

func TestProtobufEncodeSchemaModeRejectsMismatchedJSON(t *testing.T) {
	const schema = `
syntax = "proto3";
message Person { string name = 1; }
`
	_, err := ProtobufEncode(context.Background(), &ProtobufEncodeRequest{
		Mode:        ProtoModeSchema,
		Schema:      schema,
		MessageType: "Person",
		Message:     `{"nope": 1}`,
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidProtobuf)
}
