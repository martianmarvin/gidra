package sleep

import (
	"fmt"
	"time"

	"github.com/martianmarvin/gidra/task"
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
	Duration time.Duration `task:"seconds,required"`
}

func NewTask() task.Task {
	return &Task{
		BaseTask: task.BaseTask{},
		Config:   &Config{},
	}
}

func (t *Task) Execute(vars map[string]interface{}) (err error) {
	if err = task.Configure(t, vars); err != nil {
		return err
	}
	t.Config.Duration *= time.Second
	fmt.Println(t.Config)
	time.Sleep(t.Config.Duration)

	return err
}
