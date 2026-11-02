package encoding

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// protoSchemaFileName 是编译 .proto 时使用的虚拟文件名。
const protoSchemaFileName = "schema.proto"

// protoMaxMessageNames 是错误里列出的可用消息名上限。
const protoMaxMessageNames = 20

// protoSchema 是编译后的 .proto 描述。
type protoSchema struct {
	// messages 是消息全名 → 描述符（含嵌套消息）。
	messages map[string]protoreflect.MessageDescriptor
	// topLevel 是顶层消息，按声明顺序。
	topLevel []protoreflect.MessageDescriptor
}

// compileProtoSchema 编译单文件 .proto 描述。
//
// 只支持**单文件** schema：Google 标准类型（`google/protobuf/*.proto`）可用，
// 其余 import 不会被解析——需要时请自行把定义合并进同一文件。
func compileProtoSchema(schema string) (*protoSchema, error) {
	if strings.TrimSpace(schema) == "" {
		return nil, tools.MissingField("schema")
	}

	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			Accessor: protocompile.SourceAccessorFromMap(map[string]string{
				protoSchemaFileName: schema,
			}),
		}),
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	files, err := compiler.Compile(context.Background(), protoSchemaFileName)
	if err != nil {
		return nil, invalidProtobuf("schema 编译失败", protoSchemaErrorParams(err))
	}
	if len(files) == 0 {
		return nil, invalidProtobuf("schema 编译失败", nil)
	}

	out := &protoSchema{messages: map[string]protoreflect.MessageDescriptor{}}
	collectProtoMessages(files[0].Messages(), out)
	return out, nil
}

// collectProtoMessages 收集消息（含嵌套消息）。
func collectProtoMessages(messages protoreflect.MessageDescriptors, out *protoSchema) {
	for i := 0; i < messages.Len(); i++ {
		md := messages.Get(i)
		out.messages[string(md.FullName())] = md
		out.topLevel = append(out.topLevel, md)
		collectProtoMessages(md.Messages(), out)
	}
}

// protoSchemaErrorParams 从编译错误里提取位置信息。
//
// 只取行列号，**不回显** schema 原文。
func protoSchemaErrorParams(err error) map[string]any {
	params := map[string]any{}
	var withPos reporter.ErrorWithPos
	if ok := asProtoErrorWithPos(err, &withPos); ok {
		pos := withPos.GetPosition()
		params["line"] = pos.Line
		params["column"] = pos.Col
	}
	return params
}

// asProtoErrorWithPos 把错误解包成带位置的编译错误。
func asProtoErrorWithPos(err error, target *reporter.ErrorWithPos) bool {
	for err != nil {
		if withPos, ok := err.(reporter.ErrorWithPos); ok {
			*target = withPos
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

// findMessage 按名字查找消息描述。
//
// name 可以是全名，也可以是简名；留空时要求 schema 只有一个顶层消息。
func (s *protoSchema) findMessage(name string) (protoreflect.MessageDescriptor, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		if len(s.topLevel) == 1 {
			return s.topLevel[0], nil
		}
		return nil, tools.InvalidField(
			"messageType",
			s.messageNames(),
			"schema 中存在多个顶层消息",
		)
	}

	if md, ok := s.messages[name]; ok {
		return md, nil
	}

	var found []protoreflect.MessageDescriptor
	for _, md := range s.messages {
		if string(md.Name()) == name {
			found = append(found, md)
		}
	}
	switch len(found) {
	case 1:
		return found[0], nil
	case 0:
		return nil, tools.InvalidField("messageType", s.messageNames(), "未在 schema 中定义")
	default:
		return nil, tools.InvalidField("messageType", s.messageNames(), "消息名不唯一")
	}
}

// messageNames 返回可用的消息名列表（最多 20 个）。
func (s *protoSchema) messageNames() []string {
	names := make([]string, 0, len(s.messages))
	for name := range s.messages {
		names = append(names, name)
	}
	sortProtoStrings(names)
	if len(names) > protoMaxMessageNames {
		names = names[:protoMaxMessageNames]
	}
	return names
}

// sortProtoStrings 对字符串切片做字典序排序。
func sortProtoStrings(list []string) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j] < list[j-1]; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

// protoSchemaDecode 按 .proto 描述解析字节流。
//
// 返回字段树、protojson 文本与未知字段数。
func protoSchemaDecode(
	data []byte, schema *protoSchema, md protoreflect.MessageDescriptor, opts protoWireOptions,
) ([]ProtoField, string, int, error) {
	msg := dynamicpb.NewMessage(md)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, "", 0, invalidProtobuf("消息解析失败", nil)
	}

	raw, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(msg)
	if err != nil {
		return nil, "", 0, invalidProtobuf("消息转 JSON 失败", nil)
	}

	// protojson 的缩进与冒号间距与本仓库其它工具不一致，这里统一成两空格缩进。
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return nil, "", 0, invalidProtobuf("消息转 JSON 失败", nil)
	}

	fields, err := protoSchemaFields(msg, "root", 0, opts)
	if err != nil {
		return nil, "", 0, err
	}

	unknown := 0
	if unknownRaw := msg.GetUnknown(); len(unknownRaw) > 0 {
		parsed, err := parseProtoWire(unknownRaw, opts)
		if err == nil {
			unknown = parsed.fieldCount
		} else {
			unknown = 1
		}
	}

	return fields, pretty.String(), unknown, nil
}

// protoSchemaFields 按描述符遍历消息里已设置的字段。
func protoSchemaFields(
	msg protoreflect.Message, path string, depth int, opts protoWireOptions,
) ([]ProtoField, error) {
	descriptors := msg.Descriptor().Fields()
	out := make([]ProtoField, 0, descriptors.Len())

	for i := 0; i < descriptors.Len(); i++ {
		fd := descriptors.Get(i)
		if fd.IsMap() {
			continue
		}
		if !msg.Has(fd) {
			continue
		}

		value := msg.Get(fd)
		if fd.IsList() {
			list := value.List()
			for j := 0; j < list.Len(); j++ {
				field, err := protoSchemaOneField(
					fd, list.Get(j), path+"/"+strconv.Itoa(int(fd.Number()))+"["+strconv.Itoa(j)+"]",
					depth, opts,
				)
				if err != nil {
					return nil, err
				}
				out = append(out, field)
			}
			continue
		}

		field, err := protoSchemaOneField(
			fd, value, path+"/"+strconv.Itoa(int(fd.Number())), depth, opts,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, field)
	}

	return out, nil
}

// protoSchemaOneField 构造单个字段条目。
func protoSchemaOneField(
	fd protoreflect.FieldDescriptor, value protoreflect.Value, path string, depth int, opts protoWireOptions,
) (ProtoField, error) {
	field := ProtoField{
		Path:         path,
		FieldNumber:  int(fd.Number()),
		Name:         string(fd.Name()),
		WireType:     protoSchemaWireType(fd),
		DeclaredType: protoSchemaDeclaredType(fd),
	}

	switch fd.Kind() {
	case protoreflect.BoolKind:
		field.Value = strconv.FormatBool(value.Bool())

	case protoreflect.EnumKind:
		number := value.Enum()
		field.Value = strconv.FormatInt(int64(number), 10)
		if enumValue := fd.Enum().Values().ByNumber(number); enumValue != nil {
			field.Value += " (" + string(enumValue.Name()) + ")"
		}

	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		field.Value = strconv.FormatInt(value.Int(), 10)

	case protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		field.Value = strconv.FormatUint(value.Uint(), 10)

	case protoreflect.FloatKind:
		f := value.Float()
		field.Value = strconv.FormatFloat(f, 'g', -1, 32)
		field.Float = &f

	case protoreflect.DoubleKind:
		f := value.Float()
		field.Value = strconv.FormatFloat(f, 'g', -1, 64)
		field.Float = &f

	case protoreflect.StringKind:
		field.Text = true
		field.Value = value.String()

	case protoreflect.BytesKind:
		raw := value.Bytes()
		field.raw = raw
		field.Length = len(raw)
		field.Hex = protoHexPreview(raw)

	case protoreflect.MessageKind, protoreflect.GroupKind:
		sub := value.Message()
		if depth+1 <= opts.maxDepth {
			children, err := protoSchemaFields(sub, path, depth+1, opts)
			if err != nil {
				return field, err
			}
			field.Message = true
			field.Children = children
		}
	}

	return field, nil
}

// protoSchemaWireType 返回声明类型对应的线类型。
func protoSchemaWireType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.BoolKind, protoreflect.EnumKind,
		protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return ProtoWireVarint
	case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		return ProtoWireFixed32
	case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
		return ProtoWireFixed64
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return ProtoWireBytes
	default:
		return ProtoWireBytes
	}
}

// protoSchemaDeclaredType 返回声明类型名。
func protoSchemaDeclaredType(fd protoreflect.FieldDescriptor) string {
	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return string(fd.Message().FullName())
	default:
		return fd.Kind().String()
	}
}

// protoSchemaEncode 按 .proto 描述把 protojson 文本编码成字节流。
//
// 同时返回已赋值字段的总数（含嵌套字段）。
func protoSchemaEncode(
	message string, schema *protoSchema, md protoreflect.MessageDescriptor, opts protoWireOptions,
) ([]byte, int, error) {
	text := strings.TrimSpace(message)
	if text == "" {
		return nil, 0, tools.MissingField("message")
	}

	msg := dynamicpb.NewMessage(md)
	if err := protojson.Unmarshal([]byte(text), msg); err != nil {
		return nil, 0, invalidProtobuf("protojson 与描述不匹配", nil)
	}

	// 用确定性序列化：字段按**字段号**排序，输出稳定可复现。
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return nil, 0, invalidProtobuf("消息编码失败", nil)
	}

	fields, err := protoSchemaFields(msg, "root", 0, opts)
	if err != nil {
		return nil, 0, err
	}
	return raw, protoCountFields(fields), nil
}
