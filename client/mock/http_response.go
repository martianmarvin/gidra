package mock

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

var htmlTemplate = `
<html>
<head>
<title>%s</title>
</head>
<body>
%s
</body>
</html>
`

// NewHTTPResponse constructs a *fasthttp.Response
func NewHTTPResponse(status int, headers, cookies map[string]string, title, body string) *fasthttp.Response {
	resp := fasthttp.AcquireResponse()

	resp.Header.SetStatusCode(status)
	for k, v := range headers {
		resp.Header.Set(k, v)
	}
	for k, v := range cookies {
		ck := fasthttp.AcquireCookie()
		ck.SetKey(k)
		ck.SetValue(v)
		resp.Header.SetCookie(ck)
	}

	resp.AppendBodyString(fmt.Sprint(htmlTemplate, title, body))

	return resp
}
