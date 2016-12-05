// Package config includes global configuration variables and defaults
// Config should not import any other subpackages to avoid circular imports
package config

import (
	"context"
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
loop: 10
`

// The default top level config object
var cfg *config.Config

// formats yaml to make it easier to parse
func parseYaml(text string) (*config.Config, error) {
	formatted := strings.Replace(strings.TrimSpace(text), "\t", "    ", -1)
	return config.ParseYaml(formatted)
}

func init() {
	cfg = config.Must(parseYaml(defaultConfig))
}

// Global settings not used inside tasks, therefore not subject to config file
var (
	// Location of script files
	ScriptDir = "./scripts"
)

// Default returns the default config
func Default() *config.Config {
	return cfg
}

// ToContext returns a context with the provided config merged into the one in
// the context
func ToContext(ctx context.Context, c *config.Config) context.Context {
	old := FromContext(ctx)
	merged := config.Must(old.Extend(c))
	return context.WithValue(ctx, ctxConfig, merged)
}

// FromContext returns a config from the context, or the default if the context
// does not have one
func FromContext(ctx context.Context) *config.Config {
	if c, ok := ctx.Value(ctxConfig).(*config.Config); ok {
		return c
	} else {
		return Default()
	}
}
