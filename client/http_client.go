package client

import (
	"net"
	"net/url"
	"time"

	"github.com/martianmarvin/conn"
	"github.com/martianmarvin/gidra/fastcookiejar"
	"github.com/valyala/fasthttp"
)

var (
	defaultTimeout = 15 * time.Second
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
}

func NewHTTPClient() *FastHTTPClient {
	c := &FastHTTPClient{
		client:          &fasthttp.Client{},
		dialer:          conn.NewDialer(),
		Jar:             fastcookiejar.New(),
		Timeout:         defaultTimeout,
		FollowRedirects: true,
	}
	c.client.Dial = c.dialer.FastDial
	return c
}

func (c *FastHTTPClient) Dial(addr string, proxy *url.URL) (net.Conn, error) {
	return c.dialer.Dial(addr, proxy)
}

//Close closes the underlying connection
func (c *FastHTTPClient) Close() error {
	var err error
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
func (c *FastHTTPClient) Do(req *fasthttp.Request) (*fasthttp.Response, error) {
	var err error

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

	return resp, err
}
