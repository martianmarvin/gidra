package http

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/martianmarvin/gidra/client/httpclient"
	"github.com/martianmarvin/gidra/client/mock"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to run an http request
func testReq(t *testing.T, tsk task.Task, conf string, expected ...string) {
	assert := assert.New(t)

	r := strings.NewReader(conf)
	cfg := config.Must(config.ParseYaml(r))

	ctx := config.ToContext(context.Background(), cfg)
	ctx = httpclient.ToContext(ctx, httpclient.New())

	err := tsk.Execute(ctx)
	assert.NoError(err)

	client, ok := httpclient.FromContext(ctx)
	require.True(t, ok)
	assert.NotNil(client)

	resp, err := client.Response()
	assert.NoError(err)
	for _, s := range expected {
		assert.Contains(string(resp), s)
	}

}

func TestGet(t *testing.T) {
	ts := mock.NewServer()
	defer ts.Close()
	conf := fmt.Sprintf("url: %s\n", ts.URL)
	conf += `headers:
	user-agent: Gidra
	h1: v1
	h2: v2
`

	testReq(t, task.New("get"), conf, "Gidra", "v2")

}

func TestPost(t *testing.T) {
	ts := mock.NewServer()
	defer ts.Close()
	conf := fmt.Sprintf("url: %s\n", ts.URL)
	conf += `headers:
	user-agent: Gidra
	h1: v1
	h2: v2
cookies:
	c1: v1
	c2: v2
params:
	b1: v1
	b2: b2
`

	testReq(t, task.New("post"), conf, "Gidra", "v2", "c2", "b2")
}
