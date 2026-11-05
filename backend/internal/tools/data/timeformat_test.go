package data

import (
	"context"
	"errors"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// timeFormats 执行一次时间格式查询并取出结果。
func timeFormats(t *testing.T, req *TimeFormatRequest) TimeFormatResult {
	t.Helper()
	result, err := TimeFormats(context.Background(), req)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	out, ok := result.Data.(TimeFormatResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

func TestTimeFormatsByKey(t *testing.T) {
	t.Parallel()

	out := timeFormats(t, &TimeFormatRequest{Query: "unix-ms"})
	if !out.Matched {
		t.Fatal("unix-ms 应当命中")
	}
	if out.Kind != TimeFormatQueryKeyword {
		t.Errorf("kind = %s，期望 %s", out.Kind, TimeFormatQueryKeyword)
	}
	if out.Key != "unix-ms" {
		t.Errorf("key = %s，期望 unix-ms", out.Key)
	}
	if out.Category != TimeFormatCategoryEpoch {
		t.Errorf("category = %s，期望 %s", out.Category, TimeFormatCategoryEpoch)
	}
	if out.Example == "" {
		t.Error("应给出示例值")
	}
}

func TestTimeFormatsAllEntries(t *testing.T) {
	t.Parallel()

	out := timeFormats(t, &TimeFormatRequest{})
	if out.Kind != TimeFormatQueryAll {
		t.Errorf("kind = %s，期望 %s", out.Kind, TimeFormatQueryAll)
	}
	if out.Total != out.DatasetTotal || out.Total == 0 {
		t.Fatalf("total = %d，datasetTotal = %d，两者应相等且非空", out.Total, out.DatasetTotal)
	}
	for _, entry := range out.Matches {
		if entry.Category == "" {
			t.Fatalf("格式 %s 缺少分类", entry.Key)
		}
		if entry.Example == "" {
			t.Fatalf("格式 %s 缺少示例值", entry.Key)
		}
	}
}

func TestTimeFormatsCategoryFilter(t *testing.T) {
	t.Parallel()

	out := timeFormats(t, &TimeFormatRequest{Category: TimeFormatCategoryStandard})
	if out.Total == 0 {
		t.Fatal("标准日期格式类别不应为空")
	}
	for _, entry := range out.Matches {
		if entry.Category != TimeFormatCategoryStandard {
			t.Fatalf("格式 %s 的分类 = %s，期望 %s", entry.Key, entry.Category, TimeFormatCategoryStandard)
		}
	}
}

func TestTimeFormatsInvalidCategory(t *testing.T) {
	t.Parallel()

	_, err := TimeFormats(context.Background(), &TimeFormatRequest{Category: "bogus"})
	if err == nil {
		t.Fatal("非法分类应当报错")
	}
	var toolErr *toolerr.Error
	if !errors.As(err, &toolErr) || toolErr.Code != toolerr.CodeInvalidField {
		t.Fatalf("错误码不符：%v", err)
	}
}

// TestTimeFormatExamplesAreConsistent 校验数据集里的示例值互相自洽。
//
// 所有示例都对应同一时刻 2026-10-29T12:00:00Z（Unix 秒 1793275200），
// 因此 Unix 各档、Excel 序列号、儒略日与日计数之间必须能对上。
func TestTimeFormatExamplesAreConsistent(t *testing.T) {
	t.Parallel()

	out := timeFormats(t, &TimeFormatRequest{})
	examples := map[string]string{}
	for _, entry := range out.Matches {
		examples[entry.Key] = entry.Example
	}

	const wantUnix = "1793275200"
	if examples["unix"] != wantUnix {
		t.Errorf("unix 示例 = %s，期望 %s", examples["unix"], wantUnix)
	}
	if examples["unix-ms"] != wantUnix+"000" {
		t.Errorf("unix-ms 示例 = %s，与 unix 不自洽", examples["unix-ms"])
	}
	if examples["unix-us"] != wantUnix+"000000" {
		t.Errorf("unix-us 示例 = %s，与 unix 不自洽", examples["unix-us"])
	}
	if examples["unix-ns"] != wantUnix+"000000000" {
		t.Errorf("unix-ns 示例 = %s，与 unix 不自洽", examples["unix-ns"])
	}
	// 2026-10-29 距 1970-01-01 共 20755 天；Excel 纪元另加 25567 天。
	if examples["day-count"] != "20755" {
		t.Errorf("day-count 示例 = %s，期望 20755", examples["day-count"])
	}
	if examples["excel-serial"] != "46322.5" {
		t.Errorf("excel-serial 示例 = %s，期望 46322.5", examples["excel-serial"])
	}
	// 儒略日从正午起算：20755.5 + 2440587.5 = 2461343。
	if examples["julian"] != "2461343" {
		t.Errorf("julian 示例 = %s，期望 2461343", examples["julian"])
	}
}
