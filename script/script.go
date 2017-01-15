package script

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/martianmarvin/gidra/config"
	"github.com/martianmarvin/gidra/datasource"
	"github.com/martianmarvin/gidra/global"
	"github.com/martianmarvin/gidra/log"
	"github.com/martianmarvin/gidra/script/options"
	"github.com/martianmarvin/gidra/script/parser"
	"github.com/martianmarvin/gidra/sequence"
	"github.com/martianmarvin/vars"
)

var Logger = log.Logger()

//Script is the runner that executes Sequences
// Unlike a Sequence or Task, a Script is fully concurrency-safe and executes
// multiple Sequences concurrently
type Script struct {
	Options *options.ScriptOptions

	// All sequences loaded for this script
	sequences []*sequence.Sequence

	// The queue of sequences to execute
	queue chan *sequence.Sequence

	// Channel that receives results from completed sequences
	results chan *sequence.Result

	wg sync.WaitGroup
}

// New initializes a new Script
func New() *Script {
	return &Script{
		Options: options.New(),
		results: make(chan *sequence.Result),
	}
}

// Open reads a script file and Loads it into a script instance
func Open(r io.Reader) (*Script, error) {
	cfg, err := config.ParseYaml(r)
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

// OpenFile reads a script from a file
func OpenFile(fn string) (*Script, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Open(f)
}

// Load loads a script file and initializes Sequences
func (s *Script) Load(cfg *config.Config) error {
	var err error

	// Parse yaml script and merge with default config
	cfg = config.Default.Extend(cfg)

	err = parser.Configure(s.Options, cfg)
	if err != nil {
		return err
	}

	// Buffer queue for concurrency
	s.queue = make(chan *sequence.Sequence, s.Options.Threads)

	// If no inputs, create NopReader
	if len(s.Options.Input) == 0 {
		s.Options.Input = make(map[string]datasource.ReadableTable)
		s.Options.Input["main"] = datasource.NewNopReader(s.Options.Loop)
	}

	if seq := s.Options.BeforeSequence; seq != nil {
		s.Add(seq)
	}

	if s.Options.MainSequence == nil {
		return errors.New("Could not parse task list")
	}

	iter := s.Options.Input["main"]
	Logger.WithField("n", iter.Len()).Warn("Running loop")
	for {
		row, err := iter.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		seq := s.Options.MainSequence.Copy()
		seq.Row = row

		s.Add(seq)
	}
	iter.Close()

	if seq := s.Options.AfterSequence; seq != nil {
		s.Add(seq)
	}

	return err
}

// Add adds a new sequence to the Script's queue
func (s *Script) Add(seq *sequence.Sequence) {
	s.sequences = append(s.sequences, seq)
}

// Run runs all of the script's sequences
// Once Run has been called on the script, no new sequences can be added, and
// it should not be called again.
// Run returns a signal channel that can accept a boolean value to signal
// cancellation
func (s *Script) Run(ctx context.Context) {
	ctx = configureContext(ctx, s.Options)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go s.resultProcessor(ctx)
	s.enqueue()

	Logger.Info("Starting Workers")
	for i := 0; i < s.Options.Threads; i++ {
		Logger.WithField("i", i).Warn("Starting worker...")
		s.wg.Add(1)
		go s.worker(ctx)
	}
	s.wg.Wait()
	Logger.Info("All Workers Done")
}

// Add sequences to the queue
func (s *Script) enqueue() {
	for i, seq := range s.sequences {
		seq.Id = i
		s.queue <- seq
	}
	close(s.queue)
	s.sequences = make([]*sequence.Sequence, 0)
}

// DryRun prints the tasks that would be executed by the script if it ran to
// the given io.Writer, but
// does not actually run any of them
func (s *Script) DryRun(w io.Writer) {
	n := (int(s.Options.Input["main"].Len()) * s.Options.MainSequence.Size())
	fmt.Fprintf(w, "DRY RUN %d tasks\n", n)
	s.enqueue()
	for seq := range s.queue {
		fmt.Fprint(w, seq)
	}
}

// Goroutine worker that runs sequences and returns results
func (s *Script) worker(ctx context.Context) {
	defer s.wg.Done()
	for seq := range s.queue {
		seqTimeout := s.Options.TaskTimeout * time.Duration(seq.Size())
		ctx, cancel := context.WithTimeout(ctx, seqTimeout)
		select {
		case s.results <- seq.Execute(ctx):
			Logger.Warn("Sent Result")
			cancel()
		case <-ctx.Done():
			return
		}
	}
	Logger.Warn("worker shutting down...")
}

// Goroutine that receives results and processes them with OutputFunc
func (s *Script) resultProcessor(ctx context.Context, filters ...datasource.FilterFunc) {
	output := s.Options.Output
	defer Logger.Warn("processor shutting down...")
	for {
		select {
		case res := <-s.results:
			Logger.Warn(res)
			if err := res.Err(); err != nil {
				Logger.Error(err)
				continue
			}
			if res.Output.Len() > 0 {
				output.Append(res.Output)
				output.WriteTo(nil)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Instantiates context with global objects based on options
func configureContext(ctx context.Context, opts *options.ScriptOptions) context.Context {
	// Log
	log.SetLevel(opts.Verbosity)
	ctx = log.ToContext(ctx, log.Logger())

	// // HTTP Client
	// if opts.HTTP != nil {
	// 	client := httpclient.New().WithOptions(opts.HTTP)
	// 	ctx = httpclient.ToContext(ctx, client)
	// }

	// Variables
	scriptVars := vars.New()
	if opts.Vars != nil {
		scriptVars = opts.Vars.Copy()
	}

	// Save user-defined options as vars so they can be modified during
	// script execution
	// TODO Refactor to more modular approach to passing task vars
	scriptVars.Set("task_timeout", opts.TaskTimeout)
	scriptVars.Set("verbosity", opts.Verbosity)

	// scriptVars.Set("follow_redirects", opts.HTTP.FollowRedirects)
	// scriptVars.Set("proxy", opts.HTTP.Proxy)
	// scriptVars.Set("headers", opts.HTTP.Headers)
	// scriptVars.Set("params", opts.HTTP.Params)
	// scriptVars.Set("cookies", opts.HTTP.Cookies)
	// scriptVars.Set("body", opts.HTTP.Body)

	ctx = vars.ToContext(ctx, scriptVars)

	// Template Globals
	g := configureGlobal(global.New(), opts)
	ctx = global.ToContext(ctx, g)

	return ctx
}

// Configure builds globals from the provided options
func configureGlobal(g *global.Global, opts *options.ScriptOptions) *global.Global {
	if opts == nil {
		panic("global: Options must not be nil")
	}

	if opts.Vars != nil {
		g.Vars = opts.Vars.Map()
	}

	if len(opts.Input) > 0 {
		g.Inputs = opts.Input
	}

	return g
}
