package client

import (
	"bytes"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/valyala/fasthttp"
)

var (
	ErrEmpty = errors.New("Body is empty")
)

// Page represents a single http response with easier accessors that
// *fasthttp.Response
type Page struct {
	Bytes []byte

	URL *url.URL

	Redirects URLList

	Status int

	Headers map[string]string

	Cookies map[string]string

	Title string

	// Headers as a string
	Header string

	Body string

	// All of the form fields on the page in one map

	Forms FormList

	json *simplejson.Json

	dom *goquery.Selection
}

func NewPage() *Page {
	return &Page{
		Headers:   make(map[string]string),
		Cookies:   make(map[string]string),
		Forms:     make(FormList, 0),
		Redirects: make(URLList, 0),
	}
}

// Parse parses a *fasthttp.Response into this page
func (p *Page) Parse(resp *fasthttp.Response) error {
	var err error
	if resp == nil {
		return ErrEmpty
	}

	//TODO look at getting body directly if []byte(string) conversion is
	//expensive
	p.Bytes = []byte(resp.String())

	if len(p.Bytes) == 0 {
		return ErrEmpty
	}

	// Headers
	p.Status = resp.StatusCode()
	p.Header = resp.Header.String()

	var i int
	resp.Header.VisitAll(func(k, v []byte) {
		key := string(k)
		val := string(v)
		// For duplicate header values
		if _, ok := p.Headers[key]; ok {
			key = key + "." + strconv.Itoa(i)
			i += 1
		}
		p.Headers[key] = val
	})

	// Cookies
	i = 0
	resp.Header.VisitAllCookie(func(k, v []byte) {
		key := string(k)
		val := string(v)
		// For duplicate cookie values
		if _, ok := p.Headers[key]; ok {
			key = key + "." + strconv.Itoa(i)
			i += 1
		}
		p.Headers[key] = val
	})

	// Body
	var body []byte
	if p.Headers["Content-Encoding"] == "gzip" {
		body, err = resp.BodyGunzip()
		if err != nil {
			return err
		}
	} else {
		body = resp.Body()
	}
	if len(body) == 0 {
		return nil
	}
	p.Body = string(body)

	r := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		// Return nil, document is probably not html
		return nil
	}
	p.dom = doc.AndSelf()

	p.Title = p.dom.Find("title").Text()

	// Form
	//TODO parse textarea, select, and submit elements
	p.dom.Find("form").Each(func(i int, formEl *goquery.Selection) {
		form := NewForm()
		form.Method = strings.ToUpper(formEl.AttrOr("method", "POST"))
		if action, ok := formEl.Attr("action"); ok {
			u, err := url.Parse(action)
			if err == nil {
				if u.Host == "" {
					u.Host = p.URL.Host
				}
				if u.Scheme == "" {
					u.Scheme = p.URL.Scheme
				}
				form.URL = u
			}
		} else {
			form.URL, _ = url.Parse(p.URL.String())
		}

		formEl.Find("input").Each(func(i int, el *goquery.Selection) {
			if name, ok := el.Attr("name"); ok {
				form.Fields[name] = el.AttrOr("value", "")
			}
		})

		p.Forms = append(p.Forms, form)
	})

	return nil
}

// String returns the entire page(headers and body) as a string
func (p *Page) String() string {
	return string(p.Bytes)
}

// Json returns the page body in JSON format
func (p *Page) Json() (*simplejson.Json, error) {
	if p.json != nil {
		return p.json, nil
	}
	json, err := simplejson.NewJson(p.Bytes)
	if err != nil {
		return nil, err
	}
	p.json = json
	return p.json, nil

}
