package parser

import (
	"os"
	"path"
	"testing"
	"time"

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
	subcfg := cfg.Get(cfgConfig, nil)
	require.NotNil(subcfg)
	// Not equal because loop should be limited by test file size
	assert.NotEqual(opts.Input["main"].Len(), subcfg.UInt(cfgConfigLoop), "config.loop")
	assert.Equal(opts.Threads, subcfg.UInt(cfgConfigThreads), "config.threads")
	assert.Equal(opts.TaskTimeout, time.Duration(subcfg.UInt(cfgConfigTaskTimeout))*time.Second, "config.task_timeout")
	assert.Equal(opts.Verbosity, subcfg.UInt(cfgConfigVerbosity), "config.verbosity")

	// Global Vars
	gvars, err := cfg.StringMap(cfgVars)
	assert.NoError(err, "vars")
	assert.Equal(len(gvars), len(opts.Vars.Map()))

	//Sequence Templates
	subcfgs := make([][]*config.Config, 3)
	subcfgs[0], _ = cfg.CList(cfgSeqBefore)
	subcfgs[1], _ = cfg.CList(cfgSeqTasks)
	subcfgs[2], _ = cfg.CList(cfgSeqAfter)

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
