package parser

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/martianmarvin/gidra"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/script/options"
)

func init() {
	Register(cfgInputs, inputParser)
	Register(cfgOutput, outputParser)
}

func inputParser(s *options.ScriptOptions, cfg *config.Config) error {
	inputs, err := cfg.MapList("")
	if err != nil {
		return err
	}
	parsed, err := parseInput(inputs)
	if err != nil {
		return err
	}
	s.Input = parsed

	return nil
}

// Creates an input datasource based on the config
// If there is no input defined, creates a dummy reader that iterates for the
// specified number of loop iterations
func parseInput(inputs []map[string]interface{}) (map[string]datasource.ReadableTable, error) {
	readers := make(map[string]datasource.ReadableTable)
	var err error

	for _, input := range inputs {
		source, ok := input[cfgIOSource].(string)
		if !ok || len(source) == 0 {
			return nil, gidra.FieldError{cfgIOSource}
		}

		adapter, ok := input[cfgIOAdapter].(string)
		if ok {
			adapter = strings.ToLower(adapter)
		}

		name, ok := input[cfgIOVars].(string)
		if !ok || len(name) == 0 {
			// No variable name provided
			// Main input defaults to first source without an 'as' declaration
			if readers["main"] == nil {
				reader, err := datasource.FromFileType(source, adapter)
				if err != nil {
					return nil, err
				}
				readers["main"] = reader
			} else {
				return nil, gidra.FieldError{cfgIOVars}
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

func outputParser(s *options.ScriptOptions, cfg *config.Config) error {
	var err error
	outputs, err := cfg.Map(cfgOutput)
	if err != nil {
		// No outputs, writing directly to stdout
		s.Output, err = outputWriter("tsv", os.Stdout)
		if err != nil {
			return err
		}
		return nil
	}
	parsed, err := parseOutput(outputs)
	if err == nil {
		s.Output = parsed
	}
	return err
}

func parseOutput(output map[string]interface{}) (*datasource.WriteCloser, error) {
	source, ok := output[cfgIOSource].(string)
	if !ok || len(source) == 0 {
		return nil, gidra.FieldError{cfgIOSource}
	}

	adapter, ok := output[cfgIOAdapter].(string)
	if ok {
		adapter = strings.ToLower(adapter)
	} else {
		// Guess adapter from file exetension
		adapter = strings.TrimLeft(path.Ext(source), ".")
	}

	writer, err := os.OpenFile(source, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	if writer == nil {
		return nil, gidra.FieldError{cfgIOSource}
	}

	return outputWriter(adapter, writer)
}

func outputWriter(adapter string, writer io.WriteCloser) (*datasource.WriteCloser, error) {
	w, err := datasource.NewWriter(adapter)
	if err != nil {
		return nil, err
	}
	return datasource.NewWriteCloser(w, writer), nil
}
