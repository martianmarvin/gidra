package sleep

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/task"
	"github.com/stretchr/testify/assert"
)

var taskCfg = `
duration: 5s
`

func TestSleep(t *testing.T) {
	r := strings.NewReader(taskCfg)
	cfg := config.Must(config.ParseYaml(r))
	t.Log(cfg.GetDuration("duration"))

	tsk := task.New("sleep")
	ctx := config.ToContext(context.Background(), cfg)
	before := time.Now()
	err := tsk.Execute(ctx)
	after := time.Now()
	assert.NoError(t, err)
	assert.WithinDuration(t, before.Add(5*time.Second), after, time.Second)
}
