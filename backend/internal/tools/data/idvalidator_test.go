package data

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestIDValidateIDCard 断言合法身份证通过校验。
func TestIDValidateIDCard(t *testing.T) {
	result, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  IDTypeIDCard,
		Value: "11010519491231002X",
	})
	data := result.Data.(IDValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if !data.Valid {
		t.Errorf("校验应通过：%s", data.Issue)
	}
	if data.CheckDigit != "X" {
		t.Errorf("校验位 = %q，期望 X", data.CheckDigit)
	}
}

// TestIDValidateIDCardBadCheckDigit 断言校验位不匹配时报错。
func TestIDValidateIDCardBadCheckDigit(t *testing.T) {
	result, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  IDTypeIDCard,
		Value: "110105194912310021",
	})
	data := result.Data.(IDValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.Valid {
		t.Error("校验位不匹配时 valid 应为 false")
	}
	if data.Issue == "" {
		t.Error("应给出失败原因")
	}
}

// TestIDValidateBankCard 断言银行卡 Luhn 校验。
func TestIDValidateBankCard(t *testing.T) {
	result, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  IDTypeBankCard,
		Value: "4111111111111111",
	})
	data := result.Data.(IDValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if !data.Valid {
		t.Errorf("校验应通过：%s", data.Issue)
	}
}

// TestIDValidateUnknownType 断言未知类型返回 INVALID_FIELD。
func TestIDValidateUnknownType(t *testing.T) {
	_, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  "passport",
		Value: "x",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
