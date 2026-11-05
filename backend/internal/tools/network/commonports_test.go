package network

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// commonPorts 执行一次端口查询并取出结果。
func commonPorts(t *testing.T, req *CommonPortRequest) CommonPortResult {
	t.Helper()
	result, err := CommonPorts(context.Background(), req)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	out, ok := result.Data.(CommonPortResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

func TestCommonPortsByNumber(t *testing.T) {
	t.Parallel()

	out := commonPorts(t, &CommonPortRequest{Query: "443"})
	if !out.Matched {
		t.Fatal("443 应当命中")
	}
	if out.Kind != PortQueryPort {
		t.Errorf("kind = %s，期望 %s", out.Kind, PortQueryPort)
	}
	if out.Port != 443 {
		t.Errorf("port = %d，期望 443", out.Port)
	}
	if out.Label != "HTTPS" {
		t.Errorf("label = %s，期望 HTTPS", out.Label)
	}
	if len(out.Protocols) == 0 {
		t.Error("应给出传输协议")
	}
	if out.Category == "" {
		t.Error("应给出服务分类")
	}
}

func TestCommonPortsByKeyword(t *testing.T) {
	t.Parallel()

	out := commonPorts(t, &CommonPortRequest{Query: "ssh"})
	if out.Kind != PortQueryKeyword {
		t.Errorf("kind = %s，期望 %s", out.Kind, PortQueryKeyword)
	}
	if !out.Matched || out.Port != 22 {
		t.Fatalf("ssh 应当命中 22，实际 matched=%v port=%d", out.Matched, out.Port)
	}
}

func TestCommonPortsAllEntries(t *testing.T) {
	t.Parallel()

	out := commonPorts(t, &CommonPortRequest{})
	if out.Kind != PortQueryAll {
		t.Errorf("kind = %s，期望 %s", out.Kind, PortQueryAll)
	}
	if out.Total != out.DatasetTotal {
		t.Errorf("total = %d，datasetTotal = %d，两者应相等", out.Total, out.DatasetTotal)
	}
	if out.Total == 0 {
		t.Fatal("数据集不应为空")
	}
	// 每条条目都要有分类，否则前端筛选项会漏。
	for _, entry := range out.Matches {
		if entry.Category == "" {
			t.Fatalf("端口 %d 缺少服务分类", entry.Port)
		}
		if entry.Description == "" {
			t.Fatalf("端口 %d 缺少中文说明", entry.Port)
		}
	}
}

func TestCommonPortsCategoryFilter(t *testing.T) {
	t.Parallel()

	out := commonPorts(t, &CommonPortRequest{Category: PortCategoryDatabase})
	if out.Total == 0 {
		t.Fatal("数据库分类不应为空")
	}
	for _, entry := range out.Matches {
		if entry.Category != PortCategoryDatabase {
			t.Fatalf("端口 %d 的分类 = %s，期望 %s", entry.Port, entry.Category, PortCategoryDatabase)
		}
	}
}

func TestCommonPortsUnknownPort(t *testing.T) {
	t.Parallel()

	out := commonPorts(t, &CommonPortRequest{Query: "65535"})
	if out.Matched {
		t.Error("未收录的端口不应命中")
	}
	if out.Total != 0 {
		t.Errorf("total = %d，期望 0", out.Total)
	}
}

func TestCommonPortsInvalidCategory(t *testing.T) {
	t.Parallel()

	_, err := CommonPorts(context.Background(), &CommonPortRequest{Category: "bogus"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
