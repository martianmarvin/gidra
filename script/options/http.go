package options

import (
	"net/url"
	"time"
)

type HTTPOptions struct {
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
