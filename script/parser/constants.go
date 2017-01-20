package parser

// Dot-separated paths to specific config values
var (
	// Global config
	cfgConfig            = "config"
	cfgConfigLoop        = "loop"
	cfgConfigThreads     = "threads"
	cfgConfigVerbosity   = "verbosity"
	cfgConfigTaskTimeout = "task_timeout"

	// Connection
	cfgProxy           = "proxy"
	defaultProxyScheme = "http"

	//Global variablies
	cfgVars = "vars"

	//Sequences
	cfgSeqBefore = "before"
	cfgSeqTasks  = "tasks"
	cfgSeqAfter  = "finally"

	//Tasks
	cfgTaskCond        = "when"
	cfgTaskBefore      = "with"
	cfgTaskLimit       = "limit"
	cfgTaskSuccessCond = "success"
	cfgTaskSkipCond    = "skip"
	cfgTaskAbortCond   = "abort"
	cfgTaskRetryCond   = "retry"
	cfgTaskFailCond    = "fail"

	//I/O
	cfgInputs    = "inputs"
	cfgOutput    = "output"
	cfgIOSource  = "path"
	cfgIOList    = "list"
	cfgIOAdapter = "type"
	cfgIOVars    = "as"
	cfgMainInput = "main"
)

// Aliases for config keys
var cfgAliases = map[string]string{
	"repeat":  cfgConfigLoop,
	"workers": cfgConfigThreads,
	"if":      cfgTaskCond,
	"do":      cfgTaskBefore,
	"name":    cfgIOVars,
}
