package table

import (
	"bytes"
	"io"
	"sync/atomic"

	tablib "github.com/agrison/go-tablib"
	"github.com/martianmarvin/gidra/datasource"
)

// Reader provides support for all tablib data sources supporting the
// tablib.Dataset type.
// Reader implements the datasource.ReadableTable interface.

type Reader struct {
	// The current row we are iterating through
	// IMPORTANT : The index must be the first field defined in the struct
	// to prevent a panic in asm_386.s caused by alignment error on 32-bit
	index int64

	dataset *tablib.Dataset

	// Open loads raw data into a *tablib.Dataset used by this reader
	Open importFunc
}

func NewReader(importer importFunc) *Reader {
	reader := &Reader{
		Open:    importer,
		dataset: &tablib.Dataset{},
	}
	return reader
}

//Reads all data from the underlying reader
func (r *Reader) ReadFrom(reader io.Reader) (n int64, err error) {
	var buf bytes.Buffer
	n, err = buf.ReadFrom(reader)
	if err != nil {
		return n, err
	}
	r.index = 0
	r.dataset, err = r.Open(buf.Bytes())
	return n, err
}

// Builds a datasource.Row from the underlying dataset at specified index
func (r *Reader) buildRow(index int64) (row *datasource.Row, err error) {
	datarow, err := r.dataset.Row(int(index))
	if err != nil {
		return
	}
	row = datasource.NewRow()
	row.Index = index
	row.SetColumns(r.dataset.Headers())
	row.SetMap(datarow)

	return row, err

}

// Next returns rows starting from the first non-header row
func (r *Reader) Next() (*datasource.Row, error) {
	index := r.Index()
	if index >= r.Len() {
		return nil, io.EOF
	} else {
		atomic.AddInt64(&r.index, 1)
		return r.buildRow(index)
	}
}

func (r *Reader) Columns() []string {
	return r.dataset.Headers()
}

func (r *Reader) Index() int64 {
	return atomic.LoadInt64(&r.index)
}

func (r *Reader) Len() int64 {
	return int64(r.dataset.Height())
}

// Close removes the underlying dataset so it can be garbage collected
func (r *Reader) Close() error {
	var err error
	r.dataset = &tablib.Dataset{}
	return err
}
