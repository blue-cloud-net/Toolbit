package network

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestIPRangeExpand 断言展开网段。
func TestIPRangeExpand(t *testing.T) {
	result, err := IPRange(t.Context(), &IPRangeRequest{
		Mode:  IPRangeModeExpand,
		Cidr:  "192.168.1.0/30",
		Limit: 16,
	})
	data := result.Data.(IPRangeResult)
	if err != nil {
		t.Fatalf("展开失败：%v", err)
	}
	if data.Count != 4 {
		t.Errorf("地址数 = %d，期望 4", data.Count)
	}
	if len(data.Addresses) != 4 {
		t.Errorf("地址列表长度 = %d，期望 4", len(data.Addresses))
	}
}

// TestIPRangeToCIDR 断言地址区间转 CIDR。
func TestIPRangeToCIDR(t *testing.T) {
	result, err := IPRange(t.Context(), &IPRangeRequest{
		Mode:  IPRangeModeToCIDR,
		Start: "192.168.1.10",
		End:   "192.168.1.20",
	})
	data := result.Data.(IPRangeResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if len(data.Cidrs) == 0 {
		t.Fatal("应给出 CIDR 列表")
	}
	if data.Start != "192.168.1.10" || data.End != "192.168.1.20" {
		t.Errorf("区间 = %s ~ %s", data.Start, data.End)
	}
}

// TestIPRangeToRange 断言 CIDR 转地址区间。
func TestIPRangeToRange(t *testing.T) {
	result, err := IPRange(t.Context(), &IPRangeRequest{
		Mode: IPRangeModeToRange,
		Cidr: "192.168.1.0/24",
	})
	data := result.Data.(IPRangeResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.Start != "192.168.1.0" {
		t.Errorf("起始地址 = %q", data.Start)
	}
}

// TestIPRangeReversedRange 断言倒置区间被拒。
func TestIPRangeReversedRange(t *testing.T) {
	_, err := IPRange(t.Context(), &IPRangeRequest{
		Mode:  IPRangeModeToCIDR,
		Start: "192.168.1.20",
		End:   "192.168.1.10",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
