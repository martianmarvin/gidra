package parser

import (
	"strings"

	"github.com/martianmarvin/gidra/config"
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
	cookies, err := cfg.StringMap(cfgHTTPHeaders)
	if err == nil {
		opts.Cookies = cookies
	}

	timeout, err := cfg.Duration(cfgHTTPTimeout)
	if err != nil && !strings.HasPrefix(err.Error(), "Invalid path") {
		return nil, err
	}
	opts.Timeout = timeout

	proxy, err := cfg.URL(cfgHTTPProxy)
	if err != nil && !strings.HasPrefix(err.Error(), "Invalid path") {
		return nil, err
	}
	opts.Proxy = proxy

	body, err := cfg.String(cfgHTTPBody)
	if err == nil {
		opts.Body = []byte(body)
	}

	return opts, nil
}

func normalizeHeader(key string) string {
	var res []byte
	res = fasthttp.AppendNormalizedHeaderKey(res, key)
	return string(res)
}
