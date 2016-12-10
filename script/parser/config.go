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
	// Set by default config yaml in config.go, so they always exist
	s.Loop, err = cfg.Int(cfgConfigLoop)
	if err != nil {
		return err
	}

	s.Threads, err = cfg.Int(cfgConfigThreads)
	if err != nil {
		return err
	}

	s.Verbosity, err = cfg.Int(cfgConfigVerbosity)
	if err != nil {
		return err
	}

	s.TaskTimeout, err = cfg.Duration(cfgConfigTaskTimeout)
	if err != nil {
		return err
	}

	return nil
}
