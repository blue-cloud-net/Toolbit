package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRouter 构造一个带全部中间件的测试路由。
func newRouter() *gin.Engine {
	r := gin.New()
	r.Use(RequestID(), SecurityHeaders(), AccessLog(nil))
	return r
}

// decodeEnvelope 解析响应信封。
func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) Envelope {
	t.Helper()
	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("响应不是合法 JSON: %v；body=%s", err, rec.Body.String())
	}
	return env
}

func TestSuccessEnvelope(t *testing.T) {
	r := newRouter()
	r.GET("/x", func(c *gin.Context) {
		Success(c, map[string]string{"output": "ok"}, Meta{DurationMs: 12, InputBytes: 256, OutputBytes: 512})
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	env := decodeEnvelope(t, rec)
	if env.Code != 0 {
		t.Errorf("code = %d，期望 0", env.Code)
	}
	if env.Message != "ok" {
		t.Errorf("message = %q，期望 %q", env.Message, "ok")
	}
	if env.Error != nil {
		t.Errorf("error = %v，成功时必须为 null", env.Error)
	}
	if env.RequestID == "" {
		t.Error("requestId 不得为空")
	}
	if env.Meta.DurationMs != 12 || env.Meta.InputBytes != 256 || env.Meta.OutputBytes != 512 {
		t.Errorf("meta = %+v", env.Meta)
	}
	if env.Meta.Warnings == nil {
		t.Error("warnings 必须为空数组，不得为 null")
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q，期望 no-store", got)
	}
}

func TestFailureEnvelope(t *testing.T) {
	r := newRouter()
	r.GET("/x", func(c *gin.Context) {
		Failure(c, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field": "key", "expected": 16, "actual": 24,
		}), NewMeta())
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("状态码 = %d，期望 400", rec.Code)
	}
	env := decodeEnvelope(t, rec)
	if env.Code != http.StatusBadRequest {
		t.Errorf("code = %d，期望等于 HTTP 状态", env.Code)
	}
	if env.Data != nil {
		t.Errorf("data = %v，失败时必须为 null", env.Data)
	}
	if env.Error == nil {
		t.Fatal("error 不得为 null")
	}
	if env.Error.Code != "INVALID_FIELD" {
		t.Errorf("error.code = %q，期望 INVALID_FIELD", env.Error.Code)
	}
	if env.Error.Params["field"] != "key" {
		t.Errorf("error.params.field = %v", env.Error.Params["field"])
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q，期望 no-store", got)
	}
}

func TestRequestIDReused(t *testing.T) {
	r := newRouter()
	r.GET("/x", func(c *gin.Context) {
		Success(c, nil, NewMeta())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil)
	req.Header.Set(RequestIDHeader, "abc-123_XYZ")
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != "abc-123_XYZ" {
		t.Errorf("响应头 requestId = %q，期望复用入站值", got)
	}
	env := decodeEnvelope(t, rec)
	if env.RequestID != "abc-123_XYZ" {
		t.Errorf("信封 requestId = %q", env.RequestID)
	}
}

func TestRequestIDRegeneratedOnInvalidFormat(t *testing.T) {
	r := newRouter()
	r.GET("/x", func(c *gin.Context) {
		Success(c, nil, NewMeta())
	})

	for _, bad := range []string{"has space", strings.Repeat("a", 65), "bad/slash", "emoji😀"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil)
		req.Header.Set(RequestIDHeader, bad)
		r.ServeHTTP(rec, req)

		got := rec.Header().Get(RequestIDHeader)
		if got == bad {
			t.Errorf("非法 requestId %q 被原样透传", bad)
		}
		if got == "" {
			t.Errorf("非法 requestId %q 未生成新值", bad)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := newRouter()
	r.GET("/x", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil))

	if got := rec.Header().Get(HeaderContentTypeOptions); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q，期望 nosniff", got)
	}
	if got := rec.Header().Get(HeaderReferrerPolicy); got != "no-referrer" {
		t.Errorf("Referrer-Policy = %q，期望 no-referrer", got)
	}
}

func TestBodyLimit(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.POST("/x", func(c *gin.Context) {
		LimitBody(c, 10)
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				Failure(c, TooLargeError(maxErr.Limit, 0), NewMeta())
				return
			}
			Failure(c, toolerr.New(toolerr.CodeInvalidRequest), NewMeta())
			return
		}
		Success(c, map[string]int{"read": len(body)}, NewMeta())
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/x", strings.NewReader(strings.Repeat("a", 100))))

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("状态码 = %d，期望 413", rec.Code)
	}
	env := decodeEnvelope(t, rec)
	if env.Error == nil || env.Error.Code != "TOO_LARGE" {
		t.Errorf("error = %+v，期望 TOO_LARGE", env.Error)
	}
}

func TestOriginCheckBlocksCrossSite(t *testing.T) {
	r := gin.New()
	r.Use(RequestID(), OriginCheck(nil))
	r.POST("/x", func(c *gin.Context) {
		Success(c, nil, NewMeta())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/x", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("状态码 = %d，期望 403", rec.Code)
	}
	env := decodeEnvelope(t, rec)
	if env.Error == nil || env.Error.Code != "CROSS_ORIGIN_BLOCKED" {
		t.Errorf("error = %+v，期望 CROSS_ORIGIN_BLOCKED", env.Error)
	}
}

func TestOriginCheckBlocksMismatchedOrigin(t *testing.T) {
	r := gin.New()
	r.Use(RequestID(), OriginCheck(nil))
	r.POST("/x", func(c *gin.Context) {
		Success(c, nil, NewMeta())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/x", nil)
	req.Header.Set("Origin", "https://evil.example")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("状态码 = %d，期望 403", rec.Code)
	}
}

func TestOriginCheckAllowsSameOrigin(t *testing.T) {
	r := gin.New()
	r.Use(RequestID(), OriginCheck(nil))
	r.POST("/x", func(c *gin.Context) {
		Success(c, nil, NewMeta())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/x", nil)
	req.Host = "toolbit.example"
	req.Header.Set("Origin", "https://toolbit.example")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
}

func TestOriginCheckAllowsMissingOrigin(t *testing.T) {
	r := gin.New()
	r.Use(RequestID(), OriginCheck(nil))
	r.POST("/x", func(c *gin.Context) {
		Success(c, nil, NewMeta())
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/x", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
}

func TestOriginCheckAllowsConfiguredOrigin(t *testing.T) {
	r := gin.New()
	r.Use(RequestID(), OriginCheck([]string{"https://allowed.example"}))
	r.POST("/x", func(c *gin.Context) {
		Success(c, nil, NewMeta())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/x", nil)
	req.Header.Set("Origin", "https://allowed.example")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
}

func TestRecoveryReturns500(t *testing.T) {
	r := gin.New()
	r.Use(RequestID(), Recovery(nil))
	r.GET("/x", func(_ *gin.Context) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("状态码 = %d，期望 500", rec.Code)
	}
	env := decodeEnvelope(t, rec)
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("error = %+v，期望 INTERNAL_ERROR", env.Error)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Error("响应不得回显 panic 内容")
	}
}
