package parser

import (
	"net/url"

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
	// If proxy is a single url, create reader for it
	rawurl, err := cfg.GetStringE(cfgProxy)
	if err == nil {
		rawurls := make([]string, 1)
		rawurls[0] = rawurl
		return parseURLs(rawurls), nil
	}

	// Read slice into reader
	rawurls, err := cfg.GetStringSliceE(cfgProxy)
	if err == nil {
		return parseURLs(rawurls), nil
	}

	// Read from input file
	subcfg := cfg.Get(cfgProxy, nil)
	fp, err := subcfg.GetStringE(cfgIOSource)
	rawurls, err = datasource.ReadLines(fp)
	if err != nil {
		return nil, err
	}

	return parseURLs(rawurls), nil

}

// Parse a slice of string URLs, filling in nil if the url cannot be parsed
func parseURLs(rawurls []string) []*url.URL {
	urls := make([]*url.URL, len(rawurls))
	for i, rawurl := range rawurls {
		u, err := url.Parse(rawurl)
		if err != nil {
			continue
		}
		urls[i] = u
	}
	return urls
}
