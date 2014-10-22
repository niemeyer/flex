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
	if len(args) != 0 {
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

	l, err := d.Create("foo", "ubuntu", "trusty", "amd64")
	fmt.Println(l)
	return err
}
