package script

import (
	"context"
	"fmt"

	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
)

// Sequence is a series of tasks that represent a single iteration of the loop
type Sequence struct {
	//Tasks is the list of tasks in this sequence
	Tasks []task.Task

	//The sequence id of the current task
	n int

	//List of errors from tasks(not including retries)
	errors []error

	//List of conditions corresponding to tasks in the sequence
	conditions [][]Condition

	//Sequence-global variables/config, set per iteration
	Vars *vars.Vars

	//Context shared by all requests in this sequence
	ctx context.Context
}

func NewSequence() *Sequence {
	s := &Sequence{
		Tasks:      make([]task.Task, 0),
		errors:     make([]error, 0),
		conditions: make([][]Condition, 0),
		context:    defaultContext(),
		Vars:       vars.New(),
	}

	return s
}

func defaultContext() context.Context {
	ctx := context.Background()
	return ctx
}

//Success indicates whether the sequence completed successfully
func (s *Sequence) Success() bool {
	return s.Completed() && (s.ErrCount() == 0)
}

//Completed indicates whether the sequence is done or still has more tasks to do successfully
func (s *Sequence) Completed() bool {
	return s.n > s.Size()
}

//Add adds a new task to the sequence and returns the number of tasks in the
//sequence
// Like all Sequence methods, this is not concurrency-safe
func (s *Sequence) Add(task task.Task) int {
	s.Tasks = append(s.Tasks, task)
	s.errors = append(s.errors, nil)
	s.conditions = append(s.conditions, defaultConditions())
	return s.Size()
}

//Size returns the number of tasks in the sequence
func (s *Sequence) Size() int {
	return len(s.Tasks)
}

//ErrCount returns the number of errors encountered so far
func (s *Sequence) ErrCount() int {
	var n int
	for _, err := range s.errors {
		if err != nil {
			n += 1
		}
	}
	return n
}

//Step returns the step the sequence is on
func (s *Sequence) Step() int {
	return s.n
}

// Executes the specified single step/task in the sequence
func (s *Sequence) executeStep(n int) error {
	var err error
	var ok bool
	var runLatch bool //true if the task has already run once
	defer func() {
		if r := recover(); r != nil {
			if err, ok = r.(error); !ok {
				fmt.Errorf("sequence [%d]: %v", n, r)
			}
		}
		s.errors[n] = err
	}()
	// Step through conditions at this step, to determine whether this
	// task should be executed
	tsk := s.Tasks[n]
	for _, cond := range s.conditions[n] {
		vars := s.Vars.Map()
		for {
			// This for loop executes inside a single condition
			err = cond.Check(vars)
			switch err {
			case nil:
				// Only run the task if it has not yet run
				if !runLatch {
					//This must be a precondition
					s.errors[n] = tsk.Execute(s.Client, vars)
					runLatch = true
				}
				// Regardless of whether the task was run,
				// break out of this condition and
				// continue to the next one
				break
			case ErrSkip:
				return err
			case ErrAbort:
				//TODO Better handle panic? Should never get here anyway
				return err
			case ErrRetry:
				// Run task, but stay inside this condition
				// The condition is responsible for signaling when we
				// should stop retrying
				s.errors[n] = tsk.Execute(s.Client, vars)
				continue
			case ErrFail:
				return err
			}
		}
	}
	return err
}

//Execute executes all remaining incomplete tasks in the Sequence
func (s *Sequence) Execute() []error {
	for n, _ := range s.Tasks[s.n:] {
		// set number of current step
		s.n = n
		err := s.executeStep(s.n)

		//TODO Instrument logging here
		if err == ErrAbort || err == ErrFail {
			break
		}
	}
	return s.errors
}

//Clone creates a deep copy of this sequence
func (s *Sequence) Clone() (*Sequence, error) {
	var err error
	seq := &Sequence{
		Tasks:      make([]task.Task, len(s.Tasks)),
		n:          s.n,
		errors:     make([]error, len(s.errors)),
		conditions: make([][]Condition, len(s.conditions)),
		Client:     s.Client,
	}

	seq.Vars, err = s.Vars.Clone()
	if err != nil {
		return nil, err
	}

	copy(seq.Tasks, s.Tasks)
	copy(seq.errors, s.errors)

	for i, conds := range s.conditions {
		seq.conditions[i] = make([]Condition, len(conds))
		copy(seq.conditions[i], conds)
	}

	return seq, err
}

// Configure merges each map of variables, and applies them to this sequence
// This method is safe to use concurrently from multiple goroutines
func (s *Sequence) Configure(vars ...*vars.Vars) {
	for _, v := range vars {
		s.Vars.Extend(v)
	}
}
