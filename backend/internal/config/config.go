// Package config 集中解析并校验全部环境变量配置。
//
// 约定：
//   - 全部配置通过环境变量注入，前缀 TOOLBIT_；
//   - 启动时一次性解析，非法值**必须** fail-fast（禁止用「静默默认值」掩盖错误）；
//   - 错误信息**只**包含变量名，**不**包含变量值；
//   - 配置项**不得**在运行时通过接口修改（v1 无热更新）。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// 环境变量名。取值与 docs/deploy.md §3 的表格逐项对应。
const (
	EnvAddr                    = "TOOLBIT_ADDR"
	EnvWebRoot                 = "TOOLBIT_WEB_ROOT"
	EnvDBPath                  = "TOOLBIT_DB_PATH"
	EnvLogLevel                = "TOOLBIT_LOG_LEVEL"
	EnvLogFormat               = "TOOLBIT_LOG_FORMAT"
	EnvTrustProxy              = "TOOLBIT_TRUST_PROXY"
	EnvAllowedOrigins          = "TOOLBIT_ALLOWED_ORIGINS"
	EnvEnableInternalEndpoints = "TOOLBIT_ENABLE_INTERNAL_ENDPOINTS"
	EnvRateLimitEnabled        = "TOOLBIT_RATE_LIMIT_ENABLED"
	EnvStatsEnabled            = "TOOLBIT_STATS_ENABLED"
	EnvNetworkEnabled          = "TOOLBIT_NETWORK_ENABLED"
	EnvNetworkAllowlist        = "TOOLBIT_NETWORK_ALLOWLIST"
	EnvRequestTimeout          = "TOOLBIT_REQUEST_TIMEOUT"
	EnvMaxBodyBytes            = "TOOLBIT_MAX_BODY_BYTES"
	EnvShutdownTimeout         = "TOOLBIT_SHUTDOWN_TIMEOUT"
	EnvStatsToken              = "TOOLBIT_STATS_TOKEN"
)

// 默认值。
const (
	DefaultAddr            = ":30234"
	DefaultWebRoot         = "./frontend/dist"
	DefaultDBPath          = "./data/stats.db"
	DefaultLogLevel        = "info"
	DefaultLogFormat       = "json"
	DefaultRequestTimeout  = 30 * time.Second
	DefaultMaxBodyBytes    = 100 * 1024 * 1024
	DefaultShutdownTimeout = 15 * time.Second
)

// Config 是全部配置的聚合。
type Config struct {
	// Addr 是监听地址。
	Addr string
	// WebRoot 是前端静态资源目录。
	WebRoot string
	// DBPath 是 SQLite 统计库路径。
	DBPath string
	// LogLevel 是日志级别：debug / info / warn / error。
	LogLevel string
	// LogFormat 是日志格式：json / text。
	LogFormat string
	// TrustProxy 表示是否信任 X-Forwarded-For。
	TrustProxy bool
	// AllowedOrigins 是 Origin 白名单；为空表示不启用额外白名单。
	AllowedOrigins []string
	// EnableInternalEndpoints 表示是否注册 /internal/* 端点。
	EnableInternalEndpoints bool
	// RateLimitEnabled 是限流总开关。
	RateLimitEnabled bool
	// StatsEnabled 是匿名统计总开关。
	StatsEnabled bool
	// NetworkEnabled 是出网工具总开关。
	NetworkEnabled bool
	// NetworkAllowlist 是额外允许出网的主机。
	NetworkAllowlist []string
	// RequestTimeout 是全局兜底超时。
	RequestTimeout time.Duration
	// MaxBodyBytes 是全局兜底请求体上限。
	MaxBodyBytes int64
	// ShutdownTimeout 是优雅退出的等待时长。
	ShutdownTimeout time.Duration
	// StatsToken 是统计详情端点的 Token；为空表示该端点不注册。
	StatsToken string
}

// Error 是配置错误。
//
// 只携带变量名，**不**携带变量值，避免敏感值进日志。
type Error struct {
	// Var 是出问题的环境变量名。
	Var string
	// Reason 是原因描述。
	Reason string
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	return fmt.Sprintf("配置 %s 无效：%s", e.Var, e.Reason)
}

// Load 从环境读取并校验配置。
//
// 任一变量非法即返回错误，调用方**必须**据此终止启动。
func Load() (*Config, error) {
	return load(os.LookupEnv)
}

// load 是可注入的环境读取入口，便于测试。
func load(lookup func(string) (string, bool)) (*Config, error) {
	cfg := &Config{
		Addr:             stringOr(lookup, EnvAddr, DefaultAddr),
		WebRoot:          stringOr(lookup, EnvWebRoot, DefaultWebRoot),
		DBPath:           stringOr(lookup, EnvDBPath, DefaultDBPath),
		LogLevel:         stringOr(lookup, EnvLogLevel, DefaultLogLevel),
		LogFormat:        stringOr(lookup, EnvLogFormat, DefaultLogFormat),
		AllowedOrigins:   listOr(lookup, EnvAllowedOrigins),
		NetworkAllowlist: listOr(lookup, EnvNetworkAllowlist),
		StatsToken:       stringOr(lookup, EnvStatsToken, ""),
	}

	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return nil, &Error{Var: EnvLogLevel, Reason: "取值必须是 debug / info / warn / error 之一"}
	}

	switch cfg.LogFormat {
	case "json", "text":
	default:
		return nil, &Error{Var: EnvLogFormat, Reason: "取值必须是 json / text 之一"}
	}

	var err error
	if cfg.TrustProxy, err = boolOr(lookup, EnvTrustProxy, false); err != nil {
		return nil, err
	}
	if cfg.EnableInternalEndpoints, err = boolOr(lookup, EnvEnableInternalEndpoints, false); err != nil {
		return nil, err
	}
	if cfg.RateLimitEnabled, err = boolOr(lookup, EnvRateLimitEnabled, true); err != nil {
		return nil, err
	}
	if cfg.StatsEnabled, err = boolOr(lookup, EnvStatsEnabled, true); err != nil {
		return nil, err
	}
	if cfg.NetworkEnabled, err = boolOr(lookup, EnvNetworkEnabled, true); err != nil {
		return nil, err
	}

	if cfg.MaxBodyBytes, err = int64Or(lookup, EnvMaxBodyBytes, DefaultMaxBodyBytes); err != nil {
		return nil, err
	}
	if cfg.MaxBodyBytes <= 0 {
		return nil, &Error{Var: EnvMaxBodyBytes, Reason: "必须是正整数"}
	}

	if cfg.RequestTimeout, err = durationOr(lookup, EnvRequestTimeout, DefaultRequestTimeout); err != nil {
		return nil, err
	}
	if cfg.RequestTimeout <= 0 {
		return nil, &Error{Var: EnvRequestTimeout, Reason: "必须是正时长"}
	}

	if cfg.ShutdownTimeout, err = durationOr(lookup, EnvShutdownTimeout, DefaultShutdownTimeout); err != nil {
		return nil, err
	}
	if cfg.ShutdownTimeout <= 0 {
		return nil, &Error{Var: EnvShutdownTimeout, Reason: "必须是正时长"}
	}

	return cfg, nil
}

// stringOr 读取字符串配置，未设置时返回默认值。
func stringOr(lookup func(string) (string, bool), key, def string) string {
	if v, ok := lookup(key); ok {
		return v
	}
	return def
}

// boolOr 读取布尔配置，未设置时返回默认值。
func boolOr(lookup func(string) (string, bool), key string, def bool) (bool, error) {
	raw, ok := lookup(key)
	if !ok || raw == "" {
		return def, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, &Error{Var: key, Reason: "取值必须是布尔值"}
	}
	return v, nil
}

// int64Or 读取整数配置，未设置时返回默认值。
func int64Or(lookup func(string) (string, bool), key string, def int64) (int64, error) {
	raw, ok := lookup(key)
	if !ok || raw == "" {
		return def, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, &Error{Var: key, Reason: "取值必须是整数"}
	}
	return v, nil
}

// durationOr 读取时长配置，未设置时返回默认值。
func durationOr(lookup func(string) (string, bool), key string, def time.Duration) (time.Duration, error) {
	raw, ok := lookup(key)
	if !ok || raw == "" {
		return def, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, &Error{Var: key, Reason: "取值必须是合法时长"}
	}
	return v, nil
}

// listOr 读取逗号分隔的列表配置，未设置时返回空列表。
func listOr(lookup func(string) (string, bool), key string) []string {
	raw, ok := lookup(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// Summary 返回配置摘要，用于启动日志。
//
// 摘要中**只**包含非敏感项；Token 一律显示为「已配置 / 未配置」。
func (c *Config) Summary() string {
	token := "未配置"
	if c.StatsToken != "" {
		token = "已配置"
	}
	return fmt.Sprintf(
		"addr=%s webRoot=%s dbPath=%s log=%s/%s trustProxy=%t rateLimit=%t stats=%t network=%t internal=%t statsToken=%s",
		c.Addr, c.WebRoot, c.DBPath, c.LogFormat, c.LogLevel, c.TrustProxy,
		c.RateLimitEnabled, c.StatsEnabled, c.NetworkEnabled, c.EnableInternalEndpoints, token,
	)
}
