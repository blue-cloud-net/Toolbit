package formatting

import (
	"strings"
	"testing"
)

func TestJSONFormatPretty(t *testing.T) {
	res, err := JSONFormat(&JSONFormatRequest{
		Input:  TextInput{Data: `{"b":1,"a":[1,2]}`},
		Indent: 2,
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !res.Valid {
		t.Fatal("应为合法")
	}
	want := "{\n  \"b\": 1,\n  \"a\": [\n    1,\n    2\n  ]\n}"
	if res.Output != want {
		t.Errorf("输出 =\n%s\n期望 =\n%s", res.Output, want)
	}
}

func TestJSONFormatSortKeys(t *testing.T) {
	res, err := JSONFormat(&JSONFormatRequest{
		Input:    TextInput{Data: `{"b":1,"a":2}`},
		SortKeys: true,
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !strings.Contains(res.Output, `"a": 2`) {
		t.Errorf("输出应含排序后的 a 键：%s", res.Output)
	}
	if strings.Index(res.Output, `"a"`) > strings.Index(res.Output, `"b"`) {
		t.Errorf("键序未排序：%s", res.Output)
	}
}

func TestJSONFormatMinify(t *testing.T) {
	res, err := JSONFormat(&JSONFormatRequest{
		Input:  TextInput{Data: "{\n  \"a\" : 1\n}"},
		Action: ActionMinify,
	})
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}
	if res.Output != `{"a":1}` {
		t.Errorf("输出 = %q，期望 {\"a\":1}", res.Output)
	}
}

func TestJSONFormatValidate(t *testing.T) {
	res, err := JSONFormat(&JSONFormatRequest{
		Input:  TextInput{Data: `{"a":1}`},
		Action: ActionValidate,
	})
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	if !res.Valid {
		t.Error("应为合法")
	}
	if res.Output != "" {
		t.Errorf("validate 不应产出内容，实际 %q", res.Output)
	}
}

func TestJSONFormatReportsSyntaxError(t *testing.T) {
	res, err := JSONFormat(&JSONFormatRequest{
		Input: TextInput{Data: "{\n  \"a\": 1,\n}"},
	})
	if err != nil {
		t.Fatalf("不应返回工具错误: %v", err)
	}
	if res.Valid {
		t.Fatal("应判为非法")
	}
	if res.Error == nil {
		t.Fatal("应给出错误位置")
	}
	if res.Error.Line == 0 {
		t.Errorf("应给出行号，实际 %+v", res.Error)
	}
}

func TestJSONFormatRejectsDuplicateKey(t *testing.T) {
	res, err := JSONFormat(&JSONFormatRequest{
		Input: TextInput{Data: `{"a":1,"a":2}`},
	})
	if err != nil {
		t.Fatalf("不应返回工具错误: %v", err)
	}
	if res.Valid {
		t.Error("重复键应判为非法")
	}
}

func TestYAMLFormatKeepsComments(t *testing.T) {
	res, err := YAMLFormat(&YAMLFormatRequest{
		Input: TextInput{Data: "# 注释\nkey: value\n"},
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !strings.Contains(res.Output, "# 注释") {
		t.Errorf("YAML 注释应被保留：%s", res.Output)
	}
}

func TestYAMLFormatRejectsAliasBomb(t *testing.T) {
	// 经典的 billion laughs：一个锚点被反复引用。
	bomb := `
a: &a ["x","x","x","x","x","x","x","x","x"]
b: &b [*a,*a,*a,*a,*a,*a,*a,*a,*a]
c: &c [*b,*b,*b,*b,*b,*b,*b,*b,*b]
d: &d [*c,*c,*c,*c,*c,*c,*c,*c,*c]
e: &e [*d,*d,*d,*d,*d,*d,*d,*d,*d]
f: &f [*e,*e,*e,*e,*e,*e,*e,*e,*e]
g: &g [*f,*f,*f,*f,*f,*f,*f,*f,*f]
h: &h [*g,*g,*g,*g,*g,*g,*g,*g,*g]
i: &i [*h,*h,*h,*h,*h,*h,*h,*h,*h]
`
	res, err := YAMLFormat(&YAMLFormatRequest{Input: TextInput{Data: bomb}})
	if err != nil {
		// 被上限拦下也是可接受的：不得静默展开。
		return
	}
	if res.Valid && len(res.Output) > 1<<20 {
		t.Errorf("别名展开未受限，输出长度 %d", len(res.Output))
	}
}

func TestTOMLFormatDetectsComments(t *testing.T) {
	res, err := TOMLFormat(&TOMLFormatRequest{
		Input: TextInput{Data: "# 注释\nkey = \"value\"\n"},
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !res.CommentsLost {
		t.Error("应标记注释丢失")
	}
	// go-toml 的编码器默认用 literal string（单引号）：
	// 只有内容含单引号或换行时才改用 basic string。
	if !strings.Contains(res.Output, "key = 'value'") {
		t.Errorf("输出 = %q", res.Output)
	}
}

func TestXMLFormatPretty(t *testing.T) {
	res, err := XMLFormat(&XMLFormatRequest{
		Input: TextInput{Data: `<root><a>1</a><b attr="x">2</b></root>`},
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	want := "<root>\n  <a>1</a>\n  <b attr=\"x\">2</b>\n</root>"
	if res.Output != want {
		t.Errorf("输出 =\n%s\n期望 =\n%s", res.Output, want)
	}
}

func TestHTMLFormatPretty(t *testing.T) {
	res, err := HTMLFormat(&HTMLFormatRequest{
		Input: TextInput{Data: "<div><p>hello</p></div>"},
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !strings.Contains(res.Output, "<p>hello</p>") {
		t.Errorf("输出 = %s", res.Output)
	}
}

func TestJSFormatPretty(t *testing.T) {
	res, err := JSFormat(&JSFormatRequest{
		Input:  TextInput{Data: "function f(a){return a+1}"},
		Indent: 2,
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !strings.Contains(res.Output, "return a + 1;") {
		t.Errorf("输出 = %s", res.Output)
	}
	if !strings.Contains(res.Output, "\n  ") {
		t.Errorf("缩进应为 2 空格：%s", res.Output)
	}
}

func TestJSFormatMinify(t *testing.T) {
	res, err := JSFormat(&JSFormatRequest{
		Input:  TextInput{Data: "function f(a) {\n  // 注释\n  return a + 1;\n}"},
		Action: ActionMinify,
	})
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}
	if strings.Contains(res.Output, "//") {
		t.Errorf("注释应被去掉：%s", res.Output)
	}
	if !strings.Contains(res.Output, "function f(a)") {
		t.Errorf("输出 = %s", res.Output)
	}
}

func TestJSFormatReportsSyntaxError(t *testing.T) {
	res, err := JSFormat(&JSFormatRequest{
		Input: TextInput{Data: "function f( {"},
	})
	if err != nil {
		t.Fatalf("不应返回工具错误: %v", err)
	}
	if res.Valid {
		t.Error("应判为非法")
	}
	if res.Error == nil {
		t.Fatal("应给出错误位置")
	}
}

func TestSQLFormatPretty(t *testing.T) {
	res, err := SQLFormat(&SQLFormatRequest{
		Input: TextInput{Data: "select id,name from users where id=1"},
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !strings.Contains(res.Output, "SELECT") && !strings.Contains(res.Output, "select") {
		t.Errorf("输出 = %s", res.Output)
	}
	if !strings.Contains(res.Output, "FROM") && !strings.Contains(res.Output, "from") {
		t.Errorf("子句应另起一行：%s", res.Output)
	}
}

// TestSQLFormatKeepsTokenSpacing 断言关键字与后续 token 之间保留空格。
//
// 回归用例：曾经把子句关键字后的 `lineStart` 置为 true，
// 导致 `SELECT` 与列名之间丢空格（输出 `SELECTid`）。
func TestSQLFormatKeepsTokenSpacing(t *testing.T) {
	res, err := SQLFormat(&SQLFormatRequest{
		Input: TextInput{Data: "SELECT id FROM users WHERE id = 1"},
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	for _, want := range []string{"SELECT id", "FROM users", "WHERE id = 1"} {
		if !strings.Contains(res.Output, want) {
			t.Errorf("输出缺少 %q：%s", want, res.Output)
		}
	}
	for _, bad := range []string{"SELECTid", "FROMusers", "WHEREid"} {
		if strings.Contains(res.Output, bad) {
			t.Errorf("关键字与后续 token 之间丢了空格（%q）：%s", bad, res.Output)
		}
	}
}

func TestSQLFormatKeywordCase(t *testing.T) {
	res, err := SQLFormat(&SQLFormatRequest{
		Input:       TextInput{Data: "select 1 from t"},
		KeywordCase: SQLKeywordUpper,
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !strings.Contains(res.Output, "SELECT") {
		t.Errorf("关键字应转大写：%s", res.Output)
	}
}

func TestSQLFormatKeepsStringContent(t *testing.T) {
	res, err := SQLFormat(&SQLFormatRequest{
		Input: TextInput{Data: "select 'from where' as x from t"},
	})
	if err != nil {
		t.Fatalf("格式化失败: %v", err)
	}
	if !strings.Contains(res.Output, "'from where'") {
		t.Errorf("字符串内容不应被改动：%s", res.Output)
	}
}

func TestFormatConvertJSONToYAML(t *testing.T) {
	res, err := FormatConvert(&FormatConvertRequest{
		Input: TextInput{Data: `{"name":"toolbit","count":3}`},
		From:  DataFormatJSON,
		To:    DataFormatYAML,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if !strings.Contains(res.Output, "name: toolbit") {
		t.Errorf("输出 = %s", res.Output)
	}
	if !strings.Contains(res.Output, "count: 3") {
		t.Errorf("输出 = %s", res.Output)
	}
}

func TestFormatConvertYAMLToJSON(t *testing.T) {
	res, err := FormatConvert(&FormatConvertRequest{
		Input: TextInput{Data: "name: toolbit\ncount: 3\n"},
		From:  DataFormatYAML,
		To:    DataFormatJSON,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if !strings.Contains(res.Output, `"name": "toolbit"`) {
		t.Errorf("输出 = %s", res.Output)
	}
}

func TestFormatConvertXMLToJSON(t *testing.T) {
	res, err := FormatConvert(&FormatConvertRequest{
		Input: TextInput{Data: `<root><name>toolbit</name></root>`},
		From:  DataFormatXML,
		To:    DataFormatJSON,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if !strings.Contains(res.Output, `"name": "toolbit"`) {
		t.Errorf("输出 = %s", res.Output)
	}
}

func TestFormatConvertPreservesLargeIntegers(t *testing.T) {
	// 超过 float64 精度的整数必须原样保留。
	big := "123456789012345678901234567890"
	res, err := FormatConvert(&FormatConvertRequest{
		Input: TextInput{Data: `{"n":` + big + `}`},
		From:  DataFormatJSON,
		To:    DataFormatYAML,
	})
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if !strings.Contains(res.Output, big) {
		t.Errorf("大整数精度丢失：%s", res.Output)
	}
}

func TestFormatConvertRejectsInvalidSource(t *testing.T) {
	_, err := FormatConvert(&FormatConvertRequest{
		Input: TextInput{Data: "{not json"},
		From:  DataFormatJSON,
		To:    DataFormatYAML,
	})
	if err == nil {
		t.Fatal("非法源格式应报错")
	}
}
