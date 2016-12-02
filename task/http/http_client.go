package http

import (
	"bufio"
	"bytes"
	"net"
	"net/url"
	"time"

	"github.com/martianmarvin/conn"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/fastcookiejar"
	"github.com/valyala/fasthttp"
)

type FastHTTPClient struct {
	//The underlying fasthttp.Client instance
	client *fasthttp.Client

	//Underlying dialer, possible including proxy connection
	dialer *conn.Dialer

	//The cookie jar shared between tasks using this client
	Jar *fastcookiejar.Jar

	//Global headers
	Headers map[string]string

	// Request timeout
	Timeout time.Duration

	FollowRedirects bool

	//All responses from requests made by this client
	Responses []*fasthttp.Response
}

func NewHTTPClient() *FastHTTPClient {
	c := &FastHTTPClient{
		client:          &fasthttp.Client{},
		dialer:          conn.NewDialer(),
		Jar:             fastcookiejar.New(),
		Timeout:         config.Timeout,
		FollowRedirects: true,
		Responses:       make([]*fasthttp.Response, 0),
	}
	c.client.Dial = c.dialer.FastDial
	return c
}

func (c *FastHTTPClient) Dial(addr string, proxy *url.URL) (net.Conn, error) {
	return c.dialer.Dial(addr, proxy)
}

//Close closes the underlying connection and releases all responses
func (c *FastHTTPClient) Close() error {
	var err error
	for _, resp := range c.Responses {
		fasthttp.ReleaseResponse(resp)
	}
	return err
}

// Apply client's headers and cookies to request
func (c *FastHTTPClient) buildRequest(req *fasthttp.Request) *fasthttp.Request {
	//Apply global client headers if not in request
	for k, v := range c.Headers {
		if len(req.Header.Peek(k)) == 0 {
			req.Header.Set(k, v)
		}
	}

	//Apply cookies from Jar
	cookies := c.Jar.Cookies(string(req.Host()))
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
func (c *FastHTTPClient) Do(req *fasthttp.Request, proxy *url.URL) error {
	var err error

	c.dialer.Proxy = proxy

	req = c.buildRequest(req)
	resp := fasthttp.AcquireResponse()

	//follow redirects, saving cookies along the way
	for {
		err = c.client.DoTimeout(req, resp, c.Timeout)
		if err != nil {
			break
		}
		//Save updated cookies
		resp.Header.VisitAllCookie(c.Jar.SetBytes)
		statusCode := resp.Header.StatusCode()

		if !c.FollowRedirects ||
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
		c.Responses = append(c.Responses, nil)
	} else {
		c.Responses = append(c.Responses, resp)
	}

	return err
}

//Parses fasthttp response to text
func parseResponse(resp *fasthttp.Response) []byte {
	if resp == nil {
		return []byte{}
	}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	resp.WriteTo(w)
	return buf.Bytes()
}

//Response returns the response for this client's most recent request, or nil
//if the most recent request had an error
//The response is released after being read and should not be accessed again
func (c *FastHTTPClient) Response() []byte {
	var text []byte
	n := len(c.Responses)
	if n == 0 {
		return text
	} else {
		resp := c.Responses[n-1]
		text = parseResponse(resp)
		fasthttp.ReleaseResponse(resp)
		return text
	}
}
