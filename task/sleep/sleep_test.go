package sleep

import (
	"testing"

	"github.com/martianmarvin/gidra/task"
)

func TestSleep(t *testing.T) {
	params := make(map[string]interface{})
	params["seconds"] = 5
	tsk := task.New("sleep")
	err := tsk.Execute(params)
	t.Log(tsk)
	if err != nil {
		t.Fatal(err)
	}
}
