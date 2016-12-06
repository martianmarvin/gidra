package sequence

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
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

	// Appended to template conditions to coerce them to boolean
	condSuffix = ` | if }}true{{end}}`

	DefaultRetryLimit = 5
)

//Output a template into a string
func stringTemplate(t *template.Template, data interface{}) (text string, err error) {
	var b bytes.Buffer
	err = t.Execute(&b, data)
	if err != nil {
		return
	}
	text = b.String()
	return text, err
}

// CallBackFunc represents an optional function to call if condition is met
type CallBackFunc func() error

// Condition acts as a gate during the execution of a Sequence that determines
// whether the sequence should continue to the next step, quit execution, or try
// the same step again
// The type of error returned by the Condition on success signals to the
// Sequence executing it what it should do next. For example, returning ErrRetry
// will cause the same step to be repeated again.
type Condition interface {
	// Check evaluates the Condition with provided variables and returns nil(for success) or an error
	Check(vars map[string]interface{}) error

	// Parse parses and compiles a template.Template to validate this condition.
	// The template is concatenated with an if pipeline to coerce the result to
	// boolean
	Parse(cond string) error
}

//Returns a slice containing a single BaseCondition
func defaultConditions() []Condition {
	conds := make([]Condition, 0)
	conds = append(conds, NewBaseCondition())
	return conds
}

// BaseCondition is common to all conditions. Its Check() method is a no-op that
// always returns no error
// Since a Task is only executed if at least one condition returns nil(no
// errors), all tasks with no conditions specified have a BaseCondition by
// default
type BaseCondition struct {
	// The original template text for this condition
	cond string

	// The compiled template used to evaluate this Condition
	tmpl *template.Template

	// The type of error to return if the condition is NOT MET
	err error
}

func NewBaseCondition() *BaseCondition {
	return &BaseCondition{tmpl: template.New("").Option("missingkey=zero")}
}

func (c *BaseCondition) Parse(cond string) error {
	var err error
	if !strings.HasPrefix(cond, "{{") || !strings.HasSuffix(cond, "}}") {
		return ErrInvalidCondition
	}

	// Append if pipeline to the end
	cond = strings.TrimSuffix(strings.TrimSpace(cond), "}}") + condSuffix
	c.tmpl, err = c.tmpl.Parse(cond)
	if err == nil {
		c.cond = cond
	}
	return err
}

// Checks if condition is met
func (c *BaseCondition) isMet(vars map[string]interface{}) bool {
	if len(c.cond) == 0 {
		return true
	}
	res, err := stringTemplate(c.tmpl, vars)
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

func (c *BaseCondition) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		return nil
	} else {
		return c.err
	}
}

// SuccessCondition returns nil if it is met, and an error otherwise
// MET - nil
// NOT MET - ErrFail
type SuccessCondition struct {
	*BaseCondition
}

func NewSuccessCondition() *SuccessCondition {
	return &SuccessCondition{
		BaseCondition: &BaseCondition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrFail,
		},
	}
}

// SkipCondition returns nil(so execution of the task continues) if it is not
// met
// MET - ErrSkip
// NOT MET - nil
type SkipCondition struct {
	*BaseCondition
}

func NewSkipCondition() *SkipCondition {
	return &SkipCondition{
		BaseCondition: &BaseCondition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrSkip,
		},
	}
}

func (c *SkipCondition) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		return ErrSkip
	} else {
		return nil
	}
}

// AbortCondition panics if it is met, optionally executing a series of
// callbacks before panicking. If any of the callbacks return an error, the
// execution chain is halted and the script panics immediately
// met
// MET - callbacks, then panic
// NOT MET - nil

type AbortCondition struct {
	*BaseCondition

	callbacks []CallBackFunc
}

func NewAbortCondition(callbacks ...CallBackFunc) *AbortCondition {
	return &AbortCondition{
		BaseCondition: &BaseCondition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrAbort,
		},
		callbacks: callbacks,
	}
}

func (c *AbortCondition) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		for _, cb := range c.callbacks {
			err := cb()
			if err != nil {
				break
			}
		}
		panic(c.err)
	} else {
		return nil
	}
}

// RetryCondition keeps track of retries and attempts the task again as long as
// less than RetryLimit attempts have been made. It returns ErrRetry when the
// maximum number of attempts has not been made, and ErrFail otherwise,
// indicating that the task has failed, and no more attempts should be made
// MET - ErrRetry up to RetryLimit times, then ErrFail
// NOT MET - nil

type RetryCondition struct {
	*BaseCondition

	// The current attempt being made
	i int

	// RetryLimit is the maximum number of times the task should be retried
	RetryLimit int
}

func NewRetryCondition(limit int) *RetryCondition {
	if limit <= 0 {
		limit = DefaultRetryLimit
	}
	return &RetryCondition{
		BaseCondition: &BaseCondition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrRetry,
		},
		RetryLimit: limit,
	}
}

func (c *RetryCondition) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		c.i += 1
		if c.i <= c.RetryLimit {
			return ErrRetry
		} else {
			return ErrFail
		}
	} else {
		return nil
	}
}

// FailCondition is like AbortCondition, except it returns ErrFail instead of
// panicking
// MET - callbacks, then return ErrFail
// NOT MET - nil

type FailCondition struct {
	*BaseCondition

	callbacks []CallBackFunc
}

func NewFailCondition(callbacks ...CallBackFunc) *FailCondition {
	return &FailCondition{
		BaseCondition: &BaseCondition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrFail,
		},
		callbacks: callbacks,
	}
}

func (c *FailCondition) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		for _, cb := range c.callbacks {
			err := cb()
			if err != nil {
				break
			}
		}
		return c.err
	} else {
		return nil
	}
}
