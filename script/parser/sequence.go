package parser

import (
	"context"

	"github.com/martianmarvin/gidra/condition"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/martianmarvin/gidra/sequence"
	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/gidra/template"
	"github.com/martianmarvin/vars"
)

func init() {
	Register(cfgSeqBefore, beforeSeqParser)
	Register(cfgSeqTasks, mainSeqParser)
	Register(cfgSeqAfter, afterSeqParser)
}

func beforeSeqParser(s *options.ScriptOptions, cfg *config.Config) error {
	if s.BeforeSequence != nil {
		return KeyError{cfgSeqBefore}
	}
	taskList, err := cfg.CList("")
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
		return KeyError{cfgSeqTasks}
	}
	taskList, err := cfg.CList("")
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
	if s.AfterSequence != nil {
		return KeyError{cfgSeqAfter}
	}
	taskList, err := cfg.CList("")
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
		var taskVars *vars.Vars
		m, err := taskcfg.Map("")
		if err != nil {
			return nil, err
		}
		// task config should have one key, which is the task name
		for k, _ := range m {
			taskName = k
			break
		}

		tm, err := taskcfg.Map(taskName)
		if err != nil {
			return nil, err
		}

		taskVars = vars.NewFromMap(tm)

		// Parse all conditions like 'success', 'fail', etc for this task
		for k, _ := range tm {
			cond, err := parseCondition(k, taskcfg)
			if err != nil {
				if err, ok := err.(KeyError); ok {
					// KeyError simply means this item is not a
					// condition
					continue
				} else {
					return nil, err
				}
			}
			taskConds = append(taskConds, cond)
		}

		tsk := task.New(taskName)
		seq.Add(tsk, taskConds, taskVars)
	}
	return seq, nil
}

func parseCondition(key string, cfg *config.Config) (condition.Condition, error) {
	var cond condition.Condition

	if k, ok := cfgAliases[key]; ok {
		key = k
	}

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
		limit := cfg.UInt(key+"."+cfgTaskLimit, 1)
		cond = condition.NewRetry(limit, callbacks...)
	case cfgTaskFailCond:
		cond = condition.NewFail(callbacks...)
	default:
		return nil, KeyError{key}
	}

	ck := key + "." + cfgTaskCond
	tmpl, err := cfg.String(ck)
	if err != nil {
		return nil, KeyError{ck}
	}

	err = cond.Parse(tmpl)

	return cond, err
}

// Builds a callback function that executes the template provided at the
// specified key
// TODO support multiple callbacks supplied with a list?
func parseCallbacks(key string, cfg *config.Config) []condition.CallBackFunc {
	callbacks := make([]condition.CallBackFunc, 0)
	tmplText, err := cfg.String(key)
	if err != nil {
		return callbacks
	}

	tmplText = template.Format(tmplText, template.FmtAll)

	fn := func(ctx context.Context) error {
		tmpl, err := template.New(tmplText)
		if err != nil {
			return err
		}
		cvars := vars.FromContext(ctx)
		_, err = template.Execute(tmpl, cvars)
		return err
	}

	callbacks = append(callbacks, fn)

	return callbacks
}
