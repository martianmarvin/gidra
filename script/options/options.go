// Package options provides a configuration struct that holds options parsed
// from a Gidra script. Script parsers in package gidra/parser are
// responsible for configuring the ScriptOptions struct
package options

import (
	"time"

	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/sequence"
	"github.com/martianmarvin/vars"
)

type ScriptOptions struct {
	// How many times the script should loop total
	Loop int

	// Threads is the number of concurrent sequences executing
	Threads int

	// Verbosity determines the log level
	Verbosity int

	// Timeout for each individual task
	TaskTimeout time.Duration

	// Global variables available to all tasks
	Vars *vars.Vars

	// Datasources to read from
	Input map[string]datasource.ReadableTable

	// Proxies to use for requests
	Proxies datasource.ReadableTable

	// File to write output, or stdout if none
	Output datasource.WriteableTable

	// Sequence to run before the main loop
	BeforeSequence *sequence.Sequence

	// The main sequence to loop
	MainSequence *sequence.Sequence

	// Sequence to run after the main loop
	AfterSequence *sequence.Sequence
}

// New initializes an empty set of options
func New() *ScriptOptions {
	return &ScriptOptions{
		Vars:  vars.New(),
		Input: make(map[string]datasource.ReadableTable),
	}
}
