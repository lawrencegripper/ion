//nolint: golint

package env

import (
	"github.com/lawrencegripper/ion/modules/helpers/Go/log"
	"os"
	"path"
)

var (
	IonBaseDir         = "/ion"
	InputDataDir       = func() string { return path.Join(IonBaseDir, "in", "data") }
	InputEventMetaFile = func() string { return path.Join(IonBaseDir, "in", "eventmeta.json") }
	OutputDataDir      = func() string { return path.Join(IonBaseDir, "out", "data") }
	EventDir           = func() string { return path.Join(IonBaseDir, "out", "events") }
	InsightFile        = func() string { return path.Join(IonBaseDir, "out", "insights.json") }
)

func init() {
	baseDirEnv := os.Getenv("HANDLER_BASE_DIR")
	if baseDirEnv == "" {
		log.Info("HANDLER_BASE_DIR environment variable not setting using default /ion as base dir")
	} else {
		IonBaseDir = baseDirEnv
	}

}

func MakeOutputDirs() {
	err := os.MkdirAll(OutputDataDir(), 0777)
	if err != nil {
		panic("Failed making output data dir")
	}
	err = os.MkdirAll(EventDir(), 0777)
	if err != nil {
		panic("Failed making output event dir")
	}
	err = os.MkdirAll(InsightFile(), 0777)
	if err != nil {
		panic("Failed making output insights dir")
	}
}
