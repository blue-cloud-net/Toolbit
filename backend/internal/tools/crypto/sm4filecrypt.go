package crypto

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"strconv"

	"github.com/blue-cloud-net/Toolbit/backend/internal/codecsvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// SM4 文件加解密的操作类型。
const (
	// FileOpEncrypt 是加密。
	FileOpEncrypt = "encrypt"
	// FileOpDecrypt 是解密。
	FileOpDecrypt = "decrypt"
)

// 表单字段名。
const (
	// fileFieldOperation 是操作字段名。
	fileFieldOperation = "operation"
	// fileFieldMode 是模式字段名。
	fileFieldMode = "mode"
	// fileFieldKey 是密钥字段名。
	fileFieldKey = "key"
	// fileFieldIV 是 nonce 字段名。
	fileFieldIV = "iv"
	// fileFieldChunkSize 是分块大小字段名。
	fileFieldChunkSize = "chunkSize"
	// fileFieldFile 是文件字段名。
	fileFieldFile = "file"
)

// SM4FileCryptRequest 是 SM4 文件加解密的请求。
//
// 请求体是 multipart/form-data：表单字段 + 一个文件部分。
// 文件部分**必须**是最后一个 part。
type SM4FileCryptRequest struct {
	// Operation 是 encrypt 或 decrypt。
	Operation string
	// Mode 是对称模式；本工具只支持 GCM。
	Mode string
	// Key 是对称密钥（十六进制或 base64 的 JSON 串）。
	Key []byte
	// IV 是基础 nonce；为空时加密时随机生成。
	IV []byte
	// ChunkSize 是分块大小（字节）。
	ChunkSize uint32
	// Filename 是上传文件的原始文件名（**不可信**，仅用于回显）。
	Filename string
	// form 是已绑定的表单。
	form *tools.MultipartForm
	// file 是已打开的文件内容。
	file io.ReadCloser
}

// BindMultipart 绑定 multipart 表单。
func (r *SM4FileCryptRequest) BindMultipart(reader *multipart.Reader) error {
	form := tools.NewMultipartForm(reader)
	if err := form.ReadFields(); err != nil {
		return err
	}
	r.form = form

	r.Operation = form.Value(fileFieldOperation)
	r.Mode = form.Value(fileFieldMode)
	r.Filename = form.Filename()

	key, err := decodeJSONByteField(form.Value(fileFieldKey), fileFieldKey)
	if err != nil {
		return err
	}
	r.Key = key

	iv, err := decodeJSONByteField(form.Value(fileFieldIV), fileFieldIV)
	if err != nil {
		return err
	}
	r.IV = iv

	if raw := form.Value(fileFieldChunkSize); raw != "" {
		size, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return tools.InvalidField(fileFieldChunkSize, "unsigned integer", raw)
		}
		r.ChunkSize = uint32(size)
	}

	if !form.HasFile() {
		return tools.MissingField(fileFieldFile)
	}
	file, err := form.OpenFile()
	if err != nil {
		return err
	}
	r.file = file
	return nil
}

// Validate 做字段级校验。
func (r *SM4FileCryptRequest) Validate() error {
	switch r.Operation {
	case FileOpEncrypt, FileOpDecrypt:
	default:
		return tools.InvalidField(fileFieldOperation, []string{FileOpEncrypt, FileOpDecrypt}, r.Operation)
	}

	if r.Mode != "" && r.Mode != cryptosvc.ModeGCM {
		return tools.UnsupportedMode(r.Mode, []string{cryptosvc.ModeGCM})
	}

	if len(r.Key) == 0 {
		return tools.MissingField(fileFieldKey)
	}
	if len(r.Key) != 16 {
		return toolerr.Newf(toolerr.CodeKeySizeMismatch, map[string]any{
			"field":    fileFieldKey,
			"expected": 16,
			"actual":   len(r.Key),
		})
	}

	if len(r.IV) != 0 && len(r.IV) != cryptosvc.TBFCBaseNonceSize {
		return toolerr.Newf(toolerr.CodeIVSizeMismatch, map[string]any{
			"field":    fileFieldIV,
			"expected": cryptosvc.TBFCBaseNonceSize,
			"actual":   len(r.IV),
		})
	}

	if r.ChunkSize != 0 &&
		(r.ChunkSize < cryptosvc.TBFCMinChunkSize || r.ChunkSize > cryptosvc.TBFCMaxChunkSize) {
		return tools.InvalidField(fileFieldChunkSize,
			[]int{cryptosvc.TBFCMinChunkSize, cryptosvc.TBFCMaxChunkSize}, r.ChunkSize)
	}
	return nil
}

// Close 释放已打开的文件。
func (r *SM4FileCryptRequest) Close() error {
	if r.file != nil {
		err := r.file.Close()
		r.file = nil
		return err
	}
	if r.form != nil {
		return r.form.Close()
	}
	return nil
}

// SM4FileCrypt 执行 SM4 文件级加解密。
//
// 输入是上传的文件，输出是二进制流：
//   - 加密 → TBFC 容器；
//   - 解密 → 原始明文。
//
// ⚠️ 文件句柄的所有权交给**流式 goroutine**：
// handler 返回时输出还没写完，此时关闭文件会让 goroutine 读到已关闭的流。
func SM4FileCrypt(_ context.Context, req *SM4FileCryptRequest) (tools.MultipartResult, error) {
	if err := req.Validate(); err != nil {
		_ = req.Close()
		return tools.MultipartResult{}, err
	}

	if req.Operation == FileOpDecrypt {
		return decryptFile(req)
	}
	return encryptFile(req)
}

// encryptFile 加密文件。
func encryptFile(req *SM4FileCryptRequest) (tools.MultipartResult, error) {
	reader, writer := io.Pipe()

	go func() {
		defer func() { _ = req.Close() }()
		// 管道写端出错时用 CloseWithError 把错误传给读端。
		err := streamEncrypt(req, writer)
		_ = writer.CloseWithError(err)
	}()

	return tools.MultipartResult{
		ContentType: "application/octet-stream",
		Filename:    encryptedFilename(req.Filename),
		Body:        reader,
	}, nil
}

// streamEncrypt 把上传的文件加密成容器并写入管道。
func streamEncrypt(req *SM4FileCryptRequest, out *io.PipeWriter) error {
	chunkSize := req.ChunkSize
	if chunkSize == 0 {
		chunkSize = cryptosvc.TBFCDefaultChunkSize
	}

	container, err := cryptosvc.NewTBFCWriter(out, cryptosvc.SymSM4, req.Key, req.IV, chunkSize)
	if err != nil {
		return err
	}
	if _, err := io.Copy(container, req.file); err != nil {
		return err
	}
	return container.Close()
}

// decryptFile 解密文件。
//
// 解密**先同步**读容器头并校验第一块：
// 「密钥不对」「文件不是容器」这类错误因此能在响应头发出**之前**
// 被发现，返回干净的 JSON 错误，而不是 200 + 截断的下载。
func decryptFile(req *SM4FileCryptRequest) (tools.MultipartResult, error) {
	container, err := cryptosvc.NewTBFCReader(req.file, cryptosvc.SymSM4, req.Key)
	if err != nil {
		_ = req.Close()
		return tools.MultipartResult{}, err
	}
	if err := container.Probe(); err != nil {
		_ = req.Close()
		return tools.MultipartResult{}, err
	}

	reader, writer := io.Pipe()
	go func() {
		defer func() { _ = req.Close() }()
		_, err := io.Copy(writer, container)
		_ = writer.CloseWithError(err)
	}()

	return tools.MultipartResult{
		ContentType: "application/octet-stream",
		Filename:    decryptedFilename(req.Filename),
		Body:        reader,
	}, nil
}

// decodeJSONByteField 解析形如 `{"data":"…","encoding":"hex"}` 的表单字段。
//
// 空串返回 nil（表示「未提供」）。
func decodeJSONByteField(raw string, field string) ([]byte, error) {
	if raw == "" {
		return nil, nil
	}
	var in ByteInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		return nil, tools.InvalidField(field, "JSON object { data, encoding }", "malformed JSON")
	}
	if in.Encoding == "" {
		in.Encoding = codecsvc.EncodingHex
	}
	value, err := codecsvc.Decode(in.Data, in.Encoding)
	if err != nil {
		return nil, tools.InvalidField(field+".encoding", codecsvc.Encodings(), in.Encoding)
	}
	return value, nil
}

// encryptedFilename 生成加密后的下载文件名。
func encryptedFilename(name string) string {
	safe := tools.SanitizeFilename(name)
	if safe == "" {
		return "encrypted.sm4"
	}
	return safe + ".sm4"
}

// decryptedFilename 生成解密后的下载文件名。
//
// 去掉 `.sm4` 后缀；没有后缀时原样返回。
func decryptedFilename(name string) string {
	safe := tools.SanitizeFilename(name)
	if safe == "" {
		return "decrypted"
	}
	if len(safe) > 4 && safe[len(safe)-4:] == ".sm4" {
		return safe[:len(safe)-4]
	}
	return safe
}
