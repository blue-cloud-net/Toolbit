package toolerr

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestRegistryCoversAllCodes(t *testing.T) {
	// 注册表里的每个码都必须有 HTTP 状态、兜底文案与 i18n key。
	for code, spec := range Registry {
		if spec.HTTP < 100 || spec.HTTP > 599 {
			t.Errorf("错误码 %s 的 HTTP 状态非法：%d", code, spec.HTTP)
		}
		if spec.Message == "" {
			t.Errorf("错误码 %s 缺少兜底文案", code)
		}
		if spec.I18nKey != "errors."+string(code) {
			t.Errorf("错误码 %s 的 i18n key 不匹配：%s", code, spec.I18nKey)
		}
	}
}

func TestRegistryIsComplete(t *testing.T) {
	// 抽样校验关键错误码存在且映射到预期 HTTP 状态。
	cases := map[Code]int{
		CodeInvalidField:       http.StatusBadRequest,
		CodeMissingField:       http.StatusBadRequest,
		CodeInvalidBase64:      http.StatusBadRequest,
		CodeTooLarge:           http.StatusRequestEntityTooLarge,
		CodeRateLimited:        http.StatusTooManyRequests,
		CodeTimeout:            http.StatusGatewayTimeout,
		CodeToolNotFound:       http.StatusNotFound,
		CodeToolDisabled:       http.StatusNotFound,
		CodeNotImplemented:     http.StatusNotImplemented,
		CodeInternalError:      http.StatusInternalServerError,
		CodeCrossOriginBlocked: http.StatusForbidden,
		CodeStatsUnauthorized:  http.StatusUnauthorized,
	}
	for code, want := range cases {
		spec, ok := Registry[code]
		if !ok {
			t.Errorf("注册表缺少错误码 %s", code)
			continue
		}
		if spec.HTTP != want {
			t.Errorf("错误码 %s 的 HTTP 状态 = %d，期望 %d", code, spec.HTTP, want)
		}
	}
}

func TestRegistryHasNoDuplicateSpec(t *testing.T) {
	// 同一语义不得有两个错误码：检查每个码只出现一次（map 天然唯一），
	// 同时确保没有大小写变体造成的重复语义。
	seen := make(map[string]Code, len(Registry))
	for code := range Registry {
		key := strings.ToUpper(string(code))
		if prev, dup := seen[key]; dup {
			t.Errorf("错误码 %s 与 %s 重复", code, prev)
		}
		seen[key] = code
	}
}

func TestErrorMessageHasNoUserInput(t *testing.T) {
	// 兜底文案来自注册表，不得包含任何用户输入。
	e := New(CodeInvalidField).WithParam("field", "key").WithParam("actual", 24)
	msg := e.Error()
	if strings.Contains(msg, "key") && !strings.Contains(msg, "field") {
		t.Errorf("兜底文案疑似包含用户输入：%s", msg)
	}
	if msg != "invalid field value" {
		t.Errorf("兜底文案 = %q，期望 %q", msg, "invalid field value")
	}
}

func TestErrorHTTPStatusAndI18nKey(t *testing.T) {
	e := New(CodeTooLarge)
	if got := e.HTTPStatus(); got != http.StatusRequestEntityTooLarge {
		t.Errorf("HTTPStatus() = %d，期望 %d", got, http.StatusRequestEntityTooLarge)
	}
	if got := e.I18nKey(); got != "errors.TOO_LARGE" {
		t.Errorf("I18nKey() = %q，期望 %q", got, "errors.TOO_LARGE")
	}
}

func TestErrorUnknownCodeFallsBack(t *testing.T) {
	// 未登记的错误码不得泄露到传输层，一律回落 500。
	e := New(Code("NOT_REGISTERED"))
	if got := e.HTTPStatus(); got != http.StatusInternalServerError {
		t.Errorf("未登记错误码的 HTTP 状态 = %d，期望 500", got)
	}
	if got := e.I18nKey(); got != "errors.INTERNAL_ERROR" {
		t.Errorf("未登记错误码的 i18n key = %q，期望 errors.INTERNAL_ERROR", got)
	}
	if got := e.Error(); got != "unknown error" {
		t.Errorf("未登记错误码的文案 = %q，期望 %q", got, "unknown error")
	}
}

func TestErrorWrapKeepsCause(t *testing.T) {
	cause := errors.New("底层库错误")
	e := Wrap(CodeInvalidKey, cause)
	if !errors.Is(e, cause) {
		t.Error("Wrap 后应能通过 errors.Is 找到底层错误")
	}
	if e.Error() != "invalid key" {
		t.Errorf("对外文案 = %q，期望 %q", e.Error(), "invalid key")
	}
}

func TestFromNormalizes(t *testing.T) {
	// 已是 *Error 的原样返回。
	orig := New(CodeInvalidJSON)
	if got := From(orig); got != orig {
		t.Error("From 应原样返回 *Error")
	}
	// 未知错误归一化为内部错误。
	plain := errors.New("boom")
	got := From(plain)
	if got.Code != CodeInternalError {
		t.Errorf("From 的错误码 = %s，期望 INTERNAL_ERROR", got.Code)
	}
	if !errors.Is(got, plain) {
		t.Error("归一化后应保留底层错误链")
	}
	if From(nil) != nil {
		t.Error("From(nil) 应返回 nil")
	}
}

func TestWithParamBuildsParams(t *testing.T) {
	e := New(CodeKeySizeMismatch).
		WithParam("field", "key").
		WithParam("expected", 16).
		WithParam("actual", 24)
	if len(e.Params) != 3 {
		t.Fatalf("参数数量 = %d，期望 3", len(e.Params))
	}
	if e.Params["expected"] != 16 {
		t.Errorf("expected = %v，期望 16", e.Params["expected"])
	}
}

func TestHelperFunctionsMatchRegistry(t *testing.T) {
	for code, spec := range Registry {
		if got := Message(code); got != spec.Message {
			t.Errorf("Message(%s) = %q，期望 %q", code, got, spec.Message)
		}
		if got := HTTPStatusOf(code); got != spec.HTTP {
			t.Errorf("HTTPStatusOf(%s) = %d，期望 %d", code, got, spec.HTTP)
		}
		if got := I18nKeyOf(code); got != spec.I18nKey {
			t.Errorf("I18nKeyOf(%s) = %q，期望 %q", code, got, spec.I18nKey)
		}
	}
}
