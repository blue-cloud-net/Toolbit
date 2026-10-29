package pki

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 夹具目录：测试用的证书链、CSR 与 CRL。
//
// 这些材料由 `openssl` **一次性生成**，仅用于本仓库的测试，
// **不是**任何真实系统的密钥或证书（见 AGENTS.md §5）。
const fixtureDir = "../../../testdata/fixtures/pki"

// loadFixture 读取一份夹具文件。
func loadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(fixtureDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取夹具失败：%v", err)
	}
	return string(data)
}

// loadFixtureBase64 读取一份**二进制**夹具并返回 base64 文本。
func loadFixtureBase64(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(fixtureDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取夹具失败：%v", err)
	}
	return base64.StdEncoding.EncodeToString(data)
}

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

func TestCertParse(t *testing.T) {
	res, err := CertParse(context.Background(), &CertParseRequest{
		Input: PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("CertParse 失败：%v", err)
	}
	data, ok := res.Data.(*CertParseResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if !strings.Contains(data.Certificate.SubjectText, "leaf.example.com") {
		t.Errorf("主体 = %q，期望包含 leaf.example.com", data.Certificate.SubjectText)
	}
	if data.Certificate.PublicKey.Algorithm != "RSA" {
		t.Errorf("算法 = %q，期望 RSA", data.Certificate.PublicKey.Algorithm)
	}
	if data.Certificate.PublicKey.Bits != 2048 {
		t.Errorf("位数 = %d，期望 2048", data.Certificate.PublicKey.Bits)
	}
	if data.Validity.Status != "valid" {
		t.Errorf("有效性 = %q，期望 valid", data.Validity.Status)
	}
	if data.SelfSigned {
		t.Error("SelfSigned = true，期望 false")
	}
	if len(data.Fingerprints) != 1 {
		t.Fatalf("指纹数量 = %d，期望 1（默认只算 SHA-256）", len(data.Fingerprints))
	}
	if data.Fingerprints[0].Algorithm != "sha256" {
		t.Errorf("指纹算法 = %q，期望 sha256", data.Fingerprints[0].Algorithm)
	}
	if data.RawDerBase64 == "" {
		t.Error("RawDerBase64 为空")
	}
}

func TestCertParseAllFingerprintAlgorithms(t *testing.T) {
	res, err := CertParse(context.Background(), &CertParseRequest{
		Input:        PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		Fingerprints: []string{"md5", "sha1", "sha256", "sha384", "sha512", "sm3"},
	})
	if err != nil {
		t.Fatalf("CertParse 失败：%v", err)
	}
	data := res.Data.(*CertParseResult)
	if len(data.Fingerprints) != 6 {
		t.Fatalf("指纹数量 = %d，期望 6", len(data.Fingerprints))
	}
	for i, want := range []string{"md5", "sha1", "sha256", "sha384", "sha512", "sm3"} {
		if data.Fingerprints[i].Algorithm != want {
			t.Errorf("第 %d 个算法 = %q，期望 %q", i, data.Fingerprints[i].Algorithm, want)
		}
	}
}

func TestCertParseRejectsUnknownFingerprint(t *testing.T) {
	_, err := CertParse(context.Background(), &CertParseRequest{
		Input:        PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		Fingerprints: []string{"sha999"},
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedAlgorithm)
	}
}

func TestCertParseExpired(t *testing.T) {
	res, err := CertParse(context.Background(), &CertParseRequest{
		Input: PemBundle{Data: loadFixture(t, "expired.crt.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("CertParse 失败：%v", err)
	}
	data := res.Data.(*CertParseResult)
	if data.Validity.Status != "expired" {
		t.Errorf("有效性 = %q，期望 expired", data.Validity.Status)
	}
	if len(res.Warnings) == 0 {
		t.Error("过期证书应有警告")
	}
}

func TestCertParseInvalidTime(t *testing.T) {
	_, err := CertParse(context.Background(), &CertParseRequest{
		Input:  PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		AtTime: "2026/01/01",
	})
	if got := codeOf(t, err); got != toolerr.CodeInvalidTime {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidTime)
	}
}

func TestCertParseInvalidCertificate(t *testing.T) {
	_, err := CertParse(context.Background(), &CertParseRequest{
		Input: PemBundle{Data: "not a certificate", Format: "pem"},
	})
	if got := codeOf(t, err); got != toolerr.CodeInvalidCert {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidCert)
	}
}

func TestCSRParse(t *testing.T) {
	res, err := CSRParse(context.Background(), &CSRParseRequest{
		Input: PemBundle{Data: loadFixture(t, "leaf.csr.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("CSRParse 失败：%v", err)
	}
	data := res.Data.(*CSRParseResult)
	if !strings.Contains(data.SubjectText, "leaf.example.com") {
		t.Errorf("主体 = %q，期望包含 leaf.example.com", data.SubjectText)
	}
	if !data.SignatureValid {
		t.Error("SignatureValid = false，期望 true")
	}
	if data.ChallengePasswordPresent {
		t.Error("ChallengePasswordPresent = true，期望 false")
	}
	if len(data.FingerprintSha256) != 64 {
		t.Errorf("指纹长度 = %d，期望 64", len(data.FingerprintSha256))
	}
}

func TestCSRParseInvalid(t *testing.T) {
	_, err := CSRParse(context.Background(), &CSRParseRequest{
		Input: PemBundle{Data: "not a csr", Format: "pem"},
	})
	if got := codeOf(t, err); got != toolerr.CodeInvalidCSR {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidCSR)
	}
}

func TestPEMDerToDER(t *testing.T) {
	res, err := PEMDer(context.Background(), &PEMDerRequest{
		Input: PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		Type:  TypeCert,
		To:    ToDER,
	})
	if err != nil {
		t.Fatalf("PEMDer 失败：%v", err)
	}
	data := res.Data.(*PEMDerResult)
	if data.Output.Format != "der" {
		t.Errorf("格式 = %q，期望 der", data.Output.Format)
	}
	if data.Output.Encoding != "base64" {
		t.Errorf("编码 = %q，期望 base64", data.Output.Encoding)
	}
	if data.Output.Data == "" {
		t.Error("输出为空")
	}
	if data.BlockCount != 1 {
		t.Errorf("块数 = %d，期望 1", data.BlockCount)
	}
}

func TestPEMDerToPEM(t *testing.T) {
	der := loadFixtureBase64(t, "leaf.crt.der")
	res, err := PEMDer(context.Background(), &PEMDerRequest{
		Input: PemBundle{Data: der, Format: "der"},
		Type:  TypeCert,
		To:    ToPEM,
	})
	if err != nil {
		t.Fatalf("PEMDer 失败：%v", err)
	}
	data := res.Data.(*PEMDerResult)
	if !strings.Contains(data.Output.Data, "BEGIN CERTIFICATE") {
		t.Errorf("输出不含 CERTIFICATE 块：%q", data.Output.Data)
	}
}

func TestPEMDerRejectsMultipleBlocksToDER(t *testing.T) {
	_, err := PEMDer(context.Background(), &PEMDerRequest{
		Input: PemBundle{Data: loadFixture(t, "chain.pem"), Format: "pem"},
		Type:  TypeCert,
		To:    ToDER,
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

func TestPEMDerRejectsUnknownTarget(t *testing.T) {
	_, err := PEMDer(context.Background(), &PEMDerRequest{
		Input: PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		To:    "hex",
	})
	if got := codeOf(t, err); got != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeUnsupportedFormat)
	}
}

func TestASN1View(t *testing.T) {
	res, err := ASN1View(context.Background(), &ASN1ViewRequest{
		Input: PemBundle{Data: loadFixtureBase64(t, "leaf.crt.der"), Format: "der"},
	})
	if err != nil {
		t.Fatalf("ASN1View 失败：%v", err)
	}
	data := res.Data.(*ASN1ViewResult)
	if data.Tree.Number != 16 {
		t.Errorf("根节点 = %d，期望 16（SEQUENCE）", data.Tree.Number)
	}
	if data.NodeCount < 3 {
		t.Errorf("节点数 = %d，期望 >= 3", data.NodeCount)
	}
	if !strings.Contains(data.HexDump, "00000000:") {
		t.Errorf("十六进制转储格式异常：%q", data.HexDump[:minInt(len(data.HexDump), 40)])
	}
}

func TestASN1ViewRejectsBadDepth(t *testing.T) {
	_, err := ASN1View(context.Background(), &ASN1ViewRequest{
		Input:    PemBundle{Data: loadFixtureBase64(t, "leaf.crt.der"), Format: "der"},
		MaxDepth: 1000,
	})
	if got := codeOf(t, err); got != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidField)
	}
}

func TestChainParse(t *testing.T) {
	res, err := ChainParse(context.Background(), &ChainParseRequest{
		Bundle: PemBundle{Data: loadFixture(t, "chain.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("ChainParse 失败：%v", err)
	}
	data := res.Data.(*ChainParseResult)
	if data.Count != 3 {
		t.Errorf("证书数 = %d，期望 3", data.Count)
	}
	if len(data.Chain) != 3 {
		t.Fatalf("链长度 = %d，期望 3", len(data.Chain))
	}
	if !strings.Contains(data.Chain[0].Info.SubjectText, "leaf.example.com") {
		t.Errorf("链首 = %q，期望叶证书", data.Chain[0].Info.SubjectText)
	}
	if !strings.Contains(data.Chain[2].Info.SubjectText, "Root CA") {
		t.Errorf("链尾 = %q，期望根证书", data.Chain[2].Info.SubjectText)
	}
	if !data.SelfSigned {
		t.Error("SelfSigned = false，期望 true")
	}
	if data.MissingRoot {
		t.Error("MissingRoot = true，期望 false")
	}
}

func TestChainParseReordersInput(t *testing.T) {
	root := loadFixture(t, "ca.crt.pem")
	inter := loadFixture(t, "inter.crt.pem")
	leaf := loadFixture(t, "leaf.crt.pem")

	res, err := ChainParse(context.Background(), &ChainParseRequest{
		Bundle: PemBundle{Data: root + inter + leaf, Format: "pem"},
	})
	if err != nil {
		t.Fatalf("ChainParse 失败：%v", err)
	}
	data := res.Data.(*ChainParseResult)
	if data.Ordered {
		t.Error("Ordered = true，期望 false（输入是乱的）")
	}
	if !strings.Contains(data.Chain[0].Info.SubjectText, "leaf.example.com") {
		t.Errorf("链首 = %q，期望叶证书", data.Chain[0].Info.SubjectText)
	}
}

func TestChainVerify(t *testing.T) {
	res, err := ChainVerify(context.Background(), &ChainVerifyRequest{
		Bundle:       PemBundle{Data: loadFixture(t, "chain.pem"), Format: "pem"},
		TrustedRoots: []PemBundle{{Data: loadFixture(t, "ca.crt.pem"), Format: "pem"}},
		Host:         "leaf.example.com",
	})
	if err != nil {
		t.Fatalf("ChainVerify 失败：%v", err)
	}
	data := res.Data.(*ChainVerifyResult)
	if !data.Valid {
		t.Errorf("Valid = false，期望 true；errors=%v", data.Errors)
	}
	if !data.HostnameMatch {
		t.Error("HostnameMatch = false，期望 true")
	}
	if data.ImplicitTrust {
		t.Error("ImplicitTrust = true，期望 false（显式给了 trustedRoots）")
	}
}

func TestChainVerifyHostnameMismatch(t *testing.T) {
	res, err := ChainVerify(context.Background(), &ChainVerifyRequest{
		Bundle:       PemBundle{Data: loadFixture(t, "chain.pem"), Format: "pem"},
		TrustedRoots: []PemBundle{{Data: loadFixture(t, "ca.crt.pem"), Format: "pem"}},
		Host:         "other.example.com",
	})
	if err != nil {
		t.Fatalf("ChainVerify 失败：%v", err)
	}
	data := res.Data.(*ChainVerifyResult)
	if data.HostnameMatch {
		t.Error("HostnameMatch = true，期望 false")
	}
	if data.HostnameError == "" {
		t.Error("HostnameError 为空")
	}
}

func TestChainVerifyImplicitTrust(t *testing.T) {
	res, err := ChainVerify(context.Background(), &ChainVerifyRequest{
		Bundle: PemBundle{Data: loadFixture(t, "chain.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("ChainVerify 失败：%v", err)
	}
	data := res.Data.(*ChainVerifyResult)
	if !data.ImplicitTrust {
		t.Error("ImplicitTrust = false，期望 true")
	}
}

func TestChainVerifyInvalidCertificate(t *testing.T) {
	_, err := ChainVerify(context.Background(), &ChainVerifyRequest{
		Bundle: PemBundle{Data: "not a certificate", Format: "pem"},
	})
	if err == nil {
		t.Fatal("非法证书应报错")
	}
}

func TestChainRepair(t *testing.T) {
	res, err := ChainRepair(context.Background(), &ChainRepairRequest{
		Cert:          PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		Intermediates: []PemBundle{{Data: loadFixture(t, "inter.crt.pem"), Format: "pem"}},
		TrustedRoots:  []PemBundle{{Data: loadFixture(t, "ca.crt.pem"), Format: "pem"}},
	})
	if err != nil {
		t.Fatalf("ChainRepair 失败：%v", err)
	}
	data := res.Data.(*ChainRepairResult)
	if data.AddedCount != 2 {
		t.Errorf("补充数 = %d，期望 2", data.AddedCount)
	}
	if len(data.ChainPem) != 3 {
		t.Errorf("链长度 = %d，期望 3", len(data.ChainPem))
	}
	if !data.ReachedRoot {
		t.Error("ReachedRoot = false，期望 true")
	}
	if data.Note != "reached_root" {
		t.Errorf("结论 = %q，期望 reached_root", data.Note)
	}
	if len(data.AddedFroms) != 2 || data.AddedFroms[0] != "intermediates" {
		t.Errorf("来源 = %v，期望首项为 intermediates", data.AddedFroms)
	}
}

func TestChainRepairNoCandidates(t *testing.T) {
	res, err := ChainRepair(context.Background(), &ChainRepairRequest{
		Cert: PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("ChainRepair 失败：%v", err)
	}
	data := res.Data.(*ChainRepairResult)
	if data.AddedCount != 0 {
		t.Errorf("补充数 = %d，期望 0", data.AddedCount)
	}
	if data.Note != "already_complete" {
		t.Errorf("结论 = %q，期望 already_complete", data.Note)
	}
}

func TestCRLParse(t *testing.T) {
	res, err := CRLParse(context.Background(), &CRLParseRequest{
		Input: PemBundle{Data: loadFixture(t, "crl.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("CRLParse 失败：%v", err)
	}
	data := res.Data.(*CRLParseResult)
	if len(data.Revoked) != 1 {
		t.Fatalf("吊销条目数 = %d，期望 1", len(data.Revoked))
	}
	if data.Revoked[0].ReasonCode != 1 {
		t.Errorf("原因码 = %d，期望 1", data.Revoked[0].ReasonCode)
	}
	if data.Revoked[0].ReasonText != "keyCompromise" {
		t.Errorf("原因名 = %q，期望 keyCompromise", data.Revoked[0].ReasonText)
	}
	if !strings.Contains(data.IssuerText, "Root CA") {
		t.Errorf("签发者 = %q，期望包含 Root CA", data.IssuerText)
	}
}

func TestCRLParseInvalid(t *testing.T) {
	_, err := CRLParse(context.Background(), &CRLParseRequest{
		Input: PemBundle{Data: "not a crl", Format: "pem"},
	})
	if got := codeOf(t, err); got != toolerr.CodeInvalidCRL {
		t.Errorf("错误码 = %s，期望 %s", got, toolerr.CodeInvalidCRL)
	}
}

// minInt 返回两数中的较小者。
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
