package parser

import (
	"os"
	"path"
	"testing"

	// Register datasource and task types
	_ "github.com/martianmarvin/gidra/datasource/all"
	_ "github.com/martianmarvin/gidra/task/all"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/martianmarvin/gidra/sequence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The test script has all possible options
var testScript = "./scripts/template.yaml"

func TestParser(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Change to base dir for tests
	baseDir := path.Join(os.Getenv("GOPATH"), "src/github.com/martianmarvin/gidra/")
	os.Chdir(baseDir)

	f, err := os.Open(testScript)
	require.NoError(err)
	cfg, err := config.ParseYaml(f)
	require.NoError(err)
	opts := options.New()

	err = Configure(opts, cfg)
	assert.NoError(err)

	// Config
	subcfg, ok := cfg.CheckGet(cfgConfig)
	require.True(ok)
	// Not equal because loop should be limited by test file size
	assert.NotEqual(opts.Input["main"].Len(), subcfg.GetInt(cfgConfigLoop), "config.loop")
	assert.Equal(opts.Threads, subcfg.GetInt(cfgConfigThreads), "config.threads")
	assert.Equal(opts.TaskTimeout, subcfg.GetDuration(cfgConfigTaskTimeout), "config.task_timeout")
	assert.Equal(opts.Verbosity, subcfg.GetInt(cfgConfigVerbosity), "config.verbosity")

	// Global Vars
	gvars, err := cfg.GetStringMapE(cfgVars)
	assert.NoError(err, "vars")
	assert.Equal(len(gvars), len(opts.Vars.Map()))

	//Sequence Templates
	subcfgs := make([][]*config.Config, 3)
	subcfgs[0] = cfg.GetConfigSlice(cfgSeqBefore)
	subcfgs[1] = cfg.GetConfigSlice(cfgSeqTasks)
	subcfgs[2] = cfg.GetConfigSlice(cfgSeqAfter)

	//TODO fix sequence parsing
	sequences := make([]*sequence.Sequence, 3)
	sequences[0] = opts.BeforeSequence
	sequences[1] = opts.MainSequence
	sequences[2] = opts.AfterSequence

	for i, seq := range sequences {
		assert.NotNil(seq)
		n := len(subcfgs[i])
		assert.Len(seq.Tasks, n)
	}

	// spew.Dump(opts)

}
