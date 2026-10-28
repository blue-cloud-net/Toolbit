package config

import (
	"strings"
	"testing"
	"time"
)

// lookupFrom 构造一个可注入的环境读取函数。
func lookupFrom(env map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := load(lookupFrom(map[string]string{}))
	if err != nil {
		t.Fatalf("加载默认配置失败: %v", err)
	}
	if cfg.Addr != DefaultAddr {
		t.Errorf("Addr = %q，期望 %q", cfg.Addr, DefaultAddr)
	}
	if cfg.WebRoot != DefaultWebRoot {
		t.Errorf("WebRoot = %q，期望 %q", cfg.WebRoot, DefaultWebRoot)
	}
	if cfg.DBPath != DefaultDBPath {
		t.Errorf("DBPath = %q，期望 %q", cfg.DBPath, DefaultDBPath)
	}
	if cfg.LogLevel != DefaultLogLevel {
		t.Errorf("LogLevel = %q，期望 %q", cfg.LogLevel, DefaultLogLevel)
	}
	if cfg.LogFormat != DefaultLogFormat {
		t.Errorf("LogFormat = %q，期望 %q", cfg.LogFormat, DefaultLogFormat)
	}
	if cfg.TrustProxy {
		t.Error("TrustProxy 默认应为 false")
	}
	if cfg.EnableInternalEndpoints {
		t.Error("EnableInternalEndpoints 默认应为 false")
	}
	if !cfg.RateLimitEnabled {
		t.Error("RateLimitEnabled 默认应为 true")
	}
	if !cfg.StatsEnabled {
		t.Error("StatsEnabled 默认应为 true")
	}
	if !cfg.NetworkEnabled {
		t.Error("NetworkEnabled 默认应为 true")
	}
	if cfg.RequestTimeout != DefaultRequestTimeout {
		t.Errorf("RequestTimeout = %v，期望 %v", cfg.RequestTimeout, DefaultRequestTimeout)
	}
	if cfg.MaxBodyBytes != DefaultMaxBodyBytes {
		t.Errorf("MaxBodyBytes = %d，期望 %d", cfg.MaxBodyBytes, DefaultMaxBodyBytes)
	}
	if cfg.ShutdownTimeout != DefaultShutdownTimeout {
		t.Errorf("ShutdownTimeout = %v，期望 %v", cfg.ShutdownTimeout, DefaultShutdownTimeout)
	}
	if cfg.StatsToken != "" {
		t.Errorf("StatsToken 默认应为空，实际 %q", cfg.StatsToken)
	}
}

func TestLoadOverrides(t *testing.T) {
	cfg, err := load(lookupFrom(map[string]string{
		EnvAddr:                    "127.0.0.1:9000",
		EnvWebRoot:                 "/srv/www",
		EnvDBPath:                  "/var/lib/toolbit/stats.db",
		EnvLogLevel:                "debug",
		EnvLogFormat:               "text",
		EnvTrustProxy:              "true",
		EnvAllowedOrigins:          "https://a.example, https://b.example",
		EnvEnableInternalEndpoints: "1",
		EnvRateLimitEnabled:        "false",
		EnvStatsEnabled:            "0",
		EnvNetworkEnabled:          "false",
		EnvNetworkAllowlist:        "example.com",
		EnvRequestTimeout:          "5s",
		EnvMaxBodyBytes:            "1024",
		EnvShutdownTimeout:         "3s",
		EnvStatsToken:              "secret-token",
	}))
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if cfg.Addr != "127.0.0.1:9000" {
		t.Errorf("Addr = %q", cfg.Addr)
	}
	if !cfg.TrustProxy {
		t.Error("TrustProxy 应为 true")
	}
	if !cfg.EnableInternalEndpoints {
		t.Error("EnableInternalEndpoints 应为 true")
	}
	if cfg.RateLimitEnabled {
		t.Error("RateLimitEnabled 应为 false")
	}
	if cfg.StatsEnabled {
		t.Error("StatsEnabled 应为 false")
	}
	if cfg.NetworkEnabled {
		t.Error("NetworkEnabled 应为 false")
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins 数量 = %d，期望 2", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "https://a.example" {
		t.Errorf("AllowedOrigins[0] = %q，期望去除空格", cfg.AllowedOrigins[0])
	}
	if cfg.RequestTimeout != 5*time.Second {
		t.Errorf("RequestTimeout = %v，期望 5s", cfg.RequestTimeout)
	}
	if cfg.MaxBodyBytes != 1024 {
		t.Errorf("MaxBodyBytes = %d，期望 1024", cfg.MaxBodyBytes)
	}
	if cfg.ShutdownTimeout != 3*time.Second {
		t.Errorf("ShutdownTimeout = %v，期望 3s", cfg.ShutdownTimeout)
	}
	if cfg.StatsToken != "secret-token" {
		t.Errorf("StatsToken = %q", cfg.StatsToken)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
	}{
		{"非法日志级别", map[string]string{EnvLogLevel: "verbose"}},
		{"非法日志格式", map[string]string{EnvLogFormat: "xml"}},
		{"非法布尔值", map[string]string{EnvTrustProxy: "maybe"}},
		{"非法整数", map[string]string{EnvMaxBodyBytes: "big"}},
		{"非正上限", map[string]string{EnvMaxBodyBytes: "0"}},
		{"非法时长", map[string]string{EnvRequestTimeout: "5 seconds"}},
		{"非正时长", map[string]string{EnvShutdownTimeout: "0s"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := load(lookupFrom(tc.env))
			if err == nil {
				t.Fatal("期望 fail-fast，但通过了")
			}
			var cfgErr *Error
			if !asConfigError(err, &cfgErr) {
				t.Fatalf("错误类型 = %T，期望 *config.Error", err)
			}
			if cfgErr.Var == "" {
				t.Error("配置错误必须指明变量名")
			}
		})
	}
}

func TestConfigErrorDoesNotLeakValue(t *testing.T) {
	// 错误信息只含变量名与原因，**不**含变量值。
	_, err := load(lookupFrom(map[string]string{EnvLogLevel: "super-secret-value"}))
	if err == nil {
		t.Fatal("期望 fail-fast")
	}
	if strings.Contains(err.Error(), "super-secret-value") {
		t.Errorf("错误信息泄露了变量值：%s", err.Error())
	}
	if !strings.Contains(err.Error(), EnvLogLevel) {
		t.Errorf("错误信息应包含变量名：%s", err.Error())
	}
}

func TestSummaryHidesToken(t *testing.T) {
	cfg, err := load(lookupFrom(map[string]string{EnvStatsToken: "top-secret"}))
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	summary := cfg.Summary()
	if strings.Contains(summary, "top-secret") {
		t.Errorf("配置摘要泄露了 Token：%s", summary)
	}
	if !strings.Contains(summary, "已配置") {
		t.Errorf("配置摘要应标明 Token 状态：%s", summary)
	}
}

func TestListOrIgnoresEmptyEntries(t *testing.T) {
	env := lookupFrom(map[string]string{EnvAllowedOrigins: " , a.example , ,b.example, "})
	cfg, err := load(env)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins = %v，期望 2 项", cfg.AllowedOrigins)
	}
	if cfg.AllowedOrigins[0] != "a.example" || cfg.AllowedOrigins[1] != "b.example" {
		t.Errorf("AllowedOrigins = %v", cfg.AllowedOrigins)
	}
}

// asConfigError 是 errors.As 的薄封装，避免测试文件引入额外的类型断言样板。
func asConfigError(err error, target **Error) bool {
	e, ok := err.(*Error)
	if ok {
		*target = e
	}
	return ok
}
