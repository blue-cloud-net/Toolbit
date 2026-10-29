package system

import (
	"strconv"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 权限类别。
const (
	// PermUser 是属主权限。
	PermUser = "user"
	// PermGroup 是属组权限。
	PermGroup = "group"
	// PermOther 是其他人权限。
	PermOther = "other"
)

// PermClasses 返回受支持的权限类别（顺序稳定）。
func PermClasses() []string {
	return []string{PermUser, PermGroup, PermOther}
}

// 权限位。
const (
	// PermRead 是读权限。
	PermRead = "read"
	// PermWrite 是写权限。
	PermWrite = "write"
	// PermExecute 是执行权限。
	PermExecute = "execute"
)

// PermBits 返回受支持的权限位（顺序稳定）。
func PermBits() []string {
	return []string{PermRead, PermWrite, PermExecute}
}

// ChmodCalcRequest 是 chmod-calc 工具的请求。
type ChmodCalcRequest struct {
	// Mode 是权限：符号形式（`rwxr-xr-x`）或八进制形式（`755`）。
	Mode string `json:"mode"`
	// User 是属主权限位；给出时按位**设置**。
	User []string `json:"user"`
	// Group 是属组权限位。
	Group []string `json:"group"`
	// Other 是其他人权限位。
	Other []string `json:"other"`
}

// PermClass 是一个权限类别的明细。
type PermClass struct {
	// Class 是类别名：user / group / other。
	Class string `json:"class"`
	// Read 是否可读。
	Read bool `json:"read"`
	// Write 是否可写。
	Write bool `json:"write"`
	// Execute 是否可执行。
	Execute bool `json:"execute"`
	// Symbol 是符号形式的三个字符。
	Symbol string `json:"symbol"`
	// Numeric 是该类别的数字形式（4 / 2 / 1）。
	Numeric int `json:"numeric"`
	// Meaning 是文字说明。
	Meaning string `json:"meaning"`
}

// ChmodCalcResult 是 chmod-calc 工具的响应。
type ChmodCalcResult struct {
	// Octal 是八进制形式，如 `755`。
	Octal string `json:"octal"`
	// Symbolic 是符号形式，如 `rwxr-xr-x`。
	Symbolic string `json:"symbolic"`
	// Classes 是三个权限类别的明细。
	Classes []PermClass `json:"classes"`
}

// ChmodCalc 在符号权限与八进制之间互转。
//
// **只**处理低 9 位：`setuid` / `setgid` / `sticky` **不在**范围内。
func ChmodCalc(req *ChmodCalcRequest) (ChmodCalcResult, error) {
	if err := req.Validate(); err != nil {
		return ChmodCalcResult{}, err
	}

	mode, err := req.resolveMode()
	if err != nil {
		return ChmodCalcResult{}, err
	}

	return ChmodCalcResult{
		Octal:    formatOctal(mode),
		Symbolic: formatSymbolic(mode),
		Classes:  describeMode(mode),
	}, nil
}

// Validate 做字段级校验。
func (r *ChmodCalcRequest) Validate() error {
	for _, class := range []struct {
		name string
		bits []string
	}{
		{PermUser, r.User},
		{PermGroup, r.Group},
		{PermOther, r.Other},
	} {
		for _, bit := range class.bits {
			switch bit {
			case PermRead, PermWrite, PermExecute:
			default:
				return tools.InvalidField(class.name, PermBits(), bit)
			}
		}
	}

	if r.Mode == "" && len(r.User) == 0 && len(r.Group) == 0 && len(r.Other) == 0 {
		return tools.MissingField("mode")
	}
	return nil
}

// resolveMode 解析出九位权限。
//
// 优先级：显式的权限位（user / group / other）覆盖 `mode` 给出的值。
func (r *ChmodCalcRequest) resolveMode() (int, error) {
	mode := 0

	if r.Mode != "" {
		parsed, err := parseMode(r.Mode)
		if err != nil {
			return 0, err
		}
		mode = parsed
	}

	if len(r.User) > 0 || len(r.Group) > 0 || len(r.Other) > 0 {
		// 从零重建，逐类别覆盖。
		base := mode &^ 0o777
		if len(r.User) > 0 {
			base |= bitsToMode(r.User) << 6
		} else {
			base |= (mode >> 6 & 0o7) << 6
		}
		if len(r.Group) > 0 {
			base |= bitsToMode(r.Group) << 3
		} else {
			base |= (mode >> 3 & 0o7) << 3
		}
		if len(r.Other) > 0 {
			base |= bitsToMode(r.Other)
		} else {
			base |= mode & 0o7
		}
		mode = base
	}

	return mode, nil
}

// parseMode 解析符号或八进制权限。
func parseMode(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, toolerr.Newf(toolerr.CodeInvalidChmod, map[string]any{
			"reason": "empty permission value",
		})
	}

	// 八进制形式：纯数字且长度不超过 4。
	if isOctalDigits(trimmed) {
		value, err := strconv.ParseInt(trimmed, 8, 32)
		if err != nil {
			return 0, toolerr.Newf(toolerr.CodeInvalidChmod, map[string]any{
				"reason": "invalid octal permission value",
			})
		}
		if value > 0o7777 {
			return 0, toolerr.Newf(toolerr.CodeInvalidChmod, map[string]any{
				"reason": "permission value out of range",
			})
		}
		return int(value) & 0o777, nil
	}

	return parseSymbolicMode(trimmed)
}

// parseSymbolicMode 解析符号形式的权限。
//
// 接受 `rwxr-xr-x` 这样的九字符形式，也接受 `u=rwx,g=rx,o=rx` 这样的长形式。
func parseSymbolicMode(raw string) (int, error) {
	if strings.Contains(raw, "=") {
		return parseLongSymbolicMode(raw)
	}

	if len(raw) != 9 {
		return 0, toolerr.Newf(toolerr.CodeInvalidChmod, map[string]any{
			"reason": "symbolic permission must have 9 characters",
		})
	}

	mode := 0
	for i := 0; i < 3; i++ {
		bits := raw[i*3 : i*3+3]
		value := 0
		for j, c := range bits {
			switch c {
			case 'r':
				if j != 0 {
					return 0, invalidChmod()
				}
				value |= 4
			case 'w':
				if j != 1 {
					return 0, invalidChmod()
				}
				value |= 2
			case 'x':
				if j != 2 {
					return 0, invalidChmod()
				}
				value |= 1
			case '-':
			default:
				return 0, invalidChmod()
			}
		}
		mode |= value << (6 - i*3)
	}

	return mode, nil
}

// parseLongSymbolicMode 解析 `u=rwx,g=rx,o=rx` 形式。
func parseLongSymbolicMode(raw string) (int, error) {
	mode := 0
	seen := false

	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		class, bits, ok := strings.Cut(part, "=")
		if !ok {
			return 0, invalidChmod()
		}
		class = strings.TrimSpace(class)

		var shift int
		switch class {
		case "u", "user":
			shift = 6
		case "g", "group":
			shift = 3
		case "o", "other":
			shift = 0
		default:
			return 0, invalidChmod()
		}

		value := 0
		for _, c := range bits {
			switch c {
			case 'r':
				value |= 4
			case 'w':
				value |= 2
			case 'x':
				value |= 1
			case '-':
			default:
				return 0, invalidChmod()
			}
		}
		mode |= value << shift
		seen = true
	}

	if !seen {
		return 0, invalidChmod()
	}
	return mode, nil
}

// bitsToMode 把权限位列表转成三位数值。
func bitsToMode(bits []string) int {
	value := 0
	for _, bit := range bits {
		switch bit {
		case PermRead:
			value |= 4
		case PermWrite:
			value |= 2
		case PermExecute:
			value |= 1
		}
	}
	return value
}

// formatOctal 把权限渲染成八进制字符串。
func formatOctal(mode int) string {
	digits := []byte("000")
	for i := 0; i < 3; i++ {
		digits[i] = byte('0' + ((mode >> (6 - i*3)) & 0o7))
	}
	return string(digits)
}

// formatSymbolic 把权限渲染成符号形式。
func formatSymbolic(mode int) string {
	var b strings.Builder
	for i := 0; i < 3; i++ {
		value := (mode >> (6 - i*3)) & 0o7
		if value&4 != 0 {
			b.WriteByte('r')
		} else {
			b.WriteByte('-')
		}
		if value&2 != 0 {
			b.WriteByte('w')
		} else {
			b.WriteByte('-')
		}
		if value&1 != 0 {
			b.WriteByte('x')
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}

// describeMode 列出三个权限类别的明细。
func describeMode(mode int) []PermClass {
	classes := []PermClass{}
	for i, name := range PermClasses() {
		value := (mode >> (6 - i*3)) & 0o7
		classes = append(classes, PermClass{
			Class:   name,
			Read:    value&4 != 0,
			Write:   value&2 != 0,
			Execute: value&1 != 0,
			Symbol:  classSymbol(value),
			Numeric: value,
			Meaning: classMeaning(value),
		})
	}
	return classes
}

// classSymbol 渲染单个类别的符号形式。
func classSymbol(value int) string {
	var b strings.Builder
	if value&4 != 0 {
		b.WriteByte('r')
	} else {
		b.WriteByte('-')
	}
	if value&2 != 0 {
		b.WriteByte('w')
	} else {
		b.WriteByte('-')
	}
	if value&1 != 0 {
		b.WriteByte('x')
	} else {
		b.WriteByte('-')
	}
	return b.String()
}

// classMeaning 给出单个类别的文字说明。
func classMeaning(value int) string {
	switch value {
	case 0:
		return "无权限"
	case 1:
		return "仅执行"
	case 2:
		return "仅写"
	case 3:
		return "写与执行"
	case 4:
		return "仅读"
	case 5:
		return "读与执行"
	case 6:
		return "读与写"
	case 7:
		return "读写与执行"
	default:
		return "未知"
	}
}

// isOctalDigits 判断字符串是否只含八进制数字。
func isOctalDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '7' {
			return false
		}
	}
	return true
}

// invalidChmod 构造权限值非法的错误。
func invalidChmod() error {
	return toolerr.Newf(toolerr.CodeInvalidChmod, map[string]any{
		"reason": "invalid permission value",
	})
}
