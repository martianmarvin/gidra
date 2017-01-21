package parser

import (
	"context"
	"errors"

	"github.com/martianmarvin/gidra/condition"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/global"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/martianmarvin/gidra/sequence"
	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/gidra/template"
)

func init() {
	Register(cfgSeqBefore, beforeSeqParser)
	Register(cfgSeqTasks, mainSeqParser)
	Register(cfgSeqAfter, afterSeqParser)
}

func beforeSeqParser(s *options.ScriptOptions, cfg *config.Config) error {
	// Only one before sequence is allowed
	if s.BeforeSequence != nil {
		return config.KeyError{Name: cfgSeqBefore, Err: errors.New("Only one before sequence is allowed")}
	}
	taskList, err := cfg.GetConfigSliceE(cfgSeqBefore)
	if err != nil {
		return err
	}

	seq, err := parseSequence(taskList)
	if err != nil {
		return err
	}
	s.BeforeSequence = seq

	return nil
}

func mainSeqParser(s *options.ScriptOptions, cfg *config.Config) error {
	if s.MainSequence != nil {
		return config.KeyError{Name: cfgSeqTasks, Err: config.ErrRequired}
	}
	taskList, err := cfg.GetConfigSliceE(cfgSeqTasks)
	if err != nil {
		return err
	}

	seq, err := parseSequence(taskList)
	if err != nil {
		return err
	}
	s.MainSequence = seq

	return nil
}

func afterSeqParser(s *options.ScriptOptions, cfg *config.Config) error {
	// Only one after sequence is allowed
	if s.AfterSequence != nil {
		return config.KeyError{Name: cfgSeqAfter, Err: errors.New("Only one finally sequence is allowed")}
	}
	taskList, err := cfg.GetConfigSliceE(cfgSeqAfter)
	if err != nil {
		return err
	}

	seq, err := parseSequence(taskList)
	if err != nil {
		return err
	}
	s.AfterSequence = seq

	return nil
}

// Parses config into a new sequence and initializes tasks
func parseSequence(taskList []*config.Config) (*sequence.Sequence, error) {
	seq := sequence.New()
	for _, taskcfg := range taskList {
		var taskName string
		var taskConds []condition.Condition

		m := taskcfg.Map()
		// task config should have one key, which is the task name
		for k, _ := range m {
			taskName = k
			break
		}
		taskcfg = taskcfg.Get(taskName, nil)

		// Parse all conditions like 'success', 'fail', etc for this task
		for k, _ := range taskcfg.Map() {
			if !isCondition(k) {
				continue
			}
			cond, err := parseCondition(k, taskcfg)
			if err != nil {
				return nil, err
			}
			taskConds = append(taskConds, cond)
		}

		tsk := task.New(taskName)
		seq.Add(tsk, taskConds, taskcfg)
	}
	return seq, nil
}

// Whether this key is one of the condition keys (should not be passed on to
// task
func isCondition(key string) bool {
	for _, k := range conditionKeys {
		if key == k || cfgAliases[key] == k {
			return true
		}
	}
	return false
}

func parseCondition(key string, cfg *config.Config) (condition.Condition, error) {
	var cond condition.Condition

	callbacks := parseCallbacks(key+"."+cfgTaskBefore, cfg)

	switch key {
	case cfgTaskCond:
		cond = condition.NewOnly()
	case cfgTaskSkipCond:
		cond = condition.NewSkip()
	case cfgTaskSuccessCond:
		cond = condition.NewSuccess()
	case cfgTaskAbortCond:
		cond = condition.NewAbort(callbacks...)
	case cfgTaskRetryCond:
		limit := cfg.GetInt(key + "." + cfgTaskLimit)
		if limit <= 0 {
			limit = 1
		}
		cond = condition.NewRetry(limit, callbacks...)
	case cfgTaskFailCond:
		cond = condition.NewFail(callbacks...)
	case cfgTaskBefore:
		// Just run task callbacks
		callbacks = parseCallbacks(key, cfg)
		cond = condition.NewTrue(callbacks...)
		return cond, nil
	default:
		return nil, config.KeyError{Name: key}
	}

	// Condition can be a string or submap
	tmpl, err := cfg.GetStringE(key)
	if err == nil {
		err = cond.Parse(tmpl)
		return cond, err
	}

	ck := key + "." + cfgTaskCond
	tmpl, err = cfg.GetStringE(ck)
	if err != nil {
		return nil, config.KeyError{Name: ck}
	}

	err = cond.Parse(tmpl)

	return cond, err
}

// Builds a callback function that executes the template provided at the
// specified key
// TODO support multiple callbacks supplied with a list?
func parseCallbacks(key string, cfg *config.Config) []condition.CallBackFunc {
	callbacks := make([]condition.CallBackFunc, 0)
	tmplText, err := cfg.GetStringE(key)
	if err != nil {
		return callbacks
	}

	tmplText = template.Format(tmplText, template.FmtAll)

	fn := func(ctx context.Context) error {
		tmpl, err := template.New(tmplText)
		if err != nil {
			return err
		}
		g := global.FromContext(ctx)
		_, err = tmpl.Execute(g)
		return err
	}

	callbacks = append(callbacks, fn)

	return callbacks
}
