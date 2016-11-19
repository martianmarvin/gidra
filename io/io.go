package io

import (
	"io"
	"sort"
	"sync"
)

var (
	adaptersMu     sync.RWMutex
	readerAdapters = make(map[string]io.Reader)
	writerAdapters = make(map[string]io.Writer)
)

// RegisterReader registers a new adapter to read input from
func RegisterReader(name string, adapter io.Reader) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	if adapter == nil {
		panic("io: Register adapter is nil")
	}
	if _, dup := readerAdapters[name]; dup {
		panic("io: Register called twice for adapter " + name)
	}
	readerAdapters[name] = adapter
}

// RegisterWriter registers a new adapter to write output to
func RegisterWriter(name string, adapter io.Writer) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	if adapter == nil {
		panic("io: Register adapter is nil")
	}
	if _, dup := writerAdapters[name]; dup {
		panic("io: Register called twice for adapter " + name)
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
