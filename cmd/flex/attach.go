package main

import (
	"github.com/niemeyer/flex"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
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

	// read the new port from l
	// open a connection to l and connect stdin/stdout to it

	// i have no idea why the extra cruft is in the Addr.String,
	// but remove it
	for {
		if l[0] == '=' {
			newl := l[1:len(l)-1]
			l = newl
			break
		}
		l = l[1:]
	}

	// connect
	conn, err := net.Dial("tcp", l)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(secret))
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			fmt.Println("Connection read error: %s", err.Error())
			return
		}
	}()
	wg.Add(0) // stdin won't hangup when we're done, so don't wait for it
	go func() {
		defer wg.Done()
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			fmt.Println("Stdin read error: %s", err.Error())
			return
		}
	}()

	// FIXME(niemeyer): WaitGroup is being misused here. Add(0) is a NOOP,
	// and the Done below decrements a counter that was not incremented,
	// which will lead to a crash if ever executed, or will unblock the
	// wait group before the goroutines above are done. It's also not clear
	// what is the intent here. The WaitGroup would only unblock once
	// stdout EOFs or something else fails.
	wg.Wait()

	return nil
}
