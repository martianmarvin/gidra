package debug

import (
	"context"
	"strings"
	"testing"

	"github.com/martianmarvin/gidra/client"
	"github.com/martianmarvin/gidra/global"
	"github.com/martianmarvin/gidra/task"
	"github.com/martianmarvin/vars"
	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	var err error
	assert := assert.New(t)
	ctx := context.Background()

	ctx = vars.ToContext(ctx, vars.New())

	taskVars := vars.FromContext(ctx)
	taskVars.Set("host", "http://www.example.com")
	taskVars.Set("user", "John Smith")

	g := global.New()
	g.Page = client.NewPage()
	g.Page.Title = "Lorem Ipsum"
	g.Page.Headers = map[string]string{
		"h1": "v1",
		"h2": "v2",
	}
	g.Page.Body = strings.Repeat("Lorem ipsum dolor sit amet.", 20)

	ctx = global.ToContext(ctx, g)

	err = task.Run(ctx, "debug")
	assert.NoError(err)

}
