package parser

import (
	"strings"
	"testing"

	// Register datasource and task types

	_ "github.com/martianmarvin/gidra/datasource/all"
	_ "github.com/martianmarvin/gidra/task/all"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/martianmarvin/gidra/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The test script has all possible options
var testScript = "./scripts/template.yaml"

// For transition to table driven tests
type parserTest struct {
	rawcfg string
	// template for getting the result from opts, ex '{{ $.Timeout }}'
	valTmpl  string
	expected string
}

// Generic tester for parser
func testParser(t *testing.T, parser ParseFunc, v parserTest) {
	assert := assert.New(t)
	require := require.New(t)

	r := strings.NewReader(v.rawcfg)
	cfg, err := config.ParseYaml(r)
	require.NoError(err)

	opts := options.New()
	err = parser(opts, cfg)
	require.NoError(err)

	tmpl, err := template.New(v.valTmpl)
	require.NoError(err)

	res, err := tmpl.Execute(opts)
	assert.NoError(err)

	res = strings.TrimSpace(res)

	assert.Equal(v.expected, res)
}
