// Package config includes global configuration variables and defaults
// Config should not import any other subpackages to avoid circular imports
package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/olebedev/config"
)

type contextKey int

const (
	ctxConfig contextKey = iota
)

// Default config. Overridden from script file or environment
var defaultConfig = `
config:
	verbosity: 4
	threads: 100
	task_timeout: 15
	http:
		follow_redirects: false
		headers:
			user-agent: Mozilla/5.0 (Windows NT 6.1; rv:45.0) Gecko/20100101 Firefox/45.0
			accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
			accept-language: en-US,en;q=0.5
			accept-encoding: gzip, deflate
loop: 1
`

// Config wraps the config.Config struct with additional helper methods
type Config struct {
	*config.Config
}

// New initializes a new Config
func New() *Config {
	return &Config{
		Config: &config.Config{},
	}
}

func (cfg *Config) Get(path string) (*Config, error) {
	subcfg, err := cfg.Config.Get(path)
	return &Config{Config: subcfg}, err
}

// UGet returns the result of Get() or the default
func (cfg *Config) UGet(path string, def *Config) *Config {
	c, err := cfg.Get(path)
	if err == nil {
		return c
	} else {
		return def
	}
}

// StringMap returns a map with string values
func (cfg *Config) StringMap(path string) (map[string]string, error) {
	m, err := cfg.Map(path)
	if err != nil {
		return nil, err
	}
	sm := make(map[string]string)
	for k, v := range m {
		sm[k] = fmt.Sprint(v)
	}
	return sm, nil
}

//UStringMap returns a map[string]string or defaults
func (cfg *Config) UStringMap(path string, defaults ...map[string]string) map[string]string {
	defs := make([]map[string]interface{}, 0)
	for _, def := range defaults {
		dm := make(map[string]interface{})
		for k, v := range def {
			dm[k] = v
		}
		defs = append(defs, dm)
	}

	m := cfg.UMap(path, defs...)
	sm := make(map[string]string)
	for k, v := range m {
		sm[k] = fmt.Sprint(v)
	}
	return sm
}

// StringList returns a list of string values
func (cfg *Config) StringList(path string) ([]string, error) {
	list := make([]string, 0)
	l, err := cfg.List(path)
	if err != nil {
		return nil, err
	}
	for _, v := range l {
		list = append(list, fmt.Sprint(v))
	}
	return list, nil
}

// UStringList returns a list of string values or the default
func (cfg *Config) UStringList(path string, defaults ...[]string) []string {
	defs := make([][]interface{}, 0)
	for _, def := range defaults {
		dl := make([]interface{}, len(def))
		for i, v := range def {
			dl[i] = v
		}
		defs = append(defs, dl)
	}
	l := cfg.UList(path, defs...)

	list := make([]string, len(l))
	for j, v := range l {
		list[j] = fmt.Sprint(v)
	}
	return list
}

// The default top level config object
var cfg *Config

// ParseYAML creates a new config from a yaml file
func ParseYaml(text string) (*Config, error) {
	formatted := strings.Replace(strings.TrimSpace(text), "\t", "    ", -1)
	c, err := config.ParseYaml(formatted)
	if err != nil {
		return nil, err
	}
	cfg := New()
	cfg.Config = c
	return cfg, nil
}

// Must panics if there is an error
func Must(cfg *Config, err error) *Config {
	if err != nil {
		panic("config: " + err.Error())
	}
	return cfg
}

func init() {
	cfg = Must(ParseYaml(defaultConfig))
}

// Global settings not used inside tasks, therefore not subject to config file
var (
	// Location of script files
	ScriptDir = "./scripts"
)

// Default returns the default config
func Default() *Config {
	return cfg
}

// ToContext returns a context with the provided config merged into the one in
// the context
func ToContext(ctx context.Context, c *Config) context.Context {
	old := FromContext(ctx)
	merged := config.Must(old.Extend(c.Config))
	return context.WithValue(ctx, ctxConfig, merged)
}

// FromContext returns a config from the context, or the default if the context
// does not have one
func FromContext(ctx context.Context) *Config {
	if c, ok := ctx.Value(ctxConfig).(*Config); ok {
		return c
	} else {
		return Default()
	}
}
