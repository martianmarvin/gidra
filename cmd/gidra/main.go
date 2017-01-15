package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/martianmarvin/gidra/datasource"
	_ "github.com/martianmarvin/gidra/datasource/all"
	_ "github.com/martianmarvin/gidra/task/all"

	"github.com/martianmarvin/gidra/script"
)

// Global config vars
var (
	cfgThreads int
	cfgQuiet   bool
)

func cmdRun(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: gidra run [options] SCRIPT\n\n")
		os.Exit(1)
	}

	s, err := script.OpenFile(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.Options.Threads = cfgThreads

	if s.Options.Output == nil {
		// If no output specified, output tsv to Stdout
		if cfgQuiet {
			s.Options.Output = &datasource.NopWriter{}
		} else {
			w, err := datasource.NewWriter("tsv")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			s.Options.Output = datasource.NewWriteCloser(w, os.Stdout)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	ossigs := make(chan os.Signal, 1)
	signal.Notify(ossigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ossigs
		fmt.Println("Received user signal, shutting down")
		cancel()

	}()

	s.Run(ctx)
}

func cmdShow(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: gidra show [options] SCRIPT\n\n")
		os.Exit(1)
	}

	s, err := script.OpenFile(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.Show(os.Stdout)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "gidra",
		Short: "Gidra is the easiest way to automate a series of web requests.",
		Long:  `Gidra is the easiest way to automate a series of web requests.`,
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a script file",
		Long:  "Run a script file",
		Run:   cmdRun,
	}
	// runCmd.SetUsageTemplate("Usage: gidra run [options] SCRIPT\n\n")
	runCmd.Flags().IntVarP(&cfgThreads, "threads", "t", 100, "number of concurrent threads to run tasks.")
	runCmd.Flags().BoolVarP(&cfgQuiet, "quiet", "q", false, "Don't send results to standard output. Has no effect if an output is set.")
	rootCmd.AddCommand(runCmd)

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show what HTTP requests would be made if the script ran",
		Long:  "Show what HTTP requests would be made if the script ran",
		Run:   cmdShow,
	}
	rootCmd.AddCommand(showCmd)

	rootCmd.Execute()
}
