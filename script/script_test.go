package script

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	_ "github.com/martianmarvin/gidra/task/all"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testScripts = []string{
	`
config:
  loop: 5
tasks:
  - sleep:
      duration: 2s
`,
	`
default:
  http:
    headers:
      user-agent: gidra-test
      h1: v1
vars:
  host: https://www.example.com
before:
  - get:
      url: http://www.example.com/
  - sleep:
      duration: 2s
tasks:
  - get:
      url: http://www.example.com/get
      headers:
        h1: v2
  - post:
      url: http://www.example.com/post
`,
}

func TestLoad(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	for i, scr := range testScripts[1:] {
		msg := fmt.Sprintf("test script #%d \n%20s", i, scr)
		var buf bytes.Buffer

		r := strings.NewReader(scr)
		s, err := Open(r)
		require.NoError(err, msg)

		s.DryRun(&buf)
		output := string(buf.Bytes())
		assert.NotEmpty(output, msg)
		t.Log(output)
	}
}
