package data

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 归档格式。
const (
	// ArchiveFormatZip 是 ZIP 归档。
	ArchiveFormatZip = "zip"
	// ArchiveFormatTar 是 TAR 归档。
	ArchiveFormatTar = "tar"
	// ArchiveFormatTarGz 是 TAR.GZ 归档。
	ArchiveFormatTarGz = "tarGz"
	// ArchiveFormatGzip 是单独的 GZIP 文件。
	ArchiveFormatGzip = "gzip"
)

// ArchiveFormat 返回受支持的归档格式（顺序稳定）。
func ArchiveFormat() []string {
	return []string{ArchiveFormatZip, ArchiveFormatTar, ArchiveFormatTarGz, ArchiveFormatGzip}
}

// 归档安全上限。
//
// 这些限制是**必须**的：归档可以声明极大的解压体积，
// 不限制会耗尽内存与磁盘。
const (
	// maxArchiveInputBytes 是输入的字节上限。
	maxArchiveInputBytes = 32 * 1024 * 1024
	// maxArchiveEntries 是条目数上限。
	maxArchiveEntries = 5000
	// maxArchiveEntryBytes 是单条目解压后的字节上限。
	maxArchiveEntryBytes = 64 * 1024 * 1024
	// maxArchiveTotalBytes 是全部条目解压后的总字节上限。
	maxArchiveTotalBytes = 256 * 1024 * 1024
)

// ArchiveInspectRequest 是 archive-inspect 工具的请求。
type ArchiveInspectRequest struct {
	// Data 是归档内容（Base64 编码）。
	Data string `json:"data"`
	// Format 是归档格式；留空自动识别。
	Format string `json:"format"`
	// IncludeContent 是否返回条目内容预览。
	IncludeContent bool `json:"includeContent"`
}

// ArchiveEntry 是一个归档条目。
type ArchiveEntry struct {
	// Name 是条目名。
	Name string `json:"name"`
	// Size 是未压缩大小。
	Size int64 `json:"size"`
	// CompressedSize 是压缩后大小。
	CompressedSize int64 `json:"compressedSize"`
	// IsDir 是否是目录。
	IsDir bool `json:"isDir"`
	// Mode 是文件权限位。
	Mode uint32 `json:"mode"`
	// Modified 是修改时间。
	Modified string `json:"modified"`
	// ContentPreview 是内容预览；不含内容时为空。
	ContentPreview string `json:"contentPreview"`
}

// ArchiveInspectResult 是 archive-inspect 工具的响应。
type ArchiveInspectResult struct {
	// Format 是识别出的归档格式。
	Format string `json:"format"`
	// EntryCount 是条目数。
	EntryCount int `json:"entryCount"`
	// TotalSize 是全部条目未压缩大小之和。
	TotalSize int64 `json:"totalSize"`
	// CompressedSize 是归档本身的字节数。
	CompressedSize int64 `json:"compressedSize"`
	// Entries 是条目列表。
	Entries []ArchiveEntry `json:"entries"`
	// Truncated 是否因超出上限而截断。
	Truncated bool `json:"truncated"`
}

// contentPreviewLimit 是内容预览的字符上限。
const contentPreviewLimit = 512

// ArchiveInspect 查看归档内容与条目清单。
//
// **必须**限制条目数与解压体积：归档可以声明极大的解压体积，
// 不限制会耗尽资源。
func ArchiveInspect(_ context.Context, req *ArchiveInspectRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := decodeArchiveData(req.Data)
	if err != nil {
		return tools.Result{}, err
	}

	format := req.Format
	if format == "" {
		format = detectArchiveFormat(raw)
	}

	result := ArchiveInspectResult{
		Format:         format,
		CompressedSize: int64(len(raw)),
		Entries:        []ArchiveEntry{},
	}

	switch format {
	case ArchiveFormatZip:
		err = readZipEntries(raw, req.IncludeContent, &result)
	case ArchiveFormatTar:
		err = readTarEntries(raw, req.IncludeContent, &result)
	case ArchiveFormatTarGz:
		err = readTarGzEntries(raw, req.IncludeContent, &result)
	case ArchiveFormatGzip:
		err = readGzipEntry(raw, req.IncludeContent, &result)
	default:
		return tools.Result{}, tools.InvalidField("format", ArchiveFormat(), "unrecognized format")
	}
	if err != nil {
		return tools.Result{}, err
	}

	return tools.Result{Data: result}, nil
}

// decodeArchiveData 解码 Base64 数据。
func decodeArchiveData(data string) ([]byte, error) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return nil, tools.MissingField("data")
	}
	raw, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		// 允许 URL-safe 变体。
		raw, err = base64.RawStdEncoding.DecodeString(trimmed)
		if err != nil {
			return nil, tools.InvalidField("data", "Base64 encoded data", "decode error")
		}
	}
	if len(raw) > maxArchiveInputBytes {
		return nil, tools.InvalidField("data", []int{0, maxArchiveInputBytes}, len(raw))
	}
	return raw, nil
}

// detectArchiveFormat 识别归档格式。
func detectArchiveFormat(raw []byte) string {
	switch {
	case len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b:
		return ArchiveFormatGzip
	case len(raw) >= 4 && raw[0] == 'P' && raw[1] == 'K' &&
		(raw[2] == 0x03 || raw[2] == 0x05 || raw[2] == 0x07):
		return ArchiveFormatZip
	default:
		return ArchiveFormatTar
	}
}

// readZipEntries 读取 ZIP 条目。
func readZipEntries(raw []byte, includeContent bool, result *ArchiveInspectResult) error {
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return tools.InvalidField("data", "valid ZIP archive", "parse error")
	}

	total := int64(0)
	for _, file := range reader.File {
		if len(result.Entries) >= maxArchiveEntries {
			result.Truncated = true
			break
		}
		entry := ArchiveEntry{
			Name:           file.Name,
			Size:           int64(file.UncompressedSize64),
			CompressedSize: int64(file.CompressedSize64),
			IsDir:          file.FileInfo().IsDir(),
			Mode:           uint32(file.Mode().Perm()),
			Modified:       file.Modified.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if entry.Size > maxArchiveEntryBytes {
			return tools.InvalidField("data", "archive entry size limit", "entry too large")
		}
		total += entry.Size
		if total > maxArchiveTotalBytes {
			return tools.InvalidField("data", "archive total size limit", "archive too large")
		}
		if includeContent && !entry.IsDir {
			preview, err := readArchiveEntry(file.Open, int64(file.UncompressedSize64))
			if err != nil {
				return err
			}
			entry.ContentPreview = preview
		}
		result.Entries = append(result.Entries, entry)
	}

	result.EntryCount = len(result.Entries)
	result.TotalSize = total
	return nil
}

// readTarEntries 读取 TAR 条目。
func readTarEntries(raw []byte, includeContent bool, result *ArchiveInspectResult) error {
	return readTarStream(bytes.NewReader(raw), includeContent, result)
}

// readTarGzEntries 读取 TAR.GZ 条目。
func readTarGzEntries(raw []byte, includeContent bool, result *ArchiveInspectResult) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return tools.InvalidField("data", "valid GZIP archive", "parse error")
	}
	defer func() { _ = gzipReader.Close() }()

	// 限制解压后的总量。
	limited := io.LimitReader(gzipReader, maxArchiveTotalBytes+1)
	return readTarStream(limited, includeContent, result)
}

// readTarStream 从 TAR 流读取条目。
func readTarStream(reader io.Reader, includeContent bool, result *ArchiveInspectResult) error {
	tarReader := tar.NewReader(reader)
	total := int64(0)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return tools.InvalidField("data", "valid TAR archive", "parse error")
		}
		if len(result.Entries) >= maxArchiveEntries {
			result.Truncated = true
			break
		}

		entry := ArchiveEntry{
			Name:     header.Name,
			Size:     header.Size,
			IsDir:    header.Typeflag == tar.TypeDir,
			Mode:     uint32(header.Mode & 0o777),
			Modified: header.ModTime.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if entry.Size > maxArchiveEntryBytes {
			return tools.InvalidField("data", "archive entry size limit", "entry too large")
		}
		total += entry.Size
		if total > maxArchiveTotalBytes {
			return tools.InvalidField("data", "archive total size limit", "archive too large")
		}
		if includeContent && !entry.IsDir {
			preview, err := readArchiveEntry(func() (io.ReadCloser, error) {
				return io.NopCloser(tarReader), nil
			}, entry.Size)
			if err != nil {
				return err
			}
			entry.ContentPreview = preview
		}
		result.Entries = append(result.Entries, entry)
	}

	result.EntryCount = len(result.Entries)
	result.TotalSize = total
	return nil
}

// readGzipEntry 读取单个 GZIP 文件的条目。
func readGzipEntry(raw []byte, includeContent bool, result *ArchiveInspectResult) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return tools.InvalidField("data", "valid GZIP file", "parse error")
	}
	defer func() { _ = gzipReader.Close() }()

	name := gzipReader.Name
	if name == "" {
		name = "content"
	}

	preview := ""
	size := int64(0)
	if includeContent {
		preview, size, err = readArchiveEntryWithSize(func() (io.ReadCloser, error) {
			return gzipReader, nil
		}, 0)
		if err != nil {
			return err
		}
	} else {
		// 只统计大小，不保留内容。
		size, err = io.Copy(io.Discard, io.LimitReader(gzipReader, maxArchiveTotalBytes+1))
		if err != nil {
			return tools.InvalidField("data", "valid GZIP file", "parse error")
		}
		if size > maxArchiveTotalBytes {
			return tools.InvalidField("data", "archive total size limit", "archive too large")
		}
	}

	result.Entries = append(result.Entries, ArchiveEntry{
		Name:           name,
		Size:           size,
		CompressedSize: int64(len(raw)),
		ContentPreview: preview,
	})
	result.EntryCount = 1
	result.TotalSize = size
	return nil
}

// readArchiveEntry 读取条目内容并生成预览。
func readArchiveEntry(open func() (io.ReadCloser, error), size int64) (string, error) {
	preview, _, err := readArchiveEntryWithSize(open, size)
	return preview, err
}

// readArchiveEntryWithSize 读取条目内容并返回预览与实际大小。
func readArchiveEntryWithSize(open func() (io.ReadCloser, error), size int64) (string, int64, error) {
	closer, err := open()
	if err != nil {
		return "", 0, tools.InvalidField("data", "readable archive entry", "read error")
	}
	defer func() { _ = closer.Close() }()

	buffer := make([]byte, contentPreviewLimit)
	read, err := io.ReadFull(closer, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", 0, tools.InvalidField("data", "readable archive entry", "read error")
	}

	preview := string(buffer[:read])
	// 剩余部分只统计大小。
	remaining, err := io.Copy(io.Discard, io.LimitReader(closer, maxArchiveEntryBytes+1))
	if err != nil {
		return "", 0, tools.InvalidField("data", "readable archive entry", "read error")
	}

	total := int64(read) + remaining
	if total > maxArchiveEntryBytes {
		return "", 0, tools.InvalidField("data", "archive entry size limit", "entry too large")
	}
	return preview, total, nil
}

// Validate 做字段级校验。
func (r *ArchiveInspectRequest) Validate() error {
	if strings.TrimSpace(r.Data) == "" {
		return tools.MissingField("data")
	}
	if r.Format != "" && !containsString(ArchiveFormat(), r.Format) {
		return tools.InvalidField("format", ArchiveFormat(), r.Format)
	}
	return nil
}
