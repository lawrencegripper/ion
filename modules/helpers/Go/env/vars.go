package env

import (
	"github.com/lawrencegripper/ion/modules/helpers/Go/log"
	logrus "github.com/sirupsen/logrus"
	"os"
	"path"
)

var (
	//IonBaseDir The base directory for ion files
	IonBaseDir = "/ion"
	//InputDataDir holds blob data from the parent module
	InputDataDir = func() string { return path.Join(IonBaseDir, "in", "data") }
	//InputEventMetaFile holds meta data from the parent module
	InputEventMetaFile = func() string { return path.Join(IonBaseDir, "in", "eventmeta.json") }
	//OutputDataDir is used to store blob data outputted by this module
	OutputDataDir = func() string { return path.Join(IonBaseDir, "out", "data") }
	//EventDir is used to store events this module will raise
	EventDir = func() string { return path.Join(IonBaseDir, "out", "events") }
	//InsightFile is used to track insights the module has identified, for example a person seen or a object identified.
	// these are tracked against the correlationID and can be queried.
	InsightFile = func() string { return path.Join(IonBaseDir, "out", "insights.json") }
)

func init() {
	baseDirEnv := os.Getenv("HANDLER_BASE_DIR")
	if baseDirEnv == "" {
		log.Info("HANDLER_BASE_DIR environment variable not setting using default /ion as base dir")
	} else {
		IonBaseDir = baseDirEnv
	}

}

//MakeOutputDirs creates all the output dirs that an ion module can use.
func MakeOutputDirs() {
	err := os.MkdirAll(OutputDataDir(), 0777)
	if err != nil {
		logrus.WithError(err).Fatal("Failed making output data dir")
		os.Exit(13)
	}
	err = os.MkdirAll(EventDir(), 0777)
	if err != nil {
		logrus.WithError(err).Fatal("Failed making output event dir")
		os.Exit(13)
	}
}
