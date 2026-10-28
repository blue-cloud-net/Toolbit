package keys

import (
	"context"
	"errors"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
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

// generatePEM 生成一把测试私钥并返回 PEM 文本。
func generatePEM(t *testing.T, algorithm string, bits int, curve string) string {
	t.Helper()
	pair, err := pkisvc.Generate(algorithm, bits, curve)
	if err != nil {
		t.Fatalf("生成密钥失败：%v", err)
	}
	defer func() { _ = pair.Close() }()

	pemBytes, err := pair.Private.MarshalPKCS8PEM()
	if err != nil {
		t.Fatalf("导出私钥失败：%v", err)
	}
	return string(pemBytes)
}

// TestKeyGenRejectsUnknownAlgorithm 验证未知算法被拒绝。
func TestKeyGenRejectsUnknownAlgorithm(t *testing.T) {
	_, err := KeyGen(context.Background(), &KeyGenRequest{Algorithm: "DSA"})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedAlgorithm)
	}
}

// TestKeyGenRejectsBatch 验证批量生成被拒绝（响应结构只能表达一把密钥）。
func TestKeyGenRejectsBatch(t *testing.T) {
	_, err := KeyGen(context.Background(), &KeyGenRequest{Algorithm: pkisvc.AlgRSA, Count: 3})
	if got := codeOf(t, err); got != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidField)
	}
}

// TestKeyGenRequiresCurve 验证 EC 缺少曲线时报错。
func TestKeyGenRequiresCurve(t *testing.T) {
	_, err := KeyGen(context.Background(), &KeyGenRequest{Algorithm: pkisvc.AlgEC})
	if got := codeOf(t, err); got != toolerr.CodeMissingField {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeMissingField)
	}
}

// TestKeyGenProducesPair 验证生成的密钥对自洽。
func TestKeyGenProducesPair(t *testing.T) {
	res, err := KeyGen(context.Background(), &KeyGenRequest{Algorithm: pkisvc.AlgRSA, Bits: 2048})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	data, ok := res.Data.(*KeyGenResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if data.Algorithm != pkisvc.AlgRSA {
		t.Errorf("算法 = %s，期望 %s", data.Algorithm, pkisvc.AlgRSA)
	}
	if data.Bits != 2048 {
		t.Errorf("位数 = %d，期望 2048", data.Bits)
	}
	if data.PrivateKey.PEM == "" || data.PublicKey.PEM == "" {
		t.Fatal("私钥或公钥为空")
	}
}

// TestKeyGenRejectsPKCS1Public 验证 PKCS#1 公钥格式不受支持。
func TestKeyGenRejectsPKCS1Public(t *testing.T) {
	_, err := KeyGen(context.Background(), &KeyGenRequest{
		Algorithm:    pkisvc.AlgRSA,
		PublicFormat: PublicFormatPKCS1,
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

// TestKeyConvertRejectsPrivateToSPKI 验证私钥转 SPKI 被拒绝。
func TestKeyConvertRejectsPrivateToSPKI(t *testing.T) {
	pemText := generatePEM(t, pkisvc.AlgRSA, 2048, "")

	_, err := KeyConvert(context.Background(), &KeyConvertRequest{
		Key: KeyInput{Data: pemText, Format: FormatPEM},
		To:  pkisvc.FormatSPKI,
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

// TestKeyConvertReportsDetected 验证转换结果报告识别到的源格式。
func TestKeyConvertReportsDetected(t *testing.T) {
	pemText := generatePEM(t, pkisvc.AlgRSA, 2048, "")

	res, err := KeyConvert(context.Background(), &KeyConvertRequest{
		Key: KeyInput{Data: pemText, Format: FormatPEM},
		To:  pkisvc.FormatPKCS8,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	data, ok := res.Data.(*KeyConvertResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if data.Detected.Algorithm != pkisvc.AlgRSA {
		t.Errorf("算法 = %s，期望 %s", data.Detected.Algorithm, pkisvc.AlgRSA)
	}
	if data.Detected.Format != pkisvc.FormatPKCS8 {
		t.Errorf("格式 = %s，期望 %s", data.Detected.Format, pkisvc.FormatPKCS8)
	}
}

// TestKeyParseReturnsPrivateParams 验证解析私钥返回 RSA 参数。
func TestKeyParseReturnsPrivateParams(t *testing.T) {
	pemText := generatePEM(t, pkisvc.AlgRSA, 2048, "")

	res, err := KeyParse(context.Background(), &KeyParseRequest{
		Key: KeyInput{Data: pemText, Format: FormatPEM},
	})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	data, ok := res.Data.(*KeyParseResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if data.KeyType != pkisvc.KeyTypePrivate {
		t.Errorf("密钥类型 = %s，期望 %s", data.KeyType, pkisvc.KeyTypePrivate)
	}
	if data.Params == nil || data.Params.E != "10001" {
		t.Errorf("RSA 参数异常：%+v", data.Params)
	}
	if data.FingerprintSHA256 == "" {
		t.Error("指纹为空")
	}
}

// TestKeyParsePublicHasNoParams 验证公钥不返回私钥参数。
func TestKeyParsePublicHasNoParams(t *testing.T) {
	pair, err := pkisvc.Generate(pkisvc.AlgRSA, 2048, "")
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	defer func() { _ = pair.Close() }()
	pubPEM, err := pair.Public.MarshalPEM()
	if err != nil {
		t.Fatalf("导出公钥失败：%v", err)
	}

	res, err := KeyParse(context.Background(), &KeyParseRequest{
		Key: KeyInput{Data: string(pubPEM), Format: FormatPEM},
	})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	data, ok := res.Data.(*KeyParseResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if data.Params != nil {
		t.Error("公钥不应返回私钥参数")
	}
	if data.KeyType != pkisvc.KeyTypePublic {
		t.Errorf("密钥类型 = %s，期望 %s", data.KeyType, pkisvc.KeyTypePublic)
	}
}

// TestKeyOpsExtractPublic 验证提取公钥不需要口令。
func TestKeyOpsExtractPublic(t *testing.T) {
	pemText := generatePEM(t, pkisvc.AlgRSA, 2048, "")

	res, err := KeyOps(context.Background(), &KeyOpsRequest{
		Action: ActionExtractPublic,
		Key:    KeyInput{Data: pemText, Format: FormatPEM},
	})
	if err != nil {
		t.Fatalf("提取公钥失败：%v", err)
	}
	data, ok := res.Data.(*KeyOpsResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if data.Format != pkisvc.FormatSPKI {
		t.Errorf("格式 = %s，期望 %s", data.Format, pkisvc.FormatSPKI)
	}
}

// TestKeyOpsEncryptDecrypt 验证加密后再解密可用。
func TestKeyOpsEncryptDecrypt(t *testing.T) {
	pemText := generatePEM(t, pkisvc.AlgRSA, 2048, "")

	encrypted, err := KeyOps(context.Background(), &KeyOpsRequest{
		Action:     ActionEncrypt,
		Key:        KeyInput{Data: pemText, Format: FormatPEM},
		Passphrase: "secret",
	})
	if err != nil {
		t.Fatalf("加密失败：%v", err)
	}
	encData, ok := encrypted.Data.(*KeyOpsResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", encrypted.Data)
	}

	decrypted, err := KeyOps(context.Background(), &KeyOpsRequest{
		Action:     ActionDecrypt,
		Key:        KeyInput{Data: encData.Key.PEM, Format: FormatPEM},
		Passphrase: "secret",
	})
	if err != nil {
		t.Fatalf("解密失败：%v", err)
	}
	decData, ok := decrypted.Data.(*KeyOpsResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", decrypted.Data)
	}
	if decData.Key.PEM != pemText {
		t.Error("加解密往返不一致")
	}
}

// TestKeyOpsRejectsUnknownAction 验证未知动作被拒绝。
func TestKeyOpsRejectsUnknownAction(t *testing.T) {
	_, err := KeyOps(context.Background(), &KeyOpsRequest{
		Action: "sign",
		Key:    KeyInput{Data: "x", Format: FormatPEM},
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedMode {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedMode)
	}
}

// TestKeyMatchReturnsBusinessResult 验证不配对是业务结果而非错误。
func TestKeyMatchReturnsBusinessResult(t *testing.T) {
	a := generatePEM(t, pkisvc.AlgRSA, 2048, "")
	b := generatePEM(t, pkisvc.AlgEC, 0, pkisvc.CurveP256)

	res, err := KeyMatch(context.Background(), &KeyMatchRequest{
		A: KeyInput{Data: a, Format: FormatPEM},
		B: KeyInput{Data: b, Format: FormatPEM},
	})
	if err != nil {
		t.Fatalf("配对检测失败：%v", err)
	}
	data, ok := res.Data.(*KeyMatchResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if data.Match {
		t.Error("不同算法不应配对")
	}
	if data.Reason != pkisvc.MatchNotPair {
		t.Errorf("结论 = %s，期望 %s", data.Reason, pkisvc.MatchNotPair)
	}
}

// TestKeyMatchRejectsInvalidInput 验证无法解析的输入报错。
func TestKeyMatchRejectsInvalidInput(t *testing.T) {
	b := generatePEM(t, pkisvc.AlgRSA, 2048, "")

	_, err := KeyMatch(context.Background(), &KeyMatchRequest{
		A: KeyInput{Data: "not-a-key", Format: FormatPEM},
		B: KeyInput{Data: b, Format: FormatPEM},
	})
	if got := codeOf(t, err); got != toolerr.CodeInvalidKey {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidKey)
	}
}

// TestJWKXMLConvertRoundTrip 验证 PEM → JWK → PEM 往返。
func TestJWKXMLConvertRoundTrip(t *testing.T) {
	pemText := generatePEM(t, pkisvc.AlgRSA, 2048, "")

	toJWK, err := JWKXMLConvert(context.Background(), &JWKXMLConvertRequest{
		Input:   MaterialInput{Format: MaterialFormatPEM, Data: pemText},
		To:      MaterialFormatJWK,
		KeyType: KeyTypePrivate,
	})
	if err != nil {
		t.Fatalf("转 JWK 失败：%v", err)
	}
	jwkData, ok := toJWK.Data.(*JWKXMLConvertResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", toJWK.Data)
	}

	back, err := JWKXMLConvert(context.Background(), &JWKXMLConvertRequest{
		Input:   MaterialInput{Format: MaterialFormatJWK, Data: jwkData.Output.Data},
		To:      MaterialFormatPEM,
		KeyType: KeyTypePrivate,
	})
	if err != nil {
		t.Fatalf("JWK 转 PEM 失败：%v", err)
	}
	backData, ok := back.Data.(*JWKXMLConvertResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", back.Data)
	}
	if backData.Output.Data != pemText {
		t.Error("JWK 往返后 PEM 不一致")
	}
}

// TestJWKXMLConvertRejectsECToXML 验证 EC 转 XML 被拒绝。
func TestJWKXMLConvertRejectsECToXML(t *testing.T) {
	pemText := generatePEM(t, pkisvc.AlgEC, 0, pkisvc.CurveP256)

	_, err := JWKXMLConvert(context.Background(), &JWKXMLConvertRequest{
		Input:   MaterialInput{Format: MaterialFormatPEM, Data: pemText},
		To:      MaterialFormatXML,
		KeyType: KeyTypePrivate,
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}
