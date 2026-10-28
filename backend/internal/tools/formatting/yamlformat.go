package formatting

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// maxYAMLAliases 是 YAML 别名节点的上限。
//
// 别名（anchor / alias）可以指数级放大文档体积：
// 一个锚点被引用 N 次，内容就膨胀 N 倍。
// 因此在**展开之前**统计别名节点数，超限直接拒绝。
const maxYAMLAliases = 1000

// maxYAMLDepth 是 YAML 嵌套深度上限。
const maxYAMLDepth = 100

// YAMLFormatRequest 是 yaml-format 工具的请求。
type YAMLFormatRequest struct {
	// Input 是待处理的 YAML 文本。
	Input TextInput `json:"input"`
	// Action 是 format、minify 或 validate。
	Action string `json:"action"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
	// SortKeys 是否按键排序。
	SortKeys bool `json:"sortKeys"`
}

// YAMLFormatResult 是 yaml-format 工具的响应。
type YAMLFormatResult struct {
	// Output 是处理结果；validate 时为空串。
	Output string `json:"output"`
	// Valid 是内容是否合法。
	Valid bool `json:"valid"`
	// Error 是语法错误的位置与原因；合法时为 nil。
	Error *SyntaxError `json:"error,omitempty"`
}

// YAMLFormat 美化、压缩或校验 YAML。
func YAMLFormat(req *YAMLFormatRequest) (YAMLFormatResult, error) {
	if err := req.Validate(); err != nil {
		return YAMLFormatResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return YAMLFormatResult{}, err
	}

	node, synErr := parseYAML(text)
	if synErr != nil {
		return YAMLFormatResult{Output: "", Valid: false, Error: synErr}, nil
	}

	if action(req.Action) == ActionValidate {
		return YAMLFormatResult{Valid: true}, nil
	}

	var out string
	switch action(req.Action) {
	case ActionMinify:
		out = minifyYAML(node)
	default:
		out = formatYAML(node, indentWidth(req.Indent), req.SortKeys)
	}

	return YAMLFormatResult{Output: out, Valid: true}, nil
}

// Validate 做字段级校验。
func (r *YAMLFormatRequest) Validate() error {
	if err := validateAction(r.Action); err != nil {
		return err
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	return r.Input.Validate("input")
}

// parseYAML 解析 YAML，并在展开前检查别名与深度。
//
// 先按 `yaml.Node` 解析（节点树里别名是**引用**而非展开后的内容），
// 统计别名数量与嵌套深度，确认安全后再交给编码器展开。
func parseYAML(text string) (*yaml.Node, *SyntaxError) {
	var node yaml.Node

	dec := yaml.NewDecoder(strings.NewReader(text))
	if err := dec.Decode(&node); err != nil {
		if err.Error() == "EOF" {
			// 空文档：合法。
			return &yaml.Node{Kind: yaml.DocumentNode}, nil
		}
		se := yamlSyntaxError(err)
		return nil, &se
	}

	aliases := 0
	if depth := nodeDepth(&node, 0); depth > maxYAMLDepth {
		se := SyntaxError{Message: "document nesting is too deep"}
		return nil, &se
	}
	countAliases(&node, &aliases)
	if aliases > maxYAMLAliases {
		se := SyntaxError{Message: "too many aliases"}
		return nil, &se
	}

	return &node, nil
}

// countAliases 统计节点树里的别名数量。
func countAliases(node *yaml.Node, count *int) {
	if node == nil {
		return
	}
	if node.Kind == yaml.AliasNode {
		*count++
	}
	for _, child := range node.Content {
		countAliases(child, count)
	}
}

// nodeDepth 返回节点树的最大嵌套深度。
func nodeDepth(node *yaml.Node, depth int) int {
	if node == nil {
		return depth
	}
	max := depth
	for _, child := range node.Content {
		if d := nodeDepth(child, depth+1); d > max {
			max = d
		}
	}
	return max
}

// formatYAML 美化 YAML。
//
// 走 yaml.Node 而不是 map[string]any，因此**键序与注释都被保留**。
func formatYAML(node *yaml.Node, indent int, sortKeys bool) string {
	if sortKeys {
		sortYAMLKeys(node)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	if err := enc.Encode(node); err != nil {
		return ""
	}
	_ = enc.Close()

	return strings.TrimRight(buf.String(), "\n")
}

// minifyYAML 压缩 YAML。
//
// YAML 的「压缩」就是去掉多余空行与文档分隔符；
// 缩进本身是语义的，不能去掉。
func minifyYAML(node *yaml.Node) string {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(1)
	if err := enc.Encode(node); err != nil {
		return ""
	}
	_ = enc.Close()

	lines := strings.Split(buf.String(), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed == "" {
			continue
		}
		if trimmed == "---" || trimmed == "..." {
			continue
		}
		out = append(out, trimmed)
	}
	return strings.Join(out, "\n")
}

// sortYAMLKeys 递归按键排序映射节点的键。
func sortYAMLKeys(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		sortMapping(node)
	}
	for _, child := range node.Content {
		sortYAMLKeys(child)
	}
}

// sortMapping 对映射节点的键值对排序。
func sortMapping(node *yaml.Node) {
	if len(node.Content) < 2 {
		return
	}

	type pair struct {
		key   *yaml.Node
		value *yaml.Node
	}
	pairs := make([]pair, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		pairs = append(pairs, pair{key: node.Content[i], value: node.Content[i+1]})
	}

	// 稳定排序：相同键序按原顺序保留。
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0 && pairs[j].key.Value < pairs[j-1].key.Value; j-- {
			pairs[j], pairs[j-1] = pairs[j-1], pairs[j]
		}
	}

	node.Content = node.Content[:0]
	for _, p := range pairs {
		node.Content = append(node.Content, p.key, p.value)
	}
}

// yamlLineRE 匹配 yaml.v3 错误消息里的行号。
var yamlLineRE = regexp.MustCompile(`line (\d+)`)

// atoiSafe 把字符串转成整数，失败时返回错误。
func atoiSafe(s string) (int, error) {
	return strconv.Atoi(s)
}

// yamlSyntaxError 把 yaml.v3 的错误转成带行列的结构化错误。
func yamlSyntaxError(err error) SyntaxError {
	se := SyntaxError{Message: "invalid YAML"}

	msg := err.Error()
	if loc := yamlLineRE.FindStringSubmatch(msg); loc != nil {
		if line, convErr := strconv.Atoi(loc[1]); convErr == nil {
			se.Line = line
			se.Column = 1
		}
	}
	return se
}
