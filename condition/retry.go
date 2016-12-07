package condition

import "text/template"

// Retry keeps track of retries and attempts the task again as long as
// less than RetryLimit attempts have been made. It returns ErrRetry when the
// maximum number of attempts has not been made, and ErrFail otherwise,
// indicating that the task has failed, and no more attempts should be made
// MET - ErrRetry up to RetryLimit times, then ErrFail
// NOT MET - nil

type Retry struct {
	*condition

	// The current attempt being made
	i int

	// RetryLimit is the maximum number of times the task should be retried
	RetryLimit int
}

func NewRetry(limit int) Condition {
	if limit <= 0 {
		limit = DefaultRetryLimit
	}
	return &Retry{
		condition: &condition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrRetry,
		},
		RetryLimit: limit,
	}
}

func (c *Retry) Check(vars map[string]interface{}) error {
	if c.isMet(vars) {
		c.i += 1
		if c.i <= c.RetryLimit {
			return ErrRetry
		} else {
			return ErrFail
		}
	} else {
		return nil
	}
}
