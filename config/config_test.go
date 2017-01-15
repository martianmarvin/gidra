package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCfg = `version: '1'
config:
  verbosity: 4
  threads: 100
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

func TestParse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	f := strings.NewReader(testCfg)
	cfg, err := ParseYaml(f)
	require.NoError(err)

	cfg = Default.Extend(cfg)
	require.NoError(err)

	assert.Equal("1", cfg.GetString("version"))

	subcfg, ok := cfg.CheckGet("config")
	assert.True(ok)
	duration, err := subcfg.GetDurationE("task_timeout")
	assert.NoError(err)
	assert.Equal(15*time.Second, duration)
	assert.Equal(subcfg.GetInt("loop"), 1)

	subcfgs, err := cfg.GetConfigSliceE("before")
	assert.NoError(err)
	assert.Len(subcfgs, 2)
	u, err := subcfgs[0].GetURLE("get.url")
	assert.NoError(err)
	require.NotNil(u)
	assert.Equal("http://icanhazip.com", u.String())
	assert.Equal("ip", subcfgs[1].GetString("extract.as"))
}

func TestError(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	f := strings.NewReader(testCfg)
	cfg, err := ParseYaml(f)
	require.NoError(err)

	cfg = cfg.Extend(Default)
	require.NoError(err)

	line, n := cfg.findLine("config.task_timeout")
	assert.Equal(5, n)
	assert.Contains(line, "task_timeout")
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)
	m1 := map[string]interface{}{"k1": "v1", "k2": "v2",
		"km1": map[string]interface{}{
			"k1": "v1",
			"k2": "v2",
		}}
	m2 := map[string]interface{}{"k1": "v1a", "k3": "v3",
		"km1": map[string]interface{}{
			"k1": "v1a",
			"k3": "v3a",
		}}
	merged := mergeMaps(m1, m2)
	assert.NotEmpty(merged)
	assert.Equal(merged["k1"], "v1a")
	assert.Equal(merged["k3"], "v3")
	sm := merged["km1"].(map[string]interface{})
	assert.Equal(sm["k1"], "v1a")
	assert.Equal(sm["k2"], "v2")
	assert.Equal(sm["k3"], "v3a")
	cfg := FromMap(merged)
	assert.Equal(len(merged), len(cfg.AllSettings()))
	assert.Contains(cfg.AllKeys(), "km1.k1")
}
