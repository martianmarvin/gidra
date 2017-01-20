package parser

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/script/options"
)

func init() {
	Register(cfgInputs, inputParser)
	Register(cfgOutput, outputParser)
}

func inputParser(s *options.ScriptOptions, cfg *config.Config) error {
	inputs, err := cfg.GetConfigSliceE(cfgInputs)
	if err != nil {
		return err
	}
	parsed, err := parseInputs(inputs)
	if err != nil {
		return err
	}
	s.Input = parsed

	return nil
}

// Creates an input datasource based on the config
// If there is no input defined, creates a dummy reader that iterates for the
// specified number of loop iterations
func parseInputs(inputs []*config.Config) (map[string]datasource.ReadableTable, error) {
	readers := make(map[string]datasource.ReadableTable)

	var n int
	for _, subcfg := range inputs {
		name := subcfg.GetString(cfgIOVars)
		if len(name) == 0 {
			if readers[cfgMainInput] == nil {
				name = cfgMainInput
			} else {
				// Inputs without a name address by number
				name = fmt.Sprint(n)
				n += 1
			}
		}
		input, err := parseInput(subcfg)
		if err != nil {
			return readers, err
		}
		readers[name] = input
	}
	return readers, nil
}

//Parse a single input
func parseInput(inputcfg *config.Config) (datasource.ReadableTable, error) {
	source, err := inputcfg.GetStringE(cfgIOSource)
	if err != nil {
		return nil, config.KeyError{Name: cfgIOSource, Err: err}
	}

	adapter := strings.ToLower(inputcfg.GetString(cfgIOAdapter))

	return datasource.FromFileType(source, adapter)
}

func outputParser(s *options.ScriptOptions, cfg *config.Config) error {
	output := cfg.Get(cfgOutput, nil)
	if len(output.AllKeys()) == 0 {
		return nil
	}
	parsed, err := parseOutput(output)
	if err != nil {
		return err
	}
	s.Output = parsed

	return nil
}

func parseOutput(outputcfg *config.Config) (*datasource.WriteCloser, error) {
	source, err := outputcfg.GetStringE(cfgIOSource)
	if err != nil {
		return nil, config.KeyError{Name: cfgIOSource, Err: err}
	}

	adapter := strings.ToLower(outputcfg.GetString(cfgIOAdapter))
	if len(adapter) == 0 {
		// Guess adapter from file exetension
		adapter = strings.TrimLeft(path.Ext(source), ".")
	}

	writer, err := os.OpenFile(source, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	if writer == nil {
		return nil, config.KeyError{Name: cfgIOSource, Err: config.ErrRequired}
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
