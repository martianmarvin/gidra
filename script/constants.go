package script

// Dot-separated paths to specific config values
var (
	// Global config
	cfgConfig        = "config"
	cfgConfigLoop    = cfgConfig + ".loop"
	cfgConfigThreads = cfgConfig + ".threads"

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
	cfgTaskSuccessCond = "success" + "." + cfgTaskCond
	cfgTaskAbortCond   = "abort" + "." + cfgTaskCond
	cfgTaskRetryCond   = "retry" + "." + cfgTaskCond
	cfgTaskRetryChange = "retry" + "." + cfgTaskBefore
	cfgTaskRetryLimit  = "retry" + "." + cfgTaskLimit
	cfgTaskFailCond    = "fail" + "." + cfgTaskCond

	//I/O
	cfgIOIn      = "inputs"
	cfgIOOut     = "output"
	cfgIOSource  = "source"
	cfgIOAdapter = "adapter"
	cfgIOVars    = "as"
)
