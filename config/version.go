package config

import (
	"fmt"
	"runtime"
)

const Version = "0.0.1"

// VersionInfo 版本信息结构
type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersionInfo 获取版本信息
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
