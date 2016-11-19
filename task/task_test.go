package task

import (
	"testing"

	_ "github.com/martianmarvin/gidra/task/all"
)

func TestTaskList(t *testing.T) {
	tasks := Tasks()
	t.Log(tasks)
	if len(tasks) == 0 {
		t.Fatal("No registerered tasks")
	}

}
