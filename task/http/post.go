package http

import "github.com/martianmarvin/gidra/task"

func init() {
	task.Register("post", NewPost)
}

func NewPost() task.Task {
	return &Task{
		BaseTask: task.BaseTask{},
		Config: &Config{
			Method: []byte("POST"),
		},
	}
}
