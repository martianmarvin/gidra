package datasource

import "io"

type ReadableTable interface {
	Table

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

	// Value returns the most recent row requested with Next(), or nil if there
	// is not one
	Value() *Row

	//Index atomically returns the current position of the table
	Index() int64

	// Len returns the total number of rows in the datasource
	// If the total is not known, like for a streaming datasource, it should
	// return 0
	Len() int64
}
