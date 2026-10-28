package system

import (
	"strconv"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Crontab 预设。
const (
	// PresetEveryMinute 每分钟。
	PresetEveryMinute = "every-minute"
	// PresetEveryFiveMinutes 每 5 分钟。
	PresetEveryFiveMinutes = "every-5-minutes"
	// PresetHourly 每小时。
	PresetHourly = "hourly"
	// PresetDaily 每天午夜。
	PresetDaily = "daily"
	// PresetWeekly 每周日午夜。
	PresetWeekly = "weekly"
	// PresetMonthly 每月 1 日午夜。
	PresetMonthly = "monthly"
	// PresetYearly 每年 1 月 1 日午夜。
	PresetYearly = "yearly"
)

// CrontabPresets 返回受支持的预设（顺序稳定）。
func CrontabPresets() []string {
	return []string{
		PresetEveryMinute, PresetEveryFiveMinutes, PresetHourly,
		PresetDaily, PresetWeekly, PresetMonthly, PresetYearly,
	}
}

// cronFieldCount 是 crontab 表达式的字段数。
const cronFieldCount = 5

// maxCronOccurrences 是预览的触发时间条数上限。
const maxCronOccurrences = 10

// CrontabGeneratorRequest 是 crontab-generator 工具的请求。
type CrontabGeneratorRequest struct {
	// Preset 是预设名；给出时忽略五个字段。
	Preset string `json:"preset"`
	// Minute 是分钟字段。
	Minute string `json:"minute"`
	// Hour 是小时字段。
	Hour string `json:"hour"`
	// DayOfMonth 是日字段。
	DayOfMonth string `json:"dayOfMonth"`
	// Month 是月字段。
	Month string `json:"month"`
	// DayOfWeek 是星期字段。
	DayOfWeek string `json:"dayOfWeek"`
	// Timezone 是 IANA 时区名；缺省 UTC。
	Timezone string `json:"timezone"`
	// Count 是预览的触发次数；默认 5。
	Count int `json:"count"`
}

// CronOccurrence 是一次触发时间。
type CronOccurrence struct {
	// Time 是触发时刻（RFC 3339）。
	Time string `json:"time"`
}

// CrontabGeneratorResult 是 crontab-generator 工具的响应。
type CrontabGeneratorResult struct {
	// Expression 是五段式表达式。
	Expression string `json:"expression"`
	// Description 是表达式的人类可读描述。
	Description string `json:"description"`
	// Occurrences 是接下来的触发时间。
	Occurrences []CronOccurrence `json:"occurrences"`
	// Timezone 是使用的时区。
	Timezone string `json:"timezone"`
}

// CrontabGenerator 生成 crontab 表达式并预览触发时间。
//
// 本工具**不**解释命令部分，也不做任何安全评估。
func CrontabGenerator(req *CrontabGeneratorRequest) (CrontabGeneratorResult, error) {
	if err := req.Validate(); err != nil {
		return CrontabGeneratorResult{}, err
	}

	fields, err := req.fields()
	if err != nil {
		return CrontabGeneratorResult{}, err
	}

	loc, err := req.location()
	if err != nil {
		return CrontabGeneratorResult{}, err
	}

	schedule, err := parseCron(fields)
	if err != nil {
		return CrontabGeneratorResult{}, err
	}

	count := req.Count
	if count <= 0 {
		count = 5
	}
	if count > maxCronOccurrences {
		count = maxCronOccurrences
	}

	occurrences := []CronOccurrence{}
	now := time.Now().In(loc)
	cursor := now.Truncate(time.Minute).Add(time.Minute)
	for i := 0; i < count; i++ {
		next, ok := schedule.next(cursor)
		if !ok {
			break
		}
		occurrences = append(occurrences, CronOccurrence{
			Time: next.Format(time.RFC3339),
		})
		cursor = next.Add(time.Minute)
	}

	return CrontabGeneratorResult{
		Expression:  strings.Join(fields, " "),
		Description: describeCron(fields),
		Occurrences: occurrences,
		Timezone:    loc.String(),
	}, nil
}

// Validate 做字段级校验。
func (r *CrontabGeneratorRequest) Validate() error {
	if r.Preset != "" {
		switch r.Preset {
		case PresetEveryMinute, PresetEveryFiveMinutes, PresetHourly,
			PresetDaily, PresetWeekly, PresetMonthly, PresetYearly:
		default:
			return tools.InvalidField("preset", CrontabPresets(), r.Preset)
		}
		return nil
	}

	if r.Minute == "" && r.Hour == "" && r.DayOfMonth == "" &&
		r.Month == "" && r.DayOfWeek == "" {
		return tools.MissingField("minute")
	}
	if r.Count < 0 || r.Count > maxCronOccurrences {
		return tools.InvalidField("count", []int{0, maxCronOccurrences}, r.Count)
	}
	return nil
}

// fields 返回五个字段。
//
// 给出预设时用预设值；否则用请求里的字段，缺省为 `*`。
func (r *CrontabGeneratorRequest) fields() ([]string, error) {
	if r.Preset != "" {
		return presetFields(r.Preset), nil
	}

	fields := []string{
		orDefault(r.Minute, "*"),
		orDefault(r.Hour, "*"),
		orDefault(r.DayOfMonth, "*"),
		orDefault(r.Month, "*"),
		orDefault(r.DayOfWeek, "*"),
	}
	if len(fields) != cronFieldCount {
		return nil, toolerr.Newf(toolerr.CodeInvalidCron, map[string]any{
			"reason": "malformed cron expression",
		})
	}
	return fields, nil
}

// orDefault 返回 value 或缺省值。
func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

// presetFields 返回预设对应的五个字段。
func presetFields(preset string) []string {
	switch preset {
	case PresetEveryFiveMinutes:
		return []string{"*/5", "*", "*", "*", "*"}
	case PresetHourly:
		return []string{"0", "*", "*", "*", "*"}
	case PresetDaily:
		return []string{"0", "0", "*", "*", "*"}
	case PresetWeekly:
		return []string{"0", "0", "*", "*", "0"}
	case PresetMonthly:
		return []string{"0", "0", "1", "*", "*"}
	case PresetYearly:
		return []string{"0", "0", "1", "1", "*"}
	default:
		return []string{"*", "*", "*", "*", "*"}
	}
}

// location 返回时区。
//
// 缺省 UTC；时区名非法返回 INVALID_TIME。
func (r *CrontabGeneratorRequest) location() (*time.Location, error) {
	if r.Timezone == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(r.Timezone)
	if err != nil {
		return nil, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
			"field": "timezone",
		})
	}
	return loc, nil
}

// cronSchedule 是解析后的定时规则。
type cronSchedule struct {
	// minute 是分钟取值集合。
	minute map[int]bool
	// hour 是小时取值集合。
	hour map[int]bool
	// dayOfMonth 是日取值集合。
	dayOfMonth map[int]bool
	// month 是月取值集合。
	month map[int]bool
	// dayOfWeek 是星期取值集合。
	dayOfWeek map[int]bool
	// dayOfMonthAny 表示日字段未限定。
	dayOfMonthAny bool
	// dayOfWeekAny 表示星期字段未限定。
	dayOfWeekAny bool
}

// next 返回下一个触发时刻。
func (s *cronSchedule) next(from time.Time) (time.Time, bool) {
	// 最多向后找 5 年：五年内没有任何匹配时视为无解。
	limit := from.AddDate(5, 0, 0)

	cursor := from.Truncate(time.Minute).Add(time.Minute)
	for cursor.Before(limit) {
		if !s.month[int(cursor.Month())] {
			cursor = time.Date(cursor.Year(), cursor.Month(), 1, 0, 0, 0, 0, cursor.Location()).
				AddDate(0, 1, 0)
			continue
		}
		if !s.matchesDay(cursor) {
			cursor = time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, cursor.Location()).
				AddDate(0, 0, 1)
			continue
		}
		if !s.hour[cursor.Hour()] {
			cursor = time.Date(cursor.Year(), cursor.Month(), cursor.Day(), cursor.Hour(), 0, 0, 0, cursor.Location()).
				Add(time.Hour)
			continue
		}
		if !s.minute[cursor.Minute()] {
			cursor = cursor.Add(time.Minute)
			continue
		}
		return cursor, true
	}
	return time.Time{}, false
}

// matchesDay 判断日期是否满足日与星期的约束。
//
// 日与星期都限定时按 crontab 的传统语义：**两者取并集**。
func (s *cronSchedule) matchesDay(t time.Time) bool {
	domMatch := s.dayOfMonth[t.Day()]
	dowMatch := s.dayOfWeek[int(t.Weekday())]

	switch {
	case s.dayOfMonthAny && s.dayOfWeekAny:
		return true
	case s.dayOfMonthAny:
		return dowMatch
	case s.dayOfWeekAny:
		return domMatch
	default:
		return domMatch || dowMatch
	}
}

// parseCron 解析五段式 crontab 表达式。
func parseCron(fields []string) (*cronSchedule, error) {
	if len(fields) != cronFieldCount {
		return nil, toolerr.Newf(toolerr.CodeInvalidCron, map[string]any{
			"reason": "malformed cron expression",
		})
	}

	minute, err := parseCronField(fields[0], 0, 59)
	if err != nil {
		return nil, err
	}
	hour, err := parseCronField(fields[1], 0, 23)
	if err != nil {
		return nil, err
	}
	dom, err := parseCronField(fields[2], 1, 31)
	if err != nil {
		return nil, err
	}
	month, err := parseCronField(fields[3], 1, 12)
	if err != nil {
		return nil, err
	}
	dow, err := parseCronField(fields[4], 0, 7)
	if err != nil {
		return nil, err
	}

	// 星期字段的 7 与 0 都表示周日。
	if dow[7] {
		dow[0] = true
	}

	return &cronSchedule{
		minute:        minute,
		hour:          hour,
		dayOfMonth:    dom,
		month:         month,
		dayOfWeek:     dow,
		dayOfMonthAny: fields[2] == "*",
		dayOfWeekAny:  fields[4] == "*",
	}, nil
}

// parseCronField 解析一个 crontab 字段。
//
// 支持 `*`、`a`、`a-b`、`*/n`、`a-b/n` 与逗号列表。
func parseCronField(field string, min, max int) (map[int]bool, error) {
	out := map[int]bool{}

	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, invalidCron()
		}

		step := 1
		if idx := strings.Index(part, "/"); idx >= 0 {
			parsed, err := strconv.Atoi(part[idx+1:])
			if err != nil || parsed <= 0 {
				return nil, invalidCron()
			}
			step = parsed
			part = part[:idx]
		}

		start, end := min, max
		switch {
		case part == "*":
		case strings.Contains(part, "-"):
			bounds := strings.SplitN(part, "-", 2)
			low, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, invalidCron()
			}
			high, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, invalidCron()
			}
			start, end = low, high
		default:
			value, err := strconv.Atoi(part)
			if err != nil {
				return nil, invalidCron()
			}
			start, end = value, value
		}

		if start < min || end > max || start > end {
			return nil, invalidCron()
		}

		for v := start; v <= end; v += step {
			out[v] = true
		}
	}

	if len(out) == 0 {
		return nil, invalidCron()
	}
	return out, nil
}

// invalidCron 构造 crontab 表达式非法的错误。
func invalidCron() error {
	return toolerr.Newf(toolerr.CodeInvalidCron, map[string]any{
		"reason": "invalid cron expression",
	})
}

// describeCron 生成表达式的人类可读描述。
func describeCron(fields []string) string {
	parts := []string{}

	if fields[0] == "*" {
		parts = append(parts, "每分钟")
	} else {
		parts = append(parts, "第 "+fields[0]+" 分钟")
	}
	if fields[1] == "*" {
		parts = append(parts, "每小时")
	} else {
		parts = append(parts, "第 "+fields[1]+" 小时")
	}
	if fields[2] != "*" {
		parts = append(parts, "第 "+fields[2]+" 日")
	}
	if fields[3] != "*" {
		parts = append(parts, "第 "+fields[3]+" 月")
	}
	if fields[4] != "*" {
		parts = append(parts, "星期 "+fields[4])
	}

	return strings.Join(parts, "，")
}
