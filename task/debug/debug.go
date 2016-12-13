// debug is a Task that prints out the current global and local state
package debug

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/martianmarvin/gidra/global"
	"github.com/martianmarvin/gidra/task"
)

func init() {
	task.Register("debug", New)
}

type Task struct {
	task.Loggable
}

func New() task.Task {
	return &Task{Loggable: task.NewLoggable()}
}

func (t *Task) Execute(ctx context.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', tabwriter.FilterHTML|tabwriter.StripEscape)
	g := global.FromContext(ctx)

	printHeader(w, color.YellowString("VARS"))
	printMap(w, g.Vars)

	fmt.Println(g.Page)

	printHeader(w, color.YellowString("LAST PAGE"))
	fmt.Fprintf(w, "%s\t%40s\n", color.BlueString("Title"), color.YellowString(g.Page.Title))
	fmt.Fprintln(w, hr)
	fmt.Fprintln(w, color.BlueString("Headers"))
	printStringMap(w, g.Page.Headers)
	fmt.Fprintln(w, hr)
	fmt.Fprintf(w, "%s\t%80s...\n", color.BlueString("Body"), g.Page.Body)

	w.Flush()
	return nil
}

func printHeader(w io.Writer, h string) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, h)
	fmt.Fprintln(w, hr)
}

func printMap(w io.Writer, m map[string]interface{}) {
	for k, v := range m {
		fmt.Fprintf(w, "%s\t%s\n", color.BlueString(k), color.YellowString("%v", v))
		fmt.Fprintln(w, hr)
	}
}

func printStringMap(w io.Writer, m map[string]string) {
	for k, v := range m {
		fmt.Fprintf(w, "%s\t%s\n", color.BlueString(k), color.YellowString("%v", v))
	}
}
