package httpclient

import (
	"net/url"
	"time"
)

// Options are defaults that are shared by all HTTP clients
type Options struct {
	URL *url.URL

	FollowRedirects bool

	Timeout time.Duration

	Proxy *url.URL

	Headers map[string]string

	Params map[string]string

	Cookies map[string]string

	// Text body
	Body []byte
}
