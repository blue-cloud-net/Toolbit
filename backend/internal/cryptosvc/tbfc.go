package cryptosvc

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/blue-cloud-net/tongsuo-go/sym"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TBFC 是 ToolBit File Container 的魔数。
const TBFC = "TBFC"

// TBFC 容器的格式常量。
const (
	// TBFCVersion 是当前容器版本。
	TBFCVersion byte = 1
	// TBFCHeaderSize 是容器头长度：
	// magic(4) + version(1) + flags(1) + chunkSize(4) + baseNonce(12)。
	TBFCHeaderSize = 22
	// TBFCDefaultChunkSize 是默认分块大小（1 MiB）。
	TBFCDefaultChunkSize = 1 << 20
	// TBFCMinChunkSize 是允许的最小分块大小（4 KiB）。
	TBFCMinChunkSize = 4 << 10
	// TBFCMaxChunkSize 是允许的最大分块大小（16 MiB）。
	//
	// 上限保证内存占用可控：加密与解密各持有一份分块缓冲。
	TBFCMaxChunkSize = 16 << 20
	// TBFCBaseNonceSize 是容器头里基础 nonce 的长度。
	TBFCBaseNonceSize = 12
	// TBFCChunkNoncePrefixSize 是分块 nonce 中沿用基础 nonce 的前缀长度。
	//
	// 剩余 4 字节放分块索引，因此 nonce 在同一容器内**唯一**。
	TBFCChunkNoncePrefixSize = 8
)

// TBFC 容器的错误。
var (
	// ErrTBFCBadMagic 表示魔数不匹配：输入不是 TBFC 容器。
	ErrTBFCBadMagic = errors.New("tbfc: bad magic")
	// ErrTBFCBadVersion 表示版本不支持。
	ErrTBFCBadVersion = errors.New("tbfc: unsupported version")
	// ErrTBFCBadChunkSize 表示分块大小非法。
	ErrTBFCBadChunkSize = errors.New("tbfc: bad chunk size")
	// ErrTBFCTruncated 表示容器被截断。
	ErrTBFCTruncated = errors.New("tbfc: truncated container")
)

// TBFCHeader 是容器头。
type TBFCHeader struct {
	// ChunkSize 是分块大小（字节）。
	ChunkSize uint32
	// BaseNonce 是基础 nonce（12 字节）。
	//
	// 分块 nonce 由它派生：`baseNonce[0:8] || uint32be(chunkIndex)`。
	BaseNonce []byte
}

// Marshal 序列化容器头。
func (h *TBFCHeader) Marshal() []byte {
	buf := make([]byte, TBFCHeaderSize)
	copy(buf[0:4], TBFC)
	buf[4] = TBFCVersion
	buf[5] = 0 // flags：预留。
	binary.BigEndian.PutUint32(buf[6:10], h.ChunkSize)
	copy(buf[10:22], h.BaseNonce)
	return buf
}

// ParseTBFCHeader 解析容器头。
func ParseTBFCHeader(raw []byte) (*TBFCHeader, error) {
	if len(raw) < TBFCHeaderSize {
		return nil, toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCTruncated)
	}
	if string(raw[0:4]) != TBFC {
		return nil, toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCBadMagic)
	}
	if raw[4] != TBFCVersion {
		return nil, toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCBadVersion)
	}
	chunkSize := binary.BigEndian.Uint32(raw[6:10])
	if chunkSize < TBFCMinChunkSize || chunkSize > TBFCMaxChunkSize {
		return nil, toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCBadChunkSize)
	}
	baseNonce := make([]byte, TBFCBaseNonceSize)
	copy(baseNonce, raw[10:22])
	return &TBFCHeader{ChunkSize: chunkSize, BaseNonce: baseNonce}, nil
}

// ChunkNonce 派生第 index 块使用的 nonce。
//
// 结构：`baseNonce[0:8] || uint32be(index)`。
//
// 基础 nonce 随机生成，因此不同容器之间 nonce 不重复；
// 块索引让同一容器内的 nonce 也不重复。
func (h *TBFCHeader) ChunkNonce(index uint32) []byte {
	nonce := make([]byte, GCMNonceSize)
	copy(nonce[:TBFCChunkNoncePrefixSize], h.BaseNonce[:TBFCChunkNoncePrefixSize])
	binary.BigEndian.PutUint32(nonce[TBFCChunkNoncePrefixSize:], index)
	return nonce
}

// ChunkAAD 构造第 index 块的附加认证数据。
//
// AAD 绑定**容器头**与**块索引**：
//   - 绑定容器头 → 改分块大小或基础 nonce 会导致认证失败；
//   - 绑定块索引 → 块重排或删除会导致认证失败。
func (h *TBFCHeader) ChunkAAD(index uint32) []byte {
	aad := make([]byte, 0, TBFCHeaderSize+4)
	aad = append(aad, h.Marshal()...)
	var idx [4]byte
	binary.BigEndian.PutUint32(idx[:], index)
	return append(aad, idx[:]...)
}

// TBFCWriter 是 TBFC 容器的流式写入器。
//
// 它实现了 io.Writer：写入的数据按 chunkSize 分块，
// 每块独立用 GCM 加密并附上认证标签。
//
// 内存占用约 2×chunkSize（明文缓冲 + 密文缓冲），**不**随文件大小增长。
type TBFCWriter struct {
	// w 是底层写入器。
	w io.Writer
	// algorithm 是对称算法名。
	algorithm string
	// header 是容器头。
	header *TBFCHeader
	// key 是对称密钥。
	key []byte
	// buf 是明文分块缓冲。
	buf []byte
	// filled 是 buf 中已用的字节数。
	filled int
	// index 是当前块索引。
	index uint32
	// closed 标记是否已写结束块。
	closed bool
	// err 记录首个错误。
	err error
}

// NewTBFCWriter 构造容器写入器并写出容器头。
//
// algorithm 取 SymSM4 或 SymAES；baseNonce 为空时生成随机值。
func NewTBFCWriter(w io.Writer, algorithm string, key, baseNonce []byte, chunkSize uint32) (*TBFCWriter, error) {
	if chunkSize == 0 {
		chunkSize = TBFCDefaultChunkSize
	}
	if chunkSize < TBFCMinChunkSize || chunkSize > TBFCMaxChunkSize {
		return nil, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":    "chunkSize",
			"expected": []int{TBFCMinChunkSize, TBFCMaxChunkSize},
			"actual":   chunkSize,
		})
	}
	if err := checkSymKeyLength(algorithm, len(key)); err != nil {
		return nil, err
	}
	if len(baseNonce) == 0 {
		random, err := RandomBytes(TBFCBaseNonceSize)
		if err != nil {
			return nil, err
		}
		baseNonce = random
	}
	if len(baseNonce) != TBFCBaseNonceSize {
		return nil, toolerr.Newf(toolerr.CodeIVSizeMismatch, map[string]any{
			"field":    "iv",
			"expected": TBFCBaseNonceSize,
			"actual":   len(baseNonce),
		})
	}

	header := &TBFCHeader{ChunkSize: chunkSize, BaseNonce: baseNonce}
	if _, err := w.Write(header.Marshal()); err != nil {
		return nil, toolerr.Wrap(toolerr.CodeInternalError, err)
	}
	return &TBFCWriter{
		w:         w,
		algorithm: algorithm,
		header:    header,
		key:       key,
		buf:       make([]byte, chunkSize),
	}, nil
}

// Header 返回容器头。
func (w *TBFCWriter) Header() *TBFCHeader {
	return w.header
}

// Write 加密并写入数据。
//
// 数据按 chunkSize 分块；每个非空块独立加密并附认证标签。
func (w *TBFCWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	if w.closed {
		return 0, toolerr.New(toolerr.CodeInternalError)
	}
	written := 0
	for len(p) > 0 {
		n := copy(w.buf[w.filled:], p)
		w.filled += n
		p = p[n:]
		written += n
		if w.filled == len(w.buf) {
			if err := w.flushChunk(); err != nil {
				return written, err
			}
		}
	}
	return written, nil
}

// Close 写出剩余数据与结束块。
//
// 结束块带认证标签：因此**截断**容器会被检测出来，
// 而不会被误判成「文件到此为止」。
func (w *TBFCWriter) Close() error {
	if w.err != nil {
		return w.err
	}
	if w.closed {
		return nil
	}
	w.closed = true
	if w.filled > 0 {
		if err := w.flushChunk(); err != nil {
			return err
		}
	}
	return w.writeChunk(nil)
}

// flushChunk 把缓冲里的明文写成一个密文块。
func (w *TBFCWriter) flushChunk() error {
	plaintext := w.buf[:w.filled]
	w.filled = 0
	return w.writeChunk(plaintext)
}

// writeChunk 加密并写出一个块。
func (w *TBFCWriter) writeChunk(plaintext []byte) error {
	index := w.index
	w.index++

	nonce := w.header.ChunkNonce(index)
	aad := w.header.ChunkAAD(index)

	ciphertext, tag, err := SymEncrypt(w.algorithm, ModeGCM, PaddingPKCS7, w.key, nonce, plaintext, aad)
	if err != nil {
		w.err = err
		return err
	}

	frame := make([]byte, 4, 4+len(ciphertext)+len(tag))
	binary.BigEndian.PutUint32(frame[0:4], uint32(len(ciphertext)))
	frame = append(frame, ciphertext...)
	frame = append(frame, tag...)
	if _, err := w.w.Write(frame); err != nil {
		w.err = toolerr.Wrap(toolerr.CodeInternalError, err)
		return w.err
	}
	return nil
}

// TBFCReader 是 TBFC 容器的流式读取器。
//
// 它实现了 io.Reader：逐块解密并校验认证标签。
type TBFCReader struct {
	// r 是底层读取器。
	r io.Reader
	// algorithm 是对称算法名。
	algorithm string
	// header 是容器头。
	header *TBFCHeader
	// key 是对称密钥。
	key []byte
	// buf 是当前块的明文。
	buf []byte
	// pos 是 buf 中的读取位置。
	pos int
	// index 是当前块索引。
	index uint32
	// done 标记容器是否已读完。
	done bool
	// probed 标记是否已做过首块预校验。
	probed bool
	// err 记录首个错误。
	err error
}

// NewTBFCReader 构造容器读取器并解析容器头。
func NewTBFCReader(r io.Reader, algorithm string, key []byte) (*TBFCReader, error) {
	if err := checkSymKeyLength(algorithm, len(key)); err != nil {
		return nil, err
	}
	raw := make([]byte, TBFCHeaderSize)
	if _, err := io.ReadFull(r, raw); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCTruncated)
		}
		return nil, toolerr.Wrap(toolerr.CodeInvalidField, err)
	}
	header, err := ParseTBFCHeader(raw)
	if err != nil {
		return nil, err
	}
	return &TBFCReader{r: r, algorithm: algorithm, header: header, key: key}, nil
}

// Header 返回容器头。
func (r *TBFCReader) Header() *TBFCHeader {
	return r.header
}

// Probe 预读并校验第一块。
//
// 目的是让「密钥错误 / 容器损坏 / 魔数不对」在**响应头发出之前**
// 就被发现，从而返回干净的 JSON 错误信封，
// 而不是 `200` + 一个被截断的下载。
//
// 空容器（只有结束块）视为正常。
func (r *TBFCReader) Probe() error {
	if r.probed {
		return nil
	}
	r.probed = true

	err := r.nextChunk()
	if errors.Is(err, io.EOF) {
		// 只有结束块：空文件。
		r.done = true
		r.buf = nil
		r.pos = 0
		return nil
	}
	return err
}

// Read 读取解出的明文。
//
// 任一块的认证失败都返回 DECRYPT_FAILED（**不是** EOF）：
// 损坏的容器**必须**与正常结束区分开。
func (r *TBFCReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if r.done {
		return 0, io.EOF
	}
	for r.pos >= len(r.buf) {
		if err := r.nextChunk(); err != nil {
			if errors.Is(err, io.EOF) {
				r.done = true
				return 0, io.EOF
			}
			r.err = err
			return 0, err
		}
	}
	n := copy(p, r.buf[r.pos:])
	r.pos += n
	return n, nil
}

// nextChunk 读取并解密下一个块。
func (r *TBFCReader) nextChunk() error {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r.r, lenBuf[:]); err != nil {
		// 容器**必须**以结束块收尾：这里读不到块长度说明被截断。
		//
		// ⚠️ 不能把它当成正常结束——那会让被截断的容器被静默接受，
		// 而结束块的认证标签正是用来防这件事的。
		return toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCTruncated)
	}
	length := binary.BigEndian.Uint32(lenBuf[:])

	if length == 0 {
		// 结束块：校验它的认证标签。
		if err := r.readTerminator(); err != nil {
			return err
		}
		return io.EOF
	}
	if length > r.header.ChunkSize {
		return toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCBadChunkSize)
	}

	ciphertext := make([]byte, length)
	if _, err := io.ReadFull(r.r, ciphertext); err != nil {
		return toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCTruncated)
	}
	tag := make([]byte, sym.TagSize)
	if _, err := io.ReadFull(r.r, tag); err != nil {
		return toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCTruncated)
	}

	nonce := r.header.ChunkNonce(r.index)
	aad := r.header.ChunkAAD(r.index)
	r.index++

	plaintext, err := SymDecrypt(r.algorithm, ModeGCM, PaddingPKCS7, r.key, nonce, ciphertext, tag, aad)
	if err != nil {
		return err
	}
	r.buf = plaintext
	r.pos = 0
	return nil
}

// readTerminator 读取并校验结束块。
func (r *TBFCReader) readTerminator() error {
	tag := make([]byte, sym.TagSize)
	if _, err := io.ReadFull(r.r, tag); err != nil {
		return toolerr.Wrap(toolerr.CodeInvalidField, ErrTBFCTruncated)
	}
	nonce := r.header.ChunkNonce(r.index)
	aad := r.header.ChunkAAD(r.index)
	r.index++
	if _, err := SymDecrypt(r.algorithm, ModeGCM, PaddingPKCS7, r.key, nonce, nil, tag, aad); err != nil {
		return err
	}
	return nil
}

// checkSymKeyLength 校验对称密钥长度。
//
// SM4 固定 16 字节；AES 支持 16 / 24 / 32 字节。
func checkSymKeyLength(algorithm string, length int) error {
	switch algorithm {
	case SymSM4:
		if length != sym.SM4KeySize {
			return toolerr.Newf(toolerr.CodeKeySizeMismatch, map[string]any{
				"algorithm": algorithm,
				"expected":  sym.SM4KeySize,
				"actual":    length,
			})
		}
	case SymAES:
		switch length {
		case sym.AES128KeySize, sym.AES192KeySize, sym.AES256KeySize:
		default:
			return toolerr.Newf(toolerr.CodeKeySizeMismatch, map[string]any{
				"algorithm": algorithm,
				"expected":  []int{sym.AES128KeySize, sym.AES192KeySize, sym.AES256KeySize},
				"actual":    length,
			})
		}
	default:
		return unsupportedAlgorithm(algorithm, SymAlgorithms())
	}
	return nil
}
