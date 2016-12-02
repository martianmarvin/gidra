package log

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/martianmarvin/gidra/config"
)

var Logger *logrus.Entry

func init() {
	initLogger()
}

func initLogger() {
	logrus.SetOutput(os.Stderr)
	Logger = logrus.WithField("task", "gidra")
	Logger.Level = config.Verbosity
}
