package condition

import (
	"context"
	"text/template"
)

// Retry keeps track of retries and attempts the task again as long as
// less than RetryLimit attempts have been made. Optionally, callbacks are
// executed if ErrRetry will be returned. It returns ErrRetry when the
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

	callbacks []CallBackFunc
}

func NewRetry(limit int, callbacks ...CallBackFunc) Condition {
	if limit <= 0 {
		limit = DefaultRetryLimit
	}
	return &Retry{
		condition: &condition{
			tmpl: template.New("").Option("missingkey=zero"),
			err:  ErrRetry,
			flag: After,
		},
		RetryLimit: limit,
		callbacks:  callbacks,
	}
}

func (c *Retry) Check(ctx context.Context) error {
	if c.check(ctx) {
		c.i += 1
		if c.i <= c.RetryLimit {
			for _, cb := range c.callbacks {
				if err := cb(ctx); err != nil {
					break
				}
			}
			return ErrRetry
		} else {
			return ErrFail
		}
	} else {
		return nil
	}
}
