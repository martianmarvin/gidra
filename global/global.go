// Package global represents global objects available to the end-user in
// their tasks and templates
package global

import (
	"context"

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
	Proxy *List

	// Page is the page requested by the last request
	Page *client.Page

	// Loop is the number of the current iteration of the task loop
	Loop int

	// Status of the last executed task
	Status Status

	// Inputs are the user-defined datasources
	Inputs map[string]datasource.ReadableTable

	// Data represents the current row of data from the main input as a map
	Data *datasource.Row

	// Result is the resulting output returned by the task, if any
	Result *datasource.Row
}

// New instantiates a new Global object
func New() *Global {
	return &Global{
		Vars:   make(map[string]interface{}),
		Inputs: make(map[string]datasource.ReadableTable),
		Data:   datasource.NewRow(),
		Result: datasource.NewRow(),
		Page:   client.NewPage(),
	}
}

// Copy returns a shallow copy of the Global
func (g *Global) Copy() *Global {
	g2 := New()
	*g2 = *g
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

	// Use first input to set loop
	if len(g.Inputs) > 0 {
		for _, input := range g.Inputs {
			g.Loop = int(input.Index())
			break
		}
	}

	return g
}
