package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testurls = []string{"https://www.google.com", "http://www.yahoo.com",
	"socks5://127.0.0.1:9000", "http://user:pass@localhost:9000"}

func TestURLList(t *testing.T) {
	assert := assert.New(t)

	l := NewURLList()
	l.Append(testurls...)

	assert.Equal(l.Len(), len(testurls))

	for _, testurl := range testurls {
		u, ok := l.Next()
		assert.True(ok)
		assert.Equal(testurl, u.String())
	}
	u, ok := l.Next()
	assert.False(ok)
	assert.Nil(u)

	l.Rewind()
	u, ok = l.Next()
	assert.True(ok)
	assert.Equal(testurls[0], u.String())

}
