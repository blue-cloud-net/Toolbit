package data

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 目标语言。
const (
	// LangTypeScript 是 TypeScript。
	LangTypeScript = "typescript"
	// LangGo 是 Go。
	LangGo = "go"
	// LangRust 是 Rust。
	LangRust = "rust"
	// LangJava 是 Java。
	LangJava = "java"
)

// JSONToTypesLanguage 返回受支持的目标语言（顺序稳定）。
func JSONToTypesLanguage() []string {
	return []string{LangTypeScript, LangGo, LangRust, LangJava}
}

// JSONToTypesRequest 是 json-to-types 工具的请求。
type JSONToTypesRequest struct {
	// Input 是 JSON 文本。
	Input string `json:"input"`
	// Language 是目标语言；默认 typescript。
	Language string `json:"language"`
	// RootName 是根类型名；默认 Root。
	RootName string `json:"rootName"`
}

// JSONToTypesResult 是 json-to-types 工具的响应。
type JSONToTypesResult struct {
	// Language 是目标语言。
	Language string `json:"language"`
	// Output 是生成的类型定义。
	Output string `json:"output"`
	// TypeCount 是生成的类型数量。
	TypeCount int `json:"typeCount"`
}

// maxJSONToTypesBytes 是输入 JSON 的字节上限。
const maxJSONToTypesBytes = 512 * 1024

// JSONToTypes 由 JSON 样本生成语言类型定义。
//
// 这是**结构推断**，不是完整模式推导：生成的类型反映样本里出现过的字段。
func JSONToTypes(_ context.Context, req *JSONToTypesRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	language := req.Language
	if language == "" {
		language = LangTypeScript
	}
	rootName := req.RootName
	if rootName == "" {
		rootName = "Root"
	}

	decoder := json.NewDecoder(strings.NewReader(req.Input))
	var value any
	if err := decoder.Decode(&value); err != nil {
		return tools.Result{}, tools.InvalidField("input", "valid JSON", "parse error")
	}

	gen := &typeGenerator{
		language: language,
		types:    map[string]string{},
		order:    []string{},
	}

	rootType := gen.infer(rootName, value)

	lines := []string{}
	switch language {
	case LangGo:
		lines = append(lines, "package main", "")
		for _, name := range gen.order {
			lines = append(lines, gen.types[name])
			lines = append(lines, "")
		}
		lines = append(lines, "var _ "+rootType)
	case LangRust:
		for _, name := range gen.order {
			lines = append(lines, gen.types[name])
			lines = append(lines, "")
		}
		lines = append(lines, "pub type Root = "+rootType+";")
	case LangJava:
		lines = append(lines, "import java.util.List;", "import java.util.Map;", "")
		for _, name := range gen.order {
			lines = append(lines, gen.types[name])
			lines = append(lines, "")
		}
	default:
		for _, name := range gen.order {
			lines = append(lines, gen.types[name])
			lines = append(lines, "")
		}
	}

	return tools.Result{Data: JSONToTypesResult{
		Language:  language,
		Output:    strings.Join(lines, "\n"),
		TypeCount: len(gen.order),
	}}, nil
}

// typeGenerator 负责推断与渲染类型。
type typeGenerator struct {
	// language 是目标语言。
	language string
	// types 是类型名到定义文本的映射。
	types map[string]string
	// order 是类型定义的输出顺序。
	order []string
}

// infer 推断一个值对应的类型名。
func (g *typeGenerator) infer(name string, value any) string {
	switch typed := value.(type) {
	case map[string]any:
		return g.inferObject(name, typed)
	case []any:
		return g.inferArray(name, typed)
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// inferObject 推断对象类型。
func (g *typeGenerator) inferObject(name string, value map[string]any) string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fields := []string{}
	for _, key := range keys {
		fieldType := g.infer(g.nestedName(name, key), value[key])
		fields = append(fields, g.renderField(key, fieldType))
	}

	g.addType(name, g.renderObject(name, fields))
	return name
}

// inferArray 推断数组类型。
func (g *typeGenerator) inferArray(name string, value []any) string {
	if len(value) == 0 {
		g.addType(name, g.renderObject(name, []string{}))
		return name
	}

	// 逐项推断并合并：同名字段的不同类型用联合类型表示。
	itemTypes := []string{}
	for _, item := range value {
		itemType := g.infer(g.nestedName(name, "Item"), item)
		if !containsString(itemTypes, itemType) {
			itemTypes = append(itemTypes, itemType)
		}
	}

	elementType := itemTypes[0]
	if len(itemTypes) > 1 {
		elementType = g.union(itemTypes)
	}

	g.addType(name, g.renderTypeAlias(name, g.arrayOf(elementType)))
	return name
}

// nestedName 生成嵌套类型名。
func (g *typeGenerator) nestedName(parent, child string) string {
	clean := sanitizeTypeName(child)
	if clean == "" {
		return parent + "Item"
	}
	return parent + upperFirst(clean)
}

// addType 登记类型定义。
func (g *typeGenerator) addType(name string, definition string) {
	if _, ok := g.types[name]; ok {
		return
	}
	g.types[name] = definition
	g.order = append(g.order, name)
}

// renderObject 渲染对象类型定义。
func (g *typeGenerator) renderObject(name string, fields []string) string {
	switch g.language {
	case LangGo:
		lines := []string{"type " + name + " struct {"}
		for _, field := range fields {
			lines = append(lines, "\t"+field)
		}
		lines = append(lines, "}")
		return strings.Join(lines, "\n")
	case LangRust:
		lines := []string{"#[derive(Debug, Clone)]", "pub struct " + name + " {"}
		for _, field := range fields {
			lines = append(lines, "\t"+field)
		}
		lines = append(lines, "}")
		return strings.Join(lines, "\n")
	case LangJava:
		lines := []string{"public class " + name + " {"}
		for _, field := range fields {
			lines = append(lines, "\t"+field+";")
		}
		lines = append(lines, "}")
		return strings.Join(lines, "\n")
	default:
		lines := []string{"export interface " + name + " {"}
		for _, field := range fields {
			lines = append(lines, "\t"+field)
		}
		lines = append(lines, "}")
		return strings.Join(lines, "\n")
	}
}

// renderTypeAlias 渲染类型别名。
func (g *typeGenerator) renderTypeAlias(name, target string) string {
	switch g.language {
	case LangGo:
		return "type " + name + " " + target
	case LangRust:
		return "pub type " + name + " = " + target + ";"
	case LangJava:
		return "public class " + name + " {}"
	default:
		return "export type " + name + " = " + target + ";"
	}
}

// renderField 渲染单个字段。
func (g *typeGenerator) renderField(key, fieldType string) string {
	fieldName := sanitizeFieldName(key)
	switch g.language {
	case LangGo:
		goName := upperFirst(fieldName)
		return goName + " " + g.goType(fieldType) + " `json:\"" + key + "\"`"
	case LangRust:
		return fieldName + ": " + g.rustType(fieldType) + ","
	case LangJava:
		return "public " + g.javaType(fieldType) + " " + fieldName
	default:
		return fieldName + ": " + g.tsType(fieldType)
	}
}

// tsType 把类型名转成 TypeScript 类型。
func (g *typeGenerator) tsType(name string) string {
	switch name {
	case "string":
		return "string"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "null":
		return "null"
	default:
		return name
	}
}

// goType 把类型名转成 Go 类型。
func (g *typeGenerator) goType(name string) string {
	switch name {
	case "string":
		return "string"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "null":
		return "any"
	default:
		if strings.HasPrefix(name, "[]") {
			return "[]" + g.goType(strings.TrimPrefix(name, "[]"))
		}
		return name
	}
}

// rustType 把类型名转成 Rust 类型。
func (g *typeGenerator) rustType(name string) string {
	switch name {
	case "string":
		return "String"
	case "number":
		return "f64"
	case "boolean":
		return "bool"
	case "null":
		return "Option<serde_json::Value>"
	default:
		if strings.HasPrefix(name, "[]") {
			return "Vec<" + g.rustType(strings.TrimPrefix(name, "[]")) + ">"
		}
		return name
	}
}

// javaType 把类型名转成 Java 类型。
func (g *typeGenerator) javaType(name string) string {
	switch name {
	case "string":
		return "String"
	case "number":
		return "double"
	case "boolean":
		return "boolean"
	case "null":
		return "Object"
	default:
		if strings.HasPrefix(name, "[]") {
			return "List<" + g.javaType(strings.TrimPrefix(name, "[]")) + ">"
		}
		return name
	}
}

// arrayOf 包装成数组类型。
func (g *typeGenerator) arrayOf(element string) string {
	switch g.language {
	case LangGo:
		return "[]" + element
	case LangRust:
		return "Vec<" + element + ">"
	case LangJava:
		return "List<" + element + ">"
	default:
		return element + "[]"
	}
}

// union 合并多个类型。
func (g *typeGenerator) union(types []string) string {
	switch g.language {
	case LangGo:
		return "any"
	case LangRust:
		return "serde_json::Value"
	case LangJava:
		return "Object"
	default:
		return strings.Join(types, " | ")
	}
}

// sanitizeTypeName 把字段名转成合法的类型名片段。
func sanitizeTypeName(raw string) string {
	out := []rune{}
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		}
	}
	return string(out)
}

// sanitizeFieldName 把字段名转成合法的标识符。
func sanitizeFieldName(raw string) string {
	clean := sanitizeTypeName(raw)
	if clean == "" {
		return "field"
	}
	if raw[0] >= '0' && raw[0] <= '9' {
		clean = "f" + clean
	}
	return clean
}

// upperFirst 把首字母转成大写。
func upperFirst(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

// Validate 做字段级校验。
func (r *JSONToTypesRequest) Validate() error {
	language := r.Language
	if language == "" {
		language = LangTypeScript
	}
	if !containsString(JSONToTypesLanguage(), language) {
		return tools.InvalidField("language", JSONToTypesLanguage(), language)
	}
	if strings.TrimSpace(r.Input) == "" {
		return tools.MissingField("input")
	}
	if len(r.Input) > maxJSONToTypesBytes {
		return tools.InvalidField("input", []int{0, maxJSONToTypesBytes}, len(r.Input))
	}
	if r.RootName != "" && sanitizeTypeName(r.RootName) == "" {
		return tools.InvalidField("rootName", "alphanumeric name", "invalid format")
	}
	return nil
}
