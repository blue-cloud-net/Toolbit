package network

import "testing"

// TestGradeOf 覆盖等级映射的边界。
func TestGradeOf(t *testing.T) {
	cases := []struct {
		score int
		want  string
	}{
		{scoreMax, "A+"},
		{95, "A"},
		{90, "A"},
		{89, "B"},
		{85, "B"},
		{75, "B"},
		{74, "C"},
		{1, "C"},
		{0, "T"},
		{-10, "T"},
	}
	for _, tc := range cases {
		if got := gradeOf(tc.score); got != tc.want {
			t.Errorf("gradeOf(%d) = %q，期望 %q", tc.score, got, tc.want)
		}
	}
}

// TestEvaluateTLSProbePerfect 覆盖各项全过的情形。
func TestEvaluateTLSProbePerfect(t *testing.T) {
	result := &TLSProbeResult{
		NegotiatedVersion: "TLS 1.3",
		CipherSuite:       "TLS_AES_128_GCM_SHA256",
		ChainValid:        true,
		Verified:          true,
		OcspStapled:       true,
	}
	grade, checks := evaluateTLSProbe(result)
	if grade != "A+" {
		t.Errorf("评级 = %q，期望 A+", grade)
	}
	if len(checks) != 6 {
		t.Fatalf("检查项数量 = %d，期望 6", len(checks))
	}
	for _, check := range checks {
		if !check.Passed {
			t.Errorf("检查项 %q 未通过，期望通过", check.ID)
		}
	}
}

// TestEvaluateTLSProbeChainInvalid 覆盖证书链无效的致命情形。
func TestEvaluateTLSProbeChainInvalid(t *testing.T) {
	result := &TLSProbeResult{
		NegotiatedVersion: "TLS 1.3",
		CipherSuite:       "TLS_AES_128_GCM_SHA256",
		ChainValid:        false,
		ChainErrors:       []string{"expired certificate"},
		Verified:          false,
		OcspStapled:       true,
	}
	grade, checks := evaluateTLSProbe(result)
	if grade != "T" {
		t.Errorf("评级 = %q，期望 T", grade)
	}
	chain := findCheck(checks, "chain")
	if chain == nil {
		t.Fatal("缺少 chain 检查项")
	}
	if chain.Passed {
		t.Error("chain 检查项应未通过")
	}
	if chain.Detail != "expired certificate" {
		t.Errorf("chain Detail = %q，期望 %q", chain.Detail, "expired certificate")
	}
}

// TestEvaluateTLSProbeOldTLS 覆盖协议版本过低的致命情形。
func TestEvaluateTLSProbeOldTLS(t *testing.T) {
	result := &TLSProbeResult{
		NegotiatedVersion: "TLS 1.1",
		CipherSuite:       "TLS_RSA_WITH_AES_128_CBC_SHA",
		ChainValid:        true,
		Verified:          true,
		OcspStapled:       true,
	}
	grade, _ := evaluateTLSProbe(result)
	if grade != "T" {
		t.Errorf("评级 = %q，期望 T", grade)
	}
}

// TestEvaluateTLSProbeUntrusted 覆盖未通过信任根验证的降级。
func TestEvaluateTLSProbeUntrusted(t *testing.T) {
	result := &TLSProbeResult{
		NegotiatedVersion: "TLS 1.3",
		CipherSuite:       "TLS_AES_128_GCM_SHA256",
		ChainValid:        true,
		Verified:          false,
		OcspStapled:       true,
	}
	grade, checks := evaluateTLSProbe(result)
	if grade != "C" {
		t.Errorf("评级 = %q，期望 C", grade)
	}
	trust := findCheck(checks, "trust")
	if trust == nil || trust.Passed {
		t.Error("trust 检查项应未通过")
	}
}

// TestEvaluateTLSProbeImplicitTrust 覆盖「隐式信任」不扣信任分的情形。
func TestEvaluateTLSProbeImplicitTrust(t *testing.T) {
	result := &TLSProbeResult{
		NegotiatedVersion: "TLS 1.3",
		CipherSuite:       "TLS_AES_128_GCM_SHA256",
		ChainValid:        true,
		Verified:          false,
		ImplicitTrust:     true,
		OcspStapled:       true,
	}
	grade, checks := evaluateTLSProbe(result)
	// 100 - 20（隐式信任）= 80 → B
	if grade != "B" {
		t.Errorf("评级 = %q，期望 B（隐式信任扣 20 分）", grade)
	}
	trust := findCheck(checks, "trust")
	if trust == nil || trust.Detail == "" {
		t.Error("trust 检查项应给出隐式信任的说明")
	}
}

// TestEvaluateTLSProbeNoOCSP 覆盖缺少 OCSP 装订的轻微扣分。
func TestEvaluateTLSProbeNoOCSP(t *testing.T) {
	result := &TLSProbeResult{
		NegotiatedVersion: "TLS 1.2",
		CipherSuite:       "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		ChainValid:        true,
		Verified:          true,
		OcspStapled:       false,
	}
	grade, checks := evaluateTLSProbe(result)
	// 100 - 5（无 OCSP）- 5（未用 TLS 1.3）= 90 → A
	if grade != "A" {
		t.Errorf("评级 = %q，期望 A", grade)
	}
	if ocsp := findCheck(checks, "ocsp"); ocsp == nil || ocsp.Passed {
		t.Error("ocsp 检查项应未通过")
	}
}

// findCheck 按 ID 查找检查项。
func findCheck(checks []TLSCheck, id string) *TLSCheck {
	for i := range checks {
		if checks[i].ID == id {
			return &checks[i]
		}
	}
	return nil
}
