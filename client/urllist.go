package client

import (
	"io"
	"net/url"
	"regexp"
	"runtime"
)

// URLList is a list of URLs
type URLList struct {
	urls []*url.URL
	ch   chan *url.URL
}

// NewURLList initializes a new URL list
func NewURLList() *URLList {
	return &URLList{
		urls: make([]*url.URL, 0),
		ch:   make(chan *url.URL),
	}
}

// Append parses and adds URLs to the list
func (l *URLList) Append(rawurls ...string) *URLList {
	for _, rawurl := range rawurls {
		u, err := url.Parse(rawurl)
		if err == nil {
			l.urls = append(l.urls, u)
		}
	}
	l.Rewind()
	return l
}

// Contains checks if the list contains this exact URL
func (l *URLList) Contains(u *url.URL) bool {
	for _, lu := range l.urls {
		if lu.String() == u.String() {
			return true
		}
	}
	return false
}

// ContainsRegex checks if any of the URLs in the list match a given regex
func (l *URLList) ContainsRegex(re *regexp.Regexp) bool {
	for _, lu := range l.urls {
		if matched := re.MatchString(lu.String()); matched {
			return true
		}
	}
	return false
}

// Len returns the number of URLs in the list
func (l *URLList) Len() int {
	return len(l.urls)
}

// Next returns the next URL from the iterator, or io.EOF if there are no more
// exist
func (l *URLList) Next() (*url.URL, error) {
	runtime.Gosched()
	select {
	case u := <-l.ch:
		return u, nil
	default:
		return nil, io.EOF
	}
}

// Rewind resets the iterator to the beginning
func (l *URLList) Rewind() {
	// Send to channel for iteration later
	go func() {
		for _, u := range l.urls {
			l.ch <- u
		}
	}()
	runtime.Gosched()
}
