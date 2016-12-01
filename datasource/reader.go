package datasource

import "io"

// ReaderFunc returns an open connection
type ReaderFunc func(r io.Reader) (ReadableTable, error)

type ReadableTable interface {
	io.ReaderFrom

	// Columns returns the names of this table's headers, in the order they are
	// read/written.
	// If a particular column name isn't known, an empty
	// string should be returned for that entry.
	Columns() []string

	// Next reads the next row from this datasource
	// Next should return io.EOF when there are no more rows.
	// Next must advance atomically and be safe to call from multiple concurrent goroutines
	Next() (*Row, error)

	//Index atomically returns the current position of the table
	Index() int64

	// Len returns the total number of rows in the datasource
	// If the total is not known, like for a streaming datasource, it should
	// return 0
	Len() int64

	// Close closes the underlying data writer
	Close() error
}
