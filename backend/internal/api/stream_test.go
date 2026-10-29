package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
)

// dialStream 连接流式端点。
func dialStream(t *testing.T, baseURL string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/api/v1/tools/ssl-check/stream"
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("连接流式端点失败：%v", err)
	}
	return conn
}

// readFrame 读取一帧并解析。
func readFrame(t *testing.T, conn *websocket.Conn) streamFrame {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, payload, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("读取帧失败：%v", err)
	}
	var frame streamFrame
	if unmarshalErr := json.Unmarshal(payload, &frame); unmarshalErr != nil {
		t.Fatalf("解析帧失败：%v", unmarshalErr)
	}
	return frame
}

// readUntilTerminal 跳过进度帧，返回首个 result 或 error 帧。
func readUntilTerminal(t *testing.T, conn *websocket.Conn) streamFrame {
	t.Helper()
	for i := 0; i < 8; i++ {
		frame := readFrame(t, conn)
		if frame.Type == "progress" {
			continue
		}
		return frame
	}
	t.Fatal("只收到进度帧")
	return streamFrame{}
}

// TestSSLStreamInvalidFirstFrame 断言非法首帧以 1008 关闭。
func TestSSLStreamInvalidFirstFrame(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(newTestServer(t))
	defer srv.Close()

	conn := dialStream(t, srv.URL)
	defer func() { _ = conn.CloseNow() }()

	if err := conn.Write(context.Background(), websocket.MessageText, []byte("not json")); err != nil {
		t.Fatalf("写首帧失败：%v", err)
	}

	frame := readFrame(t, conn)
	if frame.Type != "error" || frame.Error == nil || frame.Error.Code != "INVALID_JSON" {
		t.Fatalf("期望 INVALID_JSON 错误帧，实际 %+v", frame)
	}

	_, _, err := conn.Read(context.Background())
	if got := websocket.CloseStatus(err); got != websocket.StatusPolicyViolation {
		t.Fatalf("关闭码 = %d，期望 %d", got, websocket.StatusPolicyViolation)
	}
}

// TestSSLStreamRejectsBlockedTarget 断言回环目标被出网策略拒绝，并以 1008 关闭。
func TestSSLStreamRejectsBlockedTarget(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(newTestServer(t))
	defer srv.Close()

	conn := dialStream(t, srv.URL)
	defer func() { _ = conn.CloseNow() }()

	payload := []byte(`{"host":"127.0.0.1","port":443}`)
	if err := conn.Write(context.Background(), websocket.MessageText, payload); err != nil {
		t.Fatalf("写首帧失败：%v", err)
	}

	frame := readUntilTerminal(t, conn)
	if frame.Type != "error" || frame.Error == nil || frame.Error.Code != "SSRF_BLOCKED" {
		t.Fatalf("期望 SSRF_BLOCKED 错误帧，实际 %+v", frame)
	}

	_, _, err := conn.Read(context.Background())
	if got := websocket.CloseStatus(err); got != websocket.StatusPolicyViolation {
		t.Fatalf("关闭码 = %d，期望 %d", got, websocket.StatusPolicyViolation)
	}
}

// TestSSLStreamProbesLocalTLS 覆盖正常路径：进度帧 + 结果帧 + 1000 关闭。
func TestSSLStreamProbesLocalTLS(t *testing.T) {
	t.Parallel()

	tlsSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer tlsSrv.Close()

	parsed, err := url.Parse(tlsSrv.URL)
	if err != nil {
		t.Fatalf("解析测试服务器地址失败：%v", err)
	}
	_, portText, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatalf("拆分端口失败：%v", err)
	}
	port := mustAtoi(t, portText)

	guard := netguard.New(netguard.Config{
		Enabled:    true,
		Allowlist:  []string{"127.0.0.1"},
		ExtraPorts: []int{port},
	})
	srv := httptest.NewServer(newTestServerWithGuard(t, guard, 1<<20))
	defer srv.Close()

	conn := dialStream(t, srv.URL)
	defer func() { _ = conn.CloseNow() }()

	payload := []byte(`{"host":"127.0.0.1","port":` + portText + `,"sni":"127.0.0.1"}`)
	if writeErr := conn.Write(context.Background(), websocket.MessageText, payload); writeErr != nil {
		t.Fatalf("写首帧失败：%v", writeErr)
	}

	sawProgress := false
	var result *streamFrame
	for i := 0; i < 8; i++ {
		frame := readFrame(t, conn)
		switch frame.Type {
		case "progress":
			sawProgress = true
		case "result":
			result = &frame
		case "error":
			t.Fatalf("探测失败：%+v", frame.Error)
		}
		if result != nil {
			break
		}
	}
	if !sawProgress {
		t.Fatal("未收到进度帧")
	}
	if result == nil {
		t.Fatal("未收到结果帧")
	}

	data, ok := result.Data.(map[string]any)
	if !ok {
		t.Fatalf("结果帧 data 类型不符：%T", result.Data)
	}
	if data["host"] != "127.0.0.1" {
		t.Errorf("host = %v", data["host"])
	}
	if certs, okCerts := data["certificates"].([]any); !okCerts || len(certs) == 0 {
		t.Errorf("未返回证书链：%v", data["certificates"])
	}

	_, _, readErr := conn.Read(context.Background())
	if got := websocket.CloseStatus(readErr); got != websocket.StatusNormalClosure {
		t.Fatalf("关闭码 = %d，期望 %d", got, websocket.StatusNormalClosure)
	}
}

// TestSSLStreamRejectsNonStreamTool 断言非流式工具没有 /stream 端点。
func TestSSLStreamRejectsNonStreamTool(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(newTestServer(t))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/tools/cidr-calc/stream") //nolint:noctx // 测试内固定 URL
	if err != nil {
		t.Fatalf("请求失败：%v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("状态码 = %d，期望 404", resp.StatusCode)
	}
}

// TestClientTransportToolHasNoEndpoint 断言 transport: client 的工具没有后端端点。
func TestClientTransportToolHasNoEndpoint(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(newTestServer(t))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		srv.URL+"/api/v1/tools/websocket-test", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("构造请求失败：%v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求失败：%v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("状态码 = %d，期望 404", resp.StatusCode)
	}
}

// TestStreamToolRejectsPost 断言流式工具不接受 JSON POST。
func TestStreamToolRejectsPost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(newTestServer(t))
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		srv.URL+"/api/v1/tools/ssl-check", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("构造请求失败：%v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求失败：%v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("状态码 = %d，期望 405", resp.StatusCode)
	}
}

// mustAtoi 解析整数并在失败时终止测试。
func mustAtoi(t *testing.T, text string) int {
	t.Helper()
	value := 0
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			t.Fatalf("非法端口：%s", text)
		}
		value = value*10 + int(ch-'0')
	}
	return value
}
