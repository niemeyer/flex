package main

import (
	"github.com/niemeyer/flex"
	"fmt"
)

type listCmd struct {}

const listUsage = `
flex list

Gets a list of containers from the flex daemon
`

func (c *listCmd) usage() string {
	return listUsage
}

func (c *listCmd) flags() {}

func (c *listCmd) run(args []string) error {
	if len(args) > 0 {
		return errArgs
	}
	config, err := flex.LoadConfig()
	if err != nil {
		return err
	}

	// NewClient will ping the server to test the connection before returning.
	d, err := flex.NewClient(config)
	if err != nil {
		return err
	}
	l, err := d.List()
	if err != nil {
		return err
	}
	fmt.Println(l)
	return err
}
