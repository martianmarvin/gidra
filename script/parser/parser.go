package parser

import (
	"sync"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/log"
	"github.com/martianmarvin/gidra/script/options"
)

// Global logger for debugging
var Logger log.Log

// ParseFunc parses a top-level item in the config
type ParseFunc func(s *options.ScriptOptions, cfg *config.Config) error

var (
	parsersMu     sync.RWMutex
	configParsers = make(map[string]ParseFunc)

	// ParseFirst defines a list of standard options. Options in Keys are parsed
	// before the other options. Keys are parsed in the
	// order they appear the slice.
	ParseFirst = []string{cfgConfig, cfgVars, cfgInputs, cfgOutput}
)

func init() {
	Logger = log.Logger().WithField("task", "gidra.parser")
}

// For returns the registered parser function for that key, and nil, false if it
// does not exist
func For(key string) (ParseFunc, bool) {
	parsersMu.RLock()
	defer parsersMu.RUnlock()
	if fn, ok := configParsers[key]; ok {
		return fn, true
	}
	if alt, ok := cfgAliases[key]; ok {
		if fn, ok := configParsers[alt]; ok {
			return fn, true
		}
	}
	return nil, false
}

// Register registers a new parser for the specified top-level key
func Register(key string, fn ParseFunc) {
	parsersMu.Lock()
	defer parsersMu.Unlock()
	if fn == nil {
		panic("Invalid parser")
	}
	if _, dup := configParsers[key]; dup {
		panic("Register called twice for parser " + key)
	}
	configParsers[key] = wrapParser(key, fn)
}

// Configure parses a config and applies it to the ScriptOptions
func Configure(s *options.ScriptOptions, cfg *config.Config) error {
	var keys []string

	parsed := make(map[string]bool)
	// Add keys in ParseFirst to the chain before the others
	for _, key := range ParseFirst {
		if _, ok := cfg.CheckGet(key); ok {
			// key is in the ParseFirst list
			parsed[key] = true
			keys = append(keys, key)
		}
	}

	// Iterate remaining top level keys
	for key, _ := range cfg.UMap("") {
		if parsed[key] {
			continue
		}
		keys = append(keys, key)
		parsed[key] = true
	}

	for _, key := range keys {
		parser, ok := For(key)
		if !ok {
			return KeyError{key}
		}
		conf, ok := cfg.CheckGet(key)
		if !ok {
			return KeyError{key}
		}
		err := parser(s, conf)
		if err != nil {
			return err
		}
	}

	return nil
}

// Wraps a ParseFunc to add better error reporting, logging, etc
func wrapParser(key string, fn ParseFunc) ParseFunc {
	return func(s *options.ScriptOptions, cfg *config.Config) error {
		err := fn(s, cfg)

		if err == nil {
			return err
		}
		entry := Logger.WithField("type", "Unknown").WithField("key", key)
		switch err := err.(type) {
		case FieldError:
			entry = entry.WithField("type", "FieldError").WithField("field", err.Name)
		case KeyError:
			entry = entry.WithField("type", "KeyError").WithField("key", err.Name)
		case ValueError:
			entry = entry.WithField("type", "ValueError").WithField("value", err.Name)
		}
		entry.Error(err)

		return err
	}

}
