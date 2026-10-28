package formatting

import (
	"strings"
	"unicode"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// SQL 方言。
const (
	// SQLDialectStandard 是通用 SQL。
	SQLDialectStandard = "standard"
	// SQLDialectMySQL 是 MySQL。
	SQLDialectMySQL = "mysql"
	// SQLDialectPostgres 是 PostgreSQL。
	SQLDialectPostgres = "pg"
	// SQLDialectSQLite 是 SQLite。
	SQLDialectSQLite = "sqlite"
)

// SQL 关键字大小写策略。
const (
	// SQLKeywordKeep 保持原样。
	SQLKeywordKeep = "keep"
	// SQLKeywordUpper 转成大写。
	SQLKeywordUpper = "upper"
	// SQLKeywordLower 转成小写。
	SQLKeywordLower = "lower"
)

// SQLDialects 返回受支持的方言（顺序稳定）。
func SQLDialects() []string {
	return []string{SQLDialectStandard, SQLDialectMySQL, SQLDialectPostgres, SQLDialectSQLite}
}

// SQLFormatRequest 是 sql-format 工具的请求。
type SQLFormatRequest struct {
	// Input 是待处理的 SQL 文本。
	Input TextInput `json:"input"`
	// Action 是 format、minify 或 validate。
	Action string `json:"action"`
	// Indent 是缩进宽度；默认 2。
	Indent int `json:"indent"`
	// Dialect 是方言；默认 standard。
	Dialect string `json:"dialect"`
	// KeywordCase 是关键字大小写策略；默认 keep。
	KeywordCase string `json:"keywordCase"`
}

// SQLFormatResult 是 sql-format 工具的响应。
type SQLFormatResult struct {
	// Output 是处理结果；validate 时为空串。
	Output string `json:"output"`
	// Valid 是内容是否合法。
	Valid bool `json:"valid"`
	// Error 是语法错误的位置与原因；合法时为 nil。
	Error *SyntaxError `json:"error,omitempty"`
}

// SQLFormat 美化、压缩或校验 SQL。
//
// 本工具**只做分词与重排空白**：token 序列与 token 文本（除关键字大小写）
// 完全不变。因此它**不执行** SQL，也**不会**改变语句语义——
// 但方言差异仍可能影响语义，改动后请复核。
func SQLFormat(req *SQLFormatRequest) (SQLFormatResult, error) {
	if err := req.Validate(); err != nil {
		return SQLFormatResult{}, err
	}

	text, err := req.Input.Text()
	if err != nil {
		return SQLFormatResult{}, err
	}

	tokens, synErr := tokenizeSQL(text)
	if synErr != nil {
		return SQLFormatResult{Output: "", Valid: false, Error: synErr}, nil
	}

	if action(req.Action) == ActionValidate {
		return SQLFormatResult{Valid: true}, nil
	}

	switch action(req.Action) {
	case ActionMinify:
		return SQLFormatResult{Output: renderSQLMinified(tokens), Valid: true}, nil
	default:
		return SQLFormatResult{
			Output: renderSQL(tokens, indentWidth(req.Indent), req.keywordCase()),
			Valid:  true,
		}, nil
	}
}

// Validate 做字段级校验。
func (r *SQLFormatRequest) Validate() error {
	if err := validateAction(r.Action); err != nil {
		return err
	}
	if err := validateIndent(r.Indent); err != nil {
		return err
	}
	switch r.Dialect {
	case "", SQLDialectStandard, SQLDialectMySQL, SQLDialectPostgres, SQLDialectSQLite:
	default:
		return tools.InvalidField("dialect", SQLDialects(), r.Dialect)
	}
	switch r.KeywordCase {
	case "", SQLKeywordKeep, SQLKeywordUpper, SQLKeywordLower:
	default:
		return tools.InvalidField("keywordCase",
			[]string{SQLKeywordKeep, SQLKeywordUpper, SQLKeywordLower}, r.KeywordCase)
	}
	return r.Input.Validate("input")
}

// keywordCase 返回关键字大小写策略；缺省为 keep。
func (r *SQLFormatRequest) keywordCase() string {
	if r.KeywordCase == "" {
		return SQLKeywordKeep
	}
	return r.KeywordCase
}

// sqlTokenKind 是 SQL token 类型。
type sqlTokenKind int

const (
	// sqlTokenWord 是标识符或关键字。
	sqlTokenWord sqlTokenKind = iota
	// sqlTokenNumber 是数字。
	sqlTokenNumber
	// sqlTokenString 是字符串字面量。
	sqlTokenString
	// sqlTokenComment 是注释。
	sqlTokenComment
	// sqlTokenPunct 是标点或运算符。
	sqlTokenPunct
	// sqlTokenSpace 是空白。
	sqlTokenSpace
)

// sqlToken 是一个 SQL token。
type sqlToken struct {
	// kind 是 token 类型。
	kind sqlTokenKind
	// text 是 token 文本。
	text string
}

// tokenizeSQL 把 SQL 拆成 token。
//
// 识别：单引号字符串（含 ” 转义）、双引号标识符、反引号标识符、
// 行注释 `--`、块注释 `/* */`、数字、标识符与标点。
func tokenizeSQL(text string) ([]sqlToken, *SyntaxError) {
	var tokens []sqlToken
	runes := []rune(text)
	i := 0

	for i < len(runes) {
		c := runes[i]

		switch {
		case unicode.IsSpace(c):
			start := i
			for i < len(runes) && unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, sqlToken{kind: sqlTokenSpace, text: string(runes[start:i])})

		case c == '\'':
			start := i
			i++
			for i < len(runes) {
				if runes[i] == '\'' {
					// '' 是转义的单引号。
					if i+1 < len(runes) && runes[i+1] == '\'' {
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			if i > len(runes) || (i == len(runes) && !endsQuote(runes[start:i])) {
				se := SyntaxError{Line: 1, Message: "unterminated string literal"}
				return nil, &se
			}
			tokens = append(tokens, sqlToken{kind: sqlTokenString, text: string(runes[start:i])})

		case c == '"' || c == '`':
			quote := c
			start := i
			i++
			for i < len(runes) && runes[i] != quote {
				i++
			}
			if i >= len(runes) {
				se := SyntaxError{Line: 1, Message: "unterminated quoted identifier"}
				return nil, &se
			}
			i++
			tokens = append(tokens, sqlToken{kind: sqlTokenString, text: string(runes[start:i])})

		case c == '-' && i+1 < len(runes) && runes[i+1] == '-':
			start := i
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
			tokens = append(tokens, sqlToken{kind: sqlTokenComment, text: string(runes[start:i])})

		case c == '/' && i+1 < len(runes) && runes[i+1] == '*':
			start := i
			i += 2
			for i < len(runes) && (runes[i] != '*' || i+1 >= len(runes) || runes[i+1] != '/') {
				i++
			}
			if i >= len(runes) {
				se := SyntaxError{Line: 1, Message: "unterminated block comment"}
				return nil, &se
			}
			i += 2
			tokens = append(tokens, sqlToken{kind: sqlTokenComment, text: string(runes[start:i])})

		case unicode.IsDigit(c):
			start := i
			for i < len(runes) && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			tokens = append(tokens, sqlToken{kind: sqlTokenNumber, text: string(runes[start:i])})

		case isIdentStart(c):
			start := i
			for i < len(runes) && isIdentPart(runes[i]) {
				i++
			}
			tokens = append(tokens, sqlToken{kind: sqlTokenWord, text: string(runes[start:i])})

		default:
			start := i
			for i < len(runes) && !unicode.IsSpace(runes[i]) &&
				!isIdentStart(runes[i]) && !unicode.IsDigit(runes[i]) &&
				runes[i] != '\'' && runes[i] != '"' && runes[i] != '`' &&
				runes[i] != '-' && runes[i] != '/' {
				i++
			}
			if i == start {
				i++
			}
			tokens = append(tokens, sqlToken{kind: sqlTokenPunct, text: string(runes[start:i])})
		}
	}

	return tokens, nil
}

// endsQuote 判断已消费的字符是否以闭合引号结尾。
func endsQuote(runes []rune) bool {
	if len(runes) < 2 {
		return false
	}
	return runes[len(runes)-1] == '\''
}

// isIdentStart 判断字符能否作为标识符首字符。
func isIdentStart(c rune) bool {
	return c == '_' || unicode.IsLetter(c)
}

// isIdentPart 判断字符能否作为标识符的一部分。
func isIdentPart(c rune) bool {
	return c == '_' || c == '$' || unicode.IsLetter(c) || unicode.IsDigit(c)
}

// sqlKeywords 是会被关键字大小写策略影响的 SQL 关键字。
//
// 集合只覆盖常用关键字；未列出的词按标识符处理，大小写保持原样。
var sqlKeywords = map[string]bool{
	"ADD": true, "ALL": true, "ALTER": true, "AND": true, "AS": true,
	"ASC": true, "BEGIN": true, "BETWEEN": true, "BY": true, "CASE": true,
	"CAST": true, "COLUMN": true, "COMMIT": true, "CONSTRAINT": true,
	"CREATE": true, "CROSS": true, "DATABASE": true, "DEFAULT": true,
	"DELETE": true, "DESC": true, "DISTINCT": true, "DROP": true, "ELSE": true,
	"END": true, "EXCEPT": true, "EXISTS": true, "FALSE": true, "FOR": true,
	"FROM": true, "FULL": true, "GROUP": true, "HAVING": true, "IF": true,
	"IN": true, "INDEX": true, "INNER": true, "INSERT": true, "INTO": true,
	"IS": true, "JOIN": true, "KEY": true, "LEFT": true, "LIKE": true,
	"LIMIT": true, "NOT": true, "NULL": true, "OFFSET": true, "ON": true,
	"OR": true, "ORDER": true, "OUTER": true, "PRIMARY": true, "REFERENCES": true,
	"RIGHT": true, "ROLLBACK": true, "SELECT": true, "SET": true, "TABLE": true,
	"THEN": true, "TRUE": true, "UNION": true, "UNIQUE": true, "UPDATE": true,
	"USING": true, "VALUES": true, "VIEW": true, "WHEN": true, "WHERE": true,
	"WITH": true,
}

// sqlClauseKeywords 是会触发换行的子句关键字。
var sqlClauseKeywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true, "GROUP": true, "HAVING": true,
	"ORDER": true, "LIMIT": true, "OFFSET": true, "UNION": true, "INSERT": true,
	"INTO": true, "VALUES": true, "UPDATE": true, "SET": true, "DELETE": true,
	"JOIN": true, "INNER": true, "LEFT": true, "RIGHT": true, "FULL": true,
	"OUTER": true, "CROSS": true, "ON": true, "WITH": true,
}

// sqlOpenParenKeywords 是后面需要换行的关键字。
var sqlOpenParenKeywords = map[string]bool{
	"SELECT": true, "VALUES": true, "SET": true,
}

// renderSQL 美化 SQL。
func renderSQL(tokens []sqlToken, indent int, keywordCase string) string {
	pad := indentString(indent)
	var buf strings.Builder

	depth := 0
	pendingSpace := false
	lineStart := true

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		if tok.kind == sqlTokenSpace {
			pendingSpace = true
			continue
		}
		if tok.kind == sqlTokenComment {
			// 注释独占一行，保持原样。
			if buf.Len() > 0 && !lineStart {
				buf.WriteString("\n")
			}
			buf.WriteString(tok.text)
			buf.WriteString("\n")
			lineStart = true
			pendingSpace = false
			depth = 0
			continue
		}

		text := tok.text
		if tok.kind == sqlTokenWord {
			text = applyKeywordCase(text, keywordCase)
			upper := strings.ToUpper(text)
			if sqlClauseKeywords[upper] {
				// 子句关键字另起一行。
				if buf.Len() > 0 && !lineStart {
					buf.WriteString("\n")
				}
				lineStart = true
				pendingSpace = false
				buf.WriteString(text)
				depth = 0
				continue
			}
		}

		if pendingSpace && !lineStart {
			buf.WriteString(" ")
		}
		pendingSpace = false

		switch text {
		case "(":
			buf.WriteString(text)
			depth++
			// 括号后换行，让列清单分行显示。
			if sqlOpenParenKeywords[strings.ToUpper(prevWord(tokens, i))] {
				buf.WriteString("\n")
				buf.WriteString(strings.Repeat(pad, depth))
				lineStart = true
			}
		case ")":
			depth--
			if depth < 0 {
				depth = 0
			}
			buf.WriteString(text)
		case ",":
			buf.WriteString(text)
			buf.WriteString("\n")
			buf.WriteString(strings.Repeat(pad, depth))
			lineStart = true
		default:
			buf.WriteString(text)
			lineStart = false
		}
	}

	return strings.TrimRight(buf.String(), "\n")
}

// prevWord 返回前一个 word 类型 token 的大写形式。
func prevWord(tokens []sqlToken, index int) string {
	for i := index - 1; i >= 0; i-- {
		if tokens[i].kind == sqlTokenWord {
			return strings.ToUpper(tokens[i].text)
		}
		if tokens[i].kind != sqlTokenSpace && tokens[i].kind != sqlTokenComment {
			return ""
		}
	}
	return ""
}

// applyKeywordCase 按策略调整关键字大小写。
func applyKeywordCase(text string, keywordCase string) string {
	upper := strings.ToUpper(text)
	if !sqlKeywords[upper] {
		return text
	}
	switch keywordCase {
	case SQLKeywordUpper:
		return upper
	case SQLKeywordLower:
		return strings.ToLower(text)
	default:
		return text
	}
}

// renderSQLMinified 压缩 SQL：去掉所有空白与注释之间的多余空格。
func renderSQLMinified(tokens []sqlToken) string {
	var buf strings.Builder

	for i, tok := range tokens {
		switch tok.kind {
		case sqlTokenSpace:
			continue
		case sqlTokenComment:
			continue
		default:
			if i > 0 && needsSQLSpace(tokens, i) {
				buf.WriteString(" ")
			}
			buf.WriteString(tok.text)
		}
	}

	return buf.String()
}

// needsSQLSpace 判断当前 token 前是否需要空格。
func needsSQLSpace(tokens []sqlToken, index int) bool {
	prev := tokens[index-1]
	if prev.kind == sqlTokenPunct && prev.text == "(" {
		return false
	}
	if tokens[index].kind == sqlTokenPunct && tokens[index].text == ")" {
		return false
	}
	if prev.kind == sqlTokenPunct && prev.text == "," {
		return false
	}
	return true
}
