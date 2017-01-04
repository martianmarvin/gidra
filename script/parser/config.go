package parser

import (
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/script/options"
)

func init() {
	Register(cfgConfig, configParser)
}

func configParser(s *options.ScriptOptions, cfg *config.Config) error {
	var err error
	cfg = cfg.Get(cfgConfig, nil)
	// Set by default config yaml in config.go, so they always exist
	s.Loop, err = cfg.GetIntE(cfgConfigLoop)
	if err != nil {
		return err
	}

	s.Threads, err = cfg.GetIntE(cfgConfigThreads)
	if err != nil {
		return err
	}

	s.Verbosity, err = cfg.GetIntE(cfgConfigVerbosity)
	if err != nil {
		return err
	}

	s.TaskTimeout, err = cfg.GetDurationE(cfgConfigTaskTimeout)
	if err != nil {
		return err
	}

	return nil
}
