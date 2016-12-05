package httpclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/martianmarvin/conn"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/fastcookiejar"
	"github.com/valyala/fasthttp"
)

var defaultTimeout time.Duration = 15 * time.Second

// Context key
type contextKey int

const (
	ctxClient contextKey = iota
	ctxProxy
	ctxHeaders
	ctxFollowRedirects
	ctxTimeout
)

// Config key names
var (
	cfgRoot            = "config.http"
	cfgFollowRedirects = cfgRoot + ".follow_redirects"
	cfgHeaders         = cfgRoot + ".headers"
	cfgTimeout         = cfgRoot + ".timeout"
	cfgProxy           = cfgRoot + ".proxy"
)

// ToContext attaches an http client to this context
func ToContext(ctx context.Context, c *Client) context.Context {
	return context.WithValue(ctx, ctxClient, c)
}

// FromContext returns the client attached to this context, setting its global proxy and
// headers according the context's globals
func FromContext(ctx context.Context) (*Client, bool) {
	c, ok := ctx.Value(ctxClient).(*Client)
	if !ok {
		return nil, false
	}
	// Apply proxy
	if p, ok := ctx.Value(ctxProxy).(*url.URL); ok {
		c.proxy = p
	}
	// Apply context headers, overwriting any on client
	if headers, ok := ctx.Value(ctxHeaders).(map[string]string); ok {
		for k, v := range headers {
			c.headers[k] = v
		}
	}
	// Other settings
	if followRedirects, ok := ctx.Value(ctxFollowRedirects).(bool); ok {
		c.followRedirects = followRedirects
	}
	// Other settings
	if timeout, ok := ctx.Value(ctxTimeout).(time.Duration); ok {
		c.timeout = timeout
	}
	return c, true
}

// WithProxy sets a proxy on the context
func WithProxy(ctx context.Context, p *url.URL) context.Context {
	return context.WithValue(ctx, ctxProxy, p)
}

// WithHeaders sets global headers on the context, merging any headers
// already set before
func WithHeaders(ctx context.Context, addheaders map[string]string) context.Context {
	headers, ok := ctx.Value(ctxHeaders).(map[string]string)
	if !ok {
		headers = make(map[string]string)
	}
	for k, v := range addheaders {
		headers[k] = v
	}
	return context.WithValue(ctx, ctxHeaders, headers)
}

// WithFollow sets the Follow Redirects setting on the context
func WithFollow(ctx context.Context, followRedirects bool) context.Context {
	return context.WithValue(ctx, ctxFollowRedirects, followRedirects)
}

// WithTimeout sets the http request timeout on the context
func WithTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, ctxTimeout, timeout)
}

// Assert map values to string
func stringMap(m map[string]interface{}) map[string]string {
	sm := make(map[string]string)
	for k, v := range m {
		sm[k] = fmt.Sprint(v)
	}
	return sm
}

// ConfigureCtx sets global defaults from the context's config in the context
func ConfigureCtx(ctx context.Context) context.Context {
	cfg := config.FromContext(ctx)
	if rawproxy, err := cfg.String(cfgProxy); err == nil && len(rawproxy) > 0 {
		u, err := url.Parse(rawproxy)
		if err == nil {
			ctx = WithProxy(ctx, u)
		}
	}

	if headers, err := cfg.Map(cfgHeaders); err == nil {
		ctx = WithHeaders(ctx, stringMap(headers))
	}

	if follow, err := cfg.Bool(cfgFollowRedirects); err == nil {
		ctx = WithFollow(ctx, follow)
	}

	// timeout is int seconds
	if timeout, err := cfg.Int(cfgTimeout); err == nil {
		ctx = WithTimeout(ctx, time.Duration(timeout)*time.Second)
	}

	return ctx
}

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
}

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
		newurl := getRedirectURL(req.URI().String(), location)
		req.SetRequestURI(newurl)
	}

	fasthttp.ReleaseRequest(req)
	if err != nil {
		fasthttp.ReleaseResponse(resp)
		c.responses = append(c.responses, nil)
	} else {
		c.responses = append(c.responses, resp)
	}

	return err
}

//Parses fasthttp response to text
func parseResponse(resp *fasthttp.Response) []byte {
	if resp == nil {
		return nil
	}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	resp.WriteTo(w)
	return buf.Bytes()
}

//Response pops the latest response from this client's list of responses
//if the most recent request had an error or there are no more responses
func (c *Client) Response() []byte {
	var text []byte
	n := len(c.responses)
	if n == 0 {
		return nil
	} else {
		resp := c.responses[n-1]
		c.responses = c.responses[:n-1]
		text = parseResponse(resp)
		fasthttp.ReleaseResponse(resp)
		return text
	}
}
