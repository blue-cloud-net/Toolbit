// Package buildinfo 承载由 ldflags 注入的版本信息。
//
// 注入方式：
//
//	go build -ldflags "-X github.com/blue-cloud-net/Toolbit/backend/internal/buildinfo.Version=1.2.3 \
//	                   -X .../buildinfo.Commit=abc1234 \
//	                   -X .../buildinfo.BuildTime=2026-10-28T08:00:00Z"
//
// 未注入时全部回落为开发默认值，保证本地构建不会因为缺少注入而失败。
package buildinfo

import "runtime"

// 默认值，用于未注入 ldflags 的本地构建。
const (
	// DevVersion 是未注入时的版本号。
	DevVersion = "0.0.0-dev"
	// DevCommit 是未注入时的提交标识。
	DevCommit = "unknown"
	// DevBuildTime 是未注入时的构建时间。
	DevBuildTime = "unknown"
)

// 以下变量由 ldflags 在构建时覆盖。
var (
	// Version 是语义化版本号。
	Version = DevVersion
	// Commit 是短 commit sha。
	Commit = DevCommit
	// BuildTime 是构建时间（RFC3339 UTC）。
	BuildTime = DevBuildTime
)

// Info 是对外暴露的版本信息快照。
type Info struct {
	// Version 是语义化版本号。
	Version string `json:"version"`
	// Commit 是短 commit sha。
	Commit string `json:"commit"`
	// BuildTime 是构建时间（RFC3339 UTC）。
	BuildTime string `json:"buildTime"`
	// GoVersion 是编译所用的 Go 版本。
	GoVersion string `json:"goVersion"`
	// Platform 是构建目标平台。
	Platform string `json:"platform"`
}

// Get 返回当前版本信息快照。
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
