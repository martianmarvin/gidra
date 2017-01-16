package datasource

import (
	"fmt"

	simplejson "github.com/bitly/go-simplejson"
)

// Row is a single row of data from a datasource
// Row wraps simplejson
type Row struct {
	// headers are the columns for this row's table
	headers []string

	data *simplejson.Json

	// Index represents this row's position in the table. For SQL data, this is
	// typically the row's id.
	Index int64
}

func NewRow() *Row {
	return &Row{
		headers: make([]string, 0),
		data:    simplejson.New(),
	}
}

// String satisfies the fmt.Stringer interface
func (row *Row) String() string {
	return fmt.Sprint(row.Strings())
}

// Columns returns this row's headers
func (row *Row) Columns() []string {
	return row.headers
}

// HasColumn returns the position of the column in the row's headers, or -1 if
// it does not exist
func (row *Row) ColumnIndex(col string) int {
	for i, v := range row.headers {
		if v == col {
			return i
		}
	}
	return -1
}

//SetColumns sets the headers for this row
func (row *Row) SetColumns(cols []string) {
	row.headers = cols
}

// Map returns this row as a map with column headers
func (row *Row) Map() map[string]interface{} {
	return row.data.MustMap()
}

// Get returns a simplejson value for the specified key
func (row *Row) Get(key string) *simplejson.Json {
	return row.data.Get(key)
}

// GetIndex returns a simplejson value for the specified column number
func (row *Row) GetIndex(i int) *simplejson.Json {
	if i >= len(row.headers) {
		return simplejson.New()
	} else {
		return row.Get(row.headers[i])
	}
}

// Values returns a slice of this row's values only
func (row *Row) Values() []*simplejson.Json {
	var vals []*simplejson.Json
	for _, k := range row.headers {
		vals = append(vals, row.data.Get(k))
	}
	return vals
}

// Bytes returns the row's values as a slice of byte arrays
func (row *Row) Bytes() [][]byte {
	var vals [][]byte
	for _, k := range row.headers {
		vals = append(vals, []byte(row.data.Get(k).MustString()))
	}
	return vals
}

// Strings returns this row's values as a slice of strings
func (row *Row) Strings() []string {
	var vals []string
	for _, k := range row.headers {
		vals = append(vals, row.data.Get(k).MustString())
	}
	return vals
}

// Len returns the number of values in the row, which may be different from the
// number of columns
func (row *Row) Len() int {
	return len(row.data.MustMap())
}

// TODO Use string values instead of interface{}
// Append adds values to the row. It returns the new row, so it can be chained
func (row *Row) Append(vals ...interface{}) *Row {
	hwidth := len(row.headers)
	width := row.Len()
	// pad headers if needed
	for i := hwidth; i < (width + len(vals)); i++ {
		row.headers = append(row.headers, fmt.Sprint(i))
	}

	for i, val := range vals {
		pos := width + i
		row.data.Set(row.headers[pos], val)
	}

	return row
}

// AppendKV adds a new header and corresponding value to the row and returns
// the new row
func (row *Row) AppendKV(key string, val interface{}) *Row {
	row.headers = append(row.headers, key)
	row.Append(val)
	return row
}

// AppendMap unzips the provided map and appends keys to row headers and values
// to row values and returns the new row
func (row *Row) AppendMap(m map[string]interface{}) *Row {
	for key, val := range m {
		row.AppendKV(key, val)
	}
	return row
}

// SetMap works like AppendMap, but replaces the existing value for each of the
// map's headers that already exist in this row's columns and returns the new
// row. In most cases, SetMap should be used after calling SetColumns to set
// the row's headers
func (row *Row) SetMap(m map[string]interface{}) *Row {
	for key, val := range m {
		if row.ColumnIndex(key) >= 0 {
			row.data.Set(key, val)
		} else {
			row.AppendKV(key, val)
		}
	}
	return row
}
