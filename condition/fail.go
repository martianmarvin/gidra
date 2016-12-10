package condition

import (
	"context"
	"text/template"
)

// Fail is like Abort, except it returns ErrFail instead of
// ErrAbort
// MET - callbacks, then return ErrFail
// NOT MET - nil

type Fail struct {
	*condition

	callbacks []CallBackFunc
}

func NewFail(callbacks ...CallBackFunc) Condition {
	return &Fail{
		condition: &condition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrFail,
			flag: After,
		},
		callbacks: callbacks,
	}
}

func (c *Fail) Check(ctx context.Context) error {
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
