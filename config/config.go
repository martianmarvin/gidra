// Package config includes global configuration variables and defaults
// Config should not import any other subpackages to avoid circular imports
package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

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
	loop: 1
	task_timeout: 15
http:
	follow_redirects: false
	headers:
		user-agent: Mozilla/5.0 (Windows NT 6.1; rv:45.0) Gecko/20100101 Firefox/45.0
		accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
		accept-language: en-US,en;q=0.5
		accept-encoding: gzip, deflate
`

// The default top level config object
var cfg *Config

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

// Get returns the config at path, or the default if not found
func (cfg *Config) Get(path string, def *Config) *Config {
	c, ok := cfg.CheckGet(path)
	if !ok {
		return def
	}
	return c
}

// CheckGet returns the config at path, or false if it does not exist
func (cfg *Config) CheckGet(path string) (*Config, bool) {
	subcfg, err := cfg.Config.Get(path)
	if err != nil {
		return nil, false
	}
	return &Config{Config: subcfg}, true
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

// MapList returns a slice of map[string]interface{}
func (cfg *Config) MapList(path string) ([]map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0)
	l, err := cfg.List(path)
	if err != nil {
		return nil, err
	}
	for i, _ := range l {
		m, err := cfg.Map(fmt.Sprintf("%s.%d", path, i))
		if err == nil {
			list = append(list, m)
		}
	}
	return list, nil
}

// CList returns a slice of Config objects that can be accessed with the
// other methods
func (cfg *Config) CList(path string) ([]*Config, error) {
	list := make([]*Config, 0)
	l, err := cfg.List(path)
	if err != nil {
		return nil, err
	}
	for i, _ := range l {
		c, err := cfg.Config.Get(fmt.Sprintf("%s.%d", path, i))
		if err == nil {
			list = append(list, &Config{Config: c})
		}
	}
	return list, nil
}

// Duration returns a time.Duration from a dictionary
func (cfg *Config) Duration(path string) (time.Duration, error) {
	var t time.Duration
	// Try parsing an int first, defaulting to seconds
	i, err := cfg.Int(path)
	if err == nil {
		t = time.Duration(i) * time.Second
		return t, nil
	}

	m, err := cfg.Map(path)
	if err != nil {
		return 0, err
	}

	for key, _ := range m {
		val, err := cfg.Int(path + "." + key)
		if err != nil {
			return 0, err
		}
		var unit time.Duration
		switch key {
		case "milliseconds", "ms":
			unit = time.Millisecond
		case "seconds", "s":
			unit = time.Second
		case "hours", "h":
			unit = time.Hour
		default:
			unit = 0
		}
		t += time.Duration(val) * unit
	}

	return t, nil
}

// URL returns a *url.URL from a string
func (cfg *Config) URL(path string) (*url.URL, error) {
	rawurl, err := cfg.String(path)
	if err != nil {
		return nil, err
	}
	return url.Parse(rawurl)
}

// ParseYAML creates a new config from a yaml file
func ParseYaml(r io.Reader) (*Config, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	text := buf.String()

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
	r := strings.NewReader(defaultConfig)
	cfg = Must(ParseYaml(r))
}

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
