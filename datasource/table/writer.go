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
	dataset *tablib.Dataset

	// Open is the opener that converts a dataset into exportable format
	Open exportFunc
}

func NewWriter(exporter exportFunc) *Writer {
	return &Writer{
		Open:    exporter,
		dataset: tablib.NewDataset(nil),
	}
}

// SetColumns sets headers on this table
func (w *Writer) SetColumns(cols []string) error {
	var err error
	height := w.dataset.Height()
	width := w.dataset.Width()
	hwidth := len(cols)

	if height == 0 {
		w.dataset = tablib.NewDataset(cols)
		return nil
	}

	// Ensure width of current dataset matches number of columns
	if width > hwidth {
		cols2 := make([]string, width-hwidth)
		cols = append(cols, cols2...)
	} else if width < hwidth {
		// Create empty value for each cell in new column
		vals := make([]interface{}, height)

		for i := width; i < hwidth; i++ {
			w.dataset.AppendColumn(cols[i-1], vals)
		}
	}

	dataset2 := tablib.NewDataset(cols)
	w.dataset, err = dataset2.Stack(w.dataset)
	return err
}

func (w *Writer) Append(row *datasource.Row) error {
	// Pad with extra columns from row if needed
	start := w.dataset.Width()
	newcols := row.Columns()[start:]
	cols := append(w.dataset.Headers(), newcols...)
	err := w.SetColumns(cols)
	if err != nil {
		return err
	}

	var vals []interface{}
	for _, val := range row.Values() {
		vals = append(vals, val.Interface())
	}
	return w.dataset.Append(vals)
}

func (w *Writer) WriteTo(writer io.Writer) (n int64, err error) {
	if w.dataset.Height() == 0 {
		return 0, err
	}
	exportable, err := w.Open(w.dataset)
	if err != nil {
		return 0, err
	}
	return exportable.WriteTo(writer)
}

// Close removes the underlying dataset so it can be garbage collected
func (w *Writer) Close() error {
	var err error
	w.dataset = tablib.NewDataset(nil)
	return err
}
