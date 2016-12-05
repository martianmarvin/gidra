package sleep

import (
	"context"
	"testing"

	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
)

func TestSleep(t *testing.T) {
	tsk := task.New("sleep")
	ctx := vars.ToContext(context.Background(), vars.New())
	ctx = vars.SetCtx(ctx, "seconds", 5)
	err := tsk.Execute(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
