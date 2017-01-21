package condition

import (
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/template"
)

// Success returns nil if it is met, and an error otherwise
// MET - nil
// NOT MET - ErrFail
type Success struct {
	*condition
}

func NewSuccess() Condition {
	tmpl, _ := template.New("")
	return &Success{
		condition: &condition{
			tmpl: tmpl,
			err:  ErrFail,
			flag: config.CondAfter,
		},
	}
}
