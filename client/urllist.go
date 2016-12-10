package client

import (
	"net/url"
	"regexp"
)

// URLList is a list of URLs
type URLList []*url.URL

// Append parses and adds URLs to the list
func (l URLList) Append(rawurls ...string) URLList {
	for _, rawurl := range rawurls {
		u, err := url.Parse(rawurl)
		if err == nil {
			l = append(l, u)
		}
	}
	return l
}

func (l URLList) Contains(u *url.URL) bool {
	for _, lu := range l {
		if lu.String() == u.String() {
			return true
		}
	}
	return false
}

// ContainsRegex checks if any of the URLs in the list match a given regex
func (l URLList) ContainsRegex(re *regexp.Regexp) bool {
	for _, lu := range l {
		if matched := re.MatchString(lu.String()); matched {
			return true
		}
	}
	return false
}
