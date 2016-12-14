package sleep

import (
	"context"
	"testing"
	"time"

	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
	"github.com/stretchr/testify/assert"
)

func TestSleep(t *testing.T) {
	tsk := task.New("sleep")
	ctx := vars.ToContext(context.Background(), vars.New())
	ctx = vars.SetCtx(ctx, "seconds", 5)
	before := time.Now()
	err := tsk.Execute(ctx)
	after := time.Now()
	assert.NoError(t, err)
	assert.WithinDuration(t, before.Add(5*time.Second), after, time.Second)
}
