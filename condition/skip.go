package condition

import "text/template"

// Skip returns nil(so execution of the task continues) if it is not
// met
// MET - ErrSkip
// NOT MET - nil

type Skip struct {
	*condition
}

func NewSkip() Condition {
	return &Skip{
		condition: &condition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrSkip,
			flag: Before,
		},
	}
}

func (c *Skip) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		return ErrSkip
	} else {
		return nil
	}
}
