package http

import (
	"fmt"
	"net/url"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/task"
	"github.com/valyala/fasthttp"
)

type Task struct {
	task.BaseTask

	Config *Config
}

type Config struct {
	Method []byte `task:"required"`

	URL string `task:"required"`

	Proxy *url.URL

	Headers map[string]string

	Cookies map[string]string

	Params map[string]string

	Body []byte

	Client client.Client
}

func (t *Task) Execute(client client.Client, vars map[string]interface{}) (err error) {
	if err = task.Configure(t, vars); err != nil {
		return err
	}
	t.Config.Client = client

	fmt.Println(t.Config)

	return err
}

//Build HTTP request based on config
func (t *Task) buildRequest() (req *fasthttp.Request) {
	if t.Config.Proxy != nil {
		//TODO Set proxy dialer
	}
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
