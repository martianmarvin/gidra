package client

import "net/url"

//HTTPClient performs http requests
type HTTPClient interface {
	Client

	// Page returns the last response parsed into a *Page
	Page() (*Page, error)
}

// Options are defaults that are shared by all HTTP clients
type HTTPOptions struct {
	*Options

	Method []byte

	URL *url.URL

	FollowRedirects bool

	Headers map[string]string

	Params map[string]string

	Cookies map[string]string

	// Text body
	Body []byte
}
