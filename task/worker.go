package task

import (
	"context"

	"github.com/martianmarvin/gidra/log"

	"github.com/Sirupsen/logrus"
	"github.com/martianmarvin/vars"
)

type worker struct {
	id int

	logger logrus.FieldLogger

	//container of task-local variables
	taskVars *vars.Vars

	requiredVars []string
}

// Worker is the basic Task type that most tasks should include. It
// encapsulates standard methods shared by most tasks
type Worker interface {
	// Id returns the number of this task in the sequence
	Id() int

	// Logger returns this task's logger
	Logger() logrus.FieldLogger

	// SetLogger attaches a logger to this task
	SetLogger(logrus.FieldLogger)

	// Vars give access to this worker's local vars
	Vars() *vars.Vars
}

// NewWorker returns a new Worker
func NewWorker() Worker {
	return &worker{
		logger:       log.Logger,
		taskVars:     vars.New(),
		requiredVars: make([]string, 0),
	}
}

// Execute wraps the other task's Execute function
func (w *worker) execute(fn ExecFunc) ExecFunc {
	return func(ctx context.Context) {
		fn(ctx)
	}
}

func (w *worker) Id() int {
	return w.id
}

func (w *worker) SetLogger(logger logrus.FieldLogger) {
	w.logger = logger
}

func (w *worker) Logger() logrus.FieldLogger {
	return w.logger
}

func (w *worker) Vars() *vars.Vars {
	return w.taskVars
}
