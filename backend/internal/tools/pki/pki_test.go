package pki

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
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
	if !strings.HasPrefix(data.Pem, "-----BEGIN CERTIFICATE-----") {
		t.Errorf("Pem = %q，期望证书 PEM", data.Pem)
	}
	if data.PublicKey == nil {
		t.Fatal("PublicKey 为空")
	}
	if data.PublicKey.Algorithm != "RSA" {
		t.Errorf("公钥算法 = %q，期望 RSA", data.PublicKey.Algorithm)
	}
	if data.PublicKey.Modulus == "" {
		t.Error("公钥模数为空")
	}
	if data.TextView == "" {
		t.Error("TextView 为空")
	}
	if !strings.Contains(data.TextView, "Subject: /C=CN/O=Toolbit Test/CN=leaf.example.com") {
		t.Errorf("TextView 缺少主体行：%s", data.TextView)
	}
	sanFound := false
	for _, ext := range data.Certificate.Extensions {
		if ext.OID == "2.5.29.17" {
			sanFound = true
			if !strings.Contains(ext.Text, "DNS:leaf.example.com") {
				t.Errorf("SAN 扩展文本 = %q，期望包含 DNS:leaf.example.com", ext.Text)
			}
		}
	}
	if !sanFound {
		t.Error("扩展列表缺少 subjectAltName")
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

	roles := []string{"leaf", "intermediate", "root"}
	for i, want := range roles {
		if data.Chain[i].Role != want {
			t.Errorf("第 %d 张的角色 = %q，期望 %q", i, data.Chain[i].Role, want)
		}
	}
	for i, cert := range data.Chain {
		if !cert.Link.Ok {
			t.Errorf("第 %d 张的连接异常：%+v", i, cert.Link)
		}
		if cert.Link.Method != "keyid" && i < 2 {
			t.Errorf("第 %d 张的连接方式 = %q，期望 keyid", i, cert.Link.Method)
		}
		if len(cert.FingerprintSha256) != 64 {
			t.Errorf("第 %d 张的指纹长度 = %d，期望 64", i, len(cert.FingerprintSha256))
		}
		if cert.PublicKey == nil || cert.PublicKey.Modulus == "" {
			t.Errorf("第 %d 张缺少公钥细节", i)
		}
	}
	if data.Chain[2].Link.Method != "self" {
		t.Errorf("根证书的连接方式 = %q，期望 self", data.Chain[2].Link.Method)
	}
	if data.Chain[0].Validity.Status != "valid" {
		t.Errorf("叶证书有效性 = %q，期望 valid", data.Chain[0].Validity.Status)
	}
}

func TestChainParseReportsBrokenLink(t *testing.T) {
	// 只给叶证书：签发者不在输入里 → 缺中间 CA。
	res, err := ChainParse(context.Background(), &ChainParseRequest{
		Bundle: PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
	})
	if err != nil {
		t.Fatalf("ChainParse 失败：%v", err)
	}
	data := res.Data.(*ChainParseResult)
	if data.MissingRoot != true {
		t.Error("MissingRoot = false，期望 true")
	}
	if len(data.Chain) != 1 {
		t.Fatalf("链长度 = %d，期望 1", len(data.Chain))
	}
	if data.Chain[0].Link.Ok {
		t.Error("断链时 Link.Ok 应为 false")
	}
	if data.Chain[0].Link.Reason != "missing_intermediate" {
		t.Errorf("断链原因 = %q，期望 missing_intermediate", data.Chain[0].Link.Reason)
	}
}

// TestChainParseDetectsMissingIntermediate 验证「叶 + 根（缺中间 CA）」的断链诊断。
//
// 这个组合是自签名判定的陷阱：根证书排在排序结果末尾且自签名，
// 若拿它判定链尾就会误报「不缺根」，并把叶证书标成根。
func TestChainParseDetectsMissingIntermediate(t *testing.T) {
	leaf := loadFixture(t, "leaf.crt.pem")
	root := loadFixture(t, "ca.crt.pem")

	res, err := ChainParse(context.Background(), &ChainParseRequest{
		Bundle: PemBundle{Data: leaf + root, Format: "pem"},
	})
	if err != nil {
		t.Fatalf("ChainParse 失败：%v", err)
	}
	data := res.Data.(*ChainParseResult)
	if data.MissingRoot != true {
		t.Error("MissingRoot = false，期望 true（根证书接不进主链）")
	}
	if data.SelfSigned != false {
		t.Error("SelfSigned = true，期望 false（主链链尾是叶证书）")
	}
	if len(data.Chain) != 2 {
		t.Fatalf("链长度 = %d，期望 2", len(data.Chain))
	}
	if data.Chain[0].Role != "leaf" {
		t.Errorf("链首角色 = %q，期望 leaf", data.Chain[0].Role)
	}
	if data.Chain[0].Link.Ok {
		t.Error("链首 Link.Ok = true，期望 false")
	}
	if data.Chain[0].Link.Reason != "missing_intermediate" {
		t.Errorf("链首断链原因 = %q，期望 missing_intermediate", data.Chain[0].Link.Reason)
	}
	if data.Chain[1].Role != "unlinked" {
		t.Errorf("链尾角色 = %q，期望 unlinked", data.Chain[1].Role)
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

// ── v0.6.0：写侧工具的单元测试 ────────────────────────────────────
//
// 覆盖「正例 + 反例」各一，重点在字段校验与错误码映射；
// 密码学正确性由 pkisvc 的测试与本文件的端到端链路负责。

func TestCSRGenAndParseBack(t *testing.T) {
	result, err := CSRGen(context.Background(), &CSRGenRequest{
		CertType: CSRTypeServer,
		Subject:  SubjectInput{C: "CN", O: "Toolbit Test", CN: "unit.example.com"},
		Key:      KeyGenInput{Type: "EC", Curve: "P-256"},
		Sans:     []string{"DNS:unit.example.com"},
	})
	if err != nil {
		t.Fatalf("生成 CSR 失败：%v", err)
	}
	out := result.Data.(*CSRGenResult)
	if out.CsrPem == "" || out.PrivateKeyPem == "" {
		t.Fatal("CSR 与私钥都应返回")
	}
	if !out.Info.SignatureValid {
		t.Error("刚生成的 CSR 自签名应校验通过")
	}
	if out.KeyType != "EC" || out.KeyCurve != "P-256" {
		t.Errorf("密钥参数 = %s/%s，期望 EC/P-256", out.KeyType, out.KeyCurve)
	}

	// 用已有私钥生成时**不**返回私钥。
	existing := CSRGenRequest{
		Subject:            SubjectInput{CN: "unit2.example.com"},
		ExistingPrivateKey: &KeyInput{Data: out.PrivateKeyPem, Format: "pem"},
	}
	reused, err := CSRGen(context.Background(), &existing)
	if err != nil {
		t.Fatalf("用已有私钥生成 CSR 失败：%v", err)
	}
	if reused.Data.(*CSRGenResult).PrivateKeyPem != "" {
		t.Error("使用已有私钥时不应返回私钥")
	}
}

func TestCSRGenRejectsUnknownType(t *testing.T) {
	_, err := CSRGen(context.Background(), &CSRGenRequest{
		CertType: "unknown",
		Subject:  SubjectInput{CN: "x.example.com"},
	})
	if err == nil {
		t.Fatal("未知生成类型应报错")
	}
}

func TestCertGenSelfSigned(t *testing.T) {
	result, err := CertGen(context.Background(), &CertGenRequest{
		SignMode:         pkisvc.SignModeSelf,
		Subject:          SubjectInput{CN: "self.unit.example.com"},
		Key:              KeyGenInput{Type: "EC", Curve: "P-256"},
		NotAfterDays:     30,
		BasicConstraints: &BasicConstraintsInput{IsCA: true, PathLen: 1},
	})
	if err != nil {
		t.Fatalf("自签证书失败：%v", err)
	}
	out := result.Data.(*CertGenResult)
	if out.Serial == "" || out.PrivateKeyPem == "" {
		t.Fatal("序列号与私钥都应返回")
	}
	if len(out.ChainPem) != 1 {
		t.Errorf("自签模式的链长度 = %d，期望 1", len(out.ChainPem))
	}
}

func TestCertGenRejectsBadSignMode(t *testing.T) {
	_, err := CertGen(context.Background(), &CertGenRequest{SignMode: "peer"})
	if err == nil {
		t.Fatal("未知签发模式应报错")
	}
}

func TestCertGenRequiresIssuerForCA(t *testing.T) {
	_, err := CertGen(context.Background(), &CertGenRequest{
		SignMode: pkisvc.SignModeCA,
		Subject:  SubjectInput{CN: "x.example.com"},
	})
	if err == nil {
		t.Fatal("ca 模式缺少 issuer 应报错")
	}
}

func TestCRLGenRejectsReservedReason(t *testing.T) {
	_, err := CRLGen(context.Background(), &CRLGenRequest{
		IssuerCert: PemBundle{Data: loadFixture(t, "ca.crt.pem"), Format: "pem"},
		IssuerKey:  KeyInput{Data: "-----BEGIN PRIVATE KEY-----\nAAAA\n-----END PRIVATE KEY-----\n", Format: "pem"},
		Revoked: []CRLRevokeInput{{
			Cert:       PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
			ReasonCode: 7,
		}},
	})
	if err == nil {
		t.Fatal("保留原因码 7 应被拒绝")
	}
}

func TestRevocationCheckRejectsGarbage(t *testing.T) {
	_, err := RevocationCheck(context.Background(), &RevocationCheckRequest{
		Cert: PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		Crl:  PemBundle{Data: "not a crl", Format: "pem"},
	})
	if err == nil {
		t.Fatal("非法 CRL 应报错")
	}
}

func TestOCSPRequestRejectsUnsupportedHash(t *testing.T) {
	_, err := OCSPRequest(context.Background(), &OCSPRequestInput{
		Cert:   PemBundle{Data: loadFixture(t, "leaf.crt.pem"), Format: "pem"},
		Issuer: PemBundle{Data: loadFixture(t, "inter.crt.pem"), Format: "pem"},
		Hash:   "sha512",
	})
	if err == nil {
		t.Fatal("不支持的 OCSP 哈希算法应报错")
	}
}

func TestP7BPackRejectsEmpty(t *testing.T) {
	_, err := P7BPack(context.Background(), &P7BPackRequest{})
	if err == nil {
		t.Fatal("空证书列表应报错")
	}
}

func TestP7BPackAndExtract(t *testing.T) {
	leaf := loadFixture(t, "leaf.crt.pem")
	ca := loadFixture(t, "ca.crt.pem")

	packed, err := P7BPack(context.Background(), &P7BPackRequest{
		Certs: []PemBundle{
			{Data: leaf, Format: "pem"},
			{Data: ca, Format: "pem"},
		},
	})
	if err != nil {
		t.Fatalf("打包 PKCS#7 失败：%v", err)
	}
	out := packed.Data.(*P7BOutput)
	if out.CertCount != 2 {
		t.Errorf("证书数 = %d，期望 2", out.CertCount)
	}

	extracted, err := P7BExtract(context.Background(), &P7BExtractRequest{
		Container: PemBundle{Data: out.Pem, Format: "pem"},
	})
	if err != nil {
		t.Fatalf("提取 PKCS#7 失败：%v", err)
	}
	if extracted.Data.(*P7BExtractResult).Count != 2 {
		t.Errorf("提取证书数 = %d，期望 2", extracted.Data.(*P7BExtractResult).Count)
	}
}

func TestP7BExtractRejectsGarbage(t *testing.T) {
	_, err := P7BExtract(context.Background(), &P7BExtractRequest{
		Container: PemBundle{Data: "MAM=", Format: "base64"},
	})
	if err == nil {
		t.Fatal("非法 PKCS#7 应报错")
	}
}

func TestCAAGenProducesZoneLines(t *testing.T) {
	result, err := CAAGen(context.Background(), &CAAGenRequest{
		Domain: "example.com",
		Ttl:    3600,
		Records: []CAARecordInput{
			{Tag: CAATagIssue, Value: "letsencrypt.org"},
			{Tag: CAATagIodef, Value: "mailto:security@example.com"},
		},
	})
	if err != nil {
		t.Fatalf("生成 CAA 记录失败：%v", err)
	}
	out := result.Data.(*CAAGenResult)
	if out.Count != 2 {
		t.Errorf("记录数 = %d，期望 2", out.Count)
	}
	if !strings.Contains(out.ZoneLines[0], "CAA 0 issue \"letsencrypt.org\"") {
		t.Errorf("区域行格式不符：%s", out.ZoneLines[0])
	}
}

func TestCAAGenRejectsUnknownTag(t *testing.T) {
	_, err := CAAGen(context.Background(), &CAAGenRequest{
		Domain:  "example.com",
		Records: []CAARecordInput{{Tag: "unknown", Value: "x"}},
	})
	if err == nil {
		t.Fatal("未知 CAA 标签应报错")
	}
}
