// Package cookiejar implements a cookie jar for fasthttp cookies
package fastcookiejar

import (
	"sync"

	"github.com/valyala/fasthttp"
)

// Jar is the main cookie jar that contains pointers to fasthttp.Cookie
type Jar struct {
	mu sync.Mutex

	// entries is a simple map of domain to cookie string
	entries map[string][]*fasthttp.Cookie
}

// New creates a new cookie jar
func New() *Jar {
	return &Jar{
		entries: make(map[string][]*fasthttp.Cookie),
	}
}

// Clear deletes all cookies in the jar, optionally matching a specific
// domain
func (j *Jar) Clear(domains ...string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	//Clear all
	if len(domains) == 0 {
		for _, cookies := range j.entries {
			for _, ck := range cookies {
				fasthttp.ReleaseCookie(ck)
			}
		}
		j.entries = make(map[string][]*fasthttp.Cookie)
	} else {
		for _, domain := range domains {
			for _, ck := range j.entries[domain] {
				fasthttp.ReleaseCookie(ck)
			}
			delete(j.entries, domain)
		}
	}
}

//SetCookies sets cookies for the specified domain
//An empty string sets cookies for all domains
// Once a cookie is deposited into the cookie jar, it may not be accessed
// again directly
func (j *Jar) SetCookies(cookies []*fasthttp.Cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for _, ck := range cookies {
		cookie := fasthttp.AcquireCookie()
		ck.CopyTo(cookie)
		fasthttp.ReleaseCookie(ck)
		domain := parseDomain(string(ck.Domain()))
		j.entries[domain] = append(j.entries[domain], cookie)
	}
}

//Set sets a cookie from a key, value pair
func (j *Jar) Set(domain, key, value string) {
	var fastcookies []*fasthttp.Cookie
	j.SetCookies(fastcookies)
	ck := newCookie(domain, key, value)
	fastcookies = append(fastcookies, ck)
	j.SetCookies(fastcookies)
}

//SetMap sets a cookie from a map of strings
func (j *Jar) SetMap(domain string, cookies map[string]string) {
	var fastcookies []*fasthttp.Cookie
	for k, v := range cookies {
		ck := newCookie(domain, k, v)
		fastcookies = append(fastcookies, ck)
	}
	j.SetCookies(fastcookies)
}

//Creates a new fasthttp.Cookie
func newCookie(domain, key, value string) *fasthttp.Cookie {
	ck := fasthttp.AcquireCookie()
	if len(domain) > 1 {
		ck.SetDomain(domain)
	}
	ck.SetKey(key)
	ck.SetValue(value)
	return ck
}

func parseDomain(domain string) string {
	if len(domain) == 0 {
		domain = "."
	}
	return domain
}

//Cookies returns a slice of all global cookies and any matching this domain
func (j *Jar) Cookies(domain string) (cookies []*fasthttp.Cookie) {
	return cookies
}
