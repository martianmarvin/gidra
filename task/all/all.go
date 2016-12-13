// Package all is a convenience package that allows importing all task sub
// packages
package all

import (
	_ "github.com/martianmarvin/gidra/task/debug"
	_ "github.com/martianmarvin/gidra/task/extract"
	_ "github.com/martianmarvin/gidra/task/http"
	_ "github.com/martianmarvin/gidra/task/sleep"
)
