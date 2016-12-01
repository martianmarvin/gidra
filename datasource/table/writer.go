package table

import (
	"io"

	tablib "github.com/agrison/go-tablib"
	"github.com/martianmarvin/gidra/datasource"
)

// Writer provides support for all tablib data sources supporting the
// tablib.Exportable type.
// Writer implements the WriteConn interface.

type Writer struct {
	// The underlying writer
	writer io.Writer

	dataset *tablib.Dataset

	// OpenFunc is the opener that converts a dataset into exportable format
	OpenFunc exportFunc
}

func NewWriter(exporter exportFunc) *Writer {
	return &Writer{
		OpenFunc: exporter,
	}
}

// SetColumns sets headers on this table
func (w *Writer) SetColumns(cols []string) {
	dataset2 := tablib.NewDataset(cols)
	w.dataset, _ = dataset2.Stack(w.dataset)
}

func (w *Writer) Append(row *datasource.Row) error {
	var err error
	var vals []interface{}
	for _, val := range row.Values() {
		vals = append(vals, val.Interface())
	}
	w.dataset.Append(vals)
	return err
}

func (w *Writer) WriteTo(writer io.Writer) (n int64, err error) {
	if w.dataset.Height() == 0 {
		return 0, err
	}
	exportable, err := w.OpenFunc(w.dataset)
	if err != nil {
		return 0, err
	}
	return exportable.WriteTo(writer)
}

// Flush exports the entire dataset to this Writer's format and writes it to the
// underlying io.Writer
func (w *Writer) Flush() error {
	var err error
	_, err = w.WriteTo(w.writer)
	return err
}

// Close removes the underlying dataset so it can be garbage collected
func (w *Writer) Close() error {
	var err error
	w.dataset = &tablib.Dataset{}
	return err
}
