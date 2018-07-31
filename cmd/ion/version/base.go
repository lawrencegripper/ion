package version

import (
	"fmt"
	"runtime"
)

// Overridden via ldflags
var (
	BasePlatform  = "unknown"
	BaseVersion   = "unknown"
	BaseGitCommit = "unknown"
	BaseBuildDate = "unknown"
	BaseGoVersion = "unknown"
)

// ClientVersionInfo holds details about the current client binary
type ClientVersionInfo struct {
	Platform  string `json:"platform"`
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

// GetClientVersion returns details about the current client binary
func GetClientVersion() ClientVersionInfo {
	return ClientVersionInfo{
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Version:   BaseVersion,
		GitCommit: BaseGitCommit,
		BuildDate: BaseBuildDate,
		GoVersion: BaseGoVersion,
	}
}
