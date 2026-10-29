package data

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// FileAnalyzeRequest 是 file-analyze 工具的请求。
type FileAnalyzeRequest struct {
	// Data 是文件内容（Base64 编码）。
	Data string `json:"data"`
	// FileName 是文件名；留空时从魔数推断类型。
	FileName string `json:"fileName"`
	// DigestAlgorithm 是摘要算法；默认 SHA256。
	DigestAlgorithm string `json:"digestAlgorithm"`
}

// FileAnalyzeResult 是 file-analyze 工具的响应。
type FileAnalyzeResult struct {
	// FileName 是文件名。
	FileName string `json:"fileName"`
	// Size 是字节数。
	Size int `json:"size"`
	// MimeType 是推断出的媒体类型。
	MimeType string `json:"mimeType"`
	// Extension 是推断出的扩展名。
	Extension string `json:"extension"`
	// Description 是类型说明。
	Description string `json:"description"`
	// IsTextual 是否是文本类文件。
	IsTextual bool `json:"isTextual"`
	// MagicHex 是文件头部的十六进制。
	MagicHex string `json:"magicHex"`
	// Digest 是文件摘要。
	Digest string `json:"digest"`
	// DigestAlgorithm 是摘要算法。
	DigestAlgorithm string `json:"digestAlgorithm"`
}

// magicSignature 是魔数签名。
type magicSignature struct {
	// magic 是魔数字节。
	magic []byte
	// offset 是魔数在文件中的偏移。
	offset int
	// mimeType 是媒体类型。
	mimeType string
	// extension 是扩展名。
	extension string
	// description 是类型说明。
	description string
	// textual 是否是文本类。
	textual bool
}

// magicSignatures 是内置的魔数签名表。
//
// 数据是**内置快照**：不读取系统 MIME 数据库，也不联网。
var magicSignatures = []magicSignature{
	{magic: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, mimeType: "image/png", extension: "png", description: "PNG 图像"},
	{magic: []byte{0xFF, 0xD8, 0xFF}, mimeType: "image/jpeg", extension: "jpg", description: "JPEG 图像"},
	{magic: []byte{'G', 'I', 'F', '8'}, mimeType: "image/gif", extension: "gif", description: "GIF 图像"},
	{magic: []byte{0x25, 0x50, 0x44, 0x46}, mimeType: "application/pdf", extension: "pdf", description: "PDF 文档"},
	{magic: []byte{'P', 'K', 0x03, 0x04}, mimeType: "application/zip", extension: "zip", description: "ZIP 压缩包"},
	{magic: []byte{0x1F, 0x8B}, mimeType: "application/gzip", extension: "gz", description: "GZIP 压缩包"},
	{magic: []byte{'B', 'Z', 'h'}, mimeType: "application/x-bzip2", extension: "bz2", description: "BZIP2 压缩包"},
	{magic: []byte{0xFD, '7', 'z', 'X', 'Z', 0x00}, mimeType: "application/x-xz", extension: "xz", description: "XZ 压缩包"},
	{magic: []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}, mimeType: "application/x-7z-compressed", extension: "7z", description: "7-Zip 压缩包"},
	{magic: []byte{'R', 'I', 'F', 'F'}, mimeType: "audio/wav", extension: "wav", description: "WAV 音频"},
	{magic: []byte{0x4F, 0x67, 0x67, 0x53}, mimeType: "audio/ogg", extension: "ogg", description: "OGG 音频"},
	{magic: []byte{0x1A, 0x45, 0xDF, 0xA3}, mimeType: "video/webm", extension: "webm", description: "Matroska / WebM 视频"},
	{magic: []byte{0x00, 0x00, 0x01, 0x00}, mimeType: "image/vnd.microsoft.icon", extension: "ico", description: "Windows 图标"},
	{magic: []byte{'P', 'K', 0x03, 0x04}, offset: 30, mimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", extension: "docx", description: "Word 文档（Open XML）"},
	{magic: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, mimeType: "application/vnd.ms-excel", extension: "xls", description: "旧版 Excel 工作簿"},
	{magic: []byte{'%', 'P', 'D', 'F'}, mimeType: "application/pdf", extension: "pdf", description: "PDF 文档（文本头）"},
}

// textHeads 是文本类文件的判定前缀。
var textHeads = []string{
	"<?xml", "<!DOCTYPE", "<html", "{", "[",
}

// maxFileAnalyzeBytes 是输入的字节上限。
const maxFileAnalyzeBytes = 32 * 1024 * 1024

// FileAnalyze 识别文件类型并提取元数据。
//
// 摘要经 cryptosvc 取；不读取系统 MIME 数据库，也不联网。
func FileAnalyze(_ context.Context, req *FileAnalyzeRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := decodeArchiveData(req.Data)
	if err != nil {
		return tools.Result{}, err
	}

	algorithm := req.DigestAlgorithm
	if algorithm == "" {
		algorithm = "SHA256"
	}
	digest, err := cryptosvc.Digest(algorithm, raw)
	if err != nil {
		return tools.Result{}, err
	}

	result := FileAnalyzeResult{
		FileName:        req.FileName,
		Size:            len(raw),
		Digest:          hexLower(digest),
		DigestAlgorithm: algorithm,
		MagicHex:        hexLower(headBytes(raw, 16)),
	}

	if signature, ok := matchMagicSignature(raw); ok {
		result.MimeType = signature.mimeType
		result.Extension = signature.extension
		result.Description = signature.description
		result.IsTextual = signature.textual
	} else if isTextualContent(raw) {
		result.MimeType = "text/plain"
		result.Extension = "txt"
		result.Description = "纯文本"
		result.IsTextual = true
	} else {
		result.MimeType = "application/octet-stream"
		result.Extension = "bin"
		result.Description = "二进制数据（未识别已知类型）"
		result.IsTextual = false
	}

	if result.FileName == "" {
		result.FileName = "file." + result.Extension
	}

	return tools.Result{Data: result}, nil
}

// matchMagicSignature 匹配魔数签名。
func matchMagicSignature(raw []byte) (magicSignature, bool) {
	for _, signature := range magicSignatures {
		offset := signature.offset
		if offset+len(signature.magic) > len(raw) {
			continue
		}
		matched := true
		for i, b := range signature.magic {
			if raw[offset+i] != b {
				matched = false
				break
			}
		}
		if matched {
			return signature, true
		}
	}
	return magicSignature{}, false
}

// isTextualContent 判断是否为文本内容。
func isTextualContent(raw []byte) bool {
	if len(raw) == 0 {
		return false
	}
	head := string(headBytes(raw, 64))
	for _, prefix := range textHeads {
		if strings.HasPrefix(head, prefix) {
			return true
		}
	}
	// 没有 NUL 字节且大部分是可打印字符时判为文本。
	printable := 0
	for _, b := range head {
		if b == 0 {
			return false
		}
		if b >= 0x20 || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}
	return printable*100/len(head) >= 90
}

// headBytes 取文件头部字节。
func headBytes(raw []byte, count int) []byte {
	if len(raw) < count {
		return raw
	}
	return raw[:count]
}

// hexLower 把字节转成小写十六进制。
func hexLower(data []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, 0, len(data)*2)
	for _, b := range data {
		out = append(out, digits[b>>4], digits[b&0x0f])
	}
	return string(out)
}

// Validate 做字段级校验。
func (r *FileAnalyzeRequest) Validate() error {
	if strings.TrimSpace(r.Data) == "" {
		return tools.MissingField("data")
	}
	if len(r.FileName) > 255 {
		return tools.InvalidField("fileName", []int{0, 255}, len(r.FileName))
	}
	if len(r.Data) > maxFileAnalyzeBytes {
		return tools.InvalidField("data", []int{0, maxFileAnalyzeBytes}, len(r.Data))
	}
	if r.DigestAlgorithm != "" {
		supported := cryptosvc.DigestAlgorithms()
		if !containsString(supported, r.DigestAlgorithm) {
			return tools.InvalidField("digestAlgorithm", supported, r.DigestAlgorithm)
		}
	}
	return nil
}
