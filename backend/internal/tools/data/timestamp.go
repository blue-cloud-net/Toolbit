package data

import (
	"context"
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
	}, nil
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
