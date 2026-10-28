package cryptosvc

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// mustRand 返回 n 字节随机数据。
func mustRand(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("随机数据生成失败：%v", err)
	}
	return b
}

// tbfcRoundTrip 加密后解密，校验内容一致。
func tbfcRoundTrip(t *testing.T, plain []byte, chunkSize uint32) {
	t.Helper()
	key := mustRand(t, 16)

	var enc bytes.Buffer
	w, err := NewTBFCWriter(&enc, SymSM4, key, nil, chunkSize)
	if err != nil {
		t.Fatalf("构造写入器失败：%v", err)
	}
	if _, err := w.Write(plain); err != nil {
		t.Fatalf("写入失败：%v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("关闭失败：%v", err)
	}

	r, err := NewTBFCReader(bytes.NewReader(enc.Bytes()), SymSM4, key)
	if err != nil {
		t.Fatalf("构造读取器失败：%v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("读取失败：%v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("往返不一致：期望 %d 字节，实际 %d 字节", len(plain), len(got))
	}
}

func TestTBFCRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		size  int
		chunk uint32
	}{
		{"空文件", 0, TBFCDefaultChunkSize},
		{"小于一块", 100, TBFCDefaultChunkSize},
		{"恰好一块", TBFCMinChunkSize, TBFCMinChunkSize},
		{"跨多块", TBFCMinChunkSize*3 + 17, TBFCMinChunkSize},
		{"恰好整块倍数", TBFCMinChunkSize * 2, TBFCMinChunkSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tbfcRoundTrip(t, mustRand(t, tc.size), tc.chunk)
		})
	}
}

func TestTBFCHeaderRoundTrip(t *testing.T) {
	h := &TBFCHeader{ChunkSize: 65536, BaseNonce: mustRand(t, TBFCBaseNonceSize)}
	parsed, err := ParseTBFCHeader(h.Marshal())
	if err != nil {
		t.Fatalf("解析容器头失败：%v", err)
	}
	if parsed.ChunkSize != h.ChunkSize {
		t.Errorf("分块大小 = %d，期望 %d", parsed.ChunkSize, h.ChunkSize)
	}
	if !bytes.Equal(parsed.BaseNonce, h.BaseNonce) {
		t.Error("基础 nonce 不一致")
	}
}

func TestTBFCParseRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		raw  []byte
	}{
		{"魔数错误", append([]byte("XXXX"), make([]byte, 18)...)},
		{"版本不支持", append([]byte(TBFC), append([]byte{99}, make([]byte, 17)...)...)},
		{"分块过小", append([]byte(TBFC), append([]byte{1, 0, 0, 0, 0}, make([]byte, 12)...)...)},
		{"长度不足", []byte(TBFC)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseTBFCHeader(tc.raw); err == nil {
				t.Fatal("期望报错，实际通过")
			}
		})
	}
}

// 分块 nonce 在同一容器内必须互不相同。
func TestTBFCChunkNonceUnique(t *testing.T) {
	h := &TBFCHeader{ChunkSize: 4096, BaseNonce: mustRand(t, TBFCBaseNonceSize)}
	seen := make(map[string]bool, 64)
	for i := uint32(0); i < 64; i++ {
		nonce := h.ChunkNonce(i)
		if len(nonce) != GCMNonceSize {
			t.Fatalf("nonce 长度 = %d，期望 %d", len(nonce), GCMNonceSize)
		}
		key := string(nonce)
		if seen[key] {
			t.Fatalf("第 %d 块的 nonce 与前面重复", i)
		}
		seen[key] = true
	}
}

// 换密钥解密必须失败，且错误码是 DECRYPT_FAILED / TAG_INVALID。
func TestTBFCDecryptWithWrongKeyFails(t *testing.T) {
	key := mustRand(t, 16)
	var enc bytes.Buffer
	w, err := NewTBFCWriter(&enc, SymSM4, key, nil, TBFCDefaultChunkSize)
	if err != nil {
		t.Fatalf("构造写入器失败：%v", err)
	}
	if _, err := w.Write([]byte("机密内容")); err != nil {
		t.Fatalf("写入失败：%v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("关闭失败：%v", err)
	}

	r, err := NewTBFCReader(bytes.NewReader(enc.Bytes()), SymSM4, mustRand(t, 16))
	if err != nil {
		t.Fatalf("构造读取器失败：%v", err)
	}
	if _, err := io.ReadAll(r); err == nil {
		t.Fatal("期望解密失败，实际通过")
	}
}

// 截断容器必须被检测出来，不能被当成「文件到此为止」。
func TestTBFCTruncationDetected(t *testing.T) {
	key := mustRand(t, 16)
	var enc bytes.Buffer
	w, err := NewTBFCWriter(&enc, SymSM4, key, nil, TBFCDefaultChunkSize)
	if err != nil {
		t.Fatalf("构造写入器失败：%v", err)
	}
	if _, err := w.Write([]byte("机密内容")); err != nil {
		t.Fatalf("写入失败：%v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("关闭失败：%v", err)
	}

	// 去掉结束块：只剩容器头 + 一个数据块。
	truncated := enc.Bytes()[:enc.Len()-20]
	r, err := NewTBFCReader(bytes.NewReader(truncated), SymSM4, key)
	if err != nil {
		t.Fatalf("构造读取器失败：%v", err)
	}
	if _, err := io.ReadAll(r); err == nil {
		t.Fatal("期望检测到截断，实际通过")
	} else if errors.Is(err, io.EOF) {
		t.Fatal("截断被误判为正常结束")
	}
}

// 篡改密文必须导致认证失败。
func TestTBFCTamperDetected(t *testing.T) {
	key := mustRand(t, 16)
	var enc bytes.Buffer
	w, err := NewTBFCWriter(&enc, SymSM4, key, nil, TBFCDefaultChunkSize)
	if err != nil {
		t.Fatalf("构造写入器失败：%v", err)
	}
	if _, err := w.Write([]byte("机密内容机密内容机密内容")); err != nil {
		t.Fatalf("写入失败：%v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("关闭失败：%v", err)
	}

	raw := enc.Bytes()
	raw[TBFCHeaderSize+5] ^= 0xff // 翻转密文中的一个字节。

	r, err := NewTBFCReader(bytes.NewReader(raw), SymSM4, key)
	if err != nil {
		t.Fatalf("构造读取器失败：%v", err)
	}
	if _, err := io.ReadAll(r); err == nil {
		t.Fatal("期望检测到篡改，实际通过")
	}
}

// 分块大小非法必须被拒绝。
func TestTBFCRejectsBadChunkSize(t *testing.T) {
	key := mustRand(t, 16)
	for _, size := range []uint32{1, TBFCMinChunkSize - 1, TBFCMaxChunkSize + 1} {
		var buf bytes.Buffer
		if _, err := NewTBFCWriter(&buf, SymSM4, key, nil, size); err == nil {
			t.Errorf("分块大小 %d 期望被拒绝", size)
		} else if codeOf(t, err) != toolerr.CodeInvalidField {
			t.Errorf("错误码 = %v，期望 %v", codeOf(t, err), toolerr.CodeInvalidField)
		}
	}
}

// 密钥长度非法必须返回 KEY_SIZE_MISMATCH。
func TestTBFCRejectsBadKeyLength(t *testing.T) {
	var buf bytes.Buffer
	_, err := NewTBFCWriter(&buf, SymSM4, make([]byte, 8), nil, TBFCDefaultChunkSize)
	if err == nil {
		t.Fatal("期望报错，实际通过")
	}
	if codeOf(t, err) != toolerr.CodeKeySizeMismatch {
		t.Errorf("错误码 = %v，期望 %v", codeOf(t, err), toolerr.CodeKeySizeMismatch)
	}
}

// 显式给定基础 nonce 时，容器头里的 nonce 必须一致。
func TestTBFCExplicitBaseNonce(t *testing.T) {
	key := mustRand(t, 16)
	nonce := mustRand(t, TBFCBaseNonceSize)
	var enc bytes.Buffer
	w, err := NewTBFCWriter(&enc, SymSM4, key, nonce, TBFCDefaultChunkSize)
	if err != nil {
		t.Fatalf("构造写入器失败：%v", err)
	}
	if !bytes.Equal(w.Header().BaseNonce, nonce) {
		t.Error("容器头里的基础 nonce 与传入的不一致")
	}
}
