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

type VersionInfo struct {
	Platform  string `json:"platform"`
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

func GetVersion() VersionInfo {
	return VersionInfo{
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Version:   BaseVersion,
		GitCommit: BaseGitCommit,
		BuildDate: BaseBuildDate,
		GoVersion: BaseGoVersion,
	}
}
