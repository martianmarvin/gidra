package task

import (
	"context"
	"errors"

	"github.com/Sirupsen/logrus"

	"github.com/martianmarvin/vars"
)

// Task wraps tasks of different types
type task struct {
	task   Task
	name   string
	logger logrus.FieldLogger
}

func (t *task) Execute(ctx context.Context) error {
	var err error

	if l, ok := t.task.(Loggable); ok {
		l.SetLogger(t.logger)
	}

	if c, ok := t.task.(Configurable); ok {
		err = configureTask(ctx, c)
		if err != nil {
			t.logger.Error(err)
			return err
		}
	}

	t.logger.Info("Starting...")
	t.task.Execute(ctx)
	t.logger.Info("Done")

	return err
}

func configureTask(ctx context.Context, t Configurable) error {
	taskVars, ok := vars.FromContext(ctx)
	if !ok {
		return errors.New("No vars found on context")
	}
	return t.Configure(taskVars)
}
