package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestTextConvertStyles 断言各命名风格的输出。
func TestTextConvertStyles(t *testing.T) {
	cases := []struct {
		name  string
		style string
		want  string
	}{
		{"大写", StyleUpper, "HTTPSERVER"},
		{"小写", StyleLower, "httpserver"},
		{"驼峰", StyleCamel, "httpServer"},
		{"帕斯卡", StylePascal, "HttpServer"},
		{"下划线", StyleSnake, "http_server"},
		{"短横线", StyleKebab, "http-server"},
		{"常量", StyleConstant, "HTTP_SERVER"},
		{"标题", StyleTitle, "Http Server"},
		{"句首大写", StyleSentence, "Http server"},
		{"点分", StyleDot, "http.server"},
		{"路径", StylePath, "http/server"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TextConvert(&TextConvertRequest{
				Input: TextInput{Data: "http-server"},
				Style: tc.style,
			})
			if err != nil {
				t.Fatalf("转换失败：%v", err)
			}
			if result.Output != tc.want {
				t.Errorf("输出 = %q，期望 %q", result.Output, tc.want)
			}
		})
	}
}

// TestTextConvertWords 断言词元切分。
func TestTextConvertWords(t *testing.T) {
	result, err := TextConvert(&TextConvertRequest{
		Input: TextInput{Data: "Hello, World! 123"},
		Style: StyleSnake,
	})
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	want := []string{"Hello", "World", "123"}
	if len(result.Words) != len(want) {
		t.Fatalf("词元 = %v，期望 %v", result.Words, want)
	}
	for i, w := range want {
		if result.Words[i] != w {
			t.Errorf("第 %d 个词元 = %q，期望 %q", i, result.Words[i], w)
		}
	}
}

// TestTextConvertUnknownStyle 断言未知风格报 INVALID_FIELD。
func TestTextConvertUnknownStyle(t *testing.T) {
	_, err := TextConvert(&TextConvertRequest{
		Input: TextInput{Data: "abc"},
		Style: "unknown",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestTextConvertMissingInput 断言缺少输入时报 MISSING_FIELD。
func TestTextConvertMissingInput(t *testing.T) {
	_, err := TextConvert(&TextConvertRequest{
		Style: StyleSnake,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}
