package http

import "github.com/martianmarvin/gidra/task"

func init() {
	task.Register("post", NewPost)
}

func NewPost() task.Task {
	t := &Task{
		Config: &Config{},
		Worker: task.NewWorker(),
	}
	t.Configurable = task.NewConfigurable(t.Config)
	t.Config.Method = []byte("POST")
	return t
}
