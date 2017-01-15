package condition

import (
	"context"

	"github.com/martianmarvin/gidra/template"
)

// Skip returns nil(so execution of the task continues) if it is not
// met
// MET - ErrSkip
// NOT MET - nil

type Skip struct {
	*condition
}

func NewSkip() Condition {
	tmpl, _ := template.New("")
	return &Skip{
		condition: &condition{
			tmpl: tmpl,
			err:  ErrSkip,
			flag: Before,
		},
	}
}

func (c *Skip) Check(ctx context.Context) error {
	if c.check(ctx) {
		return c.err
	} else {
		return nil
	}
}
