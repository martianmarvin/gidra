package script

import (
	"github.com/martianmarvin/gidra/fastcookiejar"
	"github.com/martianmarvin/gidra/task"
	"github.com/valyala/fasthttp"
)

// Sequence is a series of tasks that represent a single iteration of the loop
type Sequence struct {
	//Tasks is the list of tasks in this sequence
	Tasks []task.Task

	//The sequence id of the current task
	n int

	//List of errors from tasks(not including retries)
	errors []error

	//The cookie jar shared between tasks in this sequence
	cj *fastcookiejar.Jar

	//The persistent http client for this sequence
	client *fasthttp.Client
}

func NewSequence() *Sequence {
	return &Sequence{
		Tasks:  make([]task.Task, 0),
		errors: make([]error, 0),
		cj:     fastcookiejar.New(),
		client: &fasthttp.Client{},
	}
}

//Success indicates whether the sequence completed successfully
func (s *Sequence) Success() bool {
	return s.Completed() && (len(s.errors) == 0)
}

//Completed indicates whether the sequence is done or still has more tasks to do successfully
func (s *Sequence) Completed() bool {
	return s.n > len(s.Tasks)
}

//Add adds a new task to the sequence and returns the number of tasks in the
//sequence
// Like all Sequence methods, this is not concurrency-safe
func (s *Sequence) Add(task task.Task) int {
	s.Tasks = append(s.Tasks, task)
	s.n += 1
	return s.n
}

//Size returns the number of tasks in the sequence
func (s *Sequence) Size() int {
	return s.n
}
