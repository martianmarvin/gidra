package datasource

// FilterFunc filters or transforms a Row in some way. It either returns a
// new *Row or nil if the row was filtered out. FilterFunc can be chained to
// build composable pipelines
type FilterFunc func(row *Row) *Row

// EqualsFilter excludes rows that don't have a key equal to a particular
// string value
func StringEqualsFilter(key, val string) FilterFunc {
	return FilterFunc(func(row *Row) *Row {
		if row == nil {
			return nil
		}
		field, err := row.Get(key).String()
		if err != nil {
			return nil
		}
		if field != val {
			return nil
		}
		return row
	})
}

// PrependFilter prepends the given text to the specified key
func PrependFilter(key, val string) FilterFunc {
	return FilterFunc(func(row *Row) *Row {
		if row == nil {
			return nil
		}
		oldval, err := row.Get(key).String()
		if err != nil {
			return nil
		}
		return row.Set(key, val+oldval)
	})
}
