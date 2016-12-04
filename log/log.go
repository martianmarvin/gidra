package log

import (
	"context"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/martianmarvin/gidra/config"
)

// The default global log
var logger *logrus.Entry

type Log interface {
	logrus.FieldLogger
}

// Context key
type key int

const (
	ctxLogger key = iota
)

func init() {
	initLogger()
}

func initLogger() {
	logrus.SetOutput(os.Stderr)
	logger = logrus.WithField("task", "gidra")
	logger.Level = config.Verbosity
}

// Logger returns the default global logger
func Logger() Log {
	return logger
}

// WithContext attaches the given logger to this context
func WithContext(ctx context.Context, logger Log) context.Context {
	return context.WithValue(ctx, ctxLogger, logger)
}

// FromContext returns the logger attached to this context
func FromContext(ctx context.Context) (Log, bool) {
	l, ok := ctx.Value(ctxLogger).(Log)
	return l, ok
}
