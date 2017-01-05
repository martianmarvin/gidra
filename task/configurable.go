package task

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fatih/structs"
	"github.com/martianmarvin/gidra/config"
)

const (
	// Tag on config struct
	configTag = "task"
	//TODO Tag for help text on struct config
	helpTag = "help"
	//Separator between tag fields
	tagSeparator = ","
)

type configurable struct {
	// The struct this task's config is stored in
	config interface{}
}

// Configurable represents a task that can build its own custom Config struct
// from a map[string]interface{}
type Configurable interface {
	// Configure decodes input into the task's config struct's fields
	Configure(cfg *config.Config) error

	// String returns this task's config as  a string
	String() string
}

func NewConfigurable(configStruct interface{}) Configurable {
	return &configurable{config: configStruct}
}

func (c *configurable) String() string {
	var s string
	configStruct := structs.New(c.config)
	for k, v := range configStruct.Map() {
		s += fmt.Sprintf("%s=%v ", k, v)
	}
	return s
}

func (c *configurable) Configure(cfg *config.Config) error {
	var err error

	// Validate all fields tagged required decoded successfully
	configStruct := structs.New(c.config)
	for _, f := range configStruct.Fields() {
		if !f.IsExported() {
			continue
		}
		k := changeInitialCase(f.Name(), unicode.ToLower)
		tag := f.Tag(configTag)
		name, flags := parseFieldTag(tag)
		if len(name) > 0 {
			cfg.RegisterAlias(name, k)
		}

		ok := cfg.IsSet(k)
		if !ok {
			if flags.IsSet(config.FieldRequired) {
				return config.KeyError{Name: name, Err: config.ErrRequired}
			}
		}
	}

	// Try to unmarshal with config default
	cfg.Unmarshal(c.config)

	for _, f := range configStruct.Fields() {
		// If viper could not unmarshal, we take over
		if f.IsExported() && f.IsZero() {
			val := parseType(cfg, f)
			err = f.Set(val)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Parse custom stdlib type supported by config
func parseType(cf *config.Config, f *structs.Field) interface{} {
	key := strings.ToLower(f.Name())
	switch f.Value().(type) {
	case time.Duration:
		return cf.GetDuration(key)
	case *url.URL:
		return cf.GetURL(key)
	case int:
		return cf.GetInt(key)
	case int32:
		return int32(cf.GetInt(key))
	case int64:
		return cf.GetInt64(key)
	case float64:
		return cf.GetFloat64(key)
	case bool:
		return cf.GetBool(key)
	case string:
		return cf.GetString(key)
	case []byte:
		return []byte(cf.GetString(key))
	case []string:
		return cf.GetStringSlice(key)
	case []map[string]interface{}:
		return cf.GetMapSlice(key)
	case []interface{}:
		return cf.GetSlice(key)
	case map[string]string:
		return cf.GetStringMap(key)
	case map[string]interface{}:
		return cf.GetMap(key)
	default:
		return cf.GetInterface(key)
	}
}

func changeInitialCase(s string, mapper func(rune) rune) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(mapper(r)) + s[n:]
}

func parseFieldTag(tag string) (name string, flags config.Flag) {
	if len(tag) == 0 {
		return
	}

	fieldTags := strings.Split(tag, tagSeparator)
	for _, ft := range fieldTags {
		switch ft {
		case "-":
			flags.Set(config.FieldSkip)
		case "required":
			flags.Set(config.FieldRequired)
		case "omitempty":
			flags.Set(config.FieldOmitEmpty)
		default:
			if len(name) == 0 {
				name = ft
			}
		}
	}
	return name, flags
}
