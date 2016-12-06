package datasource

import (
	"io"
	"sync/atomic"
)

//Nopreader is a reader that returns dummy empty rows for the specified
//number of iterations

type NopReader struct {
	index int64
	len   int64
}

// NewNopReader returns a Reader that returns empty rows for n iterations, or
// forever if n<=0
func NewNopReader(n int) *NopReader {
	return &NopReader{len: int64(n)}
}

func (r *NopReader) ReadFrom(reader io.Reader) (n int64, err error) {
	return 0, nil
}

func (r *NopReader) Columns() (cols []string) {
	return cols
}

func (r *NopReader) Next() (*Row, error) {
	index := r.Index()
	len := r.Len()
	if len <= 0 || index < len {
		row := NewRow()
		row.Index = index
		atomic.AddInt64(&r.index, 1)
		return row, nil
	} else {
		return nil, io.EOF
	}
}

func (r *NopReader) Index() int64 {
	return atomic.LoadInt64(&r.index)
}
func (r *NopReader) Len() int64 {
	return r.len
}

func (r *NopReader) Close() error {
	return nil
}
