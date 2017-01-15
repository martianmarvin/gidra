// Print prints out text to the configured output source
package print

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/martianmarvin/gidra/task"
)

func init() {
	task.Register("print", New)
}

type Task struct {
	task.Worker
	task.Configurable

	Config *Config
}

type Config struct {
	// Text is any valid golang template in string form
	Text string `task:"required"`

	// Path to write to, or '-' for STDOUT
	Path string
}

func New() task.Task {
	t := &Task{
		Config: &Config{},
		Worker: task.NewWorker(),
	}
	t.Configurable = task.NewConfigurable(t.Config)
	return t
}

func (t *Task) Execute(ctx context.Context) error {
	var err error
	var w io.Writer
	if len(t.Config.Path) == 0 || t.Config.Path == "-" {
		w = os.Stdout
	} else {
		w, err = os.OpenFile(t.Config.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w, t.Config.Text)
	return err

}
