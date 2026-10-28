package encoding

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func TestNumberBaseConvert(t *testing.T) {
	cases := []struct {
		name  string
		value string
		from  int
		to    int
		want  string
	}{
		{"十进制到十六进制", "255", 10, 16, "ff"},
		{"十六进制到十进制", "ff", 16, 10, "255"},
		{"二进制到十进制", "1010", 2, 10, "10"},
		{"十进制到二进制", "10", 10, 2, "1010"},
		{"三十六进制", "zz", 36, 10, "1295"},
		{"八进制", "17", 8, 10, "15"},
		{"同进制", "123", 10, 10, "123"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := NumberBase(context.Background(), &NumberBaseRequest{
				Value: tc.value, From: tc.from, To: tc.to,
			})
			if err != nil {
				t.Fatalf("转换失败: %v", err)
			}
			got := res.Data.(NumberBaseResult).Result
			if got != tc.want {
				t.Errorf("结果 = %q，期望 %q", got, tc.want)
			}
		})
	}
}

func TestNumberBaseArbitraryPrecision(t *testing.T) {
	// 超过 64 位的大整数必须精确。
	const big = "18446744073709551616" // 2^64
	res, err := NumberBase(context.Background(), &NumberBaseRequest{
		Value: big, From: 10, To: 16,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if got := res.Data.(NumberBaseResult).Result; got != "10000000000000000" {
		t.Errorf("结果 = %q，期望 %q", got, "10000000000000000")
	}
}

func TestNumberBaseNegative(t *testing.T) {
	res, err := NumberBase(context.Background(), &NumberBaseRequest{
		Value: "-42", From: 10, To: 16,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if got := res.Data.(NumberBaseResult).Result; got != "-2a" {
		t.Errorf("结果 = %q，期望 %q", got, "-2a")
	}
}

func TestNumberBasePrefix(t *testing.T) {
	cases := []struct {
		base int
		want string
	}{
		{2, "0b1010"},
		{8, "0o12"},
		{16, "0xa"},
		{10, "10"},
	}
	for _, tc := range cases {
		res, err := NumberBase(context.Background(), &NumberBaseRequest{
			Value: "10", From: 10, To: tc.base, Prefix: true,
		})
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}
		if got := res.Data.(NumberBaseResult).Result; got != tc.want {
			t.Errorf("进制 %d 的结果 = %q，期望 %q", tc.base, got, tc.want)
		}
	}
}

func TestNumberBaseInvalidBase(t *testing.T) {
	cases := []struct {
		name string
		from int
		to   int
	}{
		{"源进制为 1", 1, 10},
		{"源进制为 37", 37, 10},
		{"目标进制为 0", 10, 0},
		{"目标进制为 99", 10, 99},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NumberBase(context.Background(), &NumberBaseRequest{
				Value: "10", From: tc.from, To: tc.to,
			})
			if err == nil {
				t.Fatal("期望校验失败")
			}
			toolErr, ok := err.(*toolerr.Error)
			if !ok {
				t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
			}
			if toolErr.Code != toolerr.CodeInvalidField {
				t.Errorf("错误码 = %s，期望 INVALID_FIELD", toolErr.Code)
			}
		})
	}
}

func TestNumberBaseMissingValue(t *testing.T) {
	_, err := NumberBase(context.Background(), &NumberBaseRequest{
		Value: "   ", From: 10, To: 16,
	})
	if err == nil {
		t.Fatal("期望校验失败")
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeMissingField {
		t.Errorf("错误码 = %s，期望 MISSING_FIELD", toolErr.Code)
	}
}

func TestNumberBaseInvalidDigits(t *testing.T) {
	_, err := NumberBase(context.Background(), &NumberBaseRequest{
		Value: "12", From: 2, To: 10,
	})
	if err == nil {
		t.Fatal("期望校验失败")
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 INVALID_FIELD", toolErr.Code)
	}
}
