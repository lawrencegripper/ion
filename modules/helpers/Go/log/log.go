package log

import (
	"os"

	logrus "github.com/sirupsen/logrus"
)

var DebugMode = false

func init() {
	if os.Getenv("MODULE_DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
		DebugMode = true
	}
}

func Debug(msg string) {
	logrus.Debug(msg)
}

func Fatal(msg string) {
	logrus.Fatal(msg)
}

func Info(msg string) {
	logrus.Info(msg)
}
