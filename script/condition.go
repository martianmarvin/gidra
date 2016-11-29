package script

import (
	"errors"
	"strings"
	"text/template"
)

var (
	ErrInvalidCondition = errors.New("Could not parse condition. A condition must be a valid Go template, including the leading {{ and trailing }} brackets")

	// Appended to template conditions to coerce them to boolean
	condSuffix = ` | if }}true{{end}}`
)

// Condition acts as a gate during the execution of a Sequence that determines
// whether the sequence should continue to the next step, quit execution, or try
// the same step again
type Condition interface {
	// Check evaluates the Condition with provided variables and returns nil(for success) or an error
	Check(vars map[string]interface{}) error

	// Parse parses and compiles a template.Template to validate this condition.
	// The template is concatenated with an if pipeline to coerce the result to
	// boolean
	Parse(cond string) error
}

// BaseCondition is common to all conditions. Its Check() method is a no-op that
// always returns no error
type BaseCondition struct {
	// The original template text for this condition
	cond string

	// The compiled template used to evaluate this Condition
	tmpl *template.Template
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

func (c *BaseCondition) Check(vars map[string]interface{}) error {
	return nil
}
