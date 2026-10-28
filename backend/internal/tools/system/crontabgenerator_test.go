package system

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCrontabGeneratorPresets 断言预设生成正确的表达式。
func TestCrontabGeneratorPresets(t *testing.T) {
	cases := []struct {
		preset string
		want   string
	}{
		{PresetEveryMinute, "* * * * *"},
		{PresetEveryFiveMinutes, "*/5 * * * *"},
		{PresetHourly, "0 * * * *"},
		{PresetDaily, "0 0 * * *"},
		{PresetWeekly, "0 0 * * 0"},
		{PresetMonthly, "0 0 1 * *"},
		{PresetYearly, "0 0 1 1 *"},
	}
	for _, tc := range cases {
		t.Run(tc.preset, func(t *testing.T) {
			result, err := CrontabGenerator(&CrontabGeneratorRequest{Preset: tc.preset})
			if err != nil {
				t.Fatalf("生成失败：%v", err)
			}
			if result.Expression != tc.want {
				t.Errorf("表达式 = %q，期望 %q", result.Expression, tc.want)
			}
			if len(result.Occurrences) != 5 {
				t.Errorf("预览条数 = %d，期望 5", len(result.Occurrences))
			}
		})
	}
}

// TestCrontabGeneratorCustomFields 断言自定义字段被解析。
func TestCrontabGeneratorCustomFields(t *testing.T) {
	result, err := CrontabGenerator(&CrontabGeneratorRequest{
		Minute:     "30",
		Hour:       "2",
		DayOfMonth: "15",
		Month:      "*",
		DayOfWeek:  "*",
		Count:      3,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if result.Expression != "30 2 15 * *" {
		t.Errorf("表达式 = %q", result.Expression)
	}
	if len(result.Occurrences) != 3 {
		t.Errorf("预览条数 = %d，期望 3", len(result.Occurrences))
	}
}

// TestCrontabGeneratorStepList 断言步长与列表语法。
func TestCrontabGeneratorStepList(t *testing.T) {
	result, err := CrontabGenerator(&CrontabGeneratorRequest{
		Minute: "0,30",
		Hour:   "*/6",
		Count:  4,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if result.Expression != "0,30 */6 * * *" {
		t.Errorf("表达式 = %q", result.Expression)
	}
}

// TestCrontabGeneratorInvalidExpression 断言非法表达式报 INVALID_CRON。
func TestCrontabGeneratorInvalidExpression(t *testing.T) {
	_, err := CrontabGenerator(&CrontabGeneratorRequest{
		Minute: "99",
	})
	assertCode(t, err, toolerr.CodeInvalidCron)
}

// TestCrontabGeneratorInvalidPreset 断言未知预设报 INVALID_FIELD。
func TestCrontabGeneratorInvalidPreset(t *testing.T) {
	_, err := CrontabGenerator(&CrontabGeneratorRequest{Preset: "nope"})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestCrontabGeneratorInvalidTimezone 断言非法时区报 INVALID_TIME。
func TestCrontabGeneratorInvalidTimezone(t *testing.T) {
	_, err := CrontabGenerator(&CrontabGeneratorRequest{
		Preset:   PresetHourly,
		Timezone: "Not/AZone",
	})
	assertCode(t, err, toolerr.CodeInvalidTime)
}

// TestCrontabGeneratorTimezone 断言时区被应用到预览。
func TestCrontabGeneratorTimezone(t *testing.T) {
	result, err := CrontabGenerator(&CrontabGeneratorRequest{
		Preset:   PresetHourly,
		Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if result.Timezone != "Asia/Shanghai" {
		t.Errorf("时区 = %q", result.Timezone)
	}
	if !strings.Contains(result.Occurrences[0].Time, "+08:00") {
		t.Errorf("触发时间未体现时区：%s", result.Occurrences[0].Time)
	}
}

// TestParseCronField 断言字段解析。
func TestParseCronField(t *testing.T) {
	set, err := parseCronField("*/15", 0, 59)
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if len(set) != 4 {
		t.Fatalf("取值数 = %d，期望 4", len(set))
	}
	for _, v := range []int{0, 15, 30, 45} {
		if !set[v] {
			t.Errorf("缺少取值 %d", v)
		}
	}
}

// TestDescribeCron 断言描述可读。
func TestDescribeCron(t *testing.T) {
	got := describeCron([]string{"0", "3", "*", "*", "*"})
	if !strings.Contains(got, "3") {
		t.Errorf("描述缺少小时：%s", got)
	}
}
