package datasource

import "io"

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
