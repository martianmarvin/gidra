package sequence

import (
	"context"
	"fmt"

	"github.com/martianmarvin/gidra/condition"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/log"
	"github.com/martianmarvin/gidra/task"
)

var Logger = log.Logger()

// Sequence is a series of tasks that represent a single iteration of the loop
type Sequence struct {
	// The number of the current loop iteration
	Id int

	// The current row of data from the main input
	Row *datasource.Row

	//Tasks is the list of tasks in this sequence
	Tasks []task.Task

	//The sequence id of the current task
	n int

	//List of Conditions corresponding to tasks in the sequence
	Conditions [][]condition.Condition

	// Options corresponding to each task in the sequence
	Configs []*config.Config
}

func New() *Sequence {
	s := &Sequence{
		Tasks:      make([]task.Task, 0),
		Conditions: make([][]condition.Condition, 0),
		Configs:    make([]*config.Config, 0),
	}

	return s
}

// Copy returns a shallow copy of the Sequence with the new context
func (s *Sequence) Copy() *Sequence {
	s2 := New()
	*s2 = *s

	return s2
}

//Completed indicates whether the sequence is done or still has more tasks to do successfully
func (s *Sequence) Completed() bool {
	return s.n > s.Size()
}

//Add adds a new task to the sequence and returns the number of tasks in the
//sequence
// Like all Sequence methods, this is not concurrency-safe
func (s *Sequence) Add(task task.Task, conds []condition.Condition, cfg *config.Config) int {
	if len(conds) == 0 {
		conds = condition.Default()
	}

	s.Tasks = append(s.Tasks, task)
	s.Conditions = append(s.Conditions, conds)
	s.Configs = append(s.Configs, cfg)
	return s.Size()
}

//Size returns the number of tasks in the sequence
func (s *Sequence) Size() int {
	return len(s.Tasks)
}

// Executes the specified single step/task in the sequence
func (s *Sequence) executeStep(ctx context.Context, n int) error {
	var err, taskErr error
	var ok bool

	defer func() {
		if r := recover(); r != nil {
			if err, ok = r.(error); !ok {
				err = fmt.Errorf("sequence [%d]: %v", n, r)
			}
		}
	}()

	tsk := s.Tasks[n]
	// Apply vars for this task to the task's context
	ctx = config.ToContext(ctx, s.Configs[n])

	// Step through pre Conditions to determine whether this
	// task should be executed
	for _, cond := range s.Conditions[n] {
		err = cond.Check(ctx)
		switch err {
		case nil:
			// Condition passed, advance to the next tone
			continue
		case condition.ErrSkip:
			return err
		}
	}

	taskErr = tsk.Execute(ctx)
	if taskErr != nil {
		return taskErr
	}

	// Step through post conditions to evaluate if the task executed
	// successfully
	for _, cond := range s.Conditions[n] {
		err = cond.Check(ctx)
		switch err {
		case nil:
			return nil
		case condition.ErrAbort, condition.ErrFail:
			return err
		case condition.ErrRetry:
			// Retry task until condition tells us not to
			for err == condition.ErrRetry {
				taskErr = tsk.Execute(ctx)
				if taskErr != nil {
					return taskErr
				}
				err = cond.Check(ctx)
			}
		}
	}

	return taskErr
}

//Execute executes all remaining incomplete tasks in the Sequence
func (s *Sequence) Execute(ctx context.Context) *Result {
	res := NewResult()
	done := make(chan error, 1)
	for n, _ := range s.Tasks {
		// set number of current step
		s.n = n
		Logger.WithField("sid", s.Id).WithField("n", n).Warn("Executing step")
		go func() {
			done <- s.executeStep(ctx, s.n)
		}()
		select {
		case err := <-done:
			//TODO Instrument logging here
			if err == condition.ErrAbort {
				panic("Aborting script!")
			} else if err == condition.ErrFail {
				return res
			} else if err != nil {
				res.Errors = append(res.Errors, err)
			}
		case <-ctx.Done():
			return res
		}
	}
	// res.ReadContext(s.ctx) TODO get result
	return res
}

// TODO make sure user variables are shown
// String implements the Stringer interface
func (s *Sequence) String() string {
	var out string
	for i := 0; i < s.Size(); i++ {
		ctx := config.ToContext(context.Background(), s.Configs[i])
		out += fmt.Sprintf("%d: %s\n", i, task.Show(ctx, s.Tasks[i]))
	}
	return out
}
