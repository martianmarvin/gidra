package parser

import (
	"fmt"
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
	var reader datasource.ReadableTable

	scheme := cfg.GetString(cfgIOAdapter)
	if len(scheme) == 0 {
		scheme = defaultProxyScheme
	}

	reader, err = parseProxy(cfg)
	if err != nil {
		return err
	}

	// Append default scheme to all rows
	// TODO other columns to mark proxy good/bad, speed, etc
	err = reader.Filter(func(r *datasource.Row) *datasource.Row {
		r.SetColumns([]string{"rawurl"})
		rawurl := r.GetIndex(0).MustString()
		r.Set("rawurl", appendScheme(rawurl, scheme))
		return r
	})

	if err != nil {
		return err
	}

	s.Proxies = reader
	return nil
}

func parseProxy(cfg *config.Config) (datasource.ReadableTable, error) {
	// If proxy is a single url, create reader for it
	rawurl, err := cfg.GetStringE(cfgProxy)
	if err == nil {
		return stringReader(rawurl)
	}

	// Read slice into reader
	rawurls, err := cfg.GetStringSliceE(cfgProxy)
	if err == nil {
		for i, rawurl := range rawurls {
			rawurls[i] = appendScheme(rawurl, defaultProxyScheme)
		}
		return stringReader(strings.Join(rawurls, "\n"))
	}

	// Read from input file
	subcfg := cfg.Get(cfgProxy, nil)
	return parseInput(subcfg)
}

// Reader from single string
func stringReader(s string) (datasource.ReadableTable, error) {
	reader, _ := datasource.NewReader("csv")
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
