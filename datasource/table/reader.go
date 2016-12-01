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
	// The underlying reader to get data from
	reader io.Reader

	dataset *tablib.Dataset

	// The current row we are iterating through
	index int64

	// OpenFunc loads raw data into a *tablib.Dataset used by this reader
	OpenFunc importFunc
}

func NewReader(importer importFunc) *Reader {
	reader := &Reader{
		OpenFunc: importer,
	}
	return reader
}

//Reads all data from the underlying reader
func (r *Reader) ReadFrom(reader io.Reader) (n int64, err error) {
	var buf *bytes.Buffer
	r.reader = reader
	n, err = buf.ReadFrom(r.reader)
	if err != nil {
		return n, err
	}
	r.dataset, err = r.OpenFunc(buf.Bytes())
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
	row.AppendMap(datarow)

	return row, err

}

func (r *Reader) Next() (*datasource.Row, error) {
	index := atomic.AddInt64(&r.index, 1)
	if index >= r.Len() {
		return nil, io.EOF
	} else {
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
