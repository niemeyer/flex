package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/niemeyer/flex"
	"github.com/niemeyer/flex/internal/gnuflag"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var verbose = gnuflag.Bool("v", false, "Enables verbose mode.")
var debug = gnuflag.Bool("debug", false, "Enables debug mode.")

func run() error {
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		os.Args[1] = "help"
	}
	if len(os.Args) < 2 || os.Args[1] == "" || os.Args[1][0] == '-' {
		return fmt.Errorf("missing subcommand")
	}
	name := os.Args[1]
	cmd, ok := commands[name]
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}
	cmd.flags()
	gnuflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n\nOptions:\n\n", strings.TrimSpace(cmd.usage()))
		gnuflag.PrintDefaults()
	}

	os.Args = os.Args[1:]
	gnuflag.Parse(true)

	if *verbose || *debug {
		flex.SetLogger(log.New(os.Stderr, "", log.LstdFlags))
		flex.SetDebug(*debug)
	}
	return cmd.run(gnuflag.Args())
}

type command interface {
	usage() string
	flags()
	run(args []string) error
}

var commands = map[string]command{
	"version": &versionCmd{},
	"help":    &helpCmd{},
	"daemon":  &daemonCmd{},
	"ping":    &pingCmd{},
	"list":    &listCmd{},
	"create":  &createCmd{},
	"attach":  &attachCmd{},
	"start": &byNameCmd{
		"start",
		func(c *flex.Client, name string) (string, error) { return c.Start(name) },
	},
	"stop": &byNameCmd{
		"stop",
		func(c *flex.Client, name string) (string, error) { return c.Stop(name) },
	},

	// This is a demo command. Drop after ideas are understood.
	"test": &testCmd{},
}

var errArgs = fmt.Errorf("too many subcommand arguments")
