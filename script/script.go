package script

import (
	"github.com/martianmarvin/gidra/log"
	"github.com/martianmarvin/gidra/sequence"
	"github.com/martianmarvin/vars"
)

var Logger = log.Logger()

//Script is the runner that executes Sequences
// Unlike a Sequence or Task, a Script is fully concurrency-safe and executes
// multiple Sequences concurrently
type Script struct {
	//How many times the script should loop total
	Loop int

	//Threads is the number of concurrent sequences executing
	Threads int

	// The queue of sequences to execute
	queue chan *Sequence

	// Script-global variables applied to all sequences
	Vars *vars.Vars

	//Sequence represents the main task sequence in this script
	Sequence *sequence.Sequence

	//BeforeSequence is the sequence to run before running the main one
	BeforeSequence *sequence.Sequence

	//AfterSequence is the sequence to run before running the main one
	AfterSequence *sequence.Sequence
}

//NewScript loads and parses config YAML
func New() *Script {
	return &Script{
		queue: make(chan *Sequence, 1),
		Vars:  vars.New(),
	}
}

//Load loads a script file and initializes Sequences
func (s *Script) Load(name string) (err error) {
	params := make(map[string]interface{})

	cfg, err := parseScript(name)
	if err != nil {
		return
	}

	//TODO get loop from number of input lines if not explicitly defined
	s.Loop = cfg.UInt(cfgConfigLoop, 1)
	s.Threads, _ = cfg.Int(cfgConfigThreads)

	s.BeforeSequence, err = parseSequence(cfgSeqBefore, cfg)
	if err == nil {
		s.Add(s.BeforeSequence)
	}

	s.Sequence, err = parseSequence(cfgSeqTasks, cfg)
	if err != nil {
		// Main sequence is required
		return err
	}

	//TODO construct vars from input for each sequence
	for i := 0; i <= s.Loop; i++ {
		ivars := parseInputVars(params)
		seq, err := s.Sequence.Clone()
		if err != nil {
			return err
		}
		seq.Configure(s.Vars, ivars)
		s.Add(seq)
	}

	s.AfterSequence, err = parseSequence(cfgSeqAfter, cfg)
	if err == nil {
		s.AfterSequence.Configure(s.Vars)
		s.Add(s.AfterSequence)
	}

	return err
}

// Add adds a new sequence to the Script's queue
func (s *Script) Add(seq *Sequence) {
	go func() {
		s.queue <- seq
	}()
}
