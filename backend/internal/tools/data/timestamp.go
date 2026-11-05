package data

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 时间戳单位。
const (
	// UnitSecond 是秒。
	UnitSecond = "s"
	// UnitMillisecond 是毫秒。
	UnitMillisecond = "ms"
	// UnitMicrosecond 是微秒。
	UnitMicrosecond = "us"
	// UnitNanosecond 是纳秒。
	UnitNanosecond = "ns"
)

// TimestampUnits 返回受支持的单位（顺序稳定）。
func TimestampUnits() []string {
	return []string{UnitSecond, UnitMillisecond, UnitMicrosecond, UnitNanosecond}
}

// 操作。
const (
	// OpToDate 把时间戳转成日期。
	OpToDate = "toDate"
	// OpToTimestamp 把日期转成时间戳。
	OpToTimestamp = "toTimestamp"
)

// TimestampOperations 返回受支持的操作（顺序稳定）。
func TimestampOperations() []string {
	return []string{OpToDate, OpToTimestamp}
}

// TimestampRequest 是 timestamp 工具的请求。
type TimestampRequest struct {
	// Value 是待转换的值：时间戳（toDate）或日期文本（toTimestamp）。
	//
	// 允许写成字符串或数字；缺省时取当前时间。
	Value Scalar `json:"value"`
	// Unit 是时间戳单位；缺省时按位数自动识别。
	Unit string `json:"unit"`
	// Timezone 是 IANA 时区名；缺省用 UTC。
	Timezone string `json:"timezone"`
	// Operation 是转换方向；缺省 toDate。
	Operation string `json:"operation"`
}

// TimestampPair 是一对时间戳。
type TimestampPair struct {
	// S 是秒级时间戳。
	S int64 `json:"s"`
	// Ms 是毫秒级时间戳。
	Ms int64 `json:"ms"`
}

// TimestampFormat 是一个日期时间格式条目。
//
// 列表里的每一项都对应同一次转换的**不同写法**，
// 供前端逐行展示与复制。
type TimestampFormat struct {
	// Key 是格式标识；前端据此取词条展示名称。
	Key string `json:"key"`
	// Value 是该格式下的时间文本。
	Value string `json:"value"`
}

// TimestampResult 是 timestamp 工具的响应。
type TimestampResult struct {
	// Timestamp 是秒与毫秒时间戳。
	Timestamp TimestampPair `json:"timestamp"`
	// ISO8601 是 ISO 8601 格式的本地时间。
	ISO8601 string `json:"iso8601"`
	// RFC3339 是 RFC 3339 格式的本地时间。
	RFC3339 string `json:"rfc3339"`
	// Local 是指定时区的本地时间。
	Local string `json:"local"`
	// UTC 是 UTC 时间。
	UTC string `json:"utc"`
	// Formats 是各常见格式的时间文本；顺序固定。
	Formats []TimestampFormat `json:"formats"`
}

// Timestamp 转换时间戳与日期。
//
// 服务端**不**猜时区：未指定时区一律按 UTC 处理，
// 并给出 TIMEZONE_ASSUMED 警告。
func Timestamp(ctx context.Context, req *TimestampRequest) (TimestampResult, error) {
	if err := req.Validate(); err != nil {
		return TimestampResult{}, err
	}

	loc, err := req.location()
	if err != nil {
		return TimestampResult{}, err
	}

	var moment time.Time
	switch req.operation() {
	case OpToTimestamp:
		moment, err = req.parseDate()
		if err != nil {
			return TimestampResult{}, err
		}
	default:
		moment, err = req.parseTimestamp()
		if err != nil {
			return TimestampResult{}, err
		}
	}

	local := moment.In(loc)
	utc := moment.UTC()

	return TimestampResult{
		Timestamp: TimestampPair{
			S:  moment.Unix(),
			Ms: moment.UnixMilli(),
		},
		ISO8601: local.Format("2006-01-02T15:04:05.000Z07:00"),
		RFC3339: local.Format(time.RFC3339),
		Local:   local.Format("2006-01-02 15:04:05"),
		UTC:     utc.Format("2006-01-02 15:04:05"),
		Formats: buildTimestampFormats(local, utc),
	}, nil
}

// 各纪元与格式的常量。
const (
	// epochExcelOffset 是 Excel 纪元（1899-12-30）与 Unix 纪元的秒差。
	epochExcelOffset = 25567 * 86400
	// epochDotNetOffset 是 .NET 纪元（0001-01-01）与 Unix 纪元的秒差。
	epochDotNetOffset = 62135596800
	// epochWin32Offset 是 Win32 纪元（1601-01-01）与 Unix 纪元的秒差。
	epochWin32Offset = 11644473600
	// julianEpochOffset 是儒略日与 Unix 纪元的日差（儒略日从正午起算）。
	julianEpochOffset = 2440587.5
	// secondsPerDay 是一天的秒数。
	secondsPerDay = 86400
)

// buildTimestampFormats 生成各常见格式的时间文本。
//
// 顺序固定：先四种时间戳，再按「带偏移 → UTC → 无时区 → 其他纪年」排列，
// 便于前端直接按序渲染列表。
//
// 只做**表示层**换算，不涉及任何密码学运算。
func buildTimestampFormats(local, utc time.Time) []TimestampFormat {
	seconds := local.Unix()
	nanos := local.UnixNano()

	return []TimestampFormat{
		{Key: "unix", Value: strconv.FormatInt(seconds, 10)},
		{Key: "unixMs", Value: strconv.FormatInt(local.UnixMilli(), 10)},
		{Key: "unixUs", Value: strconv.FormatInt(local.UnixMicro(), 10)},
		{Key: "unixNs", Value: strconv.FormatInt(nanos, 10)},
		{Key: "iso8601", Value: local.Format("2006-01-02T15:04:05.000Z07:00")},
		{Key: "iso8601Utc", Value: utc.Format("2006-01-02T15:04:05.000Z")},
		{Key: "rfc3339", Value: local.Format(time.RFC3339)},
		{Key: "rfc7231", Value: utc.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		{Key: "rfc2822", Value: local.Format(time.RFC1123Z)},
		{Key: "iso9075", Value: local.Format("2006-01-02 15:04:05")},
		{Key: "jsDateString", Value: local.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")},
		{Key: "jsDate", Value: formatJSDate(local)},
		{Key: "excelSerial", Value: formatExcelSerial(seconds)},
		{Key: "dotnetTicks", Value: strconv.FormatInt(epochTicks(nanos, epochDotNetOffset), 10)},
		{Key: "win32FileTime", Value: strconv.FormatInt(epochTicks(nanos, epochWin32Offset), 10)},
		{Key: "ldapYmd", Value: utc.Format("20060102150405") + "Z"},
		{Key: "julianDay", Value: strconv.FormatFloat(float64(seconds)/secondsPerDay+julianEpochOffset, 'f', -1, 64)},
		{Key: "dayCount", Value: strconv.FormatInt(secFloorDiv(seconds, secondsPerDay), 10)},
	}
}

// formatJSDate 生成 JS `new Date(...)` 构造式。
//
// JS 的月份从 0 开始，因此这里减一。
func formatJSDate(t time.Time) string {
	return fmt.Sprintf(
		"new Date(%d, %d, %d, %d, %d, %d, %d);",
		t.Year(),
		int(t.Month())-1,
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond()/int(time.Millisecond),
	)
}

// formatExcelSerial 生成 Excel 日期序列号。
//
// Excel 以 1899-12-30 为起点，并保留历史上的 1900 闰年缺陷，
// 因此偏移量取 25567 天。
func formatExcelSerial(seconds int64) string {
	days := float64(seconds+epochExcelOffset) / secondsPerDay
	return strconv.FormatFloat(days, 'f', -1, 64)
}

// epochTicks 把 Unix 纳秒换算为「自指定纪元起的 100 纳秒计数」。
//
// 全程用整数运算：浮点会丢失低位精度。
func epochTicks(nanos int64, offset int64) int64 {
	seconds := nanos / int64(time.Second)
	remainder := nanos % int64(time.Second)
	return (seconds+offset)*1e7 + remainder/100
}

// secFloorDiv 是向下取整的整除（Go 的整数除法向零取整，负数需修正）。
func secFloorDiv(value, divisor int64) int64 {
	quotient := value / divisor
	if value%divisor != 0 && value < 0 {
		quotient--
	}
	return quotient
}

// Validate 做字段级校验。
func (r *TimestampRequest) Validate() error {
	switch r.Operation {
	case "", OpToDate, OpToTimestamp:
	default:
		return tools.InvalidField("operation", TimestampOperations(), r.Operation)
	}
	if r.Unit != "" {
		switch r.Unit {
		case UnitSecond, UnitMillisecond, UnitMicrosecond, UnitNanosecond:
		default:
			return tools.InvalidField("unit", TimestampUnits(), r.Unit)
		}
	}
	if r.Timezone != "" {
		if _, err := time.LoadLocation(r.Timezone); err != nil {
			return toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
				"field": "timezone",
			})
		}
	}
	return nil
}

// operation 返回操作；缺省 toDate。
func (r *TimestampRequest) operation() string {
	if r.Operation == "" {
		return OpToDate
	}
	return r.Operation
}

// location 返回时区；缺省 UTC。
func (r *TimestampRequest) location() (*time.Location, error) {
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

// parseTimestamp 解析时间戳。
//
// 未指定单位时按位数识别：10 位按秒、13 位按毫秒，
// 其余按秒处理（负数与更长的值同样按秒）。
func (r *TimestampRequest) parseTimestamp() (time.Time, error) {
	raw := strings.TrimSpace(r.Value.String())
	if raw == "" {
		return time.Now().UTC(), nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
			"field": "value",
		})
	}

	unit := r.Unit
	if unit == "" {
		unit = guessUnit(raw)
	}

	switch unit {
	case UnitMillisecond:
		return time.UnixMilli(value).UTC(), nil
	case UnitMicrosecond:
		return time.UnixMicro(value).UTC(), nil
	case UnitNanosecond:
		return time.Unix(0, value).UTC(), nil
	default:
		return time.Unix(value, 0).UTC(), nil
	}
}

// guessUnit 按位数猜测单位。
func guessUnit(raw string) string {
	digits := strings.TrimPrefix(raw, "-")
	switch {
	case len(digits) == 13:
		return UnitMillisecond
	default:
		return UnitSecond
	}
}

// parseDate 解析日期文本。
//
// 支持 RFC 3339 与 `YYYY-MM-DD`。
func (r *TimestampRequest) parseDate() (time.Time, error) {
	raw := strings.TrimSpace(r.Value.String())
	if raw == "" {
		return time.Now().UTC(), nil
	}

	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t.UTC(), nil
	}

	return time.Time{}, toolerr.Newf(toolerr.CodeInvalidTime, map[string]any{
		"field": "value",
	})
}
