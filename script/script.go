package script

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/martianmarvin/gidra/client/httpclient"
	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/global"
	"github.com/martianmarvin/gidra/log"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/martianmarvin/gidra/script/parser"
	"github.com/martianmarvin/gidra/sequence"
)

var Logger = log.Logger()

//Script is the runner that executes Sequences
// Unlike a Sequence or Task, a Script is fully concurrency-safe and executes
// multiple Sequences concurrently
type Script struct {
	Options *options.ScriptOptions

	// The queue of sequences to execute
	queue chan *sequence.Sequence

	// Channel that receives results from completed sequences
	results chan *sequence.Result
}

// New initializes a new Script
func New() *Script {
	return &Script{
		queue: make(chan *sequence.Sequence, 1),
	}
}

// Open reads a script file and Loads it into a script instance
func Open(name string) (*Script, error) {
	f, err := os.Open(name)
	cfg, err := config.ParseYaml(f)
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
func (s *Script) Load(cfg *config.Config) (err error) {

	//Parse yaml script
	err = parser.Configure(s.Options, cfg)
	if err != nil {
		return err
	}

	ctx := configureContext(context.Background(), s.Options)

	if seq := s.Options.BeforeSequence; seq != nil {
		s.Add(seq.WithContext(ctx))
	}

	if s.Options.MainSequence == nil {
		return errors.New("Could not parse task list")
	}

	for i := 0; i <= s.Options.Loop; i++ {
		s.Add(s.Options.MainSequence.WithContext(ctx))
	}

	if seq := s.Options.AfterSequence; seq != nil {
		s.Add(seq.WithContext(ctx))
	}

	return err
}

// Add adds a new sequence to the Script's queue
func (s *Script) Add(seq *sequence.Sequence) {
	go func() {
		s.queue <- seq
	}()
}

// Run runs all of the script's sequences
// Once Run has been called on the script, no new sequences can be added, and
// it should not be called again.
func (s *Script) Run() {
	close(s.queue)
	var wg sync.WaitGroup
	results := make(chan *sequence.Result)
	go resultProcessor(results, s.Options.Output)

	Logger.Info("Starting Workers")
	for i := 0; i < s.Options.Threads; i++ {
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

// Instantiates context with global objects based on options
func configureContext(ctx context.Context, opts *options.ScriptOptions) context.Context {
	// Log
	log.SetLevel(logrus.Level(opts.Verbosity))
	ctx = log.ToContext(ctx, log.Logger())

	// HTTP Client
	if opts.HTTP != nil {
		client := httpclient.New().WithOptions(opts.HTTP)
		ctx = httpclient.ToContext(ctx, client)
	}

	// Template Globals
	g := global.New()
	if opts.Vars != nil {
		g.Vars = opts.Vars.Map()
	}

	if len(opts.Input) > 0 {
		g.Inputs = opts.Input
	}

	ctx = global.ToContext(ctx, g)

	return ctx
}
