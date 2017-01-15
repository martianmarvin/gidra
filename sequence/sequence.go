package sequence

import (
	"context"
	"fmt"

	"github.com/martianmarvin/gidra/condition"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/global"
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
func (s *Sequence) executeStep(ctx context.Context, n int) *Result {
	var err, taskErr error
	var ok bool
	res := &Result{}

	defer func() {
		if r := recover(); r != nil {
			if err, ok = r.(error); !ok {
				res.Err = fmt.Errorf("sequence [%d]: %v", n, r)
			}
		}
	}()

	tsk := s.Tasks[n]
	// Apply vars for this task to the task's context
	ctx = config.ToContext(ctx, s.Configs[n])

	g := global.FromContext(ctx)

	// Step through pre Conditions to determine whether this
	// task should be executed
	for _, cond := range s.Conditions[n] {
		err = cond.Check(ctx)
		switch err {
		case nil:
			// Condition passed, advance to the next tone
			continue
		case condition.ErrSkip:
			g.Status = global.StatusSkip
			res.Err = err
			return res
		}
	}

	taskErr = tsk.Execute(ctx)
	if taskErr != nil {
		res.Err = taskErr
		return res
	}

	// Step through post conditions to evaluate if the task executed
	// successfully
	for _, cond := range s.Conditions[n] {
		err = cond.Check(ctx)
		switch err {
		case nil:
			g.Status = global.StatusSuccess
			return res
		case condition.ErrAbort, condition.ErrFail:
			g.Status = global.StatusFail
			res.Err = err
			return res
		case condition.ErrRetry:
			// Retry task until condition tells us not to
			for err == condition.ErrRetry {
				taskErr = tsk.Execute(ctx)
				if taskErr != nil {
					res.Err = taskErr
					return res
				}
				err = cond.Check(ctx)
			}
		}
	}

	return res
}

//Execute executes all remaining incomplete tasks in the Sequence
func (s *Sequence) Execute(ctx context.Context) <-chan *Result {
	results := make(chan *Result, 1)
	go func() {
		defer close(results)
		for n, _ := range s.Tasks {
			logger := Logger.WithField("sid", s.Id).WithField("n", n)
			s.n = n
			logger.Warn("Executing step")
			// Wait for step to complete or cancellation
			select {
			case results <- s.executeStep(ctx, s.n):
				logger.Warn("Sent Result")
			case <-ctx.Done():
				return
			}
		}
	}()
	return results
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
