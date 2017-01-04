// Package global represents global objects available to the end-user in
// their tasks and templates
package global

import (
	"context"
	"net/url"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/client/httpclient"
	"github.com/martianmarvin/gidra/datasource"
)

// Context
type contextKey int

// Global variables available inside all tasks are stored on the context
const (
	ctxGlobal contextKey = iota
)

// Global is the global context passed to templates. It is actually global to a
// sequence, not all sequences

type Global struct {
	// Vars are all variables available to the user
	Vars map[string]interface{}

	// Proxy represents the proxy list used by the task
	Proxy *url.URL

	// Page is the page requested by the last request
	Page *client.Page

	// Status of the last executed task
	Status int

	// Inputs are the user-defined datasources
	Inputs map[string]datasource.ReadableTable
}

// New instantiates a new Global object
func New() *Global {
	return &Global{
		Vars:   make(map[string]interface{}),
		Inputs: make(map[string]datasource.ReadableTable),
		Page:   client.NewPage(),
	}
}

// Copy returns a shallow copy of the Global
func (g *Global) Copy() *Global {
	g2 := New()
	for k, v := range g.Vars {
		g2.Vars[k] = v
	}
	for k, v := range g.Inputs {
		g2.Inputs[k] = v
	}
	*g2.Proxy = *g.Proxy
	*g2.Page = *g.Page
	return g2
}

// ToContext saves the Global in the context
func ToContext(ctx context.Context, g *Global) context.Context {
	return context.WithValue(ctx, ctxGlobal, g)
}

// FromContext retrieves the Global from the context, or a new one if one does
// not exist
func FromContext(ctx context.Context) *Global {
	g, ok := ctx.Value(ctxGlobal).(*Global)
	if !ok {
		g = New()
	}

	if client, ok := httpclient.FromContext(ctx); ok {
		page, err := client.Page()
		if err == nil {
			g.Page = page
		}
	}

	return g
}
