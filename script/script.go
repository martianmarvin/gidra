package script

import (
	"context"
	"sync"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/log"
	"github.com/martianmarvin/gidra/sequence"
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
	queue chan *sequence.Sequence

	// Datasources to read from
	input map[string]datasource.ReadableTable

	output *datasource.WriteCloser
}

// New initializes a new Script
func New() *Script {
	return &Script{
		queue: make(chan *Sequence, 1),
	}
}

// Open reads a script file and Loads it into a script instance
func Open(name string) (*Script, error) {
	cfg, err := parseScript(name)
	if err != nil {
		return nil, err
	}
	s := New()
	err = s.Load(cfg)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Load loads a script file and initializes Sequences
// Once Load has been called on the script, no new sequences can be added, and
// it should not be called again.
func (s *Script) Load(cfg *config.Config) (err error) {
	ctx := context.Background()

	seqVars := parseVars(cfg)

	s.Threads, _ = cfg.Int(cfgConfigThreads)

	beforeSequence, err = parseSequence(cfgSeqBefore, cfg)
	if err == nil {
		s.Add(beforeSequence)
	}

	mainSequence, err = parseSequence(cfgSeqTasks, cfg)
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

	afterSequence, err = parseSequence(cfgSeqAfter, cfg)
	if err == nil {
		s.AfterSequence.Configure(s.Vars)
		s.Add(s.AfterSequence)
	}

	close(s.queue)

	return err
}

// Add adds a new sequence to the Script's queue
func (s *Script) Add(seq *Sequence) {
	go func() {
		s.queue <- seq
	}()
}

// Run runs all of the script's sequences
func (s *Script) Run() {
	var wg sync.WaitGroup
	results := make(chan *sequence.Result)
	go resultProcessor(results)

	Logger.Info("Starting Workers")
	for i := 0; i < s.Threads; i++ {
		wg.Add(1)
		go runner(s.queue, results, &wg)
	}
	go func() {
		wg.Wait()
		Logger.Info("All Workers Done")
		close(results)
	}()
}

// Goroutine worker that runs sequences and returns results
func runner(queue <-chan *sequence.Sequence, results chan<- *sequence.Result, wg *sync.WaitGroup) {
	for seq := range queue {
		results <- seq.Execute()
	}
	wg.Done()
}

// Goroutine that receives results and processes them with OutputFunc
func resultProcessor(results <-chan *sequence.Result, output datasource.WriteableTable, filters ...datasource.FilterFunc) {
	for res := range results {
		if err := res.Err(); err != nil {
			Logger.Error(err)
			continue
		}

	}
}
