package http

import (
	"testing"

	"github.com/martianmarvin/gidra/task"
)

func TestGet(t *testing.T) {
	c := NewHTTPClient()
	params := make(map[string]interface{})
	params["url"] = "http://httpbin.org/get"
	// params["method"] = "GET"
	params["headers"] = map[string]string{
		"user-agent": "Gidra",
		"h1":         "v1",
		"h2":         "v2",
	}
	tsk := task.New("get")
	err := tsk.Execute(c, params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(c.Response().Body()))
}

func TestPost(t *testing.T) {
	c := NewHTTPClient()
	params := make(map[string]interface{})
	params["url"] = "http://httpbin.org/post"
	params["method"] = "POST"
	params["headers"] = map[string]string{
		"user-agent": "Gidra",
		"h1":         "v1",
		"h2":         "v2",
	}
	params["cookies"] = map[string]string{
		"c1": "v1",
		"c2": "v2",
	}
	params["params"] = map[string]string{
		"b1": "v1",
		"b2": "b2",
	}
	tsk := task.New("post")
	err := tsk.Execute(c, params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(c.Response().Body()))
}
