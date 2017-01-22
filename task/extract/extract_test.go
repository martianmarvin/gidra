package extract

import (
	"context"
	"strings"
	"testing"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testText = `
<div>
  <div class="t2">
    <p class="t3" id="contentTag">
	  content
	</p>
  </div>
</div>
`

func testExtract(t *testing.T, conf, expected string) {
	assert := assert.New(t)
	require := require.New(t)

	r := strings.NewReader(conf)
	cfg := config.Must(config.ParseYaml(r))
	cfg.Set("text", testText)
	tsk := New()

	ctx := config.ToContext(context.Background(), cfg)
	cTask := tsk.(task.Configurable)
	err := cTask.Configure(cfg)
	require.NoError(err)
	err = tsk.Execute(ctx)
	require.NoError(err)

	wTask := tsk.(task.Writeable)

	res, err := wTask.Row().Get("result").StringArray()
	assert.NoError(err)
	require.Len(res, 1)
	assert.Equal(expected, res[0])
}

func TestExtractElement(t *testing.T) {
	testCfgs := []string{
		`element: '.t2>.t3' 
as: result
`,
		`element: '.t2>.t3' 
as: result
trim: false
`,
		`element: '.t2>.t3' 
as: result
attr: id
`,
	}
	expected := []string{
		"content",
		"\n\t  content\n\t",
		"contentTag",
	}

	for i, testCfg := range testCfgs {
		testExtract(t, testCfg, expected[i])
	}
}

func TestExtractRegex(t *testing.T) {
	testCfgs := []string{
		`regex: '(?s)t3" id="(.*?)"' 
as: result
`,
	}
	expected := []string{
		"contentTag",
	}

	for i, testCfg := range testCfgs {
		testExtract(t, testCfg, expected[i])
	}
}
