package network

import (
	"context"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// headerGenerator 执行一次请求头生成并取出结果。
func headerGenerator(t *testing.T, req *HeaderGeneratorRequest) HeaderGeneratorResult {
	t.Helper()
	result, err := HeaderGenerator(context.Background(), req)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	out, ok := result.Data.(HeaderGeneratorResult)
	if !ok {
		t.Fatalf("响应类型不符：%T", result.Data)
	}
	return out
}

func TestHeaderGeneratorChrome(t *testing.T) {
	t.Parallel()

	out := headerGenerator(t, &HeaderGeneratorRequest{
		Platform: HeaderPlatformChrome,
		Version:  "120.0.0.0",
	})
	if len(out.Headers) == 0 {
		t.Fatal("应给出请求头")
	}
	if !strings.Contains(out.UserAgent, "Chrome/120.0.0.0") {
		t.Errorf("UA = %q", out.UserAgent)
	}
	if out.Headers[0].Name != "User-Agent" {
		t.Errorf("首个请求头 = %q，期望 User-Agent", out.Headers[0].Name)
	}
}

func TestHeaderGeneratorDefaults(t *testing.T) {
	t.Parallel()

	// 平台留空时应回落到 chrome 的内置默认值。
	out := headerGenerator(t, &HeaderGeneratorRequest{})
	if !strings.Contains(out.UserAgent, "Chrome/"+headerDefaults[HeaderPlatformChrome].version) {
		t.Errorf("UA = %q", out.UserAgent)
	}
}

func TestHeaderGeneratorCurl(t *testing.T) {
	t.Parallel()

	out := headerGenerator(t, &HeaderGeneratorRequest{Platform: HeaderPlatformCurl})
	if !strings.HasPrefix(out.UserAgent, "curl/") {
		t.Errorf("UA = %q，期望以 curl/ 开头", out.UserAgent)
	}
}

func TestHeaderGeneratorRefererAndCustomHeaders(t *testing.T) {
	t.Parallel()

	out := headerGenerator(t, &HeaderGeneratorRequest{
		Referer: "https://example.com/",
		CustomHeaders: map[string]string{
			"X-Trace-Id":   "abc123",
			"X-Request-Id": "def456",
		},
	})

	names := make([]string, 0, len(out.Headers))
	for _, header := range out.Headers {
		names = append(names, header.Name)
	}
	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "Referer") {
		t.Errorf("应包含 Referer 请求头：%v", names)
	}
	// 自定义请求头按字典序输出，顺序稳定。
	if !strings.Contains(joined, "X-Request-Id,X-Trace-Id") {
		t.Errorf("自定义请求头顺序不稳定：%v", names)
	}
}

func TestHeaderGeneratorUnknownPlatform(t *testing.T) {
	t.Parallel()

	_, err := HeaderGenerator(context.Background(), &HeaderGeneratorRequest{Platform: "netscape"})
	assertCode(t, err, toolerr.CodeInvalidField)
}

func TestHeaderGeneratorInvalidCustomHeaderName(t *testing.T) {
	t.Parallel()

	_, err := HeaderGenerator(context.Background(), &HeaderGeneratorRequest{
		CustomHeaders: map[string]string{"bad header": "x"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
