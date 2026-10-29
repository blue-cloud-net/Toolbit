package data

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDateCalcDiff 断言日期差。
func TestDateCalcDiff(t *testing.T) {
	result, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpDiff,
		DateA:     "2024-01-01",
		DateB:     "2024-01-31",
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.DiffDays != 30 {
		t.Errorf("相差天数 = %d，期望 30", result.DiffDays)
	}
}

// TestDateCalcAdd 断言日期加法。
func TestDateCalcAdd(t *testing.T) {
	days := 10
	result, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpAdd,
		DateA:     "2024-01-01",
		DaysRaw:   &days,
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.ResultDate != "2024-01-11" {
		t.Errorf("结果日期 = %q", result.ResultDate)
	}
}

// TestDateCalcSubMonth 断言按月减法。
func TestDateCalcSubMonth(t *testing.T) {
	months := 2
	result, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpSub,
		DateA:     "2024-03-15",
		DaysRaw:   &months,
		Unit:      DateUnitMonth,
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.ResultDate != "2024-01-15" {
		t.Errorf("结果日期 = %q", result.ResultDate)
	}
}

// TestDateCalcWorkdays 断言工作日统计。
func TestDateCalcWorkdays(t *testing.T) {
	result, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpWorkdays,
		DateA:     "2024-01-01",
		DateB:     "2024-01-07",
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	// 2024-01-01 是周一，1 月 7 日是周日，共 5 个工作日。
	if result.Workdays != 5 {
		t.Errorf("工作日数 = %d，期望 5", result.Workdays)
	}
}

// TestDateCalcWorkdaysWithHolidays 断言节假日被排除。
func TestDateCalcWorkdaysWithHolidays(t *testing.T) {
	result, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpWorkdays,
		DateA:     "2024-01-01",
		DateB:     "2024-01-07",
		Holidays:  []string{"2024-01-02"},
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.Workdays != 4 {
		t.Errorf("工作日数 = %d，期望 4", result.Workdays)
	}
}

// TestDateCalcMissingDateB 断言缺少第二个日期时报 MISSING_FIELD。
func TestDateCalcMissingDateB(t *testing.T) {
	_, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpDiff,
		DateA:     "2024-01-01",
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestDateCalcMissingDays 断言缺少天数时报 MISSING_FIELD。
func TestDateCalcMissingDays(t *testing.T) {
	_, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpAdd,
		DateA:     "2024-01-01",
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestDateCalcZeroDaysAllowed 断言 0 天是合法输入。
func TestDateCalcZeroDaysAllowed(t *testing.T) {
	days := 0
	result, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpAdd,
		DateA:     "2024-01-01",
		DaysRaw:   &days,
	})
	if err != nil {
		t.Fatalf("计算失败：%v", err)
	}
	if result.ResultDate != "2024-01-01" {
		t.Errorf("结果日期 = %q", result.ResultDate)
	}
}

// TestDateCalcInvalidDate 断言非法日期报 INVALID_TIME。
func TestDateCalcInvalidDate(t *testing.T) {
	_, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpDiff,
		DateA:     "not-a-date",
		DateB:     "2024-01-01",
	})
	assertCode(t, err, toolerr.CodeInvalidTime)
}

// TestDateCalcInvalidHoliday 断言非法节假日报 INVALID_TIME。
func TestDateCalcInvalidHoliday(t *testing.T) {
	_, err := DateCalc(context.Background(), &DateCalcRequest{
		Operation: DateOpWorkdays,
		DateA:     "2024-01-01",
		DateB:     "2024-01-07",
		Holidays:  []string{"01-01"},
	})
	assertCode(t, err, toolerr.CodeInvalidTime)
}
