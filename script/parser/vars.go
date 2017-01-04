package parser

import (
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/martianmarvin/vars"
)

func init() {
	Register(cfgVars, varsParser)
}

func varsParser(s *options.ScriptOptions, cfg *config.Config) error {
	taskVars, err := cfg.GetMapE(cfgVars)
	if err != nil {
		s.Vars = vars.New()
	} else {
		s.Vars = vars.NewFromMap(taskVars)
	}

	return nil
}
