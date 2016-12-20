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

// WriteTo writes the table's data to the specified writer, or to the underlying
// writer if nil is provided as the argument
func (w *WriteCloser) WriteTo(wr io.Writer) (int64, error) {
	if wr == nil {
		wr = w.writer
	}
	return w.WriteableTable.WriteTo(wr)
}
