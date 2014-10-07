package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/niemeyer/flex"
)

type daemonCmd struct{}

const daemonUsage = `
flex daemon

Prints the daemon number of flex.
`

func (c *daemonCmd) usage() string {
	return daemonUsage
}

func (c *daemonCmd) flags() {
}

func (c *daemonCmd) run(args []string) error {
	if len(args) > 0 {
		return errArgs
	}

	config, err := flex.ReadConfig(nil)
	if err != nil {
		return err
	}

	d, err := flex.StartDaemon(config)
	if err != nil {
		return err
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	signal.Notify(ch, syscall.SIGTERM)
	<-ch
	return d.Stop()
}
