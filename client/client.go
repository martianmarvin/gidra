// Package client provides an interface for clients that access external
// resources
package client

import (
	"context"
	"net"

	"github.com/martianmarvin/gidra/config"
)

// Context key
type contextKey int

const (
	ctxClient contextKey = iota
)

//TODO Interfaces for HTTP/Mail clients

// Client is the basic client to connect to an external resource
type Client interface {
	// Configure applies options to the client from the given *config.Config
	Configure(cfg *config.Config) error

	// Dial establishes a connection, optionally through the given proxy,
	// and persists it on the client
	Dial(addr string) (net.Conn, error)

	// Close closes the client's underlying connection
	Close() error

	// Do executes a request and saves the response on the client
	Do(req interface{}) error

	// Response returns the full text of the last successful response from request made by
	//this client as a byte array
	Response() ([]byte, error)
}

// ToContext attaches a client to this context
func ToContext(ctx context.Context, c Client) context.Context {
	return context.WithValue(ctx, ctxClient, c)
}

// FromContext returns the most recent client that was saved in this context
func FromContext(ctx context.Context) (Client, bool) {
	c, ok := ctx.Value(ctxClient).(Client)
	if !ok {
		return nil, false
	}
	return c, true
}
