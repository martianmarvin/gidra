package extract

import (
	"bytes"
	"errors"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/task"
)

var (
	//Cache of compiled matchers
	regexMatchers   = make(map[string]*regexp.Regexp)
	elementMatchers = make(map[string]cascadia.Selector)
)

func init() {
	task.Register("extract", NewTask)
}

type Task struct {
	task.BaseTask

	Config *Config
}

type Config struct {
	//Regex string to match to extract
	RegexSelector string `task:"regex"`

	//Jquery selector to match to extract
	ElementSelector string `task:"element"`

	//Key to save match in, if any
	Key string `task:"as"`
}

func NewTask() task.Task {
	return &Task{
		BaseTask: task.BaseTask{},
		Config:   &Config{},
	}
}

//Return the compiled matcher for this string key
func regexMatcher(pattern string) (matcher *regexp.Regexp, err error) {
	matcher, ok := regexMatchers[pattern]
	if ok {
		return matcher, err
	}
	matcher, err = regexp.Compile(pattern)
	return matcher, err
}

//Return the compiled matcher for this string key
func elementMatcher(pattern string) (matcher cascadia.Selector, err error) {
	matcher, ok := elementMatchers[pattern]
	if ok {
		return matcher, err
	}
	matcher, err = cascadia.Compile(pattern)
	return matcher, err
}

//Functions to extract fields from results
//TODO Ability to extract a list of values instead of a single one

func matchByRegex(matcher *regexp.Regexp, text []byte) bool {
	return matcher.Match(text)
}

//ExtractByRegex extracts the specified value from the page via a regex
func extractByRegex(matcher *regexp.Regexp, text []byte) string {
	return string(matcher.Find(text))
}

func findByElement(matcher cascadia.Selector, text []byte) *goquery.Selection {
	r := bytes.NewReader(text)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil
	}
	el := doc.FindMatcher(matcher)
	return el
}

func matchByElement(matcher cascadia.Selector, text []byte) bool {
	el := findByElement(matcher, text)
	if len(el.Nodes) > 0 {
		return true
	} else {
		return false
	}
}

//ExtractByElement extracts the specified value from the page via a
//goquery selector
func extractByElement(matcher cascadia.Selector, text []byte) string {
	el := findByElement(matcher, text)
	return el.Text()
}

// Execute extracts specified text
func (t *Task) Execute(client client.Client, vars map[string]interface{}) (err error) {
	var res string
	if err = task.Configure(t, vars); err != nil {
		return err
	}
	if len(t.Config.ElementSelector) == 0 &&
		len(t.Config.RegexSelector) == 0 {
		return errors.New("A regex or element selector is required to extract")
	}

	text := client.Response()
	if len(text) == 0 {
		return errors.New("Response text is empty, nothing to extract")
	}

	if len(t.Config.ElementSelector) > 0 {
		matcher, err := elementMatcher(t.Config.ElementSelector)
		if err != nil {
			return err
		}
		res = extractByElement(matcher, text)
	} else if len(t.Config.RegexSelector) > 0 {
		matcher, err := regexMatcher(t.Config.RegexSelector)
		if err != nil {
			return err
		}
		res = extractByRegex(matcher, text)
	}

	if len(t.Config.Key) > 0 {
		vars[t.Config.Key] = res
	} else {
		_, ok := vars["extracted"]
		if !ok {
			vars["extracted"] = make([]string, 0)
		}
		extracted, ok := vars["extracted"].([]string)
		if !ok {
			return errors.New("Could not append result to extracted")
		}
		vars["extracted"] = append(extracted, res)
	}

	return err
}
