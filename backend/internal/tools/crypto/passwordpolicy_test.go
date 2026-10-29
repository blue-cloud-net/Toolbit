package crypto

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestPasswordPolicyCompliant 断言满足策略。
func TestPasswordPolicyCompliant(t *testing.T) {
	result, err := PasswordPolicy(t.Context(), &PasswordPolicyRequest{
		Password: "Zx7#quietRiver9",
		Preset:   PolicyPresetMedium,
	})
	data := result.Data.(PasswordPolicyResult)
	if err != nil {
		t.Fatalf("检查失败：%v", err)
	}
	if !data.Compliant {
		t.Errorf("应判定合规：%v", data.Suggestions)
	}
	if len(data.Checks) == 0 {
		t.Error("应给出检查项")
	}
}

// TestPasswordPolicyNonCompliant 断言不满足策略。
func TestPasswordPolicyNonCompliant(t *testing.T) {
	result, err := PasswordPolicy(t.Context(), &PasswordPolicyRequest{
		Password: "abc",
		Preset:   PolicyPresetStrict,
	})
	data := result.Data.(PasswordPolicyResult)
	if err != nil {
		t.Fatalf("检查失败：%v", err)
	}
	if data.Compliant {
		t.Error("应判定不合规")
	}
	if len(data.Suggestions) == 0 {
		t.Error("应给出改进建议")
	}
}

// TestPasswordPolicyUnsupportedPreset 断言未知预设返回 INVALID_FIELD。
func TestPasswordPolicyUnsupportedPreset(t *testing.T) {
	_, err := PasswordPolicy(t.Context(), &PasswordPolicyRequest{
		Password: "abc",
		Preset:   "extreme",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
