package http

import "github.com/martianmarvin/gidra/task"

func init() {
	task.Register("post", NewPost)
}

func NewPost() task.Task {
	t := newHTTP()
	t.Config.Method = []byte("POST")
	return t
}
