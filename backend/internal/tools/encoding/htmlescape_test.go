package encoding

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestHTMLEscapeEncode 断言转义尖括号与引号。
func TestHTMLEscapeEncode(t *testing.T) {
	result, err := HTMLEscape(t.Context(), &HTMLEscapeRequest{
		Mode:  HTMLEscapeEncode,
		Input: `<b>hi</b> & "x"`,
	})
	data := result.Data.(HTMLEscapeResult)
	if err != nil {
		t.Fatalf("转义失败：%v", err)
	}
	if data.Output != "&lt;b&gt;hi&lt;/b&gt; &amp; &quot;x&quot;" {
		t.Errorf("输出 = %q", data.Output)
	}
}

// TestHTMLEscapeDecode 断言反转义实体。
func TestHTMLEscapeDecode(t *testing.T) {
	result, err := HTMLEscape(t.Context(), &HTMLEscapeRequest{
		Mode:  HTMLEscapeDecode,
		Input: "&lt;b&gt;hi&lt;/b&gt;",
	})
	data := result.Data.(HTMLEscapeResult)
	if err != nil {
		t.Fatalf("反转义失败：%v", err)
	}
	if data.Output != "<b>hi</b>" {
		t.Errorf("输出 = %q", data.Output)
	}
}

// TestHTMLEscapeUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestHTMLEscapeUnsupportedMode(t *testing.T) {
	_, err := HTMLEscape(t.Context(), &HTMLEscapeRequest{Mode: "strip", Input: "a"})
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}
