package extract

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
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
	task.Worker
	task.Configurable

	Config *Config
}

type Config struct {
	//Regex string to match to extract
	RegexSelector string `task:"regex"`

	//Jquery selector to match to extract
	ElementSelector string `task:"element"`

	//Key to save match in, if any
	Key string `task:"as"`

	// The text to extract from
	Text []byte

	// Whether to strip leading and trailing spaces from extracted value
	TrimSpace bool `task:"trim"`
}

func NewTask() task.Task {
	t := &Task{
		Config: &Config{TrimSpace: true},
		Worker: task.NewWorker(),
	}
	t.Configurable = task.NewConfigurable(t.Config)
	return t
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
func extractByRegex(matcher *regexp.Regexp, text []byte) []string {
	var results []string
	matches := matcher.FindSubmatch(text)
	if len(matches) < 2 {
		return results
	}
	for _, result := range matches[1:] {
		results = append(results, string(result))
	}
	return results
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
func extractByElement(matcher cascadia.Selector, text []byte) []string {
	els := findByElement(matcher, text)
	return els.Map(func(n int, el *goquery.Selection) string {
		return el.Text()
	})
}

// Execute extracts specified text
func (t *Task) Execute(ctx context.Context) (err error) {
	var results []string
	var rk string
	var text []byte

	if len(t.Config.Key) > 0 {
		rk = t.Config.Key
	} else {
		rk = "extracted"
	}

	if len(t.Config.ElementSelector) == 0 &&
		len(t.Config.RegexSelector) == 0 {
		return errors.New("A regex or element selector is required to extract")
	}

	taskVars := vars.FromContext(ctx)
	extracted := taskVars.Get(rk).MustStringArray()

	// If no text is provided, default to the last response from the client
	if len(t.Config.Text) > 0 {
		text = t.Config.Text
	} else {
		client, ok := client.FromContext(ctx)
		if !ok {
			return errors.New("No client to extract from, make an http request first")
		}

		text, err = client.Response()
		if err != nil {
			return err
		}
	}

	if len(text) == 0 {
		return errors.New("Response text is empty, nothing to extract")
	}

	if len(t.Config.ElementSelector) > 0 {
		matcher, err := elementMatcher(t.Config.ElementSelector)
		if err != nil {
			return err
		}
		results = extractByElement(matcher, text)
	} else if len(t.Config.RegexSelector) > 0 {
		matcher, err := regexMatcher(t.Config.RegexSelector)
		if err != nil {
			return err
		}
		results = extractByRegex(matcher, text)
	}

	if len(results) == 0 {
		return err
	}

	for _, res := range results {
		if t.Config.TrimSpace {
			res = strings.TrimSpace(res)
		}
		extracted = append(extracted, res)
	}
	taskVars.SetPath(rk, extracted)

	return err
}
