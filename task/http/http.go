package http

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/client/httpclient"
	"github.com/martianmarvin/gidra/task"
)

var (
	ErrBadClient = errors.New("Invalid client, HTTPClient is required for this task")
)

func init() {
	task.Register("http", New)
}

type Task struct {
	task.Worker
	task.Configurable

	Config *Config
}

type Config struct {
	Method []byte

	URL *url.URL `task:"url, required"`

	FollowRedirects bool `task:"follow_redirects"`

	Headers map[string]string

	Params map[string]string

	Cookies map[string]string

	// Text body
	Body []byte

	// JSON body
	// TODO - marshal and save bytes in Body
	Json map[string]string
}

func newHTTP() *Task {
	t := &Task{
		Config: &Config{},
		Worker: task.NewWorker(),
	}
	t.Configurable = task.NewConfigurable(t.Config)
	return t
}

func New() task.Task {
	return newHTTP()
}

func (t *Task) Execute(ctx context.Context) error {
	var err error

	// Get client from the context
	c, ok := httpclient.FromContext(ctx)
	if !ok {
		return ErrBadClient
	}

	opts := &client.HTTPOptions{
		Method: t.Config.Method,
		URL:    t.Config.URL,
		//TODO only if set?
		FollowRedirects: t.Config.FollowRedirects,
		Headers:         t.Config.Headers,
		Params:          t.Config.Params,
		Cookies:         t.Config.Cookies,
		Body:            t.Config.Body,
	}

	err = c.Do(opts)

	return err
}

func (t *Task) String() string {
	if len(t.Config.Method) == 0 || t.Config.URL == nil {
		return "http: <nil>"
	}
	return fmt.Sprintf("http: %s %s", strings.ToUpper(string(t.Config.Method)), t.Config.URL)
}
