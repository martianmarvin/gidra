package table

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/martianmarvin/gidra/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sampleCSV = `id,first name,last name,email
1,Lefty,"O'Toole",lefty@example.com
2,John,,john@example.com
`
)

func TestAdapters(t *testing.T) {
	allReaders := datasource.Readers()
	allWriters := datasource.Writers()
	for format, _ := range importers {
		assert.Contains(t, allReaders, format)
	}
	for format, _ := range exporters {
		assert.Contains(t, allWriters, format)
	}
}

func TestReader(t *testing.T) {
	var err error
	format := "csv"
	r := strings.NewReader(sampleCSV)
	reader, err := datasource.NewReader(format, r)
	require.Nil(t, err)

	assert.Len(t, reader.Columns(), 4)
	assert.Contains(t, reader.Columns(), "email")

	assert.Equal(t, reader.Len(), 2)

	for {
		row, err := reader.Next()
		if err == io.EOF {
			break
		}
		assert.Nil(t, err)
		assert.Equal(t, row.Index, reader.Index())
		val := row.Get("Email")
		if assert.NotNil(t, val) {
			assert.Contains(val, "example.com")
		}
	}

	err = reader.Close()
	assert.Nil(t, err)
}

func TestWriter(t *testing.T) {
	var err error
	format := "csv"
	var buf bytes.Buffer

	w := bufio.NewWriter(&buf)
	writer, err := datasource.NewWriter(format, w)
	require.Nil(t, err)

	r := strings.NewReader(sampleCSV)
	reader, err := datasource.NewReader(format, r)

	writer.SetColumns(reader.Columns())
	for row, err := reader.Next(); err != io.EOF; {
		err = writer.Append(row)
		assert.Nil(t, err)
	}

	w.Flush()
	output := buf.String()
	assert.Equal(sampleCSV, output)

	err = writer.Close()
	assert.Nil(t, err)
}
