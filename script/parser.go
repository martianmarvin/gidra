package script

import (
	"fmt"
	"io"
	"path/filepath"

	gidraconfig "github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
	"github.com/olebedev/config"
)

// Reads a config yaml file
func parseConfig(siteName string) (cfg *config.Config, err error) {
	fn := siteName + ".yaml"
	fp := filepath.Join(gidraconfig.ScriptDir, fn)

	cfg, err = config.ParseYamlFile(fp)

	return cfg, err
}

// Reads a Sequence from a config object
func parseSequence(key string, cfg *config.Config) (seq *Sequence, err error) {
	tasksList, err := cfg.List(key)
	if err != nil {
		return nil, err
	}
	seq = NewSequence()

	for i, _ := range tasksList {
		taskPath := fmt.Sprintf("%s.%d", key, i)
		tsk, err := parseTask(taskPath, cfg)
		if err != nil {
			return nil, err
		}
		seq.Add(tsk)
	}

	return seq, err
}

// Reads an individual Task from a yaml config. Parses standard variables such
// as Conditions, but leaves parsing of task-specific parameters to that Task's
// package file
func parseTask(key string, cfg *config.Config) (task.Task, error) {
	var err error
	var taskName string

	taskConfig, err := cfg.Map(key)
	if err != nil {
		return nil, err
	}

	// First map key is the task name
	for k, _ := range taskConfig {
		taskName = k
		break
	}

	// Task configuration variables
	taskVars, err := cfg.Map(fmt.Sprintf("%s.%s", key, taskName))
	if err != nil {
		return nil, err
	}

	//Panics if task not found
	tsk := task.New(taskName)

	err = task.Configure(tsk, taskVars)

	return tsk, err
}

// Parses any number of maps into the provided Vars instance
func parseMapVars(dst *vars.Vars, src ...map[string]interface{}) *vars.Vars {
	for _, v := range src {
		dst.Extend(vars.NewFromMap(v))
	}
	return dst
}

func parseInputVars(r io.Reader) {
}
