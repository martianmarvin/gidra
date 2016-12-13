package http

import (
	"time"

	"github.com/martianmarvin/gidra/client"
)

// Options are defaults that are shared by all HTTP clients
type Options struct {
	URL *client.URLList

	FollowRedirects bool

	Timeout time.Duration

	Proxy *client.URLList

	Headers map[string]string

	Params map[string]string

	Cookies map[string]string

	// Text body
	Body []byte
}
