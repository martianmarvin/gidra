package http

import (
	"context"
	"testing"

	"github.com/martianmarvin/gidra/client/httpclient"
	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to run an http request
func testReq(t *testing.T, tsk task.Task, params map[string]interface{}) {
	assert := assert.New(t)

	taskVars := vars.NewFromMap(params)

	ctx := vars.ToContext(context.Background(), taskVars)
	ctx = httpclient.ToContext(ctx, httpclient.New())

	err := tsk.Execute(ctx)
	assert.NoError(err)

	client, ok := httpclient.FromContext(ctx)
	require.True(t, ok)
	assert.NotNil(client)

	resp, err := client.Response()
	assert.NoError(err)
	assert.NotEmpty(resp)

	// t.Log(string(resp))
}

func TestGet(t *testing.T) {
	params := make(map[string]interface{})
	params["url"] = "http://httpbin.org/get"
	params["headers"] = map[string]string{
		"user-agent": "Gidra",
		"h1":         "v1",
		"h2":         "v2",
	}

	testReq(t, task.New("get"), params)

}

func TestPost(t *testing.T) {
	params := make(map[string]interface{})
	params["url"] = "http://httpbin.org/post"
	params["headers"] = map[string]string{
		"user-agent": "Gidra",
		"h1":         "v1",
		"h2":         "v2",
	}
	params["cookies"] = map[string]string{
		"c1": "v1",
		"c2": "v2",
	}
	params["params"] = map[string]string{
		"b1": "v1",
		"b2": "b2",
	}

	testReq(t, task.New("post"), params)
}
