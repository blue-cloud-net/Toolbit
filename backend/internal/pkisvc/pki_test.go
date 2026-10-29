package pkisvc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 夹具目录：测试用的证书链、CSR 与 CRL。
//
// 这些材料由 `openssl` **一次性生成**，仅用于本仓库的测试，
// **不是**任何真实系统的密钥或证书（见 AGENTS.md §5）。
const fixtureDir = "../../testdata/fixtures/pki"

// loadFixture 读取一份夹具文件。
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(fixtureDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取夹具失败：%v", err)
	}
	return data
}

// loadCertPEM 读取一张证书并返回句柄。
func loadCertPEM(t *testing.T, name string) *Certificate {
	t.Helper()
	cert, err := LoadCertificate(loadFixture(t, name), true)
	if err != nil {
		t.Fatalf("加载证书 %s 失败：%v", name, err)
	}
	t.Cleanup(func() { _ = cert.Close() })
	return cert
}

func TestDescribeCertificate(t *testing.T) {
	cert := loadCertPEM(t, "leaf.crt.pem")

	info, err := Describe(cert, zeroTime, nil)
	if err != nil {
		t.Fatalf("Describe 失败：%v", err)
	}

	if info.Version != 2 {
		t.Errorf("Version = %d，期望 2（v3）", info.Version)
	}
	if !strings.Contains(info.SubjectText, "leaf.example.com") {
		t.Errorf("SubjectText = %q，期望包含 leaf.example.com", info.SubjectText)
	}
	if !strings.Contains(info.IssuerText, "Intermediate CA") {
		t.Errorf("IssuerText = %q，期望包含 Intermediate CA", info.IssuerText)
	}
	if info.PublicKey.Algorithm != AlgRSA {
		t.Errorf("PublicKey.Algorithm = %q，期望 %q", info.PublicKey.Algorithm, AlgRSA)
	}
	if info.PublicKey.Bits != 2048 {
		t.Errorf("PublicKey.Bits = %d，期望 2048", info.PublicKey.Bits)
	}
	if info.IsCA {
		t.Error("IsCA = true，期望 false（叶证书）")
	}
	if info.SelfSigned {
		t.Error("SelfSigned = true，期望 false（叶证书由中间 CA 签发）")
	}
	if len(info.SAN) == 0 {
		t.Error("SAN 为空，期望包含 DNS 条目")
	}
	if len(info.Subject) == 0 {
		t.Error("Subject 为空，期望包含 RDN 条目")
	}
	if info.Serial == "" {
		t.Error("Serial 为空")
	}
	if len(info.Extensions) == 0 {
		t.Error("Extensions 为空")
	}
	if info.SubjectKeyID == "" {
		t.Error("SubjectKeyID 为空，期望有 SKID 扩展")
	}
	if info.AuthorityKeyID == "" {
		t.Error("AuthorityKeyID 为空，期望有 AKID 扩展")
	}
}

func TestFingerprintsAllAlgorithms(t *testing.T) {
	cert := loadCertPEM(t, "leaf.crt.pem")

	algs := FingerprintAlgorithms()
	fps, err := Fingerprints(cert, algs)
	if err != nil {
		t.Fatalf("Fingerprints 失败：%v", err)
	}
	if len(fps) != len(algs) {
		t.Fatalf("指纹数量 = %d，期望 %d", len(fps), len(algs))
	}
	for i, fp := range fps {
		if fp.Algorithm != algs[i] {
			t.Errorf("第 %d 个算法 = %q，期望 %q（顺序须与请求一致）", i, fp.Algorithm, algs[i])
		}
		if fp.Value == "" {
			t.Errorf("算法 %q 的指纹为空", fp.Algorithm)
		}
		if fp.Value != strings.ToLower(fp.Value) {
			t.Errorf("算法 %q 的指纹 %q 不是小写", fp.Algorithm, fp.Value)
		}
	}
}

func TestFingerprintsUnsupportedAlgorithm(t *testing.T) {
	cert := loadCertPEM(t, "leaf.crt.pem")

	_, err := Fingerprints(cert, []string{"sha999"})
	if err == nil {
		t.Fatal("未知算法应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeUnsupportedAlgorithm {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeUnsupportedAlgorithm)
	}
}

func TestCheckValidityExpired(t *testing.T) {
	cert := loadCertPEM(t, "expired.crt.pem")

	info := CheckValidity(cert, zeroTime)
	if info.Status != ValidityExpired {
		t.Errorf("Status = %q，期望 %q", info.Status, ValidityExpired)
	}
}

func TestCheckValidityNotYetValid(t *testing.T) {
	cert := loadCertPEM(t, "leaf.crt.pem")

	// 取一个早于 notBefore 的时刻。
	before := cert.cert.NotBefore().Add(-time.Hour)
	info := CheckValidity(cert, before)
	if info.Status != ValidityNotYetValid {
		t.Errorf("Status = %q，期望 %q", info.Status, ValidityNotYetValid)
	}
}

func TestCheckValidityValid(t *testing.T) {
	cert := loadCertPEM(t, "leaf.crt.pem")

	info := CheckValidity(cert, zeroTime)
	if info.Status != ValidityValid {
		t.Errorf("Status = %q，期望 %q", info.Status, ValidityValid)
	}
	if info.DaysRemaining <= 0 {
		t.Errorf("DaysRemaining = %d，期望 > 0", info.DaysRemaining)
	}
}

func TestParseCSR(t *testing.T) {
	csr := loadFixture(t, "leaf.csr.pem")

	info, err := ParseCSR(csr, true)
	if err != nil {
		t.Fatalf("ParseCSR 失败：%v", err)
	}
	if !strings.Contains(info.SubjectText, "leaf.example.com") {
		t.Errorf("SubjectText = %q，期望包含 leaf.example.com", info.SubjectText)
	}
	if info.PublicKey.Algorithm != AlgRSA {
		t.Errorf("PublicKey.Algorithm = %q，期望 %q", info.PublicKey.Algorithm, AlgRSA)
	}
	if !info.SignatureValid {
		t.Error("SignatureValid = false，期望 true")
	}
	if info.FingerprintSHA256 == "" {
		t.Error("FingerprintSHA256 为空")
	}
	if len(info.FingerprintSHA256) != 64 {
		t.Errorf("FingerprintSHA256 长度 = %d，期望 64", len(info.FingerprintSHA256))
	}
}

func TestParseCSRInvalid(t *testing.T) {
	_, err := ParseCSR([]byte("not a csr"), true)
	if err == nil {
		t.Fatal("非法 CSR 应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidCSR {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeInvalidCSR)
	}
}

func TestSplitPEMBundle(t *testing.T) {
	bundle := loadFixture(t, "chain.pem")

	blocks, err := SplitPEM(bundle)
	if err != nil {
		t.Fatalf("SplitPEM 失败：%v", err)
	}
	if len(blocks) != 3 {
		t.Fatalf("块数 = %d，期望 3", len(blocks))
	}
	for _, block := range blocks {
		if block.Label != normalizePEMLabel(PEMBlockCertificate) {
			t.Errorf("块标签 = %q，期望 CERTIFICATE", block.Label)
		}
	}
}

func TestSplitPEMNotPEM(t *testing.T) {
	_, err := SplitPEM([]byte("hello"))
	if err == nil {
		t.Fatal("非 PEM 输入应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidPEM {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeInvalidPEM)
	}
}

func TestConvertPEMDER(t *testing.T) {
	cert := loadFixture(t, "leaf.crt.pem")

	// PEM → DER
	result, err := ConvertPEMDER(PEMConvertOptions{
		Source:      cert,
		SourceIsPEM: true,
		Type:        PEMTypeCert,
		To:          ConvertToDER,
	})
	if err != nil {
		t.Fatalf("PEM → DER 失败：%v", err)
	}
	if result.Encoding != "base64" {
		t.Errorf("Encoding = %q，期望 base64", result.Encoding)
	}
	if result.Labels[0] != PEMBlockCertificate {
		t.Errorf("Labels[0] = %q，期望 %q", result.Labels[0], PEMBlockCertificate)
	}

	// DER → PEM
	der := loadFixture(t, "leaf.crt.der")
	back, err := ConvertPEMDER(PEMConvertOptions{
		Source:      der,
		SourceIsPEM: false,
		Type:        PEMTypeCert,
		To:          ConvertToPEM,
	})
	if err != nil {
		t.Fatalf("DER → PEM 失败：%v", err)
	}
	if !strings.Contains(back.Data, "BEGIN CERTIFICATE") {
		t.Errorf("输出不含 CERTIFICATE 块：%q", back.Data[:minInt(len(back.Data), 80)])
	}
}

func TestConvertPEMDERMultipleBlocksToDERRejected(t *testing.T) {
	bundle := loadFixture(t, "chain.pem")

	_, err := ConvertPEMDER(PEMConvertOptions{
		Source:      bundle,
		SourceIsPEM: true,
		Type:        PEMTypeCert,
		To:          ConvertToDER,
	})
	if err == nil {
		t.Fatal("多块转 DER 应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeUnsupportedFormat)
	}
}

func TestConvertPEMDERTypeFilterMismatch(t *testing.T) {
	cert := loadFixture(t, "leaf.crt.pem")

	_, err := ConvertPEMDER(PEMConvertOptions{
		Source:      cert,
		SourceIsPEM: true,
		Type:        PEMTypeCRL,
		To:          ConvertToPEM,
	})
	if err == nil {
		t.Fatal("类型不匹配应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeUnsupportedFormat {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeUnsupportedFormat)
	}
}

func TestParseASN1(t *testing.T) {
	der := loadFixture(t, "leaf.crt.der")

	result, err := ParseASN1(der, 0)
	if err != nil {
		t.Fatalf("ParseASN1 失败：%v", err)
	}
	if result.Length != len(der) {
		t.Errorf("Length = %d，期望 %d", result.Length, len(der))
	}
	if result.NodeCount < 3 {
		t.Errorf("NodeCount = %d，期望 >= 3", result.NodeCount)
	}
	if result.MaxDepth != ASN1DefaultDepth {
		t.Errorf("MaxDepth = %d，期望 %d", result.MaxDepth, ASN1DefaultDepth)
	}
	if !strings.Contains(result.HexDump, "00000000:") {
		t.Errorf("HexDump 首行格式异常：%q", result.HexDump[:minInt(len(result.HexDump), 40)])
	}
	if result.Tree.Number != 16 {
		t.Errorf("根节点 Number = %d，期望 16（SEQUENCE）", result.Tree.Number)
	}
}

func TestParseASN1DepthLimit(t *testing.T) {
	der := loadFixture(t, "leaf.crt.der")

	result, err := ParseASN1(der, 2)
	if err != nil {
		t.Fatalf("ParseASN1 失败：%v", err)
	}
	if result.MaxDepth != 2 {
		t.Errorf("MaxDepth = %d，期望 2", result.MaxDepth)
	}
	// 深度 2 时只展开根与它的一层子节点；子节点应标记为已截断。
	if len(result.Tree.Children) == 0 {
		t.Fatal("根节点没有子节点")
	}
	if !result.Tree.Children[0].Truncated {
		t.Error("深度受限时第二层子节点应标记 Truncated")
	}
}

func TestParseASN1Invalid(t *testing.T) {
	_, err := ParseASN1([]byte{0x30, 0x05, 0x01}, 0)
	if err == nil {
		t.Fatal("截断的 DER 应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidDER {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeInvalidDER)
	}
}

func TestOrderChain(t *testing.T) {
	certs, err := ParseCertificates(loadFixture(t, "chain.pem"), true)
	if err != nil {
		t.Fatalf("ParseCertificates 失败：%v", err)
	}
	t.Cleanup(func() { closeCertificates(certs) })

	info, err := OrderChain(certs)
	if err != nil {
		t.Fatalf("OrderChain 失败：%v", err)
	}
	if info.Count != 3 {
		t.Errorf("Count = %d，期望 3", info.Count)
	}
	if len(info.Chain) != 3 {
		t.Fatalf("Chain 长度 = %d，期望 3", len(info.Chain))
	}
	if !strings.Contains(info.Chain[0].Info.SubjectText, "leaf.example.com") {
		t.Errorf("链首应为叶证书，实际 = %q", info.Chain[0].Info.SubjectText)
	}
	if !strings.Contains(info.Chain[2].Info.SubjectText, "Root CA") {
		t.Errorf("链尾应为根证书，实际 = %q", info.Chain[2].Info.SubjectText)
	}
	if !info.SelfSigned {
		t.Error("SelfSigned = false，期望 true（链尾是自签根）")
	}
	if info.MissingRoot {
		t.Error("MissingRoot = true，期望 false")
	}
}

func TestOrderChainReversedInput(t *testing.T) {
	certs, err := ParseCertificates(loadFixture(t, "chain.pem"), true)
	if err != nil {
		t.Fatalf("ParseCertificates 失败：%v", err)
	}
	t.Cleanup(func() { closeCertificates(certs) })

	// 反转输入：排序后应恢复叶 → 根。
	reversed := []*Certificate{certs[2], certs[1], certs[0]}
	info, err := OrderChain(reversed)
	if err != nil {
		t.Fatalf("OrderChain 失败：%v", err)
	}
	if !strings.Contains(info.Chain[0].Info.SubjectText, "leaf.example.com") {
		t.Errorf("链首应为叶证书，实际 = %q", info.Chain[0].Info.SubjectText)
	}
	if info.Ordered {
		t.Error("Ordered = true，期望 false（输入是反的）")
	}
}

func TestOrderChainTooMany(t *testing.T) {
	certs := make([]*Certificate, MaxChainLength+1)
	for i := range certs {
		certs[i] = &Certificate{}
	}

	_, err := OrderChain(certs)
	if err == nil {
		t.Fatal("超过链长度上限应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeTooManyItems {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeTooManyItems)
	}
}

func TestVerifyChainWithTrustedRoot(t *testing.T) {
	leaf := loadCertPEM(t, "leaf.crt.pem")
	inter := loadCertPEM(t, "inter.crt.pem")
	root := loadCertPEM(t, "ca.crt.pem")

	result, err := VerifyChain(leaf, []*Certificate{inter}, []*Certificate{root}, zeroTime, "leaf.example.com")
	if err != nil {
		t.Fatalf("VerifyChain 失败：%v", err)
	}
	if !result.Valid {
		t.Errorf("Valid = false，期望 true；errors=%v", result.Errors)
	}
	if result.ImplicitTrust {
		t.Error("ImplicitTrust = true，期望 false（显式给了 trustedRoots）")
	}
	if !result.HostnameMatch {
		t.Error("HostnameMatch = false，期望 true")
	}
}

func TestVerifyChainHostnameMismatch(t *testing.T) {
	leaf := loadCertPEM(t, "leaf.crt.pem")
	inter := loadCertPEM(t, "inter.crt.pem")
	root := loadCertPEM(t, "ca.crt.pem")

	result, err := VerifyChain(leaf, []*Certificate{inter}, []*Certificate{root}, zeroTime, "other.example.com")
	if err != nil {
		t.Fatalf("VerifyChain 失败：%v", err)
	}
	if !result.Valid {
		t.Errorf("Valid = false，期望 true（域名不匹配不影响链可信）")
	}
	if result.HostnameMatch {
		t.Error("HostnameMatch = true，期望 false")
	}
	if result.HostnameError == "" {
		t.Error("HostnameError 为空，期望有说明")
	}
}

func TestVerifyChainImplicitTrust(t *testing.T) {
	leaf := loadCertPEM(t, "leaf.crt.pem")
	inter := loadCertPEM(t, "inter.crt.pem")

	// 没有 trustedRoots：链内没有自签名根（根不在输入里）→ 验证失败。
	result, err := VerifyChain(leaf, []*Certificate{inter}, nil, zeroTime, "")
	if err != nil {
		t.Fatalf("VerifyChain 失败：%v", err)
	}
	if !result.ImplicitTrust {
		t.Error("ImplicitTrust = false，期望 true")
	}
	if result.Valid {
		t.Error("Valid = true，期望 false（找不到信任锚）")
	}
	if len(result.Errors) == 0 {
		t.Error("Errors 为空，期望有失败原因")
	}
}

func TestVerifyChainSetTimeIsolation(t *testing.T) {
	leaf := loadCertPEM(t, "leaf.crt.pem")
	inter := loadCertPEM(t, "inter.crt.pem")
	root := loadCertPEM(t, "ca.crt.pem")

	// 过期证书：由中间 CA 签发，按当前时刻应失败。
	expired := loadCertPEM(t, "expired.crt.pem")
	result, err := VerifyChain(expired, []*Certificate{inter}, []*Certificate{root}, zeroTime, "")
	if err != nil {
		t.Fatalf("VerifyChain 失败：%v", err)
	}
	if result.Valid {
		t.Error("Valid = true，期望 false（证书已过期）")
	}
	if len(result.Errors) == 0 {
		t.Fatal("Errors 为空，期望有过期原因")
	}
	if result.Errors[0].Code != "X509_V_ERR_CERT_HAS_EXPIRED" {
		t.Errorf("错误码 = %q，期望 X509_V_ERR_CERT_HAS_EXPIRED", result.Errors[0].Code)
	}

	// 回拨到证书有效期内：应通过。
	at := expired.cert.NotBefore().Add(time.Hour)
	result, err = VerifyChain(expired, []*Certificate{inter}, []*Certificate{root}, at, "")
	if err != nil {
		t.Fatalf("VerifyChain 失败：%v", err)
	}
	if !result.Valid {
		t.Errorf("Valid = false，期望 true（回拨到有效期内）；errors=%v", result.Errors)
	}

	// 后续调用**不**受前一次 SetTime 影响。
	result, err = VerifyChain(leaf, []*Certificate{inter}, []*Certificate{root}, zeroTime, "")
	if err != nil {
		t.Fatalf("VerifyChain 失败：%v", err)
	}
	if !result.Valid {
		t.Errorf("Valid = false，期望 true（SetTime 不得跨请求污染）")
	}
}

func TestRepairChain(t *testing.T) {
	leaf := loadCertPEM(t, "leaf.crt.pem")
	inter := loadCertPEM(t, "inter.crt.pem")
	root := loadCertPEM(t, "ca.crt.pem")

	result, err := RepairChain(leaf, []*Certificate{inter}, []*Certificate{root})
	if err != nil {
		t.Fatalf("RepairChain 失败：%v", err)
	}
	if result.AddedCount != 2 {
		t.Errorf("AddedCount = %d，期望 2", result.AddedCount)
	}
	if len(result.ChainPEM) != 3 {
		t.Errorf("ChainPEM 长度 = %d，期望 3", len(result.ChainPEM))
	}
	if !result.ReachedRoot {
		t.Error("ReachedRoot = false，期望 true")
	}
	if result.AddedFroms[0] != "intermediates" {
		t.Errorf("AddedFroms[0] = %q，期望 intermediates", result.AddedFroms[0])
	}
	if result.AddedFroms[1] != "trustedRoots" {
		t.Errorf("AddedFroms[1] = %q，期望 trustedRoots", result.AddedFroms[1])
	}
}

func TestRepairChainNoCandidates(t *testing.T) {
	leaf := loadCertPEM(t, "leaf.crt.pem")

	result, err := RepairChain(leaf, nil, nil)
	if err != nil {
		t.Fatalf("RepairChain 失败：%v", err)
	}
	if result.AddedCount != 0 {
		t.Errorf("AddedCount = %d，期望 0", result.AddedCount)
	}
	if result.ReachedRoot {
		t.Error("ReachedRoot = true，期望 false")
	}
	if result.Note != RepairAlreadyComplete {
		t.Errorf("Note = %q，期望 %q", result.Note, RepairAlreadyComplete)
	}
}

func TestParseCRLInfo(t *testing.T) {
	crl := loadFixture(t, "crl.pem")

	info, err := ParseCRLInfo(crl, true)
	if err != nil {
		t.Fatalf("ParseCRLInfo 失败：%v", err)
	}
	if !strings.Contains(info.IssuerText, "Root CA") {
		t.Errorf("IssuerText = %q，期望包含 Root CA", info.IssuerText)
	}
	if len(info.Revoked) != 1 {
		t.Fatalf("Revoked 长度 = %d，期望 1", len(info.Revoked))
	}
	if info.Revoked[0].ReasonCode != 1 {
		t.Errorf("ReasonCode = %d，期望 1（keyCompromise）", info.Revoked[0].ReasonCode)
	}
	// 夹具里的序列号是 20 字节（>64 位），上游用 int64 表达，
	// 因此读不出来 —— 此时序列号**为空**（见 I-18）。
	if info.Revoked[0].Serial != "" {
		t.Errorf("Serial = %q，期望空串（序列号超出 int64 范围）", info.Revoked[0].Serial)
	}
	if info.ThisUpdate.IsZero() {
		t.Error("ThisUpdate 为零值")
	}
}

func TestParseCRLInfoDER(t *testing.T) {
	crl := loadFixture(t, "crl.der")

	info, err := ParseCRLInfo(crl, false)
	if err != nil {
		t.Fatalf("ParseCRLInfo 失败：%v", err)
	}
	if len(info.Revoked) != 1 {
		t.Errorf("Revoked 长度 = %d，期望 1", len(info.Revoked))
	}
}

func TestParseCRLInfoInvalid(t *testing.T) {
	_, err := ParseCRLInfo([]byte("not a crl"), true)
	if err == nil {
		t.Fatal("非法 CRL 应报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidCRL {
		t.Errorf("错误码 = %q，期望 %q", toolErr.Code, toolerr.CodeInvalidCRL)
	}
}

func TestCertificateCloseIdempotent(t *testing.T) {
	cert, err := LoadCertificate(loadFixture(t, "leaf.crt.pem"), true)
	if err != nil {
		t.Fatalf("加载证书失败：%v", err)
	}
	if err := cert.Close(); err != nil {
		t.Errorf("首次 Close 失败：%v", err)
	}
	if err := cert.Close(); err != nil {
		t.Errorf("重复 Close 失败：%v", err)
	}
}

// minInt 返回两数中的较小者。
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
