package http

import "github.com/martianmarvin/gidra/task"

func init() {
	task.Register("get", NewGet)
}

func NewGet() task.Task {
	return &Task{
		BaseTask: task.BaseTask{},
		Config: &Config{
			Method: []byte("GET"),
		},
	}
}
