package main

import (
	"github.com/niemeyer/flex"
)

type pingCmd struct{}

const pingUsage = `
flex ping

Pings the flex daemon to check if it is up and working.
`

func (c *pingCmd) usage() string {
	return pingUsage
}

func (c *pingCmd) flags() {}

func (c *pingCmd) run(args []string) error {
	if len(args) > 0 {
		return errArgs
	}
	config, err := flex.LoadConfig()
	if err != nil {
		return err
	}

	// NewClient will ping the server to test the connection before returning.
	_, err = flex.NewClient(config)
	return err
}
