package datasource

import (
	"errors"
	"io"
)

// ReadIterator is a ReadableTable wrapper that allows reading the most recent
// value after Next() has been called, as well as rewinding
// ReadIterator is NOT concurrency safe, and should not be used from multiple
// goroutines
type ReadIterator struct {
	ReadableTable

	values []*Row

	index int64
}

func NewReadIterator(r ReadableTable) *ReadIterator {
	return &ReadIterator{
		ReadableTable: r,
		values:        make([]*Row, 0),
	}
}

func (r *ReadIterator) Next() (*Row, error) {
	if int64(len(r.values)) < r.index {
		row, err := r.ReadableTable.Next()
		if err != nil && err != io.EOF {
			return nil, err
		}
		r.index += 1
		r.values = append(r.values, row)
		return row, nil
	} else {
		r.index += 1
		return r.values[r.index-1], nil
	}
}

// Value returns the most recent value from the iterator
func (r *ReadIterator) Value() (*Row, error) {
	if r.index == 0 {
		return nil, errors.New("No value in iterator")
	}

	return r.values[r.index-1], nil
}

// Back rewinds the iterator one position
func (r *ReadIterator) Back() error {
	if r.index == 0 {
		return errors.New("Already at beginning")
	}
	r.index -= 1
	return nil
}

func (r *ReadIterator) Rewind() error {
	if len(r.values) == 0 {
		return errors.New("Iterator is empty")
	}
	r.index = 0
	return nil
}
