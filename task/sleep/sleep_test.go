package sleep

import (
	"context"
	"testing"

	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
)

func TestSleep(t *testing.T) {
	params := make(map[string]interface{})
	params["seconds"] = 5
	tsk := task.New("sleep")
	ctx := vars.WithContext(context.Background(), vars.New())
	ctx = vars.WithVar(ctx, "seconds", 5)
	tsk.Execute(ctx)
}
