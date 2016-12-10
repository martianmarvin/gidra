package http

import "github.com/martianmarvin/gidra/task"

func init() {
	task.Register("get", NewGet)
}

func NewGet() task.Task {
	t := &Task{
		Config: &Config{},
		Worker: task.NewWorker(),
	}
	t.Configurable = task.NewConfigurable(t.Config)
	t.Config.Method = []byte("GET")
	return t
}
