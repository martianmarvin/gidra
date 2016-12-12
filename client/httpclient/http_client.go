package httpclient

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/imdario/mergo"
	"github.com/martianmarvin/conn"
	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/fastcookiejar"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/valyala/fasthttp"
)

// Context key
type contextKey int

const (
	ctxClient contextKey = iota
)

var defaultTimeout time.Duration = 15 * time.Second

type Client struct {
	// Options are globally applied to each request by the client
	Options *options.HTTPOptions

	// The underlying connection
	conn net.Conn

	//The underlying fasthttp.Client instance
	client *fasthttp.Client

	//Underlying dialer, possible including proxy connection
	dialer *conn.Dialer

	//The cookie jar shared between tasks using this client
	jar *fastcookiejar.Jar

	//All responses from requests made by this client
	responses []*fasthttp.Response

	// page is the most recent page requested by this client
	page *client.Page
}

// ToContext attaches a client to this context
func ToContext(ctx context.Context, c *Client) context.Context {
	// Also save to top-level Client slot
	ctx = client.ToContext(ctx, c)
	return context.WithValue(ctx, ctxClient, c)
}

// FromContext returns the most recent client that was saved in this context
func FromContext(ctx context.Context) (*Client, bool) {
	c, ok := ctx.Value(ctxClient).(*Client)
	if !ok {
		return nil, false
	}
	return c, true
}

// New initializes an HTTP client with default settings
func New() *Client {
	c := &Client{
		client:    &fasthttp.Client{},
		dialer:    conn.NewDialer(),
		jar:       fastcookiejar.New(),
		responses: make([]*fasthttp.Response, 0),
		Options: &options.HTTPOptions{
			Timeout: defaultTimeout,
		},
	}
	c.dialer.Proxy = c.Options.Proxies[0]
	c.client.Dial = c.dialer.FastDial
	return c
}

// WithOptions applies settings from the Options struct to this client
func (c *Client) WithOptions(opts *options.HTTPOptions) *Client {
	mergo.MergeWithOverwrite(c.Options, opts)
	// Add cookies to jar for all domains
	c.jar.SetMap(".", c.Options.Cookies)
	return c

}

func (c *Client) Dial(addr string) (net.Conn, error) {
	var err error
	c.conn, err = c.dialer.Dial(addr, c.Options.Proxy)
	return c.conn, err
}

//Close closes the underlying connection and releases all responses
func (c *Client) Close() error {
	var err error
	for _, resp := range c.responses {
		fasthttp.ReleaseResponse(resp)
	}
	if c.conn != nil {
		err = c.conn.Close()
	}
	return err
}

// Apply client's headers and cookies to request
func (c *Client) buildRequest(req *fasthttp.Request) *fasthttp.Request {
	//Apply global client headers if not in request
	for k, v := range c.Options.Headers {
		if len(req.Header.Peek(k)) == 0 {
			req.Header.Set(k, v)
		}
	}

	//Apply cookies from jar
	cookies := c.jar.Cookies(string(req.Host()))
	for _, ck := range cookies {
		req.Header.SetCookieBytesKV(ck.Key(), ck.Value())
	}
	return req
}

func getRedirectURL(baseURL string, location []byte) string {
	u := fasthttp.AcquireURI()
	u.Update(baseURL)
	u.UpdateBytes(location)
	redirectURL := u.String()
	fasthttp.ReleaseURI(u)
	return redirectURL
}

//Do executes the request, applying all client-global options and returns the
//response
func (c *Client) Do(r interface{}) error {
	var err error

	req, ok := r.(*fasthttp.Request)
	if !ok {
		return errors.New("This client can only execute a request of type *fasthttp.Request")
	}

	req = c.buildRequest(req)
	resp := fasthttp.AcquireResponse()

	//follow redirects, saving cookies along the way
	requrl := req.URI().String()
	redirects := make([]string, 0)
	redirects = append(redirects, requrl)

	for {
		err = c.client.DoTimeout(req, resp, c.Options.Timeout)
		if err != nil {
			break
		}
		//Save updated cookies
		resp.Header.VisitAllCookie(c.jar.SetBytes)
		statusCode := resp.Header.StatusCode()

		if !c.Options.FollowRedirects ||
			!(statusCode == 301 || statusCode == 302 || statusCode == 303) {
			break
		}
		location := resp.Header.Peek("Location")
		if len(location) == 0 {
			break
		}
		requrl = getRedirectURL(requrl, location)
		redirects = append(redirects, requrl)
		req.SetRequestURI(requrl)
	}

	fasthttp.ReleaseRequest(req)
	if err != nil {
		fasthttp.ReleaseResponse(resp)
		c.responses = append(c.responses, nil)
	} else {
		c.responses = append(c.responses, resp)
		c.page = client.NewPage()
		c.page.URL, _ = url.Parse(requrl)
		c.page.Redirects.Append(redirects...)
	}

	return err
}

//Response pops the latest response from this client's list of responses
//if the most recent request had an error or there are no more responses
func (c *Client) Response() ([]byte, error) {
	n := len(c.responses)
	if n == 0 {
		return nil, client.ErrEmpty
	}
	resp := c.responses[n-1]
	c.responses = c.responses[:n-1]
	if resp == nil {
		return nil, client.ErrEmpty
	}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	_, err := resp.WriteTo(w)
	if err != nil {
		return nil, err
	}
	w.Flush()
	fasthttp.ReleaseResponse(resp)
	return buf.Bytes(), nil
}

// Page returns a *Page based on the most recent response from this client
func (c *Client) Page() (*client.Page, error) {
	if c.page != nil && len(c.page.Bytes) > 0 {
		return c.page, nil
	}
	if c.page == nil {
		c.page = client.NewPage()
	}
	// check non-nil responses if we haven't parsed a page in yet
	for i := len(c.responses) - 1; i >= 0; i-- {
		resp := c.responses[i]
		if resp == nil {
			continue
		}
		err := c.page.Parse(resp)
		if err != nil {
			return nil, err
		}
		return c.page, nil
	}
	return nil, client.ErrEmpty
}
