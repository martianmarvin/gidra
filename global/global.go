// Package global represents global objects available to the end-user in
// their tasks and templates
package global

import (
	"context"

	"github.com/martianmarvin/gidra/client"
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
	return g
}
