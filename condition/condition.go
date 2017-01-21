// Package condition evaluates the success/failure outcome of tasks, as well
// as whether a task should be executed or skipped
package condition

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/global"
	"github.com/martianmarvin/gidra/template"
)

var (
	ErrInvalidCondition = errors.New("Could not parse condition. A condition must be a valid Go template, including the leading {{ and trailing }} brackets")

	// Kill the entire script. Similar to a panic
	ErrAbort = errors.New("Aborting entire script now")

	// Repeat the same step again
	ErrRetry = errors.New("Task temporary failure, retrying")

	// Mark the task as failed, and continue to the next iteration of the loop
	ErrFail = errors.New("Task failed, moving on")

	// Skip this task, and advance to the next one in the sequence
	ErrSkip = errors.New("Skipping task")

	DefaultRetryLimit = 5
)

// CallBackFunc represents an optional function to call if condition is met
type CallBackFunc func(ctx context.Context) error

// Condition acts as a gate during the execution of a Sequence that determines
// whether the sequence should continue to the next step, quit execution, or try
// the same step again
// The type of error returned by the Condition on success signals to the
// Sequence executing it what it should do next. For example, returning ErrRetry
// will cause the same step to be repeated again.
type Condition interface {
	// Check evaluates the Condition with provided variables and returns nil(for success) or an error
	Check(ctx context.Context) error

	// Parse parses and compiles a template.Template to validate this condition.
	// The template is concatenated with an if pipeline to coerce the result to
	// boolean
	Parse(cond string) error

	// Flags returns the execution flags that determine how this condition
	// should run
	Flags() *config.Flag
}

//Returns a slice containing a single condition
func Default() []Condition {
	conds := make([]Condition, 0)
	conds = append(conds, New())
	return conds
}

// The basic Condition that is common to all conditions. Its Check() method is a no-op that
// always returns no error
// Since a Task is only executed if at least one condition returns nil(no
// errors), all tasks with no conditions specified have a condition by
// default
type condition struct {
	// The original template text for this condition
	cond string

	// The compiled template used to evaluate this Condition
	tmpl *template.Template

	// The type of error to return if the condition is NOT MET
	err error

	// Config flags
	flag config.Flag
}

// New returns a new empty condition
func New() Condition {
	tmpl, _ := template.New("")
	return &condition{tmpl: tmpl}
}

func (c *condition) Parse(cond string) error {
	var err error
	if !strings.HasPrefix(cond, "{{") || !strings.HasSuffix(cond, "}}") {
		return ErrInvalidCondition
	}

	// Wrap in conditional
	cond = strings.TrimPrefix(strings.TrimSpace(cond), "{{")
	cond = strings.TrimSpace(strings.TrimSuffix(cond, "}}"))
	cond = fmt.Sprintf("{{ if %s }} true {{end}}", cond)

	// Create new clone template
	tmpl, err := template.New(cond)
	if err != nil {
		return err
	}

	if err == nil {
		c.cond = cond
		c.tmpl = tmpl
	}
	return err
}

// Checks if condition is met by executing the template using the provided data
// as its context
func (c *condition) isMet(data interface{}) bool {
	if len(c.cond) == 0 {
		return true
	}
	res, err := c.tmpl.Execute(data)
	if err != nil {
		panic(fmt.Sprintf("Condition '%s' died with error %s", c.cond, err.Error()))
	}

	if len(res) > 0 {
		// template evaluates to non-empty response, condition is MET
		return true
	} else {
		// template evaluates to empty, condition NOT MET
		return false
	}
}

// Common checker, determines what from context gets passed to isMet
func (c *condition) check(ctx context.Context) bool {
	return c.isMet(global.FromContext(ctx))

}

func (c *condition) Check(ctx context.Context) error {
	if c.check(ctx) {
		return nil
	} else {
		return c.err
	}
}

func (c *condition) Flags() *config.Flag {
	return &c.flag
}
