package template

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/martianmarvin/gidra/datasource"
)

func init() {
	initFuncMap()
}

type regexCache map[string]*regexp.Regexp

func (c regexCache) Get(pattern string) (*regexp.Regexp, error) {
	re, ok := c[pattern]
	if ok {
		return re, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	c[pattern] = re
	return re, nil
}

var (
	funcMap template.FuncMap

	regexMap regexCache = make(map[string]*regexp.Regexp)
)

// Custom template functions
// Some functions silently drop errors so template continues executing
func initFuncMap() {
	funcMap = template.FuncMap{
		// Compare string equality of two interfaces
		"eq": strEq,
		// Get environment variable
		"env": getEnv,
		// Advance a reader to the next position
		"next": nextRow,
		// Return a random element of the list
		"pick": pickRand,
		// Match a regex
		"match": matchRegex,
		// Get the output of running a shell command
		"shell": runShell,
	}

}

func getEnv(k string) string {
	return os.Getenv(k)
}

func nextRow(r datasource.ReadableTable) *datasource.Row {
	row, _ := r.Next()
	return row
}

// TODO: Fix InterfaceSlice
func pickRand(list []interface{}) interface{} {
	return list[rand.Intn(len(list))]
}

func matchRegex(text, pattern string) (bool, error) {
	re, err := regexMap.Get(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(text), nil
}

func strEq(a, b interface{}) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func runShell(args ...string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("Command is required as first argument")
	}
	cmd := args[0]
	args = args[1:]
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
