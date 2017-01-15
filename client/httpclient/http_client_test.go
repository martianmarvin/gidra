package httpclient

import (
	"net/url"
	"testing"

	"github.com/martianmarvin/gidra/config"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestConfig(t *testing.T) {
	var err error
	assert := assert.New(t)

	cfg := config.New()
	cfg.Set(cfgFollowRedirects, true)
	cfg.Set(cfgHeaders, map[string]string{"h1": "v1", "h2": "v2"})

	client := New()
	err = client.Configure(cfg)
	assert.NoError(err)
	assert.Equal(client.Options.FollowRedirects, true)
	assert.Equal(client.Options.Headers["h1"], "v1")

	proxyurl, _ := url.Parse("socks5://127.0.0.1:9000")
	cfg.Set("proxy", proxyurl.String())
	err = client.Configure(cfg)
	assert.NoError(err)
	assert.Equal(proxyurl, client.Options.Proxy)

	cfg.Set(cfgHeaders, map[string]string{"h1": "v1b", "h3": "v3"})
	err = client.Configure(cfg)
	assert.NoError(err)
	assert.Equal(client.Options.Headers["h1"], "v1b")
	assert.Equal(client.Options.Headers["h2"], "v2")
	assert.Equal(client.Options.Headers["h3"], "v3")
}

func TestSimulate(t *testing.T) {
	var err error
	assert := assert.New(t)

	cfg := config.New()
	cfg.Set(cfgSimulate, true)
	cfg.Set(cfgURL, "http://www.httpbin.org/get")
	cfg.Set(cfgHeaders, map[string]string{"h1": "v1", "h2": "v2"})

	client := New()
	err = client.Configure(cfg)
	assert.NoError(err)

	req := &fasthttp.Request{}
	err = client.Do(req)
	assert.NoError(err)
	resp, err := client.Page()
	assert.NoError(err)
	assert.NotNil(resp)

	assert.Contains(resp.Body, "/get")
	assert.Contains(resp.Body, "v1")
	assert.Contains(resp.Body, "Accept-Encoding")
}
