package client

import (
	"net/url"
	"strconv"
)

type Form struct {
	Method string

	URL *url.URL

	Fields map[string]string
}

// NewForm returns a new empty Form
func NewForm() *Form {
	return &Form{Fields: make(map[string]string)}
}

type FormList []*Form

// Fields returns the combined fields of all forms in this list
func (l FormList) Fields() map[string]string {
	fields := make(map[string]string)
	for n, form := range l {
		if len(form.Fields) == 0 {
			continue
		}
		for key, val := range form.Fields {
			if _, ok := fields[key]; ok {
				key = key + "." + strconv.Itoa(n)
			}
			fields[key] = val
		}
	}
	return fields
}
