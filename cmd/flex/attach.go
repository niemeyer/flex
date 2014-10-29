package main

import (
	"github.com/niemeyer/flex"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"

	"code.google.com/p/go.crypto/ssh/terminal"
)

type attachCmd struct{}

const attachUsage = `
flex attach images:ubuntu/$release/$arch

Attaches to a container
`

func (c *attachCmd) usage() string {
	return attachUsage
}

func (c *attachCmd) flags() {}

func (c *attachCmd) run(args []string) error {
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

	// TODO - random value in place of 5 :)
	secret := "5"

	l, err := d.Attach(name, "/bin/bash", secret)
	if err != nil {
		return err
	}

	cfd := syscall.Stdout
	if terminal.IsTerminal(cfd) {
		oldttystate, err := terminal.MakeRaw(cfd)
		if err != nil {
			return err
		}
		defer terminal.Restore(cfd, oldttystate)
	}

	// open a connection to l and connect stdin/stdout to it

	// connect
	conn, err := net.Dial("tcp", l)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(secret))
	if err != nil {
		return err
	}

	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			fmt.Println("Stdin read error: %s", err.Error())
			return
		}
	}()
	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		fmt.Println("Connection read error: %s", err.Error())
		return err
	}

	return nil
}
