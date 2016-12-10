package log

import (
	"context"
	"os"

	"github.com/Sirupsen/logrus"
)

// The default global log
var logger *logrus.Entry

var (
	// Default global log level
	defaultLevel = logrus.DebugLevel

	// Config key that has log level
	cfgLogLevel = "verbosity"
)

type Log interface {
	logrus.FieldLogger
}

// Context key
type contextKey int

const (
	ctxLogger contextKey = iota
)

func init() {
	initLogger()
}

func initLogger() {
	logrus.SetOutput(os.Stderr)
	logger = logrus.WithField("task", "gidra")
}

// SetLevel sets the global log level
func SetLevel(lvl logrus.Level) {
	if lvl > 0 && lvl <= 5 {
		logrus.SetLevel(lvl)
	}
}

// Logger returns the default global logger
func Logger() Log {
	return logger
}

// ToContext attaches the given logger to this context
func ToContext(ctx context.Context, logger Log) context.Context {
	return context.WithValue(ctx, ctxLogger, logger)
}

// FromContext returns the logger attached to this context
func FromContext(ctx context.Context) Log {
	l, ok := ctx.Value(ctxLogger).(Log)
	if !ok {
		l = Logger()
	}
	return l
}
