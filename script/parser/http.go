package parser

import (
	"net/url"
	"strings"

	"github.com/martianmarvin/gidra"
	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
)

// TODO: Transparently replace a file path with a slice of results in all
// tasks, this should be unused for now
// URLs could be map, list, filename to read from, or url string
func parseURLList(cfg *config.Config) (*client.URLList, error) {
	var err error
	l := client.NewURLList()

	// URL string
	u, err := cfg.URL("")
	if err == nil {
		l.Append(u)
		_, err = l.Next()
		return l, err
	}

	// URL list
	urls, err := cfg.StringList("")
	if err == nil {
		l.AppendString(urls...)
		_, err = l.Next()
		return l, err
	}

	// Map with scheme and source
	m, err := cfg.StringMap("")
	if err == nil {
		source, ok := m[cfgIOSource]
		if !ok {
			return nil, gidra.FieldError{cfgIOSource}
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
		uType, err := cfg.String(cfgIOAdapter)
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

	return nil, gidra.ValueError{}
}
