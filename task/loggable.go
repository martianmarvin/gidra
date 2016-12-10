package task

import (
	"github.com/martianmarvin/gidra/log"
)

type loggable struct {
	logger log.Log
}

// A Loggable task is able to log its own logger
type Loggable interface {
	// Logger returns this task's logger
	Logger() log.Log

	// SetLogger attaches a logger to this task
	SetLogger(log.Log)
}

// NewLoggable initializes a new loggable task with the default global logger
func NewLoggable() Loggable {
	return &loggable{
		logger: log.Logger(),
	}
}

func (l *loggable) SetLogger(logger log.Log) {
	l.logger = logger
}

func (l *loggable) Logger() log.Log {
	return l.logger
}
