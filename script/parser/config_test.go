package parser

import "testing"

var testConfigCfg = `config:
  verbosity: 3
  threads: 200
  task_timeout: 15s
  loop: 20
`

var configTests = []parserTest{
	{testConfigCfg, `{{.Verbosity}}`, "3"},
	{testConfigCfg, `{{.Threads}}`, "200"},
	{testConfigCfg, `{{.TaskTimeout}}`, "15s"},
	{testConfigCfg, `{{.Loop}}`, "20"},
}

func TestConfig(t *testing.T) {
	for _, v := range configTests {
		testParser(t, configParser, v)
	}

}
