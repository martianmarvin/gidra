package datasource

import "io"

// WriteableTable represents a table that can be written to
type WriteableTable interface {
	io.WriterTo

	// SetColumns sets the headers for this table
	SetColumns([]string) error

	// Append atomically adds a single row to the table.
	// Append must guarantee insertion order and be safe for concurrent
	// access from multiple goroutines
	Append(*Row) error

	// Close closes the underlying data writer
	Close() error
}

// WriteCloser is a convenience struct that wraps a WriteableTable with an
// associated WriteCloser
type WriteCloser struct {
	WriteableTable
	writer io.WriteCloser
}

func NewWriteCloser(w WriteableTable, writer io.WriteCloser) *WriteCloser {
	return &WriteCloser{WriteableTable: w, writer: writer}
}

func (w *WriteCloser) Close() error {
	err := w.WriteableTable.Close()
	if err != nil {
		return err
	}
	return w.writer.Close()
}

// Flush writes the table's data to the underlying writer
func (w *WriteCloser) Flush() error {
	_, err := w.WriteTo(w.writer)
	return err
}
