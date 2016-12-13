package parser

import (
	"net/url"
	"strings"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/valyala/fasthttp"
)

func init() {
	Register(cfgHTTP, httpParser)
}

func httpParser(s *options.ScriptOptions, cfg *config.Config) error {
	opts, err := ParseHTTP(cfg)
	if err != nil {
		return err
	}
	s.HTTP = opts

	return nil
}

// Parses http options from config
func ParseHTTP(cfg *config.Config) (*options.HTTPOptions, error) {
	opts := &options.HTTPOptions{
		FollowRedirects: cfg.UBool(cfgHTTPFollowRedirects),
		Headers:         make(map[string]string),
		Cookies:         make(map[string]string),
		Params:          make(map[string]string),
	}

	headers, err := cfg.StringMap(cfgHTTPHeaders)
	if err == nil {
		for k, v := range headers {
			k := normalizeHeader(k)
			opts.Headers[k] = v
		}
	}
	cookies, err := cfg.StringMap(cfgHTTPCookies)
	if err == nil {
		opts.Cookies = cookies
	}

	timeout, err := cfg.Duration(cfgHTTPTimeout)
	if err == nil {
		opts.Timeout = timeout
	}

	if proxycfg, ok := cfg.CheckGet(cfgHTTPProxy); ok {
		fn, err := parseProxy(proxycfg)
		if err != nil {
			return nil, err
		}
		opts.Proxy = fn
	}

	body, err := cfg.String(cfgHTTPBody)
	if err == nil {
		opts.Body = []byte(body)
	}

	return opts, nil
}

// Proxy could be map, list, filename to read from, or url string
func parseProxy(cfg *config.Config) (*client.URLList, error) {
	var err error
	l := client.NewURLList()

	// URL string
	proxy, err := cfg.URL("")
	if err == nil {
		l.Append(proxy)
		_, err = l.Next()
		return l, err
	}

	// URL list
	proxies, err := cfg.StringList("")
	if err == nil {
		l.AppendString(proxies...)
		_, err = l.Next()
		return l, err
	}

	// Map with type and source
	m, err := cfg.StringMap("")
	if err == nil {
		source, ok := m[cfgIOSource]
		if !ok {
			return nil, FieldError{cfgIOSource}
		}
		lines, err := datasource.ReadLines(source)
		if err != nil {
			return nil, err
		}
		proxyurls := make([]*url.URL, 0)
		for _, line := range lines {
			u, err := url.Parse(line)
			if err != nil {
				return nil, err
			}
			proxyurls = append(proxyurls, u)
		}
		proxyType, err := cfg.String(cfgIOAdapter)
		if err == nil {
			proxyType = strings.ToLower(proxyType)
			for _, u := range proxyurls {
				u.Scheme = proxyType
			}
		}
		l.Append(proxyurls...)
		_, err = l.Next()
		return l, err
	}

	return nil, ValueError{Name: cfgHTTPProxy}
}

func normalizeHeader(key string) string {
	var res []byte
	res = fasthttp.AppendNormalizedHeaderKey(res, key)
	return string(res)
}
