package encoding

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func TestURLCodecEncode(t *testing.T) {
	cases := []struct {
		name      string
		data      string
		component string
		want      string
	}{
		{"保留保留字符", "https://a.com/p?x=1", URLComponentFull, "https://a.com/p?x=1"},
		{"转义空格", "a b", URLComponentFull, "a%20b"},
		{"路径段转义斜杠", "a/b", URLComponentPath, "a%2Fb"},
		{"查询值空格转加号", "a b", URLComponentQuery, "a+b"},
		{"片段转义井号", "a#b", URLComponentFragment, "a%23b"},
		{"中文转义", "中", URLComponentFull, "%E4%B8%AD"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := URLCodec(context.Background(), &URLCodecRequest{
				Operation: URLEncode,
				Input:     ByteInput{Data: tc.data, Encoding: EncodingUTF8},
				Component: tc.component,
			})
			if err != nil {
				t.Fatalf("编码失败: %v", err)
			}
			got := res.Data.(URLCodecResult).Output
			if got != tc.want {
				t.Errorf("输出 = %q，期望 %q", got, tc.want)
			}
		})
	}
}

func TestURLCodecDecode(t *testing.T) {
	cases := []struct {
		name        string
		data        string
		plusAsSpace *bool
		want        string
	}{
		{"普通解码", "a%20b", nil, "a b"},
		{"加号视为空格", "a+b", nil, "a b"},
		{"加号保留", "a+b", boolPtr(false), "a+b"},
		{"中文解码", "%E4%B8%AD", nil, "中"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := URLCodec(context.Background(), &URLCodecRequest{
				Operation:   URLDecode,
				Input:       ByteInput{Data: tc.data, Encoding: EncodingUTF8},
				PlusAsSpace: tc.plusAsSpace,
			})
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}
			got := res.Data.(URLCodecResult).Output
			if got != tc.want {
				t.Errorf("输出 = %q，期望 %q", got, tc.want)
			}
		})
	}
}

func TestURLCodecParse(t *testing.T) {
	res, err := URLCodec(context.Background(), &URLCodecRequest{
		Operation: URLParse,
		Input:     ByteInput{Data: "https://example.com:8443/a/b?x=1&y=2&x=3", Encoding: EncodingUTF8},
	})
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	parts := res.Data.(URLCodecResult).Parsed
	if parts == nil {
		t.Fatal("缺少 parsed 结果")
	}
	if parts.Scheme != "https" {
		t.Errorf("scheme = %q，期望 https", parts.Scheme)
	}
	if parts.Host != "example.com" {
		t.Errorf("host = %q，期望 example.com", parts.Host)
	}
	if parts.Port != "8443" {
		t.Errorf("port = %q，期望 8443", parts.Port)
	}
	if parts.Path != "/a/b" {
		t.Errorf("path = %q，期望 /a/b", parts.Path)
	}
	if parts.Query != "x=1&y=2&x=3" {
		t.Errorf("query = %q，期望 x=1&y=2&x=3", parts.Query)
	}
	if len(parts.Params) != 3 {
		t.Fatalf("参数个数 = %d，期望 3", len(parts.Params))
	}
	// 同名参数必须保留多个条目。
	if parts.Params[2].Key != "x" || parts.Params[2].Value != "3" {
		t.Errorf("第 3 个参数 = %+v，期望 x=3", parts.Params[2])
	}
}

func TestURLCodecRejectsInvalidOperation(t *testing.T) {
	_, err := URLCodec(context.Background(), &URLCodecRequest{
		Operation: "nope",
		Input:     ByteInput{Data: "x", Encoding: EncodingUTF8},
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}

func TestURLCodecRejectsInvalidPercent(t *testing.T) {
	_, err := URLCodec(context.Background(), &URLCodecRequest{
		Operation: URLDecode,
		Input:     ByteInput{Data: "%zz", Encoding: EncodingUTF8},
	})
	assertToolErrCode(t, err, toolerr.CodeInvalidEncoding)
}

// boolPtr 返回指向 v 的布尔指针。
func boolPtr(v bool) *bool {
	return &v
}

// assertToolErrCode 断言错误的错误码。
func assertToolErrCode(t *testing.T, err error, code toolerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望错误 %s，实际为 nil", code)
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("错误类型 = %T，期望 *toolerr.Error", err)
	}
	if toolErr.Code != code {
		t.Errorf("错误码 = %s，期望 %s", toolErr.Code, code)
	}
}
