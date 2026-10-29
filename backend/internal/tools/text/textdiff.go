package text

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 差异对比模式。
const (
	// DiffModeText 按行做 LCS 对比。
	DiffModeText = "text"
	// DiffModeJSON 按结构化路径标注增删改。
	DiffModeJSON = "json"
)

// DiffMode 返回受支持的对比模式（顺序稳定）。
func DiffMode() []string {
	return []string{DiffModeText, DiffModeJSON}
}

// 差异操作类型。
const (
	// DiffOpEqual 未变更。
	DiffOpEqual = "equal"
	// DiffOpInsert 新增。
	DiffOpInsert = "insert"
	// DiffOpDelete 删除。
	DiffOpDelete = "delete"
	// DiffOpUpdate 值发生变化。
	DiffOpUpdate = "update"
)

// DiffInput 是对比输入。
type DiffInput struct {
	// Data 是文本内容。
	Data string `json:"data"`
}

// TextDiffRequest 是 text-diff 工具的请求。
type TextDiffRequest struct {
	// Mode 是对比模式：text / json；默认 text。
	Mode string `json:"mode"`
	// Left 是基准文本。
	Left DiffInput `json:"left"`
	// Right 是对比文本。
	Right DiffInput `json:"right"`
	// IgnoreWhitespace 是否忽略行尾空白。
	IgnoreWhitespace bool `json:"ignoreWhitespace"`
	// ContextLines 是每处差异保留的上下文行数；默认 3。
	ContextLines int `json:"contextLines"`
}

// DiffLine 是一行差异。
type DiffLine struct {
	// Op 是操作类型：equal / insert / delete。
	Op string `json:"op"`
	// LeftLineNum 是基准侧行号；该侧不存在时为 0。
	LeftLineNum int `json:"leftLineNum"`
	// RightLineNum 是对比侧行号；该侧不存在时为 0。
	RightLineNum int `json:"rightLineNum"`
	// Text 是行内容。
	Text string `json:"text"`
}

// DiffChange 是一处结构化差异。
type DiffChange struct {
	// Path 是结构化路径，如 `$.items[2].name`。
	Path string `json:"path"`
	// Op 是操作类型：insert / delete / update。
	Op string `json:"op"`
	// Left 是基准侧的值；不存在时为空串。
	Left string `json:"left"`
	// Right 是对比侧的值；不存在时为空串。
	Right string `json:"right"`
}

// DiffStat 是差异统计。
type DiffStat struct {
	// Added 是新增行数。
	Added int `json:"added"`
	// Removed 是删除行数。
	Removed int `json:"removed"`
	// Unchanged 是未变行数。
	Unchanged int `json:"unchanged"`
	// Changes 是结构化差异条数（JSON 模式）。
	Changes int `json:"changes"`
}

// TextDiffResult 是 text-diff 工具的响应。
type TextDiffResult struct {
	// Mode 是实际使用的对比模式。
	Mode string `json:"mode"`
	// Identical 是否两侧完全一致。
	Identical bool `json:"identical"`
	// Lines 是行级差异（text 模式）。
	Lines []DiffLine `json:"lines"`
	// Changes 是结构化差异（json 模式）。
	Changes []DiffChange `json:"changes"`
	// Stat 是差异统计。
	Stat DiffStat `json:"stat"`
}

// maxDiffInputBytes 是单侧输入的字节上限。
//
// 差异对比是 O(n·m) 的 LCS，文本越长耗时越久；限制输入长度保证响应可控。
const maxDiffInputBytes = 512 * 1024

// maxDiffLines 是单侧的最大行数。
const maxDiffLines = 20000

// TextDiff 对比两段文本的差异。
func TextDiff(_ context.Context, req *TextDiffRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = DiffModeText
	}

	if mode == DiffModeJSON {
		return diffJSON(req)
	}
	return diffText(req)
}

// diffText 做行级 LCS 对比。
func diffText(req *TextDiffRequest) (tools.Result, error) {
	leftLines := splitDiffLines(req.Left.Data)
	rightLines := splitDiffLines(req.Right.Data)

	if req.IgnoreWhitespace {
		leftLines = trimLineSpace(leftLines)
		rightLines = trimLineSpace(rightLines)
	}

	ops := lcsDiff(leftLines, rightLines)

	lines := make([]DiffLine, 0, len(ops))
	stat := DiffStat{}
	for _, op := range ops {
		switch op.op {
		case DiffOpEqual:
			stat.Unchanged++
		case DiffOpInsert:
			stat.Added++
		case DiffOpDelete:
			stat.Removed++
		}
		lines = append(lines, DiffLine{
			Op:           op.op,
			LeftLineNum:  op.leftNum,
			RightLineNum: op.rightNum,
			Text:         op.text,
		})
	}

	return tools.Result{Data: TextDiffResult{
		Mode:      DiffModeText,
		Identical: stat.Added == 0 && stat.Removed == 0,
		Lines:     lines,
		Changes:   []DiffChange{},
		Stat:      stat,
	}}, nil
}

// diffJSON 做结构化路径差异对比。
func diffJSON(req *TextDiffRequest) (tools.Result, error) {
	leftValue, err := decodeJSONValue(req.Left.Data)
	if err != nil {
		return tools.Result{}, tools.InvalidField("left.data", "valid JSON", "parse error")
	}
	rightValue, err := decodeJSONValue(req.Right.Data)
	if err != nil {
		return tools.Result{}, tools.InvalidField("right.data", "valid JSON", "parse error")
	}

	changes := []DiffChange{}
	walkJSONDiff("$", leftValue, rightValue, &changes)

	// 路径排序让结果稳定，便于比对与快照。
	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].Path != changes[j].Path {
			return changes[i].Path < changes[j].Path
		}
		return changes[i].Op < changes[j].Op
	})

	stat := DiffStat{Changes: len(changes)}
	return tools.Result{Data: TextDiffResult{
		Mode:      DiffModeJSON,
		Identical: len(changes) == 0,
		Lines:     []DiffLine{},
		Changes:   changes,
		Stat:      stat,
	}}, nil
}

// diffOp 是一次 LCS 转移的结果。
type diffOp struct {
	// op 是操作类型。
	op string
	// leftNum 是基准侧行号。
	leftNum int
	// rightNum 是对比侧行号。
	rightNum int
	// text 是行内容。
	text string
}

// lcsDiff 用最长公共子序列对齐两段行。
//
// 实现是标准的动态规划回溯；为控制内存，滚动数组只保存一行。
func lcsDiff(left, right []string) []diffOp {
	n, m := len(left), len(right)
	if n == 0 && m == 0 {
		return []diffOp{}
	}

	// dp[i][j] 是 left[i:] 与 right[j:] 的 LCS 长度。
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if left[i] == right[j] {
				dp[i][j] = dp[i+1][j+1] + 1
				continue
			}
			if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	ops := make([]diffOp, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case left[i] == right[j]:
			ops = append(ops, diffOp{op: DiffOpEqual, leftNum: i + 1, rightNum: j + 1, text: left[i]})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			ops = append(ops, diffOp{op: DiffOpDelete, leftNum: i + 1, rightNum: 0, text: left[i]})
			i++
		default:
			ops = append(ops, diffOp{op: DiffOpInsert, leftNum: 0, rightNum: j + 1, text: right[j]})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, diffOp{op: DiffOpDelete, leftNum: i + 1, rightNum: 0, text: left[i]})
	}
	for ; j < m; j++ {
		ops = append(ops, diffOp{op: DiffOpInsert, leftNum: 0, rightNum: j + 1, text: right[j]})
	}
	return ops
}

// splitDiffLines 按行切分文本。
//
// 末尾空行不计入：多数编辑器在行尾换行后不再视其为一行。
func splitDiffLines(text string) []string {
	if text == "" {
		return []string{}
	}
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// trimLineSpace 去掉每行两端空白。
func trimLineSpace(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = strings.TrimSpace(line)
	}
	return out
}

// decodeJSONValue 解析 JSON 文本。
//
// 用 json.Number 保留数字原文，避免 `1` 与 `1.0` 被判成不同值。
func decodeJSONValue(text string) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

// walkJSONDiff 递归比对两个 JSON 值并收集差异。
func walkJSONDiff(path string, left, right any, changes *[]DiffChange) {
	leftMap, leftIsMap := left.(map[string]any)
	rightMap, rightIsMap := right.(map[string]any)
	if leftIsMap && rightIsMap {
		keys := map[string]struct{}{}
		for key := range leftMap {
			keys[key] = struct{}{}
		}
		for key := range rightMap {
			keys[key] = struct{}{}
		}
		names := make([]string, 0, len(keys))
		for key := range keys {
			names = append(names, key)
		}
		sort.Strings(names)
		for _, key := range names {
			childPath := path + "." + key
			leftValue, leftOK := leftMap[key]
			rightValue, rightOK := rightMap[key]
			switch {
			case leftOK && !rightOK:
				*changes = append(*changes, DiffChange{
					Path: childPath, Op: DiffOpDelete,
					Left: formatJSONValue(leftValue), Right: "",
				})
			case !leftOK && rightOK:
				*changes = append(*changes, DiffChange{
					Path: childPath, Op: DiffOpInsert,
					Left: "", Right: formatJSONValue(rightValue),
				})
			default:
				walkJSONDiff(childPath, leftValue, rightValue, changes)
			}
		}
		return
	}

	leftArr, leftIsArr := left.([]any)
	rightArr, rightIsArr := right.([]any)
	if leftIsArr && rightIsArr {
		maxLen := len(leftArr)
		if len(rightArr) > maxLen {
			maxLen = len(rightArr)
		}
		for idx := 0; idx < maxLen; idx++ {
			childPath := fmt.Sprintf("%s[%d]", path, idx)
			switch {
			case idx >= len(leftArr):
				*changes = append(*changes, DiffChange{
					Path: childPath, Op: DiffOpInsert,
					Left: "", Right: formatJSONValue(rightArr[idx]),
				})
			case idx >= len(rightArr):
				*changes = append(*changes, DiffChange{
					Path: childPath, Op: DiffOpDelete,
					Left: formatJSONValue(leftArr[idx]), Right: "",
				})
			default:
				walkJSONDiff(childPath, leftArr[idx], rightArr[idx], changes)
			}
		}
		return
	}

	if formatJSONValue(left) != formatJSONValue(right) {
		*changes = append(*changes, DiffChange{
			Path: path, Op: DiffOpUpdate,
			Left: formatJSONValue(left), Right: formatJSONValue(right),
		})
	}
}

// formatJSONValue 把 JSON 值渲染成稳定的单行文本。
func formatJSONValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return typed
	case json.Number:
		return typed.String()
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		raw, err := json.Marshal(value)
		if err != nil {
			return ""
		}
		return string(raw)
	}
}

// Validate 做字段级校验。
func (r *TextDiffRequest) Validate() error {
	mode := r.Mode
	if mode == "" {
		mode = DiffModeText
	}
	switch mode {
	case DiffModeText, DiffModeJSON:
	default:
		return tools.InvalidField("mode", DiffMode(), mode)
	}

	if len(r.Left.Data) == 0 {
		return tools.MissingField("left.data")
	}
	if len(r.Right.Data) == 0 {
		return tools.MissingField("right.data")
	}
	if len(r.Left.Data) > maxDiffInputBytes {
		return tools.InvalidField("left.data", []int{0, maxDiffInputBytes}, len(r.Left.Data))
	}
	if len(r.Right.Data) > maxDiffInputBytes {
		return tools.InvalidField("right.data", []int{0, maxDiffInputBytes}, len(r.Right.Data))
	}
	if len(splitDiffLines(r.Left.Data)) > maxDiffLines {
		return tools.InvalidField("left.data", []int{0, maxDiffLines}, maxDiffLines)
	}
	if len(splitDiffLines(r.Right.Data)) > maxDiffLines {
		return tools.InvalidField("right.data", []int{0, maxDiffLines}, maxDiffLines)
	}
	if r.ContextLines < 0 || r.ContextLines > 100 {
		return tools.InvalidField("contextLines", []int{0, 100}, r.ContextLines)
	}
	return nil
}
