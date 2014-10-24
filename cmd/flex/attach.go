package main

import (
	"github.com/niemeyer/flex"
	"fmt"
	"net"
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

	fmt.Println("Ready to attach conn to stdin/stdout")

	return nil
}
