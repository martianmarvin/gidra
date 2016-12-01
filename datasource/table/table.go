// Package table provides adapters for reading and writing tabular data
package table

import (
	"bytes"
	"io"
	"sync/atomic"

	tablib "github.com/agrison/go-tablib"
	"github.com/martianmarvin/gidra/datasource"
)

func init() {
	// Register supported adapters
	for format, importer := range importers {
		datasource.RegisterReader(format, func(r io.Reader) (datasource.ReadableTable, error) {
			var err error
			reader := NewReader(importer)
			_, err = reader.ReadFrom(r)
			return reader, err
		})
	}

	for format, exporter := range exporters {
		datasource.RegisterWriter(format, func(w io.Writer) (datasource.WriteableTable, error) {
			var err error
			writer := NewWriter(exporter)
			return writer, err
		})
	}
}

// importFunc loads provided data into a specific format
type importFunc func(data []byte) (*tablib.Dataset, error)

// exportFunc exports the dataset into a specific format
type exportFunc func(d *tablib.Dataset) (*tablib.Exportable, error)

var (
	importers = map[string]importFunc{
		"csv":  tablib.LoadCSV,
		"tsv":  tablib.LoadTSV,
		"json": tablib.LoadJSON,
		"xml":  tablib.LoadXML,
		"yaml": tablib.LoadYAML,
	}

	exporters = map[string]exportFunc{
		"csv":  dataCSV,
		"tsv":  dataTSV,
		"xlsx": dataXLSX,
		"html": dataHTML,
		"json": dataJSON,
		"sql":  dataSQL,
	}
)

// TODO Add remaining tablib supported formats
// Wrappers for tablib functions

func dataCSV(d *tablib.Dataset) (*tablib.Exportable, error) {
	return d.CSV()
}

func dataTSV(d *tablib.Dataset) (*tablib.Exportable, error) {
	return d.TSV()
}

func dataXLSX(d *tablib.Dataset) (*tablib.Exportable, error) {
	return d.XLSX()
}

func dataHTML(d *tablib.Dataset) (*tablib.Exportable, error) {
	return d.HTML(), nil
}

func dataJSON(d *tablib.Dataset) (*tablib.Exportable, error) {
	return d.JSON()
}

func dataSQL(d *tablib.Dataset) (*tablib.Exportable, error) {
	return d.MySQL("table"), nil
}

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

// Writer provides support for all tablib data sources supporting the
// tablib.Exportable type.
// Writer implements the WriteConn interface.

type Writer struct {
	// The underlying writer
	writer io.Writer

	dataset *tablib.Dataset

	// OpenFunc is the opener that converts a dataset into exportable format
	OpenFunc exportFunc
}

func NewWriter(exporter exportFunc) *Writer {
	return &Writer{
		OpenFunc: exporter,
	}
}

// SetColumns sets headers on this table
func (w *Writer) SetColumns(cols []string) {
	dataset2 := tablib.NewDataset(cols)
	w.dataset, _ = dataset2.Stack(w.dataset)
}

func (w *Writer) Append(row *datasource.Row) error {
	var err error
	var vals []interface{}
	for _, val := range row.Values() {
		vals = append(vals, val.Interface())
	}
	w.dataset.Append(vals)
	return err
}

func (w *Writer) WriteTo(writer io.Writer) (n int64, err error) {
	if w.dataset.Height() == 0 {
		return 0, err
	}
	exportable, err := w.OpenFunc(w.dataset)
	if err != nil {
		return 0, err
	}
	return exportable.WriteTo(writer)
}

// Flush exports the entire dataset to this Writer's format and writes it to the
// underlying io.Writer
func (w *Writer) Flush() error {
	var err error
	_, err = w.WriteTo(w.writer)
	return err
}

// Close removes the underlying dataset so it can be garbage collected
func (w *Writer) Close() error {
	var err error
	w.dataset = &tablib.Dataset{}
	return err
}

// TableRow wraps tablib.Dataset and implements the Row interface
type TableRow struct {
	// The dataset this row is part of
	dataset *tablib.Dataset

	//Number of the row this represents
	index int64
}

func (row *TableRow) Columns() []string {
	return row.dataset.Headers()
}

func (row *TableRow) Values() []interface{} {
	var vals []interface{}
	datarow, err := row.dataset.Row(int(row.index))
	if err != nil {
		return nil
	}
	for _, v := range datarow {
		vals = append(vals, v)
	}
	return vals
}

func (row *TableRow) Index() int64 {
	return row.index
}

// Adapters for individual input file types

// TableStringAdapter reads tabular data from a string
type TableStringAdapter struct {
}

// Open
// func (t *TableStringAdapter) Open(str string) (Conn, error) {

// }
