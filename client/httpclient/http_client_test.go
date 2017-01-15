package httpclient

import (
	"net/url"
	"testing"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	var err error
	assert := assert.New(t)

	cfg := config.New()
	cfg.Set(cfgFollowRedirects, true)
	cfg.Set(cfgHeaders, map[string]string{"h1": "v1", "h2": "v2"})

	c := New()
	err = c.Configure(cfg)
	assert.NoError(err)
	assert.Equal(c.Options.FollowRedirects, true)
	assert.Equal(c.Options.Headers["h1"], "v1")

	proxyurl, _ := url.Parse("socks5://127.0.0.1:9000")
	cfg.Set("proxy", proxyurl.String())
	err = c.Configure(cfg)
	assert.NoError(err)
	assert.Equal(proxyurl, c.Options.Proxy)

	cfg.Set(cfgHeaders, map[string]string{"h1": "v1b", "h3": "v3"})
	err = c.Configure(cfg)
	assert.NoError(err)
	assert.Equal(c.Options.Headers["h1"], "v1b")
	assert.Equal(c.Options.Headers["h2"], "v2")
	assert.Equal(c.Options.Headers["h3"], "v3")
}

func TestSimulate(t *testing.T) {
	var err error
	assert := assert.New(t)

	cfg := config.New()
	cfg.Set(cfgSimulate, true)
	cfg.Set(cfgURL, "http://www.httpbin.org/get")
	cfg.Set(cfgHeaders, map[string]string{"h1": "v1", "h2": "v2"})

	c := New()
	err = c.Configure(cfg)
	assert.NoError(err)

	err = c.Do(nil)
	assert.NoError(err)
	resp, err := c.Page()
	assert.NoError(err)
	assert.NotNil(resp)

	assert.Contains(resp.Body, "/get")
	assert.Contains(resp.Body, "v1")
	assert.Contains(resp.Body, "Accept-Encoding")

	// Test combining client options with per-request options
	u, _ := url.Parse("http://www.httpbin.org/ip")
	opts := &client.HTTPOptions{
		URL:     u,
		Headers: map[string]string{"h1": "v1a", "h3": "v3"},
	}

	err = c.Do(opts)
	assert.NoError(err)
	resp, err = c.Page()
	assert.NoError(err)
	assert.NotNil(resp)

	assert.Contains(resp.Body, "/ip")
	assert.Contains(resp.Body, "v1a")
	assert.Contains(resp.Body, "v2")
	assert.Contains(resp.Body, "v3")
}
