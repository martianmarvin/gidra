package script

import (
	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/task"
)

var (
	DefaultRetries = 5
)

// Sequence is a series of tasks that represent a single iteration of the loop
type Sequence struct {
	//Tasks is the list of tasks in this sequence
	Tasks []task.Task

	//The sequence id of the current task
	n int

	//How many times to retry the current task on fail
	retries int

	//List of errors from tasks(not including retries)
	errors []error

	//Sequence-global variables/config, set per iteration
	Vars map[string]interface{}

	//Client shared by all requests in this sequence
	Client client.Client
}

func NewSequence(c client.Client) *Sequence {
	s := &Sequence{
		Tasks:   make([]task.Task, 0),
		errors:  make([]error, 0),
		Vars:    make(map[string]interface{}),
		Client:  c,
		retries: DefaultRetries,
	}

	return s
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

//Execute executes all remaining incomplete tasks in the Sequence
func (s *Sequence) Execute() []error {
	for n, t := range s.Tasks[s.n:] {
		s.n = n
		err := t.Execute(s.Client, s.Vars)
		s.errors[n] = err
		switch err {
		case task.ErrRetry:
			for i := 0; i < s.retries; i++ {
				err = t.Execute(s.Client, s.Vars)
				s.errors[n] = err
				if err != task.ErrRetry {
					break
				}
			}
			fallthrough
		case task.ErrAbort:
			panic("Aborting script")
		case task.ErrFail:
			return s.errors
		case task.ErrSkip:
			continue
		default:
			continue
		}
	}
	return s.errors
}
