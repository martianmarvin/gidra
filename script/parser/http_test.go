package parser

import "testing"

var testProxyCfgs = map[string]string{
	"string": `proxy: socks5://1.2.3.4:9000`,

	"slice": `proxy:
  - socks5://1.2.3.4:9000
  - https://1.2.3.4:8888
`,
}

var httpTests = []parserTest{
	{testProxyCfgs["string"], `{{ (.Proxies.Next.GetIndex 0).MustString }}`, "socks5://1.2.3.4:9000"},
	{testProxyCfgs["slice"], `{{ (.Proxies.Next.GetIndex 0).MustString }}`, "socks5://1.2.3.4:9000"},
	{testProxyCfgs["slice"], `{{.Proxies.Next |null}}{{ (.Proxies.Next.GetIndex 0).MustString }}`, "https://1.2.3.4:8888"},
}

func TestProxy(t *testing.T) {
	for _, v := range httpTests {
		testParser(t, proxyParser, v)
	}
}
