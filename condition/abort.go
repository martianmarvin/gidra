package condition

import "text/template"

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
	return &Abort{
		condition: &condition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrAbort,
			flag: After,
		},
		callbacks: callbacks,
	}
}

func (c *Abort) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		for _, cb := range c.callbacks {
			if err := cb(); err != nil {
				break
			}
		}
		return c.err
	} else {
		return nil
	}
}
