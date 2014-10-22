package flex

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
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
			// Added on Go 1.3. Wait until it's more popular.
			//Timeout:   10 * time.Second,
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

func (c *Client) List() (string, error) {
	Debugf("Getting list from the daemon")
	data, err := c.getstr("/list")
	if err != nil {
		return "fail", err
	}
	return data, err
}

func (c *Client) Create(name string, distro string, release string, arch string) (string, error) {
	Debugf("Creating container")
	url := fmt.Sprintf("/create?name=%s&distro=%s&release=%s&arch=%s", name, distro, release, arch)
	data, err := c.getstr(url)
	if err != nil {
		return "fail", err
	}
	return data, err
}

// Call a function in the flex API by name (i.e. this has nothing to do with
// the parameter passing schemed :)
func (c *Client) CallByName(function string, name string) (string, error) {
	url := fmt.Sprintf("/%s?name=%s", function, name)
	data, err := c.getstr(url)
	if err != nil {
		return "fail", err
	}
	return data, err
}

func (c *Client) Start(name string) (string, error) {
	return c.CallByName("start", name)
}

func (c *Client) Stop(name string) (string, error) {
	return c.CallByName("stop", name)
}

func (c *Client) Status(name string) (string, error) {
	return c.CallByName("status", name)
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
