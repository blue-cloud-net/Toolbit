package pkisvc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// 本文件覆盖 v0.6.0 的「写」侧能力：CSR 生成、证书签发、CRL 签发、
// 吊销判定、OCSP、PKCS#12、PKCS#7。
//
// 全部材料在测试内临时生成，**不**依赖任何真实密钥或证书。

// newTestKey 生成一把测试用的密钥。
func newTestKey(t *testing.T, algorithm string, bits int, curve string) *PrivateKey {
	t.Helper()
	pair, err := Generate(algorithm, bits, curve)
	if err != nil {
		t.Fatalf("生成 %s 密钥失败：%v", algorithm, err)
	}
	t.Cleanup(func() { _ = pair.Close() })
	return pair.Private
}

// newTestPair 生成一把测试用的密钥对（保留公钥句柄）。
func newTestPair(t *testing.T, algorithm string, bits int, curve string) *KeyPair {
	t.Helper()
	pair, err := Generate(algorithm, bits, curve)
	if err != nil {
		t.Fatalf("生成 %s 密钥失败：%v", algorithm, err)
	}
	t.Cleanup(func() { _ = pair.Close() })
	return pair
}

// testSubject 返回一组测试用的 DN 条目。
func testSubject(cn string) []NameEntryInfo {
	return []NameEntryInfo{
		{Field: "C", Value: "CN"},
		{Field: "O", Value: "Toolbit Test"},
		{Field: "CN", Value: cn},
	}
}

// issueTestCA 签发一张测试用 CA 证书。
func issueTestCA(t *testing.T, pair *KeyPair, cn string) *Certificate {
	t.Helper()
	serial, err := RandomSerial()
	if err != nil {
		t.Fatalf("生成序列号失败：%v", err)
	}
	now := time.Now()
	cert, err := IssueCertificate(CertIssueOptions{
		Subject:          testSubject(cn),
		Serial:           serial,
		NotBefore:        now.Add(-time.Hour),
		NotAfter:         now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:         []string{"critical", "keyCertSign", "cRLSign"},
		BasicConstraints: &BasicConstraintsOption{IsCA: true, PathLen: 1},
		AutoKeyID:        true,
	}, pair.Public, pair.Private, nil)
	if err != nil {
		t.Fatalf("签发 CA 证书失败：%v", err)
	}
	t.Cleanup(func() { _ = cert.Close() })
	return cert
}

// issueTestLeaf 用 CA 签发一张测试用叶证书。
func issueTestLeaf(t *testing.T, leaf *KeyPair, ca *Certificate, caKey *PrivateKey,
	cn string) *Certificate {
	t.Helper()
	serial, err := RandomSerial()
	if err != nil {
		t.Fatalf("生成序列号失败：%v", err)
	}
	now := time.Now()
	cert, err := IssueCertificate(CertIssueOptions{
		Subject:          testSubject(cn),
		Serial:           serial,
		NotBefore:        now.Add(-time.Hour),
		NotAfter:         now.Add(365 * 24 * time.Hour),
		SAN:              []string{"DNS:" + cn},
		KeyUsage:         []string{"critical", "digitalSignature", "keyEncipherment"},
		ExtendedKeyUsage: []string{"serverAuth"},
		BasicConstraints: &BasicConstraintsOption{IsCA: false, PathLen: -1},
		AutoKeyID:        true,
	}, leaf.Public, caKey, ca)
	if err != nil {
		t.Fatalf("签发叶证书失败：%v", err)
	}
	t.Cleanup(func() { _ = cert.Close() })
	return cert
}

func TestCreateCSRAndParse(t *testing.T) {
	key := newTestKey(t, AlgRSA, 2048, "")

	result, err := CreateCSR(CSRGenOptions{
		Subject:           testSubject("csr.example.com"),
		SAN:               []string{"DNS:csr.example.com", "DNS:www.csr.example.com"},
		KeyUsage:          []string{"critical", "digitalSignature"},
		BasicConstraints:  &BasicConstraintsOption{IsCA: false, PathLen: -1},
		ChallengePassword: "test-only",
	}, key)
	if err != nil {
		t.Fatalf("生成 CSR 失败：%v", err)
	}

	if !strings.Contains(string(result.PEM), "CERTIFICATE REQUEST") {
		t.Fatalf("CSR PEM 标签不正确：\n%s", result.PEM)
	}
	if result.Info == nil || !result.Info.SignatureValid {
		t.Fatal("自签名校验应通过")
	}
	if !result.Info.ChallengePasswordPresent {
		t.Error("ChallengePasswordPresent 应为 true")
	}
	if len(result.Info.SAN) != 2 {
		t.Errorf("SAN 条目数 = %d，期望 2", len(result.Info.SAN))
	}

	// 重新解析应得到同样结论（PEM 可被独立读取）。
	parsed, err := ParseCSR(result.PEM, true)
	if err != nil {
		t.Fatalf("解析刚生成的 CSR 失败：%v", err)
	}
	if !parsed.SignatureValid {
		t.Error("重新解析后 signatureValid 应为 true")
	}
	if parsed.FingerprintSHA256 != result.Info.FingerprintSHA256 {
		t.Error("两次解析的指纹应一致")
	}
}

func TestCreateCSRRejectsInvalidDN(t *testing.T) {
	key := newTestKey(t, AlgEC, 0, CurveP256)

	_, err := CreateCSR(CSRGenOptions{
		Subject: []NameEntryInfo{{Field: "NotAField", Value: "x"}},
	}, key)
	if err == nil {
		t.Fatal("非法 DN 字段名应报错")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeInvalidField {
		t.Fatalf("期望 INVALID_FIELD，实际 %v", err)
	}
}

func TestIssueCertificateSelfSigned(t *testing.T) {
	pair := newTestPair(t, AlgSM2, 0, "")
	serial, err := RandomSerial()
	if err != nil {
		t.Fatalf("生成序列号失败：%v", err)
	}
	now := time.Now()

	cert, err := IssueCertificate(CertIssueOptions{
		Subject:          testSubject("self.example.com"),
		Serial:           serial,
		NotBefore:        now.Add(-time.Minute),
		NotAfter:         now.Add(24 * time.Hour),
		BasicConstraints: &BasicConstraintsOption{IsCA: true, PathLen: -1},
		AutoKeyID:        true,
	}, pair.Public, pair.Private, nil)
	if err != nil {
		t.Fatalf("自签证书失败：%v", err)
	}
	defer func() { _ = cert.Close() }()

	info, err := Describe(cert, zeroTime, nil)
	if err != nil {
		t.Fatalf("Describe 失败：%v", err)
	}
	if !info.SelfSigned {
		t.Error("SelfSigned 应为 true")
	}
	if info.SubjectKeyID == "" || info.AuthorityKeyID == "" {
		t.Errorf("SKID / AKID 应自动补齐：skid=%q akid=%q", info.SubjectKeyID, info.AuthorityKeyID)
	}
	if info.SubjectKeyID != info.AuthorityKeyID {
		t.Error("自签场景 SKID 与 AKID 应相同")
	}
	if info.Serial == "" {
		t.Error("序列号不应为空")
	}
}

func TestIssueCertificateByCA(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "Test Root CA")

	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "leaf.toolbit.test")

	info, err := Describe(leaf, zeroTime, nil)
	if err != nil {
		t.Fatalf("Describe 失败：%v", err)
	}
	if info.SelfSigned {
		t.Error("CA 签发的证书不应判为自签")
	}
	if !strings.Contains(info.IssuerText, "Test Root CA") {
		t.Errorf("IssuerText = %q，期望包含 Test Root CA", info.IssuerText)
	}

	caInfo, err := Describe(ca, zeroTime, nil)
	if err != nil {
		t.Fatalf("Describe CA 失败：%v", err)
	}
	if info.AuthorityKeyID != caInfo.SubjectKeyID {
		t.Errorf("叶证书 AKID (%q) 应等于 CA 的 SKID (%q)",
			info.AuthorityKeyID, caInfo.SubjectKeyID)
	}
	if len(info.SAN) == 0 || info.SAN[0] != "DNS:leaf.toolbit.test" {
		t.Errorf("SAN = %v，期望包含 DNS:leaf.toolbit.test", info.SAN)
	}
}

func TestIssueCertificateKeyMismatch(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "Mismatch CA")
	other := newTestPair(t, AlgEC, 0, CurveP256)

	serial, err := RandomSerial()
	if err != nil {
		t.Fatalf("生成序列号失败：%v", err)
	}
	now := time.Now()

	// 自签模式：公钥与签名私钥不配对。
	_, err = IssueCertificate(CertIssueOptions{
		Subject:   testSubject("bad.example.com"),
		Serial:    serial,
		NotBefore: now.Add(-time.Minute),
		NotAfter:  now.Add(time.Hour),
	}, other.Public, caPair.Private, nil)
	if err == nil {
		t.Fatal("自签模式下公钥与签名私钥不配对应报错")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeKeyMismatch {
		t.Fatalf("期望 KEY_MISMATCH，实际 %v", err)
	}

	// CA 模式：签名私钥与签发者证书不配对。
	_, err = IssueCertificate(CertIssueOptions{
		Subject:   testSubject("bad.example.com"),
		Serial:    serial,
		NotBefore: now.Add(-time.Minute),
		NotAfter:  now.Add(time.Hour),
	}, other.Public, other.Private, ca)
	if err == nil {
		t.Fatal("CA 模式下签名私钥与签发者证书不配对应报错")
	}
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeKeyMismatch {
		t.Fatalf("期望 KEY_MISMATCH，实际 %v", err)
	}
}

func TestIssueCertificateInvalidValidity(t *testing.T) {
	pair := newTestPair(t, AlgEC, 0, CurveP256)
	now := time.Now()

	_, err := IssueCertificate(CertIssueOptions{
		Subject:   testSubject("bad.example.com"),
		Serial:    1,
		NotBefore: now,
		NotAfter:  now.Add(-time.Hour),
	}, pair.Public, pair.Private, nil)
	if err == nil {
		t.Fatal("notAfter 早于 notBefore 应报错")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeNotAfterBeforeNow {
		t.Fatalf("期望 NOT_AFTER_BEFORE_NOW，实际 %v", err)
	}
}

func TestCreateCRLAndCheckRevocation(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "CRL CA")

	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	revoked := issueTestLeaf(t, leafPair, ca, caPair.Private, "revoked.toolbit.test")

	goodPair := newTestPair(t, AlgEC, 0, CurveP256)
	good := issueTestLeaf(t, goodPair, ca, caPair.Private, "good.toolbit.test")

	now := time.Now()
	result, err := CreateCRL(ca, caPair.Private, CRLGenOptions{
		ThisUpdate: now,
		NextUpdate: now.Add(24 * time.Hour),
	}, []CRLRevokeItem{{
		Cert:   revoked,
		Reason: ReasonKeyCompromise,
	}})
	if err != nil {
		t.Fatalf("签发 CRL 失败：%v", err)
	}
	if result.Info == nil {
		t.Fatal("CRL 元信息不应为空")
	}
	if len(result.Info.Revoked) != 1 {
		t.Fatalf("吊销条目数 = %d，期望 1", len(result.Info.Revoked))
	}
	if result.Info.Revoked[0].ReasonText != "keyCompromise" {
		t.Errorf("原因名 = %q，期望 keyCompromise", result.Info.Revoked[0].ReasonText)
	}
	if result.Info.Number == 0 {
		t.Error("CRL Number 应自动生成")
	}

	crl, err := LoadRevocationList(result.PEM, true)
	if err != nil {
		t.Fatalf("加载刚签发的 CRL 失败：%v", err)
	}
	defer func() { _ = crl.Close() }()

	check, err := CheckRevocation(revoked, crl, ca, zeroTime)
	if err != nil {
		t.Fatalf("吊销判定失败：%v", err)
	}
	if check.Status != RevocationRevoked {
		t.Errorf("已吊销证书状态 = %q，期望 %q", check.Status, RevocationRevoked)
	}
	if !check.SignatureVerified {
		t.Error("给了 issuer 时应完成验签")
	}
	if check.ReasonCode != ReasonKeyCompromise {
		t.Errorf("原因码 = %d，期望 %d", check.ReasonCode, ReasonKeyCompromise)
	}

	checkGood, err := CheckRevocation(good, crl, ca, zeroTime)
	if err != nil {
		t.Fatalf("吊销判定失败：%v", err)
	}
	if checkGood.Status != RevocationGood {
		t.Errorf("未吊销证书状态 = %q，期望 %q", checkGood.Status, RevocationGood)
	}
}

func TestCheckRevocationWithoutIssuerIsUnverified(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "Unverified CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "u.toolbit.test")

	now := time.Now()
	result, err := CreateCRL(ca, caPair.Private, CRLGenOptions{
		ThisUpdate: now,
		NextUpdate: now.Add(time.Hour),
	}, nil)
	if err != nil {
		t.Fatalf("签发空 CRL 失败：%v", err)
	}

	crl, err := LoadRevocationList(result.DER, false)
	if err != nil {
		t.Fatalf("加载 CRL 失败：%v", err)
	}
	defer func() { _ = crl.Close() }()

	check, err := CheckRevocation(leaf, crl, nil, zeroTime)
	if err != nil {
		t.Fatalf("吊销判定失败：%v", err)
	}
	if check.Status == RevocationGood {
		t.Fatal("未验签时不得返回 good")
	}
	if check.Status != RevocationUnverified {
		t.Errorf("状态 = %q，期望 %q", check.Status, RevocationUnverified)
	}
	if check.SignatureVerified {
		t.Error("未给 issuer 时 SignatureVerified 应为 false")
	}
}

func TestCreateCRLRejectsReservedReason(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "Reason CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "r.toolbit.test")

	now := time.Now()
	_, err := CreateCRL(ca, caPair.Private, CRLGenOptions{
		ThisUpdate: now,
		NextUpdate: now.Add(time.Hour),
	}, []CRLRevokeItem{{Cert: leaf, Reason: 7}})
	if err == nil {
		t.Fatal("保留原因码 7 应被拒绝")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeInvalidField {
		t.Fatalf("期望 INVALID_FIELD，实际 %v", err)
	}
}

func TestCreateCRLRejectsStaleWindow(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "Window CA")
	now := time.Now()

	_, err := CreateCRL(ca, caPair.Private, CRLGenOptions{
		ThisUpdate: now,
		NextUpdate: now.Add(-time.Hour),
	}, nil)
	if err == nil {
		t.Fatal("nextUpdate 早于 thisUpdate 应报错")
	}
}

func TestCheckRevocationRejectsStaleCRL(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "Stale CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "s.toolbit.test")

	past := time.Now().Add(-48 * time.Hour)
	result, err := CreateCRL(ca, caPair.Private, CRLGenOptions{
		ThisUpdate: past,
		NextUpdate: past.Add(time.Hour),
	}, nil)
	if err != nil {
		t.Fatalf("签发 CRL 失败：%v", err)
	}

	crl, err := LoadRevocationList(result.PEM, true)
	if err != nil {
		t.Fatalf("加载 CRL 失败：%v", err)
	}
	defer func() { _ = crl.Close() }()

	_, err = CheckRevocation(leaf, crl, ca, zeroTime)
	if err == nil {
		t.Fatal("过期 CRL 不应作为吊销依据")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeInvalidCRL {
		t.Fatalf("期望 INVALID_CRL，实际 %v", err)
	}
}

func TestOCSPRequestAndResponseError(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "OCSP CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "ocsp.toolbit.test")

	for _, hash := range OCSPHashes() {
		result, err := CreateOCSPRequestDER(leaf, ca, hash)
		if err != nil {
			t.Fatalf("生成 OCSP 请求（%s）失败：%v", hash, err)
		}
		if len(result.DER) == 0 {
			t.Fatalf("OCSP 请求（%s）的 DER 为空", hash)
		}
		if result.Hash != hash {
			t.Errorf("Hash = %q，期望 %q", result.Hash, hash)
		}
	}

	_, err := CreateOCSPRequestDER(leaf, ca, "sha512")
	if err == nil {
		t.Fatal("不支持的 OCSP 哈希算法应报错")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeUnsupportedAlgorithm {
		t.Fatalf("期望 UNSUPPORTED_ALGORITHM，实际 %v", err)
	}

	// 用 OCSP 请求字节冒充响应：必须被拒绝，而不是解析出垃圾结论。
	req, err := CreateOCSPRequestDER(leaf, ca, OCSPHashSHA256)
	if err != nil {
		t.Fatalf("生成 OCSP 请求失败：%v", err)
	}
	if _, err := ParseOCSPResponseDER(req.DER, leaf, ca); err == nil {
		t.Fatal("把请求当响应解析应报错")
	}
}

func TestPKCS12RoundTrip(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "P12 CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "p12.toolbit.test")

	der, err := PackPKCS12(leaf, leafPair.Private, PKCS12PackOptions{
		Password:     "test-pass",
		FriendlyName: "toolbit-test",
		CACerts:      []*Certificate{ca},
	})
	if err != nil {
		t.Fatalf("打包 PKCS#12 失败：%v", err)
	}

	bundle, err := ParsePKCS12(der, "test-pass")
	if err != nil {
		t.Fatalf("解析 PKCS#12 失败：%v", err)
	}
	if bundle.Certificate == nil {
		t.Fatal("主证书不应为空")
	}
	if bundle.PrivateKey == nil {
		t.Fatal("私钥不应为空")
	}
	if len(bundle.CACerts) != 1 {
		t.Errorf("CA 链长度 = %d，期望 1", len(bundle.CACerts))
	}
	if err := bundle.Close(); err != nil {
		t.Errorf("关闭 Bundle 失败：%v", err)
	}
	// 幂等：重复关闭不应报错。
	if err := bundle.Close(); err != nil {
		t.Errorf("重复关闭 Bundle 失败：%v", err)
	}
}

func TestPKCS12WrongPassphrase(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "P12 Bad CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "p12b.toolbit.test")

	der, err := PackPKCS12(leaf, leafPair.Private, PKCS12PackOptions{Password: "right"})
	if err != nil {
		t.Fatalf("打包 PKCS#12 失败：%v", err)
	}

	_, err = ParsePKCS12(der, "wrong")
	if err == nil {
		t.Fatal("错误口令应报错")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodePassphraseIncorrect {
		t.Fatalf("期望 PASSPHRASE_INCORRECT，实际 %v", err)
	}
}

func TestPKCS12ChangePassword(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "P12 Chg CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "p12c.toolbit.test")

	der, err := PackPKCS12(leaf, leafPair.Private, PKCS12PackOptions{Password: "old-pass"})
	if err != nil {
		t.Fatalf("打包 PKCS#12 失败：%v", err)
	}

	updated, err := ChangePKCS12Password(der, "old-pass", "new-pass")
	if err != nil {
		t.Fatalf("改口令失败：%v", err)
	}

	if _, err := ParsePKCS12(updated, "old-pass"); err == nil {
		t.Fatal("旧口令应失效")
	}
	bundle, err := ParsePKCS12(updated, "new-pass")
	if err != nil {
		t.Fatalf("新口令应可用：%v", err)
	}
	if err := bundle.Close(); err != nil {
		t.Errorf("关闭 Bundle 失败：%v", err)
	}
}

func TestPKCS12KeyMismatch(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "P12 Mix CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "p12m.toolbit.test")
	other := newTestPair(t, AlgEC, 0, CurveP256)

	_, err := PackPKCS12(leaf, other.Private, PKCS12PackOptions{Password: "x"})
	if err == nil {
		t.Fatal("私钥与证书不配对时应报错")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeKeyMismatch {
		t.Fatalf("期望 KEY_MISMATCH，实际 %v", err)
	}
}

func TestPKCS7RoundTrip(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "P7 CA")
	leafPair := newTestPair(t, AlgEC, 0, CurveP256)
	leaf := issueTestLeaf(t, leafPair, ca, caPair.Private, "p7.toolbit.test")

	der, err := BuildPKCS7([]*Certificate{leaf, ca})
	if err != nil {
		t.Fatalf("打包 PKCS#7 失败：%v", err)
	}

	certs, err := ExtractPKCS7(der)
	if err != nil {
		t.Fatalf("提取 PKCS#7 失败：%v", err)
	}
	if len(certs) != 2 {
		t.Fatalf("提取证书数 = %d，期望 2", len(certs))
	}
	for _, cert := range certs {
		if err := cert.Close(); err != nil {
			t.Errorf("关闭证书失败：%v", err)
		}
	}

	if pem := EncodePKCS7PEM(der); !strings.Contains(string(pem), "PKCS7") {
		t.Errorf("PKCS#7 PEM 标签不正确：\n%s", pem)
	}
}

func TestExtractPKCS7RejectsGarbage(t *testing.T) {
	_, err := ExtractPKCS7([]byte{0x30, 0x03, 0x02, 0x01, 0x01})
	if err == nil {
		t.Fatal("非法 PKCS#7 应报错")
	}
	var toolErr *toolerr.Error
	if !asToolerr(err, &toolErr) || toolErr.Code != toolerr.CodeInvalidPKCS7 {
		t.Fatalf("期望 INVALID_PKCS7，实际 %v", err)
	}
}

func TestStoreAddAndClose(t *testing.T) {
	caPair := newTestPair(t, AlgEC, 0, CurveP256)
	ca := issueTestCA(t, caPair, "Store CA")

	store := NewStore()
	if err := store.AddCertificate(ca); err != nil {
		t.Fatalf("加入信任库失败：%v", err)
	}
	if err := store.Close(); err != nil {
		t.Errorf("关闭信任库失败：%v", err)
	}
	if err := store.Close(); err != nil {
		t.Errorf("重复关闭信任库失败：%v", err)
	}
}

// asToolerr 判断 err 是否为 toolerr.Error 并取出。
func asToolerr(err error, target **toolerr.Error) bool {
	for err != nil {
		if typed, ok := err.(*toolerr.Error); ok {
			*target = typed
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

// TestGenerateToolbitP12 生成**Toolbit 自己打包**的 PKCS#12 夹具。
//
// 只有设置 TOOLBIT_V060_FIXTURES 时才执行，平时 Skip。
// 需要它的原因：上游 `PKCS12_newpass` 无法对外部实现产出的容器改密
// （见 docs/upstream/tongsuo-go-issues.md I-21），
// 因此 pfx-change-password 的契约数据必须用本库打包的容器。
func TestGenerateToolbitP12(t *testing.T) {
	dir := os.Getenv("TOOLBIT_V060_FIXTURES")
	if dir == "" {
		t.Skip("未设置 TOOLBIT_V060_FIXTURES，跳过夹具生成")
	}

	load := func(name string) []byte {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("读取 %s 失败：%v", name, err)
		}
		return data
	}

	cert, err := LoadCertificate(load("leaf2.crt.pem"), true)
	if err != nil {
		t.Fatalf("加载叶证书失败：%v", err)
	}
	defer func() { _ = cert.Close() }()

	key, err := LoadPrivateKey(load("leaf2.key.pem"), "")
	if err != nil {
		t.Fatalf("加载私钥失败：%v", err)
	}
	defer func() { _ = key.Close() }()

	ca, err := LoadCertificate(load("ca2.crt.pem"), true)
	if err != nil {
		t.Fatalf("加载 CA 证书失败：%v", err)
	}
	defer func() { _ = ca.Close() }()

	der, err := PackPKCS12(cert, key, PKCS12PackOptions{
		Password:     "test-pass",
		FriendlyName: "toolbit-test",
		CACerts:      []*Certificate{ca},
	})
	if err != nil {
		t.Fatalf("打包 PKCS#12 失败：%v", err)
	}

	path := filepath.Join(dir, "bundle-toolbit.p12")
	if err := os.WriteFile(path, der, 0o600); err != nil {
		t.Fatalf("写入 %s 失败：%v", path, err)
	}
	t.Logf("已写入 %s（%d 字节）", path, len(der))
}
