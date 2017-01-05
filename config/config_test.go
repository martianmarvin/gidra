package config

import (
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCfg = `
version: '1'
config:
  verbosity: 4
  threads: 100
  loop: 1
  task_timeout: 15s
inputs:
  - source: ./test/in.csv
    type: csv
  - source: ./test/users.txt
    type: txt
    as: users
before:
  - get:
      url: http://icanhazip.com
  - extract:
      regex: '.*'
      as: ip
tasks:
  - get:
      url: 'http://www.google.com'
      headers:
        user-agent: gidra
        connection: keep-alive
    success:
      when: '{{ eq .Page.Status 200 }}'
  - extract:
      regex: 'csrf_token="(.*)"'
      as: csrf_token
`

func TestConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	f := strings.NewReader(testCfg)
	cfg, err := ParseYaml(f)
	require.NoError(err)

	assert.Equal("1", cfg.GetString("version"))

	subcfg, ok := cfg.CheckGet("config")
	assert.True(ok)
	duration, err := subcfg.GetDurationE("task_timeout")
	assert.NoError(err)
	assert.Equal(15*time.Second, duration)

	subcfgs, err := cfg.GetConfigSliceE("before")
	assert.NoError(err)
	assert.Len(subcfgs, 2)
	spew.Dump(subcfgs[0])
	u, err := subcfgs[0].GetURLE("get.url")
	assert.NoError(err)
	require.NotNil(u)
	assert.Equal("http://icanhazip.com", u.String())
	assert.Equal("ip", subcfgs[1].GetString("extract.as"))

	line, n := cfg.findLine("config.task_timeout")
	t.Log(line, n)
}
