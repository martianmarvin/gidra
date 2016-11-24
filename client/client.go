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
	//Response returns the full text of the last successful response from request made by
	//this client as a byte array
	Response() []byte
}

type HTTPClient interface {
	Client
	//Do executes the request and saves the response on the client
	Do(req *fasthttp.Request, proxy *url.URL) error
}

type MailClient interface {
	Login(email, password string) error
	//TODO: mail Message struct
	Search(kw string) interface{}
	Messages() interface{}
}
