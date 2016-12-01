package datasource

import "io"

// WriterFunc returns an open connection
type WriterFunc func(r io.Writer) (WriteableTable, error)

// WriteableTable represents a table that can be written to
// The table is held in memory and does not write to file until Flush() is called
type WriteableTable interface {
	io.WriterTo

	// SetColumns sets the headers for this table
	SetColumns([]string)

	// Append atomically adds a single row to the table.
	// Append must guarantee insertion order and be safe for concurrent
	// access from multiple goroutines
	Append(*Row) error

	// Close closes the underlying data writer
	Close() error
}
