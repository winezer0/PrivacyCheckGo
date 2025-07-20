package core

import (
	"fmt"
	"runtime"
)

const (
	// Version 程序版本
	Version = "0.0.1"
	// BuildDate 构建日期（在构建时通过ldflags设置）
	BuildDate = "unknown"
	// GitCommit Git提交哈希（在构建时通过ldflags设置）
	GitCommit = "unknown"
)

// VersionInfo 版本信息结构
type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersionInfo 获取版本信息
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetVersionString 获取版本字符串
func GetVersionString() string {
	info := GetVersionInfo()
	return fmt.Sprintf("PrivacyCheck Go版本 v%s\n构建日期: %s\nGit提交: %s\nGo版本: %s\n平台: %s",
		info.Version, info.BuildDate, info.GitCommit, info.GoVersion, info.Platform)
}
