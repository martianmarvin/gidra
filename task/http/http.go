package http

import (
	"errors"
	"net/url"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/task"
	"github.com/valyala/fasthttp"
)

var (
	ErrBadClient = errors.New("Invalid client, HTTPClient is required for this task")
)

type Task struct {
	task.BaseTask

	Config *Config
}

type Config struct {
	Method []byte `task:"required"`

	URL string `task:"url,required"`

	Proxy *url.URL

	Headers map[string]string

	Cookies map[string]string

	Params map[string]string

	Body []byte

	Client client.HTTPClient
}

func (t *Task) Execute(c client.Client, vars map[string]interface{}) (err error) {
	if _, ok := vars["method"]; !ok {
		vars["method"] = t.Config.Method
	}
	if err = task.Configure(t, vars); err != nil {
		return err
	}
	//TODO idiomatic Dialer instead of proxy
	req := t.buildRequest()

	httpclient, ok := c.(client.HTTPClient)
	if !ok {
		return ErrBadClient
	}

	t.Config.Client = httpclient

	err = httpclient.Do(req, t.Config.Proxy)

	return err
}

//Build HTTP request based on config
func (t *Task) buildRequest() (req *fasthttp.Request) {
	req = fasthttp.AcquireRequest()
	req.Header.SetMethodBytes(t.Config.Method)
	req.SetRequestURI(t.Config.URL)

	for k, v := range t.Config.Headers {
		req.Header.Set(k, v)
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
