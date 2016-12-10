package template

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

var (
	goTemplateRegex = regexp.MustCompile(`.*{{.+}}.*`)
	singleVarRegex  = regexp.MustCompile(`\$(\w+)`)
)

// Sets of formatter functions
var (
	FmtAll []FormatFunc = []FormatFunc{formatTemplate, formatUserVar}
)

// FormatFunc transforms a text template before it is compiled
type FormatFunc func(string) string

func formatTemplate(tmpl string) string {
	if validTmpl(tmpl) {
		return tmpl
	}
	tmpl = strings.TrimSpace(tmpl)
	if !strings.Contains(tmpl, "{{") {
		tmpl = "{{ " + tmpl
	}
	if !strings.Contains(tmpl, "}}") {
		tmpl = tmpl + " }}"
	}

	// TODO Prepend dot if not exists?

	return tmpl
}

// Helper to transform '$var' to '{{ $.Vars.var }}
func formatUserVar(tmpl string) string {
	replace := `$.Vars.$1`
	if validTmpl(tmpl) {
		return tmpl
	}
	tmpl = singleVarRegex.ReplaceAllString(tmpl, replace)
	return tmpl
}

// Basic sanity check that this is already a valid Go template
func validTmpl(tmpl string) bool {
	return goTemplateRegex.MatchString(tmpl)
}

// Format formats the template by running it through the specified list of
// formatters to get it to valid Go template format
func Format(tmpl string, formatters []FormatFunc) string {
	for _, formatter := range formatters {
		// Don't touch if the template is already valid
		if validTmpl(tmpl) {
			return tmpl
		}
		tmpl = formatter(tmpl)
	}
	return tmpl
}

// New creates a new template from the given data and global
// functions, and returns the result as a compiled template
// TODO mutex and cache compiled templates?
func New(tmpl string) (*template.Template, error) {
	return template.New("").Option("missingkey=zero").Funcs(funcMap).Parse(tmpl)
}

// Execute executes the provided template and returns the result
func Execute(t *template.Template, data interface{}) (string, error) {
	var b bytes.Buffer
	err := t.Execute(&b, data)
	if err != nil {
		return "", err
	}
	return b.String(), err
}
