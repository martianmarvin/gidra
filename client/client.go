// Package client provides an interface for clients that access external
// resources
package client

import (
	"net"

	"github.com/valyala/fasthttp"
)

// Client is the basic client to connect to an external resource
type Client interface {
	// Dial establishes a connection, optionally through the given proxy,
	// and persists it on the client
	Dial(addr string) (net.Conn, error)

	// Close closes the client's underlying connection
	Close() error

	// Response returns the full text of the last successful response from request made by
	//this client as a byte array
	Response() []byte
}

type HTTPClient interface {
	Client

	// Do executes the request and saves the response on the client
	// The client is responsible for managing its own state, including
	// cookies, proxies, etc
	Do(req *fasthttp.Request) error
}

type MailClient interface {
	Login(email, password string) error
	//TODO: mail Message struct
	Search(kw string) interface{}
	Messages() interface{}
}
