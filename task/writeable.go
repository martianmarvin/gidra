package task

import "github.com/martianmarvin/gidra/datasource"

type writeable struct {
	row *datasource.Row
}

// A Writeable task writes structured results to the main output
type Writeable interface {
	// Row returns the output from this task for writing
	Row() *datasource.Row
}

func NewWriteable() Writeable {
	return &writeable{row: datasource.NewRow()}
}

func (w *writeable) Row() *datasource.Row {
	return w.row
}
