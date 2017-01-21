package parser

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/script/options"
)

func init() {
	Register(cfgProxy, proxyParser)
}

func proxyParser(s *options.ScriptOptions, cfg *config.Config) error {
	var err error

	scheme := cfg.GetString(cfgIOAdapter)
	if len(scheme) == 0 {
		scheme = defaultProxyScheme
	}

	proxies, err := parseProxy(cfg)
	if err != nil {
		return err
	}
	for _, u := range proxies {
		if len(u.Scheme) == 0 {
			u.Scheme = scheme
		}
	}

	s.Proxies = proxies
	return nil
}

func parseProxy(cfg *config.Config) ([]*url.URL, error) {
	var proxies []*url.URL
	// If proxy is a single url, create reader for it
	rawurl, err := cfg.GetStringE(cfgProxy)
	if err == nil {
		u, err := url.Parse(rawurl)
		if err != nil {
			return proxies, err
		}
		proxies = append(proxies, u)
		return proxies, nil
	}

	// Read slice into reader
	rawurls, err := cfg.GetStringSliceE(cfgProxy)
	if err == nil {
		for _, rawurl := range rawurls {
			u, err := url.Parse(rawurl)
			if err != nil {
				return proxies, err
			}
			proxies = append(proxies, u)
		}
		return proxies, nil
	}

	// Read from input file
	subcfg := cfg.Get(cfgProxy, nil)
	reader, err := parseInput(subcfg)
	if err != nil {
		return proxies, err
	}
	for {
		row, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return proxies, err
		} else {
			rawurl := row.GetIndex(0).MustString()
			u, err := url.Parse(rawurl)
			if err != nil {
				return proxies, err
			}
			proxies = append(proxies, u)
		}
	}

	return proxies, nil
}

// Reader from single string
func stringReader(s string) (datasource.ReadableTable, error) {
	reader, _ := datasource.NewReader("csv")
	// Add header
	s = "url\n" + s
	r := strings.NewReader(s)
	_, err := reader.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

// Append scheme if it is not present
func appendScheme(rawurl, scheme string) string {
	if strings.Contains(rawurl, "://") {
		return rawurl
	}
	return fmt.Sprintf("%s://%s", strings.ToLower(scheme), rawurl)
}

// Parse list into a iterable table
// URLs could be map, list, filename to read from, or url string
func parseURLList(key string, cfg *config.Config) (*client.URLList, error) {
	var err error
	l := client.NewURLList()

	// single URL string
	u, err := cfg.GetURLE(key)
	if err == nil {
		l.Append(u)
		_, err = l.Next()
		return l, err
	}

	// Otherwise look for a list
	lk := key + "." + cfgIOList

	// array/list of URLs
	urls, err := cfg.GetStringSliceE(lk)
	if err == nil {
		l.AppendString(urls...)
		_, err = l.Next()
		return l, err
	}

	// String representing an input

	// Map with list and optional scheme
	m, err := cfg.GetStringMapE(key)
	if err == nil {
		source, ok := m[cfgIOSource]
		if !ok {
			return nil, config.KeyError{Name: cfgIOSource, Err: config.ErrRequired}
		}
		lines, err := datasource.ReadLines(source)
		if err != nil {
			return nil, err
		}
		urls := make([]*url.URL, 0)
		for _, line := range lines {
			u, err := url.Parse(line)
			if err != nil {
				return nil, err
			}
			urls = append(urls, u)
		}
		uType, err := cfg.GetStringE(cfgIOAdapter)
		if err == nil {
			uType = strings.ToLower(uType)
			for _, u := range urls {
				u.Scheme = uType
			}
		}
		l.Append(urls...)
		_, err = l.Next()
		return l, err
	}

	return nil, config.ValueError{}
}
