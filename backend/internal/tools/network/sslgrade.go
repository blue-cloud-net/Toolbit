package network

// TLS 探测结果的评级与逐项清单。
//
// 评级规则（见 docs/api/tools/ssl-check.md §评级）：
//
//	每项检查按「通过 / 不通过」计分，不通过的项按严重度扣分：
//	  - 证书链无效、主机名不匹配：致命，直接判 T；
//	  - 未协商出 TLS 1.2 及以上：致命，判 T；
//	  - 未通过系统信任根验证：重扣，判 C；
//	  - 未启用 OCSP 装订：轻扣；
//	  - 未协商出 TLS 1.3：轻扣。
//
//	按剩余分映射到等级：
//	  A+（满分）、A、B、C、T（有致命项）。
//
//	评级只描述**本次握手与证书链**的观测结果，不代表站点整体安全评分。

// TLS 检查项。
type TLSCheck struct {
	// ID 是检查项的稳定标识。
	ID string `json:"id"`
	// Label 是检查项的中文标签。
	Label string `json:"label"`
	// Passed 报告该项是否通过。
	Passed bool `json:"passed"`
	// Detail 是补充说明；为空时省略。
	Detail string `json:"detail,omitempty"`
}

// 评级扣分权重。
const (
	// scoreFatal 是致命项扣分；扣满即判 T。
	scoreFatal = 100
	// scoreTrust 是未通过信任根验证的扣分。
	scoreTrust = 40
	// scoreImplicitTrust 是「隐式信任」（无法访问系统信任根）的扣分。
	scoreImplicitTrust = 20
	// scoreOCSP 是未启用 OCSP 装订的扣分。
	scoreOCSP = 5
	// scoreTLS13 是未协商出 TLS 1.3 的扣分。
	scoreTLS13 = 5
	// scoreMax 是满分。
	scoreMax = 100
)

// gradeOf 按剩余分映射等级。
func gradeOf(score int) string {
	switch {
	case score >= scoreMax:
		return "A+"
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score > 0:
		return "C"
	default:
		return "T"
	}
}
// evaluateTLSProbe 对探测结果评级并生成逐项清单。
func evaluateTLSProbe(result *TLSProbeResult) (string, []TLSCheck) {
	checks := make([]TLSCheck, 0, 6)
	score := scoreMax

	// 1) 证书链完整性。
	chainDetail := ""
	if len(result.ChainErrors) > 0 {
		chainDetail = result.ChainErrors[0]
	}
	checks = append(checks, TLSCheck{
		ID:     "chain",
		Label:  "证书链完整",
		Passed: result.ChainValid,
		Detail: chainDetail,
	})
	if !result.ChainValid {
		score -= scoreFatal
	}

	// 2) 系统信任根验证。
	trustDetail := ""
	if result.ImplicitTrust {
		trustDetail = "未暴露系统信任根，退化为链内自洽判定"
	}
	checks = append(checks, TLSCheck{
		ID:     "trust",
		Label:  "受信任颁发机构签发",
		Passed: result.Verified,
		Detail: trustDetail,
	})
	if !result.Verified {
		if result.ImplicitTrust {
			score -= scoreImplicitTrust
		} else {
			score -= scoreTrust
		}
	}

	// 3) 协议版本下限。
	versionOK := result.NegotiatedVersion == "TLS 1.3" ||
		result.NegotiatedVersion == "TLS 1.2"
	checks = append(checks, TLSCheck{
		ID:     "tlsVersion",
		Label:  "支持 TLS 1.2 及以上",
		Passed: versionOK,
		Detail: result.NegotiatedVersion,
	})
	if !versionOK {
		score -= scoreFatal
	}

	// 4) TLS 1.3 支持。
	tls13 := result.NegotiatedVersion == "TLS 1.3"
	checks = append(checks, TLSCheck{
		ID:     "tls13",
		Label:  "协商出 TLS 1.3",
		Passed: tls13,
		Detail: result.NegotiatedVersion,
	})
	if !tls13 {
		score -= scoreTLS13
	}

	// 5) OCSP 装订。
	checks = append(checks, TLSCheck{
		ID:     "ocsp",
		Label:  "携带 OCSP 装订响应",
		Passed: result.OcspStapled,
	})
	if !result.OcspStapled {
		score -= scoreOCSP
	}

	// 6) 密码套件已协商。
	cipherOK := result.CipherSuite != ""
	checks = append(checks, TLSCheck{
		ID:     "cipher",
		Label:  "协商出密码套件",
		Passed: cipherOK,
		Detail: result.CipherSuite,
	})
	if !cipherOK {
		score -= scoreFatal
	}

	if score < 0 {
		score = 0
	}
	return gradeOf(score), checks
}
