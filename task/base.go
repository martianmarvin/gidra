package task

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

// BaseTask includes fields common to all tasks
type BaseTask struct {
	//Id is the number of this task in the sequence
	Id int
}

// Try to parse interface to int
func parseInt(v interface{}) int {
	switch i := v.(type) {
	case int:
		return i
	case int32:
		return int(i)
	case int64:
		return int(i)
	case string:
		if j, err := strconv.Atoi(i); err == nil {
			return int(j)
		} else {
			return 0
		}
	default:
		return 0
	}
}

// Try to parse map interface values to string
func parseStringMap(vars map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range vars {
		res[k] = fmt.Sprint(v)
	}
	return res
}

//checks all required params are present in the config struct
func validateConfig(config interface{}, vars map[string]interface{}) (err error) {
	cfg := structs.New(config)
	if !structs.IsStruct(cfg) {
		return
	}
	for _, f := range cfg.Fields() {
		fn := strings.ToLower(f.Name())
		tag := f.Tag(configTag)
		name, flags := parseFieldTag(tag)
		if len(name) == 0 {
			name = fn
		}
		if flags.IsSet(FieldRequired) {
			if _, ok := vars[name]; !ok {
				return errors.New("Required parameter missing: " + name)
			}
		}
	}
	return err
}

//Configure parses input vars into this task's Config and Vars
//TODO refactor to use type assertions instead of reflection
func Configure(t Task, vars map[string]interface{}) (err error) {
	tsk := structs.New(t)

	//Parse id if it is provided
	if id, ok := vars["id"]; ok {
		if f, ok := tsk.FieldOk("Id"); ok && f.IsExported() {
			if err = f.Set(parseInt(id)); err != nil {
				return
			}
		}
	}

	// Initialize Vars map if exists
	if f, ok := tsk.FieldOk("Vars"); ok && f.IsExported() {
		if f.IsZero() {
			if err = f.Set(make(map[string]string)); err != nil {
				return
			}
		}

		fvars := f.Value().(map[string]string)

		if len(vars) > 0 {
			strVars := parseStringMap(vars)
			for k, v := range strVars {
				fvars[k] = v
			}
		}
	}

	if f, ok := tsk.FieldOk("Config"); ok && f.IsExported() && !f.IsZero() {
		cfg := f.Value()
		//Validate required parameters
		if err = validateConfig(f.Value(), vars); err != nil {
			return
		}
		fmt.Println("valid")
		decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: configTag, Result: cfg})
		if err = decoder.Decode(vars); err != nil {
			return
		}
	}
	return err
}
