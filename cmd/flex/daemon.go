package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/niemeyer/flex"
	"github.com/niemeyer/flex/internal/gnuflag"
)

type daemonCmd struct {
	listenAddr string
}

const daemonUsage = `
flex daemon

Prints the daemon number of flex.
`

func (c *daemonCmd) usage() string {
	return daemonUsage
}

func (c *daemonCmd) flags() {
	gnuflag.StringVar(&c.listenAddr, "tcp", "", "TCP address to listen on in addition to the unix socket")
}

func (c *daemonCmd) run(args []string) error {
	if len(args) > 0 {
		return errArgs
	}

	config, err := flex.LoadConfig()
	if err != nil {
		return err
	}
	config.ListenAddr = c.listenAddr

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
