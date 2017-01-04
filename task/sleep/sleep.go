package sleep

import (
	"context"
	"time"

	"github.com/martianmarvin/gidra/task"
)

// Sleeps for specified number of seconds
// Required Parameters:
// - seconds number of seconds to sleep

func init() {
	task.Register("sleep", New)
}

type Task struct {
	task.Worker
	task.Configurable

	Config *Config
}

type Config struct {
	Duration time.Duration `task:"required"`
}

func New() task.Task {
	t := &Task{
		Config: &Config{},
		Worker: task.NewWorker(),
	}
	t.Configurable = task.NewConfigurable(t.Config)
	return t
}

func (t *Task) Execute(ctx context.Context) error {
	t.Logger().WithField("duration", t.Config.Duration).Info("Sleeping...")
	time.Sleep(t.Config.Duration)
	return nil
}
