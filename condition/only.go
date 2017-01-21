package condition

import (
	"context"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/template"
)

// Only is the opposite of Skip. The task is executed if it is met, and
// skipped otherwise
// MET - nil
// NOT MET - ErrSkip

type Only struct {
	*condition
}

func NewOnly() Condition {
	tmpl, _ := template.New("")
	return &Only{
		condition: &condition{
			tmpl: tmpl,
			err:  ErrSkip,
			flag: config.CondBefore,
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
