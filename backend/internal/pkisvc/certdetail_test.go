package pkisvc

import (
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 夹具目录：测试用的证书链。
//
// 材料由 `openssl` 一次性生成，仅用于本仓库的测试（见 AGENTS.md §5）。
const certFixtureDir = "../../testdata/fixtures/pki"

// loadCertFixture 读取一份证书夹具。
func loadCertFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(certFixtureDir, name))
	if err != nil {
		t.Fatalf("读取夹具失败：%v", err)
	}
	return data
}

// loadCert 加载一份证书夹具（调用方负责关闭）。
func loadCert(t *testing.T, name string) *Certificate {
	t.Helper()
	cert, err := LoadCertificate(loadCertFixture(t, name), true)
	if err != nil {
		t.Fatalf("加载证书失败：%v", err)
	}
	t.Cleanup(func() { _ = cert.Close() })
	return cert
}

// bigIntFromInt 构造一个小整数（测试辅助）。
func bigIntFromInt(value int64) *big.Int {
	return big.NewInt(value)
}

func TestPublicKeyDetailOf(t *testing.T) {
	cert := loadCert(t, "leaf.crt.pem")

	detail, err := PublicKeyDetailOf(cert)
	if err != nil {
		t.Fatalf("PublicKeyDetailOf 失败：%v", err)
	}
	if detail.Algorithm != AlgRSA {
		t.Errorf("算法 = %q，期望 %s", detail.Algorithm, AlgRSA)
	}
	if detail.Bits != 2048 {
		t.Errorf("位数 = %d，期望 2048", detail.Bits)
	}
	if !strings.HasPrefix(detail.PEM, "-----BEGIN PUBLIC KEY-----") {
		t.Errorf("公钥 PEM = %q，期望 SPKI PEM", detail.PEM)
	}
	if len(detail.Fingerprint) != 64 {
		t.Errorf("指纹长度 = %d，期望 64", len(detail.Fingerprint))
	}
	if detail.Modulus == "" {
		t.Error("RSA 模数为空")
	}
	if detail.Exponent != "010001" {
		t.Errorf("指数 = %q，期望 010001", detail.Exponent)
	}
	if _, err := hex.DecodeString(detail.SPKIHex); err != nil {
		t.Errorf("SPKI hex 非法：%v", err)
	}
}

func TestPublicKeyDetailOfClosedCertificate(t *testing.T) {
	cert := loadCert(t, "leaf.crt.pem")
	if err := cert.Close(); err != nil {
		t.Fatalf("关闭证书失败：%v", err)
	}
	if _, err := PublicKeyDetailOf(cert); err == nil {
		t.Fatal("已关闭证书应返回错误")
	}
}

func TestDescribeFillsExtensionText(t *testing.T) {
	cert := loadCert(t, "leaf.crt.pem")

	info, err := Describe(cert, zeroTime, nil)
	if err != nil {
		t.Fatalf("Describe 失败：%v", err)
	}
	if info.SignatureHex == "" {
		t.Error("SignatureHex 为空")
	}
	if len(info.SignatureHex)%2 != 0 {
		t.Errorf("SignatureHex = %q，期望偶数长度十六进制", info.SignatureHex)
	}

	wants := map[string]string{
		oidBasicConstraints: "CA:FALSE",
		oidKeyUsage:         "Digital Signature",
		oidExtKeyUsage:      "TLS Web Server Authentication",
		oidSubjectAltName:   "DNS:leaf.example.com",
		oidSubjectKeyID:     "keyid:",
		oidAuthorityKeyID:   "keyid:",
	}
	for _, ext := range info.Extensions {
		want, ok := wants[ext.OID]
		if !ok {
			continue
		}
		if !strings.Contains(ext.Text, want) {
			t.Errorf("扩展 %s 的文本 = %q，期望包含 %q", ext.OID, ext.Text, want)
		}
	}
}

func TestPoliciesOfWithoutPolicies(t *testing.T) {
	cert := loadCert(t, "leaf.crt.pem")
	if policies := PoliciesOf(cert); len(policies) != 0 {
		t.Errorf("策略数 = %d，期望 0（夹具没有 certificatePolicies）", len(policies))
	}
}

func TestCertTextView(t *testing.T) {
	cert := loadCert(t, "leaf.crt.pem")

	info, err := Describe(cert, zeroTime, nil)
	if err != nil {
		t.Fatalf("Describe 失败：%v", err)
	}
	detail, err := PublicKeyDetailOf(cert)
	if err != nil {
		t.Fatalf("PublicKeyDetailOf 失败：%v", err)
	}

	view := CertTextView(info, detail, "")
	for _, want := range []string{
		"Certificate:",
		"Version: 3 (0x2)",
		"Issuer: /C=CN/O=Toolbit Test/CN=Toolbit Test Intermediate CA",
		"Subject: /C=CN/O=Toolbit Test/CN=leaf.example.com",
		"Public Key Algorithm: rsaEncryption",
		"Public-Key: (2048 bit)",
		"X509v3 extensions:",
		"basicConstraints: critical",
		"Signature Algorithm: RSA-SHA256",
		"Signature Value:",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("文本视图缺少 %q", want)
		}
	}
}

func TestCertTextViewTruncates(t *testing.T) {
	cert := loadCert(t, "leaf.crt.pem")

	info, err := Describe(cert, zeroTime, nil)
	if err != nil {
		t.Fatalf("Describe 失败：%v", err)
	}
	view := CertTextView(info, nil, "")
	if len(view) > CertTextMaxLen+64 {
		t.Errorf("文本视图长度 = %d，期望不超过 %d", len(view), CertTextMaxLen+64)
	}
}

func TestDerItemsRejectsGarbage(t *testing.T) {
	if _, ok := derItems([]byte{0xff, 0xff, 0xff}); ok {
		t.Error("非法 DER 应返回 false")
	}
	if _, ok := derSequenceContent([]byte{0x02, 0x01, 0x00}); ok {
		t.Error("非 SEQUENCE 应返回 false")
	}
}

func TestOIDText(t *testing.T) {
	content, err := hex.DecodeString("2b06010505073001")
	if err != nil {
		t.Fatalf("解码失败：%v", err)
	}
	if got := oidText(content); got != "1.3.6.1.5.5.7.48.1" {
		t.Errorf("OID = %q，期望 1.3.6.1.5.5.7.48.1", got)
	}
	if got := oidText(nil); got != "" {
		t.Errorf("空输入 = %q，期望空串", got)
	}
}

func TestBigHex(t *testing.T) {
	if got := bigHex(nil); got != "" {
		t.Errorf("nil = %q，期望空串", got)
	}
	if got := bigHex(bigIntFromInt(0)); got != "" {
		t.Errorf("零 = %q，期望空串", got)
	}
	if got := bigHex(bigIntFromInt(65537)); got != "010001" {
		t.Errorf("65537 = %q，期望 010001", got)
	}
}

func TestIPText(t *testing.T) {
	if got := ipText([]byte{192, 0, 2, 10}); got != "192.0.2.10" {
		t.Errorf("IPv4 = %q，期望 192.0.2.10", got)
	}
	if got := ipText([]byte{1, 2, 3}); got != "010203" {
		t.Errorf("异常长度 = %q，期望十六进制回落", got)
	}
}
