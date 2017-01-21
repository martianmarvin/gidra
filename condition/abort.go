package condition

import (
	"context"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/template"
)

// Abort returns ErrAbort if it is met, optionally executing a series of
// callbacks before panicking. If any of the callbacks return an error, the
// execution chain is halted and the script panics immediately
// met
// MET - callbacks, then return ErrAbort
// NOT MET - nil

type Abort struct {
	*condition

	callbacks []CallBackFunc
}

func NewAbort(callbacks ...CallBackFunc) Condition {
	tmpl, _ := template.New("")
	return &Abort{
		condition: &condition{
			tmpl: tmpl,
			err:  ErrAbort,
			flag: config.CondAfter,
		},
		callbacks: callbacks,
	}
}

func (c *Abort) Check(ctx context.Context) error {
	if c.check(ctx) {
		for _, cb := range c.callbacks {
			if err := cb(ctx); err != nil {
				break
			}
		}
		return c.err
	} else {
		return nil
	}
}
