package http

import (
	"context"
	"errors"
	"net/url"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/client/httpclient"
	"github.com/martianmarvin/gidra/task"
	"github.com/valyala/fasthttp"
)

var (
	ErrBadClient = errors.New("Invalid client, HTTPClient is required for this task")
)

type Task struct {
	task.Worker
	task.Configurable

	Config *Config
}

type Config struct {
	Method []byte

	URL *url.URL `task:"url,required"`

	Headers map[string]string

	Cookies map[string]string

	Params map[string]string

	Body []byte

	FollowRedirects bool

	Proxy *client.URLList
}

func (t *Task) Execute(ctx context.Context) (err error) {

	// Get client from the context
	c, ok := httpclient.FromContext(ctx)
	if !ok {
		return ErrBadClient
	}

	req := t.buildRequest()

	c.Options.FollowRedirects = t.Config.FollowRedirects
	c.Options.Proxy = t.Config.Proxy

	err = c.Do(req)

	return err
}

//Build HTTP request based on config
func (t *Task) buildRequest() (req *fasthttp.Request) {
	req = fasthttp.AcquireRequest()
	req.Header.SetMethodBytes(t.Config.Method)
	req.SetRequestURI(t.Config.URL.String())

	for k, v := range t.Config.Headers {
		req.Header.Set(k, v)
	}

	for k, v := range t.Config.Cookies {
		req.Header.SetCookie(k, v)
	}

	if len(t.Config.Params) == 0 && len(t.Config.Body) > 0 {
		req.SetBody(t.Config.Body)
	} else if len(t.Config.Params) > 0 {
		args := fasthttp.AcquireArgs()
		for k, v := range t.Config.Params {
			args.Set(k, v)
		}
		req.Header.SetContentType("application/x-www-form-urlencoded")
		args.WriteTo(req.BodyWriter())
		fasthttp.ReleaseArgs(args)
	}

	return req
}

func (t *Task) String() string {
	if len(t.Config.Method) == 0 || t.Config.URL == nil {
		return "http - <nil>"
	}
	return t.buildRequest().String()
}
