package template

import (
	"os"
	"testing"

	"github.com/martianmarvin/gidra/global"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testVars = map[string]interface{}{
	"host": "www.example.com",
	"Name": "John Smith",
	"ids":  []int{1, 2, 3, 4, 5},
	"colors": map[string]string{
		"black": "#000000",
		"white": "#ffffff",
	},
	"nested": "{{ .Vars.host }}",
}

func testTmpl(t *testing.T, rawtmpl string) string {
	g := global.New()
	g.Vars = testVars
	tmpl, err := New(rawtmpl)
	require.NoError(t, err, rawtmpl)
	text, err := Execute(tmpl, g)
	assert.NoError(t, err, rawtmpl)
	return text
}

func TestTemplate(t *testing.T) {
	// Map template : expected
	var tmpls = map[string]interface{}{
		`{{ .Vars.host }}`:         testVars["host"],
		`{{ .Vars.colors.black }}`: testVars["colors"].(map[string]string)["black"],
		`{{ .Vars.nested }}`:       testVars["host"],
	}
	for tmpl, expected := range tmpls {
		text := testTmpl(t, tmpl)
		assert.EqualValues(t, expected, text)
	}
}

func TestFuncs(t *testing.T) {
	os.Setenv("ENV_KEY", "abc123")
	assert := assert.New(t)

	var tmpls = []string{
		`{{ eq .Vars.Name "John Smith" }}`,
		`{{ env "ENV_KEY" }}`,
		// `{{ pick .Vars.ids }}`, // Fix InterfaceSlice
		`{{ match .Vars.host "www.*com" }}`,
		`{{ shell "/bin/echo" "ok" }}`,
	}
	results := make([]string, len(tmpls))

	for i, tmpl := range tmpls {
		text := testTmpl(t, tmpl)
		results[i] = text
	}

	assert.NotEmpty(results[0])
	assert.Equal(os.Getenv("ENV_KEY"), results[1])
	// assert.NotZero(results[2])
	assert.NotEmpty(results[2])
	assert.Equal("ok\n", results[3])
}
