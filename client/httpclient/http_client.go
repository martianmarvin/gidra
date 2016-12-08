package httpclient

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/martianmarvin/conn"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/fastcookiejar"
	"github.com/valyala/fasthttp"
)

var defaultTimeout time.Duration = 15 * time.Second

// Config key names
var (
	cfgRoot            = "config.http"
	cfgFollowRedirects = "follow_redirects"
	cfgHeaders         = "headers"
	cfgTimeout         = "timeout"
	cfgProxy           = "proxy"
)

// Errors
var (
	ErrNoConfig = errors.New("Config not found")
)

type Client struct {
	// Tthe underlying connection
	conn net.Conn

	//The underlying fasthttp.Client instance
	client *fasthttp.Client

	//Underlying dialer, possible including proxy connection
	dialer *conn.Dialer

	// The proxy this client should use
	proxy *url.URL

	//The cookie jar shared between tasks using this client
	jar *fastcookiejar.Jar

	//Global headers
	headers map[string]string

	// Request timeout
	timeout time.Duration

	followRedirects bool

	//All responses from requests made by this client
	responses []*fasthttp.Response

	// page is the most recent page requested by this client
	page *Page
}

// New initializes an HTTP client with default settings
func New() *Client {
	c := &Client{
		client:          &fasthttp.Client{},
		dialer:          conn.NewDialer(),
		jar:             fastcookiejar.New(),
		timeout:         defaultTimeout,
		followRedirects: true,
		responses:       make([]*fasthttp.Response, 0),
	}
	c.client.Dial = c.dialer.FastDial
	return c
}

// Configure applies settings from the config to this client
func (c *Client) Configure(cfg *config.Config) error {
	var err error
	cfg, ok := cfg.CheckGet(cfgRoot)
	if !ok {
		return ErrNoConfig
	}

	if rawproxy, err := cfg.String(cfgProxy); err == nil && len(rawproxy) > 0 {
		u, err := url.Parse(rawproxy)
		if err == nil {
			c.proxy = u
		} else {
			return err
		}
	}

	if headers, err := cfg.StringMap(cfgHeaders); err == nil {
		for k, v := range headers {
			c.headers[k] = v
		}
	}

	if follow, err := cfg.Bool(cfgFollowRedirects); err == nil {
		c.followRedirects = follow
	}

	// timeout is int seconds
	if timeout, err := cfg.Int(cfgTimeout); err == nil {
		c.timeout = (time.Duration(timeout) * time.Second)
	}
	return err
}

func (c *Client) Dial(addr string) (net.Conn, error) {
	var err error
	c.conn, err = c.dialer.Dial(addr, c.proxy)
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
	for k, v := range c.headers {
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
func (c *Client) Do(req *fasthttp.Request) error {
	var err error

	req = c.buildRequest(req)
	resp := fasthttp.AcquireResponse()

	//follow redirects, saving cookies along the way
	requrl := req.URI().String()
	redirects := make([]string, 0)
	redirects = append(redirects, requrl)

	for {
		err = c.client.DoTimeout(req, resp, c.timeout)
		if err != nil {
			break
		}
		//Save updated cookies
		resp.Header.VisitAllCookie(c.jar.SetBytes)
		statusCode := resp.Header.StatusCode()

		if !c.followRedirects ||
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
		c.page = NewPage()
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
		return nil, ErrEmpty
	}
	resp := c.responses[n-1]
	c.responses = c.responses[:n-1]
	if resp == nil {
		return nil, ErrEmpty
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
func (c *Client) Page() (*Page, error) {
	if c.page != nil && len(c.page.Bytes) > 0 {
		return c.page, nil
	}
	if c.page == nil {
		c.page = NewPage()
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
	return nil, ErrEmpty
}
