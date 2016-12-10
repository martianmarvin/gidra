package task

import "github.com/martianmarvin/vars"

type worker struct {
	Loggable

	//container of task-local variables
	taskVars *vars.Vars

	requiredVars []string
}

// Worker is the basic Task type that most tasks should include. It
// encapsulates standard methods shared by most tasks
type Worker interface {
	Loggable

	// Vars give access to this worker's local vars
	Vars() *vars.Vars
}

// NewWorker returns a new Worker
func NewWorker() Worker {
	return &worker{
		Loggable:     NewLoggable(),
		taskVars:     vars.New(),
		requiredVars: make([]string, 0),
	}
}

func (w *worker) Vars() *vars.Vars {
	return w.taskVars
}
