package flex

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/tomb.v2"
)

// A Daemon can respond to requests from a flex client.
type Daemon struct {
	tomb   tomb.Tomb
	config Config
	l      net.Listener
	mux    *http.ServeMux
}

// varPath returns the provided path elements joined by a slash and
// appended to the end of $FLEX_DIR, which defaults to /var/lib/flex.
func varPath(path ...string) string {
	varDir := os.Getenv("FLEX_DIR")
	if varDir == "" {
		return "/var/lib/flex"
	}
	items := []string{varDir}
	items = append(items, path...)
	return filepath.Join(items...)
}

// StartDaemon starts the flex daemon with the provided configuration.
func StartDaemon(config *Config) (*Daemon, error) {
	d := &Daemon{config: *config}
	d.mux = http.NewServeMux()
	d.mux.HandleFunc("/ping", d.servePing)

	addr, err := net.ResolveUnixAddr("unix", varPath("unix.socket"))
	if err != nil {
		return nil, fmt.Errorf("cannot resolve unix socket address: %v", err)
	}
	l, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on unix socket: %v", err)
	}
	d.l = l
	d.tomb.Go(func() error { return http.Serve(d.l, d.mux) })
	return d, nil
}

var errStop = fmt.Errorf("requested stop")

// Stop stops the flex daemon.
func (d *Daemon) Stop() error {
	d.tomb.Kill(errStop)
	d.l.Close()
	err := d.tomb.Wait()
	if err == errStop {
		return nil
	}
	return err
}

func (d *Daemon) servePing(w http.ResponseWriter, r *http.Request) {
	Debugf("responding to ping")
	w.Write([]byte("pong"))
}
