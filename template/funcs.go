package template

import (
	"container/list"
	"container/ring"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/davecgh/go-spew/spew"
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
		// Advance an iterable to the next position
		"next": next,
		// Return a random element of the list
		"pick": pickRand,
		// Match a regex
		"match": matchRegex,
		// Get the output of running a shell command
		"shell": runShell,
		// Suppress output of a command/ like piping to /dev/null
		"null": null,
		// Print to Stdout
		"echo": echo,
		// For debugging, dump the object
		"dump": sdump,
	}

}

func getEnv(k string) string {
	return os.Getenv(k)
}

func next(iter interface{}) interface{} {
	switch r := iter.(type) {
	case datasource.ReadableTable:
		res, _ := r.Next()
		return res
	case ring.Ring:
		return r.Next().Value
	case list.List:
		el := r.Front()
		if el == nil {
			return nil
		}
		return el.Next().Value
	default:
		return nil
	}
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

func null(args ...interface{}) string {
	return ""
}

func echo(args ...interface{}) string {
	fmt.Println(args)
	return ""
}

func sdump(args ...interface{}) string {
	return spew.Sdump(args)
}
