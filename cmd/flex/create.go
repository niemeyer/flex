package main

import (
	"fmt"
	"github.com/niemeyer/flex"
)

type createCmd struct{}

const createUsage = `
flex create images:ubuntu/$release/$arch

Creates a container using the specified release and arch
`

func (c *createCmd) usage() string {
	return createUsage
}

func (c *createCmd) flags() {}

func (c *createCmd) run(args []string) error {
	name := "foo"
	if len(args) > 1 {
		return errArgs
	}
	if len(args) == 1 {
		name = args[0]
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

	l, err := d.Create(name, "ubuntu", "trusty", "amd64")
	if err == nil {
		fmt.Println(l)
	}
	return err
}
