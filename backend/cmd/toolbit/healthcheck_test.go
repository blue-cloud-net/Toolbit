package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/config"
)

// lookupEnv 把 map 适配成 runHealthcheck 需要的环境读取函数。
func lookupEnv(env map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

// TestHealthcheckURL 覆盖监听地址到探活 URL 的换算。
func TestHealthcheckURL(t *testing.T) {
	cases := []struct {
		name    string
		addr    string
		want    string
		wantErr bool
	}{
		{"仅端口", ":30234", "http://127.0.0.1:30234/readyz", false},
		{"通配 IPv4", "0.0.0.0:30234", "http://127.0.0.1:30234/readyz", false},
		{"通配 IPv6", "[::]:30234", "http://127.0.0.1:30234/readyz", false},
		{"回环 IPv4", "127.0.0.1:30234", "http://127.0.0.1:30234/readyz", false},
		{"回环 IPv6", "[::1]:30234", "http://[::1]:30234/readyz", false},
		{"主机名", "localhost:9000", "http://localhost:9000/readyz", false},
		{"缺少端口", "30234", "", true},
		{"空地址", "", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := healthcheckURL(tc.addr)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("期望报错，得到 %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外错误：%v", err)
			}
			if got != tc.want {
				t.Errorf("URL = %q，期望 %q", got, tc.want)
			}
		})
	}
}

// TestListenAddr 校验地址读取与默认值回落。
func TestListenAddr(t *testing.T) {
	if got := listenAddr(lookupEnv(nil)); got != config.DefaultAddr {
		t.Errorf("未设置时 = %q，期望默认值 %q", got, config.DefaultAddr)
	}
	if got := listenAddr(lookupEnv(map[string]string{config.EnvAddr: ""})); got != config.DefaultAddr {
		t.Errorf("空值时 = %q，期望默认值 %q", got, config.DefaultAddr)
	}
	if got := listenAddr(lookupEnv(map[string]string{config.EnvAddr: ":9000"})); got != ":9000" {
		t.Errorf("显式设置时 = %q，期望 :9000", got)
	}
}

// TestRunHealthcheckOK 验证 200 时退出码为 0 且只输出 ok。
func TestRunHealthcheckOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != healthcheckPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	code := runHealthcheck(&out, &errOut, lookupEnv(map[string]string{
		config.EnvAddr: serverAddr(t, srv.URL),
	}))

	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d（stderr=%q）", code, exitOK, errOut.String())
	}
	if got := strings.TrimSpace(out.String()); got != "ok" {
		t.Errorf("stdout = %q，期望 ok", got)
	}
	if errOut.Len() != 0 {
		t.Errorf("成功时不应写 stderr，得到 %q", errOut.String())
	}
}

// TestRunHealthcheckNotReady 验证非 200 时退出码为 1，且响应体不进输出。
func TestRunHealthcheckNotReady(t *testing.T) {
	const sensitive = "sensitive-detail-must-not-leak"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(sensitive))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	code := runHealthcheck(&out, &errOut, lookupEnv(map[string]string{
		config.EnvAddr: serverAddr(t, srv.URL),
	}))

	if code != exitFail {
		t.Fatalf("退出码 = %d，期望 %d", code, exitFail)
	}
	if out.Len() != 0 {
		t.Errorf("失败时不应写 stdout，得到 %q", out.String())
	}
	if !strings.Contains(errOut.String(), "503") {
		t.Errorf("stderr 应包含状态码，得到 %q", errOut.String())
	}
	if strings.Contains(errOut.String(), sensitive) {
		t.Error("stderr 泄漏了响应体内容")
	}
}

// TestRunHealthcheckUnreachable 验证端口不可达时退出码为 1。
func TestRunHealthcheckUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	addr := serverAddr(t, srv.URL)
	srv.Close() // 主动关闭，制造连接拒绝

	var out, errOut bytes.Buffer
	code := runHealthcheck(&out, &errOut, lookupEnv(map[string]string{config.EnvAddr: addr}))

	if code != exitFail {
		t.Fatalf("退出码 = %d，期望 %d", code, exitFail)
	}
	if !strings.Contains(errOut.String(), "不可达") {
		t.Errorf("stderr 应说明不可达，得到 %q", errOut.String())
	}
}

// TestRunHealthcheckInvalidAddr 验证地址缺少端口时退出码为 1。
func TestRunHealthcheckInvalidAddr(t *testing.T) {
	var out, errOut bytes.Buffer
	code := runHealthcheck(&out, &errOut, lookupEnv(map[string]string{config.EnvAddr: "30234"}))

	if code != exitFail {
		t.Fatalf("退出码 = %d，期望 %d", code, exitFail)
	}
	if !strings.Contains(errOut.String(), config.EnvAddr) {
		t.Errorf("stderr 应指出变量名，得到 %q", errOut.String())
	}
}

// serverAddr 把 httptest 的 URL 换成 host:port 形态。
func serverAddr(t *testing.T, rawURL string) string {
	t.Helper()
	addr := strings.TrimPrefix(rawURL, "http://")
	if addr == rawURL {
		t.Fatalf("无法解析测试服务器地址：%s", rawURL)
	}
	return addr
}
