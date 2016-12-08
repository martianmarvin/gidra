package condition

import (
	"context"
	"text/template"
)

// Only is the opposite of Skip. The task is executed if it is met, and
// skipped otherwise
// MET - nil
// NOT MET - ErrSkip

type Only struct {
	*condition
}

func NewOnly() Condition {
	return &Only{
		condition: &condition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrSkip,
			flag: Before,
		},
	}
}

func (c *Only) Check(ctx context.Context) error {
	if c.check(ctx) {
		return nil
	} else {
		return c.err
	}
}
