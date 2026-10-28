package documents

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	"golang.org/x/net/html/charset"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 附件与正文的大小上限。
const (
	// defaultMaxAttachmentBytes 是单附件的默认字节上限。
	defaultMaxAttachmentBytes = 10 << 20
	// maxAttachmentBytesLimit 是单附件字节上限的绝对上限。
	maxAttachmentBytesLimit = 50 << 20
	// maxBodyBytes 是单个正文分部的字节上限。
	maxBodyBytes = 10 << 20
	// maxHeaderCount 是头字段条数上限。
	maxHeaderCount = 500
)

// EMLParserRequest 是 eml-parser 工具的请求。
type EMLParserRequest struct {
	// Input 是邮件源文。
	Input EMLInput `json:"input"`
	// MaxAttachmentBytes 是单附件的字节上限；缺省 10 MB。
	MaxAttachmentBytes int `json:"maxAttachmentBytes"`
}

// EMLInput 是邮件输入。
type EMLInput struct {
	// Text 是 .eml 源文。
	Text string `json:"text"`
	// Encoding 是源文的字符编码；缺省按 UTF-8 处理。
	Encoding string `json:"encoding"`
}

// EMLHeader 是一个头字段。
type EMLHeader struct {
	// Name 是头字段名。
	Name string `json:"name"`
	// Value 是头字段值。
	Value string `json:"value"`
}

// EMLBody 是一个正文分部。
type EMLBody struct {
	// ContentType 是内容类型。
	ContentType string `json:"contentType"`
	// Charset 是字符编码。
	Charset string `json:"charset"`
	// Text 是正文文本。
	Text string `json:"text"`
	// HTML 是 HTML 正文（仅 `text/html` 分部）。
	HTML string `json:"html,omitempty"`
}

// EMLAttachment 是一个附件。
type EMLAttachment struct {
	// Filename 是文件名。
	Filename string `json:"filename"`
	// MIME 是媒体类型。
	MIME string `json:"mime"`
	// Size 是字节数。
	Size int `json:"size"`
	// RawBase64 是附件内容（base64）；超限时不返回。
	RawBase64 string `json:"rawBase64,omitempty"`
	// Truncated 为真表示内容因超限被丢弃。
	Truncated bool `json:"truncated"`
}

// EMLParserResult 是 eml-parser 工具的响应。
type EMLParserResult struct {
	// Headers 是原始头字段。
	Headers []EMLHeader `json:"headers"`
	// HeaderMap 是头字段映射（小写键）。
	HeaderMap map[string]string `json:"headerMap"`
	// Bodies 是正文分部。
	Bodies []EMLBody `json:"bodies"`
	// Attachments 是附件。
	Attachments []EMLAttachment `json:"attachments"`
}

// EMLParser 解析 EML 邮件（RFC 5322）。
//
// 本版**只**支持文本输入：multipart 上传依赖分发器的
// multipart 支持，本轮未启用。
func EMLParser(ctx context.Context, req *EMLParserRequest) (EMLParserResult, error) {
	if err := req.Validate(); err != nil {
		return EMLParserResult{}, err
	}

	raw := []byte(req.Input.Text)

	message, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return EMLParserResult{}, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "input.text",
			"reason": "not a valid EML message",
		})
	}

	result := EMLParserResult{
		Headers:     []EMLHeader{},
		HeaderMap:   map[string]string{},
		Bodies:      []EMLBody{},
		Attachments: []EMLAttachment{},
	}

	for _, header := range parseOrderedHeaders(raw) {
		if len(result.Headers) >= maxHeaderCount {
			break
		}
		result.Headers = append(result.Headers, header)
		key := strings.ToLower(header.Name)
		if _, exists := result.HeaderMap[key]; !exists {
			result.HeaderMap[key] = header.Value
		}
	}

	if err := collectParts(message, &result, req.maxAttachmentBytes()); err != nil {
		return EMLParserResult{}, err
	}

	return result, nil
}

// Validate 做字段级校验。
func (r *EMLParserRequest) Validate() error {
	if strings.TrimSpace(r.Input.Text) == "" {
		return tools.MissingField("input.text")
	}
	if r.MaxAttachmentBytes < 0 || r.MaxAttachmentBytes > maxAttachmentBytesLimit {
		return tools.InvalidField("maxAttachmentBytes",
			[]int{0, maxAttachmentBytesLimit}, r.MaxAttachmentBytes)
	}
	return nil
}

// maxAttachmentBytes 返回单附件字节上限；缺省 10 MB。
func (r *EMLParserRequest) maxAttachmentBytes() int {
	if r.MaxAttachmentBytes <= 0 {
		return defaultMaxAttachmentBytes
	}
	return r.MaxAttachmentBytes
}

// collectParts 递归收集正文与附件。
func collectParts(message *mail.Message, result *EMLParserResult, maxAttachment int) error {
	contentType := message.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		mediaType = "text/plain"
		params = map[string]string{}
	}

	switch {
	case strings.HasPrefix(mediaType, "multipart/"):
		return collectMultipart(message, result, maxAttachment, params["boundary"])

	case strings.HasPrefix(mediaType, "text/"):
		body, err := readLimited(message.Body, maxBodyBytes)
		if err != nil {
			return err
		}
		text, err := decodeCharset(body, params["charset"])
		if err != nil {
			return err
		}
		entry := EMLBody{
			ContentType: mediaType,
			Charset:     params["charset"],
			Text:        text,
		}
		if mediaType == "text/html" {
			entry.HTML = text
		}
		result.Bodies = append(result.Bodies, entry)
		return nil

	default:
		body, err := readLimited(message.Body, maxAttachment+1)
		if err != nil {
			return err
		}
		result.Attachments = append(result.Attachments, buildAttachment(
			"", mediaType, body, maxAttachment))
		return nil
	}
}

// collectMultipart 收集 multipart 邮件的各个分部。
func collectMultipart(message *mail.Message, result *EMLParserResult, maxAttachment int, boundary string) error {
	if boundary == "" {
		return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "input.text",
			"reason": "multipart without boundary",
		})
	}

	reader := multipart.NewReader(message.Body, boundary)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"field":  "input.text",
				"reason": "malformed multipart body",
			})
		}

		partType, partParams, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil || partType == "" {
			partType = "text/plain"
			partParams = map[string]string{}
		}

		disposition, dispositionParams, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if err != nil {
			disposition = ""
			dispositionParams = map[string]string{}
		}

		isAttachment := strings.EqualFold(disposition, "attachment") ||
			partParams["name"] != "" && dispositionParams["filename"] == ""

		switch {
		case isAttachment:
			body, err := readLimited(part, maxAttachment+1)
			part.Close()
			if err != nil {
				return err
			}
			result.Attachments = append(result.Attachments, buildAttachment(
				dispositionParams["filename"], partType, body, maxAttachment))

		case strings.HasPrefix(partType, "multipart/"):
			inner := &mail.Message{
				Header: mail.Header(part.Header),
				Body:   part,
			}
			if err := collectParts(inner, result, maxAttachment); err != nil {
				part.Close()
				return err
			}
			part.Close()

		case strings.HasPrefix(partType, "text/"):
			body, err := readLimited(part, maxBodyBytes)
			part.Close()
			if err != nil {
				return err
			}
			text, err := decodeCharset(body, partParams["charset"])
			if err != nil {
				return err
			}
			entry := EMLBody{
				ContentType: partType,
				Charset:     partParams["charset"],
				Text:        text,
			}
			if partType == "text/html" {
				entry.HTML = text
			}
			result.Bodies = append(result.Bodies, entry)

		default:
			body, err := readLimited(part, maxAttachment+1)
			part.Close()
			if err != nil {
				return err
			}
			result.Attachments = append(result.Attachments, buildAttachment(
				dispositionParams["filename"], partType, body, maxAttachment))
		}
	}
}

// buildAttachment 构造附件条目。
//
// 超过上限时**不**返回内容，只返回元信息并标记 truncated。
func buildAttachment(filename string, mediaType string, body []byte, maxAttachment int) EMLAttachment {
	attachment := EMLAttachment{
		Filename: filename,
		MIME:     mediaType,
		Size:     len(body),
	}
	if len(body) > maxAttachment {
		attachment.Size = len(body)
		attachment.Truncated = true
		return attachment
	}
	attachment.RawBase64 = base64.StdEncoding.EncodeToString(body)
	return attachment
}

// readLimited 读取至多 limit 字节。
//
// 超出时按 limit 截断并返回错误，由调用方决定如何处理。
func readLimited(r io.Reader, limit int) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(r, int64(limit)))
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "input.text",
			"reason": "cannot read message body",
		})
	}
	return raw, nil
}

// decodeCharset 按指定编码解码正文。
//
// 缺省或未知编码时按 UTF-8 处理；无法解码的字节替换为 U+FFFD，
// **不**报错：邮件正文的编码声明经常不可靠。
func decodeCharset(raw []byte, name string) (string, error) {
	if name == "" {
		return string(raw), nil
	}
	decoded, err := charset.NewReaderLabel(name, bytes.NewReader(raw))
	if err != nil {
		return string(raw), nil
	}
	out, err := io.ReadAll(decoded)
	if err != nil {
		return string(raw), nil
	}
	return string(out), nil
}

// sanitizeHeaderValue 清理头字段值里的换行。
//
// 头字段里的裸换行会被邮件客户端解释成新头字段，
// 回显时必须折叠成空格。
func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

// parseOrderedHeaders 按出现顺序解析头字段。
//
// `net/mail` 的头字段是 map，**丢失顺序**；而头字段顺序对
// 「原始头字段」列表有意义（例如 Received 链），因此自己解析。
func parseOrderedHeaders(raw []byte) []EMLHeader {
	block := raw
	if idx := bytes.Index(raw, []byte("\r\n\r\n")); idx >= 0 {
		block = raw[:idx]
	} else if idx := bytes.Index(raw, []byte("\n\n")); idx >= 0 {
		block = raw[:idx]
	}

	headers := []EMLHeader{}
	decoder := &mime.WordDecoder{}

	for _, rawLine := range strings.Split(string(block), "\n") {
		line := strings.TrimSuffix(rawLine, "\r")
		if line == "" {
			break
		}

		// 续行：以空白开头，接在上一个头字段的��面。
		if line[0] == ' ' || line[0] == '\t' {
			if len(headers) == 0 {
				continue
			}
			headers[len(headers)-1].Value += " " + strings.TrimSpace(line)
			continue
		}

		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		headers = append(headers, EMLHeader{
			Name:  strings.TrimSpace(name),
			Value: decodeHeaderValue(decoder, sanitizeHeaderValue(value)),
		})
	}

	return headers
}

// decodeHeaderValue 解码 RFC 2047 编码字。
//
// 解码失败时保留原值：邮件里的编码声明经常不可靠。
func decodeHeaderValue(decoder *mime.WordDecoder, value string) string {
	if !strings.Contains(value, "=?") {
		return value
	}
	decoded, err := decoder.DecodeHeader(value)
	if err != nil {
		return value
	}
	return decoded
}
