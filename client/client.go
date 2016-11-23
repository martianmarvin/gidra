// Package client encapsulates the http client used by gidra tasks
package client

import (
	"net"
	"net/url"

	"github.com/valyala/fasthttp"
)

// Client is the basic client to connect to an external resource
type Client interface {
	Dial(addr string, proxy *url.URL) (net.Conn, error)
	Close() error
}

type HTTPClient interface {
	Client
	Do(*fasthttp.Request) (*fasthttp.Response, error)
}

type MailClient interface {
	Client
	Login(email, password string) error
	//TODO: mail Message struct
	Search(kw string) interface{}
	Messages() interface{}
}
