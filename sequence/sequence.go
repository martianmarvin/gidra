package sequence

import (
	"context"
	"fmt"

	"github.com/martianmarvin/gidra/condition"
	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
)

// Sequence is a series of tasks that represent a single iteration of the loop
type Sequence struct {
	// The number of the current loop iteration
	Id int

	//Tasks is the list of tasks in this sequence
	Tasks []task.Task

	//The sequence id of the current task
	n int

	//List of conditions corresponding to tasks in the sequence
	conditions [][]condition.Condition

	//Context shared by all requests in this sequence
	ctx context.Context

	cancel context.CancelFunc
}

func New() *Sequence {
	s := &Sequence{
		Tasks:      make([]task.Task, 0),
		conditions: make([][]condition.Condition, 0),
	}
	s.ctx, s.cancel = context.WithCancel(defaultContext())

	return s
}

// Initialize default context objects
func defaultContext() context.Context {
	ctx := context.Background()
	ctx = vars.ToContext(ctx, vars.New())
	return ctx
}

// Context returns this sequence's context
func (s *Sequence) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return defaultContext()
}

// WithContext returns a shallow copy of the Sequence with the new context
func (s *Sequence) WithContext(ctx context.Context) *Sequence {
	s2 := New()
	*s2 = *s

	seqVars := vars.FromContext(ctx)
	ctx = vars.ToContext(ctx, seqVars)

	s2.ctx = ctx

	return s2
}

//Completed indicates whether the sequence is done or still has more tasks to do successfully
func (s *Sequence) Completed() bool {
	return s.n > s.Size()
}

//Add adds a new task to the sequence and returns the number of tasks in the
//sequence
// Like all Sequence methods, this is not concurrency-safe
func (s *Sequence) Add(task task.Task, conds []condition.Condition) int {
	if len(conds) == 0 {
		conds = condition.Default()
	}

	s.Tasks = append(s.Tasks, task)
	s.conditions = append(s.conditions, conds)
	return s.Size()
}

//Size returns the number of tasks in the sequence
func (s *Sequence) Size() int {
	return len(s.Tasks)
}

// Executes the specified single step/task in the sequence
func (s *Sequence) executeStep(n int) error {
	var err error
	var ok bool
	var runLatch bool //true if the task has already run once
	defer func() {
		if r := recover(); r != nil {
			if err, ok = r.(error); !ok {
				err = fmt.Errorf("sequence [%d]: %v", n, r)
			}
		}
	}()
	// Step through conditions at this step, to determine whether this
	// task should be executed
	tsk := s.Tasks[n]
	for _, cond := range s.conditions[n] {
		for {
			// This for loop executes inside a single condition
			err = cond.Check(s.ctx)
			switch err {
			case nil:
				// Only run the task if it has not yet run
				if !runLatch {
					//This must be a precondition
					runLatch = true
					err = tsk.Execute(s.ctx)
				}
				// Regardless of whether the task was run,
				// break out of this condition and
				// continue to the next one
				break
			case condition.ErrSkip:
				return err
			case condition.ErrAbort:
				//TODO Better handle panic? Should never get here anyway
				return err
			case condition.ErrRetry:
				// Run task, but stay inside this condition
				// The condition is responsible for signaling when we
				// should stop retrying
				continue
			case condition.ErrFail:
				return err
			}
		}
	}
	return err
}

//Execute executes all remaining incomplete tasks in the Sequence
func (s *Sequence) Execute() *Result {
	var err error
	res := NewResult()
	for n, _ := range s.Tasks[s.n:] {
		// set number of current step
		s.n = n
		err = s.executeStep(s.n)
		if err != nil {
			res.Errors = append(res.Errors, err)
		}

		//TODO Instrument logging here
		if err == condition.ErrAbort || err == condition.ErrFail {
			break
		}
	}
	s.cancel()
	res.ReadContext(s.ctx)
	return res
}
