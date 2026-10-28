package tools

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// multipartFieldLimit 是单个表单字段的最大字节数。
//
// 小字段（operation / mode / key …）本就该很短；
// 超过上限直接报错，避免把大块内容塞进内存。
const multipartFieldLimit = 1 << 20

// MultipartForm 是 multipart 表单的分发器辅助器。
//
// 它把 `multipart.Reader` 的「顺序读取」语义包装成
// 「先收小字段、再收文件」的顺序，工具只需按直觉调用。
type MultipartForm struct {
	// reader 是底层读取器。
	reader *multipart.Reader
	// fields 是已读入的小字段。
	fields map[string]string
	// fileName 是文件部分的字段名；文件**必须**是最后一个 part。
	fileName string
	// filePart 是文件部分本身。
	//
	// 用 `multipart.Part` 而非 `*multipart.FileHeader`：
	// 后者只能由 `ParseMultipartForm` 产出，而那会把大文件**落盘**。
	filePart *multipart.Part
	// fileOpened 标记文件是否已打开。
	fileOpened bool
}

// NewMultipartForm 构造表单辅助器。
func NewMultipartForm(r *multipart.Reader) *MultipartForm {
	return &MultipartForm{
		reader: r,
		fields: make(map[string]string, 8),
	}
}

// ReadFields 读取所有小字段，直到遇到文件部分或表单结束。
//
// 遇到文件部分时**停止**：文件必须是最后一个 part，
// 否则后续字段无法再取。
func (f *MultipartForm) ReadFields() error {
	for {
		part, err := f.reader.NextPart()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"reason": "malformed multipart body",
			})
		}

		name := part.FormName()
		if !IsFilePart(part) {
			value, err := readMultipartField(part)
			part.Close()
			if err != nil {
				return err
			}
			f.fields[name] = value
			continue
		}

		// 文件部分：记住它，读取推迟到工具真正需要时。
		f.fileName = name
		f.filePart = part
		return nil
	}
}

// Value 返回小字段的值；不存在时返回空串。
func (f *MultipartForm) Value(name string) string {
	return f.fields[name]
}

// HasFile 报告表单是否携带文件部分。
func (f *MultipartForm) HasFile() bool {
	return f.filePart != nil
}

// Filename 返回文件部分的原始文件名。
//
// 该值**不可信**（由客户端提供），只可用于回显，
// 使用前**必须**清洗（见 SanitizeFilename）。
func (f *MultipartForm) Filename() string {
	if f.filePart == nil {
		return ""
	}
	return f.filePart.FileName()
}

// FileSize 返回文件部分的字节数。
//
// 流式读取时长度未知，恒为 -1。
func (f *MultipartForm) FileSize() int64 {
	if f.filePart == nil {
		return -1
	}
	return -1
}

// OpenFile 打开文件内容。
//
// 调用方**必须**关闭返回的读取器。
func (f *MultipartForm) OpenFile() (io.ReadCloser, error) {
	if f.filePart == nil {
		return nil, toolerr.Newf(toolerr.CodeMissingField, map[string]any{"field": "file"})
	}
	if f.fileOpened {
		return nil, toolerr.New(toolerr.CodeInternalError)
	}
	f.fileOpened = true
	return f.filePart, nil
}

// Close 关闭已打开的文件。
func (f *MultipartForm) Close() error {
	if f.filePart == nil || !f.fileOpened {
		return nil
	}
	err := f.filePart.Close()
	f.fileOpened = false
	return err
}

// readMultipartField 读取一个小字段的内容。
func readMultipartField(part *multipart.Part) (string, error) {
	limited := io.LimitReader(part, multipartFieldLimit+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return "", toolerr.Wrap(toolerr.CodeInvalidField, err)
	}
	if len(raw) > multipartFieldLimit {
		return "", toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    part.FormName(),
			"reason":   "form field too large",
			"maxBytes": multipartFieldLimit,
		})
	}
	return string(raw), nil
}

// IsFilePart 判断 part 是否是文件部分。
//
// 不能只看 `part.FileName()`：普通字段与「文件名为空的文件」
// 都返回空串。必须看 Content-Disposition 里**有没有** filename 参数。
func IsFilePart(p *multipart.Part) bool {
	_, params, err := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
	if err != nil {
		return false
	}
	_, ok := params["filename"]
	return ok
}

// SanitizeFilename 清洗客户端提供的文件名。
//
// 只保留基础名与安全字符，去掉路径分隔符与控制字符，
// 避免响应头注入与路径穿越。
func SanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	// 只取最后一段：去掉客户端可能带的目录。
	if idx := strings.LastIndexAny(name, `/\`); idx >= 0 {
		name = name[idx+1:]
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 0x20 || r == 0x7f:
			// 控制字符：丢弃。
		case r == '"' || r == '\\' || r == '/' || r == '?' || r == '*' || r == '<' || r == '>' || r == '|':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" || out == "." || out == ".." {
		return ""
	}
	if len(out) > 128 {
		out = out[:128]
	}
	return out
}

// ContentDisposition 构造带 UTF-8 文件名的 Content-Disposition 头值。
//
// filename 用 ASCII 回退（去非 ASCII 字符），
// filename* 用 RFC 5987 的 UTF-8 百分号编码，
// 兼容只认 ASCII 的旧客户端。
func ContentDisposition(filename string) string {
	ascii := asciiFallback(filename)
	if ascii == "" {
		return ""
	}
	return fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", ascii, percentEncodeUTF8(filename))
}

// asciiFallback 去掉非 ASCII 与控制字符，得到 ASCII 回退名。
func asciiFallback(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r < 0x20 || r > 0x7e {
			continue
		}
		if r == '"' || r == '\\' {
			b.WriteRune('_')
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// percentEncodeUTF8 按 RFC 5987 编码文件名。
func percentEncodeUTF8(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for _, c := range []byte(s) {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9',
			c == '-', c == '_', c == '.', c == '!', c == '#', c == '$', c == '&',
			c == '+', c == '^', c == '`', c == '|', c == '~':
			b.WriteByte(c)
		default:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f])
		}
	}
	return b.String()
}

// NewMultipartReader 从请求构造流式的 multipart 读取器。
//
// 用 `Request.MultipartReader` 而不是 `ParseMultipartForm`：
// 后者会把超限的文件**落盘**到临时目录，违反「禁止落盘」的约束。
func NewMultipartReader(r *http.Request) (*multipart.Reader, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, toolerr.Wrap(toolerr.CodeUnsupportedMediaType, err)
	}
	return reader, nil
}
