// Package task contains primitives for task modules
package task

import (
	"context"
	"errors"
	"sort"
	"sync"
)

// Context key
type contextKey int

const (
	ctxClient contextKey = iota
)

var (
	tasksMu         sync.RWMutex
	registeredTasks = make(map[string]newTaskFunc)
)

// Errors
var (
	ErrNotImplemented = errors.New("This feature is not supported by the current task")
)

// Task is a single step in a Script
type Task interface {
	// Execute executes the task and returns an error if it did not complete
	Execute(ctx context.Context) error
}

type newTaskFunc func() Task

// ExecFunc executes a task with the given context
type ExecFunc func(ctx context.Context) error

// ExecFunc also satisfies the Task inteface
func (f ExecFunc) Execute(ctx context.Context) {
	f(ctx)
}

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
//This should be the only way new tasks are launched
func New(action string) Task {
	tasksMu.RLock()
	defer tasksMu.RUnlock()
	fn, ok := registeredTasks[action]
	if !ok {
		panic("No such task: " + action)
	}
	t := fn()

	//wrap for middleware
	t = &task{task: t, name: action}

	return t
}

//Run runs a task immediately, out of sequence
func Run(ctx context.Context, action string) {
	t := New(action)
	t.Execute(ctx)
}
