package pki

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCertLifecycleMissingCertificate 断言缺少证书返回 MISSING_FIELD。
func TestCertLifecycleMissingCertificate(t *testing.T) {
	_, err := CertLifecycle(t.Context(), &CertLifecycleRequest{
		IsPEM:       true,
		WarningDays: 30,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestFormatDays 断言天数渲染口径。
//
// 与前端 `certDetail.ts` 的 `formatDays` **必须一致**：
// 不足 10 天保留 1 位小数，否则取整 —— 否则页面上的数字与
// 「续期建议」文案会对不上。
func TestFormatDays(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{9.6, "9.6"},
		{0.4, "0.4"},
		{12, "12"},
		{820.7, "820"},
		{-2.4, "-2.4"},
		{-30.5, "-30"},
	}
	for _, c := range cases {
		if got := formatDays(c.in); got != c.want {
			t.Errorf("formatDays(%v) = %q，期望 %q", c.in, got, c.want)
		}
	}
}

// TestCertLifecycleFractionalDays 断言剩余天数是**小数**（保留 1 位）。
//
// 临期证书（剩余不足 10 天）必须能区分「9.6 天」与「9 天」，
// 整数天数会把这层信息丢掉。
func TestCertLifecycleFractionalDays(t *testing.T) {
	res, err := CertLifecycle(context.Background(), &CertLifecycleRequest{
		Certificate: loadFixture(t, "leaf.crt.pem"),
		IsPEM:       true,
		WarningDays: 30,
	})
	if err != nil {
		t.Fatalf("CertLifecycle 失败：%v", err)
	}
	data, ok := res.Data.(CertLifecycleResult)
	if !ok {
		t.Fatalf("响应类型异常：%T", res.Data)
	}
	if data.DaysRemaining <= 0 {
		t.Errorf("DaysRemaining = %v，期望 > 0", data.DaysRemaining)
	}
	// 小数天数必须能被精确表示到 1 位。
	if diff := data.DaysRemaining*10 - float64(int64(data.DaysRemaining*10)); diff > 0.01 || diff < -0.01 {
		t.Errorf("DaysRemaining = %v，不是 1 位小数", data.DaysRemaining)
	}
	if data.ValidDays <= 0 {
		t.Errorf("ValidDays = %v，期望 > 0", data.ValidDays)
	}
}
