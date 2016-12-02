// Package datasource provides utilities to read and write files and other
// input/output tabular data sources
package datasource

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrUnsupportedType = errors.New("This file type is not supported")
)

// ReaderFunc returns a table adapter for the specified format
type ReaderFunc func() ReadableTable

// WriterFunc returns a table adapter for the specified format
type WriterFunc func() WriteableTable

var (
	adaptersMu     sync.RWMutex
	readerAdapters = make(map[string]ReaderFunc)
	writerAdapters = make(map[string]WriterFunc)
)

// RegisterReader registers a new adapter to read input from
func RegisterReader(name string, adapter ReaderFunc) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	if adapter == nil {
		panic("datasource: Register adapter is nil")
	}
	if _, dup := readerAdapters[name]; dup {
		panic("datasource: Register called twice for adapter " + name)
	}
	readerAdapters[name] = adapter
}

// RegisterWriter registers a new adapter to write output to
func RegisterWriter(name string, adapter WriterFunc) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	if adapter == nil {
		panic("datasource: Register adapter is nil")
	}
	if _, dup := writerAdapters[name]; dup {
		panic("datasource: Register called twice for adapter " + name)
	}
	writerAdapters[name] = adapter
}

// Readers returns a sorted list of the names of the registered read adapters.
func Readers() []string {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	var list []string
	for name := range readerAdapters {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// Writers returns a sorted list of the names of the registered write adapters.
func Writers() []string {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	var list []string
	for name := range writerAdapters {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// NewReader returns a read adapter for the specified data format
func NewReader(format string) (ReadableTable, error) {
	var err error
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	fn, ok := readerAdapters[format]
	if !ok {
		err = ErrUnsupportedType
		return nil, err
	}
	return fn(), err
}

// NewWriter returns a write adapter for the specified data format
func NewWriter(format string) (WriteableTable, error) {
	var err error
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	fn, ok := writerAdapters[format]
	if !ok {
		err = ErrUnsupportedType
		return nil, err
	}
	return fn(), err
}
