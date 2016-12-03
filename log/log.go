package log

import (
	"context"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/martianmarvin/gidra/config"
)

var Logger *logrus.Entry

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
	Logger = logrus.WithField("task", "gidra")
	Logger.Level = config.Verbosity
}

// Attaches the given logger to this context
func WithContext(ctx context.Context, logger logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, ctxLogger, logger)
}

// Returns the logger attached to this context
func FromContext(ctx context.Context) (logrus.FieldLogger, bool) {
	l, ok := ctx.Value(ctxLogger).(logrus.FieldLogger)
	return l, ok
}
