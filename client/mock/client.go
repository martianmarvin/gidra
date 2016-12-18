package mock

import (
	"bytes"
	"net"
	"time"

	"github.com/valyala/fasthttp"
)

type MockConn struct {
	net.Conn
	r bytes.Buffer
	w bytes.Buffer
}

func (c *MockConn) Close() error {
	return nil
}

func (c *MockConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *MockConn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func (c *MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4zero}
}

func (c *MockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4zero}
}

func (c *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type MockClient struct {
	resp *fasthttp.Response
}

func (c *MockClient) Dial(addr string) (net.Conn, error) {
	return &MockConn{}, nil
}

func (c *MockClient) Close() error {
	return nil
}

// Do simply stores the provided argument to be returned from Response()
// Resp should be of the form fasthttp.Response
func (c *MockClient) Do(resp interface{}) error {
	if resp, ok := resp.(*fasthttp.Response); ok {
		c.resp = resp
	}
	return nil
}

func (c *MockClient) Response() ([]byte, error) {
	return []byte(c.resp.String()), nil
}
