package http

import "github.com/martianmarvin/gidra/task"

func init() {
	task.Register("get", NewGet)
}

func NewGet() task.Task {
	t := newHTTP()
	t.Config.Method = []byte("GET")
	return t
}
