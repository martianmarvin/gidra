package datasource

// Table is the basic interface supporting readers/writers
type Table interface {
	// Filter adds a filter to the datasource to process results before they
	// are returned
	Filter(FilterFunc) error

	// Close closes the underlying datasource
	Close() error
}
