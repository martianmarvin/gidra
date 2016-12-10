package task

import (
	"errors"
	"strings"

	"github.com/fatih/structs"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/vars"
	"github.com/mitchellh/mapstructure"
)

const (
	// Tag on config struct
	configTag = "task"
	//Separator between tag fields
	tagSeparator = ","
)

type configurable struct {
	decoder *mapstructure.Decoder

	// Names of config fields tagged as required
	required []string

	// Metadata from decoding
	md *mapstructure.Metadata
}

// Configurable represents a task that can build its own custom Config struct
// from a map[string]interface{}
type Configurable interface {
	// Configure decodes input into the task's config struct's fields
	Configure(taskVars *vars.Vars) error

	// Required returns the keys of required config fields
	Required() []string
}

func NewConfigurable(configStruct interface{}) Configurable {
	var err error
	cfg := structs.New(configStruct)
	if !structs.IsStruct(cfg) {
		panic("Config must be a struct")
	}

	c := &configurable{required: make([]string, 0), md: &mapstructure.Metadata{}}

	for _, f := range cfg.Fields() {
		k := strings.ToLower(f.Name())
		tag := f.Tag(configTag)
		name, flags := parseFieldTag(tag)
		if len(name) == 0 {
			name = k
		}
		if flags.IsSet(config.FieldRequired) {
			c.required = append(c.required, name)
		}
	}

	c.decoder, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: configTag, Result: configStruct, Metadata: c.md})
	if err != nil {
		panic("Could not initialize config")
	}

	return c
}

func (c *configurable) Required() []string {
	return c.required
}

func (c *configurable) Configure(taskVars *vars.Vars) error {
	var err error
	// Check for presence of all required fields
	if err = taskVars.Require(c.required...); err != nil {
		return err
	}

	err = c.decoder.Decode(taskVars.Map())
	if err != nil {
		return err
	}
	// Validate all fields tagged required decoded successfully
	decoded := make(map[string]bool)
	for _, k := range c.md.Keys {
		decoded[k] = true
	}
	for _, k := range c.required {
		if !decoded[k] {
			return errors.New("Could not decode field " + k)
		}
	}
	return nil
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
