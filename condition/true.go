package condition

import (
	"context"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/template"
)

// True executes its callbacks and always returns nil error
// MET - callbacks, then return nil
// NOT MET - callbacks, then return nil

type True struct {
	*condition

	callbacks []CallBackFunc
}

func NewTrue(callbacks ...CallBackFunc) Condition {
	tmpl, _ := template.New("true")
	return &True{
		condition: &condition{
			tmpl: tmpl,
			err:  nil,
			flag: config.CondBefore,
		},
		callbacks: callbacks,
	}
}

func (c *True) Check(ctx context.Context) error {
	for _, cb := range c.callbacks {
		if err := cb(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *True) Parse(cond string) error {
	return nil
}
