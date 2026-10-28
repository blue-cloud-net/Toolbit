package encoding

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"io"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// GZip/Deflate 操作。
const (
	// GZipCompress 压缩。
	GZipCompress = "compress"
	// GZipDecompress 解压。
	GZipDecompress = "decompress"
)

// GZip/Deflate 容器格式。
const (
	// GZipFormatGzip 是 gzip 容器（含头部与校验尾）。
	GZipFormatGzip = "gzip"
	// GZipFormatDeflate 是 zlib 容器（RFC 1950）。
	GZipFormatDeflate = "deflate"
	// GZipFormatRaw 是裸 deflate 流（RFC 1951，无头无尾）。
	GZipFormatRaw = "raw"
)

// maxDecompressedBytes 是解压输出的字节上限。
//
// 超过上限返回 TOO_LARGE，用于拦截**解压炸弹**。
//
// 注意：结果最终要放进 JSON 响应，峰值内存约为上限的 1.33 倍
// （原始字节 + base64 文本），因此上限不宜随意放大。
const maxDecompressedBytes = 400 << 20

// GZipDeflateRequest 是 gzip-deflate 工具的请求。
type GZipDeflateRequest struct {
	// Operation 是 compress 或 decompress。
	Operation string `json:"operation"`
	// Format 是容器格式；默认 gzip。
	Format string `json:"format"`
	// Input 是待处理的内容。
	Input tools.Bytes `json:"input"`
	// Level 是压缩级别（0–9）；缺省用压缩库的默认级别。
	Level *int `json:"level"`
	// OutputEncoding 是输出编码；默认 base64。
	OutputEncoding string `json:"outputEncoding"`
}

// GZipDeflateResult 是 gzip-deflate 工具的响应。
type GZipDeflateResult struct {
	// Output 是压缩或解压的结果。
	Output string `json:"output"`
	// Encoding 是输出编码。
	Encoding string `json:"encoding"`
	// Ratio 是输出字节数与输入字节数之比。
	Ratio float64 `json:"ratio"`
}

// GZipDeflate 执行压缩与解压。
func GZipDeflate(_ context.Context, req *GZipDeflateRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	raw, err := req.Input.Decode()
	if err != nil {
		return tools.Result{}, err
	}

	var out []byte
	switch req.Operation {
	case GZipCompress:
		out, err = compressBytes(raw, req.format(), req.level())
	case GZipDecompress:
		out, err = decompressBytes(raw, req.format())
	default:
		return tools.Result{}, tools.InvalidField("operation",
			[]string{GZipCompress, GZipDecompress}, req.Operation)
	}
	if err != nil {
		return tools.Result{}, err
	}

	enc, err := codecsvc.Normalize(req.outputEncoding())
	if err != nil {
		return tools.Result{}, err
	}
	encoded, err := codecsvc.Encode(out, enc)
	if err != nil {
		return tools.Result{}, err
	}

	ratio := 0.0
	if len(raw) > 0 {
		ratio = float64(len(out)) / float64(len(raw))
	}

	return tools.Result{Data: GZipDeflateResult{
		Output:   encoded,
		Encoding: enc,
		Ratio:    ratio,
	}}, nil
}

// format 返回容器格式；缺省为 gzip。
func (r *GZipDeflateRequest) format() string {
	if r.Format == "" {
		return GZipFormatGzip
	}
	return r.Format
}

// level 返回压缩级别；缺省为 -1（库默认）。
func (r *GZipDeflateRequest) level() int {
	if r.Level == nil {
		return -1
	}
	return *r.Level
}

// outputEncoding 返回输出编码；缺省为 base64。
func (r *GZipDeflateRequest) outputEncoding() string {
	if r.OutputEncoding == "" {
		return codecsvc.EncodingBase64
	}
	return r.OutputEncoding
}

// Validate 做字段级校验。
func (r *GZipDeflateRequest) Validate() error {
	switch r.Operation {
	case GZipCompress, GZipDecompress:
	default:
		return tools.InvalidField("operation",
			[]string{GZipCompress, GZipDecompress}, r.Operation)
	}

	switch r.Format {
	case "", GZipFormatGzip, GZipFormatDeflate, GZipFormatRaw:
	default:
		return tools.UnsupportedFormat(r.Format,
			[]string{GZipFormatGzip, GZipFormatDeflate, GZipFormatRaw})
	}

	if r.Level != nil && (*r.Level < 0 || *r.Level > 9) {
		return tools.InvalidField("level", []int{0, 9}, *r.Level)
	}

	if _, err := codecsvc.Normalize(r.outputEncoding()); err != nil {
		return err
	}

	return r.Input.Validate("input")
}

// compressBytes 按指定格式压缩。
func compressBytes(raw []byte, format string, level int) ([]byte, error) {
	var buf bytes.Buffer

	var w io.WriteCloser
	switch format {
	case GZipFormatDeflate:
		zw, err := zlib.NewWriterLevel(&buf, level)
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"field":  "level",
				"reason": "invalid compression level",
			})
		}
		w = zw
	case GZipFormatRaw:
		fw, err := flate.NewWriter(&buf, level)
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"field":  "level",
				"reason": "invalid compression level",
			})
		}
		w = fw
	case GZipFormatGzip, "":
		gw, err := gzip.NewWriterLevel(&buf, level)
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
				"field":  "level",
				"reason": "invalid compression level",
			})
		}
		w = gw
	default:
		return nil, tools.UnsupportedFormat(format,
			[]string{GZipFormatGzip, GZipFormatDeflate, GZipFormatRaw})
	}

	if _, err := w.Write(raw); err != nil {
		_ = w.Close()
		return nil, toolerr.Wrap(toolerr.CodeInternalError, err)
	}
	if err := w.Close(); err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInternalError, err)
	}

	return buf.Bytes(), nil
}

// decompressBytes 按指定格式解压。
//
// 输出超过 maxDecompressedBytes 时返回 TOO_LARGE（解压炸弹防护）。
func decompressBytes(raw []byte, format string) ([]byte, error) {
	var r io.ReadCloser
	switch format {
	case GZipFormatDeflate:
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeDecryptFailed, map[string]any{
				"reason": "invalid zlib stream",
			})
		}
		r = zr
	case GZipFormatRaw:
		r = flate.NewReader(bytes.NewReader(raw))
	case GZipFormatGzip, "":
		gr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeDecryptFailed, map[string]any{
				"reason": "invalid gzip stream",
			})
		}
		r = gr
	default:
		return nil, tools.UnsupportedFormat(format,
			[]string{GZipFormatGzip, GZipFormatDeflate, GZipFormatRaw})
	}
	defer func() { _ = r.Close() }()

	// 多读一个字节用于判断是否恰好触顶。
	limited := io.LimitReader(r, maxDecompressedBytes+1)
	out, err := io.ReadAll(limited)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeDecryptFailed, map[string]any{
			"reason": "invalid compressed stream",
		})
	}
	if len(out) > maxDecompressedBytes {
		return nil, toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"maxBytes": maxDecompressedBytes,
		})
	}

	return out, nil
}
