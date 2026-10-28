package formatting

// xmlChild 是 XML 元素的一个子节点。
type xmlChild struct {
	// name 是元素名。
	name string
	// value 是元素的值。
	value *dataValue
}

// xmlFrame 是一个尚未闭合的 XML 元素。
type xmlFrame struct {
	// name 是元素名。
	name string
	// children 是子节点。
	children []xmlChild
	// text 是累积的文本内容。
	text string
	// hasText 标记是否出现过文本。
	hasText bool
}

// xmlTree 从 token 流构建 XML 树。
//
// 返回根元素名与它的子节点列表。XML **没有**「无序对象」的概念：
// 根元素固定，同名的重复子元素按出现顺序保留为多个成员。
func xmlTree(tokens []xmlToken) (string, []xmlChild) {
	var stack []*xmlFrame
	var root string
	var roots []xmlChild

	for _, tok := range tokens {
		switch tok.kind {
		case xmlTokenStart:
			stack = append(stack, &xmlFrame{name: tok.text})

		case xmlTokenEnd:
			if len(stack) == 0 {
				continue
			}
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			child := xmlChild{name: current.name, value: frameValue(current)}

			if len(stack) == 0 {
				root = current.name
				roots = append(roots, child)
				continue
			}
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, child)

		case xmlTokenText:
			if len(stack) == 0 {
				continue
			}
			current := stack[len(stack)-1]
			current.text += tok.text
			current.hasText = true

		default:
			// 注释与处理指令在结构转换中被忽略：
			// 它们不是数据。
		}
	}

	return root, roots
}

// frameValue 把一个 XML 元素帧转成中间表示的节点。
func frameValue(f *xmlFrame) *dataValue {
	if len(f.children) == 0 {
		if f.hasText {
			return &dataValue{kind: dataString, text: trimXMLText(f.text)}
		}
		return &dataValue{kind: dataObject}
	}

	out := &dataValue{kind: dataObject}
	for _, child := range f.children {
		out.members = append(out.members, dataMember{Key: child.name, Value: child.value})
	}
	return out
}

// trimXMLText 去掉文本两端空白。
func trimXMLText(s string) string {
	start := 0
	for start < len(s) && isXMLSpace(s[start]) {
		start++
	}
	end := len(s)
	for end > start && isXMLSpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

// isXMLSpace 判断字节是否为 XML 空白。
func isXMLSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}
