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
