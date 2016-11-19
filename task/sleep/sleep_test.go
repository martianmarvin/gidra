package sleep

import (
	"testing"

	"github.com/martianmarvin/gidra/task"
)

func TestSleep(t *testing.T) {
	params := make(map[string]interface{})
	params["seconds"] = 5
	err := task.Run("sleep", params)
	if err != nil {
		t.Fatal(err)
	}
}
