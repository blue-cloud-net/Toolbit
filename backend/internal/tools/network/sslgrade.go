package network

import "strings"

// 四象限评级模型。
//
// 规则见 docs/tools/ssl-check-refactor.md §10：
//
//   - Certificate / Protocols / KeyExchange / CipherStrength 各 100 分，互不抵扣；
//   - 命中档位上限时取「分数映射等级」与「上限等级」中更严的一个；
//   - `unavailable` 的检查项不参与计分，也不触发上限，但必须记入 issues。

// gradeOrder 是等级顺序（越大越好）。
var gradeOrder = []string{"A+", "A", "A-", "B", "C", "D", "E", "F", "M", "T"}

// gradeRank 返回等级的下标；未知等级返回末尾。
func gradeRank(letter string) int {
	for i, item := range gradeOrder {
		if item == letter {
			return i
		}
	}
	return len(gradeOrder) - 1
}

// lowerGrade 返回两个等级中更严的一个。
func lowerGrade(a, b string) string {
	if gradeRank(a) >= gradeRank(b) {
		return a
	}
	return b
}

// gradeOfScore 把平均分映射成等级。
func gradeOfScore(score int) string {
	switch {
	case score >= 100:
		return "A+"
	case score >= 95:
		return "A"
	case score >= 90:
		return "A-"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	case score >= 50:
		return "E"
	default:
		return "F"
	}
}

// certificateScore 计算证书象限得分。
func certificateScore(cert ScanCertificate, issues *[]GradeIssue) int {
	score := 100
	if !cert.ChainValid || !cert.Verified {
		*issues = append(*issues, GradeIssue{Id: "certUntrusted", Impact: "critical"})
		return 0
	}
	if !cert.SanMatchesHost {
		*issues = append(*issues, GradeIssue{Id: "certHostnameMismatch", Impact: "critical"})
		return 0
	}
	if cert.DaysRemaining < 0 {
		*issues = append(*issues, GradeIssue{Id: "certExpired", Impact: "critical"})
		return 0
	}
	if cert.DaysRemaining < 30 {
		score -= 15
		*issues = append(*issues, GradeIssue{Id: "certExpiringSoon", Impact: "medium"})
	}
	if weakKey(cert) {
		score -= 15
		*issues = append(*issues, GradeIssue{Id: "certWeakKey", Impact: "high"})
	}
	upper := strings.ToUpper(cert.SignatureAlgorithm)
	if strings.Contains(upper, "SHA1") || strings.Contains(upper, "MD5") {
		score -= 10
		*issues = append(*issues, GradeIssue{Id: "certWeakSignature", Impact: "high"})
	}
	if cert.SelfSigned {
		*issues = append(*issues, GradeIssue{Id: "certSelfSigned", Impact: "unknown"})
	}
	return clampScore(score)
}

// weakKey 判断证书公钥强度是否不足。
func weakKey(cert ScanCertificate) bool {
	switch strings.ToUpper(cert.KeyAlgorithm) {
	case "RSA":
		return cert.KeyBits > 0 && cert.KeyBits < 2048
	default:
		return false
	}
}

// protocolsScore 计算协议象限得分。
func protocolsScore(protocols []ProtocolResult, issues *[]GradeIssue) int {
	best := ""
	for _, item := range protocols {
		if !item.Supported {
			continue
		}
		switch item.Version {
		case "TLS1.3":
			best = "TLS1.3"
		case "TLS1.2":
			if best != "TLS1.3" {
				best = "TLS1.2"
			}
		case "TLS1.1", "TLS1.0":
			if best == "" {
				best = item.Version
			}
		}
	}
	switch best {
	case "TLS1.3":
		return 100
	case "TLS1.2":
		*issues = append(*issues, GradeIssue{Id: "noTLS13", Impact: "low"})
		return 90
	case "TLS1.1", "TLS1.0":
		*issues = append(*issues, GradeIssue{Id: "onlyLegacyTLS", Impact: "high", Detail: best})
		return 20
	default:
		*issues = append(*issues, GradeIssue{Id: "noTLS", Impact: "critical"})
		return 0
	}
}

// keyExchangeScore 计算密钥交换象限得分。
func keyExchangeScore(negotiated ScanNegotiated, hsts HSTSResult, issues *[]GradeIssue) int {
	score := 100
	upper := strings.ToUpper(negotiated.Cipher)
	switch {
	case strings.HasPrefix(upper, "RSA_"), strings.HasPrefix(upper, "RSA-"):
		score = 20
		*issues = append(*issues, GradeIssue{Id: "rsaKeyExchange", Impact: "high"})
	case negotiated.KeyExchangeGroup == "":
		*issues = append(*issues, GradeIssue{Id: "kexGroupUnknown", Impact: "unknown"})
	}
	if hsts.Present && hsts.MaxAge > 0 {
		score += 5
	}
	return clampScore(score)
}

// cipherStrengthScore 计算加密强度象限得分。
func cipherStrengthScore(negotiated ScanNegotiated, ciphers []CipherResult, issues *[]GradeIssue) int {
	family, _, _ := describeSuite(negotiated.Cipher)
	switch family {
	case "AEAD":
		return 100
	case "CBC":
		*issues = append(*issues, GradeIssue{Id: "cbcCipher", Impact: "medium"})
		return 80
	case "RC4", "3DES":
		*issues = append(*issues, GradeIssue{Id: "legacyCipher", Impact: "high", Detail: family})
		return 20
	}

	// 协商套件不在已知族里时，回落到「是否存在弱套件被接受」的判断。
	for _, item := range ciphers {
		if item.Verdict != SuiteSupported {
			continue
		}
		if item.Family == "RC4" || item.Family == "3DES" {
			*issues = append(*issues, GradeIssue{Id: "legacyCipher", Impact: "high", Detail: item.Name})
			return 20
		}
	}
	return 90
}

// gradeCaps 计算档位上限。
func gradeCaps(result *ScanResult, issues *[]GradeIssue) []string {
	caps := make([]string, 0, 4)
	if !result.Certificate.ChainValid || !result.Certificate.Verified {
		caps = append(caps, "untrusted")
		*issues = append(*issues, GradeIssue{Id: "capUntrusted", Impact: "critical"})
	}
	if !result.Certificate.SanMatchesHost {
		caps = append(caps, "hostnameMismatch")
		*issues = append(*issues, GradeIssue{Id: "capHostname", Impact: "critical"})
	}
	if result.Certificate.DaysRemaining < 0 {
		caps = append(caps, "expired")
		*issues = append(*issues, GradeIssue{Id: "capExpired", Impact: "critical"})
	}
	if revoked(result) {
		caps = append(caps, "revoked")
		*issues = append(*issues, GradeIssue{Id: "capRevoked", Impact: "critical"})
	}

	hasModern := false
	hasTLS13 := false
	for _, item := range result.Protocols {
		if !item.Supported {
			continue
		}
		switch item.Version {
		case "TLS1.3":
			hasTLS13 = true
			hasModern = true
		case "TLS1.2":
			hasModern = true
		}
	}
	if len(result.Protocols) > 0 && !hasModern {
		caps = append(caps, "onlyLegacyTLS")
	}
	if len(result.Protocols) > 0 && !hasTLS13 {
		caps = append(caps, "noTLS13")
	}
	return caps
}

// revoked 报告吊销检查是否给出「已吊销」。
func revoked(result *ScanResult) bool {
	for _, entry := range result.Revocation.Ocsp {
		if entry.Status == "revoked" {
			return true
		}
	}
	for _, entry := range result.Revocation.Crl {
		if entry.Status == "revoked" {
			return true
		}
	}
	return false
}

// capGrade 把档位上限翻译成等级上限。
func capGrade(capability string) string {
	switch capability {
	case "untrusted", "expired":
		return "T"
	case "hostnameMismatch":
		return "M"
	case "revoked":
		return "F"
	case "onlyLegacyTLS":
		return "B"
	case "noTLS13":
		return "A"
	default:
		return "A+"
	}
}

// gradeScan 计算综合评级。
//
// checks 里未启用的检查项不参与计分，但要作为 `unavailable` 记入 issues，
// 否则用户会把「没测」误读成「没问题」。
func gradeScan(result *ScanResult, checks map[string]bool) GradeResult {
	issues := make([]GradeIssue, 0, 8)

	categories := GradeCategories{
		Certificate:    certificateScore(result.Certificate, &issues),
		Protocols:      protocolsScore(result.Protocols, &issues),
		KeyExchange:    keyExchangeScore(result.Negotiated, result.Hsts, &issues),
		CipherStrength: cipherStrengthScore(result.Negotiated, result.Ciphers, &issues),
	}

	for _, id := range AllChecks() {
		if !checks[id] {
			issues = append(issues, GradeIssue{Id: "skipped:" + id, Impact: "unknown"})
		}
	}

	score := (categories.Certificate + categories.Protocols +
		categories.KeyExchange + categories.CipherStrength) / 4
	letter := gradeOfScore(score)

	caps := gradeCaps(result, &issues)
	for _, capability := range caps {
		letter = lowerGrade(letter, capGrade(capability))
	}

	if len(caps) == 0 && score >= 100 && result.Hsts.Present {
		letter = "A+"
	}

	return GradeResult{
		Letter:     letter,
		Score:      score,
		Categories: categories,
		Caps:       caps,
		Issues:     issues,
	}
}

// clampScore 把分数限制在 0..100。
func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
