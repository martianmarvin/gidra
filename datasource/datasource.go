// Package datasource provides utilities to read and write files and other
// input/output tabular data sources
package datasource

import (
	"io"
	"sort"
	"sync"
)

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

// NewReader returns a read connection for the specified data source type and name
// To use with an io.Reader, wrap it in an ioutil.NopCloser first
func NewReader(adapterName string, r io.ReadCloser) (ReadableTable, error) {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	fn, ok := readerAdapters[adapterName]
	if !ok {
		panic("datasource: No reader adapter registered of type " + adapterName)
	}
	return fn(r)
}

// NewWriter returns a read connection for the specified data source type and name
func NewWriter(adapterName string, w io.WriteCloser) (WriteableTable, error) {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	fn, ok := writerAdapters[adapterName]
	if !ok {
		panic("datasource: No writer adapter registered of type " + adapterName)
	}
	return fn(w)
}
