package pkisvc

import (
	"encoding/pem"
	"errors"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// codeOf 取出错误码。
func codeOf(t *testing.T, err error) toolerr.Code {
	t.Helper()
	if err == nil {
		t.Fatal("期望有错误，实际为 nil")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误不是 toolerr.Error：%v", err)
	}
	return toolErr.Code
}

// generate 生成一把测试私钥，返回 PEM。
func generate(t *testing.T, algorithm string, bits int, curve string) []byte {
	t.Helper()
	pair, err := Generate(algorithm, bits, curve)
	if err != nil {
		t.Fatalf("生成 %s 密钥失败：%v", algorithm, err)
	}
	defer func() { _ = pair.Close() }()

	pemBytes, err := pair.Private.MarshalPKCS8PEM()
	if err != nil {
		t.Fatalf("导出私钥失败：%v", err)
	}
	return pemBytes
}

// TestGenerateAllAlgorithms 验证七种算法都能生成并往返。
func TestGenerateAllAlgorithms(t *testing.T) {
	cases := []struct {
		algorithm string
		bits      int
		curve     string
	}{
		{AlgRSA, 2048, ""},
		{AlgSM2, 0, CurveSM2},
		{AlgEC, 0, CurveP256},
		{AlgEd25519, 0, ""},
		{AlgEd448, 0, ""},
		{AlgX25519, 0, ""},
		{AlgX448, 0, ""},
	}

	for _, c := range cases {
		t.Run(c.algorithm, func(t *testing.T) {
			pair, err := Generate(c.algorithm, c.bits, c.curve)
			if err != nil {
				t.Fatalf("生成失败：%v", err)
			}
			defer func() { _ = pair.Close() }()

			if got := pair.Private.Algorithm(); got != c.algorithm {
				t.Errorf("算法 = %s，期望 %s", got, c.algorithm)
			}
			if pair.Public == nil {
				t.Fatal("公钥句柄为空")
			}

			pemBytes, err := pair.Private.MarshalPKCS8PEM()
			if err != nil {
				t.Fatalf("导出私钥失败：%v", err)
			}
			reloaded, err := LoadPrivateKey(pemBytes, "")
			if err != nil {
				t.Fatalf("重新加载私钥失败：%v", err)
			}
			defer func() { _ = reloaded.Close() }()

			again, err := reloaded.MarshalPKCS8PEM()
			if err != nil {
				t.Fatalf("再次导出失败：%v", err)
			}
			if string(again) != string(pemBytes) {
				t.Error("私钥 PEM 往返不一致")
			}
		})
	}
}

// TestGenerateRSABitsBoundary 验证 RSA 位数边界。
func TestGenerateRSABitsBoundary(t *testing.T) {
	if _, err := Generate(AlgRSA, 1023, ""); err == nil {
		t.Error("1023 位 RSA 应当被拒绝")
	} else if got := codeOf(t, err); got != toolerr.CodeKeySizeMismatch {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeKeySizeMismatch)
	}

	if _, err := Generate(AlgRSA, 1024, ""); err != nil {
		t.Errorf("1024 位 RSA 应当被接受：%v", err)
	}
}

// TestGenerateUnknownAlgorithm 验证未知算法被拒绝。
func TestGenerateUnknownAlgorithm(t *testing.T) {
	_, err := Generate("DSA", 0, "")
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedAlgorithm)
	}
}

// TestConvertPrivateToSPKIRejected 验证私钥转 SPKI 必须报错。
func TestConvertPrivateToSPKIRejected(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	_, err := Convert(ConvertOptions{
		Source:      source,
		SourceIsPEM: true,
		To:          FormatSPKI,
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

// TestConvertPublicToPKCS8Rejected 验证公钥转 PKCS#8 必须报错。
func TestConvertPublicToPKCS8Rejected(t *testing.T) {
	pair, err := Generate(AlgRSA, 2048, "")
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	defer func() { _ = pair.Close() }()
	pubPEM, err := pair.Public.MarshalPEM()
	if err != nil {
		t.Fatalf("导出公钥失败：%v", err)
	}

	_, err = Convert(ConvertOptions{
		Source:      pubPEM,
		SourceIsPEM: true,
		To:          FormatPKCS8,
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

// TestConvertEncryptedRequiresPassphrase 验证加密私钥缺少口令时报 PASSPHRASE_REQUIRED。
func TestConvertEncryptedRequiresPassphrase(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")
	encrypted, err := EncryptPrivateKey(source, "secret", PBEAES256CBC)
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}

	_, err = Convert(ConvertOptions{
		Source:      encrypted,
		SourceIsPEM: true,
		To:          FormatPKCS8,
	})
	if got := codeOf(t, err); got != toolerr.CodePassphraseRequired {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodePassphraseRequired)
	}
}

// TestConvertEncryptedPreservesEncryption 验证转换后仍保持加密。
func TestConvertEncryptedPreservesEncryption(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")
	encrypted, err := EncryptPrivateKey(source, "secret", PBEAES256CBC)
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}

	result, err := Convert(ConvertOptions{
		Source:      encrypted,
		SourceIsPEM: true,
		Passphrase:  "secret",
		To:          FormatPEM,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if !result.Encrypted {
		t.Error("转换结果应当仍是加密私钥")
	}
	if !strings.Contains(string(result.PEM), "ENCRYPTED PRIVATE KEY") {
		t.Error("转换结果不是加密私钥块")
	}
}

// TestConvertDERRoundTrip 验证 PEM → DER → PEM 往返一致。
func TestConvertDERRoundTrip(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	toDER, err := Convert(ConvertOptions{
		Source:      source,
		SourceIsPEM: true,
		To:          FormatDER,
	})
	if err != nil {
		t.Fatalf("转 DER 失败：%v", err)
	}
	if len(toDER.DER) == 0 {
		t.Fatal("DER 输出为空")
	}

	back, err := Convert(ConvertOptions{
		Source:      toDER.DER,
		SourceIsPEM: false,
		To:          FormatPKCS8,
	})
	if err != nil {
		t.Fatalf("DER 转回失败：%v", err)
	}
	if string(back.PEM) != string(source) {
		t.Error("DER 往返后 PEM 不一致")
	}
}

// TestConvertPKCS1RoundTrip 验证 PKCS#1 私钥与 PKCS#8 互转。
func TestConvertPKCS1RoundTrip(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	toPKCS1, err := Convert(ConvertOptions{
		Source:      source,
		SourceIsPEM: true,
		To:          FormatPKCS1,
	})
	if err != nil {
		t.Fatalf("转 PKCS#1 失败：%v", err)
	}
	block, _ := pem.Decode(toPKCS1.PEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Fatalf("PKCS#1 输出块类型异常：%v", block)
	}

	back, err := Convert(ConvertOptions{
		Source:      toPKCS1.PEM,
		SourceIsPEM: true,
		To:          FormatPKCS8,
	})
	if err != nil {
		t.Fatalf("PKCS#1 转 PKCS#8 失败：%v", err)
	}
	if string(back.PEM) != string(source) {
		t.Error("PKCS#1 往返后 PEM 不一致")
	}
}

// TestConvertSourceFormatConstraint 验证 from 约束不匹配时报错。
func TestConvertSourceFormatConstraint(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	_, err := Convert(ConvertOptions{
		Source:      source,
		SourceIsPEM: true,
		From:        FormatDER,
		To:          FormatPKCS8,
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

// TestInspectReturnsRSAParams 验证解析私钥能拿到 RSA 参数与指纹。
func TestInspectReturnsRSAParams(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	info, normalized, err := Inspect(source, true, "")
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if info.Algorithm != AlgRSA {
		t.Errorf("算法 = %s，期望 %s", info.Algorithm, AlgRSA)
	}
	if info.KeyType != KeyTypePrivate {
		t.Errorf("密钥类型 = %s，期望 %s", info.KeyType, KeyTypePrivate)
	}
	if info.Bits != 2048 {
		t.Errorf("位数 = %d，期望 2048", info.Bits)
	}
	if info.RSA == nil || info.RSA.N == "" || info.RSA.E != "10001" {
		t.Errorf("RSA 参数异常：%+v", info.RSA)
	}
	if len(info.FingerprintSHA256) != 64 {
		t.Errorf("指纹长度 = %d，期望 64", len(info.FingerprintSHA256))
	}
	if !strings.Contains(string(normalized), "BEGIN PRIVATE KEY") {
		t.Error("规范化输出不是私钥 PEM")
	}
}

// TestEncryptDecryptRoundTrip 验证加密 → 解密往返一致。
func TestEncryptDecryptRoundTrip(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	encrypted, err := EncryptPrivateKey(source, "secret", PBEAES256CBC)
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	decrypted, err := DecryptPrivateKey(encrypted, "secret")
	if err != nil {
		t.Fatalf("解密失败：%v", err)
	}
	if string(decrypted) != string(source) {
		t.Error("加解密往返不一致")
	}
}

// TestDecryptWrongPassphrase 验证错误口令报 PASSPHRASE_INCORRECT。
func TestDecryptWrongPassphrase(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")
	encrypted, err := EncryptPrivateKey(source, "secret", PBEAES256CBC)
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}

	_, err = DecryptPrivateKey(encrypted, "wrong")
	if got := codeOf(t, err); got != toolerr.CodePassphraseIncorrect {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodePassphraseIncorrect)
	}
}

// TestChangePassphrase 验证改密后旧口令失效、新口令可用。
func TestChangePassphrase(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")
	encrypted, err := EncryptPrivateKey(source, "old", PBEAES256CBC)
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}

	changed, err := ChangePassphrase(encrypted, "old", "new")
	if err != nil {
		t.Fatalf("改密失败：%v", err)
	}
	if _, err := LoadPrivateKey(changed, "old"); err == nil {
		t.Error("旧口令应当失效")
	}
	if _, err := LoadPrivateKey(changed, "new"); err != nil {
		t.Errorf("新口令应当可用：%v", err)
	}
}

// TestMatchPrivateAndPublic 验证私钥与公钥配对。
func TestMatchPrivateAndPublic(t *testing.T) {
	pair, err := Generate(AlgRSA, 2048, "")
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	defer func() { _ = pair.Close() }()

	privPEM, err := pair.Private.MarshalPKCS8PEM()
	if err != nil {
		t.Fatalf("导出私钥失败：%v", err)
	}
	pubPEM, err := pair.Public.MarshalPEM()
	if err != nil {
		t.Fatalf("导出公钥失败：%v", err)
	}

	result, err := Match(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("配对检测失败：%v", err)
	}
	if !result.Match {
		t.Errorf("应当配对，实际结论 %s", result.Reason)
	}
	if result.Reason != MatchMatched {
		t.Errorf("结论 = %s，期望 %s", result.Reason, MatchMatched)
	}
	if result.AFingerprint != result.BFingerprint {
		t.Error("配对双方的指纹应当相同")
	}
}

// TestMatchIdenticalPrivate 验证两把相同私钥的结论是 identical。
func TestMatchIdenticalPrivate(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	result, err := Match(source, source)
	if err != nil {
		t.Fatalf("配对检测失败：%v", err)
	}
	if !result.Match || result.Reason != MatchIdentical {
		t.Errorf("结论 = %s，期望 %s", result.Reason, MatchIdentical)
	}
}

// TestMatchDifferentAlgorithms 验证不同算法不配对。
func TestMatchDifferentAlgorithms(t *testing.T) {
	rsa := generate(t, AlgRSA, 2048, "")
	ec := generate(t, AlgEC, 0, CurveP256)

	result, err := Match(rsa, ec)
	if err != nil {
		t.Fatalf("配对检测失败：%v", err)
	}
	if result.Match {
		t.Error("不同算法不应配对")
	}
	if result.Reason != MatchNotPair {
		t.Errorf("结论 = %s，期望 %s", result.Reason, MatchNotPair)
	}
}

// TestJWKRoundTrip 验证 PEM → JWK → PEM 往返。
func TestJWKRoundTrip(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	jwkJSON, err := JWKFromPEM(source, true)
	if err != nil {
		t.Fatalf("转 JWK 失败：%v", err)
	}
	if !strings.Contains(string(jwkJSON), `"kty": "RSA"`) {
		t.Errorf("JWK 内容异常：%s", jwkJSON)
	}

	back, err := PEMFromJWK(jwkJSON, true)
	if err != nil {
		t.Fatalf("JWK 转 PEM 失败：%v", err)
	}
	if string(back) != string(source) {
		t.Error("JWK 往返后 PEM 不一致")
	}
}

// TestJWKPublicHasNoPrivateMaterial 验证公钥 JWK 不含私钥材料。
func TestJWKPublicHasNoPrivateMaterial(t *testing.T) {
	pair, err := Generate(AlgRSA, 2048, "")
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	defer func() { _ = pair.Close() }()
	pubPEM, err := pair.Public.MarshalPEM()
	if err != nil {
		t.Fatalf("导出公钥失败：%v", err)
	}

	jwkJSON, err := JWKFromPEM(pubPEM, false)
	if err != nil {
		t.Fatalf("转 JWK 失败：%v", err)
	}
	if strings.Contains(string(jwkJSON), `"d"`) {
		t.Errorf("公钥 JWK 不应含私钥材料：%s", jwkJSON)
	}
}

// TestXMLRoundTrip 验证 RSA 的 PEM ↔ XML 往返。
func TestXMLRoundTrip(t *testing.T) {
	source := generate(t, AlgRSA, 2048, "")

	xmlBytes, err := XMLFromPEM(source, true)
	if err != nil {
		t.Fatalf("转 XML 失败：%v", err)
	}
	if !strings.Contains(string(xmlBytes), "<RSAKeyValue>") {
		t.Errorf("XML 内容异常：%s", xmlBytes)
	}

	back, err := PEMFromXML(xmlBytes, true)
	if err != nil {
		t.Fatalf("XML 转 PEM 失败：%v", err)
	}
	if string(back) != string(source) {
		t.Error("XML 往返后 PEM 不一致")
	}
}

// TestXMLRejectsNonRSA 验证 XML 只支持 RSA。
func TestXMLRejectsNonRSA(t *testing.T) {
	source := generate(t, AlgEC, 0, CurveP256)

	_, err := XMLFromPEM(source, true)
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

// TestJWKRejectsInvalidJSON 验证非法 JWK 报 INVALID_JWK。
func TestJWKRejectsInvalidJSON(t *testing.T) {
	_, err := PEMFromJWK([]byte("not json"), true)
	if got := codeOf(t, err); got != toolerr.CodeInvalidJWK {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidJWK)
	}
}
