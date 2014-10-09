package main

import (
	"fmt"

	"github.com/niemeyer/flex"
	"github.com/niemeyer/flex/internal/gnuflag"
)

type testCmd struct {
	short bool
	long  string
}

const testUsage = `
flex test

Exercises the command logic. <= One-line summary of the command.

This command should be removed once the ideas here understood
and exercised elsewhere. <= Optional long description.
`

func (c *testCmd) usage() string {
	return testUsage
}

func (c *testCmd) flags() {
	gnuflag.BoolVar(&c.short, "s", false, "Local short flag")
	gnuflag.StringVar(&c.long, "long", "", "Local long flag")
}

func (c *testCmd) run(args []string) error {
	if len(args) > 0 {
		// Command does not take any positional arguments.
		return errArgs
	}

	flex.Logf("Normal message, visible with -v.")
	flex.Debugf("Debug message, visible with --debug")

	fmt.Println("Short option (-s):", c.short)
	fmt.Println("Long option (--long):", c.long)

	config, err := flex.LoadConfig()
	if err != nil {
		return err
	}
	fmt.Println("Config option test-option:", config.TestOption)

	return nil
}
