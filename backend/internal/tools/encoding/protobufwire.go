package encoding

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// protoWireOptions 控制线格式解析。
type protoWireOptions struct {
	// maxDepth 是嵌套解析深度上限。
	maxDepth int
	// maxFields 是字段总数上限；超出后截断并置 truncated。
	maxFields int
	// inferStrings 是否**优先**把长度分隔字段识别为 UTF-8 文本。
	//
	// 为假时先尝试按嵌套消息解释，失败才回落文本。
	inferStrings bool
}

// protoWireResult 是线格式解析结果。
type protoWireResult struct {
	// fields 是顶层字段。
	fields []ProtoField
	// fieldCount 是解析出的字段总数（含嵌套字段）。
	fieldCount int
	// depth 是实际达到的最大嵌套深度。
	depth int
	// truncated 表示结果因上限被截断。
	truncated bool
}

// protoParser 是线格式解析器。
type protoParser struct {
	// opts 是解析选项。
	opts protoWireOptions
	// count 是已解析字段数。
	count int
	// depth 是已到达的最大深度。
	depth int
	// truncated 表示结果被截断。
	truncated bool
}

// parseProtoWire 解析 protobuf 线格式字节流。
//
// 结构性错误（非法标签、截断的 varint、未闭合的分组）返回
// INVALID_PROTOBUF；超出上限则截断并置 truncated。
func parseProtoWire(data []byte, opts protoWireOptions) (*protoWireResult, error) {
	p := &protoParser{opts: opts}
	fields, err := p.parseMessage(data, 0, "root", 0)
	if err != nil {
		return nil, err
	}
	return &protoWireResult{
		fields:     fields,
		fieldCount: p.count,
		depth:      p.depth,
		truncated:  p.truncated,
	}, nil
}

// parseMessage 解析一条消息的字段序列。
//
// base 是 data[0] 在根字节流中的绝对偏移；path 是当前消息的路径前缀。
func (p *protoParser) parseMessage(data []byte, base int, path string, depth int) ([]ProtoField, error) {
	if depth > p.depth {
		p.depth = depth
	}

	var fields []ProtoField
	off := 0
	for off < len(data) {
		if p.count >= p.opts.maxFields {
			p.truncated = true
			return fields, nil
		}

		num, typ, n := protowire.ConsumeTag(data[off:])
		if n < 0 {
			return nil, invalidProtobuf("标签解析失败", map[string]any{"offset": base + off})
		}
		fieldOffset := base + off
		off += n

		field, consumed, err := p.parseField(
			data[off:], fieldOffset, num, typ, path+"/"+strconv.Itoa(int(num)), depth,
		)
		if err != nil {
			return nil, err
		}
		off += consumed
		p.count++
		fields = append(fields, field)
	}
	return fields, nil
}

// parseField 解析一个字段的值，返回字段、已消费的字节数与错误。
func (p *protoParser) parseField(
	data []byte, offset int, num protowire.Number, typ protowire.Type, path string, depth int,
) (ProtoField, int, error) {
	field := ProtoField{
		Path:        path,
		Offset:      offset,
		FieldNumber: int(num),
		WireType:    protoWireTypeName(typ),
	}

	switch typ {
	case protowire.VarintType:
		v, n := protowire.ConsumeVarint(data)
		if n < 0 {
			return field, 0, invalidProtobuf("变长整数解析失败", map[string]any{"offset": offset})
		}
		field.Value = strconv.FormatUint(v, 10)
		return field, n, nil

	case protowire.Fixed32Type:
		v, n := protowire.ConsumeFixed32(data)
		if n < 0 {
			return field, 0, invalidProtobuf("32 位定长解析失败", map[string]any{"offset": offset})
		}
		field.Value = strconv.FormatUint(uint64(v), 10)
		f := float64(math.Float32frombits(v))
		field.Float = &f
		return field, n, nil

	case protowire.Fixed64Type:
		v, n := protowire.ConsumeFixed64(data)
		if n < 0 {
			return field, 0, invalidProtobuf("64 位定长解析失败", map[string]any{"offset": offset})
		}
		field.Value = strconv.FormatUint(v, 10)
		f := math.Float64frombits(v)
		field.Float = &f
		return field, n, nil

	case protowire.BytesType:
		v, n := protowire.ConsumeBytes(data)
		if n < 0 {
			return field, 0, invalidProtobuf("长度分隔字段解析失败", map[string]any{"offset": offset})
		}
		field.Length = len(v)
		field.raw = v
		p.classifyBytes(&field, v, offset+n-len(v), path, depth)
		return field, n, nil

	case protowire.StartGroupType:
		children, consumed, err := p.parseGroup(data, offset, num, path, depth)
		if err != nil {
			return field, 0, err
		}
		field.WireType = ProtoWireGroup
		field.Message = true
		field.Length = consumed
		field.Children = children
		return field, consumed, nil

	case protowire.EndGroupType:
		return field, 0, invalidProtobuf("出现未配对的分组结束标记", map[string]any{"offset": offset})

	default:
		return field, 0, invalidProtobuf("不支持的线类型", map[string]any{"offset": offset})
	}
}

// parseGroup 解析分组字段的内容，直到同号分组结束标记。
func (p *protoParser) parseGroup(
	data []byte, base int, num protowire.Number, path string, depth int,
) ([]ProtoField, int, error) {
	var children []ProtoField
	off := 0
	for {
		if off >= len(data) {
			return nil, 0, invalidProtobuf("分组未闭合", map[string]any{"offset": base})
		}
		num2, typ2, n := protowire.ConsumeTag(data[off:])
		if n < 0 {
			return nil, 0, invalidProtobuf("标签解析失败", map[string]any{"offset": base + off})
		}
		if typ2 == protowire.EndGroupType {
			if num2 != num {
				return nil, 0, invalidProtobuf("分组结束标记不匹配", map[string]any{"offset": base + off})
			}
			return children, off + n, nil
		}

		fieldOffset := base + off
		off += n
		child, consumed, err := p.parseField(
			data[off:], fieldOffset, num2, typ2, path+"/"+strconv.Itoa(int(num2)), depth+1,
		)
		if err != nil {
			return nil, 0, err
		}
		off += consumed
		p.count++
		children = append(children, child)
	}
}

// classifyBytes 判断长度分隔字段的语义：嵌套消息、UTF-8 文本还是二进制。
func (p *protoParser) classifyBytes(field *ProtoField, v []byte, valueOffset int, path string, depth int) {
	if len(v) == 0 {
		field.Value = ""
		return
	}

	if p.opts.inferStrings && protoPrintableUTF8(v) {
		field.Text = true
		field.Value = string(v)
		return
	}

	if depth+1 <= p.opts.maxDepth {
		savedCount, savedDepth, savedTruncated := p.count, p.depth, p.truncated
		children, err := p.parseMessage(v, valueOffset, path, depth+1)
		if err == nil && len(children) > 0 {
			field.Message = true
			field.Children = children
			return
		}
		p.count, p.depth, p.truncated = savedCount, savedDepth, savedTruncated
	}

	if protoValidUTF8(v) {
		field.Text = true
		field.Value = string(v)
		return
	}

	field.Hex = protoHexPreview(v)
}

// protoWireTypeName 返回线类型的展示名。
func protoWireTypeName(typ protowire.Type) string {
	switch typ {
	case protowire.VarintType:
		return ProtoWireVarint
	case protowire.Fixed32Type:
		return ProtoWireFixed32
	case protowire.Fixed64Type:
		return ProtoWireFixed64
	case protowire.BytesType:
		return ProtoWireBytes
	case protowire.StartGroupType, protowire.EndGroupType:
		return ProtoWireGroup
	default:
		return "unknown"
	}
}

// protoPrintableUTF8 判断字节是否是可打印的 UTF-8 文本。
func protoPrintableUTF8(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}
	for _, r := range string(b) {
		switch r {
		case '\t', '\n', '\r':
			continue
		}
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// protoValidUTF8 判断字节是否是合法 UTF-8。
func protoValidUTF8(b []byte) bool {
	return utf8.Valid(b)
}

// protoHexPreview 返回二进制值的十六进制预览。
func protoHexPreview(b []byte) string {
	if len(b) > protoHexPreviewBytes {
		b = b[:protoHexPreviewBytes]
	}
	return hex.EncodeToString(b)
}

// protoWireJSON 把字段树转成 JSON 文本。
//
// 键是字段号；重复出现的字段号归并为数组。
// 64 位以上的整数按字符串输出，避免 JSON 数字精度丢失。
func protoWireJSON(fields []ProtoField) string {
	obj := protoFieldsToObject(fields)
	raw, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(raw)
}

// protoFieldsToObject 把字段列表转成 JSON 对象。
func protoFieldsToObject(fields []ProtoField) map[string]any {
	obj := make(map[string]any, len(fields))
	for _, f := range fields {
		key := strconv.Itoa(f.FieldNumber)
		value := protoFieldValue(f)
		existing, ok := obj[key]
		if !ok {
			obj[key] = value
			continue
		}
		if arr, isArr := existing.([]any); isArr {
			obj[key] = append(arr, value)
			continue
		}
		obj[key] = []any{existing, value}
	}
	return obj
}

// protoFieldValue 返回单个字段的 JSON 值。
func protoFieldValue(f ProtoField) any {
	switch {
	case f.Message:
		return protoFieldsToObject(f.Children)
	case f.Text:
		return f.Value
	case f.Hex != "":
		return base64.StdEncoding.EncodeToString(f.raw)
	case f.WireType == ProtoWireVarint || f.WireType == ProtoWireFixed32 || f.WireType == ProtoWireFixed64:
		return protoJSONNumber(f.Value)
	default:
		return ""
	}
}

// protoJSONNumber 把十进制整数字符串转成 JSON 值。
//
// 超过 JavaScript 安全整数范围（2^53-1）时按字符串输出。
func protoJSONNumber(s string) any {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return s
	}
	if v > 1<<53-1 {
		return s
	}
	return v
}

// ProtoEncodeField 是待编码的一个字段规格。
type ProtoEncodeField struct {
	// Field 是字段号（1–536870911）。
	Field int `json:"field"`
	// Type 是语义类型名（int32 / string / message 等）。
	Type string `json:"type"`
	// Value 是值的文本表示；**一律用字符串**，避免 64 位整数在前端丢精度。
	Value string `json:"value"`
	// Encoding 是 bytes 类型值的编码（utf8 / hex / base64 / base64url）。
	Encoding string `json:"encoding"`
	// Fields 是 message 类型的子字段规格。
	Fields []ProtoEncodeField `json:"fields"`
}

// protoEncodeTypes 返回可用的语义类型名（顺序稳定）。
func protoEncodeTypes() []string {
	return []string{
		"int32", "int64", "uint32", "uint64", "sint32", "sint64",
		"bool", "enum", "fixed32", "sfixed32", "float",
		"fixed64", "sfixed64", "double", "string", "bytes", "message",
	}
}

// protoWireTypeOf 返回语义类型对应的线类型。
func protoWireTypeOf(t string) (protowire.Type, bool) {
	switch t {
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool", "enum":
		return protowire.VarintType, true
	case "fixed32", "sfixed32", "float":
		return protowire.Fixed32Type, true
	case "fixed64", "sfixed64", "double":
		return protowire.Fixed64Type, true
	case "string", "bytes", "message":
		return protowire.BytesType, true
	default:
		return 0, false
	}
}

// protoEncodeWire 按字段规格编码成 protobuf 线格式字节流。
//
// 返回字节流与写入的字段总数。
func protoEncodeWire(fields []ProtoEncodeField) ([]byte, int, error) {
	return protoAppendFields(nil, fields)
}

// protoAppendFields 按顺序编码一组字段。
func protoAppendFields(buf []byte, fields []ProtoEncodeField) ([]byte, int, error) {
	count := 0
	for _, f := range fields {
		out, n, err := protoAppendField(buf, f)
		if err != nil {
			return nil, 0, err
		}
		buf = out
		count += n
	}
	return buf, count, nil
}

// protoAppendField 编码单个字段。
func protoAppendField(buf []byte, f ProtoEncodeField) ([]byte, int, error) {
	if f.Field < ProtoMinFieldNumber || f.Field > ProtoMaxFieldNumber {
		return nil, 0, tools.InvalidField(
			"fields.field",
			fmt.Sprintf("%d-%d", ProtoMinFieldNumber, ProtoMaxFieldNumber),
			f.Field,
		)
	}

	typ, ok := protoWireTypeOf(f.Type)
	if !ok {
		return nil, 0, tools.InvalidField("fields.type", protoEncodeTypes(), f.Type)
	}
	num := protowire.Number(f.Field)
	value := strings.TrimSpace(f.Value)

	switch f.Type {
	case "int32":
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "32 位有符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, uint64(v))
	case "int64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "64 位有符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, uint64(v))
	case "uint32":
		v, err := protoParseUint(value, 32)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "32 位无符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, v)
	case "uint64":
		v, err := protoParseUint(value, 64)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "64 位无符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, v)
	case "sint32":
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "32 位有符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, protowire.EncodeZigZag(v))
	case "sint64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "64 位有符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, protowire.EncodeZigZag(v))
	case "bool":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "布尔值（true / false）", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, protowire.EncodeBool(b))
	case "enum":
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "32 位有符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendVarint(buf, uint64(v))
	case "fixed32":
		v, err := protoParseUint(value, 32)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "32 位无符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendFixed32(buf, uint32(v))
	case "sfixed32":
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "32 位有符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendFixed32(buf, uint32(v))
	case "float":
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "32 位浮点数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendFixed32(buf, math.Float32bits(float32(v)))
	case "fixed64":
		v, err := protoParseUint(value, 64)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "64 位无符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendFixed64(buf, v)
	case "sfixed64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "64 位有符号整数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendFixed64(buf, uint64(v))
	case "double":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, 0, tools.InvalidField("fields.value", "64 位浮点数", value)
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendFixed64(buf, math.Float64bits(v))
	case "string":
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendBytes(buf, []byte(f.Value))
	case "bytes":
		raw, err := codecsvc.Decode(f.Value, f.Encoding)
		if err != nil {
			return nil, 0, err
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendBytes(buf, raw)
	case "message":
		sub, n, err := protoAppendFields(nil, f.Fields)
		if err != nil {
			return nil, 0, err
		}
		buf = protowire.AppendTag(buf, num, typ)
		buf = protowire.AppendBytes(buf, sub)
		return buf, n + 1, nil
	default:
		return nil, 0, tools.InvalidField("fields.type", protoEncodeTypes(), f.Type)
	}

	return buf, 1, nil
}

// protoParseUint 解析无符号整数，支持十进制与 0x 前缀的十六进制。
func protoParseUint(s string, bits int) (uint64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, bits)
	}
	return strconv.ParseUint(s, 10, bits)
}
