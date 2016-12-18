package sequence

import (
	"context"
	"fmt"

	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/vars"
)

var defaultOutputVars = []string{"extracted"}

// Result is the result of running a sequence
type Result struct {
	// List of errors from tasks(not including retries)
	Errors []error

	// A row of output from this task
	Output *datasource.Row

	// Names of vars which should join the output
	outputVars []string
}

// NewResult initializes a new, empty result
func NewResult() *Result {
	return &Result{
		Errors:     make([]error, 0),
		outputVars: defaultOutputVars,
		Output:     datasource.NewRow(),
	}
}

// Success indicates whether the sequence completed successfully
func (r *Result) Success() bool {
	return (r.Output != nil && len(r.Errors) == 0)
}

// Err returns the most recent error
func (r *Result) Err() error {
	if len(r.Errors) == 0 {
		return nil
	}
	return r.Errors[len(r.Errors)-1]
}

// ReadContext populates this result's Output with the output vars found in the
// context and returns a list of variables that were saved
func (r *Result) ReadContext(ctx context.Context) []string {
	taskVars := vars.FromContext(ctx)
	return r.ReadVars(taskVars)
}

// ReadVars populates this result's Output with the provided *vars.Vars
// and returns a list of variables that were saved
func (r *Result) ReadVars(taskVars *vars.Vars) []string {
	found := make([]string, 0)

	for _, key := range r.outputVars {
		if taskVar, ok := taskVars.CheckGet(key); ok {
			if list, err := taskVar.Array(); err == nil {
				for i, v := range list {
					k := fmt.Sprintf("%s.%d", key, i)
					r.Output.AppendKV(k, v)
				}
			} else if m := taskVar.Map(); len(m) > 0 {
				for k, v := range m {
					k = fmt.Sprintf("%s.%s", key, k)
					r.Output.AppendKV(k, v)
				}
			} else if s, err := taskVar.String(); err == nil {
				r.Output.AppendKV(key, s)
			} else {
				continue
			}

			found = append(found, key)
		}
	}

	return found
}

// Keep adds the specified key to the list of variables the result reads into
// its output. It returns the result instance, so it is chainable
func (r *Result) Keep(key string) *Result {
	r.outputVars = append(r.outputVars, key)
	return r
}
