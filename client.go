package flex

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"time"
)

// Client can talk to a flex daemon.
type Client struct {
	config Config
	http   http.Client
}

// NewClient returns a new flex client.
func NewClient(config *Config) (*Client, error) {
	c := Client{
		config: *config,
		http: http.Client{
			Timeout:   10 * time.Second,
			Transport: &unixTransport,
		},
	}
	if err := c.Ping(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Ping pings the daemon to see if it is up listening and working.
func (c *Client) Ping() error {
	Debugf("pinging the daemon")
	data, err := c.getstr("/ping")
	if err != nil {
		return err
	}
	if data != "pong" {
		return fmt.Errorf("unexpected response to daemon ping: %q", data)
	}
	Debugf("pong received")
	return nil
}

func (c *Client) getstr(elem ...string) (string, error) {
	data, err := c.get(elem...)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) get(elem ...string) ([]byte, error) {
	resp, err := c.http.Get(c.url(elem...))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (c *Client) url(elem ...string) string {
	return "http://unix.socket" + path.Join(elem...)
}

var unixTransport = http.Transport{
	Dial: func(network, addr string) (net.Conn, error) {
		if addr != "unix.socket:80" {
			return nil, fmt.Errorf("non-unix-socket addresses not supported yet")
		}
		raddr, err := net.ResolveUnixAddr("unix", varPath("unix.socket"))
		if err != nil {
			return nil, fmt.Errorf("cannot resolve unix socket address: %v", err)
		}
		return net.DialUnix("unix", nil, raddr)
	},
}
