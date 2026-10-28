package crypto

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// testKey 读取 testdata 下的测试密钥。
func testKey(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("读取测试密钥失败：%v", err)
	}
	return string(b)
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

// hasWarning 报告警告列表中是否含有该警告码。
func hasWarning(warnings []toolerr.Warning, code toolerr.WarningCode) bool {
	for _, w := range warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

// ── hash-calc ──────────────────────────────────────────────────────────

func TestHashCalcSM3KnownVector(t *testing.T) {
	req := &HashCalcRequest{
		Algorithms:     []string{"SM3"},
		Input:          ByteInput{Data: "616263", Encoding: "hex"},
		OutputEncoding: "hex",
	}
	result, err := HashCalc(context.Background(), req)
	if err != nil {
		t.Fatalf("摘要计算失败：%v", err)
	}
	data, ok := result.Data.(HashCalcResult)
	if !ok {
		t.Fatalf("结果类型 = %T", result.Data)
	}
	want := "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0"
	if data.Digests[0].Value != want {
		t.Errorf("SM3(abc) = %s，期望 %s", data.Digests[0].Value, want)
	}
	if data.InputBytes != 3 {
		t.Errorf("输入字节数 = %d，期望 3", data.InputBytes)
	}
}

func TestHashCalcWeakHashWarning(t *testing.T) {
	result, err := HashCalc(context.Background(), &HashCalcRequest{
		Algorithms: []string{"MD5", "SHA1"},
		Input:      ByteInput{Data: "abc", Encoding: "utf8"},
	})
	if err != nil {
		t.Fatalf("摘要计算失败：%v", err)
	}
	if !hasWarning(result.Warnings, toolerr.WarnWeakHash) {
		t.Error("MD5 / SHA1 应产生 WEAK_HASH 警告")
	}
}

func TestHashCalcUnknownAlgorithm(t *testing.T) {
	_, err := HashCalc(context.Background(), &HashCalcRequest{
		Algorithms: []string{"SHA999"},
		Input:      ByteInput{Data: "abc", Encoding: "utf8"},
	})
	if err == nil {
		t.Fatal("未知算法应报错")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedAlgorithm)
	}
}

// ── hmac-gen ───────────────────────────────────────────────────────────

func TestHmacGen(t *testing.T) {
	result, err := HmacGen(context.Background(), &HmacGenRequest{
		Algorithm: "HMAC-SM3",
		Key:       ByteInput{Data: "6b6579", Encoding: "hex"},
		Message:   ByteInput{Data: "5468652071756963", Encoding: "hex"},
	})
	if err != nil {
		t.Fatalf("HMAC 计算失败：%v", err)
	}
	data := result.Data.(HmacGenResult)
	if data.MAC == "" {
		t.Error("MAC 不应为空")
	}
	if data.Algorithm != "HMAC-SM3" {
		t.Errorf("算法 = %q", data.Algorithm)
	}
}

func TestHmacGenWeakWarning(t *testing.T) {
	result, err := HmacGen(context.Background(), &HmacGenRequest{
		Algorithm: "HMAC-MD5",
		Key:       ByteInput{Data: "6b6579", Encoding: "hex"},
		Message:   ByteInput{Data: "abc", Encoding: "utf8"},
	})
	if err != nil {
		t.Fatalf("HMAC 计算失败：%v", err)
	}
	if !hasWarning(result.Warnings, toolerr.WarnWeakHash) {
		t.Error("HMAC-MD5 应产生 WEAK_HASH 警告")
	}
}

// ── password-gen ───────────────────────────────────────────────────────

func TestPasswordGenDefaults(t *testing.T) {
	result, err := PasswordGen(context.Background(), &PasswordGenRequest{})
	if err != nil {
		t.Fatalf("密码生成失败：%v", err)
	}
	data := result.Data.(PasswordGenResult)
	if len(data.Passwords) != 1 {
		t.Fatalf("密码数 = %d，期望 1", len(data.Passwords))
	}
	if len(data.Passwords[0]) != 16 {
		t.Errorf("密码长度 = %d，期望 16", len(data.Passwords[0]))
	}
	if data.EntropyBits <= 0 {
		t.Errorf("熵值 = %v，应大于 0", data.EntropyBits)
	}
}

func TestPasswordGenLengthAndCount(t *testing.T) {
	result, err := PasswordGen(context.Background(), &PasswordGenRequest{
		Length: 24,
		Count:  5,
	})
	if err != nil {
		t.Fatalf("密码生成失败：%v", err)
	}
	data := result.Data.(PasswordGenResult)
	if len(data.Passwords) != 5 {
		t.Fatalf("密码数 = %d，期望 5", len(data.Passwords))
	}
	seen := map[string]bool{}
	for _, p := range data.Passwords {
		if len(p) != 24 {
			t.Errorf("密码长度 = %d，期望 24", len(p))
		}
		if seen[p] {
			t.Error("生成了重复密码")
		}
		seen[p] = true
	}
}

func TestPasswordGenRequireEachClass(t *testing.T) {
	result, err := PasswordGen(context.Background(), &PasswordGenRequest{
		Length:           8,
		Charsets:         []string{"lower", "upper", "digit", "symbol"},
		RequireEachClass: true,
	})
	if err != nil {
		t.Fatalf("密码生成失败：%v", err)
	}
	pwd := result.Data.(PasswordGenResult).Passwords[0]
	classes := []string{
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"0123456789",
		"!@#$%^&*()-_=+[]{}|;:,.<>?/~",
	}
	for _, class := range classes {
		if !containsAny(pwd, class) {
			t.Errorf("密码 %q 缺少字符类 %q", pwd, class)
		}
	}
}

func TestPasswordGenExcludeAmbiguous(t *testing.T) {
	result, err := PasswordGen(context.Background(), &PasswordGenRequest{
		Length:           64,
		ExcludeAmbiguous: true,
	})
	if err != nil {
		t.Fatalf("密码生成失败：%v", err)
	}
	pwd := result.Data.(PasswordGenResult).Passwords[0]
	if containsAny(pwd, "0Oo1lI") {
		t.Errorf("密码 %q 含易混字符", pwd)
	}
}

func TestPasswordGenOutOfRange(t *testing.T) {
	if _, err := PasswordGen(context.Background(), &PasswordGenRequest{Length: 2}); err == nil {
		t.Error("过短的长度应报错")
	}
	if _, err := PasswordGen(context.Background(), &PasswordGenRequest{Length: 200}); err == nil {
		t.Error("过长的长度应报错")
	}
	if _, err := PasswordGen(context.Background(), &PasswordGenRequest{Count: 1000}); err == nil {
		t.Error("过大的数量应报错")
	}
}

// ── symmetric-crypt ────────────────────────────────────────────────────

func TestSymmetricCryptSM4CBCRoundTrip(t *testing.T) {
	key := ByteInput{Data: "0123456789ABCDEFFEDCBA9876543210", Encoding: "hex"}
	iv := ByteInput{Data: "000102030405060708090a0b0c0d0e0f", Encoding: "hex"}
	input := ByteInput{Data: "48656c6c6f", Encoding: "hex"}

	enc, err := SymmetricCrypt(context.Background(), &SymmetricCryptRequest{
		Algorithm: "SM4", Operation: "encrypt", Mode: "CBC",
		Key: key, IV: iv, Input: input,
	})
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	encData := enc.Data.(SymmetricCryptResult)

	dec, err := SymmetricCrypt(context.Background(), &SymmetricCryptRequest{
		Algorithm: "SM4", Operation: "decrypt", Mode: "CBC",
		Key: key, IV: iv,
		Input: ByteInput{Data: encData.Output, Encoding: "base64"},
	})
	if err != nil {
		t.Fatalf("解密失败：%v", err)
	}
	decData := dec.Data.(SymmetricCryptResult)
	// 输出默认是 base64；原文 "Hello" → "SGVsbG8="。
	if decData.Output != "SGVsbG8=" {
		t.Errorf("解密结果 = %q，期望 SGVsbG8=", decData.Output)
	}
}

func TestSymmetricCryptGCMGeneratesNonce(t *testing.T) {
	result, err := SymmetricCrypt(context.Background(), &SymmetricCryptRequest{
		Algorithm: "SM4", Operation: "encrypt", Mode: "GCM",
		Key:   ByteInput{Data: "0123456789ABCDEFFEDCBA9876543210", Encoding: "hex"},
		Input: ByteInput{Data: "48656c6c6f", Encoding: "hex"},
	})
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	data := result.Data.(SymmetricCryptResult)
	if data.IV == "" {
		t.Error("GCM 加密缺省 nonce 时应回传 iv")
	}
	if data.IVBytes != 12 {
		t.Errorf("IV 长度 = %d，期望 12", data.IVBytes)
	}
	if data.Tag == "" {
		t.Error("GCM 应返回认证标签")
	}
}

func TestSymmetricCryptECBWarns(t *testing.T) {
	result, err := SymmetricCrypt(context.Background(), &SymmetricCryptRequest{
		Algorithm: "SM4", Operation: "encrypt", Mode: "ECB",
		Key:   ByteInput{Data: "0123456789ABCDEFFEDCBA9876543210", Encoding: "hex"},
		Input: ByteInput{Data: "48656c6c6f", Encoding: "hex"},
	})
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	if !hasWarning(result.Warnings, toolerr.WarnInsecureECB) {
		t.Error("ECB 应产生 INSECURE_ECB 警告")
	}
}

func TestSymmetricCryptRejectsAESUnsupported(t *testing.T) {
	_, err := SymmetricCrypt(context.Background(), &SymmetricCryptRequest{
		Algorithm: "AES", Operation: "encrypt", Mode: "OFB",
		Key:   ByteInput{Data: "000102030405060708090a0b0c0d0e0f", Encoding: "hex"},
		Input: ByteInput{Data: "48656c6c6f", Encoding: "hex"},
	})
	if err == nil {
		t.Fatal("AES + OFB 应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedMode {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedMode)
	}
}

// ── signature ──────────────────────────────────────────────────────────

func TestSignatureSM2RoundTrip(t *testing.T) {
	privPEM := testKey(t, "sm2.key")
	pubPEM := testKey(t, "sm2.pub")
	msg := ByteInput{Data: "待签名内容", Encoding: "utf8"}

	signResult, err := Signature(context.Background(), &SignatureRequest{
		Algorithm: "SM2", Operation: "sign",
		Message: msg, Key: KeyInput{Data: privPEM, Format: "pem"},
	})
	if err != nil {
		t.Fatalf("签名失败：%v", err)
	}
	sig := signResult.Data.(SignatureResult).Signature

	verifyResult, err := Signature(context.Background(), &SignatureRequest{
		Algorithm: "SM2", Operation: "verify",
		Message: msg, PublicKey: KeyInput{Data: pubPEM, Format: "pem"},
		Signature: ByteInput{Data: sig, Encoding: "base64"},
	})
	if err != nil {
		t.Fatalf("验签失败：%v", err)
	}
	verifyData := verifyResult.Data.(SignatureResult)
	if verifyData.Valid == nil || !*verifyData.Valid {
		t.Error("验签应通过")
	}
}

func TestSignatureVerifyFailureIsBusinessResult(t *testing.T) {
	privPEM := testKey(t, "sm2.key")
	pubPEM := testKey(t, "sm2.pub")

	signResult, err := Signature(context.Background(), &SignatureRequest{
		Algorithm: "SM2", Operation: "sign",
		Message: ByteInput{Data: "待签名内容", Encoding: "utf8"},
		Key:     KeyInput{Data: privPEM, Format: "pem"},
	})
	if err != nil {
		t.Fatalf("签名失败：%v", err)
	}
	sig := signResult.Data.(SignatureResult).Signature

	// 换一份消息：验签不通过，但**不是**错误。
	verifyResult, err := Signature(context.Background(), &SignatureRequest{
		Algorithm: "SM2", Operation: "verify",
		Message:   ByteInput{Data: "另一段内容", Encoding: "utf8"},
		PublicKey: KeyInput{Data: pubPEM, Format: "pem"},
		Signature: ByteInput{Data: sig, Encoding: "base64"},
	})
	if err != nil {
		t.Fatalf("验签不通过不应返回错误：%v", err)
	}
	verifyData := verifyResult.Data.(SignatureResult)
	if verifyData.Valid == nil || *verifyData.Valid {
		t.Error("验签应不通过")
	}
}

func TestSignatureRejectsHashForSM2(t *testing.T) {
	_, err := Signature(context.Background(), &SignatureRequest{
		Algorithm: "SM2", Operation: "sign", Hash: "SHA256",
		Message: ByteInput{Data: "abc", Encoding: "utf8"},
	})
	if err == nil {
		t.Fatal("SM2 不接受 hash 参数")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedAlgorithm)
	}
}

// ── asymmetric-crypt ───────────────────────────────────────────────────

func TestAsymmetricCryptSM2RoundTrip(t *testing.T) {
	privPEM := testKey(t, "sm2.key")
	pubPEM := testKey(t, "sm2.pub")

	enc, err := AsymmetricCrypt(context.Background(), &AsymmetricCryptRequest{
		Algorithm: "SM2", Operation: "encrypt",
		Key:   KeyInput{Data: pubPEM, Format: "pem"},
		Input: ByteInput{Data: "待加密内容", Encoding: "utf8"},
	})
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	encData := enc.Data.(AsymmetricCryptResult)
	if encData.Order != "c1c3c2" {
		t.Errorf("密文顺序 = %q，期望 c1c3c2", encData.Order)
	}

	dec, err := AsymmetricCrypt(context.Background(), &AsymmetricCryptRequest{
		Algorithm: "SM2", Operation: "decrypt",
		Key:   KeyInput{Data: privPEM, Format: "pem"},
		Input: ByteInput{Data: encData.Output, Encoding: "base64"},
	})
	if err != nil {
		t.Fatalf("解密失败：%v", err)
	}
	decData := dec.Data.(AsymmetricCryptResult)
	// 输出默认是 base64；原文 "待加密内容" → "5b6F5Yqg5a+G5YaF5a65"。
	if decData.Output != "5b6F5Yqg5a+G5YaF5a65" {
		t.Errorf("解密结果 = %q", decData.Output)
	}
}

func TestAsymmetricCryptRejectsECIES(t *testing.T) {
	_, err := AsymmetricCrypt(context.Background(), &AsymmetricCryptRequest{
		Algorithm: "ECIES", Operation: "encrypt",
		Key:   KeyInput{Data: "x", Format: "pem"},
		Input: ByteInput{Data: "abc", Encoding: "utf8"},
	})
	if err == nil {
		t.Fatal("ECIES 应被拒绝")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedAlgorithm)
	}
}

// ── ecdh ───────────────────────────────────────────────────────────────

func TestECDHSharedSecret(t *testing.T) {
	privPEM := testKey(t, "ec.key")
	pubPEM := testKey(t, "ec.pub")

	result, err := ECDH(context.Background(), &ECDHRequest{
		Curve:         "P-256",
		PrivateKey:    KeyInput{Data: privPEM, Format: "pem"},
		PeerPublicKey: KeyInput{Data: pubPEM, Format: "pem"},
	})
	if err != nil {
		t.Fatalf("协商失败：%v", err)
	}
	data := result.Data.(ECDHResult)
	if data.SharedSecret == "" {
		t.Error("共享密钥不应为空")
	}
	if data.Length != 32 {
		t.Errorf("共享密钥长度 = %d，期望 32", data.Length)
	}
}

func TestECDHWithKDF(t *testing.T) {
	result, err := ECDH(context.Background(), &ECDHRequest{
		Curve:         "P-256",
		PrivateKey:    KeyInput{Data: testKey(t, "ec.key"), Format: "pem"},
		PeerPublicKey: KeyInput{Data: testKey(t, "ec.pub"), Format: "pem"},
		KDF: &KDFOptions{
			Type:   "hkdf",
			Hash:   "SHA256",
			Salt:   ByteInput{Data: "73616c74", Encoding: "hex"},
			Info:   ByteInput{Data: "696e666f", Encoding: "hex"},
			Length: 32,
		},
	})
	if err != nil {
		t.Fatalf("协商失败：%v", err)
	}
	data := result.Data.(ECDHResult)
	if data.DerivedKey == "" {
		t.Error("派生密钥不应为空")
	}
}

func TestECDHRejectsSM2(t *testing.T) {
	_, err := ECDH(context.Background(), &ECDHRequest{
		Curve:         "SM2",
		PrivateKey:    KeyInput{Data: testKey(t, "sm2.key"), Format: "pem"},
		PeerPublicKey: KeyInput{Data: testKey(t, "sm2.pub"), Format: "pem"},
	})
	if err == nil {
		t.Fatal("SM2 不是可协商曲线")
	}
	if code := codeOf(t, err); code != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %v，期望 %v", code, toolerr.CodeUnsupportedAlgorithm)
	}
}

// containsAny 报告字符串是否含有集合中任一字符。
func containsAny(s, set string) bool {
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(set); j++ {
			if s[i] == set[j] {
				return true
			}
		}
	}
	return false
}
