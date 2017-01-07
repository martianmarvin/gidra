// Package config includes global configuration variables and defaults
// Config should not import any other subpackages to avoid circular imports
package config

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type contextKey int

const (
	ctxConfig contextKey = iota
)

// The default top level config object
var Default *Config

func init() {
	viper.SetConfigType("yaml")
	Default = New()
	err := Default.ReadConfig(strings.NewReader(defaultConfig))
	if err != nil {
		panic(err.Error())
	}
}

// Config wraps the config.Config struct with additional helper methods
type Config struct {
	*viper.Viper

	// Original config text
	text []byte

	// Path from the master config, if this is a sub config
	path string

	// Last error message
	Error error
}

// New initializes a new Config
func New() *Config {
	return &Config{
		Viper: viper.New(),
	}
}

// Extend merges the new config with this one
func (cfg *Config) Extend(newcfg *Config) (*Config, error) {
	cf := newcfg.String()
	if len(cf) == 0 {
		return cfg, ErrParse
	}

	r := strings.NewReader(cf)
	err := cfg.Viper.MergeConfig(r)
	return cfg, err
}

func (cfg *Config) get(path string) *Config {
	subcfg := New()
	subcfg.path = cfg.path + "." + path
	subcfg.text = cfg.text
	c := cfg.Viper.Sub(path)
	if c == nil {
		subcfg.Error = cfg.keyError(path)
		return subcfg
	}
	subcfg.Viper = c
	return subcfg
}

// FIXME doens't work
// Find line number of the specific path and the line
func (cfg *Config) findLine(path string) (string, int) {
	var line string
	var n int

	s := cfg.text
	if len(cfg.text) == 0 {
		return "", 0
	}
	parts := strings.Split(path, ".")

	i := 0
	scanner := bufio.NewScanner(bytes.NewReader(s))
	for scanner.Scan() {
		n++
		re := regexp.MustCompile(parts[i] + `:`)
		line = scanner.Text()
		if re.MatchString(line) {
			i += 1
			if i >= len(parts) {
				break
			}
		}
	}

	return line, n
}

// Format a keyerror on path
func (cfg *Config) keyError(path string) KeyError {
	line, n := cfg.findLine(path)
	return KeyError{Name: path, Line: fmt.Sprintf("%d: %s", n, line)}
}

// Get the underlying value or return an error
func (cfg *Config) getValue(path string) (interface{}, error) {
	if cfg.Viper == nil || len(path) == 0 {
		return nil, KeyError{Name: path}
	}
	v := cfg.Viper.Get(path)
	if v == nil {
		return nil, KeyError{Name: path}
	}
	return v, nil
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
	return cfg.get(path), cfg.IsSet(path)
}

// GetInterface returns the value as an interface{} without casting
func (cfg *Config) GetInterfaceE(path string) (interface{}, error) {
	return cfg.getValue(path)
}

func (cfg *Config) GetInterface(path string) interface{} {
	v, err := cfg.getValue(path)
	if err != nil {
		return nil
	}
	return v
}

func (cfg *Config) GetStringE(path string) (string, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return "", err
	}
	val, err := cast.ToStringE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetBoolE(path string) (bool, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return false, err
	}
	val, err := cast.ToBoolE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetIntE(path string) (int, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return 0, err
	}
	val, err := cast.ToIntE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetInt64E(path string) (int64, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return 0, err
	}
	val, err := cast.ToInt64E(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetFloat64E(path string) (float64, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return 0.0, err
	}
	val, err := cast.ToFloat64E(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetTimeE(path string) (time.Time, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return time.Time{}, err
	}
	val, err := cast.ToTimeE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetDurationE(path string) (time.Duration, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return time.Duration(0), err
	}
	val, err := cast.ToDurationE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetSliceE(path string) ([]interface{}, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return nil, err
	}
	val, err := cast.ToSliceE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetSlice(path string) []interface{} {
	l, err := cfg.GetSliceE(path)
	if err != nil {
		return make([]interface{}, 0)
	}
	return l
}

func (cfg *Config) GetStringSliceE(path string) ([]string, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return nil, err
	}
	val, err := cast.ToStringSliceE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetMapE(path string) (map[string]interface{}, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return nil, err
	}
	val, err := cast.ToStringMapE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetMap(path string) map[string]interface{} {
	m, err := cfg.GetMapE(path)
	if err != nil {
		return make(map[string]interface{})
	}
	return m
}

func (cfg *Config) GetStringMapE(path string) (map[string]string, error) {
	v, err := cfg.getValue(path)
	if err != nil {
		return nil, err
	}
	val, err := cast.ToStringMapStringE(v)
	return val, NewValueError(path, err)
}

func (cfg *Config) GetStringMap(path string) map[string]string {
	m, err := cfg.GetStringMapE(path)
	if err != nil {
		return make(map[string]string)
	}
	return m
}

func (cfg *Config) GetMapSliceE(path string) ([]map[string]interface{}, error) {
	val, err := cfg.GetSliceE(path)
	if err != nil {
		return nil, err
	}
	l := make([]map[string]interface{}, len(val))
	for i, v := range val {
		m, err := cast.ToStringMapE(v)
		if err != nil {
			return nil, NewValueError(fmt.Sprintf("%s.%d", path, i), err)
		}
		l[i] = m
	}
	return l, nil
}

func (cfg *Config) GetMapSlice(path string) []map[string]interface{} {
	l, err := cfg.GetMapSliceE(path)
	if err != nil {
		return make([]map[string]interface{}, 0)
	}
	return l
}

// GetConfigSliceE returns a slice of Config objects that can be accessed with the
// other methods
func (cfg *Config) GetConfigSliceE(path string) ([]*Config, error) {
	list := make([]*Config, 0)
	l, err := cfg.GetMapSliceE(path)
	if err != nil {
		e := cfg.keyError(path)
		e.Err = err
		return nil, e
	}
	for _, cm := range l {
		subcfg := New()
		subcfg.text = cfg.text
		subcfg.path = cfg.path + "." + path
		for k, v := range cm {
			subcfg.Set(k, v)
		}
		list = append(list, subcfg)
	}
	return list, nil
}

func (cfg *Config) GetConfigSlice(path string) []*Config {
	l, err := cfg.GetConfigSliceE(path)
	if err != nil {
		return make([]*Config, 0)
	}
	return l
}

// GetConfigMapE returns a map of Config objects that can be accessed with the
// other methods
func (cfg *Config) GetConfigMapE(path string) (map[string]*Config, error) {
	cm := make(map[string]*Config)
	m, err := cfg.GetMapE(path)
	if err != nil {
		e := cfg.keyError(path)
		e.Err = err
		return nil, e
	}

	for k, _ := range m {
		if c, ok := cfg.CheckGet(fmt.Sprintf("%s.%s", path, k)); ok {
			cm[k] = c
		}
	}
	return cm, nil
}

func (cfg *Config) GetConfigMap(path string) map[string]*Config {
	m, err := cfg.GetConfigMapE(path)
	if err != nil {
		return make(map[string]*Config)
	}
	return m
}

func (cfg *Config) GetURLE(path string) (*url.URL, error) {
	rawurl, err := cfg.GetStringE(path)
	if err != nil {
		return nil, err
	}
	return url.Parse(rawurl)
}

func (cfg *Config) GetURL(path string) *url.URL {
	u, err := cfg.GetURLE(path)
	if err != nil {
		return nil
	}
	return u
}

// ParseYAML creates a new config from a yaml file
func ParseYaml(r io.Reader) (*Config, error) {
	// Replace tabs with spaces
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b = bytes.Replace(b, []byte("\t"), []byte("  "), -1)
	cfg := New()
	cfg.Viper.SetConfigType("yaml")
	cfg.text = b
	err = cfg.Viper.ReadConfig(bytes.NewReader(b))
	return cfg, err
}

// Must panics if there is an error
func Must(cfg *Config, err error) *Config {
	if err != nil {
		panic("config: " + err.Error())
	}
	return cfg
}

// ToContext returns a context with the provided config merged into the one in
// the context
func ToContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ctxConfig, cfg)
}

// FromContext returns a config from the context, or an empty config if the context
// does not have one
func FromContext(ctx context.Context) *Config {
	if c, ok := ctx.Value(ctxConfig).(*Config); ok {
		return c
	} else {
		return New()
	}
}

// Map returns the config as a map[string]interface
func (cfg *Config) Map() map[string]interface{} {
	return cfg.AllSettings()
}

// StringMap returns the config as a map with string values
func (cfg *Config) StringMap() map[string]string {
	m := make(map[string]string)
	for k, v := range cfg.Map() {
		m[k] = fmt.Sprint(v)
	}
	return m
}

// String implements the Stringer interface
func (cfg *Config) String() string {
	b, err := yaml.Marshal(cfg.AllSettings())
	if err != nil {
		return ""
	}
	return string(b)
}
