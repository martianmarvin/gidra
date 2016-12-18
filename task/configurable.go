package task

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fatih/structs"
	"github.com/imdario/mergo"
	"github.com/martianmarvin/gidra"
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

	// map of task vars to use as options with task config struct
	opts := make(map[string]interface{})

	// Validate all fields tagged required decoded successfully
	configStruct := structs.New(c.config)
	for _, f := range configStruct.Fields() {
		if !f.IsExported() {
			continue
		}
		k := changeInitialCase(f.Name(), unicode.ToLower)
		tag := f.Tag(configTag)
		name, flags := parseFieldTag(tag)
		if len(name) == 0 {
			name = k
		}
		// Fix for Mergo zero value issue
		if !f.IsZero() {
			opts[k] = f.Value()
			f.Zero()
		}

		cf, ok := cfg.CheckGet(name)
		if !ok {
			if flags.IsSet(config.FieldRequired) {
				return gidra.FieldError{name}
			} else {
				continue
			}
		}

		// Special case to correctly parse stdlib custom types
		opts[k], err = parseType(cf, f)
		if err != nil {
			return gidra.ValueError{name, err}
		}
	}

	err = mergo.MapWithOverwrite(c.config, opts)
	if err != nil {
		return err
	}
	log.Printf("%v\n", c.config)

	return nil
}

// Parse custom stdlib type supported by config
func parseType(cf *config.Config, f *structs.Field) (interface{}, error) {
	switch f.Value().(type) {
	case time.Duration:
		return cf.Duration("")
	case *url.URL:
		return cf.URL("")
	case int:
		return cf.Int("")
	case int32:
		i, err := cf.Int("")
		return int32(i), err
	case int64:
		i, err := cf.Int("")
		return int64(i), err
	case float64:
		return cf.Float64("")
	case bool:
		return cf.Bool("")
	case string:
		return cf.String("")
	case []byte:
		s, err := cf.String("")
		return []byte(s), err
	case []string:
		return cf.StringList("")
	case []map[string]interface{}:
		return cf.MapList("")
	case []interface{}:
		return cf.List("")
	case map[string]string:
		return cf.StringMap("")
	case map[string]interface{}:
		return cf.Map("")
	default:
		return cf.Root, nil
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
