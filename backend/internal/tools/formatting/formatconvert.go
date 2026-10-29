package formatting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 结构化数据格式。
const (
	// DataFormatJSON 是 JSON。
	DataFormatJSON = "json"
	// DataFormatYAML 是 YAML。
	DataFormatYAML = "yaml"
	// DataFormatTOML 是 TOML。
	DataFormatTOML = "toml"
	// DataFormatXML 是 XML。
	DataFormatXML = "xml"
)

// DataFormats 返回受支持的格式（顺序稳定）。
func DataFormats() []string {
	return []string{DataFormatJSON, DataFormatYAML, DataFormatTOML, DataFormatXML}
}

// FormatConvertRequest 是 format-convert 工具的请求。
type FormatConvertRequest struct {
	// Input 是待转换的文本。
	Input TextInput `json:"input"`
	// From 是源格式。
	From string `json:"from"`
	// To 是目标格式。
	To string `json:"to"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
	// Pretty 是否美化输出；默认 true。
	Pretty *bool `json:"pretty"`
}

// FormatConvertResult 是 format-convert 工具的响应。
type FormatConvertResult struct {
	// Output 是转换结果。
	Output string `json:"output"`
	// From 是源格式。
	From string `json:"from"`
	// To 是目标格式。
	To string `json:"to"`
	// Lossy 为真表示转换过程中有信息丢失。
	Lossy bool `json:"lossy,omitempty"`
	// LossyReason 说明丢失的原因。
	LossyReason string `json:"lossyReason,omitempty"`
}

// FormatConvert 在 JSON / YAML / TOML / XML 之间转换结构。
//
// 转换经由统一的中间表示，因此四种格式两两之间都能互转。
//
// 中间表示**保留键序**（JSON 的键序是用户数据，不能打乱），
// 数字也保留原始字面量（避免大整数在 float64 上丢精度）。
func FormatConvert(req *FormatConvertRequest) (FormatConvertResult, error) {
	if err := req.Validate(); err != nil {
		return FormatConvertResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return FormatConvertResult{}, err
	}

	value, lossy, reason, err := parseStructured(req.From, text)
	if err != nil {
		return FormatConvertResult{}, err
	}

	out, err := renderStructured(req.To, value, indentWidth(req.Indent), req.pretty())
	if err != nil {
		return FormatConvertResult{}, err
	}

	return FormatConvertResult{
		Output:      out,
		From:        req.From,
		To:          req.To,
		Lossy:       lossy,
		LossyReason: reason,
	}, nil
}

// Validate 做字段级校验。
func (r *FormatConvertRequest) Validate() error {
	switch r.From {
	case DataFormatJSON, DataFormatYAML, DataFormatTOML, DataFormatXML:
	default:
		return tools.InvalidField("from", DataFormats(), r.From)
	}
	switch r.To {
	case DataFormatJSON, DataFormatYAML, DataFormatTOML, DataFormatXML:
	default:
		return tools.InvalidField("to", DataFormats(), r.To)
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	return r.Input.Validate("input")
}

// pretty 返回是否美化输出；缺省为 true。
func (r *FormatConvertRequest) pretty() bool {
	if r.Pretty == nil {
		return true
	}
	return *r.Pretty
}

// dataKind 是中间表示的节点类型。
type dataKind int

const (
	// dataObject 是对象。
	dataObject dataKind = iota
	// dataArray 是数组。
	dataArray
	// dataString 是字符串。
	dataString
	// dataNumber 是数字（保留原始字面量）。
	dataNumber
	// dataBool 是布尔值。
	dataBool
	// dataNull 是空值。
	dataNull
)

// dataMember 是对象的一个键值对。
type dataMember struct {
	// Key 是键名。
	Key string
	// Value 是值。
	Value *dataValue
}

// dataValue 是中间表示的一个节点。
type dataValue struct {
	// kind 是节点类型。
	kind dataKind
	// members 是对象的成员（**有序**）。
	members []dataMember
	// items 是数组的元素。
	items []*dataValue
	// text 是字符串、字面量与布尔的原始文本。
	text string
}

// parseStructured 把文本解析为中间表示。
func parseStructured(format, text string) (*dataValue, bool, string, error) {
	switch format {
	case DataFormatJSON:
		return parseStructuredJSON(text)
	case DataFormatYAML:
		return parseStructuredYAML(text)
	case DataFormatTOML:
		return parseStructuredTOML(text)
	case DataFormatXML:
		return parseStructuredXML(text)
	default:
		return nil, false, "", tools.InvalidField("from", DataFormats(), format)
	}
}

// parseStructuredJSON 解析 JSON。
func parseStructuredJSON(text string) (*dataValue, bool, string, error) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()

	var raw any
	if err := dec.Decode(&raw); err != nil {
		se := jsonSyntaxError(text, err)
		return nil, false, "", SyntaxErr(toolerr.CodeInvalidJSON, se)
	}

	value := fromJSONValue(raw)
	return value, false, "", nil
}

// fromJSONValue 把标准库的值转成中间表示。
func fromJSONValue(raw any) *dataValue {
	switch v := raw.(type) {
	case map[string]any:
		out := &dataValue{kind: dataObject}
		for _, key := range sortedKeys(v) {
			out.members = append(out.members, dataMember{Key: key, Value: fromJSONValue(v[key])})
		}
		return out
	case []any:
		out := &dataValue{kind: dataArray}
		for _, item := range v {
			out.items = append(out.items, fromJSONValue(item))
		}
		return out
	case json.Number:
		return &dataValue{kind: dataNumber, text: v.String()}
	case string:
		return &dataValue{kind: dataString, text: v}
	case bool:
		return &dataValue{kind: dataBool, text: strconv.FormatBool(v)}
	case nil:
		return &dataValue{kind: dataNull}
	default:
		return &dataValue{kind: dataString, text: fmt.Sprintf("%v", v)}
	}
}

// sortedKeys 返回 map 的键（按字典序）。
//
// JSON 的键序在标准库里已丢失（`map[string]any`），
// 因此转换到保序格式时只能按字典序重建。
func sortedKeys(v map[string]any) []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

// parseStructuredYAML 解析 YAML。
func parseStructuredYAML(text string) (*dataValue, bool, string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(text), &node); err != nil {
		se := yamlSyntaxError(err)
		return nil, false, "", SyntaxErr(toolerr.CodeInvalidYAML, se)
	}

	if node.Kind == 0 {
		return &dataValue{kind: dataNull}, false, "", nil
	}
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return &dataValue{kind: dataNull}, false, "", nil
		}
		node = *node.Content[0]
	}

	value := fromYAMLNode(&node)
	return value, false, "", nil
}

// fromYAMLNode 把 YAML 节点转成中间表示。
func fromYAMLNode(node *yaml.Node) *dataValue {
	switch node.Kind {
	case yaml.MappingNode:
		out := &dataValue{kind: dataObject}
		for i := 0; i+1 < len(node.Content); i += 2 {
			out.members = append(out.members, dataMember{
				Key:   node.Content[i].Value,
				Value: fromYAMLNode(node.Content[i+1]),
			})
		}
		return out

	case yaml.SequenceNode:
		out := &dataValue{kind: dataArray}
		for _, item := range node.Content {
			out.items = append(out.items, fromYAMLNode(item))
		}
		return out

	case yaml.AliasNode:
		return fromYAMLNode(node.Alias)

	case yaml.ScalarNode:
		switch node.Tag {
		case "!!int", "!!float":
			return &dataValue{kind: dataNumber, text: node.Value}
		case "!!bool":
			return &dataValue{kind: dataBool, text: node.Value}
		case "!!null":
			return &dataValue{kind: dataNull}
		default:
			return &dataValue{kind: dataString, text: node.Value}
		}

	default:
		return &dataValue{kind: dataNull}
	}
}

// parseStructuredTOML 解析 TOML。
func parseStructuredTOML(text string) (*dataValue, bool, string, error) {
	var raw map[string]any
	if err := toml.Unmarshal([]byte(text), &raw); err != nil {
		se := tomlSyntaxError(err)
		return nil, false, "", SyntaxErr(toolerr.CodeInvalidTOML, se)
	}

	value := fromTOMLValue(raw)
	return value, false, "", nil
}

// fromTOMLValue 把 TOML 的值转成中间表示。
func fromTOMLValue(raw map[string]any) *dataValue {
	out := &dataValue{kind: dataObject}
	for _, key := range sortedKeys(raw) {
		out.members = append(out.members, dataMember{Key: key, Value: fromTOMLValueAny(raw[key])})
	}
	return out
}

// fromTOMLValueAny 把 TOML 的单个值转成中间表示。
func fromTOMLValueAny(raw any) *dataValue {
	switch v := raw.(type) {
	case map[string]any:
		return fromTOMLValue(v)
	case []any:
		out := &dataValue{kind: dataArray}
		for _, item := range v {
			out.items = append(out.items, fromTOMLValueAny(item))
		}
		return out
	case int64:
		return &dataValue{kind: dataNumber, text: strconv.FormatInt(v, 10)}
	case float64:
		return &dataValue{kind: dataNumber, text: strconv.FormatFloat(v, 'g', -1, 64)}
	case bool:
		return &dataValue{kind: dataBool, text: strconv.FormatBool(v)}
	case string:
		return &dataValue{kind: dataString, text: v}
	case nil:
		return &dataValue{kind: dataNull}
	default:
		return &dataValue{kind: dataString, text: fmt.Sprintf("%v", v)}
	}
}

// parseStructuredXML 解析 XML。
//
// XML **没有**「无序对象」的概念，根元素固定是对象，
// 重复的同名子元素被收进数组。
func parseStructuredXML(text string) (*dataValue, bool, string, error) {
	tokens, synErr := tokenizeXML(text)
	if synErr != nil {
		return nil, false, "", SyntaxErr(toolerr.CodeInvalidXML, *synErr)
	}

	root, children := xmlTree(tokens)
	if root == "" {
		return &dataValue{kind: dataNull}, false, "", nil
	}

	value := &dataValue{kind: dataObject}
	for _, child := range children {
		value.members = append(value.members, dataMember{Key: child.name, Value: child.value})
	}
	return value, false, "", nil
}

// renderStructured 把中间表示渲染为目标格式。
func renderStructured(format string, value *dataValue, indent int, pretty bool) (string, error) {
	switch format {
	case DataFormatJSON:
		return renderStructuredJSON(value, indent, pretty)
	case DataFormatYAML:
		return renderStructuredYAML(value, indent)
	case DataFormatTOML:
		return renderStructuredTOML(value)
	case DataFormatXML:
		return renderStructuredXML(value, indent, pretty)
	default:
		return "", tools.InvalidField("to", DataFormats(), format)
	}
}

// renderStructuredJSON 渲染成 JSON。
func renderStructuredJSON(value *dataValue, indent int, pretty bool) (string, error) {
	var buf bytes.Buffer
	if err := writeJSONValue(&buf, value, indent, pretty, 0); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// writeJSONValue 把中间表示的节点写成 JSON。
//
// depth 是当前嵌套层级；pretty 为假时全部输出成紧凑形式。
func writeJSONValue(buf *bytes.Buffer, value *dataValue, indent int, pretty bool, depth int) error {
	pad := indentString(indent)
	inner := strings.Repeat(pad, depth+1)
	outer := strings.Repeat(pad, depth)

	switch value.kind {
	case dataObject:
		if len(value.members) == 0 {
			buf.WriteString("{}")
			return nil
		}
		buf.WriteString("{")
		if pretty {
			buf.WriteString("\n")
		}
		for i, m := range value.members {
			buf.WriteString(inner)
			buf.WriteString(jsonQuote(m.Key))
			buf.WriteString(":")
			if pretty {
				buf.WriteString(" ")
			}
			if err := writeJSONValue(buf, m.Value, indent, pretty, depth+1); err != nil {
				return err
			}
			if i < len(value.members)-1 {
				buf.WriteString(",")
			}
			if pretty {
				buf.WriteString("\n")
			}
		}
		buf.WriteString(outer)
		buf.WriteString("}")
	case dataArray:
		if len(value.items) == 0 {
			buf.WriteString("[]")
			return nil
		}
		buf.WriteString("[")
		if pretty {
			buf.WriteString("\n")
		}
		for i, item := range value.items {
			buf.WriteString(inner)
			if err := writeJSONValue(buf, item, indent, pretty, depth+1); err != nil {
				return err
			}
			if i < len(value.items)-1 {
				buf.WriteString(",")
			}
			if pretty {
				buf.WriteString("\n")
			}
		}
		buf.WriteString(outer)
		buf.WriteString("]")
	case dataString:
		buf.WriteString(jsonQuote(value.text))
	case dataNumber:
		buf.WriteString(value.text)
	case dataBool:
		buf.WriteString(value.text)
	default:
		buf.WriteString("null")
	}
	return nil
}

// jsonQuote 把字符串编码为 JSON 字面量。
func jsonQuote(s string) string {
	encoded, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(encoded)
}

// renderStructuredYAML 渲染成 YAML。
func renderStructuredYAML(value *dataValue, indent int) (string, error) {
	node := toYAMLNode(value)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	if err := enc.Encode(node); err != nil {
		return "", err
	}
	_ = enc.Close()

	return strings.TrimRight(buf.String(), "\n"), nil
}

// toYAMLNode 把中间表示转成 YAML 节点。
func toYAMLNode(value *dataValue) *yaml.Node {
	switch value.kind {
	case dataObject:
		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, m := range value.members {
			node.Content = append(node.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: m.Key},
				toYAMLNode(m.Value),
			)
		}
		return node
	case dataArray:
		node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, item := range value.items {
			node.Content = append(node.Content, toYAMLNode(item))
		}
		return node
	case dataString:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value.text}
	case dataNumber:
		// 整数与浮点数用不同的 YAML tag，否则整数会被写成 `3.0`。
		if _, err := strconv.ParseInt(value.text, 10, 64); err == nil {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: value.text}
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: value.text}
	case dataBool:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: value.text}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}
	}
}

// renderStructuredTOML 渲染成 TOML。
//
// TOML 的**顶层必须是表**，因此顶层数组无法表达。
func renderStructuredTOML(value *dataValue) (string, error) {
	if value.kind != dataObject {
		return "", tools.InvalidField("to", []string{DataFormatTOML}, DataFormatTOML)
	}

	raw := toTOMLValue(value)

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(raw); err != nil {
		return "", err
	}

	return strings.TrimRight(buf.String(), "\n"), nil
}

// toTOMLValue 把中间表示转成 go-toml 可编码的值。
func toTOMLValue(value *dataValue) map[string]any {
	out := make(map[string]any, len(value.members))
	for _, m := range value.members {
		out[m.Key] = toTOMLValueAny(m.Value)
	}
	return out
}

// toTOMLValueAny 把中间表示的节点转成 go-toml 可编码的值。
func toTOMLValueAny(value *dataValue) any {
	switch value.kind {
	case dataObject:
		return toTOMLValue(value)
	case dataArray:
		out := make([]any, 0, len(value.items))
		for _, item := range value.items {
			out = append(out, toTOMLValueAny(item))
		}
		return out
	case dataString:
		return value.text
	case dataNumber:
		if i, err := strconv.ParseInt(value.text, 10, 64); err == nil {
			return i
		}
		f, err := strconv.ParseFloat(value.text, 64)
		if err != nil {
			return value.text
		}
		return f
	case dataBool:
		return value.text == "true"
	default:
		return nil
	}
}

// renderStructuredXML 渲染成 XML。
func renderStructuredXML(value *dataValue, indent int, pretty bool) (string, error) {
	if value.kind != dataObject {
		return "", tools.InvalidField("to", []string{DataFormatXML}, DataFormatXML)
	}
	if len(value.members) == 0 {
		return "", fmt.Errorf("xml requires a root element")
	}

	var buf bytes.Buffer
	writeXMLMembers(&buf, value, "root", indent, pretty, 0)
	return strings.TrimRight(buf.String(), "\n"), nil
}

// writeXMLMembers 把成员写成 XML 元素。
func writeXMLMembers(buf *bytes.Buffer, value *dataValue, name string, indent int, pretty bool, depth int) {
	pad := indentString(indent)
	prefix := strings.Repeat(pad, depth)

	if value.kind == dataObject {
		buf.WriteString(prefix + "<" + name + ">")
		if pretty {
			buf.WriteString("\n")
		}
		for _, m := range value.members {
			writeXMLMembers(buf, m.Value, m.Key, indent, pretty, depth+1)
		}
		if pretty {
			buf.WriteString(prefix)
		}
		buf.WriteString("</" + name + ">")
		if pretty {
			buf.WriteString("\n")
		}
		return
	}

	if value.kind == dataArray {
		for _, item := range value.items {
			writeXMLMembers(buf, item, name, indent, pretty, depth)
		}
		return
	}

	buf.WriteString(prefix + "<" + name + ">")
	buf.WriteString(escapeConvertXMLText(value.text))
	buf.WriteString("</" + name + ">")
	if pretty {
		buf.WriteString("\n")
	}
}

// escapeConvertXMLText 转义 XML 文本内容里的特殊字符。
func escapeConvertXMLText(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(s)
}

// sortStrings 按字典序排序字符串切片。
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
