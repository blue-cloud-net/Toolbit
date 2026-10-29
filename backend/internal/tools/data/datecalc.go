package data

import (
	"context"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 日期计算操作。
const (
	// DateOpDiff 计算两个日期相差的天数。
	DateOpDiff = "diff"
	// DateOpAdd 给日期加上若干天。
	DateOpAdd = "add"
	// DateOpSub 给日期减去若干天。
	DateOpSub = "sub"
	// DateOpWorkdays 统计区间内的工作日。
	DateOpWorkdays = "workdays"
)

// DateOperations 返回受支持的操作（顺序稳定）。
func DateOperations() []string {
	return []string{DateOpDiff, DateOpAdd, DateOpSub, DateOpWorkdays}
}

// 时间单位。
const (
	// DateUnitDay 是天。
	DateUnitDay = "day"
	// DateUnitWeek 是周。
	DateUnitWeek = "week"
	// DateUnitMonth 是月。
	DateUnitMonth = "month"
	// DateUnitYear 是年。
	DateUnitYear = "year"
)

// DateUnits 返回受支持的单位（顺序稳定）。
func DateUnits() []string {
	return []string{DateUnitDay, DateUnitWeek, DateUnitMonth, DateUnitYear}
}

// maxHolidayCount 是节假日列表的长度上限。
const maxHolidayCount = 500

// DateCalcRequest 是 date-calc 工具的请求。
type DateCalcRequest struct {
	// Operation 是操作：diff / add / sub / workdays。
	Operation string `json:"operation"`
	// DateA 是第一个日期。
	DateA string `json:"dateA"`
	// DateB 是第二个日期；diff 与 workdays 时必填。
	DateB string `json:"dateB"`
	// DaysRaw 是增减的天数；add 与 sub 时必填。
	//
	// 0 是合法值，因此用指针区分「未给出」与「给出 0」。
	DaysRaw *int `json:"days"`
	// Unit 是增减单位；缺省 day。
	Unit string `json:"unit"`
	// Holidays 是节假日（YYYY-MM-DD）；workdays 时排除。
	Holidays []string `json:"holidays"`
}

// DateCalcResult 是 date-calc 工具的响应。
type DateCalcResult struct {
	// ResultDate 是计算结果日期（add / sub）。
	ResultDate string `json:"resultDate,omitempty"`
	// DiffDays 是相差天数（diff）。
	DiffDays int `json:"diffDays"`
	// Workdays 是工作日数（workdays）。
	Workdays int `json:"workdays"`
}

// DateCalc 计算日期差与增减。
//
// `diff` 按**自然日**计算（忽略时分秒）。
// 节假日列表**只**在 workdays 中生效：本工具不内置法定节假日。
func DateCalc(ctx context.Context, req *DateCalcRequest) (DateCalcResult, error) {
	if err := req.Validate(); err != nil {
		return DateCalcResult{}, err
	}

	dateA, err := parseDate(req.DateA)
	if err != nil {
		return DateCalcResult{}, err
	}

	switch req.Operation {
	case DateOpDiff:
		dateB, err := parseDate(req.DateB)
		if err != nil {
			return DateCalcResult{}, err
		}
		return DateCalcResult{
			DiffDays: int(dateB.Sub(dateA).Hours() / 24),
		}, nil

	case DateOpAdd, DateOpSub:
		sign := 1
		if req.Operation == DateOpSub {
			sign = -1
		}
		result := addDuration(dateA, req.unit(), sign*req.days())
		return DateCalcResult{
			ResultDate: result.Format("2006-01-02"),
		}, nil

	default:
		dateB, err := parseDate(req.DateB)
		if err != nil {
			return DateCalcResult{}, err
		}
		holidays, err := parseHolidays(req.Holidays)
		if err != nil {
			return DateCalcResult{}, err
		}
		return DateCalcResult{
			Workdays: countWorkdays(dateA, dateB, holidays),
		}, nil
	}
}

// Validate 做字段级校验。
func (r *DateCalcRequest) Validate() error {
	switch r.Operation {
	case DateOpDiff, DateOpAdd, DateOpSub, DateOpWorkdays:
	default:
		return tools.InvalidField("operation", DateOperations(), r.Operation)
	}
	switch r.Unit {
	case "", DateUnitDay, DateUnitWeek, DateUnitMonth, DateUnitYear:
	default:
		return tools.InvalidField("unit", DateUnits(), r.Unit)
	}
	if strings.TrimSpace(r.DateA) == "" {
		return tools.MissingField("dateA")
	}
	switch r.Operation {
	case DateOpDiff, DateOpWorkdays:
		if strings.TrimSpace(r.DateB) == "" {
			return tools.MissingField("dateB")
		}
	case DateOpAdd, DateOpSub:
		if !r.DaysSet() {
			return tools.MissingField("days")
		}
	}
	if len(r.Holidays) > maxHolidayCount {
		return tools.InvalidField("holidays", []int{0, maxHolidayCount}, len(r.Holidays))
	}
	return nil
}

// DaysSet 报告增减天数是否给出。
//
// 0 是合法值，因此不能靠 `Days != 0` 判断。
func (r *DateCalcRequest) DaysSet() bool {
	return r.DaysRaw != nil
}

// days 返回增减天数；未给出时返回 0。
func (r *DateCalcRequest) days() int {
	if r.DaysRaw == nil {
		return 0
	}
	return *r.DaysRaw
}

// unit 返回单位；缺省 day。
func (r *DateCalcRequest) unit() string {
	if r.Unit == "" {
		return DateUnitDay
	}
	return r.Unit
}

// addDuration 按单位增减日期。
func addDuration(base time.Time, unit string, amount int) time.Time {
	switch unit {
	case DateUnitWeek:
		return base.AddDate(0, 0, 7*amount)
	case DateUnitMonth:
		return base.AddDate(0, amount, 0)
	case DateUnitYear:
		return base.AddDate(amount, 0, 0)
	default:
		return base.AddDate(0, 0, amount)
	}
}

// parseDate 解析日期。
//
// 支持 RFC 3339 与 `YYYY-MM-DD`；日期部分按 UTC 处理。
func parseDate(raw string) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
			"field": "date",
		})
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, text); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
		"field": "date",
	})
}

// parseHolidays 解析节假日列表。
func parseHolidays(raw []string) (map[string]bool, error) {
	out := map[string]bool{}
	for _, item := range raw {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		t, err := time.Parse("2006-01-02", text)
		if err != nil {
			return nil, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
				"field": "holidays",
			})
		}
		out[t.Format("2006-01-02")] = true
	}
	return out, nil
}

// countWorkdays 统计区间内的工作日。
//
// 区间是**闭区间**：两端都计入。
// 周末（周六 / 周日）与节假日都不计入。
func countWorkdays(from, to time.Time, holidays map[string]bool) int {
	if to.Before(from) {
		from, to = to, from
	}

	count := 0
	for cursor := from; !cursor.After(to); cursor = cursor.AddDate(0, 0, 1) {
		switch cursor.Weekday() {
		case time.Saturday, time.Sunday:
			continue
		}
		if holidays[cursor.Format("2006-01-02")] {
			continue
		}
		count++
	}
	return count
}
