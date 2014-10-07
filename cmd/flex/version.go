package main

import (
	"fmt"

	"github.com/niemeyer/flex"
)

type versionCmd struct{}

const versionUsage = `
flex version

Prints the version number of flex.
`

func (c *versionCmd) usage() string {
	return versionUsage
}

func (c *versionCmd) flags() {
}

func (c *versionCmd) run(args []string) error {
	if len(args) > 0 {
		return errArgs
	}
	fmt.Println(flex.Version)
	return nil
}
