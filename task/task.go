// Package task contains primitives for task modules
package task

import (
	"errors"
	"sort"
	"sync"

	"github.com/martianmarvin/gidra/client"
)

//Errors represent status of a task that failed to complete
var (
	ErrAbort = errors.New("Task aborted")
	ErrRetry = errors.New("Task temporary failure, retrying")
	ErrFail  = errors.New("Task failed, moving on")
	ErrSkip  = errors.New("Skipping task")

	tasksMu         sync.RWMutex
	registeredTasks = make(map[string]newTaskFunc)

	defaultClient = client.NewHTTPClient()
)

// Task is a single step in a Script
type Task interface {
	// Execute executes the task and returns an error if it did not complete
	Execute(client client.Client, vars map[string]interface{}) error
}

type newTaskFunc func() Task

//Register registers a new type of task, making it available to scripts
func Register(action string, fn newTaskFunc) {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	if fn == nil {
		panic("Invalid task")
	}
	if _, dup := registeredTasks[action]; dup {
		panic("Register called twice for task " + action)
	}
	registeredTasks[action] = fn
}

//Tasks returns a sorted list of all available task types
func Tasks() []string {
	tasksMu.RLock()
	defer tasksMu.RUnlock()
	var list []string
	for action := range registeredTasks {
		list = append(list, action)
	}
	sort.Strings(list)
	return list
}

//New initializes and returns a task of the specified action
func New(action string) Task {
	fn, ok := registeredTasks[action]
	if !ok {
		panic("No such task: " + action)
	}
	return fn()
}

//Run runs a task immediately, out of sequence
func Run(action string, vars map[string]interface{}) error {
	t := New(action)
	return t.Execute(defaultClient, vars)
}
