package condition

import "text/template"

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
		},
	}
}

func (c *Only) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		return nil
	} else {
		return c.err
	}
}
