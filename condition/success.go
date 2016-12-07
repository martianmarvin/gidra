package condition

import "text/template"

// Success returns nil if it is met, and an error otherwise
// MET - nil
// NOT MET - ErrFail
type Success struct {
	*condition
}

func NewSuccess() Condition {
	return &Success{
		condition: &condition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrFail,
		},
	}
}
