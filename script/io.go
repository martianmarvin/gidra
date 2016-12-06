package script

import (
	"io"
	"os"
	"strings"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
)

// Creates an input datasource based on the config
// If there is no input defined, creates a dummy reader that iterates for the
// specified number of loop iterations
func parseInput(cfg *config.Config) (map[string]datasource.ReadableTable, error) {
	readers := make(map[string]datasource.ReadableTable)
	var err error
	loopCount := cfg.UInt(cfgLoop, 1)
	inputs, err := cfg.MapList(cfgIOIn)
	if err != nil {
		// No inputs, create NopReader
		readers["main"] = datasource.NewNopReader(loopCount)
		return readers, nil
	}

	for _, input := range inputs {
		source, ok := input[cfgIOSource].(string)
		if !ok || len(source) == 0 {
			return nil, FieldError{cfgIOSource}
		}

		adapter, ok := input[cfgIOAdapter].(string)
		if ok {
			adapter = strings.ToLower(adapter)
		}

		name, ok := input[cfgIOVars].(string)
		if !ok || len(name) == 0 {
			//No variable name provided
			// Main input defaults to first source without an 'as' declaration
			if readers["main"] == nil {
				reader, err := datasource.FromFileType(source, adapter)
				if err != nil {
					return nil, err
				}
				readers["main"] = reader
			} else {
				return nil, FieldError{cfgIOVars}
			}
		} else {
			reader, err := datasource.FromFileType(source, adapter)
			if err != nil {
				return nil, err
			}
			readers[name] = reader
		}
	}

	return readers, err
}

func parseOutput(cfg *config.Config) (datasource.WriteCloser, error) {
	var adapter string
	var writer io.Writer
	output, err := cfg.MapList(cfgIOOut)
	if err != nil {
		// No outputs, writing directly to stdout
		adapter = "tsv"
		writer = os.Stdout
	} else {

		source, ok := output[cfgIOSource].(string)
		if !ok || len(source) == 0 {
			return nil, FieldError{cfgIOSource}
		}

		adapter, ok := output[cfgIOAdapter].(string)
		if ok {
			adapter = strings.ToLower(adapter)
		}

		writer, err := os.OpenFile(source, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
	}

	w, err := datasource.NewWriter(adapter)
	if err != nil {
		return err
	}
	if writer == nil {
		return nil, FieldError{cfgIOSource}
	}

	return datasource.NewWriteCloser(w, writer)

}
