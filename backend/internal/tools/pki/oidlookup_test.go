package pki

import (
	"context"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// oidLookup 执行一次 OID 查询并取出结果。
func oidLookup(t *testing.T, req *OIDLookupRequest) OIDLookupResult {
	t.Helper()
	result, err := OIDLookup(context.Background(), req)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	out, ok := result.Data.(OIDLookupResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

// oidLookupErr 执行一次预期失败的 OID 查询。
func oidLookupErr(t *testing.T, req *OIDLookupRequest) *toolerr.Error {
	t.Helper()
	_, err := OIDLookup(context.Background(), req)
	if err == nil {
		t.Fatal("期望失败，实际成功")
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型不符：%T", err)
	}
	return toolErr
}

func TestOIDLookupRoot(t *testing.T) {
	t.Parallel()

	out := oidLookup(t, &OIDLookupRequest{})
	if out.Mode != OIDModeRoot {
		t.Errorf("mode = %s，期望 %s", out.Mode, OIDModeRoot)
	}
	if out.Total != 3 {
		t.Fatalf("total = %d，期望 3", out.Total)
	}
	if len(out.Children) != 3 {
		t.Fatalf("根节点数 = %d，期望 3", len(out.Children))
	}
	for i, want := range []string{"0", "1", "2"} {
		if out.Children[i].OID != want {
			t.Errorf("第 %d 个根节点 = %s，期望 %s", i, out.Children[i].OID, want)
		}
		if out.Children[i].Kind != "root" {
			t.Errorf("%s 的 kind = %s，期望 root", want, out.Children[i].Kind)
		}
	}
}

func TestOIDLookupExactNode(t *testing.T) {
	t.Parallel()

	// 用户示例：1.3.6.1.5.5.7.3.36 应为 id-kp-documentSigning。
	out := oidLookup(t, &OIDLookupRequest{OID: "1.3.6.1.5.5.7.3.36"})
	if !out.Matched {
		t.Fatal("应命中 1.3.6.1.5.5.7.3.36")
	}
	if out.Node == nil {
		t.Fatal("命中的节点不应为空")
	}
	if out.Node.Name != "id-kp-documentSigning" {
		t.Errorf("name = %s，期望 id-kp-documentSigning", out.Node.Name)
	}
	if out.Node.Arc != "36" {
		t.Errorf("arc = %s，期望 36", out.Node.Arc)
	}
	if out.Node.NameZh == "" || out.Node.DescriptionZh == "" {
		t.Error("命中节点应带中文名与中文说明")
	}
	if !strings.Contains(out.Node.Source, "9336") {
		t.Errorf("source = %s，期望包含 RFC 9336", out.Node.Source)
	}

	// 祖先链应从根一路到父节点。
	if len(out.Path) == 0 {
		t.Fatal("祖先链不应为空")
	}
	if out.Path[0].OID != "1" {
		t.Errorf("祖先链首节点 = %s，期望 1", out.Path[0].OID)
	}
	last := out.Path[len(out.Path)-1]
	if last.OID != "1.3.6.1.5.5.7.3" {
		t.Errorf("祖先链末节点 = %s，期望 1.3.6.1.5.5.7.3", last.OID)
	}
	if _, ok := out.PathChildren["1.3.6.1.5.5.7.3"]; !ok {
		t.Error("pathChildren 应包含父节点的子节点")
	}
}

func TestOIDLookupX509Extension(t *testing.T) {
	t.Parallel()

	out := oidLookup(t, &OIDLookupRequest{OID: "2.5.29.17"})
	if !out.Matched {
		t.Fatal("应命中 2.5.29.17")
	}
	if out.Node.Name != "subjectAltName" {
		t.Errorf("name = %s，期望 subjectAltName", out.Node.Name)
	}
	if out.Node.NameZh == "" {
		t.Error("应带中文名")
	}
}

func TestOIDLookupChildrenSortedAndPaginated(t *testing.T) {
	t.Parallel()

	// 1.3.6.1.5.5.7.3 下是扩展密钥用途，弧号应按数值升序。
	out := oidLookup(t, &OIDLookupRequest{OID: "1.3.6.1.5.5.7.3"})
	if out.Total < 40 {
		t.Fatalf("扩展密钥用途数量 = %d，期望至少 40", out.Total)
	}
	for i := 1; i < len(out.Children); i++ {
		prev, cur := out.Children[i-1], out.Children[i]
		prevNum, prevOK := oidArcNumber(prev.OID)
		curNum, curOK := oidArcNumber(cur.OID)
		if prevOK && curOK && prevNum > curNum {
			t.Fatalf("子节点未按弧号数值排序：%s 在 %s 之前", prev.OID, cur.OID)
		}
	}

	// 分页：取第一页后应还有下一页，且两页不重叠。
	page1 := oidLookup(t, &OIDLookupRequest{OID: "1.3.6.1.5.5.7.3", Limit: 5})
	if len(page1.Children) != 5 {
		t.Fatalf("第一页条数 = %d，期望 5", len(page1.Children))
	}
	if !page1.HasMore {
		t.Error("第一页后应还有下一页")
	}
	page2 := oidLookup(t, &OIDLookupRequest{OID: "1.3.6.1.5.5.7.3", Limit: 5, Offset: 5})
	if len(page2.Children) != 5 {
		t.Fatalf("第二页条数 = %d，期望 5", len(page2.Children))
	}
	if page2.Children[0].OID == page1.Children[0].OID {
		t.Error("两页内容重叠")
	}
}

func TestOIDLookupUnknownOIDReturnsNearestAncestor(t *testing.T) {
	t.Parallel()

	// 语法合法但未收录：不是错误，应给出最近祖先。
	out := oidLookup(t, &OIDLookupRequest{OID: "1.3.6.1.5.5.7.3.9999"})
	if out.Matched {
		t.Error("未收录的 OID 不应命中")
	}
	if out.Node != nil {
		t.Error("未命中时 node 应为空")
	}
	if out.NearestOid != "1.3.6.1.5.5.7.3" {
		t.Errorf("nearestOid = %s，期望 1.3.6.1.5.5.7.3", out.NearestOid)
	}
	if out.MissingSuffix != ".9999" {
		t.Errorf("missingSuffix = %s，期望 .9999", out.MissingSuffix)
	}
	// 未命中也要把最近祖先的分支铺进树，前端才能点击祖先继续浏览。
	if len(out.Path) == 0 || out.Path[len(out.Path)-1].OID != out.NearestOid {
		t.Errorf("path 应以最近祖先结尾，实际长度 = %d", len(out.Path))
	}
	if _, ok := out.PathChildren[out.NearestOid]; !ok {
		t.Error("pathChildren 应包含最近祖先的子节点")
	}
}

func TestOIDLookupSearch(t *testing.T) {
	t.Parallel()

	// 按英文名搜索。
	out := oidLookup(t, &OIDLookupRequest{Query: "subjectAltName"})
	if out.Mode != OIDModeSearch {
		t.Errorf("mode = %s，期望 %s", out.Mode, OIDModeSearch)
	}
	if out.Total == 0 {
		t.Fatal("subjectAltName 应有命中")
	}
	if out.Matches[0].Node.OID != "2.5.29.17" {
		t.Errorf("首条命中 = %s，期望 2.5.29.17", out.Matches[0].Node.OID)
	}

	// 按中文名搜索。
	out = oidLookup(t, &OIDLookupRequest{Query: "主体备用名称"})
	if out.Total == 0 {
		t.Fatal("中文名应有命中")
	}

	// 按 OID 搜索。
	out = oidLookup(t, &OIDLookupRequest{Query: "1.2.840.113549.1.1.11"})
	if out.Total == 0 {
		t.Fatal("按 OID 搜索应有命中")
	}
	if out.Matches[0].Node.Name != "sha256WithRSAEncryption" {
		t.Errorf("首条命中 = %s，期望 sha256WithRSAEncryption", out.Matches[0].Node.Name)
	}
}

func TestOIDLookupSearchRanksExactFirst(t *testing.T) {
	t.Parallel()

	out := oidLookup(t, &OIDLookupRequest{Query: "sm2"})
	if out.Total == 0 {
		t.Fatal("sm2 应有命中")
	}
	// 精确的英文名应排在子串命中之前。
	found := false
	for _, match := range out.Matches {
		if match.Node.Name == "sm2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("应命中名称恰为 sm2 的节点")
	}
}

func TestOIDLookupNormalizesInput(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"1.3.6.1.5.5.7.3.36",
		" 1.3.6.1.5.5.7.3.36 ",
		"OID:1.3.6.1.5.5.7.3.36",
		".1.3.6.1.5.5.7.3.36.",
	} {
		out := oidLookup(t, &OIDLookupRequest{OID: input})
		if !out.Matched {
			t.Errorf("%s 应命中", input)
			continue
		}
		if out.OID != "1.3.6.1.5.5.7.3.36" {
			t.Errorf("%s → oid = %s，期望规范化后的 1.3.6.1.5.5.7.3.36", input, out.OID)
		}
	}
}

func TestOIDLookupValidate(t *testing.T) {
	t.Parallel()

	// oid 与 query 互斥。
	if code := oidLookupErr(t, &OIDLookupRequest{OID: "1.2.3", Query: "x"}).Code; code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", code, toolerr.CodeInvalidField)
	}
	// 非法 OID 语法。
	if code := oidLookupErr(t, &OIDLookupRequest{OID: "1.2.3.x"}).Code; code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", code, toolerr.CodeInvalidField)
	}
	// OID 太长。
	long := strings.Repeat("1.", OIDMaxLength) + "1"
	if code := oidLookupErr(t, &OIDLookupRequest{OID: long}).Code; code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", code, toolerr.CodeInvalidField)
	}
	// 弧数超限。
	manyArcs := strings.TrimSuffix(strings.Repeat("1.", OIDMaxArcs+1), ".")
	if code := oidLookupErr(t, &OIDLookupRequest{OID: manyArcs}).Code; code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", code, toolerr.CodeInvalidField)
	}
	// 搜索词太长。
	if code := oidLookupErr(t, &OIDLookupRequest{Query: strings.Repeat("x", OIDMaxQueryLength+1)}).Code; code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", code, toolerr.CodeInvalidField)
	}
	// 负偏移。
	if code := oidLookupErr(t, &OIDLookupRequest{Offset: -1}).Code; code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", code, toolerr.CodeInvalidField)
	}
	// limit 超上限。
	if code := oidLookupErr(t, &OIDLookupRequest{Limit: OIDMaxLimit + 1}).Code; code != toolerr.CodeInvalidField {
		t.Errorf("错误码 = %s，期望 %s", code, toolerr.CodeInvalidField)
	}
}

func TestOIDLookupErrorDoesNotEchoInput(t *testing.T) {
	t.Parallel()

	err := oidLookupErr(t, &OIDLookupRequest{OID: "1.2.3.secret-value"})
	msg := err.Error()
	if strings.Contains(msg, "secret-value") {
		t.Errorf("错误信息回显了用户输入：%s", msg)
	}
	for key, value := range err.Params {
		if text, ok := value.(string); ok && strings.Contains(text, "secret-value") {
			t.Errorf("错误参数 %s 回显了用户输入：%s", key, text)
		}
	}
}

func TestOIDDatasetIntegrity(t *testing.T) {
	t.Parallel()

	loadOIDDataset()

	for oid, node := range oidNodes {
		if node.Name == "" {
			t.Errorf("%s 缺少英文名", oid)
		}
		if node.NameZh == "" {
			t.Errorf("%s 缺少中文名", oid)
		}
		if node.DescriptionZh == "" {
			t.Errorf("%s 缺少中文说明", oid)
		}
		if node.Kind == "" {
			t.Errorf("%s 缺少 kind", oid)
		}
		parent := oidParent(oid)
		if parent != "" {
			if _, ok := oidNodes[parent]; !ok {
				t.Errorf("%s 的父节点 %s 不存在", oid, parent)
			}
		}
	}

	for _, root := range []string{"0", "1", "2"} {
		if _, ok := oidNodes[root]; !ok {
			t.Errorf("缺少根节点 %s", root)
		}
	}
}

func TestOIDDatasetChineseCoverage(t *testing.T) {
	t.Parallel()

	loadOIDDataset()

	missing := 0
	for oid, node := range oidNodes {
		if node.NameZh == "" || node.DescriptionZh == "" {
			missing++
			t.Errorf("%s 缺少中文内容", oid)
		}
	}
	if missing > 0 {
		t.Fatalf("共有 %d 个节点缺少中文内容", missing)
	}
}
