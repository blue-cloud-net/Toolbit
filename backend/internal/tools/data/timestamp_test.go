package data

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestTimestampToDate 断言时间戳转日期。
func TestTimestampToDate(t *testing.T) {
	result, err := Timestamp(context.Background(), &TimestampRequest{
		Value: Scalar{Value: "1700000000", Set: true},
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if result.Timestamp.S != 1700000000 {
		t.Errorf("秒时间戳 = %d", result.Timestamp.S)
	}
	if result.Timestamp.Ms != 1700000000000 {
		t.Errorf("毫秒时间戳 = %d", result.Timestamp.Ms)
	}
	if result.UTC != "2023-11-14 22:13:20" {
		t.Errorf("UTC = %q", result.UTC)
	}
}

// TestTimestampMillisecondAutoDetect 断言按位数识别毫秒。
func TestTimestampMillisecondAutoDetect(t *testing.T) {
	result, err := Timestamp(context.Background(), &TimestampRequest{
		Value: Scalar{Value: "1700000000123", Set: true},
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if result.Timestamp.S != 1700000000 {
		t.Errorf("秒时间戳 = %d，期望 1700000000", result.Timestamp.S)
	}
	if result.Timestamp.Ms != 1700000000123 {
		t.Errorf("毫秒时间戳 = %d", result.Timestamp.Ms)
	}
}

// TestTimestampTimezone 断言时区生效。
func TestTimestampTimezone(t *testing.T) {
	result, err := Timestamp(context.Background(), &TimestampRequest{
		Value:    Scalar{Value: "1700000000", Set: true},
		Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if result.Local != "2023-11-15 06:13:20" {
		t.Errorf("本地时间 = %q", result.Local)
	}
}

// TestTimestampToTimestampOperation 断言日期转时间戳。
func TestTimestampToTimestampOperation(t *testing.T) {
	result, err := Timestamp(context.Background(), &TimestampRequest{
		Value:     Scalar{Value: "2023-11-14", Set: true},
		Operation: OpToTimestamp,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if result.Timestamp.S != 1699920000 {
		t.Errorf("秒时间戳 = %d，期望 1699920000", result.Timestamp.S)
	}
}

// TestTimestampInvalidValue 断言非法值报 INVALID_TIME。
func TestTimestampInvalidValue(t *testing.T) {
	_, err := Timestamp(context.Background(), &TimestampRequest{
		Value: Scalar{Value: "not-a-number", Set: true},
	})
	assertCode(t, err, toolerr.CodeInvalidTime)
}

// TestTimestampInvalidTimezone 断言非法时区报 INVALID_TIME。
func TestTimestampInvalidTimezone(t *testing.T) {
	_, err := Timestamp(context.Background(), &TimestampRequest{
		Value:    Scalar{Value: "1700000000", Set: true},
		Timezone: "Not/AZone",
	})
	assertCode(t, err, toolerr.CodeInvalidTime)
}

// TestTimestampInvalidUnit 断言未知单位报 INVALID_FIELD。
func TestTimestampInvalidUnit(t *testing.T) {
	_, err := Timestamp(context.Background(), &TimestampRequest{
		Value: Scalar{Value: "1700000000", Set: true},
		Unit:  "minutes",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestScalarUnmarshal 断言标量同时接受字符串与数字。
func TestScalarUnmarshal(t *testing.T) {
	var s Scalar
	if err := s.UnmarshalJSON([]byte(`"1700000000"`)); err != nil {
		t.Fatalf("解析字符串失败：%v", err)
	}
	if s.Value != "1700000000" {
		t.Errorf("值 = %q", s.Value)
	}

	var n Scalar
	if err := n.UnmarshalJSON([]byte(`1700000000`)); err != nil {
		t.Fatalf("解析数字失败：%v", err)
	}
	if n.Value != "1700000000" {
		t.Errorf("值 = %q", n.Value)
	}
}
