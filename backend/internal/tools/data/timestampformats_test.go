package data

import (
	"context"
	"testing"
)

// TestTimestampFormats 断言多格式列表的取值。
//
// 参考时刻：2026-10-29T12:00:00Z（Unix 秒 1793275200）。
func TestTimestampFormats(t *testing.T) {
	result, err := Timestamp(context.Background(), &TimestampRequest{
		Value: Scalar{Value: "1793275200", Set: true},
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}

	formats := map[string]string{}
	for _, item := range result.Formats {
		formats[item.Key] = item.Value
	}

	cases := []struct {
		key  string
		want string
	}{
		{"unix", "1793275200"},
		{"unixMs", "1793275200000"},
		{"unixUs", "1793275200000000"},
		{"unixNs", "1793275200000000000"},
		{"iso8601Utc", "2026-10-29T12:00:00.000Z"},
		{"rfc3339", "2026-10-29T12:00:00Z"},
		{"rfc7231", "Thu, 29 Oct 2026 12:00:00 GMT"},
		{"iso9075", "2026-10-29 12:00:00"},
		{"ldapYmd", "20261029120000Z"},
		{"excelSerial", "46322.5"},
		{"julianDay", "2461343"},
		{"dayCount", "20755"},
	}
	for _, c := range cases {
		got, ok := formats[c.key]
		if !ok {
			t.Errorf("格式 %s 缺失", c.key)
			continue
		}
		if got != c.want {
			t.Errorf("格式 %s = %q，期望 %q", c.key, got, c.want)
		}
	}

	// 格式清单的键集合固定：前端按顺序逐行渲染，多一项少一项都要能看出来。
	if len(result.Formats) != len(cases)+6 {
		t.Errorf("格式条目数 = %d，期望 %d", len(result.Formats), len(cases)+6)
	}
}

// TestTimestampFormatsWithTimezone 断言带时区偏移的格式跟随所选时区。
func TestTimestampFormatsWithTimezone(t *testing.T) {
	result, err := Timestamp(context.Background(), &TimestampRequest{
		Value:    Scalar{Value: "1793275200", Set: true},
		Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}

	formats := map[string]string{}
	for _, item := range result.Formats {
		formats[item.Key] = item.Value
	}

	if formats["iso8601"] != "2026-10-29T20:00:00.000+08:00" {
		t.Errorf("带偏移的 ISO 8601 = %q", formats["iso8601"])
	}
	if formats["iso8601Utc"] != "2026-10-29T12:00:00.000Z" {
		t.Errorf("UTC 的 ISO 8601 = %q", formats["iso8601Utc"])
	}
	if formats["rfc7231"] != "Thu, 29 Oct 2026 12:00:00 GMT" {
		t.Errorf("HTTP 日期 = %q", formats["rfc7231"])
	}
}
