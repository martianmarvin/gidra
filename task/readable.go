package task

import "github.com/martianmarvin/gidra/datasource"

type readable struct {
	row *datasource.Row
}

// A Readable task reads structured variables from an input table
type Readable interface {
	// SetRow provides this task with structured input
	SetRow(row *datasource.Row)
}

func NewReadable() Readable {
	return &readable{row: datasource.NewRow()}
}

func (r *readable) SetRow(row *datasource.Row) {
	r.row = row
}
