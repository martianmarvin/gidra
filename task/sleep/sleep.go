package sleep

import (
	"time"

	"github.com/martianmarvin/gidra/task"
	"github.com/mitchellh/mapstructure"
)

// Sleeps for specified number of seconds
// Required Parameters:
// - seconds number of seconds to sleep

func init() {
	task.Register("sleep", NewTask)
}

type Task struct {
	task.BaseTask

	Config *Config
}

type Config struct {
	Seconds time.Duration
}

func NewTask() task.Task {
	return &Task{
		BaseTask: task.BaseTask{},
		Config:   &Config{},
	}
}

func (t *Task) Execute(vars map[string]interface{}) (err error) {
	if err = mapstructure.Decode(vars, t.Config); err != nil {
		return
	}
	time.Sleep(t.Config.Seconds * time.Second)

	return err
}
