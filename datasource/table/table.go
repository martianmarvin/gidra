// Package table provides adapters for reading and writing tabular data
package table

import (
	tablib "github.com/agrison/go-tablib"
	"github.com/martianmarvin/gidra/datasource"
)

func init() {
	// Register supported adapters
	for format, importer := range importers {
		func(fn importFunc) {
			datasource.RegisterReader(format,
				func() datasource.ReadableTable {
					reader := NewReader(fn)
					return reader
				})
		}(importer)
	}

	for format, exporter := range exporters {
		func(fn exportFunc) {
			datasource.RegisterWriter(format,
				func() datasource.WriteableTable {
					writer := NewWriter(fn)
					return writer
				})
		}(exporter)
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
