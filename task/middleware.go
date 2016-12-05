package task

import (
	"context"
	"errors"
	"time"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/log"
	"github.com/martianmarvin/vars"
)

var (
	cfgTimeout = "config.task_timeout"
)

// Default
var (
	defaultTimeout time.Duration = 30 * time.Second
)

// Task wraps tasks of different types
type task struct {
	task   Task
	name   string
	logger log.Log
}

func (t *task) Execute(ctx context.Context) error {
	var err error

	if ctx == nil {
		panic("nil context")
	}

	if t.logger == nil {
		t.logger = log.FromContext(ctx).WithField("task", t.name)
	}

	// Read global config
	cfg := config.FromContext(ctx)
	timeout := time.Duration(cfg.UInt(cfgTimeout, 0)) * time.Second
	if timeout < time.Second {
		timeout = defaultTimeout
	}

	ctx = log.ToContext(ctx, t.logger)

	if l, ok := t.task.(Loggable); ok {
		l.SetLogger(t.logger)
	}

	// Populate task's config
	if c, ok := t.task.(Configurable); ok {
		err = configureTask(ctx, c)
		if err != nil {
			t.logger.Error(err)
			return err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute task supporting cancellation via context
	done := make(chan error)
	go func() {
		t.logger.Info("Starting...")

		done <- t.task.Execute(ctx)

		t.logger.Info("Done")
	}()

	// Block waiting for completion or cancellation
	for {
		select {
		case err = <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func configureTask(ctx context.Context, t Configurable) error {
	taskVars, ok := vars.FromContext(ctx)
	if !ok {
		return errors.New("No vars found on context")
	}
	return t.Configure(taskVars)
}
