package cryptosvc

import (
	"context"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// testKey 读取 testdata 下的测试密钥。
func testKey(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("读取测试密钥失败：%v", err)
	}
	return b
}

// codeOf 返回错误的错误码。
func codeOf(t *testing.T, err error) toolerr.Code {
	t.Helper()
	if err == nil {
		t.Fatal("期望有错误，实际为 nil")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("期望 *toolerr.Error，实际为 %T", err)
	}
	return toolErr.Code
}

// mustHex 把十六进制串解码为字节。
func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("十六进制解码失败：%v", err)
	}
	return b
}

// ── Session ────────────────────────────────────────────────────────────

func TestSessionCloseIsIdempotent(t *testing.T) {
	s := NewSession(context.Background())
	calls := 0
	s.Defer(func() error {
		calls++
		return nil
	})
	if err := s.Close(); err != nil {
		t.Fatalf("首次 Close 返回错误：%v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("二次 Close 返回错误：%v", err)
	}
	if calls != 1 {
		t.Errorf("释放器执行次数 = %d，期望 1", calls)
	}
}

func TestSessionCloseIsLIFO(t *testing.T) {
	s := NewSession(context.Background())
	var order []string
	s.Defer(func() error { order = append(order, "first"); return nil })
	s.Defer(func() error { order = append(order, "second"); return nil })
	s.Defer(func() error { order = append(order, "third"); return nil })
	if err := s.Close(); err != nil {
		t.Fatalf("Close 返回错误：%v", err)
	}
	want := []string{"third", "second", "first"}
	if len(order) != len(want) {
		t.Fatalf("释放顺序 = %v，期望 %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("释放顺序 = %v，期望 %v", order, want)
		}
	}
}

func TestSessionWarnings(t *testing.T) {
	s := NewSession(context.Background())
	s.AddWarning(toolerr.WarnInsecureECB, map[string]any{"mode": "ECB"})
	s.AddWarning(toolerr.WarnWeakHash, nil)
	got := s.Warnings()
	if len(got) != 2 {
		t.Fatalf("警告数 = %d，期望 2", len(got))
	}
	if got[0].Code != toolerr.WarnInsecureECB {
		t.Errorf("警告[0] = %v，期望 %v", got[0].Code, toolerr.WarnInsecureECB)
	}
}

func TestSessionContextRoundTrip(t *testing.T) {
	s := NewSession(context.Background())
	got := FromContext(s.Context())
	if got != s {
		t.Errorf("FromContext 返回的不是同一个 Session")
	}
	if FromContext(context.Background()) != nil {
		t.Error("无 Session 的上下文应返回 nil")
	}
}

func TestSessionRegisterKeyReleases(t *testing.T) {
	key, err := LoadPrivateKey(testKey(t, "sm2.key"), "")
	if err != nil {
		t.Fatalf("加载私钥失败：%v", err)
	}
	s := NewSession(context.Background())
	s.RegisterKey(key)
	if err := s.Close(); err != nil {
		t.Fatalf("Close 返回错误：%v", err)
	}
	// 释放后再次使用应报错而非崩溃。
	if _, err := key.PublicKeyPEM(); err == nil {
		t.Error("已释放的私钥仍可导出公钥")
	}
}

// ── 信号量 ─────────────────────────────────────────────────────────────

func TestSessionAcquireRespectsContext(t *testing.T) {
	s := NewSession(context.Background())
	// 占满许可。
	for i := 0; i < heavyConcurrency; i++ {
		if err := s.Acquire(context.Background()); err != nil {
			t.Fatalf("第 %d 次 Acquire 失败：%v", i, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Acquire(ctx); err == nil {
		t.Error("许可耗尽且 ctx 取消时 Acquire 应失败")
	} else if code := codeOf(t, err); code != toolerr.CodeTimeout {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeTimeout)
	}
	for i := 0; i < heavyConcurrency; i++ {
		s.Release("test")
	}
}

// ── 摘要 ───────────────────────────────────────────────────────────────

func TestDigestSM3KnownVector(t *testing.T) {
	sum, err := Digest("SM3", []byte("abc"))
	if err != nil {
		t.Fatalf("SM3 计算失败：%v", err)
	}
	want := "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0"
	if got := hexEncode(sum); got != want {
		t.Errorf("SM3(abc) = %s，期望 %s", got, want)
	}
}

func TestDigestUnknownAlgorithm(t *testing.T) {
	_, err := Digest("SHA999", []byte("x"))
	if err == nil {
		t.Fatal("未知摘要算法应报错")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedAlgorithm)
	}
}

// ── 对称加密 ───────────────────────────────────────────────────────────

func TestSymEncryptSM4ECBKnownVector(t *testing.T) {
	// 国标向量对应**不填充**的 ECB：16 字节明文 → 16 字节密文。
	key := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")
	plain := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")
	ct, _, err := SymEncrypt(SymSM4, ModeECB, PaddingNone, key, nil, plain, nil)
	if err != nil {
		t.Fatalf("SM4-ECB 加密失败：%v", err)
	}
	want := "681edf34d206965e86b3e94f536e4246"
	if got := hexEncode(ct); got != want {
		t.Errorf("SM4-ECB 密文 = %s，期望 %s", got, want)
	}
}

func TestSymEncryptSM4ECBPkcs7AddsPadding(t *testing.T) {
	key := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")
	plain := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")
	ct, _, err := SymEncrypt(SymSM4, ModeECB, PaddingPKCS7, key, nil, plain, nil)
	if err != nil {
		t.Fatalf("SM4-ECB 加密失败：%v", err)
	}
	// PKCS#7 追加一整块 0x10。
	if len(ct) != 32 {
		t.Errorf("密文长度 = %d，期望 32", len(ct))
	}
}

func TestSymRoundTripAllModes(t *testing.T) {
	key := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")
	iv := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	nonce := mustHex(t, "000102030405060708090a0b")
	plain := []byte("蓝云工具箱 Toolbit")

	for _, mode := range SymModes(SymSM4) {
		useIV := iv
		if mode == ModeGCM {
			useIV = nonce
		}
		ct, tag, err := SymEncrypt(SymSM4, mode, PaddingPKCS7, key, useIV, plain, nil)
		if err != nil {
			t.Fatalf("SM4-%s 加密失败：%v", mode, err)
		}
		got, err := SymDecrypt(SymSM4, mode, PaddingPKCS7, key, useIV, ct, tag, nil)
		if err != nil {
			t.Fatalf("SM4-%s 解密失败：%v", mode, err)
		}
		if string(got) != string(plain) {
			t.Errorf("SM4-%s 往返不一致：%q", mode, string(got))
		}
	}
}

func TestSymAESGCMRoundTrip(t *testing.T) {
	key := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	nonce := mustHex(t, "000102030405060708090a0b")
	plain := []byte("hello")

	ct, tag, err := SymEncrypt(SymAES, ModeGCM, PaddingPKCS7, key, nonce, plain, []byte("aad"))
	if err != nil {
		t.Fatalf("AES-GCM 加密失败：%v", err)
	}
	if len(tag) != 16 {
		t.Errorf("tag 长度 = %d，期望 16", len(tag))
	}
	got, err := SymDecrypt(SymAES, ModeGCM, PaddingPKCS7, key, nonce, ct, tag, []byte("aad"))
	if err != nil {
		t.Fatalf("AES-GCM 解密失败：%v", err)
	}
	if string(got) != string(plain) {
		t.Errorf("AES-GCM 往返不一致：%q", string(got))
	}
}

func TestSymAESRejectsOFBCFB(t *testing.T) {
	key := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	for _, mode := range []string{ModeOFB, ModeCFB} {
		_, _, err := SymEncrypt(SymAES, mode, PaddingPKCS7, key, nil, []byte("x"), nil)
		if err == nil {
			t.Fatalf("AES-%s 应被拒绝", mode)
		}
		if code := codeOf(t, err); code != toolerr.CodeUnsupportedMode {
			t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedMode)
		}
	}
}

func TestSymPaddingMatrix(t *testing.T) {
	key := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")

	// AES + zero 必须被拒绝（否则库会静默忽略 padding）。
	_, _, err := SymEncrypt(SymAES, ModeCBC, PaddingZero, key, make([]byte, 16), []byte("x"), nil)
	if err == nil {
		t.Fatal("AES + zero 填充应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedPadding {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedPadding)
	}

	// SM4 + CBC + none 允许，但要求块对齐。
	_, _, err = SymEncrypt(SymSM4, ModeCBC, PaddingNone, key, make([]byte, 16), []byte("short"), nil)
	if err == nil {
		t.Fatal("SM4 + none 填充的非对齐输入应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeInvalidField)
	}

	// 块对齐时通过。
	if _, _, err := SymEncrypt(SymSM4, ModeCBC, PaddingNone, key, make([]byte, 16), make([]byte, 32), nil); err != nil {
		t.Errorf("SM4 + none 的对齐输入应通过：%v", err)
	}
}

func TestSymIVSizeMismatch(t *testing.T) {
	key := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")
	_, _, err := SymEncrypt(SymSM4, ModeCBC, PaddingPKCS7, key, []byte("short"), []byte("x"), nil)
	if err == nil {
		t.Fatal("IV 长度错误应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodeIVSizeMismatch {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeIVSizeMismatch)
	}
}

func TestSymGCMRequiresTag(t *testing.T) {
	key := mustHex(t, "0123456789ABCDEFFEDCBA9876543210")
	nonce := mustHex(t, "000102030405060708090a0b")
	_, err := SymDecrypt(SymSM4, ModeGCM, PaddingPKCS7, key, nonce, []byte("ct"), nil, nil)
	if err == nil {
		t.Fatal("GCM 解密缺少 tag 应报错")
	}
	if code := codeOf(t, err); code != toolerr.CodeTagRequired {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeTagRequired)
	}
}

func TestSymKeySizeMismatch(t *testing.T) {
	_, _, err := SymEncrypt(SymSM4, ModeECB, PaddingPKCS7, []byte("short"), nil, []byte("x"), nil)
	if err == nil {
		t.Fatal("密钥长度错误应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodeKeySizeMismatch {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeKeySizeMismatch)
	}
}

// ── 签名 ───────────────────────────────────────────────────────────────

func TestSignSM2Verify(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "sm2.key"), "")
	if err != nil {
		t.Fatalf("加载 SM2 私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()

	msg := []byte("待签名内容")
	sig, err := Sign(SignSM2, "", "", priv, msg, "")
	if err != nil {
		t.Fatalf("SM2 签名失败：%v", err)
	}
	pub, err := LoadPublicKey(testKey(t, "sm2.pub"))
	if err != nil {
		t.Fatalf("加载 SM2 公钥失败：%v", err)
	}
	ok, err := Verify(SignSM2, "", "", pub, msg, sig, "")
	if err != nil {
		t.Fatalf("SM2 验签失败：%v", err)
	}
	if !ok {
		t.Error("SM2 验签应通过")
	}
}

func TestSignSM2WrongUserIDIsBusinessFailure(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "sm2.key"), "")
	if err != nil {
		t.Fatalf("加载 SM2 私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()

	msg := []byte("待签名内容")
	sig, err := Sign(SignSM2, "", "", priv, msg, "custom-id")
	if err != nil {
		t.Fatalf("SM2 签名失败：%v", err)
	}
	pub, err := LoadPublicKey(testKey(t, "sm2.pub"))
	if err != nil {
		t.Fatalf("加载 SM2 公钥失败：%v", err)
	}
	ok, err := Verify(SignSM2, "", "", pub, msg, sig, "")
	if err != nil {
		t.Fatalf("userId 不一致不应返回错误：%v", err)
	}
	if ok {
		t.Error("userId 不一致时验签应不通过")
	}
}

func TestSignRSARoundTrip(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "rsa.key"), "")
	if err != nil {
		t.Fatalf("加载 RSA 私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()
	pub, err := LoadPublicKey(testKey(t, "rsa.pub"))
	if err != nil {
		t.Fatalf("加载 RSA 公钥失败：%v", err)
	}

	msg := []byte("待签名内容")
	for _, hash := range SignHashes(SignRSA) {
		sig, err := Sign(SignRSA, hash, "", priv, msg, "")
		if err != nil {
			t.Fatalf("RSA-%s 签名失败：%v", hash, err)
		}
		ok, err := Verify(SignRSA, hash, "", pub, msg, sig, "")
		if err != nil {
			t.Fatalf("RSA-%s 验签失败：%v", hash, err)
		}
		if !ok {
			t.Errorf("RSA-%s 验签应通过", hash)
		}
	}
}

func TestSignEd25519RoundTrip(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "ed25519.key"), "")
	if err != nil {
		t.Fatalf("加载 Ed25519 私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()
	pub, err := LoadPublicKey(testKey(t, "ed25519.pub"))
	if err != nil {
		t.Fatalf("加载 Ed25519 公钥失败：%v", err)
	}

	msg := []byte("待签名内容")
	sig, err := Sign(SignED25519, "", "", priv, msg, "")
	if err != nil {
		t.Fatalf("Ed25519 签名失败：%v", err)
	}
	if len(sig) != 64 {
		t.Errorf("Ed25519 签名长度 = %d，期望 64", len(sig))
	}
	ok, err := Verify(SignED25519, "", "", pub, msg, sig, "")
	if err != nil {
		t.Fatalf("Ed25519 验签失败：%v", err)
	}
	if !ok {
		t.Error("Ed25519 验签应通过")
	}
}

func TestSignRejectsHashForFixedDigestAlgorithms(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "sm2.key"), "")
	if err != nil {
		t.Fatalf("加载 SM2 私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()

	_, err = Sign(SignSM2, HashSHA256, "", priv, []byte("x"), "")
	if err == nil {
		t.Fatal("SM2 不接受 hash 参数")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedAlgorithm)
	}
}

// ── 非对称加密 ─────────────────────────────────────────────────────────

func TestAsymEncryptDecryptSM2(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "sm2.key"), "")
	if err != nil {
		t.Fatalf("加载 SM2 私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()
	pubPEM, err := priv.PublicKeyPEM()
	if err != nil {
		t.Fatalf("导出公钥失败：%v", err)
	}
	pub, err := LoadPublicKey(pubPEM)
	if err != nil {
		t.Fatalf("加载公钥失败：%v", err)
	}

	plain := []byte("待加密内容")
	ct, err := AsymEncrypt(AsymSM2, "", "", pub, plain)
	if err != nil {
		t.Fatalf("SM2 加密失败：%v", err)
	}
	// 裸格式 + 未压缩点前缀。
	if len(ct) == 0 || ct[0] != 0x04 {
		t.Fatalf("SM2 密文应以 04 开头，实际首字节 = %#x", ct[0])
	}
	got, err := AsymDecrypt(AsymSM2, "", "", priv, ct)
	if err != nil {
		t.Fatalf("SM2 解密失败：%v", err)
	}
	if string(got) != string(plain) {
		t.Errorf("SM2 往返不一致：%q", string(got))
	}
}

func TestAsymDecryptRejectsCompressedPoint(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "sm2.key"), "")
	if err != nil {
		t.Fatalf("加载 SM2 私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()

	_, err = AsymDecrypt(AsymSM2, "", "", priv, append([]byte{0x02}, make([]byte, 64)...))
	if err == nil {
		t.Fatal("压缩点密文应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedFormat)
	}
}

func TestAsymRSAPlaintextTooLong(t *testing.T) {
	pub, err := LoadPublicKey(testKey(t, "rsa.pub"))
	if err != nil {
		t.Fatalf("加载 RSA 公钥失败：%v", err)
	}
	_, err = AsymEncrypt(AsymRSA, "oaep", "", pub, make([]byte, 1024))
	if err == nil {
		t.Fatal("过长的明文应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodePlaintextTooLong {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodePlaintextTooLong)
	}
}

// ── ECDH ───────────────────────────────────────────────────────────────

func TestECDHSharedSecretMatches(t *testing.T) {
	priv, err := LoadPrivateKey(testKey(t, "ec.key"), "")
	if err != nil {
		t.Fatalf("加载私钥失败：%v", err)
	}
	defer func() { _ = priv.Close() }()
	pubA, err := priv.PublicKeyPEM()
	if err != nil {
		t.Fatalf("导出公钥失败：%v", err)
	}

	// 双方用同一对密钥互相协商，结果应一致。
	s1, err := ECDHSharedSecret(CurveP256, testKey(t, "ec.key"), pubA)
	if err != nil {
		t.Fatalf("ECDH 协商失败：%v", err)
	}
	s2, err := ECDHSharedSecret(CurveP256, testKey(t, "ec.key"), pubA)
	if err != nil {
		t.Fatalf("ECDH 协商失败：%v", err)
	}
	if hexEncode(s1) != hexEncode(s2) {
		t.Error("同一对密钥的协商结果应一致")
	}
	Zero(s1)
	Zero(s2)
}

func TestECDHDerive(t *testing.T) {
	out, err := ECDHDerive(HashSHA256, []byte("secret"), []byte("salt"), []byte("info"), 32)
	if err != nil {
		t.Fatalf("HKDF 派生失败：%v", err)
	}
	if len(out) != 32 {
		t.Errorf("派生长度 = %d，期望 32", len(out))
	}
	if _, err := ECDHDerive(HashSHA256, []byte("secret"), nil, nil, MaxDerivedKeyBytes+1); err == nil {
		t.Error("超长派生应被拒绝")
	}
}

// ── 随机数 ─────────────────────────────────────────────────────────────

func TestRandomBytes(t *testing.T) {
	b, err := RandomBytes(32)
	if err != nil {
		t.Fatalf("生成随机字节失败：%v", err)
	}
	if len(b) != 32 {
		t.Errorf("随机字节长度 = %d，期望 32", len(b))
	}
	if _, err := RandomBytes(-1); err == nil {
		t.Error("负长度应被拒绝")
	}
}

// ── 密钥加载 ───────────────────────────────────────────────────────────

func TestLoadPrivateKeyInvalid(t *testing.T) {
	_, err := LoadPrivateKey([]byte("not a pem"), "")
	if err == nil {
		t.Fatal("非法私钥应报错")
	}
	if code := codeOf(t, err); code != toolerr.CodeInvalidKey {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeInvalidKey)
	}
}

func TestPrivateKeyCloseIsIdempotent(t *testing.T) {
	key, err := LoadPrivateKey(testKey(t, "sm2.key"), "")
	if err != nil {
		t.Fatalf("加载私钥失败：%v", err)
	}
	if err := key.Close(); err != nil {
		t.Fatalf("首次 Close 失败：%v", err)
	}
	if err := key.Close(); err != nil {
		t.Fatalf("二次 Close 失败：%v", err)
	}
}
