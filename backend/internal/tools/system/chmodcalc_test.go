package system

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestChmodCalcOctal 断言八进制解析。
func TestChmodCalcOctal(t *testing.T) {
	result, err := ChmodCalc(&ChmodCalcRequest{Mode: "755"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.Octal != "755" {
		t.Errorf("八进制 = %q，期望 755", result.Octal)
	}
	if result.Symbolic != "rwxr-xr-x" {
		t.Errorf("符号形式 = %q，期望 rwxr-xr-x", result.Symbolic)
	}
	if len(result.Classes) != 3 {
		t.Fatalf("类别数 = %d，期望 3", len(result.Classes))
	}
	if !result.Classes[0].Read || !result.Classes[0].Write || !result.Classes[0].Execute {
		t.Errorf("属主权限不正确：%+v", result.Classes[0])
	}
	if !result.Classes[1].Read || result.Classes[1].Write || !result.Classes[1].Execute {
		t.Errorf("属组权限不正确：%+v", result.Classes[1])
	}
	if !result.Classes[2].Read || result.Classes[2].Write || !result.Classes[2].Execute {
		t.Errorf("其他人权限不正确：%+v", result.Classes[2])
	}
}

// TestChmodCalcSymbolic 断言符号形式解析。
func TestChmodCalcSymbolic(t *testing.T) {
	result, err := ChmodCalc(&ChmodCalcRequest{Mode: "rwxr-x---"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.Octal != "750" {
		t.Errorf("八进制 = %q，期望 750", result.Octal)
	}
}

// TestChmodCalcLongSymbolic 断言长形式符号解析。
func TestChmodCalcLongSymbolic(t *testing.T) {
	result, err := ChmodCalc(&ChmodCalcRequest{Mode: "u=rwx,g=rx,o=r"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.Octal != "754" {
		t.Errorf("八进制 = %q，期望 754", result.Octal)
	}
}

// TestChmodCalcBitOverride 断言权限位覆盖 mode。
func TestChmodCalcBitOverride(t *testing.T) {
	result, err := ChmodCalc(&ChmodCalcRequest{
		Mode:  "755",
		Other: []string{PermRead},
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.Octal != "754" {
		t.Errorf("八进制 = %q，期望 754", result.Octal)
	}
}

// TestChmodCalcIgnoresSpecialBits 断言只处理低 9 位。
//
// setuid / setgid / sticky **不在**工具范围内。
func TestChmodCalcIgnoresSpecialBits(t *testing.T) {
	result, err := ChmodCalc(&ChmodCalcRequest{Mode: "4755"})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.Octal != "755" {
		t.Errorf("八进制 = %q，期望 755", result.Octal)
	}
}

// TestChmodCalcInvalidMode 断言非法权限报 INVALID_CHMOD。
func TestChmodCalcInvalidMode(t *testing.T) {
	_, err := ChmodCalc(&ChmodCalcRequest{Mode: "rwxr-xr"})
	assertCode(t, err, toolerr.CodeInvalidChmod)
}

// TestChmodCalcInvalidBit 断言未知权限位报 INVALID_FIELD。
func TestChmodCalcInvalidBit(t *testing.T) {
	_, err := ChmodCalc(&ChmodCalcRequest{
		Mode: "755",
		User: []string{"append"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestChmodCalcMissingMode 断言缺少输入时报 MISSING_FIELD。
func TestChmodCalcMissingMode(t *testing.T) {
	_, err := ChmodCalc(&ChmodCalcRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}

// assertCode 断言错误的错误码。
func assertCode(t *testing.T, err error, want toolerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望错误 %s，实际无错误", want)
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("期望 *toolerr.Error，实际 %T", err)
	}
	if toolErr.Code != want {
		t.Errorf("错误码 = %s，期望 %s", toolErr.Code, want)
	}
}
