package task

import (
	"context"
	"fmt"
	"time"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/log"
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
	// TODO refactor to avoid using config
	cfg := config.FromContext(ctx)
	timeout, err := cfg.GetDurationE(cfgTimeout)
	if err != nil || timeout < time.Second {
		timeout = defaultTimeout
	}

	ctx = log.ToContext(ctx, t.logger)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Apply this context to task
	err = t.configure(ctx)
	if err != nil {
		t.logger.Error(err)
		return err
	}

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

// Print the task configured with the given context
func (t *task) show(ctx context.Context) string {
	var fields string
	t.configure(ctx)
	if c, ok := t.task.(Configurable); ok {
		fields = c.String()
	}
	return fmt.Sprintf("%s: %s", t.name, fields)
}

func (t *task) configure(ctx context.Context) error {
	if l, ok := t.task.(Loggable); ok {
		l.SetLogger(t.logger)
	}

	// Populate task's config
	if c, ok := t.task.(Configurable); ok {
		err := c.Configure(config.FromContext(ctx))
		if err != nil {
			return err
		}
	}
	return nil
}
